package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestFirestorePersistence_RestoreOrSeed(t *testing.T) {
	// In-memory mock Firestore server
	docs := make(map[string]map[string]fsValue)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch r.Method {
		case http.MethodGet:
			// List documents in collection
			col := path[strings.LastIndex(path, "/")+1:]
			var list []fsDocument
			for docPath, fields := range docs {
				if strings.Contains(docPath, "/"+col+"/") {
					list = append(list, fsDocument{
						Name:   docPath,
						Fields: fields,
					})
				}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(fsListResponse{Documents: list})

		case http.MethodPatch:
			// Save document
			var doc fsDocument
			if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			docs[path] = doc.Fields
			w.WriteHeader(http.StatusOK)

		case http.MethodDelete:
			delete(docs, path)
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	// 1. Initial run: Firestore is empty, RestoreOrSeed should seed default data
	p := &FirestorePersistence{
		projectID:  "test-project",
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	store := NewStore()
	ctx := context.Background()

	err := p.RestoreOrSeed(ctx, store)
	if err != nil {
		t.Fatalf("expected nil error on RestoreOrSeed, got: %v", err)
	}

	// Verify that docs were seeded into mock Firestore
	if len(docs) == 0 {
		t.Fatalf("expected seeded docs in Firestore, got 0")
	}

	// 2. Add a new manual wildlife incident and verify it's persisted
	store.SetPersistence(p)
	createdWildlife := store.CreateWildlife(AdminWildlifeIncident{
		ID:          "wildlife-custom-01",
		Species:     "Asian Elephant",
		ThreatLevel: "TIER_3_CRITICAL",
		Status:      "VERIFIED",
		Reporter:    "Officer Test",
		Description: "Single bull elephant spotted near road",
	})

	if createdWildlife.ID != "wildlife-custom-01" {
		t.Errorf("expected ID 'wildlife-custom-01', got: %s", createdWildlife.ID)
	}

	// Wait briefly for async persistence goroutine to finish
	time.Sleep(50 * time.Millisecond)

	// Verify custom wildlife exists in docs
	found := false
	for path, fields := range docs {
		if strings.Contains(path, "wildlife-custom-01") {
			found = true
			if v, ok := fields["species"]; !ok || *v.StringValue != "Asian Elephant" {
				t.Errorf("expected species 'Asian Elephant' in fields, got: %v", fields)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected wildlife-custom-01 in Firestore docs, not found")
	}

	// 3. Simulate a new container restart / deployment:
	// A brand new store starts up and restores from Firestore.
	// It MUST contain the custom wildlife incident!
	newStore := NewStore()
	// Before restore, default seed should not have our custom item
	if _, exists := newStore.wildlife["wildlife-custom-01"]; exists {
		t.Fatalf("clean store should not have wildlife-custom-01 before restore")
	}

	err = p.RestoreOrSeed(ctx, newStore)
	if err != nil {
		t.Fatalf("expected nil error on RestoreOrSeed for new store: %v", err)
	}

	if _, exists := newStore.wildlife["wildlife-custom-01"]; !exists {
		t.Errorf("CRITICAL: custom wildlife incident 'wildlife-custom-01' was NOT restored across simulated redeployment!")
	}

	// 4. Test Environmental & Weather persistence
	createdEnv := store.CreateEnvironmental(AdminEnvironmentalData{
		ID:       "env-custom-persimmon-01",
		Category: "PERSIMMON_KAKI",
		Name:     "Abandoned Kaki Orchard",
	})
	if createdEnv.ID != "env-custom-persimmon-01" {
		t.Errorf("expected ID 'env-custom-persimmon-01', got: %s", createdEnv.ID)
	}

	createdWeather := store.CreateWeather(AdminWeatherData{
		ID:          "weather-custom-station-01",
		StationName: "Nagano Foothills Sensor 9",
		Region:      "Nagano, Japan",
	})
	if createdWeather.ID != "weather-custom-station-01" {
		t.Errorf("expected ID 'weather-custom-station-01', got: %s", createdWeather.ID)
	}

	time.Sleep(50 * time.Millisecond)

	// Verify new store restored environmental and weather as well
	restartedStore := NewStore()
	_ = p.RestoreOrSeed(ctx, restartedStore)

	if _, exists := restartedStore.environmental["env-custom-persimmon-01"]; !exists {
		t.Errorf("CRITICAL: environmental data was not restored across simulated redeployment")
	}
	if _, exists := restartedStore.weather["weather-custom-station-01"]; !exists {
		t.Errorf("CRITICAL: weather data was not restored across simulated redeployment")
	}

	// 5. Test Delete persistence
	store.DeleteWildlife("wildlife-custom-01")
	time.Sleep(50 * time.Millisecond)

	postDeleteStore := NewStore()
	_ = p.RestoreOrSeed(ctx, postDeleteStore)
	if _, exists := postDeleteStore.wildlife["wildlife-custom-01"]; exists {
		t.Errorf("CRITICAL: deleted wildlife incident should not be present after restore")
	}
}
