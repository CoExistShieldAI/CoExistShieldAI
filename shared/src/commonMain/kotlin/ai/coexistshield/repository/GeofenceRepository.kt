package ai.coexistshield.repository

import ai.coexistshield.model.GeofenceZone
import ai.coexistshield.network.GeofenceRecalculateRequest
import ai.coexistshield.network.GeofenceRecalculateResponse
import ai.coexistshield.network.TopHazardsResponse
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.get
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.http.ContentType
import io.ktor.http.contentType

class GeofenceRepository(
    private val httpClient: HttpClient
) {
    suspend fun getActiveGeofences(
        latitude: Double,
        longitude: Double,
        radiusKm: Double = 10.0
    ): Result<List<GeofenceZone>> {
        return runCatching {
            httpClient.get("/api/v1/geofences/active") {
                url {
                    parameters.append("lat", latitude.toString())
                    parameters.append("long", longitude.toString())
                    parameters.append("radiusKm", radiusKm.toString())
                }
            }.body()
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
}
