#!/usr/bin/env bash
# ==============================================================================
# CoExist Shield AI: Seed Firestore Emulator / Local Backend (Issue #38)
# Populates canonical test incidents, roadkills, and geofence entities
# ==============================================================================
set -euo pipefail

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
FIRESTORE_EMULATOR_HOST="${FIRESTORE_EMULATOR_HOST:-localhost:8080}"

echo "🌱 Seeding CoExist Shield AI development baseline..."
echo "Target API: $API_BASE_URL"

# 1. Bear Sighting in Morioka City
curl -s -X POST "${API_BASE_URL}/api/v1/incidents" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "inc_iwate_01",
    "category": "WILDLIFE_SIGHTING",
    "species": "BEAR",
    "location": {
      "latitude": 39.7020,
      "longitude": 141.1520
    },
    "timestamp": "2026-10-14T07:20:00Z",
    "region": "JAPAN",
    "description": "Adult Asian Black Bear spotted feeding on unpicked persimmon tree near Matsugaeda Elementary School perimeter.",
    "reporterId": "user_citizen_01",
    "reporterRole": "CITIZEN",
    "verificationStatus": "VERIFIED"
  }' > /dev/null
echo "  ✓ Seeded Incident: inc_iwate_01 (Asian Black Bear)"

# 2. Roadkill Carcass Hazard
curl -s -X POST "${API_BASE_URL}/api/v1/incidents" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "inc_roadkill_01",
    "category": "ROADKILL",
    "species": "WILD_BOAR",
    "location": {
      "latitude": 39.7085,
      "longitude": 141.1490
    },
    "timestamp": "2026-10-14T08:10:00Z",
    "region": "JAPAN",
    "description": "Bloated wild boar carcass on Route 4 bypass shoulder emitting olfactory attractant plume toward residential fringe.",
    "reporterId": "user_official_02",
    "reporterRole": "OFFICIAL",
    "verificationStatus": "VERIFIED"
  }' > /dev/null
echo "  ✓ Seeded Roadkill: inc_roadkill_01 (Wild Boar)"

# 3. Tiger Depredation Incident in Sumatra
curl -s -X POST "${API_BASE_URL}/api/v1/incidents" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "inc_sumatra_01",
    "category": "LIVESTOCK_DEPREDATION",
    "species": "TIGER",
    "location": {
      "latitude": 0.8251,
      "longitude": 101.5302
    },
    "timestamp": "2026-10-14T03:45:00Z",
    "region": "INDONESIA",
    "description": "Sumatran Tiger depredation on cattle pasture buffer bordering Block 14 oil palm concession. High risk of return encounter.",
    "reporterId": "user_ranger_05",
    "reporterRole": "OFFICIAL",
    "verificationStatus": "VERIFIED"
  }' > /dev/null
echo "  ✓ Seeded Incident: inc_sumatra_01 (Sumatran Tiger)"

echo "✅ Database seeding completed successfully."
