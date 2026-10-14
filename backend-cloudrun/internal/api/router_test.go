package api

import (
	"bytes"
	"context"
	"encoding/json"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai.coexistshield/backend/internal/auth"
	"ai.coexistshield/backend/internal/coldstart"
	"ai.coexistshield/backend/internal/compliance"
	"ai.coexistshield/backend/internal/corroboration"
	"ai.coexistshield/backend/internal/elephant"
	"ai.coexistshield/backend/internal/gemini"
	"ai.coexistshield/backend/internal/geofence"
	"ai.coexistshield/backend/internal/memory"
	"ai.coexistshield/backend/internal/ml"
	"ai.coexistshield/backend/internal/model"
	"ai.coexistshield/backend/internal/repository"
	"ai.coexistshield/backend/internal/tiger"
)

func setupTestServer() *Router {
	geminiClient := gemini.NewClientWithConfig(gemini.Config{
		MockEnabled: true,
	})
	geofenceEngine := geofence.NewEngine()
	memoryBank := memory.NewBank()
	mlPipeline := ml.NewRetrainingPipeline()
	mlPipeline.LoadSamples(ml.GenerateSeedTrainingDataset())
	elephantEngine := elephant.NewEngine()
	tigerEngine := tiger.NewEngine()
	complianceEngine := compliance.NewEngine()
	corroborationEngine := corroboration.NewEngine(geminiClient)
	corroborationEngine.LoadFixtureStreams()
	return NewRouter(geminiClient, geofenceEngine, memoryBank, mlPipeline, elephantEngine, tigerEngine, complianceEngine, corroborationEngine)
}

func TestHealthz(t *testing.T) {
	router := setupTestServer()
	for _, path := range []string{"/healthz", "/health", "/api/v1/health", "/"} {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("path %s: expected 200 OK, got %d", path, w.Code)
		}

		var body map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("path %s: failed to parse json response: %v", path, err)
		}

		if body["status"] != "healthy" {
			t.Errorf("path %s: expected status healthy, got %s", path, body["status"])
		}
	}
}

func TestForensicTriageEndpoint(t *testing.T) {
	router := setupTestServer()

	reqBody := model.ForensicTriageRequest{
		ImageBase64:  "mockBase64Data",
		CategoryHint: "ATTRACTANT_FRUIT_TREE",
		Latitude:     38.256,
		Longitude:    140.321,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/forensics/triage", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp model.ForensicTriageResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.DetectedSpecies != "Diospyros kaki" {
		t.Errorf("expected Diospyros kaki, got %s", resp.DetectedSpecies)
	}
	if resp.Confidence < 0.90 {
		t.Errorf("expected confidence >= 0.90, got %f", resp.Confidence)
	}
	if resp.Attributes == nil || resp.Attributes.FruitLoadLevel != "ABUNDANT" {
		t.Errorf("expected fruitLoadLevel ABUNDANT, got %+v", resp.Attributes)
	}
}

func TestForensicTriageFailsWhenMockDisabledWithoutKey(t *testing.T) {
	geminiClient := gemini.NewClientWithConfig(gemini.Config{
		ApiKey:      "",
		MockEnabled: false,
	})
	router := NewRouter(
		geminiClient,
		geofence.NewEngine(),
		memory.NewBank(),
		ml.NewRetrainingPipeline(),
		elephant.NewEngine(),
		tiger.NewEngine(),
		compliance.NewEngine(),
		corroboration.NewEngine(geminiClient),
	)

	reqBody := model.ForensicTriageRequest{
		ImageBase64:  "data:image/jpeg;base64,mockImagePayload",
		CategoryHint: "ATTRACTANT_FRUIT_TREE",
		Latitude:     39.712,
		Longitude:    141.158,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/forensics/triage", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 InternalServerError when mock is disabled, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "gemini API key is not configured") {
		t.Errorf("expected error message about unconfigured API key, got %s", w.Body.String())
	}
}

func TestTopHazardsPersimmonTrees(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/analytics/japan-kaki/top-hazards?municipalityCode=03201&limit=2", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp model.TopHazardsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.TotalCatalogedTrees != 840 {
		t.Errorf("expected 840 trees, got %d", resp.TotalCatalogedTrees)
	}
	if len(resp.PriorityTrees) != 2 {
		t.Errorf("expected 2 priority trees due to limit=2, got %d", len(resp.PriorityTrees))
	}
	if resp.PriorityTrees[0].TreeId != "kt_1120" {
		t.Errorf("expected kt_1120, got %s", resp.PriorityTrees[0].TreeId)
	}
}

func TestGeofencesRecalculateEndpoint(t *testing.T) {
	router := setupTestServer()

	reqBody := model.GeofenceRecalculateRequest{
		TriggerEntityRef:     "kaki_trees/kt_1120",
		TargetCenterLatitude: 39.702,
		TargetCenterLongitude: 141.152,
		RadiusKm:             5.0,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/geofences/recalculate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp model.GeofenceRecalculateResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.RecalculatedGeofencesCount < 3 {
		t.Errorf("expected at least 3 recalculated geofences, got %d", resp.RecalculatedGeofencesCount)
	}
}

func TestRspoExportEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/compliance/rspo/export?plantationId=sumatra_test_01", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "text/csv" {
		t.Errorf("expected text/csv content type, got %s", w.Header().Get("Content-Type"))
	}
}

func TestScentPlumeGeofencesEndpoint(t *testing.T) {
	router := setupTestServer()

	reqBody := model.GeneratePlumeGeofencesRequest{
		Weather: &model.WeatherConditions{
			WindSpeedKmh:     15.0,
			WindDirectionDeg: 270.0, // Westerly wind (blowing towards East: bearing 90°)
			MoonIllumination: 0.05,  // Near New Moon
		},
		Attractants: []model.AttractantHub{
			{
				Id:           "kt_01",
				Type:         "KAKI_TREE",
				Latitude:     38.2560,
				Longitude:    140.3210,
				FruitLoad:    "ABUNDANT",
				IsTop5Hazard: true,
			},
			{
				Id:              "rk_01",
				Type:            "ROADKILL",
				Latitude:        38.2600,
				Longitude:       140.3250,
				HoursSinceDeath: 4.0,
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/geofences/scent-plume", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp model.GeneratePlumeGeofencesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Count != 2 {
		t.Fatalf("expected 2 zones generated, got %d", resp.Count)
	}

	// Verify RFC 7946 Polygon geometry
	zone1 := resp.Zones[0]
	if zone1.PolygonGeoJson == nil || zone1.PolygonGeoJson.Type != "Polygon" {
		t.Fatalf("expected Polygon geometry, got %+v", zone1.PolygonGeoJson)
	}
	coords := zone1.PolygonGeoJson.Coordinates[0]
	if len(coords) < 30 {
		t.Fatalf("expected at least 30 boundary vertices, got %d", len(coords))
	}

	// First and last vertex must be identical for valid GeoJSON LinearRing
	first := coords[0]
	last := coords[len(coords)-1]
	if first[0] != last[0] || first[1] != last[1] {
		t.Errorf("LinearRing not closed: first [%f, %f], last [%f, %f]", first[0], first[1], last[0], last[1])
	}

	// Check downwind elongation:
	// Westerly wind blows towards East (+longitude). Max longitude should be extended further east from center
	centerLon := 140.3210
	maxLon := -180.0
	minLon := 180.0
	for _, pt := range coords {
		if pt[0] > maxLon {
			maxLon = pt[0]
		}
		if pt[0] < minLon {
			minLon = pt[0]
		}
	}

	eastSpread := maxLon - centerLon
	westSpread := centerLon - minLon

	if eastSpread <= westSpread {
		t.Errorf("scent plume should elongate eastward (downwind). eastSpread=%f, westSpread=%f", eastSpread, westSpread)
	}
}

func TestMemorySearchEndpoint(t *testing.T) {
	router := setupTestServer()

	reqBody := model.MemorySearchRequest{
		QueryText:     "Asian Black Bear incursion near unharvested persimmon trees in Iwate valley with West wind",
		Region:        "JAPAN",
		Species:       "BEAR",
		MinSimilarity: 0.50,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/memory/search", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp model.MemorySearchResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.QueryVectorGenerated {
		t.Errorf("expected QueryVectorGenerated to be true")
	}
	if len(resp.Matches) == 0 {
		t.Fatalf("expected at least 1 historical analog match, got 0")
	}

	primary := resp.Matches[0]
	if primary.Species != "BEAR" || primary.Region != "JAPAN" {
		t.Errorf("expected BEAR in JAPAN, got %s in %s", primary.Species, primary.Region)
	}
	if resp.ElevatedTier == "" {
		t.Errorf("expected autonomous elevated tier recommendation")
	}
}

func TestAlertExplanationEndpoint(t *testing.T) {
	router := setupTestServer()

	reqBody := model.CausalAlertExplanationRequest{
		Tier:          "TIER_1_NOTICE",
		TargetSpecies: "Asian Black Bear",
		Region:        "JAPAN",
		Weather: map[string]interface{}{
			"windSpeedKmh":     14.0,
			"windDirection":    "W",
			"moonIllumination": 0.02,
		},
		NearbyAttractants: []map[string]interface{}{
			{
				"type":      "abandoned_persimmon",
				"distanceM": 620,
				"fruitLoad": "ABUNDANT",
			},
		},
		RetrieveMemory: true,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/alerts/explain", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp model.CausalAlertExplanationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Explanation == "" {
		t.Errorf("expected non-empty explanation")
	}
	if len(resp.HistoricalAnalogs) == 0 {
		t.Errorf("expected retrieved historical analogs")
	}
	if resp.EffectiveTier == "" {
		t.Errorf("expected effective tier")
	}
	if len(resp.DispatchRecommendations) == 0 {
		t.Errorf("expected non-empty dispatch recommendations")
	} else {
		rec := resp.DispatchRecommendations[0]
		if rec.ActionType == "" || rec.Priority == "" || rec.Justification == "" {
			t.Errorf("expected populated dispatch recommendation, got %+v", rec)
		}
	}
}

func TestMlStatusEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/ml/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse json: %v", err)
	}

	if resp["status"] != "HEALTHY" {
		t.Errorf("expected status HEALTHY, got %v", resp["status"])
	}
	if resp["activeModel"] == nil {
		t.Errorf("expected activeModel to be present")
	}
}

func TestMlRetrainEndpoint(t *testing.T) {
	router := setupTestServer()

	reqBody := MlRetrainApiRequest{
		Trigger: "test_manual_admin",
		Force:   false,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/ml/retrain", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result ml.RetrainResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode RetrainResult: %v", err)
	}

	if !result.GateCheckPassed {
		t.Errorf("expected validation gate checks to pass, notes: %s", result.GateEvaluationNotes)
	}
	if result.CandidateModel.Metrics.AucRoc < 0.82 {
		t.Errorf("expected AUC-ROC >= 0.82, got %.3f", result.CandidateModel.Metrics.AucRoc)
	}
}

func TestMlPredictEndpoint(t *testing.T) {
	router := setupTestServer()

	reqBody := MlPredictApiRequest{
		Latitude:               39.7020,
		Longitude:              141.1520,
		DistNearestPersimmonM:  180.0,
		DistNearestCarcassM:    350.0,
		DistNearestCaneFruitM:  2500.0,
		WindScentAlignmentDeg:  15.0,
		MoonDarknessPct:        90.0,
		PrecipitationMm24h:      0.0,
		HistoricalIncidents30d: 5.0,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/ml/predict", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var pred ml.PredictResult
	if err := json.Unmarshal(w.Body.Bytes(), &pred); err != nil {
		t.Fatalf("failed to decode PredictResult: %v", err)
	}

	if pred.H3Index == "" {
		t.Errorf("expected valid H3 index in prediction result")
	}
	if pred.IncursionProbability < 0.55 {
		t.Errorf("expected high hazard probability >= 0.55, got %.3f", pred.IncursionProbability)
	}
	if pred.AlertTier != "TIER_2_WARNING" && pred.AlertTier != "TIER_3_DANGER" {
		t.Errorf("expected TIER_2_WARNING or TIER_3_DANGER, got %s", pred.AlertTier)
	}
}

func TestMlQueryEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/ml/query", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	sql, ok := resp["query"].(string)
	if !ok || len(sql) == 0 {
		t.Fatalf("expected non-empty SQL query string")
	}
	if !strings.Contains(sql, "h3_geo_to_h3") {
		t.Errorf("expected query to contain h3_geo_to_h3")
	}
}

func TestBigQueryStreamSyncEndpoint(t *testing.T) {
	router := setupTestServer()

	event := ml.IncidentStreamingEvent{
		EventID:           "sync_evt_test_01",
		IncidentID:        "inc_test_01",
		Species:           "BEAR",
		Category:          "WILDLIFE_SIGHTING",
		Latitude:          39.7020,
		Longitude:         141.1520,
		Timestamp:         time.Now().UTC(),
		IsIncursion:       1,
		VerificationState: "VERIFIED",
	}
	bodyBytes, _ := json.Marshal(event)

	req := httptest.NewRequest("POST", "/api/v1/bigquery/stream-sync", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	syncedCount, ok := resp["syncedEventsCount"].(float64)
	if !ok || syncedCount < 1 {
		t.Errorf("expected syncedEventsCount >= 1, got %v", resp["syncedEventsCount"])
	}

	// Verify empty body is rejected with 400 Bad Request
	emptyReq := httptest.NewRequest("POST", "/api/v1/bigquery/stream-sync", bytes.NewReader([]byte("")))
	emptyReq.Header.Set("Content-Type", "application/json")
	emptyW := httptest.NewRecorder()
	router.ServeHTTP(emptyW, emptyReq)
	if emptyW.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request on empty body, got %d", emptyW.Code)
	}
}

func TestBigQuerySyncStatusEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/bigquery/sync-status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var status ml.BigQuerySyncStatus
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("failed to decode status: %v", err)
	}

	if status.Status != "ACTIVE_STREAMING" {
		t.Errorf("expected ACTIVE_STREAMING, got %s", status.Status)
	}
	if status.SlaTargetSeconds != 60 {
		t.Errorf("expected 60s SLA, got %d", status.SlaTargetSeconds)
	}
	if status.ReplicationLagSeconds > 60.0 {
		t.Errorf("replication lag exceeds SLA: %f", status.ReplicationLagSeconds)
	}
}

func TestBigQueryStagedSamplesEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/bigquery/staged-samples", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if _, ok := resp["samples"]; !ok {
		t.Errorf("expected 'samples' field in response")
	}
}

func TestThailandCropPlotsEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/analytics/thailand-hec/crop-plots", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var resp struct {
		Plots          []elephant.ElephantCropPlot `json:"plots"`
		Count          int                         `json:"count"`
		MonitoredParks []string                    `json:"monitoredParks"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Count < 4 {
		t.Errorf("expected at least 4 plots, got %d", resp.Count)
	}
	if len(resp.MonitoredParks) != 2 {
		t.Errorf("expected 2 monitored national parks, got %d", len(resp.MonitoredParks))
	}
}

func TestThailandHecRecalculateEndpoint(t *testing.T) {
	router := setupTestServer()

	reqBody := ThailandHecRecalculateRequest{
		Weather: &model.WeatherConditions{
			WindSpeedKmh:     20.0,
			WindDirectionDeg: 90.0, // Blowing towards West into Kui Buri forest
			MoonIllumination: 0.05, // New Moon
			PrecipitationMmH: 0.0,
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/analytics/thailand-hec/recalculate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Plots []elephant.ElephantCropPlot `json:"plots"`
		Count int                         `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Count < 4 {
		t.Errorf("expected at least 4 plots, got %d", resp.Count)
	}
}

func TestThailandPlotActionEndpoint(t *testing.T) {
	router := setupTestServer()

	reqBody := UpdateHarvestStageRequest{
		GrowthStage: "HARVESTED_CLEARED",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/analytics/thailand-hec/plots/th_plot_kb_01/harvest-stage", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var updated elephant.ElephantCropPlot
	if err := json.Unmarshal(w.Body.Bytes(), &updated); err != nil {
		t.Fatalf("failed to decode updated plot: %v", err)
	}

	if updated.GrowthStage != "HARVESTED_CLEARED" {
		t.Errorf("expected HARVESTED_CLEARED stage, got %s", updated.GrowthStage)
	}
	if updated.ActiveRiskScore > 0.40 {
		t.Errorf("expected risk score <= 0.40 after clearing, got %.3f", updated.ActiveRiskScore)
	}
}

func TestTigerPasturesEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/analytics/sumatra-tiger/pastures", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Pastures []tiger.LivestockEnclosure `json:"pastures"`
		Count    int                        `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Count < 3 {
		t.Errorf("expected at least 3 enclosures, got %d", resp.Count)
	}
}

func TestTigerDepredationReportEndpoint(t *testing.T) {
	router := setupTestServer()

	reqBody := TigerDepredationReportRequestBody{
		ReportDepredationRequest: tiger.ReportDepredationRequest{
			EnclosureID:           "ls_sumatra_02",
			LivestockSpecies:      "WATER_BUFFALO",
			HeadCountLost:         1,
			Latitude:              0.1320,
			Longitude:             101.4580,
			HasNightCorral:        true,
			DragTrailLengthMeters: 45.0,
			CanineBiteSpreadMm:    74.0, // Valid Sumatran tiger canine distance
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/analytics/sumatra-tiger/depredation/report", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Enclosure tiger.LivestockEnclosure `json:"enclosure"`
		Alert     tiger.TigerDepredationAlert `json:"alert"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Enclosure.IsActiveKillSite {
		t.Errorf("expected enclosure to be active kill site")
	}
	if resp.Alert.Tier != "TIER_3_DANGER" {
		t.Errorf("expected TIER_3_DANGER alert, got %s", resp.Alert.Tier)
	}
	if !resp.Alert.CanineForensicsConfirmed {
		t.Errorf("expected canine forensics confirmed for 74mm bite spread")
	}
	if len(resp.Alert.VillagesIn3kmRadius) == 0 {
		t.Errorf("expected at least one village in 3km radius alert")
	}
}

func TestTigerRecalculateEndpoint(t *testing.T) {
	router := setupTestServer()

	reqBody := TigerRecalculateRequest{
		Weather: &model.WeatherConditions{
			WindSpeedKmh:     15.0,
			WindDirectionDeg: 180.0,
			MoonIllumination: 0.05,
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/analytics/sumatra-tiger/recalculate", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Pastures []tiger.LivestockEnclosure `json:"pastures"`
		Count    int                        `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Count < 3 {
		t.Errorf("expected at least 3 pastures, got %d", resp.Count)
	}
}

func TestTigerPastureClearEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("POST", "/api/v1/analytics/sumatra-tiger/pastures/ls_sumatra_01/clear", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var cleared tiger.LivestockEnclosure
	if err := json.Unmarshal(w.Body.Bytes(), &cleared); err != nil {
		t.Fatalf("failed to decode cleared pasture: %v", err)
	}

	if cleared.IsActiveKillSite {
		t.Errorf("expected isActiveKillSite to be false after clear")
	}
	if cleared.ReturnProbability > 0.0 {
		t.Errorf("expected return probability 0.0 after clear, got %.2f", cleared.ReturnProbability)
	}
}

func TestRspoPackageEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/compliance/rspo/package?plantationId=wilmar_sumatra_02", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var pkg compliance.RspoAuditPackage
	if err := json.Unmarshal(w.Body.Bytes(), &pkg); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if pkg.PlantationID != "wilmar_sumatra_02" {
		t.Errorf("expected wilmar_sumatra_02, got %s", pkg.PlantationID)
	}
	if pkg.ZeroHarmComplianceScore != 1.0 {
		t.Errorf("expected 1.0 score, got %.2f", pkg.ZeroHarmComplianceScore)
	}
}

func TestRspoExportCsvEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/compliance/rspo/export?plantationId=wilmar_sumatra_02", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/csv") {
		t.Errorf("expected text/csv Content-Type, got %s", contentType)
	}

	csv := w.Body.String()
	if !strings.Contains(csv, "wilmar_sumatra_02") {
		t.Errorf("expected plantation id in csv output: %s", csv)
	}
}

func TestGcfBudgetImpactEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/compliance/gcf/budget-impact?municipalityCode=03201", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp compliance.GcfBudgetImpactResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.MunicipalityCode != "03201" {
		t.Errorf("expected 03201, got %s", resp.MunicipalityCode)
	}
	if resp.CurrentRaisedYen < 10000000 {
		t.Errorf("expected >= 10M yen raised, got %d", resp.CurrentRaisedYen)
	}
	if resp.CostPerTreeYen != 35000 {
		t.Errorf("expected 35000 yen cost per tree, got %d", resp.CostPerTreeYen)
	}
}

func TestPriorityKakiLayerEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/municipal/kaki/priority-layer?municipalityCode=03201", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp compliance.PriorityKakiLayerResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.PriorityTrees) < 3 {
		t.Errorf("expected at least 3 priority trees, got %d", len(resp.PriorityTrees))
	}
	if resp.PriorityTrees[0].TreeID != "kt_1120" {
		t.Errorf("expected kt_1120 first, got %s", resp.PriorityTrees[0].TreeID)
	}
}

func TestMunicipalKakiActionEndpoint(t *testing.T) {
	router := setupTestServer()

	// 1. Citizen role rejected with 403 Forbidden
	citizenBody, _ := json.Marshal(compliance.MunicipalKakiActionRequest{
		TreeID:       "kt_1120",
		Action:       "FELLED",
		OfficerID:    "citizen_user_1",
		OfficerRole:  "CITIZEN",
		PermitNumber: "UNAUTHORIZED",
	})
	req := httptest.NewRequest("POST", "/api/v1/municipal/kaki/action", bytes.NewReader(citizenBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for CITIZEN role, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 2. Official role approved with 200 OK and risk reduction
	officialBody, _ := json.Marshal(compliance.MunicipalKakiActionRequest{
		TreeID:            "kt_1120",
		Action:            "FELLED",
		OfficerID:         "OFFICER-IWATE-42",
		OfficerRole:       "OFFICIAL",
		PermitNumber:      "PERMIT-MRK-2026-1120",
		BiomassDisposedKg: 480.0,
		Notes:             "Felled and chipped near school route",
	})
	req2 := httptest.NewRequest("POST", "/api/v1/municipal/kaki/action", bytes.NewReader(officialBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for OFFICIAL role, got %d. Body: %s", w2.Code, w2.Body.String())
	}

	var resp compliance.MunicipalKakiActionResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.UpdatedTree.ManagementStatus != "FELLED" {
		t.Errorf("expected FELLED status, got %s", resp.UpdatedTree.ManagementStatus)
	}
	if resp.TreeRiskDropPct != 100.0 {
		t.Errorf("expected 100%% risk drop, got %.1f", resp.TreeRiskDropPct)
	}
}

func TestUnstructuredNlpIngestEndpoint(t *testing.T) {
	router := setupTestServer()

	nlpBody, _ := json.Marshal(model.UnstructuredNlpIngestRequest{
		RawText:       "盛岡市防災無線：松ヶ枝町付近でクマ目撃。柿の木の近くを徘徊中。児童は通学路を避けてください。",
		SourceChannel: "MUNICIPAL_DISASTER_RADIO",
		RegionHint:    "JAPAN",
	})

	req := httptest.NewRequest("POST", "/api/v1/nlp/ingest/unstructured", bytes.NewReader(nlpBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp model.UnstructuredNlpIngestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ModelUsed != "gemini-3.8-flash" {
		t.Errorf("expected gemini-3.8-flash, got %s", resp.ModelUsed)
	}
	if resp.Entity.Species != "BEAR" {
		t.Errorf("expected species BEAR, got %s", resp.Entity.Species)
	}
	if resp.Entity.UrgencyLevel != "HIGH" {
		t.Errorf("expected urgency HIGH, got %s", resp.Entity.UrgencyLevel)
	}
	if resp.Entity.Confidence < 0.90 {
		t.Errorf("expected confidence >= 0.90, got %f", resp.Entity.Confidence)
	}
}

func TestCorroborationStreamsAndEvaluateEndpoints(t *testing.T) {
	router := setupTestServer()

	// 1. GET /api/v1/corroboration/streams
	req1 := httptest.NewRequest("GET", "/api/v1/corroboration/streams?region=JAPAN", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w1.Code, w1.Body.String())
	}

	var streams []model.CausalStreamItem
	if err := json.Unmarshal(w1.Body.Bytes(), &streams); err != nil {
		t.Fatalf("failed to parse streams: %v", err)
	}
	if len(streams) < 2 {
		t.Errorf("expected >= 2 streams, got %d", len(streams))
	}

	// 2. POST /api/v1/corroboration/evaluate
	evalBody, _ := json.Marshal(model.CorroborationRequest{
		TargetLatitude:  39.7030,
		TargetLongitude: 141.1530,
		RadiusKm:        4.0,
		Region:          "JAPAN",
	})

	req2 := httptest.NewRequest("POST", "/api/v1/corroboration/evaluate", bytes.NewReader(evalBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w2.Code, w2.Body.String())
	}

	var evalResp model.CorroborationResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &evalResp); err != nil {
		t.Fatalf("failed to parse eval response: %v", err)
	}
	if evalResp.ModelUsed != "gemini-3.8-flash" {
		t.Errorf("expected gemini-3.8-flash, got %s", evalResp.ModelUsed)
	}
	if len(evalResp.Clusters) == 0 {
		t.Fatalf("expected at least 1 corroborated cluster")
	}
	if evalResp.Clusters[0].CombinedConfidence < 0.90 {
		t.Errorf("expected combined confidence >= 0.90, got %f", evalResp.Clusters[0].CombinedConfidence)
	}

	// 3. POST /api/v1/corroboration/poll-cycle
	req3 := httptest.NewRequest("POST", "/api/v1/corroboration/poll-cycle", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w3.Code)
	}
}

func TestGetActiveGeofencesEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/geofences/active?lat=39.7020&long=141.1520&radiusKm=20.0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var zones []model.ActiveGeofenceZone
	if err := json.Unmarshal(w.Body.Bytes(), &zones); err != nil {
		t.Fatalf("failed to decode zones: %v", err)
	}

	if len(zones) == 0 {
		t.Errorf("expected active geofence zones, got 0")
	}
	if zones[0].Tier == "" || zones[0].RadiusMeters <= 0 {
		t.Errorf("expected valid zone properties, got %+v", zones[0])
	}
}

func TestGetCurrentMeteorologyEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/meteorology/current?region=JAPAN", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var meteo model.MeteorologicalConditions
	if err := json.Unmarshal(w.Body.Bytes(), &meteo); err != nil {
		t.Fatalf("failed to decode meteorology: %v", err)
	}

	if meteo.MoonPhase.PhaseName == "" {
		t.Errorf("expected non-empty moon phase name")
	}
	if meteo.MoonPhase.NocturnalRiskMultiplier < 1.0 {
		t.Errorf("expected nocturnal risk multiplier >= 1.0, got %f", meteo.MoonPhase.NocturnalRiskMultiplier)
	}
	if meteo.WindSpeedKmh <= 0 {
		t.Errorf("expected positive wind speed, got %f", meteo.WindSpeedKmh)
	}
}

func TestGetIncidentsEndpoint(t *testing.T) {
	router := setupTestServer()

	// 1. Get all incidents
	req := httptest.NewRequest("GET", "/api/v1/incidents", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var incidents []model.IncidentDto
	if err := json.Unmarshal(w.Body.Bytes(), &incidents); err != nil {
		t.Fatalf("failed to decode incidents: %v", err)
	}
	if len(incidents) < 3 {
		t.Errorf("expected at least 3 seed incidents, got %d", len(incidents))
	}

	// 2. Filter by category
	req2 := httptest.NewRequest("GET", "/api/v1/incidents?category=ROADKILL", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w2.Code)
	}

	var roadkillIncidents []model.IncidentDto
	if err := json.Unmarshal(w2.Body.Bytes(), &roadkillIncidents); err != nil {
		t.Fatalf("failed to decode roadkill incidents: %v", err)
	}
	for _, inc := range roadkillIncidents {
		if inc.Category != "ROADKILL" {
			t.Errorf("expected category ROADKILL, got %s", inc.Category)
		}
	}
}

func TestSubmitIncidentEndpoint(t *testing.T) {
	router := setupTestServer()

	newIncident := model.IncidentDto{
		Category: "WILDLIFE_SIGHTING",
		Species:  "BEAR",
		Location: model.GeoLocationDto{
			Latitude:  39.7050,
			Longitude: 141.1510,
		},
		Region:       "JAPAN",
		Description:  "Large male bear seen crossing forest edge road.",
		ReporterId:   "usr_test_101",
		ReporterRole: "CITIZEN",
	}
	bodyBytes, _ := json.Marshal(newIncident)

	req := httptest.NewRequest("POST", "/api/v1/incidents", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	var created model.IncidentDto
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode created incident: %v", err)
	}

	if created.Id == "" {
		t.Errorf("expected generated incident ID")
	}
	if created.GeminiTriageResult == nil {
		t.Fatalf("expected Gemini triage result attached")
	}
	if created.GeminiTriageResult.ModelUsed != "gemini-3.8-flash" {
		t.Errorf("expected gemini-3.8-flash, got %s", created.GeminiTriageResult.ModelUsed)
	}
	if created.GeminiTriageResult.Confidence < 0.90 {
		t.Errorf("expected confidence >= 0.90, got %f", created.GeminiTriageResult.Confidence)
	}
}

func TestSubmitRoadkillIncidentAutoRegistersCarcass(t *testing.T) {
	router := setupTestServer()

	newRoadkill := model.IncidentDto{
		Category: "ROADKILL",
		Species:  "DEER",
		Location: model.GeoLocationDto{
			Latitude:  39.7200,
			Longitude: 141.1600,
		},
		Region:       "JAPAN",
		Description:  "Freshly struck deer carcass in Route 4 median.",
		ReporterId:   "usr_commuter_55",
		ReporterRole: "CITIZEN",
	}
	bodyBytes, _ := json.Marshal(newRoadkill)

	req := httptest.NewRequest("POST", "/api/v1/incidents", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify carcass registered in /api/v1/roadkills
	reqRk := httptest.NewRequest("GET", "/api/v1/roadkills", nil)
	wRk := httptest.NewRecorder()
	router.ServeHTTP(wRk, reqRk)

	var roadkills []model.RoadkillRecordDto
	if err := json.Unmarshal(wRk.Body.Bytes(), &roadkills); err != nil {
		t.Fatalf("failed to decode roadkills: %v", err)
	}

	found := false
	for _, rk := range roadkills {
		if rk.Species == "DEER" && rk.Location.Latitude == 39.7200 {
			found = true
			if rk.ClearanceStatus != "REPORTED" {
				t.Errorf("expected clearanceStatus REPORTED, got %s", rk.ClearanceStatus)
			}
			if rk.AttractantThreatScore <= 0 {
				t.Errorf("expected positive attractant threat score")
			}
			break
		}
	}
	if !found {
		t.Errorf("expected auto-registered roadkill record in roadkills list")
	}
}

func TestGetRoadkillsEndpoint(t *testing.T) {
	router := setupTestServer()

	req := httptest.NewRequest("GET", "/api/v1/roadkills", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var roadkills []model.RoadkillRecordDto
	if err := json.Unmarshal(w.Body.Bytes(), &roadkills); err != nil {
		t.Fatalf("failed to decode roadkills: %v", err)
	}

	if len(roadkills) < 3 {
		t.Errorf("expected at least 3 seed roadkills, got %d", len(roadkills))
	}
}

func TestUpdateRoadkillStatusEndpoint(t *testing.T) {
	router := setupTestServer()

	// 1. Dispatch sanitation crew: REPORTED -> DISPATCHED
	dispatchBody, _ := json.Marshal(model.UpdateRoadkillStatusRequestDto{
		Status:             "DISPATCHED",
		AssignedOfficialId: "OFFICER-SANITATION-09",
	})
	req := httptest.NewRequest("POST", "/api/v1/roadkills/rk_4412/status", bytes.NewReader(dispatchBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var dispatched model.RoadkillRecordDto
	if err := json.Unmarshal(w.Body.Bytes(), &dispatched); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if dispatched.ClearanceStatus != "DISPATCHED" {
		t.Errorf("expected DISPATCHED status, got %s", dispatched.ClearanceStatus)
	}
	if dispatched.AssignedOfficialId != "OFFICER-SANITATION-09" {
		t.Errorf("expected assigned official OFFICER-SANITATION-09, got %s", dispatched.AssignedOfficialId)
	}

	// 2. Mark carcass as Cleared: DISPATCHED -> CLEARED
	clearBody, _ := json.Marshal(model.UpdateRoadkillStatusRequestDto{
		Status: "CLEARED",
	})
	req2 := httptest.NewRequest("POST", "/api/v1/roadkills/rk_4412/status", bytes.NewReader(clearBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w2.Code, w2.Body.String())
	}

	var cleared model.RoadkillRecordDto
	if err := json.Unmarshal(w2.Body.Bytes(), &cleared); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if cleared.ClearanceStatus != "CLEARED" {
		t.Errorf("expected CLEARED status, got %s", cleared.ClearanceStatus)
	}
	if cleared.ClearedAt == "" {
		t.Errorf("expected ClearedAt timestamp set upon clearance")
	}
}

func TestColdStartStatusEndpoint(t *testing.T) {
	router := setupTestServer()
	req := httptest.NewRequest("GET", "/api/v1/coldstart/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var status coldstart.ColdStartStatus
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("failed to decode ColdStartStatus: %v", err)
	}

	if status.TotalGbifOccurrences == 0 {
		t.Errorf("expected pre-seeded GBIF occurrences > 0, got %d", status.TotalGbifOccurrences)
	}
	if status.TotalFirmsAnomalies == 0 {
		t.Errorf("expected pre-seeded FIRMS anomalies > 0, got %d", status.TotalFirmsAnomalies)
	}
	if !status.ReadyForDayOneDeployment {
		t.Errorf("expected ReadyForDayOneDeployment=true")
	}
}

func TestColdStartBootstrapEndpoint(t *testing.T) {
	router := setupTestServer()

	// Bootstrap for JAPAN region
	reqBody, _ := json.Marshal(map[string]string{
		"region": "JAPAN",
	})
	req := httptest.NewRequest("POST", "/api/v1/coldstart/gbif-firms/bootstrap", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var status coldstart.ColdStartStatus
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil {
		t.Fatalf("failed to decode ColdStartStatus: %v", err)
	}

	if !status.Bootstrapped {
		t.Errorf("expected Bootstrapped=true")
	}
	if status.LastBootstrapTimestamp == "" {
		t.Errorf("expected LastBootstrapTimestamp to be set")
	}

	// Verify geofences are active
	geoReq := httptest.NewRequest("GET", "/api/v1/geofences/active?latitude=39.7&longitude=141.1&radiusKm=5000", nil)
	wGeo := httptest.NewRecorder()
	router.ServeHTTP(wGeo, geoReq)
	if wGeo.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for active geofences, got %d", wGeo.Code)
	}
}

func TestNgoBulkImportEndpoint_CSV(t *testing.T) {
	router := setupTestServer()

	csvPayload := `date,species,lat,lon,observation_type,tier,notes
2026-10-14T06:00:00Z,BEAR,39.7020,141.1520,DIRECT_SIGHTING,TIER_2_WARNING,Spotted near school boundary
2026-10-14T08:00:00Z,TIGER,0.8250,101.5300,PUGMARK_TRACK,TIER_3_DANGER,Corral intrusion
`
	importReq := coldstart.NgoBulkImportRequest{
		Format:        "CSV",
		Organization:  "International Ranger Federation",
		RangerStation: "Iwate-Station-01",
		Region:        "JAPAN",
		Payload:       csvPayload,
	}
	reqBytes, _ := json.Marshal(importReq)
	req := httptest.NewRequest("POST", "/api/v1/ngo/patrol-logs/bulk-import", bytes.NewReader(reqBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp coldstart.NgoBulkImportResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode NgoBulkImportResponse: %v", err)
	}

	if resp.ImportedCount != 2 {
		t.Errorf("expected 2 records imported, got %d", resp.ImportedCount)
	}
	if resp.FormatDetected != "CSV" {
		t.Errorf("expected format CSV, got %s", resp.FormatDetected)
	}
	if resp.GeofencesActivated == 0 {
		t.Errorf("expected geofences to be activated")
	}
}

func TestNgoTemplatesEndpoint(t *testing.T) {
	router := setupTestServer()
	req := httptest.NewRequest("GET", "/api/v1/ngo/patrol-logs/templates", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result struct {
		Templates map[string]string `json:"templates"`
		Formats   []string          `json:"formats"`
		Sla       string            `json:"sla"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode templates response: %v", err)
	}

	if len(result.Templates) < 3 {
		t.Errorf("expected at least 3 templates (CSV, GEOJSON, GPX), got %d", len(result.Templates))
	}
	if result.Sla == "" {
		t.Errorf("expected SLA specification in template response")
	}
}

func makeTestToken(t *testing.T, key *rsa.PrivateKey, project string, role string) string {
	t.Helper()
	now := time.Now()
	payload := map[string]any{
		"iss":   "https://securetoken.google.com/" + project,
		"aud":   project,
		"sub":   "u-test-admin",
		"iat":   now.Add(-time.Minute).Unix(),
		"exp":   now.Add(time.Hour).Unix(),
		"email": "tester@coexistshield.ai",
		"role":  role,
	}
	hb, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "k1"})
	pb, _ := json.Marshal(payload)
	signing := base64.RawURLEncoding.EncodeToString(hb) + "." + base64.RawURLEncoding.EncodeToString(pb)
	d := sha256.Sum256([]byte(signing))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, d[:])
	if err != nil {
		t.Fatal(err)
	}
	return signing + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func TestAdminOnlyEndpointsAndRoleAssignment(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	const testProj = "coexistshield-auth-test"

	jwkServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kid": "k1",
			"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		}}})
	}))
	defer jwkServer.Close()

	router := setupTestServer()
	verifier := auth.NewVerifier(testProj, jwkServer.URL)
	router.SetAuthVerifier(verifier)

	// 1. Unauthenticated request to /api/v1/admin/set-role -> 401
	setRolePayload := `{"uid":"user-to-promote","role":"OFFICIAL"}`
	req := httptest.NewRequest("POST", "/api/v1/admin/set-role", strings.NewReader(setRolePayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for missing token, got %d", w.Code)
	}

	// 2. CITIZEN request to /api/v1/admin/set-role -> 403
	citizenToken := makeTestToken(t, key, testProj, "CITIZEN")
	req = httptest.NewRequest("POST", "/api/v1/admin/set-role", strings.NewReader(setRolePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+citizenToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for CITIZEN, got %d", w.Code)
	}

	// 3. OFFICIAL request to adminOnly endpoint -> 403 (adminOnly strictly requires ADMIN)
	officialToken := makeTestToken(t, key, testProj, "OFFICIAL")
	req = httptest.NewRequest("POST", "/api/v1/admin/set-role", strings.NewReader(setRolePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+officialToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for OFFICIAL on admin-only route, got %d", w.Code)
	}

	// 4. ADMIN request to /api/v1/admin/set-role -> 200 OK
	adminToken := makeTestToken(t, key, testProj, "ADMIN")
	req = httptest.NewRequest("POST", "/api/v1/admin/set-role", strings.NewReader(setRolePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for ADMIN, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["success"] != true || resp["uid"] != "user-to-promote" || resp["role"] != "OFFICIAL" {
		t.Errorf("unexpected set-role response: %+v", resp)
	}
}

func TestRouterCustomRepositoryPersistence(t *testing.T) {
	router := setupTestServer()
	customRepo := repository.NewMemoryRepository()
	router.SetRepositories(customRepo, customRepo)

	// 1. Submit incident via HTTP POST
	postPayload := `{
		"id": "inc-custom-repo-test",
		"species": "BEAR",
		"category": "WILDLIFE_SIGHTING",
		"location": {"latitude": 39.75, "longitude": 141.20},
		"region": "JAPAN"
	}`
	req := httptest.NewRequest("POST", "/api/v1/incidents", strings.NewReader(postPayload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", w.Code)
	}

	// 2. Verify stored directly in repository
	inc, err := customRepo.GetIncidentByID(context.Background(), "inc-custom-repo-test")
	if err != nil {
		t.Fatalf("incident was not persisted to custom repository: %v", err)
	}
	if inc.Species != "BEAR" || inc.Location.Latitude != 39.75 {
		t.Errorf("unexpected incident data in repository: %+v", inc)
	}

	// 3. Verify retrieved via HTTP GET
	getReq := httptest.NewRequest("GET", "/api/v1/incidents", nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", getW.Code)
	}
	var incidents []model.IncidentDto
	if err := json.Unmarshal(getW.Body.Bytes(), &incidents); err != nil {
		t.Fatalf("failed to decode incidents: %v", err)
	}
	found := false
	for _, item := range incidents {
		if item.Id == "inc-custom-repo-test" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find inc-custom-repo-test in GET response, got %+v", incidents)
	}
}









