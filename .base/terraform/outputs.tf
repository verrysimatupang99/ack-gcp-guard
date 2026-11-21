output "ram_scanner_user" {
  description = "RAM user for the Alibaba key scanner"
  value       = module.alibaba_ram.ram_user_name
}

output "ram_scanner_access_key_id" {
  description = "Access key ID for the Alibaba scanner"
  value       = module.alibaba_ram.access_key_id
}

output "ram_scanner_sts_role_arn" {
  description = "STS role ARN the scanner can assume"
  value       = module.alibaba_ram.sts_role_arn
}

output "billing_reader_sa" {
  description = "Service account email that reads the billing export"
  value       = module.gcp_billing.billing_reader_sa_email
}

output "billing_dataset" {
  description = "BigQuery dataset that holds the billing export"
  value       = module.gcp_billing.billing_dataset_id
}
