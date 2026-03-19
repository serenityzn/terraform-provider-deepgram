# Terraform Provider – Deepgram

[![Terraform Registry](https://img.shields.io/badge/Terraform%20Registry-serenityzn%2Fdeepgram-blue)](https://registry.terraform.io/providers/serenityzn/deepgram/latest)
[![License: MPL-2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

A Terraform provider for managing [Deepgram](https://deepgram.com) resources as infrastructure-as-code.

Deepgram powers the Voice AI Economy — use this provider to manage your Deepgram projects, API keys, and members declaratively.

## Usage

```hcl
terraform {
  required_providers {
    deepgram = {
      source  = "serenityzn/deepgram"
      version = "~> 0.1"
    }
  }
}

provider "deepgram" {
  # api_key can also be set via DEEPGRAM_API_KEY environment variable
  api_key = var.deepgram_api_key
}
```

```bash
terraform init
export DEEPGRAM_API_KEY="your-api-key"
terraform apply
```

## Provider configuration

| Attribute  | Type   | Required | Description |
|------------|--------|----------|-------------|
| `api_key`  | string | yes      | Deepgram API key. Can also be set via `DEEPGRAM_API_KEY` env var |
| `base_url` | string | no       | Override API base URL. Defaults to `https://api.deepgram.com` |

The API key must have `keys:read` and `keys:write` scopes to manage keys.

---

## Resources

### `deepgram_key`

Creates and manages an API key inside a Deepgram project.

> **Note:** Deepgram API keys are immutable after creation. Any change to
> `comment`, `scopes`, `tags`, or `expiration_date` will destroy and re-create
> the key. The secret `key` value is only available immediately after creation:
> ```bash
> terraform output -raw <output_name>
> ```

#### Example

```hcl
resource "deepgram_key" "ci" {
  project_id      = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  comment         = "CI pipeline key"
  scopes          = toset(["usage:read", "usage:write"])
  tags            = ["ci", "terraform"]
  expiration_date = "2027-01-01T00:00:00Z"
}

output "ci_key_secret" {
  value     = deepgram_key.ci.key
  sensitive = true
}
```

#### Arguments

| Attribute         | Type         | Required | Description |
|-------------------|--------------|----------|-------------|
| `project_id`      | string       | yes      | Deepgram project ID |
| `comment`         | string       | yes      | Human-readable label for the key |
| `scopes`          | set(string)  | yes      | Permission scopes (e.g. `toset(["usage:read", "keys:write"])`) |
| `tags`            | list(string) | no       | Optional tags |
| `expiration_date` | string       | no       | RFC 3339 expiry (e.g. `"2027-01-01T00:00:00Z"`). Omit for non-expiring |

#### Computed attributes

| Attribute | Description |
|-----------|-------------|
| `id`      | The `api_key_id` assigned by Deepgram |
| `key`     | The secret API key value (**sensitive** — only available at creation time) |

#### Import

```bash
terraform import deepgram_key.example <project_id>/<key_id>
```

> The `key` secret cannot be recovered on import and will be empty in state.

---

## Data sources

### `deepgram_keys`

Lists all API keys for a project.

```hcl
data "deepgram_keys" "all" {
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

output "key_ids" {
  value = [for k in data.deepgram_keys.all.api_keys : k.api_key_id]
}
```

#### Arguments

| Attribute    | Type   | Required | Description |
|--------------|--------|----------|-------------|
| `project_id` | string | yes      | Deepgram project ID |

#### Computed attributes (each item in `api_keys`)

| Attribute    | Description |
|--------------|-------------|
| `api_key_id` | Unique key identifier |
| `comment`    | Human-readable label |
| `scopes`     | Set of permission scopes |
| `created`    | Creation timestamp |
| `member_id`  | ID of the member who owns the key |
| `email`      | Email of the member who owns the key |

---

### `deepgram_projects`

Lists all Deepgram projects accessible with the API key.

```hcl
data "deepgram_projects" "all" {}

output "project_ids" {
  value = [for p in data.deepgram_projects.all.projects : p.project_id]
}
```

#### Computed attributes (each item in `projects`)

| Attribute     | Description |
|---------------|-------------|
| `project_id`  | Unique project identifier |
| `name`        | Project name |
| `mip_opt_out` | Whether opted out of the Model Improvement Program |

---

### `deepgram_project`

Fetches a single project by name.

```hcl
data "deepgram_project" "main" {
  name = "My Production Project"
}

resource "deepgram_key" "app" {
  project_id = data.deepgram_project.main.project_id
  comment    = "App key"
  scopes     = toset(["usage:read"])
}
```

Returns an error if no project matches, or if multiple projects share the same name (use `deepgram_projects` to list all and select by `project_id` in that case).

#### Arguments

| Attribute | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `name`    | string | yes      | Exact project name to look up |

#### Computed attributes

| Attribute     | Description |
|---------------|-------------|
| `project_id`  | Unique project identifier |
| `mip_opt_out` | Whether opted out of the Model Improvement Program |

---

## Requirements

| Tool      | Version |
|-----------|---------|
| Go        | ≥ 1.21  |
| Terraform | ≥ 1.5   |

## Testing

**Unit tests** (no credentials needed):

```bash
go test ./internal/deepgram/... -v
```

**Acceptance tests** (run against real Deepgram API):

```bash
export DEEPGRAM_API_KEY="your-api-key"
export DEEPGRAM_PROJECT_ID="your-project-id"
TF_ACC=1 go test ./internal/provider/... -v -timeout 5m
```

## Local development

```bash
# Build and install locally
go mod tidy
go install .

# Add dev override to ~/.terraformrc
cat > ~/.terraformrc << 'EOF'
provider_installation {
  dev_overrides {
    "serenityzn/deepgram" = "/path/to/your/gopath/bin"
  }
  direct {}
}
EOF

# Skip terraform init and apply directly (dev overrides bypass the registry)
export DEEPGRAM_API_KEY="your-api-key"
terraform apply
```

## License

[Mozilla Public License 2.0](LICENSE)
