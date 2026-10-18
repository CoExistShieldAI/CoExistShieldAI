import SwiftUI

/// Threat severity tier matching CoExist Shield AI cross-platform classification.
public enum AlertTier: String, Codable, CaseIterable {
    case safe = "SAFE"
    case tier1Notice = "TIER_1_NOTICE"
    case tier2Warning = "TIER_2_WARNING"
    case tier3Danger = "TIER_3_DANGER"

    public var displayName: String {
        switch self {
        case .safe:
            return "SAFE ZONE"
        case .tier1Notice:
            return "TIER 1 NOTICE"
        case .tier2Warning:
            return "TIER 2 WARNING"
        case .tier3Danger:
            return "IMMEDIATE DANGER"
        }
    }

    public var icon: String {
        switch self {
        case .safe:
            return "🛡️"
        case .tier1Notice:
            return "⚠️"
        case .tier2Warning:
            return "⚡"
        case .tier3Danger:
            return "🚨"
        }
    }

    /// Modern tactical HUD color palette aligned with Android/WearOS & Web
    public var tacticalColor: Color {
        switch self {
        case .safe:
            // Emerald 500 (#10B981)
            return Color(red: 16.0 / 255.0, green: 185.0 / 255.0, blue: 129.0 / 255.0)
        case .tier1Notice:
            // Sky 400 (#38BDF8)
            return Color(red: 56.0 / 255.0, green: 189.0 / 255.0, blue: 248.0 / 255.0)
        case .tier2Warning:
            // Amber 500 (#F59E0B)
            return Color(red: 245.0 / 255.0, green: 158.0 / 255.0, blue: 11.0 / 255.0)
        case .tier3Danger:
            // Red 500 (#EF4444)
            return Color(red: 239.0 / 255.0, green: 68.0 / 255.0, blue: 68.0 / 255.0)
        }
    }

    public static func parse(_ raw: String?) -> AlertTier {
        guard let raw = raw?.trimmingCharacters(in: .whitespacesAndNewlines), !raw.isEmpty else {
            return .safe
        }
        return AlertTier(rawValue: raw) ?? .safe
    }
}
