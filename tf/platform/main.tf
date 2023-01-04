terraform {
  backend "gcs" {}
}

module "config" {
  source = "../config"
}

locals {
  project = module.config.env.project
}

provider "google" {
  project = local.project
}

module "appengine" {
  project = local.project
  source  = "../module/appengine"
}

module "secrets" {
  project = local.project
  source = "../module/secrets"

  notion_client_id = var.notion_client_id
  notion_client_secret = var.notion_client_secret
  service_account = module.appengine.service_account_email
}
