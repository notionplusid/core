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
