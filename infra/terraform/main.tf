terraform {
  required_version = ">= 1.5"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
    }
  }

  # State lives in GCS. Bucket is supplied at init time:
  #   terraform init -backend-config="bucket=<state-bucket>"
  backend "gcs" {
    prefix = "coexistshield/auth"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

resource "google_project_service" "apis" {
  for_each = toset([
    "identitytoolkit.googleapis.com",
    "securetoken.googleapis.com",
    "apikeys.googleapis.com",
    "bigquery.googleapis.com",
  ])
  project            = var.project_id
  service            = each.value
  disable_on_destroy = false
}


# Enable Email/Password sign-in (Identity Platform / Firebase Auth).
resource "google_identity_platform_config" "auth" {
  project = var.project_id

  sign_in {
    allow_duplicate_emails = false

    email {
      enabled           = true
      password_required = true
    }
  }

  depends_on = [google_project_service.apis]
}

# Browser API key used by the web client. It is a public identifier, but is restricted to
# the Auth APIs and to our hosting origins to limit abuse.
resource "google_apikeys_key" "web" {
  name         = "coexistshield-web-auth"
  display_name = "CoExist Shield AI web client (auth)"
  project      = var.project_id

  restrictions {
    browser_key_restrictions {
      allowed_referrers = var.allowed_referrers
    }

    api_targets {
      service = "identitytoolkit.googleapis.com"
    }
    api_targets {
      service = "securetoken.googleapis.com"
    }
  }

  depends_on = [google_project_service.apis]
}

data "google_project" "current" {
  project_id = var.project_id
}

locals {
  cloud_run_sa = var.cloud_run_service_account != "" ? var.cloud_run_service_account : "${data.google_project.current.number}-compute@developer.gserviceaccount.com"
}

# =============================================================================
# BigQuery Spatio-Temporal Data Lake (Issue #20, #41)
# =============================================================================

resource "google_bigquery_dataset" "spatiotemporal_lake" {
  dataset_id                 = var.bigquery_dataset_id
  friendly_name              = "CoExist Shield Spatio-Temporal Data Lake"
  description                = "Spatio-temporal data lake for wildlife incidents, attractants, and continuous retraining."
  location                   = var.bigquery_location
  project                    = var.project_id
  delete_contents_on_destroy = false

  depends_on = [google_project_service.apis]
}

# Streaming Incidents Table (with Partitioning by timestamp and Clustering by h3_res8, species)
resource "google_bigquery_table" "wildlife_incidents_streaming" {
  dataset_id          = google_bigquery_dataset.spatiotemporal_lake.dataset_id
  table_id            = "wildlife_incidents_streaming"
  project             = var.project_id
  description         = "Real-time streaming ingested wildlife encounters, human attacks, and forensics."
  deletion_protection = false

  time_partitioning {
    type  = "DAY"
    field = "timestamp"
  }

  clustering = ["h3_res8", "species"]

  schema = jsonencode([
    { name = "event_id", type = "STRING", mode = "REQUIRED", description = "Unique stream event ID" },
    { name = "incident_id", type = "STRING", mode = "REQUIRED", description = "Originating incident ID" },
    { name = "species", type = "STRING", mode = "NULLABLE", description = "Target wildlife species" },
    { name = "category", type = "STRING", mode = "NULLABLE", description = "Incident category (SIGHTING, ATTACK, etc.)" },
    { name = "geo_location", type = "GEOGRAPHY", mode = "NULLABLE", description = "GIS Point geography (POINT(lon lat))" },
    { name = "latitude", type = "FLOAT", mode = "REQUIRED", description = "WGS84 Latitude" },
    { name = "longitude", type = "FLOAT", mode = "REQUIRED", description = "WGS84 Longitude" },
    { name = "timestamp", type = "TIMESTAMP", mode = "REQUIRED", description = "Event occurrence timestamp" },
    { name = "is_incursion", type = "INTEGER", mode = "NULLABLE", description = "1 if incursion into human settlement, 0 otherwise" },
    { name = "verification_state", type = "STRING", mode = "NULLABLE", description = "Verification workflow status" },
    { name = "h3_res8", type = "STRING", mode = "NULLABLE", description = "Uber H3 Resolution 8 spatial cell index" }
  ])
}

# Unmanaged Persimmon Trees Table (Kaki attractant)
resource "google_bigquery_table" "unmanaged_persimmon_trees" {
  dataset_id          = google_bigquery_dataset.spatiotemporal_lake.dataset_id
  table_id            = "unmanaged_persimmon_trees"
  project             = var.project_id
  description         = "Geolocated abandoned persimmon (kaki) trees serving as bear attractants."
  deletion_protection = false

  clustering = ["status", "fruit_load"]

  schema = jsonencode([
    { name = "tree_id", type = "STRING", mode = "REQUIRED" },
    { name = "location", type = "GEOGRAPHY", mode = "REQUIRED" },
    { name = "latitude", type = "FLOAT", mode = "REQUIRED" },
    { name = "longitude", type = "FLOAT", mode = "REQUIRED" },
    { name = "status", type = "STRING", mode = "NULLABLE" },
    { name = "fruit_load", type = "STRING", mode = "NULLABLE" },
    { name = "updated_at", type = "TIMESTAMP", mode = "NULLABLE" }
  ])
}

# Roadkill Carcasses Table (Scavenger predator magnet)
resource "google_bigquery_table" "roadkill_carcasses" {
  dataset_id          = google_bigquery_dataset.spatiotemporal_lake.dataset_id
  table_id            = "roadkill_carcasses"
  project             = var.project_id
  description         = "Roadkill animal carcasses acting as predatory magnets."
  deletion_protection = false

  time_partitioning {
    type  = "DAY"
    field = "discovered_at"
  }

  schema = jsonencode([
    { name = "carcass_id", type = "STRING", mode = "REQUIRED" },
    { name = "location", type = "GEOGRAPHY", mode = "REQUIRED" },
    { name = "species", type = "STRING", mode = "NULLABLE" },
    { name = "decay_stage", type = "STRING", mode = "NULLABLE" },
    { name = "cleared_at", type = "TIMESTAMP", mode = "NULLABLE" },
    { name = "discovered_at", type = "TIMESTAMP", mode = "REQUIRED" }
  ])
}

# Agricultural Attractants Table (Sugarcane, banana, jackfruit)
resource "google_bigquery_table" "agricultural_attractants" {
  dataset_id          = google_bigquery_dataset.spatiotemporal_lake.dataset_id
  table_id            = "agricultural_attractants"
  project             = var.project_id
  description         = "Agricultural crops attracting megaherbivores (elephants)."
  deletion_protection = false

  schema = jsonencode([
    { name = "plot_id", type = "STRING", mode = "REQUIRED" },
    { name = "crop_type", type = "STRING", mode = "REQUIRED" },
    { name = "location", type = "GEOGRAPHY", mode = "REQUIRED" },
    { name = "harvest_stage", type = "STRING", mode = "NULLABLE" },
    { name = "updated_at", type = "TIMESTAMP", mode = "NULLABLE" }
  ])
}

# Meteorological Observations Table (Weather and lunar illumination)
resource "google_bigquery_table" "meteorological_observations" {
  dataset_id          = google_bigquery_dataset.spatiotemporal_lake.dataset_id
  table_id            = "meteorological_observations"
  project             = var.project_id
  description         = "Localized meteorological observations discretized by H3 cell."
  deletion_protection = false

  time_partitioning {
    type  = "DAY"
    field = "observation_time"
  }

  clustering = ["h3_index"]

  schema = jsonencode([
    { name = "h3_index", type = "STRING", mode = "REQUIRED" },
    { name = "observation_time", type = "TIMESTAMP", mode = "REQUIRED" },
    { name = "wind_bearing_deg", type = "FLOAT", mode = "NULLABLE" },
    { name = "wind_speed_mps", type = "FLOAT", mode = "NULLABLE" },
    { name = "precipitation_mm_24h", type = "FLOAT", mode = "NULLABLE" },
    { name = "moon_illumination_fraction", type = "FLOAT", mode = "NULLABLE" }
  ])
}

# IAM: Grant BigQuery dataEditor and jobUser roles to Cloud Run Service Account
resource "google_project_iam_member" "cloudrun_bigquery_editor" {
  project = var.project_id
  role    = "roles/bigquery.dataEditor"
  member  = "serviceAccount:${local.cloud_run_sa}"
}

resource "google_project_iam_member" "cloudrun_bigquery_user" {
  project = var.project_id
  role    = "roles/bigquery.jobUser"
  member  = "serviceAccount:${local.cloud_run_sa}"
}


