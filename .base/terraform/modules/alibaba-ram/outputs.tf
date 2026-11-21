output "ram_user_name" {
  description = "Name of the RAM scanner user"
  value       = alicloud_ram_user.scanner.name
}

output "access_key_id" {
  description = "Access key ID for the scanner user"
  value       = alicloud_ram_access_key.scanner.id
}

output "access_key_secret" {
  description = "Access key secret for the scanner user (sensitive)"
  value       = alicloud_ram_access_key.scanner.secret
  sensitive   = true
}

output "sts_role_arn" {
  description = "ARN of the STS role the scanner can assume"
  value       = alicloud_ram_role.scanner_sts.arn
}

output "policy_name" {
  description = "Name of the read-only key-scanning policy"
  value       = alicloud_ram_policy.key_scanner.policy_name
}
