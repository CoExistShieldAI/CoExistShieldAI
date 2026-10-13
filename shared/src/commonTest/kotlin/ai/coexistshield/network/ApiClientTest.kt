package ai.coexistshield.network

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
import kotlinx.serialization.json.Json
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class ApiClientTest {

    private fun createMockHttpClient(responseJson: String, statusCode: HttpStatusCode = HttpStatusCode.OK): HttpClient {
        val mockEngine = MockEngine { _ ->
            respond(
                content = responseJson,
                status = statusCode,
                headers = headersOf(HttpHeaders.ContentType, ContentType.Application.Json.toString())
            )
        }

        return HttpClient(mockEngine) {
            install(ContentNegotiation) {
                json(Json {
                    prettyPrint = true
                    isLenient = true
                    ignoreUnknownKeys = true
                })
            }
        }
    }

    @Test
    fun testGeminiTriageServiceSuccess() = runTest {
        val mockJsonResponse = """
            {
                "detectedSpecies": "Diospyros kaki",
                "commonName": "Persimmon Tree",
                "confidence": 0.96,
                "attributes": {
                    "fruitLoadLevel": "ABUNDANT",
                    "ripenessStage": "OVERRIPE_FALLING"
                },
                "suggestedTags": ["abundant_fruit", "overripe_falling"],
                "preliminaryThreatIndex": 0.88,
                "notes": "Verified high bear attractant hazard."
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val service = GeminiTriageService(httpClient)

        val result = service.triagePhoto(
            imageBase64 = "mockBase64String",
            categoryHint = "ATTRACTANT_FRUIT_TREE",
            latitude = 38.256,
            longitude = 140.321
        )

        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertEquals("Diospyros kaki", data.detectedLabel)
        assertEquals(0.96, data.confidence)
        assertEquals("ABUNDANT", data.fruitLoadLevel)
        assertTrue(data.forensicTags.contains("abundant_fruit"))
    }

    @Test
    fun testTopHazardPersimmonTreesApi() = runTest {
        val mockJsonResponse = """
            {
                "municipalityName": "Morioka City",
                "totalCatalogedTrees": 840,
                "top5PercentCutoffHazardScore": 0.89,
                "priorityTrees": [
                    {
                        "treeId": "kt_1120",
                        "hazardScore": 0.95,
                        "distanceToSchoolRouteMeters": 120,
                        "estimatedBearAttractantRadiusM": 800,
                        "recommendedClearanceOrder": 1
                    }
                ]
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.getTopHazardPersimmonTrees("03201")
        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertEquals("Morioka City", data.municipalityName)
        assertEquals(840, data.totalCatalogedTrees)
        assertEquals(1, data.priorityTrees.size)
        assertEquals("kt_1120", data.priorityTrees.first().treeId)
        assertEquals(0.95, data.priorityTrees.first().hazardScore)
    }

    @Test
    fun testScentPlumeGeofencesApi() = runTest {
        val mockJsonResponse = """
            {
                "zones": [
                    {
                        "id": "geo_kt_01_1",
                        "tier": "TIER_2_WARNING",
                        "targetSpecies": ["BEAR"],
                        "polygonGeoJson": {
                            "type": "Polygon",
                            "coordinates": [
                                [[140.321, 38.256], [140.325, 38.257], [140.321, 38.256]]
                            ]
                        },
                        "activeRiskScore": 0.88,
                        "causalFactors": ["Attractant: KAKI_TREE", "Wind: 15.0 km/h"],
                        "expiresAt": "2026-10-16T00:00:00Z",
                        "createdAt": "2026-10-14T00:00:00Z"
                    }
                ],
                "count": 1,
                "createdAt": "2026-10-14T00:00:00Z"
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.generateScentPlumeGeofences(
            GeneratePlumeGeofencesRequest(
                attractants = listOf(
                    AttractantHubDto(
                        id = "kt_01",
                        type = "KAKI_TREE",
                        latitude = 38.256,
                        longitude = 140.321,
                        fruitLoad = "ABUNDANT"
                    )
                ),
                weather = WeatherConditionsDto(
                    windSpeedKmh = 15.0,
                    windDirectionDeg = 270.0,
                    moonIllumination = 0.05
                )
            )
        )

        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertEquals(1, data.count)
        val zone = data.zones.first()
        assertEquals("geo_kt_01_1", zone.id)
        assertEquals(ai.coexistshield.model.AlertTier.TIER_2_WARNING, zone.tier)
        assertEquals(0.88, zone.activeRiskScore)
        assertEquals("Polygon", zone.polygonGeoJson?.type)
    }
}

