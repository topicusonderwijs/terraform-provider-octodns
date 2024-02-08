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
var _ datasource.DataSource = &RecordDataSource{}

func NewARecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_A}
}
func NewAAAARecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_AAAA}
}
func NewCAARecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_CAA}
}
func NewCNAMERecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_CNAME}
}
func NewDNAMERecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_DNAME}
}
func NewLOCRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_LOC}
}
func NewMXRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_MX}
}
func NewNAPTRRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_NAPTR}
}
func NewNSRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_NS}
}
func NewPTRRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_PTR}
}
func NewSPFRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_SPF}
}
func NewSRVRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_SRV}
}
func NewSSHFPRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_SSHFP}
}
func NewTXTRecordDataSource() datasource.DataSource {
	return &RecordDataSource{rtype: &models.TYPE_TXT}
}
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
		MarkdownDescription: d.rtype.String() + " record data source",

		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone of the record",
				Required:            true,
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "Scope of zone",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Record Name",
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
				MarkdownDescription: "Values of the record, should confirm to record type",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "TTL of the record, if not set the zone's or dns server setting is used",
				Computed:            true,
			},
			"octodns": schema.SingleNestedAttribute{
				MarkdownDescription: "Additional octodns config for the records",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"cloudflare": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"proxied": schema.BoolAttribute{
								MarkdownDescription: "Should cloudflare proxy this record (only for A/AAAA records)",
								Computed:            true,
							},
							"auto_ttl": schema.BoolAttribute{
								MarkdownDescription: "Use cloudflare's auto-ttl *feature*, aka: set to 300",
								Computed:            true,
							},
						},
						Computed: true,
					},
					"azuredns": schema.SingleNestedAttribute{
						MarkdownDescription: "Azure healthcheck configuration",
						Attributes: map[string]schema.Attribute{
							"hc_interval": schema.Int64Attribute{
								MarkdownDescription: "Azure healthcheck interval",
								Computed:            true,
							},
							"hc_timeout": schema.Int64Attribute{
								MarkdownDescription: "Azure healthcheck timeout",
								Computed:            true,
							},
							"hc_numfailures": schema.Int64Attribute{
								MarkdownDescription: "Azure healthcheck number of failures allowed",
								Computed:            true,
							},
						},
						Computed: true,
					},
				},
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

	r, err := zone.FindSubdomain(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read record, got error: %s", err))
		return
	}

	//resp.Diagnostics.AddWarning("Client Debug", fmt.Sprintf("Could not retrieve record type: %v", r.Types))

	record, err := r.GetType(d.rtype.String())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not retrieve record type: %s", err.Error()))
		return
	}

	if record.TTL > 0 {
		data.TTL = types.Int64Value(int64(record.TTL))
	} else {
		data.TTL = types.Int64Null()
	}

	for _, v := range record.ValuesAsString() {
		data.Values = append(data.Values, types.StringValue(v))
	}

	odns := OctodnsConfigModel{}

	if record.Octodns.Cloudflare != nil {
		odns.Cloudflare = &OctodnsCloudflareModel{}

		if record.Octodns.Cloudflare.Proxied {
			odns.Cloudflare.Proxied = types.BoolValue(true)
		}
		if record.Octodns.Cloudflare.AutoTTL {
			odns.Cloudflare.AutoTTL = types.BoolValue(true)
		}

	}

	if record.Octodns.AzureDNS != nil {
		AzureDNS := &OctodnsAzureDNSModel{}
		isSet := false

		if record.Octodns.AzureDNS.Healthcheck.Interval > 0 {
			odns.AzureDNS.HCInterval = types.Int64Value(int64(record.Octodns.AzureDNS.Healthcheck.Interval))
			isSet = true
		}
		if record.Octodns.AzureDNS.Healthcheck.Timeout > 0 {
			odns.AzureDNS.HCTimeout = types.Int64Value(int64(record.Octodns.AzureDNS.Healthcheck.Timeout))
			isSet = true
		}
		if record.Octodns.AzureDNS.Healthcheck.NumFailures > 0 {
			odns.AzureDNS.HCNumFailures = types.Int64Value(int64(record.Octodns.AzureDNS.Healthcheck.NumFailures))
			isSet = true
		}
		if isSet {
			odns.AzureDNS = AzureDNS
		}

	}

	if odns.HasConfig() {
		data.Octodns = &odns
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
