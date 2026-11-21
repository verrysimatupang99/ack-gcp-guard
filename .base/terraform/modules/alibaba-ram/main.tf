terraform {
  required_providers {
    alicloud = {
      source  = "aliyun/alicloud"
      version = "~> 1.230.0"
    }
  }
}

# ---------------------------------------------------------------------------
# RAM user used ONLY by ack-gcp-guard to scan access keys (least privilege).
# ---------------------------------------------------------------------------
resource "alicloud_ram_user" "scanner" {
  name         = var.ram_user_name
  display_name = var.display_name
  force        = true
}

# Access key pair for the scanner user (output as sensitive).
resource "alicloud_ram_access_key" "scanner" {
  user_name = alicloud_ram_user.scanner.name
  status    = "Active"
}

# Read-only policy: exactly the two actions the ram-scanner needs.
resource "alicloud_ram_policy" "key_scanner" {
  policy_name = var.policy_name
  description  = "Read-only access to list RAM users and access keys (ack-gcp-guard)"
  document     = <<EOF
{
  "Version": "1",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ram:ListUsers",
        "ram:ListAccessKeys"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "alicloud_ram_user_policy_attachment" "scanner" {
  user_name   = alicloud_ram_user.scanner.name
  policy_name = alicloud_ram_policy.key_scanner.policy_name
  policy_type = "Custom"
}

# ---------------------------------------------------------------------------
# STS role the guard can assume (optional, for cross-account / central scans).
# ---------------------------------------------------------------------------
resource "alicloud_ram_role" "scanner_sts" {
  name     = var.sts_role_name
  document = <<EOF
{
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "RAM": ["acs:ram::${var.account_id}:user/${var.ram_user_name}"]
      }
    }
  ],
  "Version": "1"
}
EOF
}

resource "alicloud_ram_role_policy_attachment" "scanner_sts" {
  role_name   = alicloud_ram_role.scanner_sts.name
  policy_name = alicloud_ram_policy.key_scanner.policy_name
  policy_type = "Custom"
}
