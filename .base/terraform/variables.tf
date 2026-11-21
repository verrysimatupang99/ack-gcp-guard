variable "alibaba_account_id" {
  type        = string
  description = "Alibaba Cloud account ID"
}

variable "alibaba_region" {
  type        = string
  description = "Alibaba Cloud region"
  default     = "cn-hangzhou"
}

variable "gcp_project_id" {
  type        = string
  description = "GCP project ID"
}

variable "gcp_region" {
  type        = string
  description = "GCP region (default for resources)"
  default     = "asia-southeast1"
}

variable "gcp_billing_dataset" {
  type        = string
  description = "BigQuery dataset ID for the billing export"
  default     = "billing_export"
}

variable "ram_user_name" {
  type        = string
  description = "Name of the Alibaba RAM scanner user"
  default     = "ack-gcp-guard-scanner"
}
