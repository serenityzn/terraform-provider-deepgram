terraform {
  required_providers {
    deepgram = {
      source  = "serenityzn/deepgram"
      version = "~> 0.1"
    }
  }
}

provider "deepgram" {
  # api_key can also be provided via DEEPGRAM_API_KEY environment variable
  api_key = var.deepgram_api_key
}

variable "deepgram_api_key" {
  description = "Deepgram API key"
  type        = string
  sensitive   = true
}

variable "project_id" {
  description = "Deepgram project ID"
  type        = string
}

# -------------------------------------------------------------------
# Resource: create an API key
# -------------------------------------------------------------------
resource "deepgram_key" "example" {
  project_id = var.project_id
  comment    = "Managed by Terraform"
  scopes     = toset(["usage:read", "keys:write"])
  tags       = ["terraform", "example"]

  # Optional – remove to create a non-expiring key
  expiration_date = "2027-01-01T00:00:00Z"
}

output "api_key_id" {
  description = "The ID of the newly created key"
  value       = deepgram_key.example.id
}

output "api_key_secret" {
  description = "The secret key value (only available after creation)"
  value       = deepgram_key.example.key
  sensitive   = true
}

# -------------------------------------------------------------------
# Data source: list all keys for the project
# -------------------------------------------------------------------
data "deepgram_keys" "all" {
  project_id = var.project_id
}

output "all_key_ids" {
  description = "IDs of every API key in the project"
  value       = [for k in data.deepgram_keys.all.api_keys : k.api_key_id]
}
