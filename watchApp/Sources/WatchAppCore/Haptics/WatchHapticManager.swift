import Foundation
#if os(watchOS)
import WatchKit
#endif

/// Protocol abstracting wearable haptic playback for cross-platform testability.
public protocol HapticDeviceProtocol: AnyObject {
    func playHaptic(type: AlertTier)
    func stop()
}

/// Manages tactile haptics tailored specifically to alert tiers on Apple Watch.
/// Designed so field workers can instantly discern threat severity by touch alone.
public final class WatchHapticManager: HapticDeviceProtocol {
    public static let shared = WatchHapticManager()

    private var loopTimer: Timer?
    private let queue = DispatchQueue(label: "ai.coexistshield.watch.haptics")

    public init() {}

    deinit {
        stop()
    }

    /// Trigger distinctive haptic feedback matching the threat tier.
    public func playHaptic(type: AlertTier) {
        stop()

        switch type {
        case .safe:
            // Subtle confirming tick
            playNativeHaptic(category: .safe)

        case .tier1Notice:
            // Double pulse (Tap-Tap)
            playNativeHaptic(category: .notice)
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.15) { [weak self] in
                self?.playNativeHaptic(category: .notice)
            }

        case .tier2Warning:
            // Escalating 3-pulse ramp
            playNativeHaptic(category: .warningStep1)
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.15) { [weak self] in
                self?.playNativeHaptic(category: .warningStep2)
            }
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.40) { [weak self] in
                self?.playNativeHaptic(category: .warningStep3)
            }

        case .tier3Danger:
            // Aggressive SOS emergency siren pulse looping continuously
            playNativeHaptic(category: .danger)
            DispatchQueue.main.async { [weak self] in
                self?.loopTimer = Timer.scheduledTimer(withTimeInterval: 0.45, repeats: true) { [weak self] _ in
                    self?.playNativeHaptic(category: .danger)
                }
            }
        }
    }

    /// Immediately stop all ongoing vibrations and repeat timers.
    public func stop() {
        DispatchQueue.main.async { [weak self] in
            self?.loopTimer?.invalidate()
            self?.loopTimer = nil
        }
    }

    private enum HapticCategory {
        case safe
        case notice
        case warningStep1
        case warningStep2
        case warningStep3
        case danger
    }

    private func playNativeHaptic(category: HapticCategory) {
        #if os(watchOS)
        let device = WKInterfaceDevice.current()
        switch category {
        case .safe:
            device.play(.click)
        case .notice:
            device.play(.directionUp)
        case .warningStep1:
            device.play(.directionUp)
        case .warningStep2:
            device.play(.retry)
        case .warningStep3:
            device.play(.warning)
        case .danger:
            device.play(.failure)
        }
        #else
        // Mock fallback for unit tests and non-watchOS preview builds
        _ = category
        #endif
    }
}
