import Foundation

/// Immutable snapshot of the HUD state synchronized from the companion iOS app.
public struct WatchHudState: Equatable, Codable {
    public var tier: AlertTier
    public var closestDistanceMeters: Double?
    public var targetSpecies: [WildlifeSpecies]
    public var isSirenPlaying: Bool
    public var lastUpdatedMillis: Int64

    public init(
        tier: AlertTier = .safe,
        closestDistanceMeters: Double? = nil,
        targetSpecies: [WildlifeSpecies] = [],
        isSirenPlaying: Bool = false,
        lastUpdatedMillis: Int64 = Int64(Date().timeIntervalSince1970 * 1000)
    ) {
        self.tier = tier
        self.closestDistanceMeters = closestDistanceMeters
        self.targetSpecies = targetSpecies
        self.isSirenPlaying = isSirenPlaying
        self.lastUpdatedMillis = lastUpdatedMillis
    }

    /// Calculate the progress arc sweep ratio (0.0 to 1.0)
    /// SAFE: 1.0 (Full 360-degree closed ring)
    /// Threat: Distance 50m to 2000m maps to 0.92 to 0.20
    public var proximityFillRatio: Double {
        if tier == .safe {
            return 1.0
        }
        guard let distance = closestDistanceMeters else {
            return 0.40
        }
        let clamped = max(50.0, min(2000.0, distance))
        let ratio = (2000.0 - clamped) / 1950.0
        return max(0.20, min(0.92, ratio))
    }

    /// Parses dictionary payload matching Wear OS `/hud_state` data layer specifications.
    public static func from(dictionary: [String: Any]) -> WatchHudState {
        let tierStr = dictionary["tier"] as? String
        let tier = AlertTier.parse(tierStr)

        let rawDist = dictionary["closest_distance_meters"] as? Double
            ?? (dictionary["distanceMeters"] as? Double)
            ?? -1.0
        let distanceMeters = (rawDist >= 0.0) ? rawDist : nil

        var speciesList: [WildlifeSpecies] = []
        if let rawArray = dictionary["species"] as? [String] {
            speciesList = rawArray.compactMap { WildlifeSpecies.parse($0) }
        }

        let isSiren = (dictionary["is_siren_playing"] as? Bool)
            ?? (dictionary["isSiren"] as? Bool)
            ?? false

        let timestamp = (dictionary["timestamp"] as? Int64)
            ?? (dictionary["timestamp"] as? NSNumber)?.int64Value
            ?? Int64(Date().timeIntervalSince1970 * 1000)

        return WatchHudState(
            tier: tier,
            closestDistanceMeters: distanceMeters,
            targetSpecies: speciesList,
            isSirenPlaying: isSiren,
            lastUpdatedMillis: timestamp
        )
    }

    /// Serializes to dictionary for WCSession transmission.
    public func toDictionary() -> [String: Any] {
        return [
            "tier": tier.rawValue,
            "closest_distance_meters": closestDistanceMeters ?? -1.0,
            "species": targetSpecies.map { $0.rawValue },
            "is_siren_playing": isSirenPlaying,
            "timestamp": lastUpdatedMillis
        ]
    }
}
