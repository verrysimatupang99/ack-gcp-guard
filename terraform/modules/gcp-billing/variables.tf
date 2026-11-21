variable "project_id" {
  type        = string
  description = "GCP project ID that owns the billing export"
}

variable "region" {
  type        = string
  description = "BigQuery dataset location"
  default     = "US"
}

variable "billing_dataset_id" {
  type        = string
  description = "BigQuery dataset ID for the billing export"
  default     = "billing_export"
}

variable "sa_account_id" {
  type        = string
  description = "Account ID (prefix) of the billing reader service account"
  default     = "ack-gcp-guard-billing"
}
