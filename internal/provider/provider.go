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

		Attributes: map[string]schema.Attribute{
			"git_provider": schema.StringAttribute{
				MarkdownDescription: "Git provider, only accepted value is github",
				Optional:            true,
			},
			"github_access_token": schema.StringAttribute{
				MarkdownDescription: "Github personal access token, if empty GithubCli (gh) will be used to get a token",
				Optional:            true,
				Sensitive:           true,
			},
			"github_org": schema.StringAttribute{
				MarkdownDescription: "Github personal access token",
				Required:            true,
			},
			"github_repo": schema.StringAttribute{
				MarkdownDescription: "Github personal access token",
				Required:            true,
			},
			"branch": schema.StringAttribute{
				MarkdownDescription: "The git branch to use",
				Optional:            true,
			},
			"author_name": schema.StringAttribute{
				MarkdownDescription: "The git branch to use",
				Optional:            true,
			},
			"author_email": schema.StringAttribute{
				MarkdownDescription: "The git branch to use",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"scope": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Name of this scope, leave empty for default scope",
						},
						"path": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The git path to the folder containing the yaml files",
						},
						"branch": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The git branch to use for this scope",
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

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	gitprovider := "github"
	githubToken := ""

	// Configuration values are now available.
	if data.GitProvider.IsNull() || data.GitProvider.ValueString() == "github" { /* ... */
		gitprovider = "github"

		if data.GithubAccessToken.IsNull() {

			var err error
			githubToken, err = tokenFromGhCli("https://api.github.com/", true)

			if err != nil || githubToken == "" {
				resp.Diagnostics.AddError(
					"Missing Github API access Configuration",
					"While configuring the provider, the Github access token was not found in "+
						"provider configuration block github_access_token attribute.",
				)
			}

		} else {
			githubToken = data.GithubAccessToken.ValueString()
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
	var err error
	switch gitprovider {
	default:
		client, err = models.NewGitHubClient(githubToken, data.GithubOrg.ValueString(), data.GithubRepo.ValueString())
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
		err = client.AddScope("default", "/zones", "", "")
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
	out, err := exec.Command(ghCliPath, "auth", "token", "--hostname", hostname).Output()
	if err != nil {
		// GH CLI is either not installed or there was no `gh auth login` command issued,
		// which is fine. don't return the error to keep the flow going
		return "", nil
	}

	log.Printf("[INFO] Using the token from GitHub CLI")
	return strings.TrimSpace(string(out)), nil
}
