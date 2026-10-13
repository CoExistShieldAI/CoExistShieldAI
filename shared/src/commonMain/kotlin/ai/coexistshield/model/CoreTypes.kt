package ai.coexistshield.model

import kotlinx.serialization.Serializable

@Serializable
enum class UserRole {
    CITIZEN,
    OFFICIAL,
    ADMIN
}

@Serializable
enum class AlertTier {
    SAFE,
    TIER_1_NOTICE,
    TIER_2_WARNING,
    TIER_3_DANGER
}

@Serializable
enum class WildlifeSpecies {
    TIGER,
    ELEPHANT,
    BEAR,
    WILD_BOAR,
    DEER,
    MONKEY,
    UNKNOWN
}

@Serializable
data class GeoLocation(
    val latitude: Double,
    val longitude: Double,
    val altitude: Double? = null,
    val geohash: String? = null
)
