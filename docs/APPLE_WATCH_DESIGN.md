# Apple Watch Companion App & Modern Tactical HUD Integration

**Status:** IMPLEMENTED & VERIFIED  
**Target Platform:** watchOS 9.0+ / iOS 16+ (WatchConnectivity WCSession)  
**Companion Host Apps:** `ionic-app` (Capacitor/Angular & iOS) & `composeApp`  
**Package / Application ID:** `ai.coexistshield.ionic` / `ai.coexistshield.watch`  

---

## 1. Overview & Operational Vision

Field workers (tea/rubber plantation laborers, rural villagers, and wildlife rangers) operating along human-wildlife conflict borders frequently keep smartphones in heavy pockets, backpacks, or vehicle mounts. In acute situations—such as walking downwind toward an active tiger kill-site or entering a dense night elephant corridor—audio sirens may be muffled by diesel tractors, chain machinery, waterfalls, or wind.

The **CoExist Shield Apple Watch Companion Application** delivers immediate, wrist-worn tactile awareness:
1. **High-Precision Modern Tactical Meter:** Renders a clean, circular HUD gauge perfectly optimized for 40mm–49mm Apple Watch displays (Apple Watch Series 4+, SE, and Ultra) with zero visual clutter.
2. **Distinctive Tier-Based Haptics:** Discern threat severity strictly through Apple Watch Taptic Engine vibration patterns without glancing at the display.
3. **Real-time Mobile-to-Watch Synchronization:** Powered by the Apple `WatchConnectivity` framework (`WCSession.updateApplicationContext` and `WCSession.sendMessage`).
4. **Wrist Tactical Quick-Actions:** Instant one-tap acknowledgment (`SILENCE / ACK`) and distress broadcast (`EVACUATE SOS`) directly from the watch face.
5. **Notification Hygiene:** Native heads-up notifications with dynamic threat tier telemetry.

---

## 2. Architecture & Communication Protocol

```mermaid
flowchart TD
    subgraph MobileDevice ["Mobile Host Application (ionic-app iOS)"]
        UI["ShieldStateService / MainHudScreen<br/>(Signals: currentTier, closestDistance, species)"]
        GeoReceiver["GeofenceBridgePlugin / LocationManager<br/>(Background Geofence Incursion Events)"]
        SyncMgr["WearSyncPlugin.swift<br/>(WCSession Bridge)"]
        
        UI -->|effect() Continuous State Sync| SyncMgr
        UI -->|Manual Siren / SOS Trigger| SyncMgr
        GeoReceiver -->|Direct Background ENTER / DWELL| SyncMgr
    end

    subgraph WatchDevice ["Smartwatch (:watchApp - watchOS 9.0+)"]
        Listener["WatchSessionManager<br/>(WCSessionDelegate)"]
        Repo["WatchStateRepository<br/>(ObservableObject singleton)"]
        Haptic["WatchHapticManager<br/>(Taptic Engine Patterns)"]
        Activity["WatchHudView / ModernTacticalMeterView<br/>(SwiftUI View Model)"]
        
        Listener -->|updateApplicationContext /hud_state| Repo
        Listener -->|/alert/trigger sendMessage| Haptic
        Listener -->|Immediate Alert Event| Activity
        Repo --> Activity
        Activity -.->|/alert/dismiss WCSession Message| SyncMgr
        Activity -.->|/alert/sos WCSession Message| SyncMgr
    end

    SyncMgr -->|"① WCSession.updateApplicationContext(/hud_state)<br/>(Persistent latest state sync)"| Listener
    SyncMgr -->|"② WCSession.sendMessage(/alert/trigger)<br/>(Instant wake & haptic alarm)"| Listener
```

---

## 3. Communication Endpoints & Data Payloads

### 3.1 WCSession: Application Context Synchronization (`/hud_state`)
Used for continuous synchronization of the HUD view model. Application contexts are persisted and automatically delivered by watchOS whenever the watch wakes or reconnects:

| Key | Type | Description |
| :--- | :--- | :--- |
| `tier` | `String` | `'SAFE'`, `'TIER_1_NOTICE'`, `'TIER_2_WARNING'`, `'TIER_3_DANGER'` |
| `closest_distance_meters` | `Double` | Distance to nearest threat in meters (or `-1.0` if safe/unknown) |
| `species` | `Array<String>` | List of active target species (e.g. `["TIGER", "ELEPHANT"]`) |
| `is_siren_playing` | `Boolean` | Whether the 95dB acoustic deterrent siren is currently sounding |
| `timestamp` | `Int64` | Epoch millisecond timestamp of last state transition |

### 3.2 WCSession: Urgent Incursion Trigger (`/alert/trigger`)
Used for immediate, sub-50ms low-latency wake-up and Taptic vibration trigger when entering a hazard geofence:
- **Payload Format:** `"${TIER}|${DISTANCE_METERS}|${ALERT_MESSAGE}"` (or Dictionary payload)
- **Example:** `"TIER_3_DANGER|180|Apex predator fresh kill-site detected"`

### 3.3 WCSession: Reverse Wrist Action (`/alert/dismiss` & `/alert/sos`)
Sent from the Apple Watch back to the paired iPhone when the user taps tactical actions on the wrist:
- `/alert/dismiss`: Action `"DISMISS"`. Silences and mutes the iPhone audio siren and acknowledges local alert.
- `/alert/sos`: Action `"SOS"`. Dispatches high-priority emergency evacuation broadcast with GPS coordinates.

---

## 4. Modern Tactical Meter Visual Specifications (SwiftUI)

Implemented in [`ModernTacticalMeterView.swift`](file:///Users/kat/Projects/CoExistShieldAI/watchApp/Sources/WatchAppCore/Views/ModernTacticalMeterView.swift) matching the cross-platform design:

```
┌────────────────────────────────────────────────────────┐
│               ModernTacticalMeter Anatomy              │
├────────────────────────────────────────────────────────┤
│  1. [Ambient Radial Glow]    Threat-tinted blur halo   │
│  2. [Outer Track Ring]       Subtle ring (alpha=0.18)  │
│  3. [Dynamic Proximity Arc]  Round cap, sweeps 0°-360° │
│  4. [Inner Dashed Ring]      Tactical radar dots (3 6) │
│  5. [Central Orb]            Icon, Tier, Distance, m   │
└────────────────────────────────────────────────────────┘
```

1. **Ambient Radial Threat Glow:** Dynamic `RadialGradient` pulsing at 450ms intervals during Tier 3 Danger (`opacity: 0.14 - 0.40`).
2. **Outer Track Ring:** Subtle circular track (`stroke(lineWidth: 7.0 - 9.0)`, `opacity: 0.18`).
3. **Dynamic Proximity Arc:** Fills clockwise from 12 o'clock (`-90°`). In `SAFE` mode, renders a complete 360° closed ring of security. In threat mode, dynamically scales between 20% and 92% of perimeter based on distance (`50m - 2000m`).
4. **Inner Dashed Tactical Ring:** Rendered via `StrokeStyle(lineWidth: 1.8, dash: [3, 6])` at 83% of outer diameter.
5. **Color Standards:**
   - **Safe:** Emerald Green (`#10B981`)
   - **Tier 1 (Notice):** Vibrant Cyan Sky (`#38BDF8`)
   - **Tier 2 (Warning):** Amber Gold (`#F59E0B`)
   - **Tier 3 (Danger):** Crimson Red (`#EF4444`)

---

## 5. Tier-Differentiated Wrist Haptic Patterns (Taptic Engine)

Implemented in [`WatchHapticManager.swift`](file:///Users/kat/Projects/CoExistShieldAI/watchApp/Sources/WatchAppCore/Haptics/WatchHapticManager.swift) using watchOS `WKInterfaceDevice.play(WKHapticType)`:

| Threat Tier | Intended Perception | Taptic Feedback Implementation | Loop Behavior |
| :--- | :--- | :--- | :--- |
| **SAFE** | Subtle confirmation | `WKInterfaceDevice.current().play(.click)` (single short tap) | Single shot |
| **TIER 1 (Notice)** | Informative tap-tap | `play(.directionUp)` followed by second pulse at `150ms` | Single shot (non-intrusive) |
| **TIER 2 (Warning)** | Escalating urgency | 3-pulse ramp: `play(.directionUp)` → `play(.retry)` (+150ms) → `play(.warning)` (+400ms) | Single shot (ramping urgency) |
| **TIER 3 (Danger)** | SOS / Siren evacuation | High-amplitude pulse `play(.failure)` repeated every `450ms` via timer | **Continuous Loop** until silenced or acknowledged |

---

## 6. Directory Structure & Key Components

```
watchApp/
├── Package.swift                                      # Swift Package configuration (macOS/watchOS/iOS)
├── Sources/WatchAppCore/
│   ├── App/
│   │   └── CoExistShieldWatchApp.swift               # @main watchOS App entrypoint
│   ├── Models/
│   │   ├── AlertTier.swift                           # Threat enum, icons, and tactical color palette
│   │   ├── WildlifeSpecies.swift                     # Monitored species (Tiger, Elephant, etc.)
│   │   └── WatchHudState.swift                       # Immutable HUD snapshot & proximity math
│   ├── Repository/
│   │   └── WatchStateRepository.swift                # ObservableObject reactive state holder
│   ├── Haptics/
│   │   └── WatchHapticManager.swift                  # Tier-based Taptic feedback & repeating loop
│   ├── Connectivity/
│   │   └── WatchSessionManager.swift                 # WCSession delegate & bidirectional messaging
│   └── Views/
│       ├── ModernTacticalMeterView.swift              # High-precision circular HUD meter
│       ├── WatchHudView.swift                        # Fullscreen HUD & tactical quick-action overlay
│       └── WatchNotificationView.swift               # Native watchOS hazard notification
└── Tests/WatchAppCoreTests/
    └── WatchAppCoreTests.swift                       # XCTest suite verifying models, math & logic
```

### iOS Host Integration:
- **Plugin Implementation:** [`ionic-app/ios/App/App/Plugins/WearSyncPlugin.swift`](file:///Users/kat/Projects/CoExistShieldAI/ionic-app/ios/App/App/Plugins/WearSyncPlugin.swift)
- **Objective-C Bridge:** [`ionic-app/ios/App/App/Plugins/WearSyncPlugin.m`](file:///Users/kat/Projects/CoExistShieldAI/ionic-app/ios/App/App/Plugins/WearSyncPlugin.m)
- **TypeScript Interface:** [`ionic-app/src/app/plugins/wear-sync.plugin.ts`](file:///Users/kat/Projects/CoExistShieldAI/ionic-app/src/app/plugins/wear-sync.plugin.ts) (Exports unified `WatchSync` and `WearSync`)

---

## 7. Build & Verification Commands

```bash
# 1. Run watchOS / Swift Package Unit Tests
swift test --package-path watchApp

# 2. Build Ionic Mobile App (Web & PWA bundle)
npm --prefix ionic-app run build

# 3. Synchronize Capacitor iOS Native Assets & Plugins
npx --prefix ionic-app cap sync ios

# 4. Open iOS & watchOS Workspace in Xcode
npx --prefix ionic-app cap open ios
```
