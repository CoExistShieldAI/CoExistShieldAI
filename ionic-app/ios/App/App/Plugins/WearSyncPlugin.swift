import Foundation
import Capacitor
import WatchConnectivity

/**
 * Capacitor Plugin bridging Angular/Ionic state with Apple Watch (watchOS) smartwatches
 * via Apple WatchConnectivity framework (WCSession).
 * Matches the Android WearSyncPlugin specifications for cross-platform companion symmetry.
 */
@objc(WearSyncPlugin)
public class WearSyncPlugin: CAPPlugin, CAPBridgedPlugin, WCSessionDelegate {
    public let identifier = "WearSyncPlugin"
    public let jsName = "WearSync"
    public let pluginMethods: [CAPPluginMethod] = [
        CAPPluginMethod(name: "syncHudState", returnType: CAPPluginReturnPromise),
        CAPPluginMethod(name: "sendImmediateAlert", returnType: CAPPluginReturnPromise)
    ]

    public static weak var shared: WearSyncPlugin?

    public static let pathHudState = "/hud_state"
    public static let pathAlertTrigger = "/alert/trigger"
    public static let pathAlertDismiss = "/alert/dismiss"
    public static let pathAlertSos = "/alert/sos"

    override public func load() {
        super.load()
        WearSyncPlugin.shared = self
        setupWatchSession()
    }

    private func setupWatchSession() {
        guard WCSession.isSupported() else { return }
        let session = WCSession.default
        session.delegate = self
        session.activate()
    }

    /**
     * Static helper allowing iOS Background Location / Geofence delegates to broadcast
     * urgent alerts directly to the Apple Watch even if the Capacitor WebView is backgrounded.
     */
    @objc public static func broadcastUrgentAlertToWatch(tier: String, distanceMeters: Double, message: String) {
        guard WCSession.isSupported() else { return }
        let session = WCSession.default
        guard session.activationState == .activated, session.isPaired, session.isWatchAppInstalled else { return }

        let payload: [String: Any] = [
            "path": pathAlertTrigger,
            "payload": "\(tier)|\(distanceMeters)|\(message)",
            "tier": tier,
            "distanceMeters": distanceMeters,
            "message": message,
            "timestamp": Int64(Date().timeIntervalSince1970 * 1000)
        ]

        if session.isReachable {
            session.sendMessage(payload, replyHandler: nil) { _ in
                session.transferUserInfo(payload)
            }
        } else {
            session.transferUserInfo(payload)
        }
    }

    /**
     * Synchronizes current HUD state from Angular/Tailwind to connected Apple Watch.
     */
    @objc func syncHudState(_ call: CAPPluginCall) {
        let tier = call.getString("tier") ?? "SAFE"
        let distance = call.getDouble("distanceMeters") ?? -1.0
        let speciesArray = call.getArray("species") ?? []
        let isSiren = call.getBool("isSiren") ?? false

        guard WCSession.isSupported() else {
            call.resolve([
                "success": true,
                "tier": tier,
                "note": "WatchConnectivity not supported on this device"
            ])
            return
        }

        let session = WCSession.default
        let payload: [String: Any] = [
            "tier": tier,
            "closest_distance_meters": distance,
            "species": speciesArray,
            "is_siren_playing": isSiren,
            "timestamp": Int64(Date().timeIntervalSince1970 * 1000)
        ]

        do {
            try session.updateApplicationContext(payload)
        } catch {
            // Context update may fail if session not yet activated; fallback to transferUserInfo
            session.transferUserInfo(payload)
        }

        // If watch is actively awake on wrist, also send via low-latency interactive message
        if session.isReachable {
            session.sendMessage(payload, replyHandler: nil, errorHandler: nil)
        }

        call.resolve([
            "success": true,
            "tier": tier
        ])
    }

    /**
     * Sends an immediate, high-priority alert message causing the watch to vibrate instantly.
     */
    @objc func sendImmediateAlert(_ call: CAPPluginCall) {
        let tier = call.getString("tier") ?? "TIER_2_WARNING"
        let distance = call.getDouble("distanceMeters") ?? -1.0
        let message = call.getString("message") ?? "Wildlife threat alert"

        guard WCSession.isSupported() else {
            call.resolve([
                "success": true,
                "nodesCount": 0
            ])
            return
        }

        let session = WCSession.default
        let payload: [String: Any] = [
            "path": WearSyncPlugin.pathAlertTrigger,
            "payload": "\(tier)|\(distance)|\(message)",
            "tier": tier,
            "distanceMeters": distance,
            "message": message,
            "timestamp": Int64(Date().timeIntervalSince1970 * 1000)
        ]

        if session.isReachable {
            session.sendMessage(payload, replyHandler: { _ in
                call.resolve([
                    "success": true,
                    "nodesCount": 1
                ])
            }, errorHandler: { _ in
                session.transferUserInfo(payload)
                call.resolve([
                    "success": true,
                    "nodesCount": 1
                ])
            })
        } else {
            session.transferUserInfo(payload)
            call.resolve([
                "success": true,
                "nodesCount": (session.isPaired && session.isWatchAppInstalled) ? 1 : 0
            ])
        }
    }

    // MARK: - WCSessionDelegate

    public func session(_ session: WCSession, activationDidCompleteWith activationState: WCSessionActivationState, error: Error?) {
        // Ready to communicate
    }

    public func sessionDidBecomeInactive(_ session: WCSession) {}

    public func sessionDidDeactivate(_ session: WCSession) {
        WCSession.default.activate()
    }

    /**
     * Listens for reverse commands sent from the Apple Watch
     * (e.g. user tapped "SILENCE / ACK" or "EVACUATE SOS" on wrist).
     */
    public func session(_ session: WCSession, didReceiveMessage message: [String : Any]) {
        handleIncomingWatchAction(message)
    }

    public func session(_ session: WCSession, didReceiveUserInfo userInfo: [String : Any] = [:]) {
        handleIncomingWatchAction(userInfo)
    }

    private func handleIncomingWatchAction(_ data: [String: Any]) {
        let action = data["action"] as? String
        let path = data["path"] as? String ?? ""

        if action == "DISMISS" || path.hasSuffix("dismiss") || path == WearSyncPlugin.pathAlertDismiss {
            notifyListeners("onWatchAction", data: [
                "action": "DISMISS",
                "timestamp": Int64(Date().timeIntervalSince1970 * 1000)
            ])
        } else if action == "SOS" || path.hasSuffix("sos") || path == WearSyncPlugin.pathAlertSos {
            notifyListeners("onWatchAction", data: [
                "action": "SOS",
                "timestamp": Int64(Date().timeIntervalSince1970 * 1000)
            ])
        }
    }
}
