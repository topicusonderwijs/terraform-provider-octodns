## UNRELEASED (TBA)

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
