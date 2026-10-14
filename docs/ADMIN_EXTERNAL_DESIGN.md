# CoExist Shield AI Admin Web Portal - External Design Specification

**Document Version:** 1.0.0  
**Status:** Approved  
**Language:** English  
**Parent Epic:** [Issue #52](https://github.com/CoExistShieldAI/CoExistShieldAI/issues/52)

---

## 1. System Overview & Architecture

### 1.1 Purpose & Role
The **CoExist Shield AI Admin Web Portal** is the mission-critical operational center for rangers, plantation safety officers, municipal wildlife controllers, and platform engineers. It provides:
1. Real-time visual monitoring of AI-generated predictive geofences across the JAPAC theater.
2. Manual debugging and emergency override tools (circle and polygon geofences).
3. Field incident lifecycle management (Wildlife encounters and Roadkill reporting).
4. Ecological attractant data control (sugar crops, abandoned persimmon trees, livestock grazing).
5. Operational tools: AI engine recalculation triggers, emergency push broadcasts, and GIS/CSV export.

#### 1.2 Technology Stack
- **Frontend Framework:** Angular 22 (Latest version) (Standalone Components, Signals, Reactive Forms, Injectable Services)
- **Map Engine:** Google Maps JavaScript API with `@angular/google-maps` & Drawing Library
- **UI Styling:** Tailwind CSS & Lucide Icons (clean, high-density dashboard UI)
- **Cloud Storage:** Google Cloud Storage (`corald-coexistshield-media`) for multi-photo evidence uploads
- **Backend:** Go (`backend-cloudrun/internal/admin`), REST API endpoints with structured JSON responses
- **Data Store:** Cloud Firestore & In-Memory Geospatial Spatial Index

```
+-------------------------------------------------------------+
|             Angular 22 Admin Web Portal (English)           |
|  +---------------+  +------------------+  +--------------+  |
|  | Dashboard     |  | JAPAC Map Console|  | Incident Hub |  |
|  | (KPIs/Charts) |  | (Circle/Polygon) |  | (Wildlife/RK)|  |
|  +---------------+  +------------------+  +--------------+  |
+------------------------------+------------------------------+
                               | HTTPS / REST (JSON) & Multipart Uploads
+------------------------------v------------------------------+
|                 Go Backend (backend-cloudrun)                |
|  +------------------+  +------------------+  +-----------+  |
|  | Admin API Router |  | Geofence Service |  | GCS Media |  |
|  | (/api/v1/admin)  |  | (Manual/AI Persist)| (Uploader) |  |
|  +------------------+  +------------------+  +-----------+  |
+------------------------------+------------------------------+
                               |
               Cloud Firestore & Google Cloud Storage
```

---

## 2. Navigation & Layout Design

### 2.1 Global Shell Layout & Workspace Conventions
The web application uses a persistent left sidebar with desktop-optimized layout:
- **Header:**
  - Active Theater indicator: `JAPAC Operational Theater`
  - Quick Action: `[Run AI Geofence Engine]` button (with spinner and status badge)
  - Admin Profile & Sign-out
- **Sidebar Menu:**
  - `Dashboard` (`/dashboard`)
  - `Geofence Map` (`/geofences`)
  - `Wildlife Sightings` (`/incidents/wildlife`)
  - `Roadkill Management` (`/incidents/roadkill`)
  - `Environmental Attractants` (`/environmental`)
  - `Operations & Export` (`/operations`)
- **Unified Workspace Pattern:**
  - All data management views (Wildlife, Roadkill, Environmental) share a consistent dual-tab navigation pattern:
    - **Tab 1: History / Registry Table** with comprehensive search, filters, credibility toggles, and status badges.
    - **Tab 2: Entry Workspace (Map-Picker)** featuring a 2-column layout with input forms on the left and a full-height interactive Google Map on the right for pinpoint coordinate selection and reverse geocoding.

---

## 3. Screen Specifications & User Interactions

### 3.1 Screen 1: Executive Dashboard (`/dashboard`)
**Objective:** Deliver immediate situational awareness regarding active threats, roadkill decay risks, and platform activity.

- **KPI Cards:**
  1. **Active Geofences:** Total count, broken down by Tier 1 (Caution), Tier 2 (Warning), Tier 3 (Critical), and Manual Overrides.
  2. **Active Threats (24h):** Wildlife sightings logged in the past 24 hours.
  3. **High-Risk Roadkill:** Roadkill count with status `Pending Removal` (active predator magnets).
  4. **Registered App Users:** Total mobile early-warning recipients.
- **Incident Severity Distribution:** Interactive visual chart showing incident shares across target species (Sumatran Tiger, Asian Elephant, Brown Bear, Asian Black Bear).
- **Recent Incident Feed:** Scrollable real-time event log with severity badges and direct navigation links.

---

### 3.2 Screen 2: JAPAC Geofence Map Console (`/geofences`)
**Objective:** Interactive GIS map covering Japan, Southeast Asia, and South Asia for observing AI predictive zones and authoring manual geofences.

- **Map Configuration:**
  - **Center:** `15.0° N, 105.0° E` (JAPAC regional center covering Japan to Sumatra/India)
  - **Initial Zoom Level:** `4`
  - **Map Controls:** Satellite / Terrain / Roadmap toggle, Fullscreen, Reset View to JAPAC.
- **Visual Styling of Geofences:**
  - **Tier 1 (Notice / Caution):** Yellow fill (`#FACC15`, opacity 0.25), yellow border.
  - **Tier 2 (Alert / Warning):** Orange fill (`#FB923C`, opacity 0.35), orange border.
  - **Tier 3 (Critical Danger):** Red fill (`#EF4444`, opacity 0.45), red pulsing border.
  - **Manual Geofences:** Distinct Cyan/Purple border with a `[MANUAL]` badge in info windows.
- **Multi-Layer Empirical GIS Overlay (Issue #60, #61):**
  - **Layer Visibility Toggles (Checkboxes in Toolbar):**
    - `AI Fences` (Blue indicator)
    - `Manual Fences` (Purple indicator)
    - `Wildlife Sightings` (Cyan circular symbol with Tier 1/2/3 badge)
    - `Roadkill Carcasses` (Rose/Amber symbol with pulsing `[FRESH]` scent-plume alert for <6h carcasses)
    - `Attractant Layers` (Emerald/Orange/Purple/Cyan circles calibrated to ecological radius)
  - **Dynamic Entity Inspector:** Clicking any marker or circle displays entity metadata, attached field photos, and immediate action buttons.
- **Temporal & Date Filtering (Issue #62):**
  - Quick-select time windows: `Past 24h`, `Past 7d`, `Past 30d`, `All Time`.
  - Custom date range picker: From Date / To Date with real-time incident filtering.
  - Dynamic signal counter: `X Active Geofences • Y Empirical Signals`.
  - Secondary filters: Target species dropdown and threat tier selector.
- **One-Click Geofence from Incident & Attractant (Issue #63):**
  - Instant transition from an inspected sighting, carcass, or crop zone to an active geofence.
  - Auto-populates coordinates, target species, threat level, and context-calibrated radii (e.g., Fresh Roadkill 4500m, Wildlife 3000-4000m, Attractant layer radius).
- **Interactive Map-Click Placement (Issue #63):**
  - Toggle interactive drawing mode (`🎯 Click on Map to Position`).
  - Click map canvas directly to set Circle center point or append Polygon vertices while observing adjacent incident markers.
- **Bulk Cleanup Action:**
  - `[Clear Manual Fences]` button in toolbar with confirmation dialog.

---

### 3.3 Screen 3: Wildlife Sighting & Encounter Hub (`/incidents/wildlife`)
**Objective:** Manage, verify, filter, and register wildlife sightings and attack incidents across JAPAC.

- **Dual-Tab Workspace Layout:**
  - **Tab 1: Sightings Registry:**
    - **Filter & Search Bar:** Search by keyword (species, road, reporter); filter by Species (`Asian Elephant`, `Sumatran Tiger`, `Bengal Tiger`, `Brown Bear`, `Asian Black Bear`) and Threat Tier (`Tier 1 Caution`, `Tier 2 Alert`, `Tier 3 Critical`).
    - **Data Table:** Displays Reported At, Species, Location & Coordinates, Threat Level badge, Verification Status (`Unverified`, `Verified`, `Resolved`, `False Alarm`), Reporter, and Actions (`Verify`, `View Map`, `Delete`).
  - **Tab 2: Log Wildlife Sighting (2-Column Map-Picker Workspace):**
    - **Left Column (Incident Form):**
      - Species selection, Incident Timestamp, Threat Level, Reporter ID, Incident Description.
      - **Multi-Photo Evidence Uploader (`ImageUploadComponent`):** Drag-and-drop zone supporting up to 5 photos, automatic canvas compression for files >1MB, preview thumbnail grid with deletion, and direct upload to Google Cloud Storage (`corald-coexistshield-media`).
    - **Right Column (Interactive Map Pinpoint):**
      - Full-height Google Map: Click anywhere or drag marker to set coordinates.
      - Automatically populates Latitude, Longitude, and reverse-geocodes locality / road name.

---

### 3.4 Screen 4: Roadkill Records & Big-Screen Entry (`/incidents/roadkill`)
**Objective:** Audit historical roadkill incidents, filter out low-credibility noise, and provide a rapid, desktop-optimized entry workflow.

- **Dual-Tab Workspace Layout:**
  - **Tab 1: History & Data Cleansing:**
    - Table displaying Roadkill incidents with **Reliability Score (0 - 100%)** calculated from reporter track record and photo metadata.
    - **Credibility Filter:** Toggle to exclude/hide low-credibility records from AI ingestion (`Show Low-Credibility / Excluded Reports`).
    - **Active Predator Magnets Counter:** Displays live count of carcasses with status `Pending Removal`.
    - **Instant Actions:** Toggle exclusion (`is_excluded = true/false`) and mark disposal (`Pending Removal` -> `Disposed`).
  - **Tab 2: Log Roadkill (2-Column Map-Picker Workspace):**
    - **Left Column (Data Form):**
      - Species identification (Wild Boar, Deer, Small Mammal, Monkey, Cattle, Unknown)
      - Estimated Size / Weight
      - **Carcass Decay Stage:** `Fresh (<6h)`, `Moderate (6-24h)`, `Advanced (>24h)` (critical input for predator attraction modeling)
      - Road Type / Speed Limit
      - **Multi-Photo Evidence Uploader:** Up to 5 forensic photos with client compression and GCS storage.
    - **Right Column (Interactive Map Pinpoint):**
      - Full-height Google Map: Click anywhere to drop pin.
      - Automatically populates Latitude, Longitude, and reverse-geocodes road/locality name.

---

### 3.5 Screen 5: Environmental & Ecological Attractants (`/environmental`)
**Objective:** Manage causal attractant layers that fuel predictive geofencing models with context-aware presets.

- **Dual-Tab Workspace Layout:**
  - **Tab 1: Attractant Layers Registry:**
    - **Regional Overview Cards:** Thailand & APAC Core (Sugarcane & Fruit), Indonesia & India Core (Livestock Corrals), Japan Validation Core (Abandoned Persimmon Trees).
    - **Attractants Table:** Displays Category, Layer Name, Location Description (with center coordinates and radius in meters), Risk Weight (0.1 - 1.0), Active Season / Harvest Window, Last Updated, and Deletion action.
  - **Tab 2: Log Attractant Layer (Category-First Wizard Flow):**
    - **Step 1: Select Ecological Context (Poka-Yoke Design):**
      - Selecting a category preset automatically configures context-aware baseline parameters:
        - `Sugarcane & Fruit Phenology`: Default radius **3,000m**, Risk **0.85**, Harvest season peak, map focus on Kaeng Krachan (Thailand).
        - `Abandoned Persimmon (Kaki) Clusters`: Default radius **500m**, Risk **0.90**, Autumn hyperphagia peak, map focus on Morioka (Japan).
        - `Unfenced Livestock Corrals & Pastures`: Default radius **1,500m**, Risk **0.80**, Year-round nocturnal ambush, map focus on Siak (Sumatra).
        - `Dry-Season Wildlife Watering Holes`: Default radius **1,000m**, Risk **0.70**, Dry season peak.
    - **Step 2: 2-Column Registration Workspace:**
      - **Left Column:** Zone Name, Risk Weight slider, Active Season, Location Description, Radius (meters).
      - **Right Column:** Interactive Google Map with draggable center marker and real-time circle radius overlay preview.

---

### 3.6 Screen 6: Operations & Emergency Tools (`/operations`)
**Objective:** High-level tactical dispatch and data interoperability.

- **Emergency Push Broadcast Console:**
  - **Target Geofence Selector:** Pick an active Tier 3 geofence from dropdown or draw an ad-hoc emergency circle.
  - **Notification Composer:**
    - Alert Title (e.g., `CRITICAL ALERT: Tiger Spotted Near Sector 3`)
    - Alert Body (Clear instructions, e.g., `Evacuate plantation boundary immediately`)
    - Urgency Level: `High (Push Notification)` or `Urgent (Siren & Push)`
  - **Dispatch Summary:** Estimated recipient count, confirmation modal with safety guard.
- **GIS & Data Export Utility:**
  - **Export Dataset:** `Active Geofences`, `Wildlife Incidents`, `Roadkill Records`, `Environmental Layers`
  - **Export Format:**
    - `GeoJSON` (Standard RFC 7946, compatible with QGIS, ArcGIS, Google Earth)
    - `CSV` (Tabular format with WKT geometry for spreadsheet analytics)
  - Date Range picker and Instant Download trigger.

---

## 4. REST API Interface Specification (Go Backend)

Base Path: `/api/v1/admin`

| Endpoint | Method | Description |
| :--- | :--- | :--- |
| `/dashboard/stats` | `GET` | Return summary metrics for KPI cards and threat distribution |
| `/geofences` | `GET` | List active geofences (supports `?source=all\|ai\|manual`) |
| `/geofences` | `POST` | Create a new manual geofence (Circle or Polygon) |
| `/geofences/{id}` | `DELETE` | Delete a specific geofence |
| `/geofences/manual/all` | `DELETE` | Bulk purge all manual geofences |
| `/geofences/recalculate` | `POST` | Trigger asynchronous AI predictive geofence recalculation |
| `/incidents/wildlife` | `GET` | List wildlife sighting & encounter reports with filters |
| `/incidents/wildlife` | `POST` | Create a new wildlife incident report |
| `/incidents/wildlife/{id}` | `DELETE` | Delete a wildlife incident record |
| `/incidents/wildlife/{id}/verify`| `PATCH` | Update verification status (`Verified`, `False Alarm`, `Resolved`) |
| `/incidents/roadkill` | `GET` | List roadkill records (supports filtering low credibility) |
| `/incidents/roadkill` | `POST` | Register new roadkill report via desktop form |
| `/incidents/roadkill/{id}/exclude`| `PATCH` | Toggle exclusion flag (`is_excluded = true/false`) |
| `/incidents/roadkill/{id}/dispose`| `PATCH` | Mark carcass as disposed (`Pending Removal` -> `Disposed`) |
| `/environmental` | `GET` | List all ecological attractant records |
| `/environmental` | `POST` | Register new attractant area |
| `/environmental/{id}` | `DELETE` | Remove an attractant record |
| `/operations/broadcast` | `POST` | Send emergency push broadcast to users in target geofence |
| `/export/{dataset}` | `GET` | Stream GeoJSON or CSV file (`?format=geojson\|csv`) |
| `/upload` | `POST` | Upload field photos to Cloud Storage (`files` field, multipart/form-data, max 5, 5MB each) |

---

## 5. Data Models (JSON Schema)

### 5.1 Geofence Model
```json
{
  "id": "geo-manual-101",
  "name": "Buffer Kaeng Krachan Sector 4",
  "source": "manual",
  "is_persistent": true,
  "geometry_type": "polygon",
  "coordinates": [
    {"lat": 12.885, "lng": 99.645},
    {"lat": 12.890, "lng": 99.660},
    {"lat": 12.875, "lng": 99.670},
    {"lat": 12.870, "lng": 99.650}
  ],
  "circle_center": null,
  "circle_radius_meters": 0,
  "threat_level": "TIER_3_CRITICAL",
  "target_species": "Asian Elephant",
  "created_at": "2026-10-15T06:00:00Z",
  "expires_at": null,
  "reason": "Observed 8-elephant herd advancing toward sugar crops"
}
```

### 5.2 Roadkill Model
```json
{
  "id": "rk-2026-0042",
  "species": "Wild Boar",
  "estimated_weight_kg": 45,
  "decay_stage": "FRESH",
  "location": {
    "lat": 12.854,
    "lng": 99.682,
    "road_name": "Rural Route 3219, KM 14"
  },
  "photo_url": "https://storage.googleapis.com/corald-coexistshield-media/uploads/20261015_053000_a1b2c3d4_rk0042.jpg",
  "photo_urls": [
    "https://storage.googleapis.com/corald-coexistshield-media/uploads/20261015_053000_a1b2c3d4_rk0042.jpg"
  ],
  "reliability_score": 0.92,
  "is_excluded": false,
  "status": "PENDING_REMOVAL",
  "reported_at": "2026-10-15T05:30:00Z"
}
```

### 5.3 Wildlife Incident Model
```json
{
  "id": "inc-wildlife-001",
  "species": "Sumatran Tiger",
  "threat_level": "TIER_3_CRITICAL",
  "location": {
    "lat": 0.515,
    "lng": 101.465,
    "road_name": "Siak Oil Palm Division 3 Boundary"
  },
  "status": "VERIFIED",
  "reporter": "officer_sumatra_04",
  "description": "Adult male tiger observed near communal goat corral. Cattle carcass found nearby.",
  "photo_urls": [
    "https://storage.googleapis.com/corald-coexistshield-media/uploads/20261015_061500_e5f6g7h8_tiger_track.jpg"
  ],
  "reported_at": "2026-10-15T06:15:00Z"
}
```

### 5.4 Environmental Attractant Layer Model
```json
{
  "id": "env-cane-01",
  "category": "CROP_PHENOLOGY",
  "name": "Kaeng Krachan Border Sugarcane Sector A",
  "location_desc": "Sugarcane plantation bordering Kaeng Krachan National Park East Gate",
  "circle_center": {
    "lat": 12.875,
    "lng": 99.655
  },
  "radius_meters": 3000,
  "risk_weight": 0.85,
  "active_season": "November - March (Maturity & Harvest Peak)",
  "photo_urls": [],
  "last_updated": "2026-10-15T06:00:00Z"
}
```

### 5.5 Upload Media Response Model
```json
{
  "urls": [
    "https://storage.googleapis.com/corald-coexistshield-media/uploads/20261015_063000_1a2b3c4d_field_photo_1.jpg",
    "https://storage.googleapis.com/corald-coexistshield-media/uploads/20261015_063000_5e6f7a8b_field_photo_2.jpg"
  ]
}
```
