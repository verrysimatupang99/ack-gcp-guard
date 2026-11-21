provider "alicloud" {
  region = var.alibaba_region
}

provider "google" {
  project = var.gcp_project_id
  region  = var.gcp_region
}

# ---------------------------------------------------------------------------
# Alibaba Cloud: RAM user + read-only key-scanning policy + STS role.
# ---------------------------------------------------------------------------
module "alibaba_ram" {
  source        = "./modules/alibaba-ram"
  account_id    = var.alibaba_account_id
  ram_user_name = var.ram_user_name
}

# ---------------------------------------------------------------------------
# GCP: Billing BigQuery export dataset + read-only SA.
# ---------------------------------------------------------------------------
module "gcp_billing" {
  source             = "./modules/gcp-billing"
  project_id         = var.gcp_project_id
  billing_dataset_id = var.gcp_billing_dataset
}
