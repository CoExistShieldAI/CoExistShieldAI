package elephant

import (
	"testing"

	"ai.coexistshield/backend/internal/model"
)

func TestElephantEngineInitialization(t *testing.T) {
	engine := NewEngine()
	engine.LoadFixturePlots()
	plots := engine.GetRegisteredPlots()

	if len(plots) < 4 {
		t.Fatalf("expected at least 4 registered elephant crop plots, got %d", len(plots))
	}

	kbSugarcane, err := engine.GetPlot("th_plot_kb_01")
	if err != nil {
		t.Fatalf("expected to find plot th_plot_kb_01: %v", err)
	}

	if kbSugarcane.CropType != "SUGARCANE" {
		t.Errorf("expected SUGARCANE crop type, got %s", kbSugarcane.CropType)
	}
	if kbSugarcane.ActiveRiskScore < 0.50 {
		t.Errorf("expected high baseline risk score for ready sugarcane near forest, got %.3f", kbSugarcane.ActiveRiskScore)
	}
	if kbSugarcane.ScentPlumePolygon == nil {
		t.Fatalf("expected generated GeoJSON polygon geometry")
	}

	coords := kbSugarcane.ScentPlumePolygon.Coordinates[0]
	if len(coords) != 25 { // 24 vertices + 1 closed
		t.Errorf("expected 25 coordinates in closed LinearRing, got %d", len(coords))
	}
	if coords[0][0] != coords[24][0] || coords[0][1] != coords[24][1] {
		t.Errorf("expected first and last coordinate to match for closed LinearRing")
	}
}

func TestWindDirectionCorridorCoupling(t *testing.T) {
	engine := NewEngine()
	engine.LoadFixturePlots()

	// 1. Wind blowing FROM East (90 deg) -> TOWARDS West (270 deg) straight into Kui Buri forest reserve
	favorableWind := model.WeatherConditions{
		WindSpeedKmh:     18.0,
		WindDirectionDeg: 90.0,
		MoonIllumination: 0.10,
	}
	plotsAligned := engine.RecalculateAllPlots(favorableWind)
	var plotAligned ElephantCropPlot
	for _, p := range plotsAligned {
		if p.ID == "th_plot_kb_01" {
			plotAligned = p
			break
		}
	}

	// 2. Wind blowing FROM West (270 deg) -> TOWARDS East (90 deg) AWAY from forest reserve
	adverseWind := model.WeatherConditions{
		WindSpeedKmh:     18.0,
		WindDirectionDeg: 270.0,
		MoonIllumination: 0.10,
	}
	plotsOpposite := engine.RecalculateAllPlots(adverseWind)
	var plotOpposite ElephantCropPlot
	for _, p := range plotsOpposite {
		if p.ID == "th_plot_kb_01" {
			plotOpposite = p
			break
		}
	}

	if plotAligned.ActiveRiskScore <= plotOpposite.ActiveRiskScore {
		t.Errorf("expected aligned wind into forest to have higher risk (aligned=%.3f, opposite=%.3f)",
			plotAligned.ActiveRiskScore, plotOpposite.ActiveRiskScore)
	}
	if plotAligned.ScentBufferMeters <= plotOpposite.ScentBufferMeters {
		t.Errorf("expected aligned wind buffer to expand (aligned=%.1f m, opposite=%.1f m)",
			plotAligned.ScentBufferMeters, plotOpposite.ScentBufferMeters)
	}
}

func TestUpdateHarvestStage(t *testing.T) {
	engine := NewEngine()
	engine.LoadFixturePlots()

	// Update Kaeng Krachan Cane from Ripening to HARVESTED_CLEARED
	updated, err := engine.UpdateHarvestStage("th_plot_kk_01", "HARVESTED_CLEARED", nil)
	if err != nil {
		t.Fatalf("failed to update harvest stage: %v", err)
	}

	if updated.GrowthStage != "HARVESTED_CLEARED" {
		t.Errorf("expected HARVESTED_CLEARED stage, got %s", updated.GrowthStage)
	}
	if updated.MaturityScore != 0.05 {
		t.Errorf("expected cleared maturity score 0.05, got %.2f", updated.MaturityScore)
	}
	if updated.ActiveRiskScore >= 0.50 {
		t.Errorf("expected low risk score after clearing harvest attractant, got %.3f", updated.ActiveRiskScore)
	}
	if updated.AlertTier != "SAFE" && updated.AlertTier != "TIER_1_NOTICE" {
		t.Errorf("expected SAFE or TIER_1_NOTICE after harvest, got %s", updated.AlertTier)
	}
}
