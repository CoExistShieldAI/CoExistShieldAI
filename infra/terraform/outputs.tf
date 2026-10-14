output "firebase_web_api_key" {
  description = "Web API key injected into config.js (COEXIST_FIREBASE_API_KEY)"
  value       = google_apikeys_key.web.key_string
  sensitive   = true
}
