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
    "https://coexistshield-ai.web.app/*",
    "https://coexistshield-ai.firebaseapp.com/*",
    "http://localhost:*/*",
  ]
}
