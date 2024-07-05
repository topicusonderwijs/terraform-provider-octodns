package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/topicusonderwijs/terraform-provider-octodns/internal/models"
)

// SubdomainModel describes the data source data model.
type SubdomainModel struct {
	Zone  types.String `tfsdk:"zone"`
	Scope types.String `tfsdk:"scope"`
	Name  types.String `tfsdk:"name"`
	Id    types.String `tfsdk:"id"`
	Type  []TypeModel  `tfsdk:"type"`
}
type TypeModel struct {
	Type   types.String   `tfsdk:"type"`
	Values []types.String `tfsdk:"values"`
	TTL    types.Int64    `tfsdk:"ttl"`
}

type RecordModel struct {
	Zone    types.String   `tfsdk:"zone"`
	Scope   types.String   `tfsdk:"scope"`
	Name    types.String   `tfsdk:"name"`
	Id      types.String   `tfsdk:"id"`
	Values  []types.String `tfsdk:"values"`
	TTL     types.Int64    `tfsdk:"ttl"`
	Octodns types.Object   `tfsdk:"octodns"`
}

type OctodnsConfigModel struct {
	Cloudflare types.Object `tfsdk:"cloudflare"`
	AzureDNS   types.Object `tfsdk:"azuredns"`
}

/*
	func (o OctodnsConfigModel) HasConfig() bool {
		if o.Cloudflare != nil || o.AzureDNS != nil {
			return true
		} else {
			return false
		}
	}
*/
type OctodnsCloudflareModel struct {
	Proxied types.Bool `tfsdk:"proxied"`
	AutoTTL types.Bool `tfsdk:"auto_ttl"`
}

func (o OctodnsCloudflareModel) Attributes() (attributes map[string]attr.Type) {

	attributes = make(map[string]attr.Type)

	attributes["proxied"] = types.BoolType
	attributes["auto_ttl"] = types.BoolType

	return attributes
}

type OctodnsAzureDNSModel struct {
	HCInterval    types.Int64 `tfsdk:"hc_interval"`
	HCTimeout     types.Int64 `tfsdk:"hc_timeout"`
	HCNumFailures types.Int64 `tfsdk:"hc_numfailures"`
}

func (o OctodnsAzureDNSModel) Attributes() (attributes map[string]attr.Type) {

	attributes = make(map[string]attr.Type)

	attributes["hc_interval"] = types.Int64Type
	attributes["hc_timeout"] = types.Int64Type
	attributes["hc_numfailures"] = types.Int64Type

	return attributes
}

func RecordToDataModel(ctx context.Context, data *RecordModel, record *models.Record) diag.Diagnostics {

	retDiags := diag.Diagnostics{}
	var diags diag.Diagnostics

	if record.TTL > 0 {
		data.TTL = types.Int64Value(int64(record.TTL))
	} else {
		data.TTL = types.Int64Null()
	}

	data.Values = []types.String{}
	for _, v := range record.ValuesAsString() {
		data.Values = append(data.Values, types.StringValue(v))
	}

	octodnsTFObj := make(map[string]attr.Value)

	Cloudflare := OctodnsCloudflareModel{}
	AzureDNS := OctodnsAzureDNSModel{}
	octodnsTFObj["cloudflare"] = types.ObjectNull(Cloudflare.Attributes())
	octodnsTFObj["azuredns"] = types.ObjectNull(AzureDNS.Attributes())

	if record.Octodns.Cloudflare != nil {
		if record.Octodns.Cloudflare.Proxied {
			Cloudflare.Proxied = types.BoolValue(true)
		}
		if record.Octodns.Cloudflare.AutoTTL {
			Cloudflare.AutoTTL = types.BoolValue(true)
		}

		octodnsTFObj["cloudflare"], diags = types.ObjectValueFrom(ctx, Cloudflare.Attributes(), Cloudflare)
		retDiags.Append(diags...)
	}

	if record.Octodns.AzureDNS != nil {
		isSet := false

		if record.Octodns.AzureDNS.Healthcheck.Interval > 0 {
			AzureDNS.HCInterval = types.Int64Value(int64(record.Octodns.AzureDNS.Healthcheck.Interval))
			isSet = true
		}
		if record.Octodns.AzureDNS.Healthcheck.Timeout > 0 {
			AzureDNS.HCTimeout = types.Int64Value(int64(record.Octodns.AzureDNS.Healthcheck.Timeout))
			isSet = true
		}
		if record.Octodns.AzureDNS.Healthcheck.NumFailures > 0 {
			AzureDNS.HCNumFailures = types.Int64Value(int64(record.Octodns.AzureDNS.Healthcheck.NumFailures))
			isSet = true
		}
		if isSet {
			octodnsTFObj["azuredns"], diags = types.ObjectValueFrom(ctx, AzureDNS.Attributes(), AzureDNS)
			retDiags.Append(diags...)
		}
	}

	if octodnsTFObj["cloudflare"].IsNull() && octodnsTFObj["azuredns"].IsNull() {
		data.Octodns = types.ObjectNull(data.Octodns.AttributeTypes(ctx))
	} else {
		data.Octodns, diags = types.ObjectValue(data.Octodns.AttributeTypes(ctx), octodnsTFObj)
		retDiags.Append(diags...)
	}

	return retDiags

}

func RecordFromDataModel(ctx context.Context, data *RecordModel, record *models.Record) (diags diag.Diagnostics) {

	record.Name = data.Name.ValueString()
	record.TTL = int(data.TTL.ValueInt64())

	record.ClearValues()
	for _, v := range data.Values {
		_ = record.AddValueFromString(v.ValueString())
	}

	record.Octodns = models.OctodnsRecordConfig{}

	if !data.Octodns.IsUnknown() && !data.Octodns.IsNull() {

		var octodns OctodnsConfigModel
		diags.Append(data.Octodns.As(ctx, &octodns, basetypes.ObjectAsOptions{})...)

		if !octodns.Cloudflare.IsUnknown() && !octodns.Cloudflare.IsNull() {
			var dCF OctodnsCloudflareModel
			diags.Append(octodns.Cloudflare.As(ctx, &dCF, basetypes.ObjectAsOptions{})...)

			oCF := models.OctodnsCloudflare{}
			if !dCF.Proxied.IsNull() {
				oCF.Proxied = dCF.Proxied.ValueBool()
			}
			if !dCF.AutoTTL.IsNull() {
				oCF.AutoTTL = dCF.AutoTTL.ValueBool()
			}

			record.Octodns.Cloudflare = &oCF

		}

		if !octodns.AzureDNS.IsUnknown() && !octodns.AzureDNS.IsNull() {
			var dAZ OctodnsAzureDNSModel
			diags.Append(octodns.AzureDNS.As(ctx, &dAZ, basetypes.ObjectAsOptions{})...)

			oAZ := models.OctodnsAzureDNS{Healthcheck: models.OctodnsAzureDNSHealthcheck{}}

			if !dAZ.HCTimeout.IsNull() {
				oAZ.Healthcheck.Timeout = int(dAZ.HCTimeout.ValueInt64())
			}
			if !dAZ.HCInterval.IsNull() {
				oAZ.Healthcheck.Interval = int(dAZ.HCInterval.ValueInt64())
			}
			if !dAZ.HCNumFailures.IsNull() {
				oAZ.Healthcheck.NumFailures = int(dAZ.HCNumFailures.ValueInt64())
			}

			record.Octodns.AzureDNS = &oAZ
		}

	}

	return
}
