import Foundation
#if canImport(WatchConnectivity)
import WatchConnectivity
#endif
#if canImport(UserNotifications)
import UserNotifications
#endif

/// Communication manager handling real-time WatchConnectivity between watchOS and iOS.
public final class WatchSessionManager: NSObject {
    public static let shared = WatchSessionManager()

    public static let pathHudState = "/hud_state"
    public static let pathAlertTrigger = "/alert/trigger"
    public static let pathAlertDismiss = "/alert/dismiss"
    public static let pathAlertSos = "/alert/sos"

    private let repository: WatchStateRepository
    private let hapticManager: WatchHapticManager

    public init(
        repository: WatchStateRepository = .shared,
        hapticManager: WatchHapticManager = .shared
    ) {
        self.repository = repository
        self.hapticManager = hapticManager
        super.init()
    }

    /// Activate WCSession if supported on current platform.
    public func startSession() {
        #if canImport(WatchConnectivity)
        if WCSession.isSupported() {
            let session = WCSession.default
            session.delegate = self
            session.activate()
        }
        #endif
    }

    /// Send wrist acknowledgment / silence command to the iPhone companion app.
    public func sendDismissAction(completion: ((Bool) -> Void)? = nil) {
        hapticManager.stop()
        repository.silenceSiren()

        let payload: [String: Any] = [
            "path": Self.pathAlertDismiss,
            "action": "DISMISS",
            "timestamp": Int64(Date().timeIntervalSince1970 * 1000)
        ]

        sendPayloadToPhone(payload, completion: completion)
    }

    /// Send emergency SOS evacuation broadcast request to the iPhone companion app.
    public func sendSosAction(completion: ((Bool) -> Void)? = nil) {
        hapticManager.stop()

        let payload: [String: Any] = [
            "path": Self.pathAlertSos,
            "action": "SOS",
            "timestamp": Int64(Date().timeIntervalSince1970 * 1000)
        ]

        sendPayloadToPhone(payload, completion: completion)
    }

    private func sendPayloadToPhone(_ payload: [String: Any], completion: ((Bool) -> Void)?) {
        #if canImport(WatchConnectivity)
        guard WCSession.isSupported() else {
            completion?(false)
            return
        }

        let session = WCSession.default
        if session.isReachable {
            session.sendMessage(payload, replyHandler: { _ in
                completion?(true)
            }, errorHandler: { _ in
                // Fallback to queued transferUserInfo
                session.transferUserInfo(payload)
                completion?(true)
            })
        } else {
            session.transferUserInfo(payload)
            completion?(true)
        }
        #else
        completion?(false)
        #endif
    }

    /// Process state updates received from mobile host
    public func handleReceivedStatePayload(_ dict: [String: Any]) {
        let previousTier = repository.hudState.tier
        repository.updateFromDictionary(dict)
        let newTier = repository.hudState.tier

        // If tier escalated, trigger wrist haptic
        if newTier != previousTier && newTier != .safe {
            hapticManager.playHaptic(type: newTier)
        }
    }

    /// Process immediate urgent alert received from mobile host
    public func handleReceivedAlertTrigger(payload: String) {
        // Format: "TIER|DISTANCE|MESSAGE"
        let parts = payload.components(separatedBy: "|")
        let tier = AlertTier.parse(parts.first)
        let distanceMeters = (parts.count > 1) ? Double(parts[1]) : nil
        let message = (parts.count > 2) ? parts[2] : "Geofence Hazard Alert Detected"

        // 1. Immediately trigger distinct wrist haptic
        hapticManager.playHaptic(type: tier)

        // 2. Post heads-up watch notification
        postWatchAlertNotification(tier: tier, distanceMeters: distanceMeters, message: message)
    }

    private func postWatchAlertNotification(tier: AlertTier, distanceMeters: Double?, message: String) {
        #if os(watchOS)
        let center = UNUserNotificationCenter.current()
        let content = UNMutableNotificationContent()

        let title: String
        switch tier {
        case .tier3Danger: title = "🚨 IMMEDIATE DANGER"
        case .tier2Warning: title = "⚡ TIER 2 WARNING"
        case .tier1Notice: title = "⚠️ TIER 1 NOTICE"
        case .safe: title = "🛡️ ALL CLEAR"
        }

        let distanceText = distanceMeters.map { " (\(Int($0))m)" } ?? ""
        content.title = title
        content.body = "\(message)\(distanceText)"
        content.sound = .default

        let request = UNNotificationRequest(
            identifier: "ai.coexistshield.watch.hazard.\(Date().timeIntervalSince1970)",
            content: content,
            trigger: nil // Deliver immediately
        )

        center.add(request, withCompletionHandler: nil)
        #else
        // In macOS test runner or non-watchOS host, notifications are mocked or handled by system OS
        _ = tier
        _ = distanceMeters
        _ = message
        #endif
    }
}

#if canImport(WatchConnectivity)
extension WatchSessionManager: WCSessionDelegate {
    public func session(
        _ session: WCSession,
        activationDidCompleteWith activationState: WCSessionActivationState,
        error: Error?
    ) {
        // Session activated
    }

    #if os(iOS)
    public func sessionDidBecomeInactive(_ session: WCSession) {}
    public func sessionDidDeactivate(_ session: WCSession) {
        WCSession.default.activate()
    }
    #endif

    public func session(
        _ session: WCSession,
        didReceiveApplicationContext applicationContext: [String: Any]
    ) {
        handleReceivedStatePayload(applicationContext)
    }

    public func session(
        _ session: WCSession,
        didReceiveUserInfo userInfo: [String: Any] = [:]
    ) {
        handleReceivedStatePayload(userInfo)
    }

    public func session(
        _ session: WCSession,
        didReceiveMessage message: [String: Any],
        replyHandler: @escaping ([String: Any]) -> Void
    ) {
        if let path = message["path"] as? String, path == Self.pathAlertTrigger {
            if let payload = message["payload"] as? String {
                handleReceivedAlertTrigger(payload: payload)
            } else if let tierStr = message["tier"] as? String {
                let dist = message["distanceMeters"] as? Double ?? -1.0
                let msg = message["message"] as? String ?? "Wildlife incursion detected"
                let payload = "\(tierStr)|\(dist)|\(msg)"
                handleReceivedAlertTrigger(payload: payload)
            }
            replyHandler(["success": true])
            return
        }

        // Generic state message
        handleReceivedStatePayload(message)
        replyHandler(["success": true])
    }

    public func session(
        _ session: WCSession,
        didReceiveMessage message: [String: Any]
    ) {
        session(session, didReceiveMessage: message, replyHandler: { _ in })
    }

    public func session(
        _ session: WCSession,
        didReceiveMessageData messageData: Data
    ) {
        if let raw = String(data: messageData, encoding: .utf8) {
            handleReceivedAlertTrigger(payload: raw)
        }
    }
}
#endif
