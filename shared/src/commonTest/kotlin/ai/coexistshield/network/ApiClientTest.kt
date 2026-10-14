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

    @Test
    fun testEpisodicMemorySearchApi() = runTest {
        val mockJsonResponse = """
            {
                "queryVectorGenerated": true,
                "matches": [
                    {
                        "memoryId": "mem_jp_01",
                        "region": "JAPAN",
                        "species": "BEAR",
                        "category": "ATTACK",
                        "severityTier": "TIER_3_DANGER",
                        "narrative": "Backyard Incursion near school route",
                        "outcome": "Severe human injury",
                        "cosineSimilarity": 0.912
                    }
                ],
                "elevatedTier": "TIER_3_DANGER",
                "elevationReason": "Elevated to Tier 3 Danger: Conditions match severe historical attack pattern."
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.searchEpisodicMemory(
            MemorySearchRequestDto(
                queryText = "Asian Black Bear incursion near unharvested persimmon trees in Iwate valley with West wind",
                region = "JAPAN",
                species = "BEAR",
                minSimilarity = 0.50
            )
        )

        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertTrue(data.queryVectorGenerated)
        assertEquals(1, data.matches.size)
        assertEquals("mem_jp_01", data.matches.first().memoryId)
        assertEquals(ai.coexistshield.model.AlertTier.TIER_3_DANGER, data.elevatedTier)
    }

    @Test
    fun testExplainAlertApi() = runTest {
        val mockJsonResponse = """
            {
                "explanation": "Tier 2 Warning: Conditions match 1 historical bear incident with attack attractants; elevated caution advised.",
                "historicalAnalogs": [
                    {
                        "memoryId": "mem_jp_01",
                        "region": "JAPAN",
                        "species": "BEAR",
                        "category": "ATTACK",
                        "severityTier": "TIER_3_DANGER",
                        "narrative": "Backyard Incursion",
                        "outcome": "Severe injury",
                        "cosineSimilarity": 0.89
                    }
                ],
                "elevatedFromTier": "TIER_1_NOTICE",
                "effectiveTier": "TIER_2_WARNING"
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.explainAlert(
            CausalAlertExplanationRequestDto(
                tier = ai.coexistshield.model.AlertTier.TIER_1_NOTICE,
                targetSpecies = "Asian Black Bear",
                region = "JAPAN",
                retrieveMemory = true
            )
        )

        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertEquals(ai.coexistshield.model.AlertTier.TIER_2_WARNING, data.effectiveTier)
        assertEquals(ai.coexistshield.model.AlertTier.TIER_1_NOTICE, data.elevatedFromTier)
        assertEquals(1, data.historicalAnalogs.size)
        assertTrue(data.explanation.contains("Tier 2 Warning"))
    }

    @Test
    fun testMlStatusApi() = runTest {
        val mockJsonResponse = """
            {
                "status": "HEALTHY",
                "activeModel": {
                    "version": "v2026.10.14-prod",
                    "trainedAt": "2026-10-14T08:00:00Z",
                    "trainingSampleSize": 1200,
                    "bias": -2.15,
                    "weights": {
                        "persimmon_proximity": 2.45
                    },
                    "featureImportance": {
                        "persimmon_proximity": 0.28
                    },
                    "metrics": {
                        "aucRoc": 0.865,
                        "precisionTop20": 0.792,
                        "recall": 0.820,
                        "f1Score": 0.805,
                        "validationLoss": 0.320,
                        "testSampleSize": 240,
                        "promoted": true,
                        "promotionReason": "Validation criteria met"
                    }
                },
                "history": []
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.getMlStatus()
        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertEquals("HEALTHY", data.status)
        assertEquals("v2026.10.14-prod", data.activeModel.version)
        assertEquals(0.865, data.activeModel.metrics.aucRoc)
        assertTrue(data.activeModel.metrics.aucRoc >= 0.82)
    }

    @Test
    fun testMlRetrainApi() = runTest {
        val mockJsonResponse = """
            {
                "runId": "train-run-12345",
                "trainedAt": "2026-10-14T08:30:00Z",
                "trigger": "admin_manual",
                "totalSamples": 100,
                "trainSamples": 80,
                "holdoutSamples": 20,
                "candidateModel": {
                    "version": "v2026.10.14.0830",
                    "trainedAt": "2026-10-14T08:30:00Z",
                    "metrics": {
                        "aucRoc": 0.875,
                        "precisionTop20": 0.800,
                        "promoted": true,
                        "promotionReason": "Validation gate passed"
                    }
                },
                "activeModelVersion": "v2026.10.14.0830",
                "promotedToProd": true,
                "gateCheckPassed": true,
                "gateEvaluationNotes": "AUC: 0.875 >= 0.82 | Prec: 0.800 >= 0.75 => PROMOTED",
                "regressionDetected": false
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.triggerMlRetrain(MlRetrainRequestDto(trigger = "admin_manual"))
        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertTrue(data.gateCheckPassed)
        assertTrue(data.promotedToProd)
        assertEquals("v2026.10.14.0830", data.activeModelVersion)
        assertTrue(data.candidateModel.metrics.aucRoc >= 0.82)
    }

    @Test
    fun testMlPredictRiskApi() = runTest {
        val mockJsonResponse = """
            {
                "h3Index": "882f0a123456789",
                "incursionProbability": 0.84,
                "alertTier": "TIER_3_DANGER",
                "dominantFeature": "carcass_proximity",
                "featureContributions": {
                    "carcass_proximity": 2.15,
                    "persimmon_proximity": 1.20
                }
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.predictRisk(
            MlPredictRequestDto(
                latitude = 39.7020,
                longitude = 141.1520,
                distNearestCarcassM = 300.0
            )
        )
        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertEquals("882f0a123456789", data.h3Index)
        assertEquals(0.84, data.incursionProbability)
        assertEquals(ai.coexistshield.model.AlertTier.TIER_3_DANGER, data.alertTier)
        assertEquals("carcass_proximity", data.dominantFeature)
    }

    @Test
    fun testThailandCropPlotsApi() = runTest {
        val mockJsonResponse = """
            {
                "plots": [
                    {
                        "id": "th_plot_kb_01",
                        "name": "Kui Buri Valley Sugarcane Estate #1",
                        "province": "Prachuap Khiri Khan",
                        "parkBoundary": "Kui Buri National Park",
                        "cropType": "SUGARCANE",
                        "growthStage": "READY_FOR_HARVEST",
                        "latitude": 12.0520,
                        "longitude": 99.6210,
                        "areaHectares": 14.5,
                        "distanceToForestEdgeM": 650.0,
                        "forestBearingDeg": 285.0,
                        "aromaVolatilityIndex": 0.85,
                        "maturityScore": 1.0,
                        "activeRiskScore": 0.88,
                        "alertTier": "TIER_3_DANGER",
                        "scentBufferMeters": 2800.0,
                        "updatedAt": "2026-10-14T08:00:00Z"
                    }
                ],
                "count": 1,
                "monitoredParks": ["Kui Buri National Park"]
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.getThailandCropPlots()
        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertEquals(1, data.count)
        assertEquals("th_plot_kb_01", data.plots[0].id)
        assertEquals("SUGARCANE", data.plots[0].cropType)
        assertEquals(ai.coexistshield.model.AlertTier.TIER_3_DANGER, data.plots[0].alertTier)
    }

    @Test
    fun testUpdateCropHarvestStageApi() = runTest {
        val mockJsonResponse = """
            {
                "id": "th_plot_kb_01",
                "name": "Kui Buri Valley Sugarcane Estate #1",
                "province": "Prachuap Khiri Khan",
                "parkBoundary": "Kui Buri National Park",
                "cropType": "SUGARCANE",
                "growthStage": "HARVESTED_CLEARED",
                "latitude": 12.0520,
                "longitude": 99.6210,
                "areaHectares": 14.5,
                "distanceToForestEdgeM": 650.0,
                "forestBearingDeg": 285.0,
                "aromaVolatilityIndex": 0.85,
                "maturityScore": 0.05,
                "activeRiskScore": 0.12,
                "alertTier": "SAFE",
                "scentBufferMeters": 400.0,
                "updatedAt": "2026-10-14T08:05:00Z"
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.updateCropHarvestStage("th_plot_kb_01", "HARVESTED_CLEARED")
        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertEquals("HARVESTED_CLEARED", data.growthStage)
        assertEquals(ai.coexistshield.model.AlertTier.SAFE, data.alertTier)
        assertEquals(0.12, data.activeRiskScore)
    }
}




