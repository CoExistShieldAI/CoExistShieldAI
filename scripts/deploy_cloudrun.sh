#!/usr/bin/env bash
set -euo pipefail

# Deploy Go Backend & Continuous Retraining Job to Google Cloud Run
# Usage: ./scripts/deploy_cloudrun.sh <GCP_PROJECT_ID> [REGION]

PROJECT_ID="${1:-${GCP_PROJECT_ID:-}}"
REGION="${2:-asia-southeast1}"
SERVICE_NAME="coexistshield-backend"
JOB_NAME="coexistshield-retrain-job"

if [[ -z "${PROJECT_ID}" ]]; then
  echo "Error: GCP Project ID is required."
  echo "Usage: ./scripts/deploy_cloudrun.sh <PROJECT_ID> [REGION]"
  exit 1
fi

echo "=========================================================="
echo " Deploying CoExist Shield AI to Google Cloud Run (Singapore)"
echo " Project: ${PROJECT_ID}"
echo " Region:  ${REGION} (Singapore Demo Day Location)"
echo " Service: ${SERVICE_NAME}"
echo " Job:     ${JOB_NAME}"
echo "=========================================================="

cd "$(dirname "$0")/../backend-cloudrun"

# Check Secret Manager for optional API secrets
SECRETS=()
if gcloud secrets describe GEMINI_API_KEY --project="${PROJECT_ID}" &>/dev/null; then
  echo "✓ Detected Secret Manager secret 'GEMINI_API_KEY'. Mounting into container."
  SECRETS+=("GEMINI_API_KEY=GEMINI_API_KEY:latest")
else
  echo "⚠ Secret 'GEMINI_API_KEY' not found in Secret Manager; fallback heuristic AI will activate if key is omitted."
fi

if gcloud secrets describe GOOGLE_MAPS_API_KEY --project="${PROJECT_ID}" &>/dev/null; then
  echo "✓ Detected Secret Manager secret 'GOOGLE_MAPS_API_KEY'. Mounting into container."
  SECRETS+=("GOOGLE_MAPS_API_KEY=GOOGLE_MAPS_API_KEY:latest")
elif gcloud secrets describe GEOCODING_API_KEY --project="${PROJECT_ID}" &>/dev/null; then
  echo "✓ Detected Secret Manager secret 'GEOCODING_API_KEY'. Mounting into container."
  SECRETS+=("GEOCODING_API_KEY=GEOCODING_API_KEY:latest")
fi

if gcloud secrets describe FIRMS_MAP_KEY --project="${PROJECT_ID}" &>/dev/null; then
  echo "✓ Detected Secret Manager secret 'FIRMS_MAP_KEY'. Mounting into container."
  SECRETS+=("FIRMS_MAP_KEY=FIRMS_MAP_KEY:latest")
fi

SECRET_FLAG=""
if [[ ${#SECRETS[@]} -gt 0 ]]; then
  SECRET_FLAG="--set-secrets=$(IFS=,; echo "${SECRETS[*]}")"
fi

MEDIA_BUCKET="${GCS_MEDIA_BUCKET:-corald-coexistshield-media}"

# 1. Deploy Cloud Run Service (Microservice APIs & Ingestion Pipelines)
echo "--> Deploying Cloud Run Service: ${SERVICE_NAME}..."
gcloud run deploy "${SERVICE_NAME}" \
  --source . \
  --platform managed \
  --region "${REGION}" \
  --project "${PROJECT_ID}" \
  --allow-unauthenticated \
  --min-instances 0 \
  --max-instances 10 \
  --cpu 1 \
  --memory 512Mi \
  --timeout 300 \
  --quiet \
  --set-env-vars="BIGQUERY_STREAMING_ENABLED=true,BIGQUERY_PROJECT_ID=${PROJECT_ID},BIGQUERY_DATASET=coexist_spatiotemporal_lake,BIGQUERY_TABLE=wildlife_incidents_streaming,GCP_PROJECT_ID=${PROJECT_ID},AUTH_PROJECT_ID=${PROJECT_ID},GCS_MEDIA_BUCKET=${MEDIA_BUCKET}" \
  ${SECRET_FLAG}

SERVICE_URL=$(gcloud run services describe "${SERVICE_NAME}" \
  --platform managed \
  --region "${REGION}" \
  --project "${PROJECT_ID}" \
  --format="value(status.url)")

echo "✓ Service deployed successfully: ${SERVICE_URL}"

# 2. Deploy Cloud Run Job (Continuous Retraining Job)
echo "--> Deploying Continuous Retraining Cloud Run Job: ${JOB_NAME}..."
gcloud run jobs deploy "${JOB_NAME}" \
  --source . \
  --command "go" \
  --args "run,./cmd/job,--task=all" \
  --region "${REGION}" \
  --project "${PROJECT_ID}" \
  --cpu 1 \
  --memory 512Mi \
  --max-retries 1 \
  --quiet \
  --set-env-vars="BIGQUERY_STREAMING_ENABLED=true,BIGQUERY_PROJECT_ID=${PROJECT_ID},BIGQUERY_DATASET=coexist_spatiotemporal_lake,BIGQUERY_TABLE=wildlife_incidents_streaming,GCP_PROJECT_ID=${PROJECT_ID},AUTH_PROJECT_ID=${PROJECT_ID},GCS_MEDIA_BUCKET=${MEDIA_BUCKET}" \
  ${SECRET_FLAG} || echo "⚠ Job deploy completed with warning"


# 3. Health check verification
echo "--> Verifying Service Health via /health..."
HEALTH_RESPONSE=$(curl -s "${SERVICE_URL}/health" || echo "FAILED")
echo "Healthcheck response: ${HEALTH_RESPONSE}"

echo "=========================================================="
echo " Cloud Run Deployment Complete!"
echo " Public Endpoint: ${SERVICE_URL}"
echo " Health Status:   ${HEALTH_RESPONSE}"
echo "=========================================================="

