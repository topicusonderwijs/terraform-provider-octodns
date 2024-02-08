package models

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
	"strings"
	"sync"
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
}

type GitHubClient struct {
	*github.Client
	Owner       string
	Repo        string
	Scopes      map[string]Scope
	Counter     int
	Zones       map[string]*Zone
	Mutex       sync.RWMutex
	Branch      string
	AuthorName  string
	AuthorEmail string
}

func NewGitHubClient(accessToken, owner, repo string) (GitClient, error) {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &GitHubClient{
		Client:      github.NewClient(tc),
		Owner:       owner,
		Repo:        repo,
		Zones:       map[string]*Zone{},
		Scopes:      map[string]Scope{},
		Branch:      "main",
		AuthorEmail: "",
		AuthorName:  "",
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

	// Trim off "/" characters from the front and end
	path = strings.Trim(path, "/")

	g.Scopes[name] = Scope{
		Name:   name,
		Path:   path,
		Branch: branch,
		Ext:    ext,
	}

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

func (g *GitHubClient) createFilePath(zone, scope string) (string, error) {

	sc, err := g.GetScope(scope)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s.%s", sc.Path, zone, sc.Ext), nil

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
	fileContent, directoryContent, resp, err := g.Client.Repositories.GetContents(ctx, g.Owner, g.Repo, filepath, options)
	if err != nil {
		return nil, err
	}
	_ = resp

	//@todo: Handle multiple pages
	//fmt.Println("Nextpage:", resp.NextPage)

	contents, err := base64.StdEncoding.DecodeString(*fileContent.Content)
	if err != nil {
		return nil, err
	}

	z := Zone{}

	z.name = zone
	z.scope = scope

	if len(directoryContent) > 0 {
		z.sha = *directoryContent[0].SHA
	}

	err = z.ReadYaml(contents)
	g.Zones[filepath] = &z
	if err != nil {
		return nil, err
	} else {
		return &z, nil
	}

}

func (g *GitHubClient) getSHAForFile(filepath string) (string, error) {

	opt := &github.CommitsListOptions{
		Path: filepath,
	}
	commits, _, err := g.Client.Repositories.ListCommits(context.Background(), g.Owner, g.Repo, opt)
	if err != nil {
		return "", err
	}
	commit := commits[0]
	t, _, err := g.Client.Git.GetTree(context.Background(), g.Owner, g.Repo, commit.GetSHA(), true)
	if err != nil {
		return "", err
	}

	for _, entry := range t.Entries {
		if *entry.Path == filepath {
			return *entry.SHA, nil
		}
	}

	return "", nil
}

func (g *GitHubClient) SaveZone(zone *Zone, comment string) error {

	content, err := zone.WriteYaml()
	if err != nil {
		return err
	}

	//err = os.WriteFile("test.out.yaml", content, 0666)
	//return fmt.Errorf("Failure")

	var scope Scope
	scope, err = g.GetScope(zone.scope)
	if err != nil {
		return err
	}

	filepath := scope.CreateFilePath(zone.name)

	sha, err := g.getSHAForFile(filepath)
	if err != nil {
		return err
	}

	_ = filepath

	_ = content

	if comment == "" {
		comment = fmt.Sprintf("Updating records for %s", zone.name)
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
	repositoryContentResponse, response, err := g.Client.Repositories.UpdateFile(ctx, g.Owner, g.Repo, filepath, commitOption)
	if err != nil {
		return err
	}

	_ = repositoryContentResponse
	_ = response

	return err
}
