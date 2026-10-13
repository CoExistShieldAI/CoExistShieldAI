#!/usr/bin/env bash
set -euo pipefail

# Deploy Go Backend to Google Cloud Run
# Usage: ./scripts/deploy_cloudrun.sh <GCP_PROJECT_ID> [REGION]

PROJECT_ID="${1:-${GCP_PROJECT_ID:-}}"
REGION="${2:-asia-northeast1}"
SERVICE_NAME="coexistshield-backend"

if [[ -z "${PROJECT_ID}" ]]; then
  echo "Error: GCP Project ID is required."
  echo "Usage: ./scripts/deploy_cloudrun.sh <PROJECT_ID> [REGION]"
  exit 1
fi

echo "=========================================================="
echo " Deploying CoExist Shield AI Backend to Cloud Run"
echo " Project: ${PROJECT_ID}"
echo " Region:  ${REGION}"
echo " Service: ${SERVICE_NAME}"
echo "=========================================================="

cd "$(dirname "$0")/../backend-cloudrun"

gcloud run deploy "${SERVICE_NAME}" \
  --source . \
  --platform managed \
  --region "${REGION}" \
  --project "${PROJECT_ID}" \
  --allow-unauthenticated \
  --min-instances 0 \
  --max-instances 10 \
  --cpu 1 \
  --memory 256Mi

echo "Deployment finished successfully!"
gcloud run services describe "${SERVICE_NAME}" --platform managed --region "${REGION}" --project "${PROJECT_ID}" --format="value(status.url)"
