// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/topicusonderwijs/terraform-provider-octodns/internal/models"
	"strings"
	"time"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RecordResource{}
var _ resource.ResourceWithImportState = &RecordResource{}

func NewARecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_A}
}
func NewAAAARecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_AAAA}
}
func NewCAARecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_CAA}
}
func NewCNAMERecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_CNAME}
}
func NewDNAMERecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_DNAME}
}
func NewLOCRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_LOC}
}
func NewMXRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_MX}
}
func NewNAPTRRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_NAPTR}
}
func NewNSRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_NS}
}
func NewPTRRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_PTR}
}
func NewSPFRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_SPF}
}
func NewSRVRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_SRV}
}
func NewSSHFPRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_SSHFP}
}
func NewTXTRecordResource() resource.Resource {
	return &RecordResource{rtype: &models.TYPE_TXT}
}
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
		MarkdownDescription: r.rtype.String() + " record resource",

		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				MarkdownDescription: "Zone of the record. eq: example.com",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "Scope of zone",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(models.DEFAULT_SCOPE),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Record name. eq: <name>.example.com",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Record identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"values": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"ttl": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(3600),
				MarkdownDescription: "TTL of the record, leave empty for zone of server defaults",
			},
			"octodns": schema.SingleNestedAttribute{
				MarkdownDescription: "Additional provider specific record meta config.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"cloudflare": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "Meta config for [cloudflare provider](https://github.com/octodns/octodns-cloudflare/?tab=readme-ov-file#configuration)",
						Attributes: map[string]schema.Attribute{
							"proxied": schema.BoolAttribute{
								MarkdownDescription: "Should cloudflare proxy this record (only for A/AAAA/CNAME records)",
								Optional:            true,
							},
							"auto_ttl": schema.BoolAttribute{
								MarkdownDescription: "Use cloudflare's auto-ttl *feature*, aka: set to 300",
								Optional:            true,
							},
						},
					},
					"azuredns": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "Healthcheck configuration for [Azure provider](https://github.com/octodns/octodns-azure/?tab=readme-ov-file#healthchecks)",
						Attributes: map[string]schema.Attribute{
							"hc_interval": schema.Int64Attribute{
								MarkdownDescription: "Azure healthcheck interval",
								Optional:            true,
							},
							"hc_timeout": schema.Int64Attribute{
								MarkdownDescription: "Azure healthcheck timeout",
								Optional:            true,
							},
							"hc_numfailures": schema.Int64Attribute{
								MarkdownDescription: "Azure healthcheck number of failures allowed",
								Optional:            true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *RecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	tflog.Trace(ctx, "- Resource Configure")

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
	tflog.Trace(ctx, "- Resource Create")
	var data *RecordModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.client.Mutex.Lock()
	defer r.client.Mutex.Unlock()

	// Retry loop
	var zone *models.Zone
	var subdomain models.Subdomain
	var record *models.Record
	var err error

	retryCounter := 0
	retryLimit := 5

	if retryCounter <= retryLimit {

		zone, err = r.client.GetZone(data.Zone.ValueString(), data.Scope.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not retrieve zone: %s", err.Error()))
			return
		}

		subdomain, err = zone.CreateSubdomain(data.Name.ValueString())
		if err != nil {
			if !errors.Is(err, models.SubdomainAlreadyExistsError) {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create subdomain, got error: %s", err))
				return
			}
		}

		record, err = subdomain.CreateType(r.rtype.String())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create type record, got error: %s", err))
			return
		}

		resp.Diagnostics.Append(RecordFromDataModel(ctx, data, record)...)
		if resp.Diagnostics.HasError() {
			return
		}

		err = subdomain.UpdateYaml()
		if err != nil {
			resp.Diagnostics.AddError("Yaml Error", fmt.Sprintf("Unable to update subdomain in yaml, got error: %s", err))
			return
		}

		err = r.client.SaveZone(zone, fmt.Sprintf("chore(%s): create %s record for %s", data.Zone.ValueString(), r.rtype.String(), data.Name.ValueString()))
		if err != nil {
			retryCounter++
			if retryCounter <= retryLimit {
				resp.Diagnostics.AddWarning("Client Warning", fmt.Sprintf("Failed to save zone, going to fully retry: %s", err.Error()))
				time.Sleep(time.Duration(retryCounter*5) * time.Second)
				err = nil
			}
		}

	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not save zone: %s", err.Error()))
		return
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
	var data *RecordModel
	tflog.Trace(ctx, "- Resource Read")

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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Retreiving zone %s from scope %s resulted in error: %s", data.Zone.ValueString(), data.Scope.ValueString(), err.Error()))
		return
	}

	subdomain, err := zone.FindSubdomain(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read subdomain %s, got error: %s", data.Name.ValueString(), err))
		return
	}

	record, err := subdomain.GetType(r.rtype.String())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read type %s, got error: %s", r.rtype.String(), err))
		return
	}

	resp.Diagnostics.Append(RecordToDataModel(ctx, data, record)...)

	// UpdateYaml updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *RecordModel
	var state *RecordModel
	tflog.Trace(ctx, "- Resource Update")

	r.client.Mutex.Lock()
	defer r.client.Mutex.Unlock()

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform state data into the model so it can be compared against plan
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retry loop
	var zone *models.Zone
	var subdomain models.Subdomain
	var record *models.Record
	var err error

	retryCounter := 0
	retryLimit := 5

	if retryCounter <= retryLimit {

		zone, err = r.client.GetZone(state.Zone.ValueString(), state.Scope.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not retrieve zone: %s", err.Error()))
			return
		}

		subdomain, err = zone.FindSubdomain(state.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find subdomain, got error: %s", err))
			return
		}

		record, err = subdomain.GetType(r.rtype.String())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find type record, got error: %s", err))
			return
		}

		resp.Diagnostics.Append(RecordFromDataModel(ctx, data, record)...)

		err = subdomain.UpdateYaml()
		if err != nil {
			resp.Diagnostics.AddError("Yaml Error", fmt.Sprintf("Unable to update subdomain in yaml, got error: %s", err))
			return
		}

		err = r.client.SaveZone(zone, fmt.Sprintf("chore(%s): update %s record for %s", data.Zone.ValueString(), r.rtype.String(), data.Name.ValueString()))
		if err != nil {
			retryCounter++
			if retryCounter <= retryLimit {
				resp.Diagnostics.AddWarning("Client Warning", fmt.Sprintf("Failed to save zone, going to fully retry: %s", err.Error()))
				time.Sleep(time.Duration(retryCounter*5) * time.Second)
				err = nil
			}
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not save zone: %s", err.Error()))
		return
	}

	// Set ID
	data.Id = types.StringValue(fmt.Sprintf("%s %s %s", data.Scope.ValueString(), data.Zone.ValueString(), data.Name.ValueString()))

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
	var data *RecordModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.client.Mutex.Lock()
	defer r.client.Mutex.Unlock()

	// Retry loop
	var zone *models.Zone
	var subdomain models.Subdomain
	var err error

	retryCounter := 0
	retryLimit := 5

	if retryCounter <= retryLimit {

		zone, err = r.client.GetZone(data.Zone.ValueString(), data.Scope.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not retrieve zone: %s", err.Error()))
			return
		}

		subdomain, err = zone.FindSubdomain(data.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find subdomain, got error: %s", err))
			return
		}

		err = subdomain.DeleteType(r.rtype.String())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find type record, got error: %s", err))
			return
		}

		err = subdomain.FindAllType()
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could refresh all types of subdmain: %s", err.Error()))
		}

		if len(subdomain.Types) == 0 {
			//resp.Diagnostics.AddWarning("Client Warning", fmt.Sprintf("Trying to delete subdomain: %s / %s", subdomain.Name, data.Name))
			err = zone.DeleteSubdomain(subdomain.Name)
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete subdomain, got error: %s", err))
				return
			}
		}

		err = r.client.SaveZone(zone, fmt.Sprintf("chore(%s): delete %s record for %s", data.Zone.ValueString(), r.rtype.String(), data.Name.ValueString()))
		if err != nil {
			retryCounter++
			if retryCounter <= retryLimit {
				resp.Diagnostics.AddWarning("Client Warning", fmt.Sprintf("Failed to save zone , going to fully retry: %s", err.Error()))
				time.Sleep(time.Duration(retryCounter*5) * time.Second)
				err = nil
			}
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Could not save zone: %s", err.Error()))
		return
	}

	// Set ID
	data.Id = types.StringNull()

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
