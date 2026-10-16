# CoExist Shield AI — Angular + Ionic + Capacitor App (`ionic-app/`)

This directory contains the **Angular + Ionic + Capacitor** implementation of the CoExist Shield AI user application, designed for both:
- **Mobile Native App (Android / iOS via Capacitor)** with OS-level background geofencing (`GeofencingClient` / `CLLocationManager`).
- **Web Demo & Hackathon Evaluation (`https://corald.web.app/`)** with a wide-screen 2-pane split layout (**Left: Tactical HUD / Incident Report / Official Console**, **Right: Live Google Maps GIS Canvas**).

> Note: The original Kotlin Multiplatform + Compose Multiplatform (KMP + CMP) web edition is preserved and published at **`https://corald-kmp.web.app/`** (`kmp` hosting target in [`firebase.json`](../firebase.json) and [`.firebaserc`](../.firebaserc)).

---

## Architecture Highlights

1. **Responsive Split-Pane Layout ([`app.component.ts`](src/app/app.component.ts))**
   - **Desktop / Tablet Landscape (`lg+`)**: Left pane (`440px–470px`) continuously displays the Tactical HUD, Forensic Incident Report, or Official Decision-Support Console, while the right pane renders the full-screen interactive Google Maps (`@angular/google-maps`) canvas.
   - **Mobile Portrait (`< lg`)**: Native-style bottom navigation bar (`Alert HUD`, `GIS Map`, `Report`, `Console`) with safe-area support.

2. **Hybrid Capacitor Geofence Bridge ([`geofence-bridge.plugin.ts`](src/app/plugins/geofence-bridge.plugin.ts))**
   - **Prioritized OS Quota Allocation**: Ranks candidate geofences using the threat-weighted distance decay formula (`Score = TierWeight / (1 + DistanceKm)`) to stay within strict OS limits (20 active zones).
   - **Android Native Layer ([`GeofenceBridgePlugin.kt`](android/app/src/main/java/ai/coexistshield/ionic/GeofenceBridgePlugin.kt) & [`GeofenceBroadcastReceiver.kt`](android/app/src/main/java/ai/coexistshield/ionic/GeofenceBroadcastReceiver.kt))**: Registers circular/polygon hazard regions with Google Play Services `GeofencingClient` for background `ENTER`, `DWELL`, and `EXIT` transitions.
   - **Web & Judge Simulation Layer**: Executes real-time Haversine + ray-casting point-in-polygon checks in the browser. Clicking anywhere on the Google Map (when `Click Map = Move GPS` is active) teleports the virtual device location and triggers immediate HUD and geofence transitions.

---

## Development & Build Commands

```bash
cd ionic-app

# Install dependencies
npm install --legacy-peer-deps

# Start local dev server (http://localhost:4201)
npm start

# Production web build + Google Maps / Firebase API key injection
GOOGLE_MAPS_API_KEY="your-key" npm run build

# Sync web assets to Capacitor Android project
npm run cap:sync
```
