package repository

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai.coexistshield/backend/internal/model"
)

func TestFirestoreRepositoryCRUDWithMockServer(t *testing.T) {
	storage := make(map[string]firestoreDocument)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// e.g. /v1/projects/test-proj/databases/(default)/documents/incidents/...
		parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
		collection := ""
		docID := ""
		for i, part := range parts {
			if part == "documents" && i+1 < len(parts) {
				collection = parts[i+1]
				if i+2 < len(parts) {
					docID = parts[i+2]
				}
				break
			}
		}

		key := collection + "/" + docID

		switch r.Method {
		case http.MethodPatch:
			var doc firestoreDocument
			if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			doc.Name = "projects/test-proj/databases/(default)/documents/" + key
			storage[key] = doc
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(doc)

		case http.MethodGet:
			if docID != "" {
				doc, ok := storage[key]
				if !ok {
					http.Error(w, "Not Found", http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(doc)
			} else {
				// List collection
				docs := make([]firestoreDocument, 0)
				for k, v := range storage {
					if strings.HasPrefix(k, collection+"/") {
						docs = append(docs, v)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(firestoreListResponse{Documents: docs})
			}
		}
	}))
	defer ts.Close()

	repo := &FirestoreRepository{
		projectID:  "test-proj",
		baseURL:    ts.URL + "/v1/projects/test-proj/databases/(default)/documents",
		httpClient: ts.Client(),
	}

	ctx := context.Background()

	// 1. Save and Get Incident
	inc := model.IncidentDto{
		Id:                 "inc-001",
		Species:            "BEAR",
		Category:           "WILDLIFE_SIGHTING",
		VerificationStatus: "PENDING",
		Location: model.GeoLocationDto{
			Latitude:  39.702,
			Longitude: 141.152,
		},
		Region: "JAPAN",
	}
	if err := repo.SaveIncident(ctx, inc); err != nil {
		t.Fatalf("failed to save incident: %v", err)
	}

	fetchedInc, err := repo.GetIncidentByID(ctx, "inc-001")
	if err != nil {
		t.Fatalf("failed to get incident: %v", err)
	}
	if fetchedInc.Species != "BEAR" || fetchedInc.Location.Latitude != 39.702 {
		t.Errorf("unexpected incident data: %+v", fetchedInc)
	}

	// 2. Update Incident Verification Status
	if err := repo.UpdateIncidentVerification(ctx, "inc-001", "VERIFIED"); err != nil {
		t.Fatalf("failed to update verification: %v", err)
	}
	updatedInc, _ := repo.GetIncidentByID(ctx, "inc-001")
	if updatedInc.VerificationStatus != "VERIFIED" {
		t.Errorf("expected VERIFIED, got %s", updatedInc.VerificationStatus)
	}

	// 3. Save and Get Roadkill
	rk := model.RoadkillRecordDto{
		Id:                    "rk-001",
		IncidentId:            "inc-001",
		Species:               "WILD_BOAR",
		DecayState:            "FRESH",
		ClearanceStatus:       "REPORTED",
		AttractantThreatScore: 0.85,
		Location: model.GeoLocationDto{
			Latitude:  38.256,
			Longitude: 140.321,
		},
	}
	if err := repo.SaveRoadkill(ctx, rk); err != nil {
		t.Fatalf("failed to save roadkill: %v", err)
	}

	roadkills, err := repo.GetRoadkills(ctx, "REPORTED")
	if err != nil {
		t.Fatalf("failed to list roadkills: %v", err)
	}
	if len(roadkills) != 1 || roadkills[0].Species != "WILD_BOAR" {
		t.Errorf("unexpected roadkills: %+v", roadkills)
	}

	// 4. Update Roadkill Status
	updatedRk, err := repo.UpdateRoadkillStatus(ctx, "rk-001", "CLEARED", "Incinerated by official patrol")
	if err != nil {
		t.Fatalf("failed to update roadkill status: %v", err)
	}
	if updatedRk.ClearanceStatus != "CLEARED" {
		t.Errorf("expected status CLEARED, got %s", updatedRk.ClearanceStatus)
	}

	// 5. ML Model Checkpoint and History Persistence
	mockModelJSON := []byte(`{"version":"v2026.10.16.1200","bias":-0.45,"metrics":{"aucRoc":0.885,"precisionTop20":0.82}}`)
	if err := repo.SaveMLModelCheckpoint(ctx, mockModelJSON, "v2026.10.16.1200", "2026-10-16T12:00:00Z", 0.885, 0.82, 120); err != nil {
		t.Fatalf("failed to save ML model checkpoint: %v", err)
	}

	latestModelBytes, err := repo.GetLatestMLModel(ctx)
	if err != nil {
		t.Fatalf("failed to get latest ML model: %v", err)
	}
	if !strings.Contains(string(latestModelBytes), "v2026.10.16.1200") {
		t.Errorf("expected latest model to contain version v2026.10.16.1200, got: %s", string(latestModelBytes))
	}

	// 6. ML Retraining History and Samples
	mockResultJSON := []byte(`{"runId":"run-999","activeModelVersion":"v2026.10.16.1200","promotedToProd":true}`)
	if err := repo.SaveMLRetrainResult(ctx, mockResultJSON, "run-999", "v2026.10.16.1200", "cron_scheduled", true, "2026-10-16T12:00:00Z"); err != nil {
		t.Fatalf("failed to save ML retrain result: %v", err)
	}
	history, err := repo.ListMLRetrainHistory(ctx)
	if err != nil || len(history) != 1 {
		t.Fatalf("expected 1 history record, got %d (err: %v)", len(history), err)
	}

	mockSampleJSON := []byte(`{"isIncursion":1,"features":{"h3Index":"882f5b4d93fffff"}}`)
	if err := repo.SaveMLTrainingSample(ctx, mockSampleJSON, "sample-001"); err != nil {
		t.Fatalf("failed to save ML training sample: %v", err)
	}
	samples, err := repo.ListMLTrainingSamples(ctx)
	if err != nil || len(samples) != 1 {
		t.Fatalf("expected 1 sample, got %d (err: %v)", len(samples), err)
	}
}
