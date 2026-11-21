terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

# ---------------------------------------------------------------------------
# BigQuery dataset that will hold the billing export.
#
# NOTE: GCP does not expose the "enable billing export" switch as a native
# Terraform resource. After applying this module, enable the export in the
# Cloud Billing console (Billing > Cost Table > BigQuery export) pointing at
# `google_bigquery_dataset.billing.dataset_id`. This module prepares the
# dataset and the least-privilege service account that reads it.
# ---------------------------------------------------------------------------
resource "google_bigquery_dataset" "billing" {
  dataset_id                 = var.billing_dataset_id
  project                    = var.project_id
  location                   = var.region
  delete_contents_on_destroy = true
}

# Service account with read-only access to the billing dataset.
resource "google_service_account" "billing_reader" {
  account_id   = var.sa_account_id
  display_name = "ack-gcp-guard billing reader"
  project      = var.project_id
}

# roles/billing.viewer lets the SA read cost data.
resource "google_project_iam_member" "billing_viewer" {
  project = var.project_id
  role    = "roles/billing.viewer"
  member  = "serviceAccount:${google_service_account.billing_reader.email}"
}

# Read access to the billing dataset itself.
resource "google_bigquery_dataset_iam_member" "billing_reader" {
  dataset_id = google_bigquery_dataset.billing.dataset_id
  role       = "roles/bigquery.dataViewer"
  member     = "serviceAccount:${google_service_account.billing_reader.email}"
}

resource "google_bigquery_dataset_iam_member" "billing_job_user" {
  dataset_id = google_bigquery_dataset.billing.dataset_id
  role       = "roles/bigquery.jobUser"
  member     = "serviceAccount:${google_service_account.billing_reader.email}"
}

# JSON key for the billing reader SA (used by billing-watcher locally).
resource "google_service_account_key" "billing_reader" {
  service_account_id = google_service_account.billing_reader.name
}
