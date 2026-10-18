<div align="center">

<img src="docs/branding/app_icon_master.svg" alt="CoExist Shield AI logo" width="110" />

# CoExist Shield AI

**Predictive geofencing and early-warning alerts that keep people and wildlife apart before conflict happens.**

*Built on Google Cloud for the AI Builder Cup 2026 (JAPAC)*

[![Cloud Run](https://img.shields.io/badge/Google_Cloud_Run-asia--southeast1-4285F4?logo=googlecloud&logoColor=white)](https://coexistshield-backend-4cerxfo4wq-as.a.run.app/health)
[![Gemini](https://img.shields.io/badge/Gemini-3.8_Flash-8E75B2?logo=googlegemini&logoColor=white)](https://ai.google.dev/gemini-api/docs/models/gemini-3.8-flash)
[![Firebase](https://img.shields.io/badge/Firebase-Hosting_%7C_Firestore-FFCA28?logo=firebase&logoColor=black)](https://corald.web.app)
[![BigQuery](https://img.shields.io/badge/BigQuery-GIS_%2B_ML-669DF6?logo=googlebigquery&logoColor=white)](#architecture)
[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)](backend-cloudrun)
[![Angular](https://img.shields.io/badge/Angular-22-DD0031?logo=angular&logoColor=white)](web-admin)
[![Kotlin](https://img.shields.io/badge/Kotlin-Multiplatform-7F52FF?logo=kotlin&logoColor=white)](shared)
[![Terraform](https://img.shields.io/badge/Terraform-IaC-844FBA?logo=terraform&logoColor=white)](infra)

[**🌐 Field App**](https://corald.web.app) &nbsp;·&nbsp; [**🛠️ Admin Portal**](https://corald-admin.web.app) &nbsp;·&nbsp; [**⚡ Live API**](https://coexistshield-backend-4cerxfo4wq-as.a.run.app/health) &nbsp;·&nbsp; [**🎬 Demo Video**](https://youtu.be/F83EepRVbiQ) &nbsp;·&nbsp; [**📘 Technical Spec**](docs/TECHNICAL_SPEC.md)

</div>

<p align="center">
  <img src="promo-video/public/cinematics/01_tropical_mist.jpg" alt="Rainforest meeting plantation land in Southeast Asia" width="100%" />
</p>

## 📚 Contents

- [Demo Video](#demo-video)
- [The Problem](#the-problem)
- [Our Solution](#our-solution)
- [How It Works](#how-it-works)
- [3-Tier Alerts](#3-tier-alerts)
- [Architecture](#architecture)
- [CI/CD & Infrastructure](#cicd--infrastructure)
- [Tech Stack](#tech-stack)
- [Try It Live](#try-it-live)
- [Quick Start](#quick-start)
- [Repository Structure](#repository-structure)
- [Documentation](#documentation)

---

<a id="demo-video"></a>

## 🎬 Demo Video

<p align="center">
  <a href="https://youtu.be/F83EepRVbiQ">
    <img src="https://img.youtube.com/vi/F83EepRVbiQ/hqdefault.jpg" alt="Watch the CoExist Shield AI demo on YouTube" width="720" />
  </a>
  <br />
  <sub>3 minute walkthrough: the conflict, wrist alerts on Wear OS and Apple Watch, roadkill triage and the JAPAC operations console. Click to play.</sub>
</p>

| Mobile Field App | Operations Admin Portal | End-to-End System |
| :---: | :---: | :---: |
| [▶️ ionic_mobile_demo.mp4](promo-video/public/screen-recordings/ionic_mobile_demo.mp4) | [▶️ admin_demo.mp4](promo-video/public/screen-recordings/admin_demo.mp4) | [▶️ system_demo.mp4](promo-video/public/screen-recordings/system_demo.mp4) |

---

<a id="the-problem"></a>

## 🐾 The Problem

Wild animals do not enter villages at random. They follow food: ripening crops, unguarded livestock, roadkill and abandoned fruit trees. When people and animals meet at these spots, both sides die. Farmers lose crops and lives, and endangered animals are killed in retaliation.

<table>
  <tr>
    <td width="50%" align="center">
      <img src="promo-video/public/cinematics/03_elephant_raid.jpg" alt="Elephant herd raiding a crop field at night" />
      <br /><b>Elephants</b> raid sugarcane, paddy and fruit crops at harvest time<br /><sub>Sri Lanka · India · Thailand</sub>
    </td>
    <td width="50%" align="center">
      <img src="promo-video/public/cinematics/02_tiger_stalk.jpg" alt="Tiger in dense jungle" />
      <br /><b>Tigers</b> follow livestock and return to kill sites for 24 to 72 hours<br /><sub>Sumatra · India</sub>
    </td>
  </tr>
  <tr>
    <td width="50%" align="center">
      <img src="promo-video/public/cinematics/04_roadkill_rain.jpg" alt="Roadkill on a rainy highway attracting scavengers" />
      <br /><b>Roadkill</b> on new highways pulls predators toward roads and settlements<br /><sub>Across JAPAC</sub>
    </td>
    <td width="50%" align="center">
      <img src="promo-video/public/cinematics/05_japan_bear_persimmon.jpg" alt="Black bear eating persimmons in a Japanese village" />
      <br /><b>Bears</b> raid abandoned persimmon trees in depopulated villages<br /><sub>Japan</sub>
    </td>
  </tr>
</table>

---

<a id="our-solution"></a>

## 💡 Our Solution

CoExist Shield AI finds these **attractants** before animals do, predicts where conflict is likely, and warns people in time to stay away.

| | What it does |
| :--- | :--- |
| 🔮 **Predict** | Combines weather, wildfire satellite data, harvest seasons, species records and field reports to score risk on an H3-style hexagon grid |
| 📸 **Verify** | Field workers snap a photo. Gemini 3.8 Flash identifies the species, crop ripeness or carcass decay stage and returns structured JSON |
| 🚨 **Warn** | Dynamic geofences push 3-tier alerts to phones, Wear OS and Apple Watch, with distinct wrist vibrations for each level |
| 🧹 **Act** | Rangers and municipal teams get a priority queue: clear the carcass, harvest the crop, prune the tree, or hold a 72 hour exclusion zone |
| 🔁 **Learn** | Confirmed incidents stream into BigQuery GIS and retrain the risk model on Cloud Run Jobs |

---

<a id="how-it-works"></a>

## ⚙️ How It Works

```mermaid
flowchart LR
    A["Data Ingestion<br/>NASA FIRMS · Open-Meteo<br/>GBIF · Harvest calendars<br/>Field reports"] --> B{"Photo<br/>attached?"}
    B -- Yes --> C["Gemini 3.8 Flash<br/>Species · Ripeness<br/>Carcass decay"]
    B -- No --> D["Environmental model<br/>Wind · Rain · Moon phase"]
    C --> E["Cloud Run Geofence Engine<br/>H3 grid · Risk weighting<br/>Danger polygons"]
    D --> E
    E --> F["3-Tier Alerts<br/>Phone · Wear OS · Apple Watch"]
    F --> G["Field Action<br/>Clear · Harvest · Detour"]
    G --> H[("BigQuery GIS")]
    H -. "Retrain model" .-> E
```

### Example: a tiger kill site

```mermaid
sequenceDiagram
    autonumber
    actor Herder as Herd Owner
    participant App as Mobile App
    participant API as Cloud Run API
    participant AI as Gemini 3.8 Flash
    participant DB as Firestore
    participant Watch as Workers' Watches
    actor Ranger

    Herder->>App: Photo of livestock kill + GPS
    App->>API: Upload report
    API->>AI: Analyse photo
    AI-->>API: Tiger, cached kill, high risk
    API->>DB: Create Tier 3 zone (500 m, 72 h)
    API-->>Watch: Push Tier 3 alert
    Watch-->>Watch: Looping SOS vibration
    API-->>Ranger: Priority dispatch
    Note over DB,Watch: Zone auto-expires after 72 h or when cleared
    Ranger->>App: Site cleared + verification photo
    App->>API: Close incident
    API->>DB: Remove zone, log to BigQuery
```

---

<a id="3-tier-alerts"></a>

## 🚨 3-Tier Alerts

<table>
  <tr>
    <td width="50%">
      <img src="promo-video/public/cinematics/08_wearos_watch_closeup.jpg" alt="Wear OS watch showing a Tier 3 tiger danger alert" />
    </td>
    <td width="50%">

| Tier | Meaning | Wrist feedback |
| :--- | :--- | :--- |
| 🟢 **1 Notice** | Risk nearby, stay alert | Gentle double tap |
| 🟠 **2 Warning** | High chance of encounter, detour | Escalating 3-pulse ramp |
| 🔴 **3 Danger** | Animal or kill site close, leave now | Looping SOS + siren |

Workers can feel the level without looking at a screen. One tap on the watch silences the phone siren so it does not startle a nearby animal.

</td>
  </tr>
</table>

```mermaid
stateDiagram-v2
    direction LR
    [*] --> Safe
    Safe --> Tier1: Enter outer buffer
    Tier1 --> Tier2: Move closer
    Tier2 --> Tier3: Danger zone or kill site
    Tier3 --> Tier2: Move away
    Tier2 --> Tier1: Move away
    Tier1 --> Safe: Leave zone
    Tier3 --> Cleared: Attractant removed
    Tier3 --> Cleared: 72 h TTL expires
    Cleared --> Safe
```

---

<a id="architecture"></a>

## 🏗️ Architecture

```mermaid
flowchart TB
    subgraph Clients["Client Layer"]
        direction LR
        M["Field App<br/>Ionic + Capacitor"]
        K["Field Web / Android<br/>Kotlin Multiplatform"]
        W["Wear OS<br/>Compose for Wear"]
        AW["Apple Watch<br/>SwiftUI"]
        AD["Admin Portal<br/>Angular 22 + Maps JS"]
    end

    subgraph GCP["Google Cloud · asia-southeast1"]
        direction LR
        CR["Cloud Run<br/>Go 1.24 API"]
        GM["Gemini 3.8 Flash<br/>Structured JSON"]
        FS[("Cloud Firestore<br/>3-role RBAC")]
        GCS[("Cloud Storage<br/>Field photos")]
        BQ[("BigQuery GIS<br/>Event lake")]
        JOB["Cloud Run Job<br/>Model retraining"]
        AUTH["Firebase Auth"]
        SM["Secret Manager"]
    end

    subgraph EXT["External Data"]
        direction LR
        FIRMS["NASA FIRMS<br/>Wildfires"]
        METEO["Open-Meteo<br/>Weather"]
        GBIF["GBIF<br/>Species records"]
        OSM["OpenStreetMap<br/>Roads & barriers"]
    end

    W <-->|Data Layer| M
    AW <-->|WCSession| M
    M --> CR
    K --> CR
    AD --> CR
    Clients -.-> AUTH
    CR --> GM
    CR --> FS
    CR --> GCS
    CR --> BQ
    SM -.-> CR
    BQ --> JOB
    JOB -->|Updated weights| CR
    FIRMS --> CR
    METEO --> CR
    GBIF --> CR
    OSM --> CR
```

---

<a id="cicd--infrastructure"></a>

## 🔄 CI/CD & Infrastructure

Everything ships through GitHub Actions. Infrastructure is defined in Terraform and secrets live in Secret Manager, never in the repo.

```mermaid
flowchart LR
    DEV["Push / Pull Request"] --> CI

    subgraph CI["ci.yml"]
        direction TB
        T1["Go tests<br/>race detector"]
        T2["KMP JVM tests"]
        T3["Web bundle +<br/>Playwright E2E"]
        T4["Android APK build"]
    end

    CI -->|merge to main| CD

    subgraph CD["Deploy"]
        direction TB
        D1["deploy-cloudrun.yml<br/>Go API to Cloud Run"]
        D2["deploy-hosting.yml<br/>Field + Admin to Firebase Hosting"]
        D3["infra.yml<br/>Terraform plan on PR, apply on main"]
    end

    D1 --> RUN["Cloud Run<br/>Singapore"]
    D2 --> HOST["corald.web.app<br/>corald-admin.web.app"]
    D3 --> TF["Firebase Auth · API keys<br/>Identity Platform"]
    SEC["Secret Manager<br/>Gemini · Maps · FIRMS keys"] -.-> D1
```

| Workflow | Trigger | What it does |
| :--- | :--- | :--- |
| `ci.yml` | Push and PR to `main` | Go tests with race detection, KMP tests, Android build, web bundle and Playwright E2E |
| `deploy-cloudrun.yml` | Backend change on `main` / manual | Builds the distroless Go image and deploys to Cloud Run with secrets from Secret Manager |
| `deploy-hosting.yml` | Merge to `main` | Builds and deploys the field app and admin portal to Firebase Hosting, then smoke tests the URLs |
| `infra.yml` | PR (plan) / `main` (apply) | Terraform for Firebase Auth, API keys and Identity Platform |

---

<a id="tech-stack"></a>

## 🧰 Tech Stack

| Layer | Technologies |
| :--- | :--- |
| **AI** | Gemini 3.8 Flash (multimodal, structured JSON outputs), BigQuery GIS + ML retraining |
| **Backend** | Go 1.24 on Cloud Run (distroless container), H3-style hexagonal grid (Res 8) |
| **Data** | Cloud Firestore, Cloud Storage, BigQuery |
| **Field apps** | Ionic + Capacitor, Kotlin Multiplatform + Compose (Android and Web) |
| **Wearables** | Wear OS (Compose for Wear, Data Layer API), Apple Watch (SwiftUI, WatchConnectivity) |
| **Admin** | Angular 22, Google Maps JavaScript API |
| **DevOps** | GitHub Actions, Terraform, Firebase Hosting, Secret Manager |
| **Testing** | Go race tests, Playwright E2E, Karma, Swift XCTest, Firestore rules emulator tests |

### Benchmarks

| Metric | Result |
| :--- | :--- |
| In-memory geofence evaluation (Go benchmark) | **289 ns/op**, zero allocations |
| Live Cloud Run round trip, Japan to Singapore | **166 to 181 ms** TTFB |
| Gemini response parsing and validation | **3.48 µs/op** |
| Container image | **~4.6 MB** distroless |

---

<a id="try-it-live"></a>

## 🚀 Try It Live

| Surface | URL | What to try |
| :--- | :--- | :--- |
| 🌐 **Field Web App** | [corald.web.app](https://corald.web.app) | 3-tier geofence HUD, hazard map, Gemini photo triage |
| 🛠️ **Operations Admin Portal** | [corald-admin.web.app](https://corald-admin.web.app) | JAPAC map console, manual geofences, roadkill registry |
| ⚡ **Cloud Run API** | [/health](https://coexistshield-backend-4cerxfo4wq-as.a.run.app/health) | [Active geofences as GeoJSON](https://coexistshield-backend-4cerxfo4wq-as.a.run.app/api/v1/geofences/active?latitude=39.7&longitude=141.1&radiusKm=5000) |
| 🎬 **Demo Video** | [youtu.be/F83EepRVbiQ](https://youtu.be/F83EepRVbiQ) | 3 minute product walkthrough |

More endpoints are listed in the [Technical Spec](docs/TECHNICAL_SPEC.md#99-live-public-cloud-run-endpoints-singapore-asia-southeast1).

---

<a id="quick-start"></a>

## ⚡ Quick Start

**Prerequisites:** Go 1.24+, Node.js 20+, JDK 21, Android SDK 35 (for mobile and Wear OS), Xcode (for Apple Watch).

```bash
git clone https://github.com/CoExistShieldAI/CoExistShieldAI.git
cd CoExistShieldAI
```

<details>
<summary><b>Backend (Go / Cloud Run)</b></summary>

```bash
cd backend-cloudrun
go test -race ./...
PORT=8080 GEMINI_API_KEY="your-gemini-key" go run ./cmd/server
```
</details>

<details>
<summary><b>Operations Admin Portal (Angular 22)</b></summary>

```bash
cd web-admin
npm install
npm start          # http://localhost:4200
```
</details>

<details>
<summary><b>Mobile Field App (Ionic + Capacitor)</b></summary>

```bash
cd ionic-app
npm install
npm run build
npm run cap:sync
```
</details>

<details>
<summary><b>Kotlin Multiplatform, Android and Wear OS</b></summary>

```bash
export JAVA_HOME=$(/usr/libexec/java_home -v 21)
./gradlew :shared:jvmTest
./gradlew :shared:jsBrowserDistribution
./gradlew :composeApp:assembleDebug
./gradlew :wearApp:assembleDebug
```
</details>

<details>
<summary><b>Apple Watch (watchOS)</b></summary>

```bash
swift test --package-path watchApp
cd ionic-app && npx cap sync ios && npx cap open ios
```
</details>

Full build, test and deployment steps are in the [Technical Spec](docs/TECHNICAL_SPEC.md#9-local-setup--build-guide).

---

<a id="repository-structure"></a>

## 📁 Repository Structure

CoExistShieldAI/
├── backend-cloudrun/ Go Cloud Run API: geofence engine, Gemini client, ML, connectors
├── web-admin/ Angular 22 operations portal
├── ionic-app/ Ionic + Capacitor field app with native geofencing and watch sync
├── shared/ Kotlin Multiplatform models, API client and Compose UI
├── composeApp/ Android app
├── wearApp/ Wear OS companion app
├── watchApp/ Apple Watch companion app (Swift Package)
├── infra/terraform/ Infrastructure as Code
├── firebase-tests/ Firestore security rules tests
├── web-e2e/ Playwright end-to-end tests
├── promo-video/ Demo video source and assets
├── scripts/ Deploy and config helper scripts
├── docs/ Technical spec and design documents
└── .github/workflows/ CI and deployment pipelines


---

<a id="documentation"></a>

## 📖 Documentation

| Document | Covers |
| :--- | :--- |
| [Technical Specification](docs/TECHNICAL_SPEC.md) | Full product vision, ecology, architecture, business model, setup and endpoints |
| [Data Model](docs/DATA_MODEL.md) | Firestore collections and domain schemas |
| [Wear OS Design](docs/WEAR_OS_DESIGN.md) | Wear OS app, haptic waveforms and Data Layer protocol |
| [Apple Watch Design](docs/APPLE_WATCH_DESIGN.md) | Apple Watch app, Taptic patterns and WCSession protocol |
| [Admin Portal Design](docs/ADMIN_EXTERNAL_DESIGN.md) | Operations portal design |
| [Implementation Guidelines](docs/IMPLEMENTATION_GUIDELINES.md) | Development workflow and branch rules |

---

<p align="center"><sub>Built for the Google Cloud AI Builder Cup 2026 · Build in JAPAC. Build for the world.</sub></p>
