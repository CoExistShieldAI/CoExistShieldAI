package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai.coexistshield/backend/internal/elephant"
	"ai.coexistshield/backend/internal/gemini"
	"ai.coexistshield/backend/internal/geofence"
	"ai.coexistshield/backend/internal/memory"
	"ai.coexistshield/backend/internal/ml"
	"ai.coexistshield/backend/internal/model"
)

func setupTestServer() *Router {
	geminiClient := gemini.NewClient()
	geofenceEngine := geofence.NewEngine()
	memoryBank := memory.NewBank()
	mlPipeline := ml.NewRetrainingPipeline()
	elephantEngine := elephant.NewEngine()
	return NewRouter(geminiClient, geofenceEngine, memoryBank, mlPipeline, elephantEngine)
}

func TestHealthz(t *testing.T) {
	router := setupTestServer()
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if body["status"] != "healthy" {
		t.Errorf("expected status healthy, got %s", body["status"])
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

	if resp.RecalculatedGeofencesCount != 3 {
		t.Errorf("expected 3 recalculated geofences, got %d", resp.RecalculatedGeofencesCount)
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



