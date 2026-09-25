## UNRELEASED (TBA)

BEHAVIOUR CHANGES:
- Fix: records that exist in the Terraform state but no longer exist in the zone file (removed outside of Terraform, including from an empty zone `{}`) no longer fail `plan`/`apply` with a `Client Error`. They are removed from the state with a `Record not found` warning and recreated on the next apply, as recommended by the Terraform Plugin Framework.
- Fix: destroying a record that no longer exists in the zone file now succeeds with a `Record already removed` warning instead of failing.
- If you depend on the previous behaviour, set `error_on_missing_records = true` in the provider configuration. This option is deprecated on introduction and will be removed in version 2.0.0.

FEATURES:
- New resource and data source `octodns_alias_record` for ALIAS records
- Apply-time warning when a record is created next to record types it cannot coexist with according to octodns (CNAME next to any other type, ALIAS next to A or AAAA). All conflicting types are reported in a single warning. The record is still written.

NOTES:
- The ALIAS and CNAME resource documentation now describes the octodns coexistence rules and the current provider behaviour
- Building the provider from source now requires Go 1.26 or newer (required by `golang.org/x/oauth2` v0.37.0)

FIXES:
- Plan-time validation of record names: CNAME records at the zone root and ALIAS records outside the zone root are now rejected, matching the octodns validators
- Provider configuration stops at the first configuration error instead of building a client from invalid settings
- Errors from setting branch, author and the default scope are reported instead of discarded
- Duplicate scope detection now treats an unnamed scope as `default`: two unnamed scopes, or an unnamed scope plus a scope named `default`, are rejected instead of silently overwriting each other
- Creating a record in an empty zone file (`{}`) no longer fails, and the zone is written in block style instead of flow style

SECURITY:
- Bump `google.golang.org/grpc` to v1.83.2 (GO-2026-6443) and `golang.org/x/text` to v0.42.0 (GO-2026-5970); both vulnerabilities were reachable from the provider according to `govulncheck`
- Bump `golang.org/x/crypto` to v0.57.0 and `golang.org/x/net` to v0.59.0, resolving the open Dependabot security alerts (not reachable from the provider)
- `google.golang.org/grpc` is intentionally kept on the 1.83 line: v1.84.0 is still affected by GO-2026-6443

DEPRECATIONS:
- Provider attribute `error_on_missing_records`: backwards compatibility option for the missing record behaviour of versions up to 1.2.0, will be removed in version 2.0.0

INTERNAL:
- Dependency updates: `terraform-plugin-docs` v0.25.0, `terraform-plugin-log` v0.11.0, `golang.org/x/oauth2` v0.37.0
- CI: `actions/checkout` v7, `actions/setup-go` v7, `hashicorp/setup-terraform` v4, `golangci-lint-action` v9.3.0, `goreleaser-action` v7.2.3, `ghaction-import-gpg` v7.0.0
- Added unit tests for record name validation, record type conflict detection, missing record handling, empty zones and duplicate scopes
- ALIAS test data moved to a separate fixture (`testdata/unit.tests.alias.yaml`), so the octodns based fixtures stay valid

## 1.2.0 (2026-04-20)

FEATURES:
- Plan-time validation: invalid record values (e.g. malformed IPs) are now reported during `terraform plan` instead of at `apply`

CHANGES:
- Multiple record changes in a single `terraform apply` are batched into one GitHub commit per zone
- Commit messages include the scope: `chore(scope/zone): ...`
- Batched commits use a summary: `N changes (X creates, Y updates, Z deletes)` instead of concatenated individual messages

FIXES:
- Roll back in-memory zone state on error so partially-modified records no longer leak into GitHub
- Zone cache is only populated after successful YAML parsing, preventing stale corrupted zones from persisting across operations
- Propagate value validation errors from `RecordFromDataModel` instead of silently discarding invalid values

INTERNAL:
- Added unit tests for batching, rollback paths, and plan-time validation (coverage 67%→77%)
- Replaced `log.Printf` with `tflog`
- Removed dead code

## 1.1.4 (2025-11-14)

FIXES:
- Fix `illegal base64 data at input byte 0` error

## 1.1.3 (2025-11-13)

CHANGES:
- Add retry mode to fix 409 errors when applying multiple changes
- Dependencies upgrade

## 0.1.2 (2024-07-08)

FIXES:
- apply results in `Provider returned invalid result object after apply` error

CHANGES:
- Add documentation about `import`

## 0.1.1 (2024-07-04)

FIXES:
- apply results in `Provider returned invalid result object after apply` error

## 0.1.0 (2024-02-26) 

FEATURES:
