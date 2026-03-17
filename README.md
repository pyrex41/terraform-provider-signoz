<!-- markdownlint-disable first-line-h1 no-inline-html -->
<a href="https://terraform.io">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".github/terraform_logo_dark.svg">
    <source media="(prefers-color-scheme: light)" srcset=".github/terraform_logo_light.svg">
    <img src=".github/terraform_logo_light.svg" alt="Terraform logo" title="Terraform" align="right" height="50">
  </picture>
</a>

# Terraform SigNoz Provider

The **SigNoz Provider** enables [Terraform](https://terraform.io) to manage [SigNoz](https://signoz.io) resources.

## Compatibility

| Provider Version | SigNoz Version | Notes |
|-----------------|----------------|-------|
| >= 0.0.12 | >= 0.104.0 | Current release; handles array-wrapped API responses, v5 dashboard mutation shielding, notification channels |
| < 0.0.12 | < 0.104.0 | Legacy single-object API response format only |

**Tested against:** SigNoz v0.110.1

### SigNoz v5 Dashboard Migration

SigNoz >= 0.92 uses v5 query format internally. The provider preserves your Terraform plan values to prevent SigNoz's server-side v5 migration from mutating widget JSON (template variable rewriting, orderBy injection, operator changes). No special configuration is needed — this is handled automatically.

## Resources

| Resource | Description |
|----------|-------------|
| `signoz_dashboard` | Manage SigNoz dashboards (v4 and v5 query formats) |
| `signoz_alert` | Manage SigNoz alert rules |
| `signoz_notification_channel` | Manage notification channels (Slack, webhook) |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `signoz_dashboard` | Read existing dashboards |
| `signoz_alert` | Read existing alert rules |
| `signoz_notification_channel` | Read existing notification channels |

## Configuration

```hcl
provider "signoz" {
  endpoint     = "http://localhost:3301"   # or SIGNOZ_ENDPOINT env var
  access_token = "your-api-key"            # or SIGNOZ_ACCESS_TOKEN env var
}
```

| Attribute | Environment Variable | Default | Description |
|-----------|---------------------|---------|-------------|
| `endpoint` | `SIGNOZ_ENDPOINT` | `http://localhost:3301` | SigNoz API URL |
| `access_token` | `SIGNOZ_ACCESS_TOKEN` | — | API key or bearer token |
| `http_timeout` | `SIGNOZ_HTTP_TIMEOUT` | `35` | HTTP timeout in seconds |
| `http_max_retry` | `SIGNOZ_HTTP_MAX_RETRY` | `10` | Max retry count |

## Known Issues

- **UUIDv7 collisions:** SigNoz uses timestamp-based UUIDs. Creating multiple dashboards within the same millisecond can produce duplicate IDs. The provider detects this and retries automatically.
- **POST response may return wrong dashboard:** The provider validates by name and falls back to a list query if needed.
- **HTML escaping in JSON:** Go's default JSON marshaling escapes `<`, `>`, `&`, causing drift on ClickHouse queries. The provider uses custom marshaling to avoid this.
- **Empty notification channel update responses:** SigNoz may return empty bodies on successful PATCH; the provider handles this gracefully.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Documentation

To update the documentation edit the files in templates/ and then run make docs.
The files in docs/ are auto-generated and should not be updated manually.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org)
installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the
provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.
