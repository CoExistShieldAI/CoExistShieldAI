variable "project_id" {
  description = "Google Cloud / Firebase project ID"
  type        = string
  default     = "coexistshield-ai"
}

variable "region" {
  description = "Default region"
  type        = string
  default     = "asia-northeast1"
}

variable "allowed_referrers" {
  description = "HTTP referrers allowed to use the web API key"
  type        = list(string)
  default = [
    "https://corald.web.app/*",
    "https://corald.firebaseapp.com/*",
    "https://corald-kmp.web.app/*",
    "https://corald-kmp.firebaseapp.com/*",
    "https://corald-admin.web.app/*",
    "https://corald-admin.firebaseapp.com/*",
    "https://coexistshield-ai.web.app/*",
    "https://coexistshield-ai.firebaseapp.com/*",
    "http://localhost:*/*",
    "https://localhost/*",
    "capacitor://localhost/*",
  ]
}

variable "bigquery_dataset_id" {
  description = "BigQuery dataset ID for spatio-temporal data lake"
  type        = string
  default     = "coexist_spatiotemporal_lake"
}

variable "bigquery_location" {
  description = "Geographic location for BigQuery dataset"
  type        = string
  default     = "asia-southeast1"
}

variable "cloud_run_service_account" {
  description = "Service account email running Cloud Run microservices (defaults to compute default SA if empty)"
  type        = string
  default     = ""
}

