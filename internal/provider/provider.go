// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/topicusonderwijs/terraform-provider-octodns/internal/models"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

const ENVPREFIX = "OCTODNS_"

// Ensure OctodnsProvider satisfies various provider interfaces.
var _ provider.Provider = &OctodnsProvider{}

// OctodnsProvider defines the provider implementation.
type OctodnsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// OctodnsProviderModel describes the provider data model.
type OctodnsProviderModel struct {
	GitProvider       types.String `tfsdk:"git_provider"`
	GithubAccessToken types.String `tfsdk:"github_access_token"`
	GithubOrg         types.String `tfsdk:"github_org"`
	GithubRepo        types.String `tfsdk:"github_repo"`
	GithubRetryLimit  types.Int32  `tfsdk:"github_retry_limit"`

	GitBranch      types.String `tfsdk:"branch"`
	GitAuthorName  types.String `tfsdk:"author_name"`
	GitAuthorEmail types.String `tfsdk:"author_email"`

	Scopes []struct {
		Name   types.String `tfsdk:"name"`
		Path   types.String `tfsdk:"path"`
		Branch types.String `tfsdk:"branch"`
	} `tfsdk:"scope"`
}

func (p *OctodnsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "octodns"
	resp.Version = p.version
}

func (p *OctodnsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "**Warning**: This provider is still a work-in-progress so use at your own risk\n\n" +
			"This provider allows you to modify your OctoDNS zone yaml files within a github repo,\n" +
			"and can handle multiple zone directories within one git repo by defining multiple scopes\n\n" +
			"For github authentication you can use a personal access token (PAT) or use the [Github Cli](https://cli.github.com) to provide a token.\n" +
			"If you don't have `gh` in your $PATH, you can point to the executable using the GH_PATH environment variable.   \n*Example*: ```GH_PATH=/opt/homebrew/bin/gh terraform plan```\n\n" +
			"note: This provider can only manage records within existing zone files, it **cannot** manage/create zone files or alter the OctoDNS config.\n\n" +
			"Also this provider does not run OctoDNS after a modification, so you need your own automation for that like the OctoDNS github action",
		Attributes: map[string]schema.Attribute{
			"git_provider": schema.StringAttribute{
				MarkdownDescription: "Git provider, only accepted/supported value for now is github",
				Optional:            true,
			},
			"github_access_token": schema.StringAttribute{
				MarkdownDescription: "Github personal access token, if not set the environment variable `GITHUB_TOKEN` or the `Github Cli (gh)` command will be used to get a token",
				Optional:            true,
				Sensitive:           true,
			},
			"github_org": schema.StringAttribute{
				MarkdownDescription: "Github organisation",
				Required:            true,
			},
			"github_repo": schema.StringAttribute{
				MarkdownDescription: "Github repository",
				Required:            true,
			},
			"github_retry_limit": schema.Int32Attribute{
				MarkdownDescription: "How many times to retry updating files in github",
				Optional:            true,
			},
			"branch": schema.StringAttribute{
				MarkdownDescription: "The git branch to use, defaults to main",
				Optional:            true,
			},
			"author_name": schema.StringAttribute{
				MarkdownDescription: "The Author name used in commits, defaults to owner of github token",
				Optional:            true,
			},
			"author_email": schema.StringAttribute{
				MarkdownDescription: "The Author email used in commits, defaults to owner of github token",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"scope": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Unique name of this scope, leave empty for default scope.",
						},
						"path": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The git path to the folder containing the yaml files",
						},
						"branch": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The git branch to use for this scope, defaults to provider branch setting",
						},
					},
				},
				CustomType:          nil,
				Description:         "",
				MarkdownDescription: "",
				DeprecationMessage:  "",
				Validators:          nil,
			},
		},
	}
}

func (p *OctodnsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data OctodnsProviderModel
	var err error

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	gitprovider := "github"
	githubToken := ""
	githubRetryLimit := 5

	// Configuration values are now available.
	if data.GitProvider.IsNull() || data.GitProvider.ValueString() == "github" { /* ... */
		gitprovider = "github"

		// First check if accesstoken is configured
		if !data.GithubAccessToken.IsNull() {
			githubToken = data.GithubAccessToken.ValueString()
		}

		// If not check if env variable GITHUB_TOKEN is set
		if githubToken == "" {
			githubToken = os.Getenv("GITHUB_TOKEN")
		}

		// If still no token set try GitHub CLI command
		if githubToken == "" {
			githubToken, err = tokenFromGhCli("https://api.github.com/", true)
		}

		// No more sources for a token so error
		if err != nil || githubToken == "" {
			resp.Diagnostics.AddError(
				"Missing Github API access Configuration",
				"While configuring the provider, the Github access token was not found in "+
					"provider configuration block github_access_token attribute.",
			)
		}

		if !data.GithubRetryLimit.IsNull() {
			githubRetryLimit = int(data.GithubRetryLimit.ValueInt32())
		}

		if data.GithubOrg.IsNull() {
			resp.Diagnostics.AddError(
				"Missing Github Organisation Configuration",
				"While configuring the provider, the Github Organisation was not found in "+
					"provider configuration block github_org attribute.",
			)
		}
		if data.GithubRepo.IsNull() {
			resp.Diagnostics.AddError(
				"Missing Github repo Configuration",
				"While configuring the provider, the Github repo was not found in "+
					"provider configuration block github_repo attribute.",
			)
		}

	} else { /* ... */
		resp.Diagnostics.AddError(
			"Unsupported Git Provider Configuration",
			"While configuring the provider, an invalid value was found for git_provider attribute. "+
				"Allowed values: github ",
		)
	}

	if data.GitBranch.IsNull() {
		data.GitBranch = types.StringValue("main")
	}

	var client models.GitClient

	switch gitprovider {
	default:
		client, err = models.NewGitHubClient(githubToken, data.GithubOrg.ValueString(), data.GithubRepo.ValueString(), githubRetryLimit)
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create Github client",
			"While configuring the provider, the Github client failed to configure: "+
				err.Error(),
		)
	}

	_ = client.SetBranch(data.GitBranch.ValueString())
	_ = client.SetAuthor(data.GitAuthorName.ValueString(), data.GitAuthorEmail.ValueString())

	if len(data.Scopes) == 0 {
		// Add scope will add the default values for "" parameters
		_ = client.AddScope("", "", "", "")
	} else {
		for _, v := range data.Scopes {

			err = client.AddScope(v.Name.ValueString(), v.Path.ValueString(), v.Branch.ValueString(), "")
			if err != nil {
				resp.Diagnostics.AddError("Could not add scope", err.Error())
			}

		}
	}

	// Record client configuration for data sources and resources

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OctodnsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewARecordResource,
		NewAAAARecordResource,
		NewCAARecordResource,
		NewCNAMERecordResource,
		NewDNAMERecordResource,
		NewLOCRecordResource,
		NewMXRecordResource,
		NewNAPTRRecordResource,
		NewNSRecordResource,
		NewPTRRecordResource,
		NewSPFRecordResource,
		NewSRVRecordResource,
		NewSSHFPRecordResource,
		NewTXTRecordResource,
		NewURLFWDRecordResource,
	}
}

func (p *OctodnsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewARecordDataSource,
		NewAAAARecordDataSource,
		NewCAARecordDataSource,
		NewCNAMERecordDataSource,
		NewDNAMERecordDataSource,
		NewLOCRecordDataSource,
		NewMXRecordDataSource,
		NewNAPTRRecordDataSource,
		NewNSRecordDataSource,
		NewPTRRecordDataSource,
		NewSPFRecordDataSource,
		NewSRVRecordDataSource,
		NewSSHFPRecordDataSource,
		NewTXTRecordDataSource,
		NewURLFWDRecordDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OctodnsProvider{
			version: version,
		}
	}
}

// See https://github.com/integrations/terraform-provider-github/issues/1822
func tokenFromGhCli(baseURL string, isGithubDotCom bool) (string, error) {
	ghCliPath := os.Getenv("GH_PATH")
	if ghCliPath == "" {
		ghCliPath = "gh"
	}
	hostname := ""
	if isGithubDotCom {
		hostname = "github.com"
	} else {
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			return "", fmt.Errorf("parse %s: %w", baseURL, err)
		}
		hostname = parsedURL.Host
	}
	// GitHub CLI uses different base URLs in ~/.config/gh/hosts.yml, so when
	// we're using the standard base path of this provider, it doesn't align
	// with the way `gh` CLI stores the credentials. The following doesn't work:
	//
	// $ gh auth token --hostname api.github.com
	// > no oauth token
	//
	// ... but the following does work correctly
	//
	// $ gh auth token --hostname github.com
	// > gh..<valid token>
	hostname = strings.TrimPrefix(hostname, "api.")
	out, _ := exec.Command(ghCliPath, "auth", "token", "--hostname", hostname).Output()
	// GH CLI is either not installed or there was no `gh auth login` command issued,
	// which is fine. don't return the error to keep the flow going

	log.Printf("[INFO] Using the token from GitHub CLI")
	return strings.TrimSpace(string(out)), nil
}
