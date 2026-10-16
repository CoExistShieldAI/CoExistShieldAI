package ai.coexistshield.geofence

import ai.coexistshield.model.AlertTier
import ai.coexistshield.model.GeofenceZone
import ai.coexistshield.model.GeoLocation
import ai.coexistshield.model.WildlifeSpecies
import ai.coexistshield.network.GeofenceEventDto
import ai.coexistshield.repository.GeofenceRepository
import io.ktor.client.HttpClient
import io.ktor.client.engine.mock.MockEngine
import io.ktor.client.engine.mock.respond
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.http.ContentType
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.headersOf
import io.ktor.serialization.kotlinx.json.json
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class GeofenceManagerTest {

    private fun createSampleZone(
        id: String,
        lat: Double,
        lon: Double,
        tier: AlertTier
    ): GeofenceZone {
        return GeofenceZone(
            id = id,
            tier = tier,
            targetSpecies = listOf(WildlifeSpecies.BEAR),
            centroid = GeoLocation(lat, lon),
            radiusMeters = 500.0,
            causalExplanation = "Sample zone $id",
            validFrom = "2026-10-15T00:00:00Z",
            validUntil = "2026-10-18T00:00:00Z",
            generatedAt = "2026-10-15T00:00:00Z"
        )
    }

    @Test
    fun testSmartFilteringAndQuotaLimit() {
        val userLoc = GeoLocation(39.7020, 141.1520)

        // 10 zones with varied tiers and distances
        val zones = listOf(
            createSampleZone("far_danger", 39.7500, 141.1520, AlertTier.TIER_3_DANGER),   // ~5.3 km away
            createSampleZone("close_notice", 39.7030, 141.1520, AlertTier.TIER_1_NOTICE), // ~0.1 km away
            createSampleZone("close_danger", 39.7040, 141.1520, AlertTier.TIER_3_DANGER), // ~0.2 km away
            createSampleZone("mid_warning", 39.7200, 141.1520, AlertTier.TIER_2_WARNING),  // ~2.0 km away
            createSampleZone("far_safe", 39.9000, 141.1520, AlertTier.SAFE),              // ~22 km away
            createSampleZone("extra_1", 39.7100, 141.1520, AlertTier.TIER_1_NOTICE),
            createSampleZone("extra_2", 39.7150, 141.1520, AlertTier.TIER_2_WARNING)
        )

        val registrar = InMemoryNativeGeofenceRegistrar()
        val mockHttpClient = HttpClient(MockEngine { respond("") })
        val repo = GeofenceRepository(mockHttpClient)
        val manager = GeofenceManager(
            repository = repo,
            registrar = registrar,
            maxGeofences = 3 // Limit to 3 (simulating OS quota restriction)
        )

        val ranked = manager.filterAndRankGeofences(zones, userLoc, limit = 3)

        assertEquals(3, ranked.size, "Should strictly conform to maxGeofences limit of 3")
        // "close_danger" should be ranked highest due to TIER_3 and close distance
        assertEquals("close_danger", ranked[0].id)
        // Verify low priority far safe zone is pruned
        assertTrue(ranked.none { it.id == "far_safe" })
    }

    @Test
    fun testMovementHysteresisThreshold() = runTest {
        val initialLoc = GeoLocation(39.7020, 141.1520)
        var fetchCount = 0

        val sampleZones = listOf(
            createSampleZone("zone_1", 39.7030, 141.1520, AlertTier.TIER_3_DANGER)
        )

        val jsonSerializer = Json { ignoreUnknownKeys = true }
        val mockEngine = MockEngine { request ->
            fetchCount++
            respond(
                content = jsonSerializer.encodeToString(sampleZones),
                status = HttpStatusCode.OK,
                headers = headersOf(HttpHeaders.ContentType, ContentType.Application.Json.toString())
            )
        }

        val httpClient = HttpClient(mockEngine) {
            install(ContentNegotiation) {
                json(jsonSerializer)
            }
        }
        val repo = GeofenceRepository(httpClient)
        val registrar = InMemoryNativeGeofenceRegistrar()

        val manager = GeofenceManager(
            repository = repo,
            registrar = registrar,
            maxGeofences = 20,
            movementThresholdMeters = 500.0
        )

        // Initial location set
        manager.onLocationChanged(initialLoc)
        assertEquals(1, fetchCount, "Should fetch geofences on initial location")
        assertEquals(1, registrar.getRegisteredZoneIds().size)

        // Small move: ~50 meters north (latitude +0.00045 degrees is ~50m)
        val slightMove = GeoLocation(39.70245, 141.1520)
        manager.onLocationChanged(slightMove)
        assertEquals(1, fetchCount, "Should skip re-registration when moving under 500m threshold")

        // Significant move: ~1000 meters north (latitude +0.009 degrees is ~1000m)
        val significantMove = GeoLocation(39.7110, 141.1520)
        manager.onLocationChanged(significantMove)
        assertEquals(2, fetchCount, "Should trigger re-registration after moving past 500m threshold")
    }

    @Test
    fun testBackendEventTriggersImmediateReconciliation() = runTest {
        val userLoc = GeoLocation(39.7020, 141.1520)
        var fetchCount = 0

        val initialZones = listOf(
            createSampleZone("zone_a", 39.7030, 141.1520, AlertTier.TIER_2_WARNING)
        )
        val updatedZones = listOf(
            createSampleZone("zone_a", 39.7030, 141.1520, AlertTier.TIER_2_WARNING),
            createSampleZone("zone_new_ai", 39.7025, 141.1520, AlertTier.TIER_3_DANGER)
        )

        val jsonSerializer = Json { ignoreUnknownKeys = true }
        val mockEngine = MockEngine { _ ->
            fetchCount++
            val zonesToReturn = if (fetchCount == 1) initialZones else updatedZones
            respond(
                content = jsonSerializer.encodeToString(zonesToReturn),
                status = HttpStatusCode.OK,
                headers = headersOf(HttpHeaders.ContentType, ContentType.Application.Json.toString())
            )
        }

        val httpClient = HttpClient(mockEngine) {
            install(ContentNegotiation) {
                json(jsonSerializer)
            }
        }
        val repo = GeofenceRepository(httpClient)
        val registrar = InMemoryNativeGeofenceRegistrar()

        val manager = GeofenceManager(
            repository = repo,
            registrar = registrar,
            maxGeofences = 20
        )

        // Setup user location
        manager.onLocationChanged(userLoc)
        assertEquals(1, fetchCount)
        assertEquals(setOf("zone_a"), registrar.getRegisteredZoneIds())

        // Backend notifies an AI recalculation or manual geofence addition
        manager.handleBackendGeofenceEvent(
            GeofenceEventDto(
                type = "RECALCULATED",
                action = "recalculated",
                message = "New high risk zone detected"
            )
        )

        assertEquals(2, fetchCount, "Backend event must trigger immediate geofence refresh")
        assertEquals(setOf("zone_a", "zone_new_ai"), registrar.getRegisteredZoneIds())
        assertEquals(GeofenceSyncState.SYNCED, manager.syncState.value)
    }
}
