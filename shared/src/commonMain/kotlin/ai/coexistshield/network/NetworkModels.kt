package ai.coexistshield.network

import ai.coexistshield.model.AlertTier
import ai.coexistshield.model.FruitLoadLevel
import ai.coexistshield.model.RipenessStage
import kotlinx.serialization.Serializable

@Serializable
data class ForensicTriageRequest(
    val imageBase64: String,
    val categoryHint: String? = null,
    val latitude: Double,
    val longitude: Double
)

@Serializable
data class TriageAttributes(
    val fruitLoadLevel: FruitLoadLevel? = null,
    val ripenessStage: RipenessStage? = null,
    val decayState: String? = null,
    val traumaType: String? = null
)

@Serializable
data class ForensicTriageResponse(
    val detectedSpecies: String,
    val commonName: String? = null,
    val confidence: Double,
    val attributes: TriageAttributes? = null,
    val suggestedTags: List<String> = emptyList(),
    val preliminaryThreatIndex: Double = 0.0,
    val notes: String? = null
)

@Serializable
data class GeofenceRecalculateRequest(
    val triggerEntityRef: String,
    val targetCenterLatitude: Double,
    val targetCenterLongitude: Double,
    val radiusKm: Double = 5.0
)

@Serializable
data class GeofenceRecalculateResponse(
    val recalculatedGeofencesCount: Int,
    val affectedTiers: List<AlertTier> = emptyList(),
    val updatedAt: String
)

@Serializable
data class PriorityKakiTreeDto(
    val treeId: String,
    val hazardScore: Double,
    val distanceToSchoolRouteMeters: Int,
    val estimatedBearAttractantRadiusM: Int,
    val recommendedClearanceOrder: Int
)

@Serializable
data class TopHazardsResponse(
    val municipalityName: String,
    val totalCatalogedTrees: Int,
    val top5PercentCutoffHazardScore: Double,
    val priorityTrees: List<PriorityKakiTreeDto> = emptyList()
)
