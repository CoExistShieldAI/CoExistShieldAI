package ai.coexistshield.edge

import ai.coexistshield.model.AlertTier
import ai.coexistshield.model.GeoLocation
import ai.coexistshield.model.WildlifeSpecies
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class EdgeVisionEngineTest {

    @Test
    fun `TC-01 decodeYoloV8Tensor suppresses duplicate overlapping anchors via NMS`() {
        val candidates = listOf(
            EdgeDetectionBox(0, EdgeObjectClass.BEAR, "Ursus thibetanus", "Asian Black Bear", 0.948, 0.22, 0.30, 0.55, 0.32),
            EdgeDetectionBox(0, EdgeObjectClass.BEAR, "Ursus thibetanus", "Asian Black Bear", 0.850, 0.23, 0.31, 0.54, 0.31),
            EdgeDetectionBox(3, EdgeObjectClass.WILD_BOAR, "Sus scrofa", "Wild Boar", 0.120, 0.10, 0.10, 0.20, 0.20),
            EdgeDetectionBox(7, EdgeObjectClass.PERSON_FACE, "Privacy Region (Face)", "Anonymized Face", 0.890, 0.05, 0.08, 0.12, 0.11)
        )

        val iou = EdgeVisionEngine.computeIoU(candidates[0], candidates[1])
        assertTrue(iou > 0.45)

        val decoded = EdgeVisionEngine.decodeYoloV8Tensor(candidates, confThreshold = 0.25, iouThreshold = 0.45)
        assertEquals(2, decoded.size)
        assertEquals("Ursus thibetanus", decoded[0].taxonomy)
        assertEquals(EdgeObjectClass.PERSON_FACE, decoded[1].objectClass)
    }

    @Test
    fun `TC-02 estimateOpticalRange calculates distance and uncertainty from bbox height and tilt`() {
        val range = EdgeVisionEngine.estimateOpticalRange(
            bboxHeightRatio = 0.3191,
            objectClass = EdgeObjectClass.BEAR,
            focalLengthMm = 26.0,
            tiltDegrees = -6.4
        )
        assertTrue(range.rangeMeters in 38..46)
        assertTrue(range.uncertaintyMeters >= 3)
        assertEquals(-6.4, range.tiltDegrees)
    }

    @Test
    fun `TC-03 evaluateKillSiteCachingHeuristic escalates when CARCASS leafCoverScore exceeds 0_65`() {
        val eval = EdgeVisionEngine.evaluateKillSiteCachingHeuristic(
            objectClass = EdgeObjectClass.CARCASS,
            leafCoverScore = 0.78
        )
        assertTrue(eval.cachingDetected)
        assertEquals(AlertTier.TIER_3_DANGER, eval.threatTier)
        assertEquals(500, eval.evacRadiusMeters)
        assertTrue(eval.rangerDispatchRequired)
        assertEquals("STEALTH_TACTICAL", eval.audioMode)
    }

    @Test
    fun `TC-04 evaluateKillSiteCachingHeuristic boundary at 0_65 does not trigger caching alert`() {
        val eval = EdgeVisionEngine.evaluateKillSiteCachingHeuristic(
            objectClass = EdgeObjectClass.CARCASS,
            leafCoverScore = 0.65
        )
        assertFalse(eval.cachingDetected)
        assertEquals(AlertTier.TIER_2_WARNING, eval.threatTier)
        assertEquals(250, eval.evacRadiusMeters)
        assertEquals("NORMAL", eval.audioMode)
    }

    @Test
    fun `TC-05 analyzeFrame produces complete EdgeVisionAnalysisResult with SHA-256 and GeoJSON`() {
        val res = EdgeVisionEngine.analyzeFrame(
            mode = EdgeCameraMode.KILL_SITE,
            speciesHint = WildlifeSpecies.BEAR,
            location = GeoLocation(39.7186, 140.1025),
            leafCoverScoreOverride = 0.78,
            incidentId = "INC-892"
        )
        assertEquals("yolov8n_wildlife_int8.tflite", res.modelName)
        assertTrue(res.inferenceLatencyMs < 120)
        assertTrue(res.privacyMaskApplied)
        assertEquals(2, res.blurredRegionsCount)
        assertEquals(64, res.sha256.length)
        assertTrue(res.geoJsonPayload.contains("\"incident_id\": \"INC-892\""))
        assertTrue(res.geoJsonPayload.contains("\"caching_detected\": true"))
        assertTrue(res.geoJsonPayload.contains("\"evac_radius_m\": 500"))
    }
}
