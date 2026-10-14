package tiger

import (
	"testing"
	"time"

	"ai.coexistshield/backend/internal/model"
)

func TestTigerEngineInitialization(t *testing.T) {
	engine := NewEngine()
	engine.LoadFixtureData()
	enclosures := engine.GetEnclosures()

	if len(enclosures) < 3 {
		t.Fatalf("expected at least 3 initial enclosures, got %d", len(enclosures))
	}

	enc1, err := engine.GetEnclosure("ls_sumatra_01")
	if err != nil {
		t.Fatalf("expected to find ls_sumatra_01: %v", err)
	}

	if !enc1.IsActiveKillSite {
		t.Errorf("expected ls_sumatra_01 to be an active kill site")
	}
	if enc1.AlertTier != "TIER_3_DANGER" {
		t.Errorf("expected active kill site within 14h to be TIER_3_DANGER, got %s", enc1.AlertTier)
	}
	if enc1.ReturnProbability < 0.60 {
		t.Errorf("expected return probability >= 0.60, got %.3f", enc1.ReturnProbability)
	}
	if enc1.KillSiteGeofence == nil {
		t.Fatalf("expected generated GeoJSON polygon")
	}

	ring := enc1.KillSiteGeofence.Coordinates[0]
	if len(ring) != 25 {
		t.Errorf("expected 25 coordinates in closed LinearRing, got %d", len(ring))
	}
	if ring[0][0] != ring[24][0] || ring[0][1] != ring[24][1] {
		t.Errorf("expected closed LinearRing where first and last vertex match")
	}
}

func TestNightCorralRiskReduction(t *testing.T) {
	engine := NewEngine()
	engine.LoadFixtureData()

	// Compare Sungai Rawa (has night corral) vs Lubuk Mandi (no night corral)
	encWithCorral, _ := engine.GetEnclosure("ls_sumatra_02")
	encNoCorral, _ := engine.GetEnclosure("ls_sumatra_03")

	if !encWithCorral.HasNightCorral {
		t.Errorf("expected encWithCorral to have night corral")
	}
	if encNoCorral.HasNightCorral {
		t.Errorf("expected encNoCorral to not have night corral")
	}

	// Threat score should be substantially lower with reinforced night corral
	if encWithCorral.ActiveThreatScore >= encNoCorral.ActiveThreatScore {
		t.Errorf("expected night corral to reduce threat score (corral=%.3f, noCorral=%.3f)",
			encWithCorral.ActiveThreatScore, encNoCorral.ActiveThreatScore)
	}
}

func TestReportDepredationAnd3kmAlert(t *testing.T) {
	engine := NewEngine()
	engine.LoadFixtureData()

	weather := model.WeatherConditions{
		WindSpeedKmh:     12.0,
		WindDirectionDeg: 180.0,
		MoonIllumination: 0.10, // New Moon
	}

	req := ReportDepredationRequest{
		EnclosureID:           "test_kill_01",
		LivestockSpecies:      "WATER_BUFFALO",
		HeadCountLost:         2,
		Latitude:              0.1300,
		Longitude:             101.4450,
		HasNightCorral:        false,
		DragTrailLengthMeters: 45.0,
		CanineBiteSpreadMm:    74.0, // Confirmed tiger puncture spread
	}

	enc, alert, err := engine.ReportDepredation(req, weather)
	if err != nil {
		t.Fatalf("failed to report depredation: %v", err)
	}

	if !enc.IsActiveKillSite {
		t.Errorf("expected enc to be active kill site")
	}
	if enc.AlertTier != "TIER_3_DANGER" {
		t.Errorf("expected TIER_3_DANGER alert tier, got %s", enc.AlertTier)
	}
	if enc.ScentBufferMeters < 2500.0 {
		t.Errorf("expected >= 2500m buffer for fresh kill site, got %.1f", enc.ScentBufferMeters)
	}

	if alert == nil {
		t.Fatalf("expected alert payload to be generated")
	}
	if !alert.CanineForensicsConfirmed {
		t.Errorf("expected canine forensics to be confirmed for 74mm bite spread")
	}
	if len(alert.VillagesIn3kmRadius) == 0 {
		t.Errorf("expected at least one village notified in 3km radius")
	}
}

func TestClearKillSiteAction(t *testing.T) {
	engine := NewEngine()
	engine.LoadFixtureData()

	cleared, err := engine.ClearKillSite("ls_sumatra_01")
	if err != nil {
		t.Fatalf("failed to clear kill site: %v", err)
	}

	if cleared.IsActiveKillSite {
		t.Errorf("expected kill site to be deactivated after clearing")
	}
	if cleared.KillReportedAt != nil {
		t.Errorf("expected KillReportedAt to be cleared")
	}
	if cleared.AlertTier == "TIER_3_DANGER" {
		t.Errorf("expected tier to drop from DANGER after clearing carcass, got %s", cleared.AlertTier)
	}
}

func Test72HourDecayThreshold(t *testing.T) {
	engine := NewEngine()
	engine.LoadFixtureData()
	now := time.Now().UTC()

	// Check enclosure 80 hours after kill (past 72h window)
	killTime80hAgo := now.Add(-80 * time.Hour)
	encPast72h := LivestockEnclosure{
		ID:                    "past_72h_enc",
		Name:                  "Old Kill Pasture",
		Province:              "Riau",
		Village:               "Test Village",
		LivestockType:         "CATTLE",
		HeadCount:             10,
		Latitude:              0.1300,
		Longitude:             101.4450,
		DistanceToForestEdgeM: 1500.0,
		IsActiveKillSite:      true,
		KillReportedAt:        &killTime80hAgo,
	}

	weather := model.WeatherConditions{
		WindSpeedKmh:     10.0,
		WindDirectionDeg: 180.0,
		MoonIllumination: 0.50,
	}

	computed := engine.calculateEnclosureRisk(encPast72h, weather, now)
	if computed.IsActiveKillSite {
		t.Errorf("expected kill site to be deactivated after 80h (past 72h window)")
	}
	if computed.ReturnProbability > 0.20 {
		t.Errorf("expected return probability to decay below 0.20 after 80h, got %.3f", computed.ReturnProbability)
	}
	if computed.AlertTier == "TIER_3_DANGER" {
		t.Errorf("expected tier not to be DANGER after 80h, got %s", computed.AlertTier)
	}
}
