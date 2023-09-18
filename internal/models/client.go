package models

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
	"strings"
)

const (
	DEFAULT_SCOPE = "default"
	DEFAULT_PATH  = ""
)

type GitClient interface {
	AddScope(name string, path string) error
	SetScope(name string, path string) error
	GetZone(zone, scope string) (*Zone, error)
}

type GitHubClient struct {
	*github.Client
	Owner   string
	Repo    string
	Scopes  map[string]string
	Counter int
}

func NewGitHubClient(accessToken, owner, repo string) (GitClient, error) {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &GitHubClient{
		Client: github.NewClient(tc),
		Owner:  owner,
		Repo:   repo,
	}, nil

}

func (g *GitHubClient) AddScope(name string, path string) error {

	if _, ok := g.Scopes[name]; ok {
		return fmt.Errorf("duplicate scope name found for name `%s`", name)
	}

	return g.SetScope(name, path)
}

func (g *GitHubClient) SetScope(name string, path string) error {

	if g.Scopes == nil {
		g.Scopes = make(map[string]string)
	}

	if name == "" {
		name = DEFAULT_SCOPE
	}

	if path == "" {
		path = DEFAULT_PATH
	}

	if strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") {
		return fmt.Errorf("scope path must not start or end with a `/`, got '%s' for scope '%s'", path, name)
	}

	g.Scopes[name] = path

	return nil
}

func (g *GitHubClient) GetZoneContents(zone, scope string) ([]byte, error) {

	var path string
	var ok bool
	if path, ok = g.Scopes[scope]; !ok {
		return nil, fmt.Errorf("undefined scope `%s`", scope)
	}
	filepath := path + "/" + zone + ".yaml"

	ctx := context.Background()
	fileContent, _, resp, err := g.Client.Repositories.GetContents(ctx, g.Owner, g.Repo, filepath, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("Nextpage:", resp.NextPage)

	content, err := base64.StdEncoding.DecodeString(*fileContent.Content)
	if err != nil {
		return nil, err
	}

	return content, nil

}

func (g *GitHubClient) GetZone(zone, scope string) (*Zone, error) {

	if scope == "" {
		scope = DEFAULT_SCOPE
	}
	contents, err := g.GetZoneContents(zone, scope)
	if err != nil {
		return nil, err
	}

	z := Zone{}

	err = z.ReadYaml(contents)
	if err != nil {
		return nil, err
	} else {
		return &z, nil
	}

}
