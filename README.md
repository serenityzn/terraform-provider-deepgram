# Terraform Provider – Deepgram

[![Terraform Registry](https://img.shields.io/badge/Terraform%20Registry-serenityzn%2Fdeepgram-blue)](https://registry.terraform.io/providers/serenityzn/deepgram/latest)
[![License: MPL-2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

A Terraform provider for managing [Deepgram](https://deepgram.com) project API keys.

Deepgram powers the Voice AI Economy — this provider lets you manage your Deepgram project API keys as infrastructure-as-code.

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
  api_key = var.deepgram_api_key  # or set DEEPGRAM_API_KEY env var
}
```

Run:

```bash
terraform init
terraform apply \
  -var="deepgram_api_key=your-key-here" \
  -var="project_id=your-project-id"
```

Or use environment variables to avoid being prompted:

```bash
export TF_VAR_deepgram_api_key="your-key-here"
export TF_VAR_project_id="your-project-id"
terraform apply
```

## Provider configuration

| Attribute  | Type   | Required | Description |
|------------|--------|----------|-------------|
| `api_key`  | string | yes      | Deepgram API key. Can also be set via `DEEPGRAM_API_KEY` env var |
| `base_url` | string | no       | Override API base URL. Defaults to `https://api.deepgram.com` |

## Resources

### `deepgram_key`

Creates an API key inside a Deepgram project.

> **Note:** Deepgram API keys are immutable after creation. Any change to
> `comment`, `scopes`, `tags`, or `expiration_date` will destroy and
> re-create the key. The secret key value (`key`) is only available
> immediately after creation — retrieve it with:
> ```bash
> terraform output -raw api_key_secret
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

#### Argument reference

| Attribute        | Type         | Required | Description |
|------------------|--------------|----------|-------------|
| `project_id`     | string       | yes      | Deepgram project ID |
| `comment`        | string       | yes      | Human-readable label for the key |
| `scopes`         | set(string)  | yes      | Permission scopes (e.g. `toset(["usage:read", "keys:write"])`) |
| `tags`           | list(string) | no       | Optional tags |
| `expiration_date`| string       | no       | RFC 3339 expiry (e.g. `"2027-01-01T00:00:00Z"`). Omit for a non-expiring key |

#### Computed attributes

| Attribute | Description |
|-----------|-------------|
| `id`      | The `api_key_id` assigned by Deepgram |
| `key`     | The secret API key value (**sensitive** — only available at creation time) |

## Data sources

### `deepgram_keys`

Lists all API keys for a project.

#### Example

```hcl
data "deepgram_keys" "all" {
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}

output "key_ids" {
  value = [for k in data.deepgram_keys.all.api_keys : k.api_key_id]
}
```

#### Argument reference

| Attribute    | Type   | Required | Description |
|--------------|--------|----------|-------------|
| `project_id` | string | yes      | Deepgram project ID |

#### Computed attributes

Each object in `api_keys` contains:

| Attribute    | Description |
|--------------|-------------|
| `api_key_id` | Unique key identifier |
| `comment`    | Human-readable label |
| `scopes`     | Set of permission scopes |
| `created`    | Creation timestamp |
| `member_id`  | ID of the member who owns the key |
| `email`      | Email of the member who owns the key |

## Requirements

| Tool      | Version |
|-----------|---------|
| Go        | ≥ 1.21  |
| Terraform | ≥ 1.5   |

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

# Apply without terraform init (dev overrides skip the registry)
terraform apply -var="deepgram_api_key=..." -var="project_id=..."
```

## License

[Mozilla Public License 2.0](LICENSE)
