// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/topicusonderwijs/terraform-provider-octodns/internal/models"
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

	Scopes []struct {
		Name types.String `tfsdk:"name"`
		Path types.String `tfsdk:"path"`
	} `tfsdk:"scope"`
}

func (p *OctodnsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "octodns"
	resp.Version = p.version
}

/*
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ""},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opt := &github.RepositoryListOptions{ListOptions: github.ListOptions{PerPage: 2}}
	opt.Page = 2

	//(fileContent *RepositoryContent, directoryContent []*RepositoryContent, resp *Response, err error) {

	fileContent, dirContent, resp, err := client.Repositories.GetContents(ctx, "topicusonderwijs", "dns-topicus.education", "overlays/pdc/zones/pdc.topicus.education.yaml", nil)


*/

func (p *OctodnsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{

		Attributes: map[string]schema.Attribute{
			"git_provider": schema.StringAttribute{
				MarkdownDescription: "Git provider, only accepted value is github",
				Optional:            true,
			},
			"github_access_token": schema.StringAttribute{
				MarkdownDescription: "Github personal access token",
				Required:            true,
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
		},
		Blocks: map[string]schema.Block{
			"scope": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional: true,
						},
						"path": schema.StringAttribute{
							Required: true,
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

	// Configuration values are now available.
	if data.GitProvider.IsNull() { /* ... */
		gitprovider = "github"

		if data.GithubAccessToken.IsNull() {
			resp.Diagnostics.AddError(
				"Missing Github API access Configuration",
				"While configuring the provider, the Github access token was not found in "+
					"provider configuration block github_access_token attribute.",
			)
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

	} else if data.GitProvider.ValueString() != "github" { /* ... */
		resp.Diagnostics.AddWarning(
			"Unsupported Git Provider Configuration",
			"While configuring the provider, an invalid value was found for git_provider attribute. "+
				"Allowed values: github ",
		)
	}

	var client models.GitClient
	var err error
	switch gitprovider {
	default:
		client, err = models.NewGitHubClient(data.GithubAccessToken.ValueString(), data.GithubOrg.ValueString(), data.GithubRepo.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create Github client",
			"While configuring the provider, the Github client failed to configure: "+
				err.Error(),
		)
	}

	if len(data.Scopes) == 0 {
		err = client.AddScope("default", "/zones")
	} else {
		for _, v := range data.Scopes {

			err = client.AddScope(v.Name.ValueString(), v.Path.ValueString())
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
		NewSubdomainResource,
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
		NewSubdomainDataSource,
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
