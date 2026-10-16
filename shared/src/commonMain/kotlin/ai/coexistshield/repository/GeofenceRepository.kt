package ai.coexistshield.repository

import ai.coexistshield.model.GeofenceZone
import ai.coexistshield.network.GeofenceEventDto
import ai.coexistshield.network.GeofenceRecalculateRequest
import ai.coexistshield.network.GeofenceRecalculateResponse
import ai.coexistshield.network.TopHazardsResponse
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.get
import io.ktor.client.request.header
import io.ktor.client.request.prepareGet
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.bodyAsChannel
import io.ktor.http.ContentType
import io.ktor.http.contentType
import io.ktor.utils.io.readUTF8Line
import kotlinx.coroutines.currentCoroutineContext
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.isActive
import kotlinx.serialization.json.Json

class GeofenceRepository(
    private val httpClient: HttpClient
) {
    private val jsonParser = Json {
        ignoreUnknownKeys = true
        isLenient = true
    }

    suspend fun getActiveGeofences(
        latitude: Double,
        longitude: Double,
        radiusKm: Double = 10.0,
        limit: Int? = null
    ): Result<List<GeofenceZone>> {
        return runCatching {
            httpClient.get("/api/v1/geofences/active") {
                url {
                    parameters.append("lat", latitude.toString())
                    parameters.append("long", longitude.toString())
                    parameters.append("radiusKm", radiusKm.toString())
                    limit?.let { parameters.append("limit", it.toString()) }
                }
            }.body<List<GeofenceZone>?>() ?: emptyList()
        }
    }

    /**
     * Observes real-time geofence events streamed from backend Server-Sent Events (SSE).
     * Automatically reconnects with exponential/fixed backoff upon disconnection.
     */
    fun observeGeofenceEvents(): Flow<GeofenceEventDto> = flow {
        while (currentCoroutineContext().isActive) {
            try {
                httpClient.prepareGet("/api/v1/geofences/events") {
                    header("Accept", "text/event-stream")
                }.execute { response ->
                    val channel = response.bodyAsChannel()
                    while (!channel.isClosedForRead && currentCoroutineContext().isActive) {
                        val line = channel.readUTF8Line() ?: break
                        if (line.startsWith("data:")) {
                            val jsonStr = line.substring(5).trim()
                            if (jsonStr.isNotEmpty()) {
                                runCatching {
                                    jsonParser.decodeFromString<GeofenceEventDto>(jsonStr)
                                }.onSuccess { event ->
                                    emit(event)
                                }
                            }
                        }
                    }
                }
            } catch (e: Throwable) {
                // Connection dropped or unreachable; wait before reconnecting
                delay(3000)
            }
        }
    }

    suspend fun triggerGeofenceRecalculation(
        triggerEntityRef: String,
        latitude: Double,
        longitude: Double
    ): Result<GeofenceRecalculateResponse> {
        return runCatching {
            httpClient.post("/api/v1/geofences/recalculate") {
                contentType(ContentType.Application.Json)
                setBody(
                    GeofenceRecalculateRequest(
                        triggerEntityRef = triggerEntityRef,
                        targetCenterLatitude = latitude,
                        targetCenterLongitude = longitude
                    )
                )
            }.body()
        }
    }

    suspend fun getTopHazardPersimmonTrees(
        municipalityCode: String = "03201",
        limit: Int = 50
    ): Result<TopHazardsResponse> {
        return runCatching {
            httpClient.get("/api/v1/analytics/japan-kaki/top-hazards") {
                url {
                    parameters.append("municipalityCode", municipalityCode)
                    parameters.append("limit", limit.toString())
                }
            }.body()
        }
    }

    suspend fun generateScentPlumeGeofences(
        request: ai.coexistshield.network.GeneratePlumeGeofencesRequest
    ): Result<ai.coexistshield.network.GeneratePlumeGeofencesResponse> {
        return runCatching {
            httpClient.post("/api/v1/geofences/scent-plume") {
                contentType(ContentType.Application.Json)
                setBody(request)
            }.body()
        }
    }

    suspend fun searchEpisodicMemory(
        request: ai.coexistshield.network.MemorySearchRequestDto
    ): Result<ai.coexistshield.network.MemorySearchResponseDto> {
        return runCatching {
            httpClient.post("/api/v1/memory/search") {
                contentType(ContentType.Application.Json)
                setBody(request)
            }.body()
        }
    }

    suspend fun explainAlert(
        request: ai.coexistshield.network.CausalAlertExplanationRequestDto
    ): Result<ai.coexistshield.network.CausalAlertExplanationResponseDto> {
        return runCatching {
            httpClient.post("/api/v1/alerts/explain") {
                contentType(ContentType.Application.Json)
                setBody(request)
            }.body()
        }
    }

    suspend fun getMlStatus(): Result<ai.coexistshield.network.MlStatusResponseDto> {
        return runCatching {
            httpClient.get("/api/v1/ml/status").body()
        }
    }

    suspend fun triggerMlRetrain(
        request: ai.coexistshield.network.MlRetrainRequestDto = ai.coexistshield.network.MlRetrainRequestDto()
    ): Result<ai.coexistshield.network.MlRetrainResponseDto> {
        return runCatching {
            httpClient.post("/api/v1/ml/retrain") {
                contentType(ContentType.Application.Json)
                setBody(request)
            }.body()
        }
    }

    suspend fun predictRisk(
        request: ai.coexistshield.network.MlPredictRequestDto
    ): Result<ai.coexistshield.network.MlPredictResponseDto> {
        return runCatching {
            httpClient.post("/api/v1/ml/predict") {
                contentType(ContentType.Application.Json)
                setBody(request)
            }.body()
        }
    }

    suspend fun getBigQueryFeatureSql(): Result<ai.coexistshield.network.MlQueryResponseDto> {
        return runCatching {
            httpClient.get("/api/v1/ml/query").body()
        }
    }

    suspend fun getThailandCropPlots(): Result<ai.coexistshield.network.ThailandCropPlotsResponseDto> {
        return runCatching {
            httpClient.get("/api/v1/analytics/thailand-hec/crop-plots").body()
        }
    }

    suspend fun updateCropHarvestStage(
        plotId: String,
        newStage: String
    ): Result<ai.coexistshield.network.ThailandCropPlotDto> {
        return runCatching {
            httpClient.post("/api/v1/analytics/thailand-hec/plots/$plotId/harvest-stage") {
                contentType(ContentType.Application.Json)
                setBody(ai.coexistshield.network.UpdateHarvestStageRequestDto(growthStage = newStage))
            }.body()
        }
    }

    suspend fun getTigerPastures(): Result<ai.coexistshield.network.TigerPasturesResponseDto> {
        return runCatching {
            httpClient.get("/api/v1/analytics/sumatra-tiger/pastures").body()
        }
    }

    suspend fun reportTigerDepredation(
        request: ai.coexistshield.network.ReportDepredationRequestDto
    ): Result<ai.coexistshield.network.TigerDepredationResponseDto> {
        return runCatching {
            httpClient.post("/api/v1/analytics/sumatra-tiger/depredation/report") {
                contentType(ContentType.Application.Json)
                setBody(request)
            }.body()
        }
    }

    suspend fun clearTigerCarcass(
        enclosureId: String
    ): Result<ai.coexistshield.network.LivestockEnclosureDto> {
        return runCatching {
            httpClient.post("/api/v1/analytics/sumatra-tiger/pastures/$enclosureId/clear").body()
        }
    }

    suspend fun getRspoAuditPackage(
        plantationId: String = "wilmar_sumatra_02"
    ): Result<ai.coexistshield.network.RspoAuditPackageDto> {
        return runCatching {
            httpClient.get("/api/v1/compliance/rspo/package?plantationId=$plantationId").body()
        }
    }

    suspend fun getGcfBudgetImpact(
        municipalityCode: String = "03201"
    ): Result<ai.coexistshield.network.GcfBudgetImpactResponseDto> {
        return runCatching {
            httpClient.get("/api/v1/compliance/gcf/budget-impact?municipalityCode=$municipalityCode").body()
        }
    }

    suspend fun getPriorityKakiLayer(
        municipalityCode: String = "03201"
    ): Result<ai.coexistshield.network.PriorityKakiLayerResponseDto> {
        return runCatching {
            httpClient.get("/api/v1/municipal/kaki/priority-layer?municipalityCode=$municipalityCode").body()
        }
    }

    suspend fun logMunicipalKakiAction(
        request: ai.coexistshield.network.MunicipalKakiActionRequestDto
    ): Result<ai.coexistshield.network.MunicipalKakiActionResponseDto> {
        return runCatching {
            httpClient.post("/api/v1/municipal/kaki/action") {
                contentType(ContentType.Application.Json)
                setBody(request)
            }.body()
        }
    }

    suspend fun ingestUnstructuredNlp(
        rawText: String,
        sourceChannel: String? = null,
        regionHint: String? = null
    ): Result<ai.coexistshield.network.UnstructuredNlpIngestResponseDto> {
        return runCatching {
            httpClient.post("/api/v1/nlp/ingest/unstructured") {
                contentType(ContentType.Application.Json)
                setBody(
                    ai.coexistshield.network.UnstructuredNlpIngestRequestDto(
                        rawText = rawText,
                        sourceChannel = sourceChannel,
                        regionHint = regionHint
                    )
                )
            }.body()
        }
    }

    suspend fun getCorroborationStreams(
        region: String? = null,
        streamType: String? = null
    ): Result<List<ai.coexistshield.network.CausalStreamItemDto>> {
        return runCatching {
            httpClient.get("/api/v1/corroboration/streams") {
                url {
                    region?.let { parameters.append("region", it) }
                    streamType?.let { parameters.append("type", it) }
                }
            }.body()
        }
    }

    suspend fun evaluateCorroboration(
        targetLatitude: Double,
        targetLongitude: Double,
        radiusKm: Double = 5.0,
        region: String? = null
    ): Result<ai.coexistshield.network.CorroborationResponseDto> {
        return runCatching {
            httpClient.post("/api/v1/corroboration/evaluate") {
                contentType(ContentType.Application.Json)
                setBody(
                    ai.coexistshield.network.CorroborationRequestDto(
                        targetLatitude = targetLatitude,
                        targetLongitude = targetLongitude,
                        radiusKm = radiusKm,
                        region = region
                    )
                )
            }.body()
        }
    }

    suspend fun runCorroborationPollCycle(): Result<Boolean> {
        return runCatching {
            val response = httpClient.post("/api/v1/corroboration/poll-cycle")
            response.status.value in 200..299
        }
    }

    suspend fun getCurrentMeteorology(
        region: String = "JAPAN"
    ): Result<ai.coexistshield.network.MeteorologicalConditionsDto> {
        return runCatching {
            httpClient.get("/api/v1/meteorology/current?region=$region").body()
        }
    }
}


