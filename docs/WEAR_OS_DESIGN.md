# Wear OS Companion App & Modern Tactical HUD Integration

**Status:** IMPLEMENTED & VERIFIED  
**Target Platform:** Wear OS 3.0+ / Android 14+ (Wear API 30+)  
**Companion Host Apps:** `ionic-app` (Capacitor/Angular) & `composeApp` (Compose Multiplatform)  
**Package / Application ID:** `ai.coexistshield.ionic` / `ai.coexistshield`  

> 💡 **Related Architecture:** For the companion Apple Watch (watchOS) implementation and WCSession protocol, see [APPLE_WATCH_DESIGN.md](file:///Users/kat/Projects/CoExistShieldAI/docs/APPLE_WATCH_DESIGN.md).

---

## 1. Overview & Operational Vision

Field workers (plantation laborers, rural villagers, wildlife rangers) operating in wildlife corridors frequently keep smartphones in pockets, backpacks, or vehicle mounts. In acute situations—such as walking downwind toward an active tiger kill-site or entering a night elephant crop-raiding corridor—audio alerts may be muffled by engine noise, machinery, or dense vegetation.

The **CoExist Shield Wear OS Companion Application** delivers immediate, wrist-worn tactile awareness:
1. **High-Precision Modern Tactical Meter:** Renders a clean, circular HUD gauge designed specifically for round smartwatch displays with zero visual clutter.
2. **Distinctive Tier-Based Haptics:** Discern threat severity strictly through tactile vibration patterns on the wrist without needing to glance at a screen.
3. **Real-time Mobile-to-Watch Synchronization:** Powered by Google Play Services Wearable Data Layer API (`DataClient` and `MessageClient`).
4. **Wrist Tactical Quick-Actions:** Instant one-tap acknowledgment (`SILENCE / ACK`) and distress broadcast (`EVACUATE SOS`) directly from the watch face.
5. **Notification Hygiene:** Automatic suppression of duplicate Android notification mirroring (`setLocalOnly(true)`).

---

## 2. Architecture & Communication Protocol

```mermaid
flowchart TD
    subgraph MobileDevice ["Mobile Host Application (ionic-app / composeApp)"]
        UI["ShieldStateService / MainHudScreen<br/>(Signals: currentTier, closestDistance, species)"]
        GeoReceiver["GeofenceBroadcastReceiver<br/>(OS-level GeofencingClient Transitions)"]
        SyncMgr["WearSyncPlugin / WearDataSyncManager<br/>(Wearable Data Layer Bridge)"]
        
        UI -->|effect() Continuous State Sync| SyncMgr
        UI -->|Manual Siren / SOS Trigger| SyncMgr
        GeoReceiver -->|Direct Background ENTER / DWELL| SyncMgr
    end

    subgraph WearDevice ["Smartwatch (:wearApp - Wear OS 3.0+)"]
        Listener["WearDataListenerService<br/>(WearableListenerService)"]
        Repo["WearStateRepository<br/>(StateFlow singleton)"]
        Haptic["WatchHapticManager<br/>(VibrationEffect Waveforms)"]
        Activity["WearMainActivity / WearHudScreen<br/>(ModernTacticalMeter Composable)"]
        
        Listener -->|/hud_state DataItem| Repo
        Listener -->|/alert/trigger Message| Haptic
        Listener -->|Immediate Alert Event| Activity
        Repo --> Activity
        Activity -.->|/alert/dismiss Reverse Message| SyncMgr
    end

    SyncMgr -->|"① DataClient.putDataItem(/hud_state)<br/>(Cached persistent state)"| Listener
    SyncMgr -->|"② MessageClient.sendMessage(/alert/trigger)<br/>(Instant wake & haptics)"| Listener
```

---

## 3. Communication Endpoints & Data Payloads

### 3.1 DataClient: State Synchronization (`/hud_state`)
Used for continuous synchronization of the HUD view model. Data items are persisted and automatically delivered even if the watch was temporarily disconnected or sleeping.

| Key | Type | Description |
| :--- | :--- | :--- |
| `tier` | `String` | `'SAFE'`, `'TIER_1_NOTICE'`, `'TIER_2_WARNING'`, `'TIER_3_DANGER'` |
| `closest_distance_meters` | `Double` | Distance to nearest threat in meters (or `-1.0` if safe/unknown) |
| `species` | `ArrayList<String>` | List of active target species (e.g. `["TIGER", "ELEPHANT"]`) |
| `is_siren_playing` | `Boolean` | Whether the 95dB acoustic deterrent siren is currently sounding |
| `timestamp` | `Long` | Epoch millisecond timestamp of last state transition |

### 3.2 MessageClient: Urgent Incursion Trigger (`/alert/trigger`)
Used for immediate, low-latency wake-up and haptic vibration trigger when entering a geofence.
- **Payload Format:** `"${TIER}|${DISTANCE_METERS}|${ALERT_MESSAGE}"` (UTF-8 encoded string)
- **Example:** `"TIER_3_DANGER|180|Apex predator fresh kill-site detected"`

### 3.3 MessageClient: Reverse Wrist Action (`/alert/dismiss` & `/alert/sos`)
Sent from the watch back to the paired smartphone when the user taps tactical actions on the wrist:
- `/alert/dismiss`: Mutes and stops the smartphone audio siren and acknowledges local alert.
- `/alert/sos`: Triggers high-priority emergency SOS broadcast with GPS coordinates to municipal rangers.

---

## 4. Modern Tactical Meter Visual Specifications

Ported from the SVG implementation in `ionic-app` (`hud-panel.component.ts`) into the unified Compose Multiplatform component [`ModernTacticalMeter.kt`](file:///Users/kat/Projects/CoExistShieldAI/shared/src/commonMain/kotlin/ai/coexistshield/ui/hud/ModernTacticalMeter.kt):

```
┌────────────────────────────────────────────────────────┐
│               ModernTacticalMeter Anatomy              │
├────────────────────────────────────────────────────────┤
│  1. [Ambient Radial Glow]    Threat-tinted blur halo   │
│  2. [Outer Track Ring]       Subtle ring (alpha=0.18)  │
│  3. [Dynamic Proximity Arc]  Round cap, sweeps 0°-360° │
│  4. [Inner Dashed Ring]      Tactical radar dots (4 8) │
│  5. [Central Orb]            Icon, Tier, Distance, m   │
└────────────────────────────────────────────────────────┘
```

1. **Ambient Radial Glow:** Dynamic radial gradient pulsing at 450ms intervals during Tier 3 Danger (`alpha: 0.14 - 0.40`).
2. **Outer Track Ring:** Subtle circular track (`strokeWidth: 8-10dp`, `opacity: 0.18`).
3. **Dynamic Proximity Arc:** Fills clockwise from 12 o'clock (`-90°`). In `SAFE` mode, renders a complete 360° closed ring of security. In threat mode, dynamically scales between 20% and 92% of perimeter based on distance (`50m - 2000m`).
4. **Inner Dashed Tactical Ring:** Rendered via `PathEffect.dashPathEffect(floatArrayOf(3dp, 6dp), 0f)` at 83% of the outer radius, evoking military radar telemetry.
5. **Color Standards:**
   - **Safe:** Emerald Green (`#10B981`)
   - **Tier 1 (Notice):** Vibrant Cyan Sky (`#38BDF8`)
   - **Tier 2 (Warning):** Amber Gold (`#F59E0B`)
   - **Tier 3 (Danger):** Crimson Red (`#EF4444`)

---

## 5. Tier-Differentiated Wrist Haptic Patterns

Implemented in [`WatchHapticManager.kt`](file:///Users/kat/Projects/CoExistShieldAI/wearApp/src/androidMain/kotlin/ai/coexistshield/wear/WatchHapticManager.kt) using Android's `VibrationEffect.createWaveform()`:

| Threat Tier | Intended Perception | Vibration Waveform Pattern (Timings & Amplitudes) | Loop Behavior |
| :--- | :--- | :--- | :--- |
| **SAFE** | Subtle confirmation | `50ms` tick at standard amplitude | Single shot |
| **TIER 1 (Notice)** | Informative tap-tap | `[0, 150, 150, 150]ms` @ amplitude 130 | Single shot (non-intrusive) |
| **TIER 2 (Warning)** | Escalating urgency | `[0, 150, 100, 250, 100, 420]ms` @ amplitudes `[0, 150, 0, 200, 0, 255]` | Single shot (ramping intensity) |
| **TIER 3 (Danger)** | SOS / Siren evacuation | `[0, 400, 120, 400, 120, 400, 350]ms` @ maximum amplitude 255 | **Continuous Loop** until user acknowledges or silences |

---

## 6. Duplicate Notification Suppression Strategy

When a mobile app and companion watch app are both active, standard Android notification mirroring would cause notifications to ring twice on the user's wrist.

**Solution:**
- Mobile notifications in [`WearDataSyncManager.kt`](file:///Users/kat/Projects/CoExistShieldAI/composeApp/src/androidMain/kotlin/ai/coexistshield/wear/WearDataSyncManager.kt#L104) are posted with:
  ```kotlin
  builder.setLocalOnly(true)
  ```
- The Wear OS watch receives the alert via [`WearDataListenerService.kt`](file:///Users/kat/Projects/CoExistShieldAI/wearApp/src/androidMain/kotlin/ai/coexistshield/wear/WearDataListenerService.kt) and posts its own dedicated, wearable-optimized Heads-Up Notification that launches directly into `WearMainActivity`.

---

## 7. Build & Verification Commands

```bash
# Build Wear OS Companion App Debug APK
./gradlew :wearApp:assembleDebug

# Build Mobile Compose Multiplatform Debug APK
./gradlew :composeApp:assembleDebug

# Run KMP Unit Tests
./gradlew :shared:jvmTest

# Build and test Ionic / Capacitor Mobile App
npm --prefix ionic-app run build
npm --prefix ionic-app test -- --watch=false --browsers=ChromeHeadless
```
