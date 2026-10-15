# CoExist Shield AI - Admin Operations Web Portal

> **Mission-Critical Web Operations Center for Dynamic Geofencing, Human-Wildlife Coexistence & One Health Surveillance across JAPAC**

Built with **Angular 22 (Latest version)**, **Google Maps API**, and **Tailwind CSS**, communicating with the **Go Cloud Run Backend**.

---

## 1. Key Capabilities & Feature Modules

### 🗺️ JAPAC Dynamic Geofence Console (`/geofences`)
- **Interactive Google Map:** Centered on the entire JAPAC theater (`lat: 15.0, lng: 105.0, zoom: 4`) covering Japan, Southeast Asia, and South Asia.
- **AI & Manual Geofence Visualization:** Render 3-tier predictive danger boundaries (Tier 1 Caution, Tier 2 Alert, Tier 3 Critical Danger) with circle/polygon authoring.
- **Multi-Layer Empirical GIS Overlay (Epic #60):**
  - **Layer Visibility Checkboxes:** Toggle AI Fences, Manual Fences, Wildlife Sightings, Roadkill Carcasses, and Attractant Layers directly in the toolbar.
  - **Temporal & Date Filtering:** Quick-select time windows (`Past 24h`, `Past 7d`, `Past 30d`, `All Time`) plus custom date ranges (`From` - `To`) for incident investigation.
  - **Secondary Filters:** Narrow by Target Species (`Asian Elephant`, `Sumatran Tiger`, `Bear`, etc.) and Threat Level (`Tier 1`, `Tier 2`, `Tier 3`).
  - **One-Click Geofence from Empirical Incidents:** Convert inspected wildlife encounters, fresh roadkills, or crop attractants into active geofences with pre-populated coordinates and context-calibrated radii.
  - **Interactive Map-Click Placement:** Click anywhere on the map canvas to set circular centers or draw polygon boundary vertices with live visual preview.
  - **High-Risk Visual Callouts:** Pulsing alerts for fresh roadkill (<6h) scent-plume attractants and Tier 3 Critical incursions.
  - **Bulk Purge & Reset:** One-click `"Clear Manual Fences"` and `"Reset JAPAC View"`.

### 🦌 Wildlife Encounter Registry (`/incidents/wildlife`)
- Real-time logging of apex carnivores and megafauna incursions (Asian Elephants, Sumatran Tigers, Bengal Tigers, Brown/Black Bears).
- **Dual-Tab Workspace:**
  - `Sightings Registry`: Filterable and searchable table with status workflows (`Unverified` -> `Verified` -> `Resolved` -> `False Alarm`).
  - `+ Log Wildlife Sighting`: 2-column Map-Picker workspace with interactive Google Map pin drop, reverse geocoding, and multi-photo upload.
- Detail inspection, verify, and deletion actions.

### 🚗 Roadkill Surveillance & Desktop Big-Screen Entry (`/incidents/roadkill`)
- **Predator Magnet Mitigation:** Real-time tracking of decaying roadside carcasses that act as leading indicators of carnivore attacks.
- **Data Cleansing:** Instant toggle to exclude low-credibility entries (`is_excluded: true`) to avoid corrupting AI scent-plume dispersion training.
- **Dual-Tab Workspace:**
  - `History & Data Cleansing`: Credibility filters, active predator magnets counter (`Pending Removal`), and carcass disposal workflow (`Disposed`).
  - `+ Log Roadkill`: 2-column operations workspace with interactive Google Map click-to-pin, auto-populated GPS/address, decay stages (`Fresh <6h`, `Moderate 6-24h`, `Advanced >24h`), and multi-photo attachments.

### 🌱 Ecological & Crop Attractant Layers (`/environmental`)
- Management of causal root attractants driving wildlife incursions:
  - **Thailand & APAC:** Sugarcane and fruit phenology harvesting windows.
  - **Indonesia & India:** Unfenced livestock corrals & night pastures.
  - **Japan:** Abandoned persimmon (*kaki*) tree hazard clusters in aging rural valleys.
- **Dual-Tab Workspace & Category-First Wizard:**
  - `Attractant Layers Registry`: Regional overview cards and searchable table with risk weights and active harvest windows.
  - `+ Log Attractant Layer`: Poka-Yoke 2-step wizard. Step 1 selects ecological context; Step 2 auto-configures context-aware default radii (Sugarcane: 3000m, Livestock: 1500m, Kaki: 500m, Water Holes: 1000m) with real-time map circle radius overlay preview.

### 📸 Multi-Photo Evidence Upload & Google Cloud Storage
- Integrated `ImageUploadComponent` across Wildlife, Roadkill, and Environmental forms.
- Up to 5 photos per incident with client-side canvas compression for files >1MB.
- Direct multipart upload to backend (`/api/v1/admin/upload`) storing into Cloud Storage (`corald-coexistshield-media`).

### ⚡ Operations, Emergency Tools & GIS Export (`/operations`)
- **Emergency Area Push Broadcast:** Immediate dispatch of audio siren / push warning nets to all mobile users within designated geofences.
- **GIS & CSV Data Export:** One-click download of GeoJSON (RFC 7946) and CSV formats for QGIS, ArcGIS, and municipal reporting.

---

## 2. Live Deployment & Review

- **Live Operations Portal URL:** **`https://corald-admin.web.app`** (Mirror: `https://corald-admin.firebaseapp.com`)
- **Live Backend API (Cloud Run):** `https://coexistshield-backend-4cerxfo4wq-as.a.run.app/api/v1/admin`

---

## 3. Getting Started

### Prerequisites
- Node.js `v22.14.0+`
- npm `10.9.0+`
- Go `1.24+` (for backend)

### Google Maps API Key Configuration

The Google Maps API key is never committed to git in plain text.

- **CI/CD Deployment (GitHub Actions):**
  Set the repository secret **`GOOGLE_MAPS_API_KEY`** in:
  `GitHub Repository -> Settings -> Secrets and variables -> Actions -> New repository secret`
  The deployment workflow (`.github/workflows/deploy-hosting.yml`) automatically builds `web-admin` and injects the secret into the production distribution using `scripts/render_admin_config.sh`.

- **Local Development:**
  Create a `.env` file in `web-admin/` or the repository root (both ignored by git):
  ```env
  GOOGLE_MAPS_API_KEY=AIzaSy...
  ```
  Or pass it as an environment variable when building:
  ```bash
  GOOGLE_MAPS_API_KEY=AIzaSy... npm run build
  ../scripts/render_admin_config.sh
  ```

### Running the Admin Portal Locally

```bash
# In web-admin directory
cd web-admin

# Install dependencies
npm install

# Start Angular development server
npm start
# -> Access console at http://localhost:4200
```

### Building & Deploying to Firebase Hosting

```bash
# Build production bundle
npm run build

# Inject Maps API key into built bundle (reads GOOGLE_MAPS_API_KEY from env or .env)
../scripts/render_admin_config.sh

# Deploy to Firebase Hosting target 'admin'
firebase deploy --only hosting:admin
```

### Running with Go Backend
```bash
# In backend-cloudrun directory
cd backend-cloudrun
PORT=8080 go run ./cmd/server
# -> REST API listens on http://localhost:8080/api/v1/admin
```
