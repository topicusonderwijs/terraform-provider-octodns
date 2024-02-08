package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Zone    types.String        `tfsdk:"zone"`
	Scope   types.String        `tfsdk:"scope"`
	Name    types.String        `tfsdk:"name"`
	Id      types.String        `tfsdk:"id"`
	Values  []types.String      `tfsdk:"values"`
	TTL     types.Int64         `tfsdk:"ttl"`
	Octodns *OctodnsConfigModel `tfsdk:"octodns"`
}

type OctodnsConfigModel struct {
	Cloudflare *OctodnsCloudflareModel `tfsdk:"cloudflare"`
	AzureDNS   *OctodnsAzureDNSModel   `tfsdk:"azuredns"`
}

func (o OctodnsConfigModel) HasConfig() bool {
	if o.Cloudflare != nil || o.AzureDNS != nil {
		return true
	} else {
		return false
	}
}

type OctodnsCloudflareModel struct {
	Proxied types.Bool `tfsdk:"proxied"`
	AutoTTL types.Bool `tfsdk:"auto_ttl"`
}

type OctodnsAzureDNSModel struct {
	HCInterval    types.Int64 `tfsdk:"hc_interval"`
	HCTimeout     types.Int64 `tfsdk:"hc_timeout"`
	HCNumFailures types.Int64 `tfsdk:"hc_numfailures"`
}
