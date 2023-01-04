resource "google_secret_manager_secret" "notion_client_id" {
  secret_id = "NOTION_CLIENT_ID"

  replication {
    automatic = true
  }
}

resource "google_secret_manager_secret_version" "notion_client_id" {
  secret = google_secret_manager_secret.notion_client_id.id

  secret_data = var.notion_client_id
}

resource "google_secret_manager_secret" "notion_client_secret" {
  secret_id = "NOTION_CLIENT_SECRET"

  replication {
    automatic = true
  }
}

resource "google_secret_manager_secret_version" "notion_client_secret" {
  secret = google_secret_manager_secret.notion_client_secret.id

  secret_data = var.notion_client_secret
}

resource "google_secret_manager_secret_iam_member" "notion_client_id" {
  secret_id = google_secret_manager_secret.notion_client_id.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.service_account}"
}

resource "google_secret_manager_secret_iam_member" "notion_client_secret" {
  secret_id = google_secret_manager_secret.notion_client_secret.secret_id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${var.service_account}"
}

