package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai.coexistshield/backend/internal/gemini"
	"ai.coexistshield/backend/internal/geofence"
	"ai.coexistshield/backend/internal/memory"
	"ai.coexistshield/backend/internal/model"
)

func setupTestServer() *Router {
	geminiClient := gemini.NewClient()
	geofenceEngine := geofence.NewEngine()
	memoryBank := memory.NewBank()
	return NewRouter(geminiClient, geofenceEngine, memoryBank)
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


