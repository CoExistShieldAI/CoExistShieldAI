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

### 1.2 Technology Stack
- **Frontend Framework:** Angular 19+ (Standalone Components, Signals, Reactive Forms, Injectable Services)
- **Map Engine:** Google Maps JavaScript API with `@angular/google-maps` & Drawing Library
- **UI Styling:** Tailwind CSS & Lucide Icons (clean, high-density dashboard UI)
- **Backend:** Go (`backend-cloudrun/internal/admin`), REST API endpoints with structured JSON responses
- **Data Store:** Cloud Firestore & In-Memory Geospatial Spatial Index

```
+-------------------------------------------------------------+
|             Angular 19 Admin Web Portal (English)           |
|  +---------------+  +------------------+  +--------------+  |
|  | Dashboard     |  | JAPAC Map Console|  | Incident Hub |  |
|  | (KPIs/Charts) |  | (Circle/Polygon) |  | (Wildlife/RK)|  |
|  +---------------+  +------------------+  +--------------+  |
+------------------------------+------------------------------+
                               | HTTPS / REST (JSON)
+------------------------------v------------------------------+
|                 Go Backend (backend-cloudrun)                |
|  +------------------+  +------------------+  +-----------+  |
|  | Admin API Router |  | Geofence Service |  | ML Engine |  |
|  | (/api/v1/admin)  |  | (Manual/AI Persist)| (Trigger)  |  |
|  +------------------+  +------------------+  +-----------+  |
+------------------------------+------------------------------+
                               |
              Cloud Firestore / GeoJSON Data Store
```

---

## 2. Navigation & Layout Design

### 2.1 Global Shell Layout
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
- **Manual Geofence Authoring Tools:**
  - **Circle Tool:** Click center point, enter radius in meters (or drag perimeter).
  - **Polygon Tool:** Freeform polygon drawing by clicking coordinate vertices, double-click to close.
  - **Properties Modal on Creation:**
    - Name / Label (e.g., `Emergency Buffer - Kaeng Krachan Sector 4`)
    - Target Species (Asian Elephant, Tiger, Bear, Multi-species)
    - Threat Tier (Tier 1, Tier 2, Tier 3)
    - Reason / Debug Notes
    - Expiration (Optional date/time or Indefinite)
    - Persistence Flag: `is_persistent = true` (Ensures AI auto-refresh sweeps never overwrite or prune this geofence).
- **Bulk Cleanup Action:**
  - `[Clear All Manual Geofences]` button in toolbar.
  - Requires explicit confirmation dialog (`"Are you sure you want to purge all manual geofences? AI-generated fences will remain untouched."`).
- **AI Recalculation Button:**
  - `[Recalculate AI Geofences Now]` executes real-time pipeline inference, refreshes map layers via SSE/WebSocket/polling.

---

### 3.3 Screen 3: Wildlife Sighting & Encounter Hub (`/incidents/wildlife`)
**Objective:** Manage, verify, and filter wildlife sightings and attack incidents.

- **Filter & Search Bar:**
  - Search by Keyword (Location, Description, Reporter ID)
  - Species Filter: `All`, `Asian Elephant`, `Sumatran Tiger`, `Bengal Tiger`, `Brown Bear`, `Asian Black Bear`
  - Severity Level: `Tier 1 (Sighting)`, `Tier 2 (Crop/Property Damage)`, `Tier 3 (Direct Human Threat/Encounter)`
  - Verification Status: `All`, `Unverified`, `Verified`, `Resolved`, `False Alarm`
- **Data Table:**
  - Columns: Timestamp, Species, Location (Lat/Lng & Region), Threat Tier, Verification Status, Reporter, Actions (`Verify`, `View Map`, `Delete`).
- **Create Sighting Modal:**
  - Species dropdown, Incident Timestamp, Coordinates (with mini-map pin selector), Description, Action taken.

---

### 3.4 Screen 4: Roadkill Records & Big-Screen Entry (`/incidents/roadkill`)
**Objective:** Audit historical roadkill incidents, filter out low-credibility noise, and provide a rapid, desktop-optimized entry workflow.

- **Historical Log & Data Cleansing:**
  - Table displaying Roadkill incidents with **Reliability Score (0 - 100%)** calculated from reporter track record and photo metadata.
  - **Credibility Filter:** Toggle to exclude/hide low-credibility records from AI ingestion.
  - **Exclude Toggle Button:** Instant `Exclude from AI / Low Credibility` switch per row.
  - **Carcass Disposal Workflow:** Status tags: `Pending Disposal` (Warning: Predator Magnet) vs `Disposed / Cleared`.
- **Desktop Big-Screen Registration Workspace:**
  - 2-Column Desktop Layout:
    - **Left Column (Data Form):**
      - Species identification (Wild Boar, Deer, Small Mammal, Monkey, Cattle, Unknown)
      - Estimated Size / Weight
      - **Carcass Decay Stage:** `Fresh (<6h)`, `Moderate (6-24h)`, `Advanced (>24h)` (critical input for predator attraction modeling)
      - Road Type / Speed Limit
      - Photo Drag-and-Drop Uploader with instant preview
    - **Right Column (Interactive Map Pinpoint):**
      - Full-height Google Map: Click anywhere to drop a pin.
      - Automatically populates Latitude, Longitude, and reverse-geocodes road/locality name.

---

### 3.5 Screen 5: Environmental & Ecological Attractants (`/environmental`)
**Objective:** Manage causal attractant layers that fuel predictive geofencing models.

- **Attractant Categories:**
  - **Crop Phenology (APAC):** Sugarcane, banana, ripe jackfruit fields with harvest/ripening date ranges.
  - **Abandoned Persimmon Trees (Japan):** High-risk unharvested *kaki* trees in depopulated valleys.
  - **Livestock Corrals & Pastures (APAC):** Open cattle/goat grazing areas attracting apex predators.
- **Table & Management:**
  - List with columns: Category, Name/Location, Risk Weight (0.1 - 1.0), Active Season, Last Updated, Actions (`Edit`, `Delete`).
  - Add Attractant Modal: Geographic polygon or center point, type, phenology calendar, and risk weight.

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
  "photo_url": "https://storage.googleapis.com/coexist-roadkill/rk-0042.jpg",
  "reliability_score": 0.92,
  "is_excluded": false,
  "status": "PENDING_REMOVAL",
  "reported_at": "2026-10-15T05:30:00Z"
}
```
