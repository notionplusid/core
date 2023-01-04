resource "google_app_engine_application" "app" {
  project       = var.project
  location_id   = var.location
  database_type = "CLOUD_DATASTORE_COMPATIBILITY"
}

data "google_app_engine_default_service_account" "service_account" {
  depends_on = [
    google_app_engine_application.app
  ]
}