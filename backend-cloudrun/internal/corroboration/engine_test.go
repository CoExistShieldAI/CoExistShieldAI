package corroboration

import (
	"context"
	"math"
	"testing"
	"time"

	"ai.coexistshield/backend/internal/gemini"
	"ai.coexistshield/backend/internal/model"
)

func TestEngineInitializationAndCausalStreams(t *testing.T) {
	geminiClient := gemini.NewClient()
	engine := NewEngine(geminiClient)

	// Verify pre-seeded streams for Japan
	jpStreams := engine.GetActiveStreamItems("JAPAN", "")
	if len(jpStreams) < 2 {
		t.Fatalf("expected at least 2 pre-seeded Japan streams, got %d", len(jpStreams))
	}

	// Verify Thailand streams
	thStreams := engine.GetActiveStreamItems("THAILAND", "")
	if len(thStreams) < 2 {
		t.Fatalf("expected at least 2 pre-seeded Thailand streams, got %d", len(thStreams))
	}

	// Verify Indonesia streams
	idStreams := engine.GetActiveStreamItems("INDONESIA", "")
	if len(idStreams) < 1 {
		t.Fatalf("expected at least 1 pre-seeded Indonesia stream, got %d", len(idStreams))
	}
}

func TestBayesianCorroborationFormulaAndTierElevation(t *testing.T) {
	geminiClient := gemini.NewClient()
	engine := NewEngine(geminiClient)
	ctx := context.Background()

	// Target coordinates around Morioka, Japan where pre-seeded Agricultural (0.88) and Municipal Waste (0.82) are within 2km
	req := model.CorroborationRequest{
		TargetLatitude:  39.7030,
		TargetLongitude: 141.1530,
		RadiusKm:        3.0,
		Region:          "JAPAN",
	}

	resp, err := engine.Corroborate(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error during corroboration: %v", err)
	}

	if len(resp.Clusters) == 0 {
		t.Fatalf("expected at least 1 corroborated cluster, got 0")
	}

	cluster := resp.Clusters[0]
	if cluster.IndependentSourceCount != 2 {
		t.Errorf("expected 2 independent sources, got %d", cluster.IndependentSourceCount)
	}

	// Math check:
	// C_corroborated = 1.0 - (1 - 0.88) * (1 - 0.82) = 1.0 - (0.12 * 0.18) = 1.0 - 0.0216 = 0.9784 (~0.978)
	expectedConf := 1.0 - (1.0-0.88)*(1.0-0.82)
	if math.Abs(cluster.CombinedConfidence-expectedConf) > 0.005 {
		t.Errorf("expected combined confidence ~%f, got %f", expectedConf, cluster.CombinedConfidence)
	}

	// 2 sources should elevate to TIER_2_WARNING
	if cluster.ElevatedTier != "TIER_2_WARNING" {
		t.Errorf("expected TIER_2_WARNING, got %s", cluster.ElevatedTier)
	}

	// Now ingest a 3rd independent stream (Satellite Thermal FIRMS burn) in the same area
	engine.IngestStreamItem(model.CausalStreamItem{
		ItemId:        "stream-firms-jp-test",
		StreamType:    model.StreamSatelliteThermal,
		Region:        "JAPAN",
		Title:         "NASA FIRMS Thermal Fire Anomaly",
		Latitude:      39.7035,
		Longitude:     141.1532,
		Confidence:    0.85,
		TargetSpecies: "BEAR",
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		ExpiresAt:     time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	})

	resp3, err := engine.Corroborate(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error during 3-source corroboration: %v", err)
	}

	cluster3 := resp3.Clusters[0]
	if cluster3.IndependentSourceCount != 3 {
		t.Errorf("expected 3 independent sources, got %d", cluster3.IndependentSourceCount)
	}

	// 3 sources should elevate to TIER_3_DANGER
	if cluster3.ElevatedTier != "TIER_3_DANGER" {
		t.Errorf("expected TIER_3_DANGER, got %s", cluster3.ElevatedTier)
	}

	// Combined confidence with 3 sources should exceed 0.99
	if cluster3.CombinedConfidence < 0.99 {
		t.Errorf("expected confidence >= 0.99 with 3 sources, got %f", cluster3.CombinedConfidence)
	}

	if cluster3.AiJustification == "" {
		t.Errorf("expected non-empty AI justification")
	}
}

func TestScheduledIngestionCycle(t *testing.T) {
	geminiClient := gemini.NewClient()
	engine := NewEngine(geminiClient)

	initialCount := len(engine.GetActiveStreamItems("", ""))
	ingested := engine.RunIngestionCycle()

	if ingested <= 0 {
		t.Errorf("expected positive number of ingested items in cycle, got %d", ingested)
	}

	newCount := len(engine.GetActiveStreamItems("", ""))
	if newCount <= initialCount {
		t.Errorf("expected stream item count to increase from %d, got %d", initialCount, newCount)
	}
}
