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
import kotlin.test.assertFalse
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

    @Test
    fun testTigerPasturesApi() = runTest {
        val mockJsonResponse = """
            {
                "pastures": [
                    {
                        "id": "ls_sumatra_01",
                        "name": "Bukit Kesuma Cattle Pasture",
                        "province": "Riau",
                        "village": "Desa Bukit Kesuma",
                        "livestockType": "CATTLE",
                        "headCount": 35,
                        "hasNightCorral": false,
                        "latitude": 0.1280,
                        "longitude": 101.4480,
                        "distanceToForestEdgeM": 320.0,
                        "isActiveKillSite": true,
                        "hoursSinceKill": 14.0,
                        "returnProbability": 0.643,
                        "activeThreatScore": 0.95,
                        "alertTier": "TIER_3_DANGER",
                        "scentBufferMeters": 3000.0,
                        "updatedAt": "2026-10-14T08:00:00Z"
                    }
                ],
                "count": 1,
                "updatedAt": "2026-10-14T08:00:00Z"
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.getTigerPastures()
        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertEquals(1, data.count)
        assertEquals("ls_sumatra_01", data.pastures[0].id)
        assertTrue(data.pastures[0].isActiveKillSite)
        assertEquals(ai.coexistshield.model.AlertTier.TIER_3_DANGER, data.pastures[0].alertTier)
    }

    @Test
    fun testReportTigerDepredationApi() = runTest {
        val mockJsonResponse = """
            {
                "enclosure": {
                    "id": "ls_sumatra_02",
                    "name": "Sungai Rawa Water Buffalo Corral",
                    "province": "Riau",
                    "village": "Dusun Sungai Rawa",
                    "livestockType": "WATER_BUFFALO",
                    "headCount": 20,
                    "hasNightCorral": true,
                    "latitude": 0.1320,
                    "longitude": 101.4580,
                    "distanceToForestEdgeM": 850.0,
                    "isActiveKillSite": true,
                    "hoursSinceKill": 0.0,
                    "returnProbability": 0.95,
                    "activeThreatScore": 0.95,
                    "alertTier": "TIER_3_DANGER",
                    "scentBufferMeters": 3000.0,
                    "updatedAt": "2026-10-14T08:10:00Z"
                },
                "alert": {
                    "alertId": "alert_tiger_12345",
                    "tier": "TIER_3_DANGER",
                    "killSiteLat": 0.1320,
                    "killSiteLon": 101.4580,
                    "villagesIn3kmRadius": ["Dusun Sungai Rawa", "Desa Bukit Kesuma"],
                    "returnWindowHoursRemaining": 72.0,
                    "canineForensicsConfirmed": true,
                    "causalReasoning": "Fresh WATER_BUFFALO kill cached. 72-hour tiger return feeding cycle active.",
                    "issuedAt": "2026-10-14T08:10:00Z"
                }
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.reportTigerDepredation(
            ReportDepredationRequestDto(
                enclosureId = "ls_sumatra_02",
                livestockSpecies = "WATER_BUFFALO",
                headCountLost = 1,
                latitude = 0.1320,
                longitude = 101.4580,
                hasNightCorral = true,
                canineBiteSpreadMm = 75.0
            )
        )

        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertTrue(data.enclosure.isActiveKillSite)
        assertEquals(ai.coexistshield.model.AlertTier.TIER_3_DANGER, data.alert.tier)
        assertTrue(data.alert.canineForensicsConfirmed)
        assertEquals(2, data.alert.villagesIn3kmRadius.size)
    }

    @Test
    fun testClearTigerCarcassApi() = runTest {
        val mockJsonResponse = """
            {
                "id": "ls_sumatra_01",
                "name": "Bukit Kesuma Cattle Pasture",
                "province": "Riau",
                "village": "Desa Bukit Kesuma",
                "livestockType": "CATTLE",
                "headCount": 35,
                "hasNightCorral": false,
                "latitude": 0.1280,
                "longitude": 101.4480,
                "distanceToForestEdgeM": 320.0,
                "isActiveKillSite": false,
                "hoursSinceKill": 0.0,
                "returnProbability": 0.0,
                "activeThreatScore": 0.65,
                "alertTier": "TIER_2_WARNING",
                "scentBufferMeters": 800.0,
                "updatedAt": "2026-10-14T08:15:00Z"
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.clearTigerCarcass("ls_sumatra_01")
        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertFalse(data.isActiveKillSite)
        assertEquals(0.0, data.returnProbability)
        assertEquals(ai.coexistshield.model.AlertTier.TIER_2_WARNING, data.alertTier)
    }

    @Test
    fun testRspoAuditPackageApi() = runTest {
        val mockJsonResponse = """
            {
                "auditId": "AUD-RSPO-2026-WM02",
                "plantationId": "wilmar_sumatra_02",
                "plantationName": "PT Wilmar Bio-Corridor Concession (Riau)",
                "province": "Riau",
                "reportingYear": 2026,
                "totalMonitoredHectares": 42500.0,
                "hcvBufferForestHectares": 8200.0,
                "humanFatalities": 0,
                "wildlifeFatalities": 0,
                "acousticDeterrentDeployments": 18,
                "bioAcousticDeflections": 42,
                "verifiedSafePassages": 342,
                "carcassRemovals": 7,
                "zeroHarmComplianceScore": 1.0,
                "rspoCertificateStatus": "CERTIFIED_ZERO_HARM_HCV",
                "sha256AuditSignature": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
                "generatedAt": "2026-10-14T08:00:00Z"
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.getRspoAuditPackage("wilmar_sumatra_02")
        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertEquals("AUD-RSPO-2026-WM02", data.auditId)
        assertEquals(0, data.humanFatalities)
        assertEquals(1.0, data.zeroHarmComplianceScore)
        assertEquals("CERTIFIED_ZERO_HARM_HCV", data.rspoCertificateStatus)
    }

    @Test
    fun testGcfBudgetImpactApi() = runTest {
        val mockJsonResponse = """
            {
                "municipalityCode": "03201",
                "municipalityName": "Morioka City (岩手県盛岡市)",
                "campaignTitle": "ふるさと納税 GCF: 通学路・住宅街の放置柿緊急伐採＆ツキノワグマ遭遇ゼロプロジェクト",
                "totalFundingGoalYen": 15000000,
                "currentRaisedYen": 11850000,
                "fundingProgressPct": 79.0,
                "donorCount": 482,
                "costPerTreeYen": 35000,
                "targetTreesCount": 420,
                "completedFelledTreesCount": 214,
                "remainingTreesCount": 206,
                "schoolChildrenProtectedCount": 3420,
                "quantifiedBearEncounterRiskReduction": 64.8,
                "furusatoTaxPortalUrl": "https://www.furusato-tax.jp/gcf/03201-bear-persimmon",
                "recentFellingLogs": [
                    {
                        "logId": "log_fell_01",
                        "treeId": "kt_0710",
                        "permitId": "PERMIT-MRK-2026-0710",
                        "officerId": "OFFICER-IWATE-42",
                        "action": "FELLED",
                        "schoolDistanceM": 85,
                        "riskDropPct": 18.5,
                        "biomassKg": 450.0,
                        "timestamp": "2026-10-12T04:00:00Z"
                    }
                ],
                "updatedAt": "2026-10-14T08:00:00Z"
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.getGcfBudgetImpact("03201")
        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertEquals("03201", data.municipalityCode)
        assertEquals(15000000L, data.totalFundingGoalYen)
        assertEquals(214, data.completedFelledTreesCount)
        assertEquals(64.8, data.quantifiedBearEncounterRiskReduction)
        assertEquals(1, data.recentFellingLogs.size)
    }

    @Test
    fun testPriorityKakiLayerApi() = runTest {
        val mockJsonResponse = """
            {
                "municipalityCode": "03201",
                "municipalityName": "Morioka City (岩手県盛岡市)",
                "totalCatalogedTrees": 840,
                "top5PercentCutoff": 0.89,
                "priorityTrees": [
                    {
                        "treeId": "kt_1120",
                        "address": "盛岡市高松3丁目12-4",
                        "latitude": 39.7125,
                        "longitude": 141.1620,
                        "hazardScore": 0.95,
                        "isTop5Percent": true,
                        "distanceToSchoolRouteMeters": 120,
                        "estimatedBearAttractantRadiusM": 800,
                        "ownerConsentStatus": "GRANTED",
                        "permitId": "PERMIT-MRK-2026-1120",
                        "managementStatus": "UNMANAGED",
                        "recommendedClearanceOrder": 1,
                        "assignedSquad": "Morioka Chōjū Higai Squad 3",
                        "fruitLoad": "ABUNDANT_FALLING"
                    }
                ],
                "activeBaselineThreat": 0.91,
                "updatedAt": "2026-10-14T08:00:00Z"
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.getPriorityKakiLayer("03201")
        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertEquals(840, data.totalCatalogedTrees)
        assertEquals(1, data.priorityTrees.size)
        assertEquals("kt_1120", data.priorityTrees[0].treeId)
        assertEquals("GRANTED", data.priorityTrees[0].ownerConsentStatus)
    }

    @Test
    fun testLogMunicipalKakiActionApi() = runTest {
        val mockJsonResponse = """
            {
                "auditLogId": "log_fell_999",
                "updatedTree": {
                    "treeId": "kt_1120",
                    "address": "盛岡市高松3丁目12-4",
                    "latitude": 39.7125,
                    "longitude": 141.1620,
                    "hazardScore": 0.0,
                    "isTop5Percent": true,
                    "distanceToSchoolRouteMeters": 120,
                    "estimatedBearAttractantRadiusM": 800,
                    "ownerConsentStatus": "GRANTED",
                    "permitId": "PERMIT-MRK-2026-1120",
                    "managementStatus": "FELLED",
                    "recommendedClearanceOrder": 1,
                    "assignedSquad": "Morioka Chōjū Higai Squad 3",
                    "fruitLoad": "ABUNDANT_FALLING"
                },
                "treeRiskDropPct": 100.0,
                "newMunicipalityRiskScore": 0.60,
                "gcfCompletedFelledTrees": 215,
                "gcfRiskReductionPct": 65.2,
                "message": "Municipal Action Logged: Tree kt_1120 FELLED successfully",
                "processedAt": "2026-10-14T08:20:00Z"
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(mockJsonResponse)
        val repository = GeofenceRepository(httpClient)

        val result = repository.logMunicipalKakiAction(
            MunicipalKakiActionRequestDto(
                treeId = "kt_1120",
                action = "FELLED",
                officerId = "OFFICER-IWATE-42",
                officerRole = "OFFICIAL",
                permitNumber = "PERMIT-MRK-2026-1120",
                biomassDisposedKg = 480.0
            )
        )

        assertTrue(result.isSuccess)
        val data = result.getOrNull()!!
        assertEquals("FELLED", data.updatedTree.managementStatus)
        assertEquals(100.0, data.treeRiskDropPct)
        assertEquals(215, data.gcfCompletedFelledTrees)
    }
}




