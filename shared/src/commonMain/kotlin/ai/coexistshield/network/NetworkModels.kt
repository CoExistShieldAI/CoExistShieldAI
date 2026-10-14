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
    val traumaType: String? = null,
    val traceType: String? = null,
    val traceAgeHours: Double? = null
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
data class MoonPhaseInfoDto(
    val phaseName: String,
    val illumination: Double,
    val nocturnalRiskMultiplier: Double,
    val isNocturnalPeakRisk: Boolean
)

@Serializable
data class MeteorologicalConditionsDto(
    val temperatureCelsius: Double,
    val humidityPct: Double,
    val windSpeedKmh: Double,
    val windDirectionDeg: Double,
    val precipitationMmH: Double,
    val isFloodedRiverRisk: Boolean,
    val moonPhase: MoonPhaseInfoDto,
    val timestamp: String
)

@Serializable
data class WeatherConditionsDto(
    val windSpeedKmh: Double,
    val windDirectionDeg: Double,
    val moonIllumination: Double,
    val precipitationMmH: Double = 0.0,
    val temperatureCelsius: Double? = null,
    val humidityPct: Double? = null
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
data class DispatchRecommendationDto(
    val actionType: String,
    val priority: String,
    val targetEntityRef: String,
    val location: String,
    val justification: String
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
    val effectiveTier: AlertTier,
    val dispatchRecommendations: List<DispatchRecommendationDto> = emptyList()
)

@Serializable
data class BigQuerySyncStatusDto(
    val status: String,
    val replicationLagSeconds: Double,
    val slaTargetSeconds: Int = 60,
    val totalEventsReplicated: Int,
    val stagedSamplesCount: Int,
    val tablePartitioning: String,
    val spatialIndexing: String,
    val lastSyncAt: String
)

@Serializable
data class IncidentStreamingEventDto(
    val eventId: String,
    val incidentId: String,
    val species: String,
    val category: String,
    val latitude: Double,
    val longitude: Double,
    val timestamp: String,
    val isIncursion: Int = 1,
    val verificationState: String = "VERIFIED"
)

@Serializable
data class BigQueryStreamSyncResponseDto(
    val syncedEventsCount: Int,
    val syncStatus: BigQuerySyncStatusDto,
    val stagedLakeSamples: Int
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

@Serializable
data class ThailandCropPlotDto(
    val id: String,
    val name: String,
    val province: String,
    val parkBoundary: String,
    val cropType: String,
    val growthStage: String,
    val latitude: Double,
    val longitude: Double,
    val areaHectares: Double,
    val distanceToForestEdgeM: Double,
    val forestBearingDeg: Double,
    val aromaVolatilityIndex: Double,
    val maturityScore: Double,
    val activeRiskScore: Double,
    val alertTier: AlertTier,
    val scentBufferMeters: Double,
    val scentPlumePolygon: GeoJsonPolygonDto? = null,
    val updatedAt: String
)

@Serializable
data class ThailandCropPlotsResponseDto(
    val plots: List<ThailandCropPlotDto> = emptyList(),
    val count: Int,
    val monitoredParks: List<String> = emptyList()
)

@Serializable
data class UpdateHarvestStageRequestDto(
    val growthStage: String
)

@Serializable
data class LivestockEnclosureDto(
    val id: String,
    val name: String,
    val province: String,
    val village: String,
    val livestockType: String,
    val headCount: Int,
    val hasNightCorral: Boolean,
    val latitude: Double,
    val longitude: Double,
    val distanceToForestEdgeM: Double,
    val isActiveKillSite: Boolean,
    val killReportedAt: String? = null,
    val hoursSinceKill: Double = 0.0,
    val cachingRiskExpiry: String? = null,
    val returnProbability: Double = 0.0,
    val activeThreatScore: Double = 0.0,
    val alertTier: AlertTier,
    val scentBufferMeters: Double = 0.0,
    val killSiteGeofence: GeoJsonPolygonDto? = null,
    val forensicNotes: String? = null,
    val updatedAt: String
)

@Serializable
data class ReportDepredationRequestDto(
    val enclosureId: String,
    val livestockSpecies: String,
    val headCountLost: Int,
    val latitude: Double,
    val longitude: Double,
    val hasNightCorral: Boolean,
    val photoUrl: String? = null,
    val dragTrailLengthMeters: Double = 0.0,
    val canineBiteSpreadMm: Double = 0.0
)

@Serializable
data class TigerDepredationAlertDto(
    val alertId: String,
    val tier: AlertTier,
    val killSiteLat: Double,
    val killSiteLon: Double,
    val villagesIn3kmRadius: List<String> = emptyList(),
    val returnWindowHoursRemaining: Double = 0.0,
    val canineForensicsConfirmed: Boolean = false,
    val causalReasoning: String = "",
    val issuedAt: String = ""
)

@Serializable
data class TigerPasturesResponseDto(
    val pastures: List<LivestockEnclosureDto> = emptyList(),
    val count: Int,
    val updatedAt: String
)

@Serializable
data class TigerDepredationResponseDto(
    val enclosure: LivestockEnclosureDto,
    val alert: TigerDepredationAlertDto
)

@Serializable
data class RspoAuditPackageDto(
    val auditId: String,
    val plantationId: String,
    val plantationName: String,
    val province: String,
    val reportingYear: Int,
    val totalMonitoredHectares: Double,
    val hcvBufferForestHectares: Double,
    val humanFatalities: Int,
    val wildlifeFatalities: Int,
    val acousticDeterrentDeployments: Int,
    val bioAcousticDeflections: Int,
    val verifiedSafePassages: Int,
    val carcassRemovals: Int,
    val zeroHarmComplianceScore: Double,
    val rspoCertificateStatus: String,
    val sha256AuditSignature: String,
    val generatedAt: String
)

@Serializable
data class FellingAuditLogDto(
    val logId: String,
    val treeId: String,
    val permitId: String,
    val officerId: String,
    val action: String,
    val schoolDistanceM: Int,
    val riskDropPct: Double,
    val biomassKg: Double,
    val timestamp: String
)

@Serializable
data class GcfBudgetImpactResponseDto(
    val municipalityCode: String,
    val municipalityName: String,
    val campaignTitle: String,
    val totalFundingGoalYen: Long,
    val currentRaisedYen: Long,
    val fundingProgressPct: Double,
    val donorCount: Int,
    val costPerTreeYen: Int,
    val targetTreesCount: Int,
    val completedFelledTreesCount: Int,
    val remainingTreesCount: Int,
    val schoolChildrenProtectedCount: Int,
    val quantifiedBearEncounterRiskReduction: Double,
    val furusatoTaxPortalUrl: String,
    val recentFellingLogs: List<FellingAuditLogDto> = emptyList(),
    val updatedAt: String
)

@Serializable
data class MunicipalPriorityKakiTreeDto(
    val treeId: String,
    val address: String,
    val latitude: Double,
    val longitude: Double,
    val hazardScore: Double,
    val isTop5Percent: Boolean,
    val distanceToSchoolRouteMeters: Int,
    val estimatedBearAttractantRadiusM: Int,
    val ownerConsentStatus: String,
    val permitId: String,
    val managementStatus: String,
    val recommendedClearanceOrder: Int,
    val assignedSquad: String,
    val fruitLoad: String,
    val felledAt: String? = null
)

@Serializable
data class PriorityKakiLayerResponseDto(
    val municipalityCode: String,
    val municipalityName: String,
    val totalCatalogedTrees: Int,
    val top5PercentCutoff: Double,
    val priorityTrees: List<MunicipalPriorityKakiTreeDto> = emptyList(),
    val activeBaselineThreat: Double,
    val updatedAt: String
)

@Serializable
data class MunicipalKakiActionRequestDto(
    val treeId: String,
    val action: String,
    val officerId: String,
    val officerRole: String,
    val permitNumber: String,
    val biomassDisposedKg: Double = 0.0,
    val notes: String? = null
)

@Serializable
data class MunicipalKakiActionResponseDto(
    val auditLogId: String,
    val updatedTree: MunicipalPriorityKakiTreeDto,
    val treeRiskDropPct: Double,
    val newMunicipalityRiskScore: Double,
    val gcfCompletedFelledTrees: Int,
    val gcfRiskReductionPct: Double,
    val message: String,
    val processedAt: String
)

// Issue #10: Gemini 3.8 Flash NLP Ingestion DTOs

@Serializable
data class ExtractedIncidentEntityDto(
    val entityId: String,
    val sourceChannel: String,
    val rawText: String,
    val species: String,
    val scientificName: String,
    val commonName: String,
    val incidentType: String,
    val locationDescription: String,
    val estimatedLatitude: Double,
    val estimatedLongitude: Double,
    val individualCount: Int,
    val attractantContext: String,
    val urgencyLevel: String,
    val confidence: Double,
    val extractedTimestamp: String,
    val corroborationTags: List<String> = emptyList(),
    val notes: String
)

@Serializable
data class UnstructuredNlpIngestRequestDto(
    val rawText: String,
    val sourceChannel: String? = null,
    val regionHint: String? = null
)

@Serializable
data class UnstructuredNlpIngestResponseDto(
    val entity: ExtractedIncidentEntityDto,
    val modelUsed: String,
    val createdAt: String
)

// Issue #4: Multi-Source Causal Corroboration Engine DTOs

@Serializable
data class CausalStreamItemDto(
    val itemId: String,
    val streamType: String,
    val region: String,
    val title: String,
    val description: String,
    val latitude: Double,
    val longitude: Double,
    val radiusMeters: Double,
    val confidence: Double,
    val targetSpecies: String,
    val timestamp: String,
    val expiresAt: String
)

@Serializable
data class CorroborationRequestDto(
    val targetLatitude: Double,
    val targetLongitude: Double,
    val radiusKm: Double = 5.0,
    val maxAgeHours: Double = 12.0,
    val region: String? = null
)

@Serializable
data class CorroboratedClusterDto(
    val clusterId: String,
    val centerLatitude: Double,
    val centerLongitude: Double,
    val radiusKm: Double,
    val targetSpecies: String,
    val corroboratedSources: List<String> = emptyList(),
    val independentSourceCount: Int,
    val combinedConfidence: Double,
    val elevatedTier: AlertTier,
    val aiJustification: String,
    val contributingEvents: List<CausalStreamItemDto> = emptyList(),
    val actionRecommendation: String,
    val evaluatedAt: String
)

@Serializable
data class CorroborationResponseDto(
    val clusters: List<CorroboratedClusterDto> = emptyList(),
    val totalActiveStreams: Int,
    val evaluatedAt: String,
    val modelUsed: String
)

@Serializable
data class UpdateRoadkillStatusRequestDto(
    val status: String,
    val assignedOfficialId: String? = null
)

// Issue #23: Open-to-Enterprise Cold-Start Strategy & Conservation NGO Data Ingestion DTOs

@Serializable
data class ColdStartStatusDto(
    val bootstrapped: Boolean = false,
    val totalGbifOccurrences: Int = 0,
    val totalFirmsAnomalies: Int = 0,
    val totalNgoPatrolRecords: Int = 0,
    val activeBaselineGeofences: Int = 0,
    val regionsCovered: List<String> = emptyList(),
    val lastBootstrapTimestamp: String = "",
    val readyForDayOneDeployment: Boolean = true
)

@Serializable
data class NgoPatrolRecordDto(
    val recordId: String,
    val organization: String,
    val rangerUnit: String,
    val species: String,
    val observationType: String,
    val latitude: Double,
    val longitude: Double,
    val timestamp: String,
    val individualCount: Int = 1,
    val corridorRiskTier: String,
    val notes: String = ""
)

@Serializable
data class NgoBulkImportRequestDto(
    val format: String = "AUTO",
    val organization: String,
    val rangerStation: String,
    val region: String = "JAPAN",
    val payload: String
)

@Serializable
data class NgoBulkImportResponseDto(
    val importedCount: Int,
    val formatDetected: String,
    val processingDurationMs: Long,
    val speciesBreakdown: Map<String, Int> = emptyMap(),
    val geofencesActivated: Int,
    val importedRecords: List<NgoPatrolRecordDto> = emptyList(),
    val message: String
)

@Serializable
data class NgoTemplatesResponseDto(
    val templates: Map<String, String> = emptyMap(),
    val formats: List<String> = emptyList(),
    val sla: String = "5 minutes"
)








