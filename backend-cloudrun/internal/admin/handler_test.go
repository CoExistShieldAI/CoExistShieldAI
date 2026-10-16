package admin

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
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

	// Verify BroadcastListener is triggered
	listenerFired := false
	handler := NewHandler(nil)
	handler.SetBroadcastListener(func(r EmergencyBroadcastRequest, bID string, rec int) {
		listenerFired = true
		if r.Title != bcastReq.Title || bID == "" || rec == 0 {
			t.Errorf("BroadcastListener received unexpected args: %+v, id: %s, rec: %d", r, bID, rec)
		}
	})
	mux2 := http.NewServeMux()
	handler.RegisterRoutes(mux2)
	ts2 := httptest.NewServer(mux2)
	defer ts2.Close()

	bcastResp2, err := http.Post(ts2.URL+"/api/v1/admin/operations/broadcast", "application/json", bytes.NewReader(body))
	if err == nil {
		bcastResp2.Body.Close()
	}
	if !listenerFired {
		t.Errorf("Expected BroadcastListener to be fired on emergency broadcast")
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

func TestUploadPhotos(t *testing.T) {
	ts, _ := setupTestServer()
	defer ts.Close()

	// 1. Successful upload of 2 images
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part1, err := writer.CreateFormFile("files", "test1.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	// JPEG magic bytes: \xFF\xD8\xFF
	_, _ = part1.Write([]byte("\xFF\xD8\xFF\xE0\x00\x10JFIF\x00\x01\x01\x01\x00`\x00`\x00\x00\xFF\xDB"))

	part2, err := writer.CreateFormFile("files", "test2.png")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	// PNG magic bytes: \x89PNG\r\n\x1a\n
	_, _ = part2.Write([]byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"))

	_ = writer.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/admin/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Upload request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var uploadResp UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		t.Fatalf("Failed to decode upload response: %v", err)
	}

	if len(uploadResp.URLs) != 2 {
		t.Fatalf("Expected 2 URLs, got %d", len(uploadResp.URLs))
	}
	for _, u := range uploadResp.URLs {
		if !strings.HasPrefix(u, "https://storage.googleapis.com/") {
			t.Errorf("Invalid GCS URL format: %s", u)
		}
	}

	// 2. Reject non-image file
	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	partTxt, _ := writer.CreateFormFile("files", "malicious.sh")
	_, _ = partTxt.Write([]byte("#!/bin/bash\necho hello"))
	_ = writer.Close()

	req, _ = http.NewRequest(http.MethodPost, ts.URL+"/api/v1/admin/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	badResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer badResp.Body.Close()
	if badResp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for non-image, got %d", badResp.StatusCode)
	}

	// 3. Reject more than 5 files
	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	for i := 0; i < 6; i++ {
		p, _ := writer.CreateFormFile("files", "img.jpg")
		_, _ = p.Write([]byte("\xFF\xD8\xFF\xE0\x00\x10JFIF"))
	}
	_ = writer.Close()

	req, _ = http.NewRequest(http.MethodPost, ts.URL+"/api/v1/admin/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	tooManyResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer tooManyResp.Body.Close()
	if tooManyResp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for >5 files, got %d", tooManyResp.StatusCode)
	}
}

func TestWeatherCRUDAndExport(t *testing.T) {
	ts, _ := setupTestServer()
	defer ts.Close()

	// 1. List seeded weather data
	resp, err := http.Get(ts.URL + "/api/v1/admin/weather")
	if err != nil {
		t.Fatalf("Failed to GET weather: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	var weatherList []AdminWeatherData
	if err := json.NewDecoder(resp.Body).Decode(&weatherList); err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	if len(weatherList) < 5 {
		t.Errorf("Expected at least 5 seeded weather records, got %d", len(weatherList))
	}

	// 2. Filter by region
	respFiltered, err := http.Get(ts.URL + "/api/v1/admin/weather?region=Thailand")
	if err != nil {
		t.Fatalf("Failed to GET filtered weather: %v", err)
	}
	defer respFiltered.Body.Close()
	var thaiWeather []AdminWeatherData
	_ = json.NewDecoder(respFiltered.Body).Decode(&thaiWeather)
	if len(thaiWeather) == 0 || !strings.Contains(thaiWeather[0].Region, "Thailand") {
		t.Errorf("Expected Thailand weather station, got %+v", thaiWeather)
	}

	// 3. Create new Weather record
	newStation := AdminWeatherData{
		StationName:        "Valparai High Forest Sensor",
		Region:             "India (Anamalai)",
		Latitude:           10.325,
		Longitude:          76.955,
		TemperatureCelsius: 22.0,
		HumidityPct:        85.0,
		WindSpeedKmh:       12.0,
		WindDirectionDeg:   120.0,
		PrecipitationMmH:   30.0, // Should trigger FloodedRiverRisk = true
		MoonIllumination:   0.04,
		MoonPhase:          "NEW_MOON",
		WeatherCondition:   "HEAVY_RAIN",
		Notes:              "High rainfall leading to river rise across elephant migration path",
	}
	createBody, _ := json.Marshal(newStation)
	postResp, err := http.Post(ts.URL+"/api/v1/admin/weather", "application/json", bytes.NewReader(createBody))
	if err != nil {
		t.Fatalf("Failed to POST weather: %v", err)
	}
	defer postResp.Body.Close()
	if postResp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected 201 Created, got %d", postResp.StatusCode)
	}
	var created AdminWeatherData
	_ = json.NewDecoder(postResp.Body).Decode(&created)
	if created.ID == "" {
		t.Errorf("Expected auto-generated ID, got empty")
	}
	if !created.IsFloodedRiverRisk {
		t.Errorf("Expected IsFloodedRiverRisk to be true for 30mm/h precipitation")
	}

	// 4. Get Weather Detail
	getResp, err := http.Get(ts.URL + "/api/v1/admin/weather/" + created.ID)
	if err != nil {
		t.Fatalf("Failed to GET single weather: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", getResp.StatusCode)
	}

	// 5. Update Weather
	created.TemperatureCelsius = 24.5
	created.Notes = "Updated observation"
	putBody, _ := json.Marshal(created)
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/admin/weather/"+created.ID, bytes.NewReader(putBody))
	req.Header.Set("Content-Type", "application/json")
	putResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to PUT weather: %v", err)
	}
	defer putResp.Body.Close()
	if putResp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK on PUT, got %d", putResp.StatusCode)
	}
	var updated AdminWeatherData
	_ = json.NewDecoder(putResp.Body).Decode(&updated)
	if updated.TemperatureCelsius != 24.5 {
		t.Errorf("Expected updated temp 24.5, got %f", updated.TemperatureCelsius)
	}

	// 6. Export Weather CSV & GeoJSON
	csvResp, err := http.Get(ts.URL + "/api/v1/admin/export/csv?dataset=weather")
	if err != nil {
		t.Fatalf("Failed to export weather CSV: %v", err)
	}
	defer csvResp.Body.Close()
	if csvResp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK on CSV export, got %d", csvResp.StatusCode)
	}

	geoResp, err := http.Get(ts.URL + "/api/v1/admin/export/geojson?dataset=weather")
	if err != nil {
		t.Fatalf("Failed to export weather GeoJSON: %v", err)
	}
	defer geoResp.Body.Close()
	if geoResp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK on GeoJSON export, got %d", geoResp.StatusCode)
	}

	// 7. Delete Weather
	delReq, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/admin/weather/"+created.ID, nil)
	delResp, err := http.DefaultClient.Do(delReq)
	if err != nil {
		t.Fatalf("Failed to DELETE weather: %v", err)
	}
	defer delResp.Body.Close()
	if delResp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK on DELETE, got %d", delResp.StatusCode)
	}
}

func TestLiveOpenMeteoWeatherSync(t *testing.T) {
	// Setup mock Open-Meteo REST server
	mockMeteoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := OpenMeteoWeatherResponse{
			Current: struct {
				Temperature2m      float64 `json:"temperature_2m"`
				RelativeHumidity2m float64 `json:"relative_humidity_2m"`
				Precipitation      float64 `json:"precipitation"`
				WindSpeed10m       float64 `json:"wind_speed_10m"`
				WindDirection10m   float64 `json:"wind_direction_10m"`
				WeatherCode        int     `json:"weather_code"`
			}{
				Temperature2m:      23.4,
				RelativeHumidity2m: 78.0,
				Precipitation:      0.0,
				WindSpeed10m:       14.2,
				WindDirection10m:   180.0,
				WeatherCode:        0, // Clear
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockMeteoServer.Close()

	t.Setenv("OPEN_METEO_API_ENDPOINT", mockMeteoServer.URL)
	t.Setenv("DISABLE_PERIODIC_SYNC", "true")

	ts, store := setupTestServer()
	defer ts.Close()

	// Call /api/v1/admin/weather/sync-live
	resp, err := http.Post(ts.URL+"/api/v1/admin/weather/sync-live", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to call sync-live: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var syncResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&syncResult); err != nil {
		t.Fatalf("Failed to decode sync-live response: %v", err)
	}

	if syncResult["status"] != "synced" {
		t.Errorf("Expected status 'synced', got %v", syncResult["status"])
	}
	updatedCount := int(syncResult["updated_count"].(float64))
	if updatedCount != 40 {
		t.Errorf("Expected 40 stations updated, got %d", updatedCount)
	}

	// Verify store data was updated with mock values
	station, exists := store.GetWeather("weather-thai-01")
	if !exists {
		t.Fatalf("weather-thai-01 not found")
	}
	if station.TemperatureCelsius != 23.4 {
		t.Errorf("Expected synced temp 23.4, got %f", station.TemperatureCelsius)
	}
	if station.WindSpeedKmh != 14.2 {
		t.Errorf("Expected synced wind speed 14.2, got %f", station.WindSpeedKmh)
	}
}

func TestRecalculateAiGeofences_NaraBearKakiCorrelation(t *testing.T) {
	ts, _ := setupTestServer()
	defer ts.Close()

	// 1. Register Asian Black Bear sighting in Nara
	wildlifeJSON := `{
		"species": "Asian Black Bear",
		"location": {
			"lat": 34.685,
			"lng": 135.832,
			"region": "Nara, Japan",
			"road_name": "Kasugayama Primeval Forest Route"
		},
		"threat_level": "TIER_2_ALERT",
		"status": "VERIFIED",
		"reporter": "Nara Prefecture Forest Patrol",
		"description": "Adult Asian black bear observed foraging near residential edge."
	}`
	wResp, err := http.Post(ts.URL+"/api/v1/admin/incidents/wildlife", "application/json", strings.NewReader(wildlifeJSON))
	if err != nil || wResp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create Nara wildlife incident: %v, status: %d", err, wResp.StatusCode)
	}
	wResp.Body.Close()

	// 2. Register Persimmon (Kaki) Orchard in Nara (~500m from sighting)
	envJSON := `{
		"category": "PERSIMMON_KAKI",
		"name": "Nara Foothill Persimmon Orchard",
		"location_desc": "Abandoned sweet persimmon orchard near Kasugayama slope",
		"geometry_type": "circle",
		"circle_center": {
			"lat": 34.688,
			"lng": 135.836
		},
		"radius_meters": 1200,
		"risk_weight": 0.88,
		"active_season": "Autumn (Oct-Nov)"
	}`
	eResp, err := http.Post(ts.URL+"/api/v1/admin/environmental", "application/json", strings.NewReader(envJSON))
	if err != nil || eResp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create Nara environmental kaki layer: %v, status: %d", err, eResp.StatusCode)
	}
	eResp.Body.Close()

	// 3. Trigger Run AI Geofence Engine recalculation
	recalcResp, err := http.Post(ts.URL+"/api/v1/admin/geofences/recalculate", "application/json", nil)
	if err != nil || recalcResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to trigger recalculate: %v, status: %d", err, recalcResp.StatusCode)
	}
	var recalcResult map[string]interface{}
	_ = json.NewDecoder(recalcResp.Body).Decode(&recalcResult)
	recalcResp.Body.Close()

	if recalcResult["status"] != "recalculated" {
		t.Errorf("Expected status recalculated, got %v", recalcResult["status"])
	}

	// 4. Verify that Nara bear-kaki AI geofence was dynamically generated
	listResp, err := http.Get(ts.URL + "/api/v1/admin/geofences?source=ai")
	if err != nil || listResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to list geofences: %v", err)
	}
	var aiFences []AdminGeofence
	_ = json.NewDecoder(listResp.Body).Decode(&aiFences)
	listResp.Body.Close()

	foundNaraFence := false
	for _, f := range aiFences {
		if strings.Contains(f.Name, "Nara") && strings.Contains(f.Reason, "persimmon (kaki)") {
			foundNaraFence = true
			if f.ThreatLevel != "TIER_3_CRITICAL" {
				t.Errorf("Expected TIER_3_CRITICAL for correlated bear-kaki zone, got %s", f.ThreatLevel)
			}
			if f.TargetSpecies != "Asian Black Bear" {
				t.Errorf("Expected target species Asian Black Bear, got %s", f.TargetSpecies)
			}
			if f.CircleRadiusMeters <= 0 {
				t.Errorf("Expected positive radius, got %f", f.CircleRadiusMeters)
			}
			break
		}
	}

	if !foundNaraFence {
		t.Errorf("Expected dynamically generated Nara Bear-Kaki AI Geofence, but none was found in %+v", aiFences)
	}
}

func TestRecalculateAiGeofences_BusyMutex(t *testing.T) {
	ts, store := setupTestServer()
	defer ts.Close()

	// Simulate engine is currently calculating
	store.mu.Lock()
	store.isRecalculating = true
	store.mu.Unlock()

	resp, err := http.Post(ts.URL+"/api/v1/admin/geofences/recalculate", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to call recalculate: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("Expected 409 Conflict when busy, got %d", resp.StatusCode)
	}

	var res map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&res)
	if res["status"] != "in_progress" {
		t.Errorf("Expected status in_progress, got %v", res["status"])
	}
}


