// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/topicusonderwijs/terraform-provider-octodns/internal/models"
	"log"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SubdomainDataSource{}

func NewSubdomainDataSource() datasource.DataSource {
	return &SubdomainDataSource{}
}

// SubdomainDataSource defines the data source implementation.
type SubdomainDataSource struct {
	client *models.GitHubClient
}

func (d *SubdomainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subdomain"
}

func (d *SubdomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Subdomain data source",

		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone of subdomain",
				Required:            true,
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "Scope of zone",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of subdomain",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Subdomain identifier",
				Computed:            true,
			},
			"type": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Computed: true,
						},
						"values": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"ttl": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *SubdomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*models.GitHubClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *models.GitHubClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *SubdomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubdomainModel

	if data.Name.String() == "" {
		data.Name = types.StringValue("@")
	}

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {

		return
	}

	tflog.Trace(ctx, fmt.Sprintf("==== Trying to load %s from  %s/%s", data.Name.ValueString(), data.Scope.ValueString(), data.Zone.ValueString()))

	zone, err := d.client.GetZone(data.Zone.ValueString(), data.Scope.ValueString())
	tflog.Trace(ctx, fmt.Sprintf("==== After Zone ==== %s", ""))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not retrieve zone: %s", err.Error()))
		return
	}

	r, err := zone.FindRecord(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read record, got error: %s", err))
		return
	}

	r.FindAllType()

	//	data.Type = make(map[string]TypeModel)

	for _, t := range r.Types {

		log.Print("Loop Types ", t.Type)

		rt := TypeModel{
			Type: types.StringValue(t.Type),
			TTL:  types.Int64Value(int64(t.TTL)),
		}
		for _, v := range t.ValuesAsString() {
			rt.Values = append(rt.Values, types.StringValue(v))
		}

		data.Type = append(data.Type, rt)
	}

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue(fmt.Sprintf("%s %s %s", data.Scope.ValueString(), data.Zone.ValueString(), data.Name.ValueString()))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// UpdateYaml data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
