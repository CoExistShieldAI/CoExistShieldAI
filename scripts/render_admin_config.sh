#!/usr/bin/env bash
# Injects GOOGLE_MAPS_API_KEY into the built Angular web-admin distribution index.html.
# Usage: GOOGLE_MAPS_API_KEY=... scripts/render_admin_config.sh [dist_index_html]
set -euo pipefail

# Auto-detect target path whether run from repo root or web-admin/
if [ -n "${1:-}" ]; then
  DIST_INDEX="$1"
elif [ -f "web-admin/dist/web-admin/browser/index.html" ]; then
  DIST_INDEX="web-admin/dist/web-admin/browser/index.html"
elif [ -f "dist/web-admin/browser/index.html" ]; then
  DIST_INDEX="dist/web-admin/browser/index.html"
else
  DIST_INDEX="web-admin/dist/web-admin/browser/index.html"
fi

# Optional fallback to local .env file if key is not yet set in environment
if [ -z "${GOOGLE_MAPS_API_KEY:-}" ]; then
  if [ -f "web-admin/.env" ]; then
    # shellcheck disable=SC2046
    export $(grep -v '^#' web-admin/.env | xargs) 2>/dev/null || true
  elif [ -f ".env" ]; then
    # shellcheck disable=SC2046
    export $(grep -v '^#' .env | xargs) 2>/dev/null || true
  fi
fi

KEY="${GOOGLE_MAPS_API_KEY:-}"

if [ ! -f "$DIST_INDEX" ]; then
  echo "Error: $DIST_INDEX does not exist. Please run 'npm run build' in web-admin first." >&2
  exit 1
fi

if [[ -n "$KEY" && ! "$KEY" =~ ^[A-Za-z0-9_-]+$ ]]; then
  echo "Error: GOOGLE_MAPS_API_KEY has unexpected characters" >&2
  exit 1
fi

if [ -n "$KEY" ]; then
  if [ "${GITHUB_ACTIONS:-false}" = "true" ]; then
    echo "::add-mask::$KEY"
  fi
  echo "Injecting provided GOOGLE_MAPS_API_KEY into $DIST_INDEX..."
else
  echo "::warning::GOOGLE_MAPS_API_KEY is empty or not set. Replacing placeholder with empty string."
fi

export INJECT_TARGET="$DIST_INDEX"
export INJECT_KEY="$KEY"
node -e '
const fs = require("fs");
const target = process.env.INJECT_TARGET;
const key = process.env.INJECT_KEY || "";
let content = fs.readFileSync(target, "utf8");
content = content.replace(/__GOOGLE_MAPS_API_KEY__/g, key);
fs.writeFileSync(target, content, "utf8");
'

echo "Successfully rendered Google Maps configuration in $DIST_INDEX"
