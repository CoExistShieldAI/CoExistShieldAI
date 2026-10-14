package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestServer() (*httptest.Server, *Store) {
	store := NewStore()
	handler := NewHandler(store)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)
	server := httptest.NewServer(mux)
	return server, store
}

func TestDashboardStats(t *testing.T) {
	ts, _ := setupTestServer()
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/v1/admin/dashboard/stats")
	if err != nil {
		t.Fatalf("Failed to GET stats: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var stats DashboardStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if stats.TotalGeofences == 0 || stats.TotalRoadkills == 0 || stats.TotalUsers == 0 {
		t.Errorf("Dashboard stats missing seeded counts: %+v", stats)
	}
}

func TestGeofencesManualAndBulkPurge(t *testing.T) {
	ts, _ := setupTestServer()
	defer ts.Close()

	// 1. Create a manual polygon geofence
	newGeo := AdminGeofence{
		Name:         "Test Manual Polygon",
		GeometryType: "polygon",
		Coordinates: []Coordinate{
			{Lat: 13.0, Lng: 100.0},
			{Lat: 13.1, Lng: 100.1},
			{Lat: 13.0, Lng: 100.2},
		},
		ThreatLevel:   "TIER_3_CRITICAL",
		TargetSpecies: "Asian Elephant",
		IsPersistent:  true,
		Reason:        "Unit test manual creation",
	}
	body, _ := json.Marshal(newGeo)
	postResp, err := http.Post(ts.URL+"/api/v1/admin/geofences", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to POST geofence: %v", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201 Created, got %d", postResp.StatusCode)
	}

	var created AdminGeofence
	_ = json.NewDecoder(postResp.Body).Decode(&created)
	if created.ID == "" || created.Source != "manual" || !created.IsPersistent {
		t.Errorf("Created geofence properties incorrect: %+v", created)
	}

	// 2. Trigger AI recalculation
	recalcResp, err := http.Post(ts.URL+"/api/v1/admin/geofences/recalculate", "application/json", nil)
	if err != nil || recalcResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to recalculate AI geofences: %v", err)
	}
	recalcResp.Body.Close()

	// 3. Clear all manual geofences
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/admin/geofences/manual/all", nil)
	delResp, err := http.DefaultClient.Do(req)
	if err != nil || delResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to purge manual geofences: %v", err)
	}
	delResp.Body.Close()

	// 4. Verify manual fences are purged but AI fences remain
	listResp, _ := http.Get(ts.URL + "/api/v1/admin/geofences?source=manual")
	var manualFences []AdminGeofence
	_ = json.NewDecoder(listResp.Body).Decode(&manualFences)
	listResp.Body.Close()

	if len(manualFences) != 0 {
		t.Errorf("Expected 0 manual fences after purge, got %d", len(manualFences))
	}

	listAiResp, _ := http.Get(ts.URL + "/api/v1/admin/geofences?source=ai")
	var aiFences []AdminGeofence
	_ = json.NewDecoder(listAiResp.Body).Decode(&aiFences)
	listAiResp.Body.Close()

	if len(aiFences) == 0 {
		t.Errorf("AI fences should remain intact, got 0")
	}
}

func TestRoadkillReliabilityExclusionAndDisposal(t *testing.T) {
	ts, _ := setupTestServer()
	defer ts.Close()

	// 1. Exclude roadkill record
	rkID := "rk-2026-001"
	excludePayload := []byte(`{"is_excluded": true}`)
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/api/v1/admin/incidents/roadkill/"+rkID+"/exclude", bytes.NewReader(excludePayload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to exclude roadkill: %v", err)
	}
	var updated AdminRoadkillRecord
	_ = json.NewDecoder(resp.Body).Decode(&updated)
	resp.Body.Close()

	if !updated.IsExcluded {
		t.Errorf("Expected IsExcluded to be true, got false")
	}

	// 2. Mark as disposed
	reqDispose, _ := http.NewRequest(http.MethodPatch, ts.URL+"/api/v1/admin/incidents/roadkill/"+rkID+"/dispose", nil)
	respDispose, err := http.DefaultClient.Do(reqDispose)
	if err != nil || respDispose.StatusCode != http.StatusOK {
		t.Fatalf("Failed to dispose roadkill: %v", err)
	}
	_ = json.NewDecoder(respDispose.Body).Decode(&updated)
	respDispose.Body.Close()

	if updated.Status != "DISPOSED" {
		t.Errorf("Expected status DISPOSED, got %s", updated.Status)
	}
}

func TestOperationsEmergencyBroadcastAndExport(t *testing.T) {
	ts, _ := setupTestServer()
	defer ts.Close()

	// 1. Emergency Broadcast
	bcastReq := EmergencyBroadcastRequest{
		Title:    "URGENT: Tiger Sighting Near Camp",
		Body:     "All personnel retreat to vehicle immediately",
		Priority: "URGENT",
	}
	body, _ := json.Marshal(bcastReq)
	bcastResp, err := http.Post(ts.URL+"/api/v1/admin/operations/broadcast", "application/json", bytes.NewReader(body))
	if err != nil || bcastResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to send broadcast: %v", err)
	}
	var bcastResult EmergencyBroadcastResponse
	_ = json.NewDecoder(bcastResp.Body).Decode(&bcastResult)
	bcastResp.Body.Close()

	if bcastResult.BroadcastID == "" || bcastResult.EstimatedRecipients == 0 {
		t.Errorf("Invalid broadcast result: %+v", bcastResult)
	}

	// 2. GeoJSON Export
	expResp, err := http.Get(ts.URL + "/api/v1/admin/export/geofences?format=geojson")
	if err != nil || expResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to export GeoJSON: %v", err)
	}
	if expResp.Header.Get("Content-Type") != "application/geo+json" {
		t.Errorf("Unexpected content type: %s", expResp.Header.Get("Content-Type"))
	}
	expResp.Body.Close()

	// 3. CSV Export
	csvResp, err := http.Get(ts.URL + "/api/v1/admin/export/roadkill?format=csv")
	if err != nil || csvResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to export CSV: %v", err)
	}
	if csvResp.Header.Get("Content-Type") != "text/csv" {
		t.Errorf("Unexpected content type: %s", csvResp.Header.Get("Content-Type"))
	}
	csvResp.Body.Close()
}
