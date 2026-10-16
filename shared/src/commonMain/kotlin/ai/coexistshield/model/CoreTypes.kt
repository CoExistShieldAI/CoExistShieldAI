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

fun GeoLocation.distanceToMeters(other: GeoLocation): Double {
    val earthRadius = 6371000.0
    val dLat = (other.latitude - latitude) * (kotlin.math.PI / 180.0)
    val dLon = (other.longitude - longitude) * (kotlin.math.PI / 180.0)
    val lat1 = latitude * (kotlin.math.PI / 180.0)
    val lat2 = other.latitude * (kotlin.math.PI / 180.0)

    val a = kotlin.math.sin(dLat / 2) * kotlin.math.sin(dLat / 2) +
            kotlin.math.cos(lat1) * kotlin.math.cos(lat2) *
            kotlin.math.sin(dLon / 2) * kotlin.math.sin(dLon / 2)
    val c = 2 * kotlin.math.atan2(kotlin.math.sqrt(a), kotlin.math.sqrt(1 - a))
    return earthRadius * c
}

fun GeoLocation.distanceToKm(other: GeoLocation): Double = distanceToMeters(other) / 1000.0

