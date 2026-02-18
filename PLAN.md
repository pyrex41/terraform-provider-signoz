# Fix SigNoz Dashboard/Alert API Response Parsing

## Problem

The terraform-provider-signoz (v0.0.11) fails against SigNoz chart >= 0.104.0:

- **Dashboards**: `json: cannot unmarshal array into Go struct field dashboardResponse.data of type client.dashboardData`
- **Alerts**: `cannot find id in tfstate` (Create response parsing fails silently, ID never stored)

## Root Cause

1. **Dashboard**: SigNoz API now returns `data` as an array (`[{...}]`) instead of a single object (`{...}`). The provider's `dashboardResponse.Data` is typed as `dashboardData` (struct), which can't unmarshal a JSON array.

2. **Alert**: The response parsing is fragile — any unmarshal failure causes the ID to be lost. The upstream `GettableRule.MarshalJSON()` also strips `evaluation`, `schemaVersion`, and `notificationSettings` for v1 schema rules, causing Terraform plan/apply inconsistencies (issue #75).

## Fix (already implemented locally)

- Changed `dashboardResponse.Data` and `alertResponse.Data` from concrete types to `json.RawMessage`
- Added `parseDashboardData()` and `parseAlertData()` that handle both object and array formats
- Added explicit ID validation on Create responses
- Added better error messages with truncated response bodies
- Added 12 unit tests covering all parsing scenarios

## Execution Plan

### Step 1: Fork the repo
```
gh repo fork SigNoz/terraform-provider-signoz --clone=false
git remote add fork git@github.com:pyrex41/terraform-provider-signoz.git
```

### Step 2: Create branch, push changes
```
git checkout -b fix/api-response-parsing
git add -A
git commit -m "fix: handle array-wrapped API responses for dashboards and alerts"
git push -u fork fix/api-response-parsing
```

### Step 3: Tag a pre-release on the fork
```
git tag v0.0.12-rc1
git push fork v0.0.12-rc1
```

### Step 4: Open PR upstream to SigNoz/terraform-provider-signoz
Target: `main`
Title: `fix: handle array-wrapped API responses from SigNoz >= 0.104.0`

### Step 5: Update upjet provider
In `github.com/pyrex41/provider-signoz`:
- Update `go.mod` to point at the fork tag
- Regenerate and rebuild: `ghcr.io/pyrex41/provider-signoz:v0.2.0`
- Deploy and validate against SigNoz 0.104.0 cluster

## Files Changed

| File | Change |
|------|--------|
| `signoz/internal/client/types.go` | `Data` fields → `json.RawMessage`, added parse helpers |
| `signoz/internal/client/dashboard.go` | Use `parseDashboardData()`, better errors |
| `signoz/internal/client/alert.go` | Use `parseAlertData()`, ID validation, better errors |
| `signoz/internal/client/types_test.go` | **New** — 12 unit tests |

## Validation

- `go build ./...` — passes
- `go vet ./...` — passes
- `go test ./signoz/internal/client/ -v` — 12/12 tests pass
