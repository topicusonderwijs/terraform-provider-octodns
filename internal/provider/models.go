package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

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
	Zone   types.String   `tfsdk:"zone"`
	Scope  types.String   `tfsdk:"scope"`
	Name   types.String   `tfsdk:"name"`
	Id     types.String   `tfsdk:"id"`
	Values []types.String `tfsdk:"values"`
	TTL    types.Int64    `tfsdk:"ttl"`
}
