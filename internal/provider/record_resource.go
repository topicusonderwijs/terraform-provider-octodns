// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/topicusonderwijs/terraform-provider-octodns/internal/models"
	"log"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RecordResource{}
var _ resource.ResourceWithImportState = &RecordResource{}

// RTYPE_A      RType = "a"
func NewARecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_A}
}

// RTYPE_AAAA   RType = "aaaa"
func NewAAAARecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_AAAA}
}

// RTYPE_CAA    RType = "caa"
func NewCAARecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_CAA}
}

// RTYPE_CNAME  RType = "cname"
func NewCNAMERecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_CNAME}
}

// RTYPE_DNAME  RType = "dname"
func NewDNAMERecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_DNAME}
}

// RTYPE_LOC    RType = "loc"
func NewLOCRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_LOC}
}

// RTYPE_MX     RType = "mx"
func NewMXRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_MX}
}

// RTYPE_NAPTR  RType = "naptr"
func NewNAPTRRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_NAPTR}
}

// RTYPE_NS     RType = "ns"
func NewNSRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_NS}
}

// RTYPE_PTR    RType = "ptr"
func NewPTRRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_PTR}
}

// RTYPE_SPF    RType = "spf"
func NewSPFRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_SPF}
}

// RTYPE_SRV    RType = "srv"
func NewSRVRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_SRV}
}

// RTYPE_SSHFP  RType = "sshfp"
func NewSSHFPRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_SSHFP}
}

// RTYPE_TXT    RType = "txt"
func NewTXTRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_TXT}
}

// RTYPE_URLFWD RType = "urlfwd"
func NewURLFWDRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_URLFWD}
}

func NewRecordResource() resource.Resource {
	return &RecordResource{rtype: nil}
}

// RecordResource defines the resource implementation.
type RecordResource struct {
	rtype  *models.RType
	client *models.GitHubClient
}

func (r *RecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.rtype.LowerString() + "_record"
}

func (r *RecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Record resource",

		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				MarkdownDescription: "Record Zone",
				Required:            true,
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "Scope of zone",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Record Zone",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Record identifier",
				Computed:            true,
			},
			"values": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"ttl": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(3600),
			},
		},
	}
}

func (r *RecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*models.GitHubClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *RecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SubdomainModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	zone, err := r.client.GetZone(data.Zone.ValueString(), data.Scope.ValueString())
	tflog.Trace(ctx, fmt.Sprintf("==== After Zone ==== %s", ""))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not retrieve zone: %s", err.Error()))
		return
	}

	record, err := zone.FindRecord(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read record, got error: %s", err))
		return
	}

	record.FindAllType()

	//data.Type = []TypeModel{}

	for _, t := range record.Types {

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

	// Set ID
	data.Id = types.StringValue(fmt.Sprintf("%s %s %s", data.Scope.ValueString(), data.Zone.ValueString(), data.Name.ValueString()))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// UpdateYaml data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *SubdomainModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parts := strings.Split(data.Id.ValueString(), " ")
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Malformed ID: %s", data.Id.String()))
		return
	}

	data.Scope = types.StringValue(parts[0])
	data.Zone = types.StringValue(parts[1])
	data.Name = types.StringValue(parts[2])

	tflog.Trace(ctx, fmt.Sprintf("==== Trying to load %s from  %s/%s", data.Name.ValueString(), data.Scope.ValueString(), data.Zone.ValueString()))

	zone, err := r.client.GetZone(data.Zone.ValueString(), data.Scope.ValueString())
	tflog.Trace(ctx, fmt.Sprintf("==== After Zone ==== %s", ""))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not retrieve zone: %s", err.Error()))
		return
	}

	record, err := zone.FindRecord(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read record, got error: %s", err))
		return
	}

	record.FindAllType()

	for _, t := range record.Types {

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
	// UpdateYaml updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SubdomainModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update record, got error: %s", err))
	//     return
	// }

	// UpdateYaml updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SubdomainModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete record, got error: %s", err))
	//     return
	// }
}

func (r *RecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
