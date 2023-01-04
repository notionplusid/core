output "service_account_email" {
  value = data.google_app_engine_default_service_account.service_account.email
}