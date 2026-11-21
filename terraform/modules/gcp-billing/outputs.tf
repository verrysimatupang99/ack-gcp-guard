output "billing_dataset_id" {
  description = "BigQuery dataset ID that will hold the billing export"
  value       = google_bigquery_dataset.billing.dataset_id
}

output "billing_reader_sa_email" {
  description = "Email of the billing reader service account"
  value       = google_service_account.billing_reader.email
}

output "billing_reader_sa_key" {
  description = "JSON key for the billing reader SA (sensitive)"
  value       = google_service_account_key.billing_reader.private_key
  sensitive   = true
}
