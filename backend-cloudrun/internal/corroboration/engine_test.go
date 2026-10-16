package corroboration

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ai.coexistshield/backend/internal/gemini"
	"ai.coexistshield/backend/internal/model"
)

func TestEngineInitializationAndCausalStreams(t *testing.T) {
	geminiClient := gemini.NewClient()
	engine := NewEngine(geminiClient)
	engine.LoadFixtureStreams()

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
	engine.LoadFixtureStreams()
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

	// 1. In production mode (no seed/demo flags), RunIngestionCycle must NOT inject synthetic ghost incidents
	t.Setenv("ENABLE_SEED_DATA", "")
	t.Setenv("DEV_MODE", "")
	t.Setenv("DEMO_MODE", "")
	prodIngested := engine.RunIngestionCycle()
	if prodIngested != 0 {
		t.Errorf("expected 0 synthetic ghost items in production mode, got %d", prodIngested)
	}

	// 2. In demo/seed mode, simulated items are injected
	t.Setenv("ENABLE_SEED_DATA", "true")
	initialCount := len(engine.GetActiveStreamItems("", ""))
	ingested := engine.RunIngestionCycle()

	if ingested <= 0 {
		t.Errorf("expected positive number of ingested items in cycle under seed mode, got %d", ingested)
	}

	newCount := len(engine.GetActiveStreamItems("", ""))
	if newCount <= initialCount {
		t.Errorf("expected stream item count to increase from %d, got %d", initialCount, newCount)
	}
}

func TestPurgeExpiredItems(t *testing.T) {
	geminiClient := gemini.NewClient()
	engine := NewEngine(geminiClient)

	// Ingest an expired item
	engine.IngestStreamItem(model.CausalStreamItem{
		ItemId:    "expired-item-1",
		Title:     "Old Expired Burn",
		ExpiresAt: time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339),
	})
	// Ingest an active item
	engine.IngestStreamItem(model.CausalStreamItem{
		ItemId:    "active-item-1",
		Title:     "Current Burn",
		ExpiresAt: time.Now().UTC().Add(10 * time.Hour).Format(time.RFC3339),
	})

	purged := engine.PurgeExpiredItems()
	if purged != 1 {
		t.Errorf("expected 1 purged item, got %d", purged)
	}

	items := engine.GetActiveStreamItems("", "")
	if len(items) != 1 || items[0].ItemId != "active-item-1" {
		t.Errorf("expected only active-item-1 to remain, got %+v", items)
	}
}

func TestFetchOpenMeteoWeather(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"current": {
				"time": "2026-10-14T11:00",
				"temperature_2m": 16.2,
				"relative_humidity_2m": 62.0,
				"precipitation": 0.5,
				"wind_speed_10m": 14.5,
				"wind_direction_10m": 290.0
			}
		}`))
	}))
	defer mockServer.Close()

	t.Setenv("OPEN_METEO_API_ENDPOINT", mockServer.URL)

	weather, err := FetchOpenMeteoWeather(context.Background(), mockServer.Client(), 39.7020, 141.1520)
	if err != nil {
		t.Fatalf("FetchOpenMeteoWeather failed: %v", err)
	}

	if weather.WindSpeedKmh != 14.5 {
		t.Errorf("expected wind speed 14.5, got %f", weather.WindSpeedKmh)
	}
	if weather.WindDirectionDeg != 290.0 {
		t.Errorf("expected wind direction 290.0, got %f", weather.WindDirectionDeg)
	}
	if weather.PrecipitationMmH != 0.5 {
		t.Errorf("expected precip 0.5, got %f", weather.PrecipitationMmH)
	}
}

func TestFetchFirmsThermalAnomalies(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		csvData := "country_id,latitude,longitude,bright_ti4,scan,track,acq_date,acq_time,satellite,instrument,confidence,version,bright_ti5,frp,daynight\n" +
			"JPN,39.7150,141.1620,335.4,0.4,0.4,2026-10-14,0315,N,VIIRS,nominal,2.0NRT,290.1,12.3,D\n"
		_, _ = w.Write([]byte(csvData))
	}))
	defer mockServer.Close()

	t.Setenv("FIRMS_API_ENDPOINT", mockServer.URL)

	items, err := FetchFirmsThermalAnomalies(context.Background(), mockServer.Client(), "JPN", "dummy-key")
	if err != nil {
		t.Fatalf("FetchFirmsThermalAnomalies failed: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 anomaly item, got %d", len(items))
	}
	if items[0].StreamType != model.StreamSatelliteThermal {
		t.Errorf("expected StreamSatelliteThermal, got %s", items[0].StreamType)
	}
	if items[0].Latitude != 39.7150 || items[0].Longitude != 141.1620 {
		t.Errorf("coordinates mismatch: got (%f, %f)", items[0].Latitude, items[0].Longitude)
	}
}
