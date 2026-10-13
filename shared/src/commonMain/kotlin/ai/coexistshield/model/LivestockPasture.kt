package ai.coexistshield.model

import kotlinx.serialization.Serializable

@Serializable
enum class LivestockType {
    CATTLE,
    WATER_BUFFALO,
    GOAT,
    PIG,
    OTHER
}

@Serializable
data class LivestockPasture(
    val id: String,
    val location: GeoLocation,
    val livestockType: LivestockType = LivestockType.CATTLE,
    val headCount: Int = 1,
    val hasNightCorral: Boolean = false,
    val isKillSite: Boolean = false,
    val killReportedAt: String? = null,
    val cachingRiskExpiry: String? = null, // 72-hour decay timestamp
    val activeThreatTier: AlertTier = AlertTier.SAFE
)
