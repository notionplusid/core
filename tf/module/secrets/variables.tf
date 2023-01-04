variable "location" {
  type        = string
  description = "(optional) App Engine location"
  # as closest as possible to the Notion datacentre.
  default = "us-west2"
}

variable "project" {
  type        = string
  description = "(required) App Engine GCloud project"
}

variable "notion_client_id" {
  type        = string
  description = "(optional) Notion Extension Client ID"
  default     = "internal"
  sensitive   = true
}

variable "notion_client_secret" {
  type        = string
  description = "(required) Notion Extension Client Secret"
  sensitive   = true
}

variable "service_account" {
  type        = string
  description = "(required) Service account email that would be accessing the secrets"
}

