package geofence

import (
	"testing"

	"ai.coexistshield/backend/internal/model"
)

func TestGenerateForecastPlumeGeofences(t *testing.T) {
	engine := NewEngine()

	attractant := model.AttractantHub{
		Id:           "kt_morioka_99",
		Type:         "KAKI_TREE",
		Latitude:     39.7020,
		Longitude:    141.1520,
		FruitLoad:    "ABUNDANT",
		IsTop5Hazard: true,
	}

	req := model.ForecastPlumeGeofencesRequest{
		Attractants: []model.AttractantHub{attractant},
		TargetHours: []int{1, 3, 6},
	}

	// Mock forecasted wind shifting from West (270°) to North (0°) to East (90°)
	forecastPoints := []model.HourlyForecastPoint{
		{OffsetHours: 1, WindSpeedKmh: 10.0, WindDirectionDeg: 270.0, PrecipitationMmH: 0.0}, // blows toward East (90°)
		{OffsetHours: 3, WindSpeedKmh: 15.0, WindDirectionDeg: 0.0, PrecipitationMmH: 0.0},   // blows toward South (180°)
		{OffsetHours: 6, WindSpeedKmh: 20.0, WindDirectionDeg: 90.0, PrecipitationMmH: 30.0},  // blows toward West (270°) + heavy rain
	}

	resp := engine.GenerateForecastPlumeGeofences(req, forecastPoints)

	if resp.TotalSteps != 3 {
		t.Fatalf("expected 3 forecast steps, got %d", resp.TotalSteps)
	}

	step1 := resp.TimelineSteps[0]
	if step1.OffsetHours != 1 {
		t.Errorf("expected offset 1, got %d", step1.OffsetHours)
	}
	if step1.WindBearingDeg != 90.0 {
		t.Errorf("expected wind bearing 90° (blowing East), got %f", step1.WindBearingDeg)
	}
	if step1.Zone.PolygonGeoJson == nil {
		t.Errorf("expected non-nil polygon geometry for step 1")
	}

	step2 := resp.TimelineSteps[1]
	if step2.OffsetHours != 3 {
		t.Errorf("expected offset 3, got %d", step2.OffsetHours)
	}
	if step2.WindBearingDeg != 180.0 {
		t.Errorf("expected wind bearing 180° (blowing South), got %f", step2.WindBearingDeg)
	}

	step3 := resp.TimelineSteps[2]
	if step3.OffsetHours != 6 {
		t.Errorf("expected offset 6, got %d", step3.OffsetHours)
	}
	if step3.WindBearingDeg != 270.0 {
		t.Errorf("expected wind bearing 270° (blowing West), got %f", step3.WindBearingDeg)
	}
	// Under 30mm/h heavy rain, river barrier risk should elevate tier to Tier 3 Danger
	if step3.Zone.Tier != "TIER_3_DANGER" {
		t.Errorf("expected TIER_3_DANGER for torrential rain condition, got %s", step3.Zone.Tier)
	}
}
