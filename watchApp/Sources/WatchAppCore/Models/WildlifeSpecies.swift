import Foundation

/// Target wildlife species monitored in human-wildlife conflict zones.
public enum WildlifeSpecies: String, Codable, CaseIterable {
    case tiger = "TIGER"
    case elephant = "ELEPHANT"
    case leopard = "LEOPARD"
    case bear = "BEAR"
    case boar = "BOAR"
    case other = "OTHER"

    public var displayName: String {
        switch self {
        case .tiger: return "Tiger"
        case .elephant: return "Elephant"
        case .leopard: return "Leopard"
        case .bear: return "Bear"
        case .boar: return "Wild Boar"
        case .other: return "Wildlife"
        }
    }

    public static func parse(_ raw: String?) -> WildlifeSpecies? {
        guard let raw = raw?.trimmingCharacters(in: .whitespacesAndNewlines).uppercased(), !raw.isEmpty else {
            return nil
        }
        return WildlifeSpecies(rawValue: raw)
    }
}
