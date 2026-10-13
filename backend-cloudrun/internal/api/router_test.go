package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai.coexistshield/backend/internal/gemini"
	"ai.coexistshield/backend/internal/geofence"
	"ai.coexistshield/backend/internal/model"
)

func setupTestServer() *Router {
	geminiClient := gemini.NewClient()
	geofenceEngine := geofence.NewEngine()
	return NewRouter(geminiClient, geofenceEngine)
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
