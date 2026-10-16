output "firebase_web_api_key" {
  description = "Web API key injected into config.js (COEXIST_FIREBASE_API_KEY)"
  value       = google_apikeys_key.web.key_string
  sensitive   = true
}

output "bigquery_dataset_id" {
  description = "BigQuery dataset ID"
  value       = google_bigquery_dataset.spatiotemporal_lake.dataset_id
}

output "bigquery_incidents_table_id" {
  description = "BigQuery streaming incidents table ID"
  value       = google_bigquery_table.wildlife_incidents_streaming.table_id
}

