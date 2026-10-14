package ml

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestGeoToH3Res8(t *testing.T) {
	// Morioka, Iwate coordinates
	lat := 39.7020
	lon := 141.1520
	h3Index := GeoToH3Res8(lat, lon)

	if h3Index == "" {
		t.Fatalf("expected non-empty H3 index")
	}
	if !strings.HasPrefix(h3Index, "88") {
		t.Errorf("expected H3 Res-8 index to start with '88', got: %s", h3Index)
	}

	cell := GetH3CellDetails(lat, lon)
	if cell.Resolution != 8 {
		t.Errorf("expected resolution 8, got %d", cell.Resolution)
	}
	if len(cell.Boundary) != 7 {
		t.Errorf("expected 7 boundary coordinates (closed hexagon), got %d", len(cell.Boundary))
	}
	// Check first and last vertex match to ensure valid closed GeoJSON LinearRing
	if cell.Boundary[0] != cell.Boundary[6] {
		t.Errorf("expected closed LinearRing where first and last vertices match")
	}
}

func TestBearingAndWindAlignment(t *testing.T) {
	// Attractant directly North of cell (bearing 0 deg)
	cellLat, cellLon := 39.700, 141.150
	attrLat, attrLon := 39.710, 141.150

	bearing := BearingDegrees(cellLat, cellLon, attrLat, attrLon)
	if bearing < 355 && bearing > 5 {
		t.Errorf("expected approximate North bearing (~0 deg), got %.2f", bearing)
	}

	// Prevailing wind is from South (180 deg) -> blowing toward North (0 deg)
	// Scent blows downwind toward North, so scent alignment is 0 deg (maximum risk)
	alignment := WindScentAlignmentAngle(180.0, 0.0)
	if alignment > 5.0 {
		t.Errorf("expected alignment near 0 deg, got %.2f", alignment)
	}

	// Wind from North (0 deg) -> blowing toward South (180 deg), away from attractant
	// Alignment is 180 deg (minimum risk)
	alignmentOpposite := WindScentAlignmentAngle(0.0, 0.0)
	if alignmentOpposite < 175.0 {
		t.Errorf("expected opposite alignment near 180 deg, got %.2f", alignmentOpposite)
	}
}

func TestBigQuerySQLGeneration(t *testing.T) {
	cfg := DefaultBigQueryConfig()
	sql := GenerateBigQueryFeatureExtractionSQL(cfg)

	expectedKeywords := []string{
		"H3 Level 8",
		"h3_geo_to_h3",
		"dist_nearest_persimmon_m",
		"dist_nearest_carcass_m",
		"dist_nearest_cane_fruit_m",
		"wind_scent_alignment_deg",
		"moon_darkness_pct",
		"historical_incidents_30d",
		"ST_DISTANCE",
		"ST_DWITHIN",
	}

	for _, kw := range expectedKeywords {
		if !strings.Contains(sql, kw) {
			t.Errorf("expected generated BigQuery SQL to contain '%s'", kw)
		}
	}
}

func TestModelPrediction(t *testing.T) {
	model := DefaultBaselineProductionModel()

	// High hazard scenario: persimmon 200m away, wind aligned (10 deg), new moon (95% darkness)
	highHazard := FeatureVector{
		H3Index:                "882f0a123456789",
		DistNearestPersimmonM:  200.0,
		DistNearestCarcassM:    450.0,
		DistNearestCaneFruitM:  2000.0,
		WindScentAlignmentDeg:  10.0,
		MoonDarknessPct:        95.0,
		PrecipitationMm24h:      0.0,
		HistoricalIncidents30d: 6.0,
	}

	resHigh := model.Predict(highHazard)
	if resHigh.IncursionProbability < 0.60 {
		t.Errorf("expected high hazard probability >= 0.60, got %.3f", resHigh.IncursionProbability)
	}
	if resHigh.AlertTier != "TIER_2_WARNING" && resHigh.AlertTier != "TIER_3_DANGER" {
		t.Errorf("expected Warning or Danger tier, got %s", resHigh.AlertTier)
	}

	// Safe control scenario: far away attractants, opposite wind (170 deg), full moon (5% darkness)
	safeScenario := FeatureVector{
		H3Index:                "882f0a123456789",
		DistNearestPersimmonM:  6000.0,
		DistNearestCarcassM:    8000.0,
		DistNearestCaneFruitM:  9000.0,
		WindScentAlignmentDeg:  170.0,
		MoonDarknessPct:        5.0,
		PrecipitationMm24h:      25.0,
		HistoricalIncidents30d: 0.0,
	}

	resSafe := model.Predict(safeScenario)
	if resSafe.IncursionProbability > 0.30 {
		t.Errorf("expected safe scenario probability <= 0.30, got %.3f", resSafe.IncursionProbability)
	}
	if resSafe.AlertTier != "SAFE" && resSafe.AlertTier != "TIER_1_NOTICE" {
		t.Errorf("expected SAFE or TIER_1_NOTICE, got %s", resSafe.AlertTier)
	}
}

func TestRetrainingPipelineEndToEnd(t *testing.T) {
	pipeline := NewRetrainingPipeline()
	pipeline.LoadSamples(GenerateSeedTrainingDataset())

	activeBefore := pipeline.ActiveModel()
	if activeBefore.Version == "" {
		t.Fatalf("expected initial active model to have a version")
	}

	ctx := context.Background()
	result, err := pipeline.Retrain(ctx, "test_continuous_trigger", false)
	if err != nil {
		t.Fatalf("retrain failed: %v", err)
	}

	if result.TotalSamples < 100 {
		t.Errorf("expected at least 100 total samples, got %d", result.TotalSamples)
	}
	if result.CandidateModel.Metrics.AucRoc < 0.82 {
		t.Errorf("acceptance criteria violation: AUC-ROC %.3f < 0.82", result.CandidateModel.Metrics.AucRoc)
	}
	if result.CandidateModel.Metrics.PrecisionTop20 < 0.75 {
		t.Errorf("acceptance criteria violation: Precision@Top20%% %.3f < 0.75", result.CandidateModel.Metrics.PrecisionTop20)
	}
	if !result.GateCheckPassed {
		t.Errorf("expected validation gate check to pass, notes: %s", result.GateEvaluationNotes)
	}

	// Verify promotion
	activeAfter := pipeline.ActiveModel()
	if activeAfter.Version != result.CandidateModel.Version {
		t.Errorf("expected active model to be promoted to candidate version %s, got %s",
			result.CandidateModel.Version, activeAfter.Version)
	}
}

func TestModelSerializationCheckpoint(t *testing.T) {
	pipeline := NewRetrainingPipeline()
	data, err := pipeline.ExportModelJSON()
	if err != nil {
		t.Fatalf("failed to export model JSON: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("expected non-empty model JSON")
	}

	pipeline2 := NewRetrainingPipeline()
	if err := pipeline2.LoadModelFromJSON(data); err != nil {
		t.Fatalf("failed to load model from JSON: %v", err)
	}

	m1 := pipeline.ActiveModel()
	m2 := pipeline2.ActiveModel()
	if m1.Version != m2.Version {
		t.Errorf("expected version %s, got %s", m1.Version, m2.Version)
	}
	if m1.Bias != m2.Bias {
		t.Errorf("expected bias %f, got %f", m1.Bias, m2.Bias)
	}
}

func TestFeatureExtractionAndNegativeSampling(t *testing.T) {
	kaki, carcasses, crops := DefaultKnownAttractants()
	now := time.Now().UTC()

	// 1. Positive feature extraction near Morioka
	posFeatures := ExtractFeatureVector(39.7020, 141.1520, now, kaki, carcasses, crops)
	if posFeatures.H3Index == "" {
		t.Errorf("expected valid H3 index")
	}
	if posFeatures.DistNearestPersimmonM > 10.0 {
		t.Errorf("expected persimmon distance <= 10m at (39.7020, 141.1520), got %.1f", posFeatures.DistNearestPersimmonM)
	}

	// 2. Negative sampling creates safe control points
	negSamples := GenerateNegativeSamples(10, 39.7020, 141.1520, kaki, carcasses, crops)
	if len(negSamples) != 10 {
		t.Fatalf("expected 10 negative samples, got %d", len(negSamples))
	}

	for i, s := range negSamples {
		if s.IsIncursion != 0 {
			t.Errorf("sample %d: expected IsIncursion == 0, got %d", i, s.IsIncursion)
		}
		// Negative samples are placed > 2.5km away from human attractants
		if s.Features.DistNearestPersimmonM < 1000.0 {
			t.Errorf("sample %d: expected distant persimmon > 1000m, got %.1f", i, s.Features.DistNearestPersimmonM)
		}
	}
}
