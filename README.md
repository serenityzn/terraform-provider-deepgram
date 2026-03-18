# Terraform Provider – Deepgram

A Terraform provider for managing [Deepgram](https://deepgram.com) project API keys.

## Requirements

| Tool      | Version  |
|-----------|----------|
| Go        | ≥ 1.21   |
| Terraform | ≥ 1.5    |

## Building the provider

```bash
go mod tidy
go build -o terraform-provider-deepgram .
```

## Local development install

```bash
# build and install into your local mirror
go install .
```

Add a `~/.terraformrc` dev override so Terraform picks up the local binary:

```hcl
provider_installation {
  dev_overrides {
    "volodymyrlapada/deepgram" = "<path to your GOPATH/bin>"
  }
  direct {}
}
```

## Provider configuration

```hcl
provider "deepgram" {
  api_key = "your-deepgram-api-key"   # or set DEEPGRAM_API_KEY env var
  # base_url = "https://api.deepgram.com"  # optional override
}
```

### Environment variables

| Variable          | Description                      |
|-------------------|----------------------------------|
| `DEEPGRAM_API_KEY` | Deepgram API key (auth token)   |

## Resources

### `deepgram_key`

Creates an API key inside a Deepgram project.

> **Note:** Deepgram API keys are immutable after creation. Any change to
> `comment`, `scopes`, `tags`, or `expiration_date` will destroy and
> re-create the key. The secret key value (`key`) is only available
> immediately after creation.

#### Example

```hcl
resource "deepgram_key" "ci" {
  project_id      = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  comment         = "CI pipeline key"
  scopes          = ["usage:read", "usage:write"]
  tags            = ["ci", "terraform"]
  expiration_date = "2027-01-01T00:00:00Z"
}

output "ci_key_secret" {
  value     = deepgram_key.ci.key
  sensitive = true
}
```

#### Argument reference

| Attribute        | Type           | Required | Description                                               |
|------------------|----------------|----------|-----------------------------------------------------------|
| `project_id`     | string         | yes      | Deepgram project ID                                       |
| `comment`        | string         | yes      | Human-readable label for the key                         |
| `scopes`         | list(string)   | yes      | Permission scopes (e.g. `["usage:read"]`)                 |
| `tags`           | list(string)   | no       | Optional tags                                             |
| `expiration_date`| string         | no       | RFC 3339 expiry (e.g. `"2027-01-01T00:00:00Z"`)          |

#### Attribute reference (computed)

| Attribute    | Description                                      |
|--------------|--------------------------------------------------|
| `id`         | The `api_key_id` assigned by Deepgram            |
| `key`        | The secret API key value (**sensitive**)         |

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

| Attribute    | Type   | Required | Description           |
|--------------|--------|----------|-----------------------|
| `project_id` | string | yes      | Deepgram project ID   |

#### Attribute reference (computed)

Each object in `api_keys` contains:

| Attribute    | Description                                   |
|--------------|-----------------------------------------------|
| `api_key_id` | Unique key identifier                         |
| `comment`    | Human-readable label                          |
| `scopes`     | List of permission scopes                     |
| `created`    | Creation timestamp                            |
| `member_id`  | ID of the member who owns the key             |
| `email`      | Email of the member who owns the key          |
