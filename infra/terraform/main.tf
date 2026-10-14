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
