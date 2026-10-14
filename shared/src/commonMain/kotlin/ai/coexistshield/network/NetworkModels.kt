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

@Serializable
data class WeatherConditionsDto(
    val windSpeedKmh: Double,
    val windDirectionDeg: Double,
    val moonIllumination: Double,
    val precipitationMmH: Double = 0.0
)

@Serializable
data class AttractantHubDto(
    val id: String,
    val type: String,
    val latitude: Double,
    val longitude: Double,
    val fruitLoad: String? = null,
    val isTop5Hazard: Boolean = false,
    val hoursSinceDeath: Double? = null
)

@Serializable
data class GeneratePlumeGeofencesRequest(
    val attractants: List<AttractantHubDto>,
    val weather: WeatherConditionsDto? = null
)

@Serializable
data class GeoJsonPolygonDto(
    val type: String,
    val coordinates: List<List<List<Double>>>
)

@Serializable
data class GeofenceZoneDto(
    val id: String,
    val tier: AlertTier,
    val targetSpecies: List<String> = emptyList(),
    val polygonGeoJson: GeoJsonPolygonDto? = null,
    val activeRiskScore: Double,
    val causalFactors: List<String> = emptyList(),
    val expiresAt: String,
    val createdAt: String
)

@Serializable
data class GeneratePlumeGeofencesResponse(
    val zones: List<GeofenceZoneDto> = emptyList(),
    val count: Int,
    val createdAt: String
)

@Serializable
data class IncidentMemoryRecordDto(
    val memoryId: String,
    val region: String,
    val species: String,
    val category: String,
    val severityTier: AlertTier,
    val narrative: String,
    val outcome: String,
    val cosineSimilarity: Double = 0.0
)

@Serializable
data class MemorySearchRequestDto(
    val queryText: String,
    val region: String? = null,
    val species: String? = null,
    val topK: Int = 3,
    val minSimilarity: Double = 0.50
)

@Serializable
data class MemorySearchResponseDto(
    val queryVectorGenerated: Boolean = true,
    val matches: List<IncidentMemoryRecordDto> = emptyList(),
    val elevatedTier: AlertTier? = null,
    val elevationReason: String? = null
)

@Serializable
data class CausalAlertExplanationRequestDto(
    val tier: AlertTier,
    val targetSpecies: String,
    val region: String = "JAPAN",
    val retrieveMemory: Boolean = true
)

@Serializable
data class CausalAlertExplanationResponseDto(
    val explanation: String,
    val historicalAnalogs: List<IncidentMemoryRecordDto> = emptyList(),
    val elevatedFromTier: AlertTier? = null,
    val effectiveTier: AlertTier
)

@Serializable
data class MlMetricsDto(
    val aucRoc: Double,
    val precisionTop20: Double,
    val recall: Double = 0.0,
    val f1Score: Double = 0.0,
    val validationLoss: Double = 0.0,
    val testSampleSize: Int = 0,
    val promoted: Boolean = false,
    val promotionReason: String = ""
)

@Serializable
data class MlModelDto(
    val version: String,
    val trainedAt: String,
    val trainingSampleSize: Int = 0,
    val bias: Double = 0.0,
    val weights: Map<String, Double> = emptyMap(),
    val featureImportance: Map<String, Double> = emptyMap(),
    val metrics: MlMetricsDto
)

@Serializable
data class MlRetrainResponseDto(
    val runId: String,
    val trainedAt: String,
    val trigger: String,
    val totalSamples: Int,
    val trainSamples: Int,
    val holdoutSamples: Int,
    val candidateModel: MlModelDto,
    val activeModelVersion: String,
    val promotedToProd: Boolean,
    val gateCheckPassed: Boolean,
    val gateEvaluationNotes: String,
    val regressionDetected: Boolean
)

@Serializable
data class MlStatusResponseDto(
    val status: String,
    val activeModel: MlModelDto,
    val history: List<MlRetrainResponseDto> = emptyList()
)

@Serializable
data class MlRetrainRequestDto(
    val trigger: String = "admin_manual",
    val force: Boolean = false
)

@Serializable
data class MlPredictRequestDto(
    val latitude: Double,
    val longitude: Double,
    val distNearestPersimmonM: Double = 9999.0,
    val distNearestCarcassM: Double = 9999.0,
    val distNearestCaneFruitM: Double = 9999.0,
    val windScentAlignmentDeg: Double = 90.0,
    val moonDarknessPct: Double = 50.0,
    val precipitationMm24h: Double = 0.0,
    val historicalIncidents30d: Double = 0.0
)

@Serializable
data class MlPredictResponseDto(
    val h3Index: String,
    val incursionProbability: Double,
    val alertTier: AlertTier,
    val dominantFeature: String,
    val featureContributions: Map<String, Double> = emptyMap()
)

@Serializable
data class MlQueryResponseDto(
    val query: String,
    val resolution: Int = 8,
    val slaTargetSeconds: Int = 45
)



