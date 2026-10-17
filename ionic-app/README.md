# CoExist Shield AI — Angular + Ionic + Capacitor App (`ionic-app/`)

This directory contains the **Angular + Ionic + Capacitor** implementation of the CoExist Shield AI user application, designed for both:
- **Mobile Native App (Android / iOS via Capacitor)** with OS-level background geofencing (`GeofencingClient` / `CLLocationManager`).
- **Web Demo & Hackathon Evaluation (`https://corald.web.app/`)** with a wide-screen 2-pane split layout (**Left: Tactical HUD / Incident Report / Official Console**, **Right: Live Google Maps GIS Canvas**).

> Note: The original Kotlin Multiplatform + Compose Multiplatform (KMP + CMP) web edition is preserved and published at **`https://corald-kmp.web.app/`** (`kmp` hosting target in [`firebase.json`](../firebase.json) and [`.firebaserc`](../.firebaserc)).

---

## Architecture Highlights

1. **Responsive Split-Pane Layout & Foldable Support ([`app.component.ts`](src/app/app.component.ts), [`foldable-device.service.ts`](src/app/services/foldable-device.service.ts))**
   - **Foldable Devices (Pixel Fold, Galaxy Z Fold, Surface Duo)**:
     - **Book Mode (Horizontal Dual-Screen / Spanned)**: Seamless side-by-side split respecting the center hinge seam (`--fold-hinge-width`, `env(viewport-segment-*)`). Left screen hosts Alert HUD / Field Report / Console, right screen hosts the GIS Map Canvas without bottom tabs.
     - **Toggleable Wide Map View (⛶ Wide Map / ◧ Split View)**: Users can toggle between 2-screen Split View (Left: HUD/Report/Console, Right: Map) and Full-Width Wide Map View via the `Wide Map` button on the Left Pane, floating map controls, or demo toolbar.
     - **Tabletop Mode (Vertical Half-Opened / Laptop)**: Top screen displays full interactive GIS Map & Tactical Radar; bottom screen hosts Controls, D-Pad GPS steering, and Report Form.
     - **Cover Screen / Single Screen**: Standard mobile view with bottom tab navigation.
     - **Demo Simulator**: `📱 FOLD` and `VIEW` buttons in the header demo bar allow evaluators to simulate and toggle postures (Auto / Book / Tabletop / Phone) and view modes (Split / Wide Map) instantly without physical foldable hardware.
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
