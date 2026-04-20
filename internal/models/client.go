package models

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/oauth2"
)

const (
	DEFAULT_SCOPE     = "default"
	DEFAULT_PATH      = ""
	DEFAULT_EXTENSION = "yaml"
)

type GitClient interface {
	AddScope(name, path, branch, ext string) error
	SetScope(name, path, branch, ext string) error
	GetZone(zone, scope string) (*Zone, error)
	SetBranch(branch string) error
	SetAuthor(name, email string) error
	MarkZoneDirty(zone *Zone, comment string)
	FlushIfLast() error
}

type GitHubClient struct {
	*github.Client
	Owner         string
	Repo          string
	Scopes        map[string]Scope
	Zones         map[string]*Zone
	Mutex         sync.RWMutex
	Branch        string
	AuthorName    string
	AuthorEmail   string
	RetryLimit    int
	BatchWindow   time.Duration
	dirtyZones    map[string]*Zone
	dirtyComments map[string][]string
	InFlight      atomic.Int64
}

func NewGitHubClient(accessToken, owner, repo string, retryLimit int) (GitClient, error) {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &GitHubClient{
		Client:        github.NewClient(tc),
		Owner:         owner,
		Repo:          repo,
		Zones:         map[string]*Zone{},
		Scopes:        map[string]Scope{},
		Branch:        "main",
		AuthorEmail:   "",
		AuthorName:    "",
		RetryLimit:    retryLimit,
		BatchWindow:   100 * time.Millisecond,
		dirtyZones:    map[string]*Zone{},
		dirtyComments: map[string][]string{},
	}, nil

}

func (g *GitHubClient) SetBranch(branch string) error {
	g.Branch = branch
	return nil
}

func (g *GitHubClient) SetAuthor(name, email string) error {
	g.AuthorName = name
	g.AuthorEmail = email
	return nil
}

func (g *GitHubClient) AddScope(name, path, branch, ext string) error {
	if _, ok := g.Scopes[name]; ok {
		return fmt.Errorf("duplicate scope name found for name `%s`", name)
	}
	return g.SetScope(name, path, branch, ext)
}

func (g *GitHubClient) SetScope(name, path, branch, ext string) error {
	if name == "" {
		name = DEFAULT_SCOPE
	}
	if path == "" {
		path = DEFAULT_PATH
	}
	if ext == "" {
		ext = DEFAULT_EXTENSION
	}
	g.Scopes[name] = NewScope(name, path, branch, ext)
	return nil
}

func (g *GitHubClient) GetScope(name string) (scope Scope, err error) {
	var ok bool
	if name == "" {
		name = DEFAULT_SCOPE
	}
	if scope, ok = g.Scopes[name]; !ok {
		err = fmt.Errorf("undefined scope `%s`", name)
	}
	return
}

func (g *GitHubClient) GetZone(zone, scope string) (*Zone, error) {
	sc, err := g.GetScope(scope)
	if err != nil {
		return nil, err
	}

	filepath := sc.CreateFilePath(zone)

	if _, ok := g.Zones[filepath]; ok {
		return g.Zones[filepath], nil
	}

	options := &github.RepositoryContentGetOptions{Ref: sc.GetBranch(g.Branch)}
	ctx := context.Background()
	fileContent, _, _, err := g.Repositories.GetContents(ctx, g.Owner, g.Repo, filepath, options)
	if err != nil {
		return nil, err
	}

	contents, err := fileContent.GetContent()
	if err != nil {
		return nil, err
	}

	z := Zone{}
	z.name = zone
	z.scope = scope
	z.sha = fileContent.GetSHA()

	err = z.ReadYaml([]byte(contents))
	if err != nil {
		return nil, err
	}
	g.Zones[filepath] = &z
	return &z, nil
}

// MarkZoneDirty queues a zone to be written. Must be called with Mutex held.
func (g *GitHubClient) MarkZoneDirty(zone *Zone, comment string) {
	tflog.Debug(context.Background(), "MarkZoneDirty", map[string]interface{}{"inFlight": g.InFlight.Load()})
	sc, err := g.GetScope(zone.scope)
	if err != nil {
		return
	}
	filepath := sc.CreateFilePath(zone.name)
	g.dirtyZones[filepath] = zone
	g.dirtyComments[filepath] = append(g.dirtyComments[filepath], comment)
}

// FlushIfLast decrements InFlight and, if this was the last operation,
// writes all dirty zones to GitHub in a single call per zone.
//
// Call pattern in each CRUD method — note NO separate defer for InFlight:
//
//	InFlight.Add(+1)          // BEFORE Lock — counts self as queued
//	Mutex.Lock()
//	defer Mutex.Unlock()
//	... do work ...
//	MarkZoneDirty(...)
//	FlushIfLast()             // owns the InFlight.Add(-1)
//
// With 99 parallel creates (parallelism ≥ 2), all goroutines call
// InFlight.Add(+1) before blocking on Lock. Each FlushIfLast decrements
// the counter. When a goroutine decrements to 0 it might not be truly
// last: Terraform dispatches the next wave of goroutines moments after
// the current wave finishes, and those goroutines call Add(+1) before
// acquiring the mutex.
//
// To bridge this gap: when InFlight reaches 0, sleep for BatchWindow
// while still holding the mutex. Newly-dispatched goroutines can call
// InFlight.Add(+1) immediately (before Lock) but then block on Lock,
// making their arrival visible in InFlight. After the sleep, if
// InFlight > 0 the flush is skipped and the later goroutines take over.
// If InFlight is still 0, no new work is coming and we flush.
//
// The sleep is synchronous inside Create/Update/Delete, so the provider
// process cannot exit during it — Terraform waits for these calls to
// return before considering the apply complete.
//
// Must be called with Mutex held.
func (g *GitHubClient) FlushIfLast() error {
	remaining := g.InFlight.Add(-1)
	tflog.Debug(context.Background(), "FlushIfLast", map[string]interface{}{"remaining": remaining, "dirty": len(g.dirtyZones)})
	if remaining > 0 {
		return nil
	}
	// InFlight just hit 0. Wait one grace window for any goroutines that
	// Terraform is about to dispatch — they will call InFlight.Add(+1)
	// before trying to Lock, so we'll see them after the sleep.
	time.Sleep(g.BatchWindow)
	if g.InFlight.Load() > 0 {
		tflog.Debug(context.Background(), "FlushIfLast: new operations arrived during grace window, skipping flush")
		return nil
	}
	if len(g.dirtyZones) == 0 {
		return nil
	}
	tflog.Debug(context.Background(), "FlushIfLast: flushing dirty zones", map[string]interface{}{"count": len(g.dirtyZones)})
	for filepath, zone := range g.dirtyZones {
		comments := g.dirtyComments[filepath]
		var comment string
		if len(comments) == 1 {
			comment = comments[0]
		} else {
			var creates, updates, deletes int
			for _, c := range comments {
				switch {
				case strings.Contains(c, ": create "):
					creates++
				case strings.Contains(c, ": update "):
					updates++
				case strings.Contains(c, ": delete "):
					deletes++
				}
			}
			parts := []string{}
			if creates > 0 {
				parts = append(parts, fmt.Sprintf("%d creates", creates))
			}
			if updates > 0 {
				parts = append(parts, fmt.Sprintf("%d updates", updates))
			}
			if deletes > 0 {
				parts = append(parts, fmt.Sprintf("%d deletes", deletes))
			}
			comment = fmt.Sprintf("chore(%s/%s): %d changes (%s)", zone.scope, zone.name, len(comments), strings.Join(parts, ", "))
		}
		if err := g.SaveZone(zone, comment); err != nil {
			return err
		}
		delete(g.dirtyZones, filepath)
		delete(g.dirtyComments, filepath)
	}
	return nil
}

func (g *GitHubClient) SaveZone(zone *Zone, comment string) error {
	content, err := zone.WriteYaml()
	if err != nil {
		return err
	}

	var scope Scope
	scope, err = g.GetScope(zone.scope)
	if err != nil {
		return err
	}

	filepath := scope.CreateFilePath(zone.name)
	sha := zone.sha

	if comment == "" {
		comment = fmt.Sprintf("chore(%s/%s): updating records", zone.scope, zone.name)
	}

	var author *github.CommitAuthor = nil
	if g.AuthorName != "" || g.AuthorEmail != "" {
		author = &github.CommitAuthor{}
		if g.AuthorName != "" {
			author.Name = github.String(g.AuthorName)
		}
		if g.AuthorEmail != "" {
			author.Email = github.String(g.AuthorEmail)
		}
	}

	commitOption := &github.RepositoryContentFileOptions{
		Branch:    github.String(scope.GetBranch(g.Branch)),
		Message:   github.String(comment),
		Committer: author,
		Author:    author,
		Content:   content,
		SHA:       &sha,
	}

	ctx := context.Background()
	_, response, err := g.Repositories.UpdateFile(ctx, g.Owner, g.Repo, filepath, commitOption)
	if response != nil && response.StatusCode == 409 {
		return fmt.Errorf("409 error:`%v`", err)
	}
	if err != nil {
		return err
	}

	delete(g.Zones, filepath)

	return nil
}
