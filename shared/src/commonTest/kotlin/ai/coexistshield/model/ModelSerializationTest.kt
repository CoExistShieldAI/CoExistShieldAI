package ai.coexistshield.model

import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class ModelSerializationTest {

    private val json = Json {
        prettyPrint = true
        ignoreUnknownKeys = true
    }

    @Test
    fun testGeofenceSerialization() {
        val geofence = GeofenceZone(
            id = "geo_001",
            tier = AlertTier.TIER_2_WARNING,
            targetSpecies = listOf(WildlifeSpecies.BEAR),
            coordinates = listOf(
                GeoLocation(38.25, 140.31),
                GeoLocation(38.27, 140.33)
            ),
            centroid = GeoLocation(38.26, 140.32),
            radiusMeters = 1200.0,
            causalExplanation = "Tier 2 Warning: New Moon night + westerly wind blowing toward unharvested persimmon grove 600m away.",
            primaryAttractantRef = "kaki_trees/kt_001",
            validFrom = "2026-10-14T00:00:00Z",
            validUntil = "2026-10-14T06:00:00Z",
            generatedAt = "2026-10-14T00:00:00Z"
        )

        val serialized = json.encodeToString(geofence)
        assertTrue(serialized.contains("TIER_2_WARNING"))
        assertTrue(serialized.contains("BEAR"))

        val deserialized = json.decodeFromString<GeofenceZone>(serialized)
        assertEquals(geofence.id, deserialized.id)
        assertEquals(geofence.tier, deserialized.tier)
        assertEquals(geofence.radiusMeters, deserialized.radiusMeters)
    }

    @Test
    fun testKakiTreeTop5PercentSerialization() {
        val tree = KakiTree(
            id = "kt_001",
            location = GeoLocation(38.261, 140.319),
            fruitType = FruitType.PERSIMMON,
            fruitLoadLevel = FruitLoadLevel.ABUNDANT,
            ripenessStage = RipenessStage.OVERRIPE_FALLING,
            isTop5PercentHazard = true,
            hazardScore = 0.95,
            propertyContext = PropertyContextTag.VACANT_HOUSE,
            isNearSchoolRoute = true,
            status = TreeManagementStatus.UNMANAGED,
            gcfCampaignId = "gcf_iwate_01",
            reportedBy = "citizen_01",
            reportedAt = "2026-10-14T00:10:00Z"
        )

        val serialized = json.encodeToString(tree)
        assertTrue(serialized.contains("ABUNDANT"))
        assertTrue(serialized.contains("VACANT_HOUSE"))

        val deserialized = json.decodeFromString<KakiTree>(serialized)
        assertEquals(true, deserialized.isTop5PercentHazard)
        assertEquals(0.95, deserialized.hazardScore)
    }

    @Test
    fun testLivestockKillSiteCachingSerialization() {
        val pasture = LivestockPasture(
            id = "ls_001",
            location = GeoLocation(-0.521, 101.442),
            livestockType = LivestockType.CATTLE,
            headCount = 30,
            hasNightCorral = false,
            isKillSite = true,
            killReportedAt = "2026-10-13T22:00:00Z",
            cachingRiskExpiry = "2026-10-16T22:00:00Z",
            activeThreatTier = AlertTier.TIER_3_DANGER
        )

        val serialized = json.encodeToString(pasture)
        val deserialized = json.decodeFromString<LivestockPasture>(serialized)
        assertEquals(true, deserialized.isKillSite)
        assertEquals(AlertTier.TIER_3_DANGER, deserialized.activeThreatTier)
    }
}
