// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/topicusonderwijs/terraform-provider-octodns/internal/models"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SubdomainDataSource{}

// RTYPE_A      RType = "a"
func NewARecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_A}
}

// RTYPE_AAAA   RType = "aaaa"
func NewAAAARecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_AAAA}
}

// RTYPE_CAA    RType = "caa"
func NewCAARecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_CAA}
}

// RTYPE_CNAME  RType = "cname"
func NewCNAMERecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_CNAME}
}

// RTYPE_DNAME  RType = "dname"
func NewDNAMERecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_DNAME}
}

// RTYPE_LOC    RType = "loc"
func NewLOCRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_LOC}
}

// RTYPE_MX     RType = "mx"
func NewMXRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_MX}
}

// RTYPE_NAPTR  RType = "naptr"
func NewNAPTRRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_NAPTR}
}

// RTYPE_NS     RType = "ns"
func NewNSRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_NS}
}

// RTYPE_PTR    RType = "ptr"
func NewPTRRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_PTR}
}

// RTYPE_SPF    RType = "spf"
func NewSPFRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_SPF}
}

// RTYPE_SRV    RType = "srv"
func NewSRVRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_SRV}
}

// RTYPE_SSHFP  RType = "sshfp"
func NewSSHFPRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_SSHFP}
}

// RTYPE_TXT    RType = "txt"
func NewTXTRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_TXT}
}

// RTYPE_URLFWD RType = "urlfwd"
func NewURLFWDRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_URLFWD}
}

func NewRecordDataSource() datasource.DataSource {
	return &RecordDataSource{}
}

// RecordDataSource defines the data source implementation.
type RecordDataSource struct {
	rtype  *models.RType
	client *models.GitHubClient
}

func (d *RecordDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.rtype.LowerString() + "_record"
}

func (d *RecordDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Record data source",

		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				MarkdownDescription: "Record Zone",
				Required:            true,
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "Scope of zone",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Record Zone",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Record identifier",
				Computed:            true,
			},
			"values": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"ttl": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *RecordDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RecordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RecordModel

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

	rt, err := r.GetType(d.rtype.String())

	if rt.TTL > 0 {
		data.TTL = types.Int64Value(int64(rt.TTL))
	} else {
		data.TTL = types.Int64Null()
	}

	for _, v := range rt.ValuesAsString() {
		data.Values = append(data.Values, types.StringValue(v))
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
