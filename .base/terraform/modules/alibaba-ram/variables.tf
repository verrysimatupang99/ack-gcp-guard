variable "account_id" {
  type        = string
  description = "Alibaba Cloud account ID (for the STS role trust policy)"
}

variable "ram_user_name" {
  type        = string
  description = "Name of the RAM user created for scanning"
  default     = "ack-gcp-guard-scanner"
}

variable "display_name" {
  type        = string
  description = "Display name for the RAM user"
  default     = "ack-gcp-guard scanner"
}

variable "policy_name" {
  type        = string
  description = "Name of the read-only key-scanning policy"
  default     = "AckGcpGuardKeyScannerReadOnly"
}

variable "sts_role_name" {
  type        = string
  description = "Name of the STS role the scanner can assume"
  default     = "ack-gcp-guard-scanner-sts"
}
