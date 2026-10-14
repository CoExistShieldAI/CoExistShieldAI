# CoExist Shield AI - Admin Operations Web Portal

> **Mission-Critical Web Operations Center for Dynamic Geofencing, Human-Wildlife Coexistence & One Health Surveillance across JAPAC**

Built with **Angular 22 (Latest version)**, **Google Maps API**, and **Tailwind CSS**, communicating with the **Go Cloud Run Backend**.

---

## 1. Key Capabilities & Feature Modules

### 🗺️ JAPAC Dynamic Geofence Console (`/geofences`)
- **Interactive Google Map:** Centered on the entire JAPAC theater (`lat: 15.0, lng: 105.0, zoom: 4`) covering Japan, Southeast Asia, and South Asia.
- **AI Geofence Visualization:** Render 3-tier predictive danger boundaries (Tier 1 Caution, Tier 2 Alert, Tier 3 Critical Danger).
- **Manual Debug & Exception Overrides:**
  - Create manual geofences with **both Circles (Center + Radius in meters) and Polygons (Multi-point vertices)**.
  - **Persistent Protection:** Manual fences are marked `is_persistent: true` and are preserved during AI automated model sweeps.
  - **Bulk Purge:** One-click `"Clear All Manual Fences"` button for rapid debug reset.
- **AI Recalculation Trigger:** Instant execution of predictive geofencing models across multi-source streams.

### 🦌 Wildlife Encounter Registry (`/incidents/wildlife`)
- Real-time logging of apex carnivores and megafauna incursions (Asian Elephants, Sumatran Tigers, Bengal Tigers, Brown/Black Bears).
- Filterable and searchable table with status workflows (`Unverified` -> `Verified` -> `Resolved`).
- Detail inspect and deletion actions.

### 🚗 Roadkill Surveillance & Desktop Big-Screen Entry (`/incidents/roadkill`)
- **Predator Magnet Mitigation:** Real-time tracking of decaying roadside carcasses that act as leading indicators of carnivore attacks.
- **Data Cleansing:** Instant toggle to exclude low-credibility entries (`is_excluded: true`) to avoid corrupting AI scent-plume dispersion training.
- **Desktop Big-Screen Registration Workspace:**
  - 2-column layout optimized for operations consoles.
  - Interactive Google Map click-to-pin drop with automatic latitude/longitude and address reverse geocoding.
  - Carcass decay stages (`Fresh <6h`, `Moderate 6-24h`, `Advanced >24h`) for scent plume modeling.

### 🌱 Ecological & Crop Attractant Layers (`/environmental`)
- Management of causal root attractants driving wildlife incursions:
  - **Thailand & APAC:** Sugarcane and fruit phenology harvesting windows.
  - **Indonesia & India:** Unfenced livestock corrals & night pastures.
  - **Japan:** Abandoned persimmon (*kaki*) tree hazard clusters in aging rural valleys.

### ⚡ Operations, Emergency Tools & GIS Export (`/operations`)
- **Emergency Area Push Broadcast:** Immediate dispatch of audio siren / push warning nets to all mobile users within designated geofences.
- **GIS & CSV Data Export:** One-click download of GeoJSON (RFC 7946) and CSV formats for QGIS, ArcGIS, and municipal reporting.

---

## 2. Getting Started

### Prerequisites
- Node.js `v22.14.0+`
- npm `10.9.0+`
- Go `1.22+` (for backend)

### Running the Admin Portal Locally

```bash
# In web-admin directory
cd web-admin

# Install dependencies (already resolved)
npm install

# Start Angular development server
npm start
# -> Access console at http://localhost:4200
```

### Running with Go Backend
```bash
# In backend-cloudrun directory
cd backend-cloudrun
go run ./cmd/server
# -> REST API listens on http://localhost:8080/api/v1/admin
```
