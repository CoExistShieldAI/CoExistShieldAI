package ml

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBigQueryGISReplicatorStreamSync(t *testing.T) {
	pipeline := NewRetrainingPipeline()
	replicator := NewBigQueryGISReplicator(pipeline)

	// Ingest a high-risk bear incident near Takamatsu
	event := IncidentStreamingEvent{
		EventID:           "test_evt_01",
		IncidentID:        "inc_iwate_test",
		Species:           "BEAR",
		Category:          "WILDLIFE_SIGHTING",
		Latitude:          39.7020,
		Longitude:         141.1520,
		Timestamp:         time.Now().UTC(),
		IsIncursion:       1,
		VerificationState: "VERIFIED",
	}

	sample, err := replicator.ReplicateIncidentEvent(event)
	if err != nil {
		t.Fatalf("failed to replicate incident event: %v", err)
	}

	if sample.Features.H3Index == "" {
		t.Errorf("expected non-empty H3 index on enriched sample")
	}
	// Nearest persimmon in test seed is at 39.7020, 141.1520 (distance should be very close to 0)
	if sample.Features.DistNearestPersimmonM > 50.0 {
		t.Errorf("expected nearest persimmon distance < 50m, got %.1f", sample.Features.DistNearestPersimmonM)
	}

	status := replicator.GetSyncStatus()
	if status.Status != "ACTIVE_STREAMING" {
		t.Errorf("expected status ACTIVE_STREAMING, got %s", status.Status)
	}
	if status.ReplicationLagSeconds > float64(status.SlaTargetSeconds) {
		t.Errorf("replication lag %.2fs exceeded SLA target %ds", status.ReplicationLagSeconds, status.SlaTargetSeconds)
	}
	if status.TotalEventsReplicated != 1 {
		t.Errorf("expected 1 total replicated event, got %d", status.TotalEventsReplicated)
	}

	staged := replicator.GetStagedSamples()
	if len(staged) != 1 {
		t.Errorf("expected 1 staged sample, got %d", len(staged))
	}
}

func TestBigQueryStreamSyncBatch(t *testing.T) {
	pipeline := NewRetrainingPipeline()
	replicator := NewBigQueryGISReplicator(pipeline)

	events := []IncidentStreamingEvent{
		{
			EventID:           "evt_b_01",
			IncidentID:        "inc_01",
			Species:           "BEAR",
			Category:          "PHYSICAL_TRACE",
			Latitude:          39.7045,
			Longitude:         141.1490,
			Timestamp:         time.Now().UTC(),
			IsIncursion:       1,
			VerificationState: "VERIFIED",
		},
		{
			EventID:           "evt_b_02",
			IncidentID:        "inc_02",
			Species:           "WILD_BOAR",
			Category:          "ROADKILL",
			Latitude:          39.7120,
			Longitude:         141.1480,
			Timestamp:         time.Now().UTC(),
			IsIncursion:       0,
			VerificationState: "VERIFIED",
		},
	}

	count, err := replicator.StreamSyncBatch(events)
	if err != nil {
		t.Fatalf("batch sync failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 synced events, got %d", count)
	}

	status := replicator.GetSyncStatus()
	if status.TotalEventsReplicated != 2 {
		t.Errorf("expected 2 replicated events in status, got %d", status.TotalEventsReplicated)
	}
}

func TestBigQueryInsertAllWithHttpServer(t *testing.T) {
	receivedPayloads := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPayloads++
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-bq-token" {
			t.Errorf("expected test token in header, got %s", r.Header.Get("Authorization"))
		}

		var req BigQueryInsertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode insert request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(req.Rows) != 1 {
			t.Errorf("expected 1 row, got %d", len(req.Rows))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(BigQueryInsertResponse{
			Kind: "bigquery#tableDataInsertAllResponse",
		})
	}))
	defer server.Close()

	t.Setenv("BIGQUERY_API_ENDPOINT", server.URL)
	t.Setenv("BIGQUERY_AUTH_TOKEN", "test-bq-token")
	t.Setenv("BIGQUERY_STREAMING_ENABLED", "true")

	pipeline := NewRetrainingPipeline()
	replicator := NewBigQueryGISReplicator(pipeline)

	events := []IncidentStreamingEvent{
		{
			EventID:           "test_bq_evt_1",
			IncidentID:        "inc_live_1",
			Species:           "BEAR",
			Category:          "SIGHTING",
			Latitude:          39.7020,
			Longitude:         141.1520,
			Timestamp:         time.Now().UTC(),
			IsIncursion:       1,
			VerificationState: "VERIFIED",
		},
	}

	count, err := replicator.StreamSyncBatch(events)
	if err != nil {
		t.Fatalf("StreamSyncBatch failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 synced event, got %d", count)
	}
	if receivedPayloads != 1 {
		t.Errorf("expected server to receive 1 payload, got %d", receivedPayloads)
	}
}

func TestRetrainingPipelineEventAccumulatorTrigger(t *testing.T) {
	pipeline := NewRetrainingPipeline()

	// Initially accumulator should be 0
	if pipeline.EventAccumulator() != 0 {
		t.Errorf("expected initial accumulator 0, got %d", pipeline.EventAccumulator())
	}

	// Ingest 49 samples: should NOT trigger auto-retraining yet
	for i := 0; i < 49; i++ {
		pipeline.IngestIncidentSample(TrainingSample{
			Features: FeatureVector{
				H3Index:                GeoToH3Res8(39.700+float64(i)*0.001, 141.150),
				DistNearestPersimmonM:  300.0,
				DistNearestCarcassM:    500.0,
				DistNearestCaneFruitM:  2000.0,
				WindScentAlignmentDeg:  15.0,
				MoonDarknessPct:        80.0,
				PrecipitationMm24h:      0.0,
				HistoricalIncidents30d: 4.0,
			},
			IsIncursion: 1,
		})
	}

	if pipeline.EventAccumulator() != 49 {
		t.Errorf("expected accumulator 49, got %d", pipeline.EventAccumulator())
	}

	// Ingest 50th sample: should trigger retrain
	pipeline.IngestIncidentSample(TrainingSample{
		Features: FeatureVector{
			H3Index:                GeoToH3Res8(39.750, 141.200),
			DistNearestPersimmonM:  250.0,
			DistNearestCarcassM:    400.0,
			DistNearestCaneFruitM:  2000.0,
			WindScentAlignmentDeg:  10.0,
			MoonDarknessPct:        85.0,
			PrecipitationMm24h:      0.0,
			HistoricalIncidents30d: 5.0,
		},
		IsIncursion: 1,
	})

	// Give goroutine a moment to complete retraining
	time.Sleep(100 * time.Millisecond)

	history := pipeline.RetrainingHistory()
	if len(history) == 0 {
		t.Errorf("expected at least 1 retraining run triggered by accumulator reaching 50")
	} else {
		lastRun := history[len(history)-1]
		if lastRun.Trigger != "event_accumulated_50" {
			t.Errorf("expected trigger 'event_accumulated_50', got '%s'", lastRun.Trigger)
		}
	}
}

func TestDynamicModelPromotionListener(t *testing.T) {
	pipeline := NewRetrainingPipeline()

	promotedCalled := false
	var promotedModel IncursionModel

	pipeline.RegisterPromotionListener(func(m IncursionModel) {
		promotedCalled = true
		promotedModel = m
	})

	result, err := pipeline.Retrain(context.Background(), "test_promotion_listener", false)
	if err != nil {
		t.Fatalf("retrain failed: %v", err)
	}

	if !result.PromotedToProd {
		t.Fatalf("expected candidate model to be promoted")
	}

	// Give callback goroutine a moment to run
	time.Sleep(50 * time.Millisecond)

	if !promotedCalled {
		t.Errorf("expected promotion listener to be invoked")
	}
	if promotedModel.Version != result.CandidateModel.Version {
		t.Errorf("expected promoted model version %s, got %s", result.CandidateModel.Version, promotedModel.Version)
	}
}
