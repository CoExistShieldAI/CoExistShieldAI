#!/usr/bin/env bash
# ==============================================================================
# CoExist Shield AI: User Role Provisioning Script (Issue #46)
# Grants CITIZEN | OFFICIAL | ADMIN roles to Firebase Auth users
# ==============================================================================
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
ADMIN_TOKEN="${ADMIN_TOKEN:-}"

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <USER_UID> <ROLE: CITIZEN|OFFICIAL|ADMIN> [API_BASE_URL]"
  echo "Example: $0 user_abc123 OFFICIAL https://api-coexistshield-cloudrun.a.run.app"
  exit 1
fi

UID_ARG="$1"
ROLE_ARG="$(echo "$2" | tr '[:lower:]' '[:upper:]')"
if [[ -n "${3:-}" ]]; then
  API_BASE_URL="$3"
fi

if [[ "$ROLE_ARG" != "CITIZEN" && "$ROLE_ARG" != "OFFICIAL" && "$ROLE_ARG" != "ADMIN" ]]; then
  echo "Error: Invalid role '$ROLE_ARG'. Must be CITIZEN, OFFICIAL, or ADMIN."
  exit 1
fi

echo "Setting role for UID: $UID_ARG -> $ROLE_ARG on $API_BASE_URL..."

AUTH_HEADER=()
if [[ -n "$ADMIN_TOKEN" ]]; then
  AUTH_HEADER=(-H "Authorization: Bearer $ADMIN_TOKEN")
fi

RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE_URL}/api/v1/admin/set-role" \
  -H "Content-Type: application/json" \
  "${AUTH_HEADER[@]}" \
  -d "{\"uid\":\"${UID_ARG}\",\"role\":\"${ROLE_ARG}\"}")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [[ "$HTTP_CODE" == "200" ]]; then
  echo "✅ Role successfully assigned: $BODY"
else
  echo "❌ Failed to set role (HTTP $HTTP_CODE): $BODY"
  exit 1
fi
