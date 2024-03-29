---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "octodns_a_record Data Source - terraform-provider-octodns"
subcategory: ""
description: |-
  A record data source
---

# octodns_a_record (Data Source)

A record data source

## Example Usage

```terraform
data "octodns_a_record" "root" {
  zone = "unit.tests"
  name = "@"
}
output "a_record" {
  value = data.octodns_a_record.root
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Record Name
- `zone` (String) Zone of the record

### Optional

- `scope` (String) Scope of zone

### Read-Only

- `id` (String) Record identifier
- `octodns` (Attributes) Additional provider specific record meta config. (see [below for nested schema](#nestedatt--octodns))
- `ttl` (Number) TTL of the record, if not set the zone's or dns server setting is used
- `values` (List of String) Values of the record, should confirm to record type

<a id="nestedatt--octodns"></a>
### Nested Schema for `octodns`

Read-Only:

- `azuredns` (Attributes) Healthcheck configuration for [Azure provider](https://github.com/octodns/octodns-azure/?tab=readme-ov-file#healthchecks) (see [below for nested schema](#nestedatt--octodns--azuredns))
- `cloudflare` (Attributes) Meta config for [cloudflare provider](https://github.com/octodns/octodns-cloudflare/?tab=readme-ov-file#configuration) (see [below for nested schema](#nestedatt--octodns--cloudflare))

<a id="nestedatt--octodns--azuredns"></a>
### Nested Schema for `octodns.azuredns`

Read-Only:

- `hc_interval` (Number) Azure healthcheck interval
- `hc_numfailures` (Number) Azure healthcheck number of failures allowed
- `hc_timeout` (Number) Azure healthcheck timeout


<a id="nestedatt--octodns--cloudflare"></a>
### Nested Schema for `octodns.cloudflare`

Read-Only:

- `auto_ttl` (Boolean) Use cloudflare's auto-ttl *feature*, aka: set to 300
- `proxied` (Boolean) Should cloudflare proxy this record (only for A/AAAA records)
