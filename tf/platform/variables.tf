variable "notion_client_id" {
  type = string
  sensitive = true
  description = "(optional) Notion Client ID"
  default = "internal"
}

variable "notion_client_secret" {
  type = string
  sensitive = true
  description = "(required) Notion Client Secret"
}
