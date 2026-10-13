package ai.coexistshield.repository

import ai.coexistshield.model.Incident
import ai.coexistshield.model.ReportCategory
import ai.coexistshield.model.VerificationStatus
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.request.get
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.http.ContentType
import io.ktor.http.contentType
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow

class IncidentRepository(
    private val httpClient: HttpClient
) {
    private val _offlineIncidentsQueue = MutableStateFlow<List<Incident>>(emptyList())
    val offlineIncidentsQueue: Flow<List<Incident>> = _offlineIncidentsQueue.asStateFlow()

    suspend fun submitIncident(incident: Incident): Result<Unit> {
        return runCatching {
            httpClient.post("/api/v1/incidents") {
                contentType(ContentType.Application.Json)
                setBody(incident)
            }
            Unit
        }.onFailure {
            // Enqueue locally for offline persistence
            _offlineIncidentsQueue.value = _offlineIncidentsQueue.value + incident
        }
    }

    suspend fun getRecentIncidents(
        category: ReportCategory? = null,
        limit: Int = 50
    ): Result<List<Incident>> {
        return runCatching {
            httpClient.get("/api/v1/incidents") {
                url {
                    category?.let { parameters.append("category", it.name) }
                    parameters.append("limit", limit.toString())
                }
            }.body()
        }
    }

    suspend fun syncOfflineQueue(): Int {
        val queue = _offlineIncidentsQueue.value
        if (queue.isEmpty()) return 0

        var syncedCount = 0
        val remaining = mutableListOf<Incident>()

        for (item in queue) {
            val result = submitIncident(item)
            if (result.isSuccess) {
                syncedCount++
            } else {
                remaining.add(item)
            }
        }
        _offlineIncidentsQueue.value = remaining
        return syncedCount
    }
}
