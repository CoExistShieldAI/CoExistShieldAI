package ai.coexistshield.model

import kotlinx.serialization.Serializable

@Serializable
data class GeoPolygonCoordinates(
    val ring: List<GeoLocation>
)

@Serializable
data class GeofenceZone(
    val id: String,
    val tier: AlertTier,
    val targetSpecies: List<WildlifeSpecies> = emptyList(),
    val coordinates: List<GeoLocation> = emptyList(),
    val centroid: GeoLocation,
    val radiusMeters: Double,
    val causalExplanation: String,
    val primaryAttractantRef: String? = null,
    val validFrom: String,
    val validUntil: String,
    val generatedAt: String
)
