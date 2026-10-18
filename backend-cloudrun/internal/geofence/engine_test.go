package geofence

import (
	"math"
	"strings"
	"testing"
	"time"

	"ai.coexistshield/backend/internal/model"
)

func TestCalculateMoonPhaseAstronomical(t *testing.T) {
	// 1. Reference New Moon: Jan 11, 2024, 11:57 UTC
	refNewMoon := time.Date(2024, 1, 11, 11, 57, 0, 0, time.UTC)
	moonAtNew := CalculateMoonPhase(refNewMoon)

	if moonAtNew.PhaseName != "NEW_MOON" {
		t.Errorf("expected NEW_MOON, got %s", moonAtNew.PhaseName)
	}
	if moonAtNew.Illumination > 0.05 {
		t.Errorf("expected illumination near 0.0 for New Moon, got %f", moonAtNew.Illumination)
	}
	if moonAtNew.NocturnalRiskMultiplier < 1.55 {
		t.Errorf("expected high nocturnal multiplier (~1.6) during New Moon, got %f", moonAtNew.NocturnalRiskMultiplier)
	}
	if !moonAtNew.IsNocturnalPeakRisk {
		t.Errorf("expected IsNocturnalPeakRisk to be true during New Moon")
	}

	// 2. Full Moon (~14.76 days after New Moon)
	fullMoonTime := refNewMoon.Add(time.Duration(14.765 * 24 * float64(time.Hour)))
	moonAtFull := CalculateMoonPhase(fullMoonTime)

	if moonAtFull.PhaseName != "FULL_MOON" {
		t.Errorf("expected FULL_MOON, got %s", moonAtFull.PhaseName)
	}
	if moonAtFull.Illumination < 0.95 {
		t.Errorf("expected illumination near 1.0 for Full Moon, got %f", moonAtFull.Illumination)
	}
	if moonAtFull.NocturnalRiskMultiplier > 1.10 {
		t.Errorf("expected low nocturnal multiplier (~1.0) during Full Moon, got %f", moonAtFull.NocturnalRiskMultiplier)
	}
	if moonAtFull.IsNocturnalPeakRisk {
		t.Errorf("expected IsNocturnalPeakRisk to be false during Full Moon")
	}
}

func TestCalculateSolarPhase(t *testing.T) {
	lat := 39.7020
	lon := 141.1520 // Morioka, Japan (~UTC+9)

	// 1. Solar Noon on Spring Equinox (March 20, 2026, 11:45 JST = 02:45 UTC)
	noonTime := time.Date(2026, 3, 20, 2, 45, 0, 0, time.UTC)
	solarNoon := CalculateSolarPhase(lat, lon, noonTime)

	if solarNoon.PhaseName != "DAYLIGHT" {
		t.Errorf("expected DAYLIGHT at solar noon, got %s", solarNoon.PhaseName)
	}
	if solarNoon.ElevationDegrees < 40.0 {
		t.Errorf("expected solar elevation > 40° at equinox noon in Morioka, got %f", solarNoon.ElevationDegrees)
	}
	if solarNoon.CrepuscularRiskMultiplier > 0.90 {
		t.Errorf("expected low crepuscular multiplier (<0.90) at midday, got %f", solarNoon.CrepuscularRiskMultiplier)
	}
	if solarNoon.IsCrepuscularPeakRisk {
		t.Errorf("expected IsCrepuscularPeakRisk false at midday")
	}

	// 2. Midnight (March 20, 2026, 00:00 JST = March 19, 15:00 UTC)
	midnightTime := time.Date(2026, 3, 19, 15, 0, 0, 0, time.UTC)
	solarMidnight := CalculateSolarPhase(lat, lon, midnightTime)

	if solarMidnight.PhaseName != "NIGHT" {
		t.Errorf("expected NIGHT at midnight, got %s", solarMidnight.PhaseName)
	}
	if solarMidnight.ElevationDegrees > -20.0 {
		t.Errorf("expected negative elevation < -20° at midnight, got %f", solarMidnight.ElevationDegrees)
	}
	if solarMidnight.IsCrepuscularPeakRisk {
		t.Errorf("expected IsCrepuscularPeakRisk false at midnight")
	}

	// 3. Dusk Twilight Window (~18:00 JST = 09:00 UTC)
	duskTime := time.Date(2026, 3, 20, 9, 0, 0, 0, time.UTC)
	solarDusk := CalculateSolarPhase(lat, lon, duskTime)

	if !solarDusk.IsCrepuscularPeakRisk {
		t.Errorf("expected IsCrepuscularPeakRisk true during dusk twilight (elevation: %f)", solarDusk.ElevationDegrees)
	}
	if solarDusk.CrepuscularRiskMultiplier < 1.40 {
		t.Errorf("expected elevated crepuscular multiplier >= 1.40 at dusk, got %f", solarDusk.CrepuscularRiskMultiplier)
	}
}

func TestMeteorologyAndFloodedRiverBarrierDeflection(t *testing.T) {
	engine := NewEngine()

	// 1. Routine conditions
	meteo := engine.GetCurrentMeteorology("JAPAN")
	if meteo.WindSpeedKmh <= 0 {
		t.Errorf("expected positive wind speed, got %f", meteo.WindSpeedKmh)
	}
	if meteo.IsFloodedRiverRisk {
		t.Errorf("expected false for routine precipitation, got true")
	}

	// 2. Extreme precipitation flooded river test
	extremeWeather := &model.WeatherConditions{
		WindSpeedKmh:     18.0,
		WindDirectionDeg: 270.0,
		MoonIllumination: 0.15,
		PrecipitationMmH: 35.0, // Torrential rain causing river barrier flooding
	}

	req := model.GeneratePlumeGeofencesRequest{
		Attractants: []model.AttractantHub{
			{
				Id:        "plot_flood_01",
				Type:      "ELEPHANT_CROP",
				Latitude:  13.5200,
				Longitude: 101.4500,
				FruitLoad: "ABUNDANT",
			},
		},
		Weather: extremeWeather,
	}

	resp := engine.GenerateScentPlumeGeofences(req)
	if resp.Count == 0 {
		t.Fatalf("expected generated zones, got 0")
	}

	zone := resp.Zones[0]
	hasRiverBarrierFactor := false
	for _, factor := range zone.CausalFactors {
		if strings.HasPrefix(factor, "Flooded River Barrier") {
			hasRiverBarrierFactor = true
			break
		}
	}

	if !hasRiverBarrierFactor {
		t.Errorf("expected Flooded River Barrier causal factor under 35mm/h rain, got: %+v", zone.CausalFactors)
	}

	// Tier should elevate due to bottleneck compression
	if zone.Tier != "TIER_3_DANGER" {
		t.Errorf("expected TIER_3_DANGER under flooded river bottleneck compression, got %s", zone.Tier)
	}
}

func TestDownwindPlumeDispersionShape(t *testing.T) {
	engine := NewEngine()

	// Wind from West (270 deg) -> blows TOWARD East (90 deg)
	weather := &model.WeatherConditions{
		WindSpeedKmh:     20.0,
		WindDirectionDeg: 270.0,
		MoonIllumination: 0.50,
		PrecipitationMmH: 0.0,
	}

	hubLat := 39.7000
	hubLon := 141.1500

	resp := engine.GenerateScentPlumeGeofences(model.GeneratePlumeGeofencesRequest{
		Attractants: []model.AttractantHub{
			{
				Id:        "kaki_wind_test",
				Type:      "KAKI_TREE",
				Latitude:  hubLat,
				Longitude: hubLon,
				FruitLoad: "ABUNDANT",
			},
		},
		Weather: weather,
	})

	if resp.Count == 0 {
		t.Fatalf("expected zone, got 0")
	}

	zone := resp.Zones[0]
	poly := zone.PolygonGeoJson
	if poly == nil || len(poly.Coordinates) == 0 {
		t.Fatalf("expected non-nil polygon coordinates")
	}

	ring := poly.Coordinates[0]
	// Verify closed loop
	if ring[0][0] != ring[len(ring)-1][0] || ring[0][1] != ring[len(ring)-1][1] {
		t.Errorf("polygon ring must be closed (start == end)")
	}

	// Find maximum Eastern displacement (downwind) and Western displacement (upwind)
	var maxEast, minWest float64
	minWest = 999.0
	for _, pt := range ring {
		lon := pt[0]
		dLon := lon - hubLon
		if dLon > maxEast {
			maxEast = dLon
		}
		if dLon < minWest {
			minWest = dLon
		}
	}

	// Downwind East displacement should be significantly greater than upwind West displacement
	upwindDist := math.Abs(minWest)
	downwindDist := math.Abs(maxEast)
	if downwindDist <= upwindDist*1.5 {
		t.Errorf("expected downwind elongation > 1.5x upwind compression, got downwind=%f, upwind=%f", downwindDist, upwindDist)
	}
}

func TestTemporalRiskDecayModeling(t *testing.T) {
	// 1. Roadkill / Kill site decay: lambda = ln(2)/24 ~= 0.02888
	hubFresh := model.AttractantHub{
		Id:              "rk_fresh",
		Type:            "ROADKILL",
		HoursSinceDeath: 0.0,
	}
	_, tier0, _, risk0 := calculateHubProperties(hubFresh)
	if tier0 != "TIER_3_DANGER" || risk0 < 0.89 {
		t.Errorf("expected fresh roadkill risk ~0.90, got %f", risk0)
	}

	// At 24h (one half-life): risk should halve
	hub24h := model.AttractantHub{
		Id:              "rk_24h",
		Type:            "ROADKILL",
		HoursSinceDeath: 24.0,
	}
	_, _, _, risk24 := calculateHubProperties(hub24h)
	expected24 := risk0 * 0.50
	if math.Abs(risk24-expected24) > 0.05 {
		t.Errorf("expected risk after 24h to be ~%f, got %f", expected24, risk24)
	}

	// At 72h (3 half-lives): risk drops to ~12.5%
	hub72h := model.AttractantHub{
		Id:              "rk_72h",
		Type:            "ROADKILL",
		HoursSinceDeath: 72.0,
	}
	_, _, _, risk72 := calculateHubProperties(hub72h)
	if risk72 > 0.15 {
		t.Errorf("expected risk after 72h to decay below 0.15, got %f", risk72)
	}
}

func TestGeofenceRecalculationAndActiveQuery(t *testing.T) {
	engine := NewEngine()
	engine.LoadFixtureZones()

	// Initial active zones
	initialActive := engine.GetActiveGeofences(39.7020, 141.1520, 10.0)
	if len(initialActive) == 0 {
		t.Fatalf("expected initial active zones in Japan, got 0")
	}

	// Trigger emergency 5-minute recalibration
	recalcResp := engine.Recalculate(model.GeofenceRecalculateRequest{
		TriggerEntityRef:      "bear_sighting_matsugaeda",
		TargetCenterLatitude:  39.7050,
		TargetCenterLongitude: 141.1550,
		RadiusKm:              5.0,
	})

	if recalcResp.RecalculatedGeofencesCount <= 0 {
		t.Errorf("expected positive recalculated count, got %d", recalcResp.RecalculatedGeofencesCount)
	}

	// Query active geofences around the newly recalibrated zone
	activeAfter := engine.GetActiveGeofences(39.7050, 141.1550, 5.0)
	foundRecalibrated := false
	for _, z := range activeAfter {
		if z.Id == "recalc_bear_sighting_matsugaeda" {
			foundRecalibrated = true
			if z.Tier != "TIER_3_DANGER" {
				t.Errorf("expected TIER_3_DANGER for emergency recalibration, got %s", z.Tier)
			}
			if z.ActiveRiskScore < 0.95 {
				t.Errorf("expected high active risk score, got %f", z.ActiveRiskScore)
			}
		}
	}

	if !foundRecalibrated {
		t.Errorf("expected to find recalibrated zone in active geofences query")
	}
}

func TestGraduatedConcentricBufferZones(t *testing.T) {
	engine := NewEngine()

	core := model.ActiveGeofenceZone{
		Id:                "test_core_zone",
		CausalExplanation: "Morioka Test Critical Incursion Zone",
		Tier:              "TIER_3_DANGER",
		TargetSpecies:     []string{"Asian Black Bear"},
		RadiusMeters:      1000.0,
		Centroid:          model.ActiveGeofencePoint{Latitude: 39.7020, Longitude: 141.1520},
	}

	buffers := GenerateConcentricBufferZones(core)
	if len(buffers) != 2 {
		t.Fatalf("expected 2 buffer zones (Tier 2 and Tier 1), got %d", len(buffers))
	}

	tier2 := buffers[0]
	if tier2.Tier != "TIER_2_WARNING" {
		t.Errorf("expected tier 2 warning, got %s", tier2.Tier)
	}
	if math.Abs(tier2.RadiusMeters-1750.0) > 1.0 {
		t.Errorf("expected tier 2 radius ~1750m (1.75x), got %f", tier2.RadiusMeters)
	}

	tier1 := buffers[1]
	if tier1.Tier != "TIER_1_NOTICE" {
		t.Errorf("expected tier 1 notice, got %s", tier1.Tier)
	}
	if math.Abs(tier1.RadiusMeters-2600.0) > 1.0 {
		t.Errorf("expected tier 1 radius ~2600m (2.60x), got %f", tier1.RadiusMeters)
	}

	// Clear existing default seeded zones to isolate test_core_zone
	engine.activeZones = make(map[string]model.ActiveGeofenceZone)
	engine.registerZoneWithGraduatedBuffers(core)

	// Test EvaluateLocationTier:
	// 1. Inside core (lat: 39.7020, lon: 141.1520 -> 0m)
	tier, dist, matchedID := engine.EvaluateLocationTier(39.7020, 141.1520)
	if tier != "TIER_3_DANGER" || matchedID != "test_core_zone" {
		t.Errorf("expected TIER_3_DANGER at centroid, got %s (id: %s)", tier, matchedID)
	}

	// 2. Just outside core but inside Tier 2 buffer (~1.5km north from centroid: 500m from boundary <= 1200m)
	latT2 := 39.7020 + (1500.0 / 111320.0)
	tier, dist, _ = engine.EvaluateLocationTier(latT2, 141.1520)
	if tier != "TIER_2_WARNING" {
		t.Errorf("expected TIER_2_WARNING at 1.5km, got %s (dist: %.1fm)", tier, dist)
	}

	// 3. Outside Tier 2 but inside Tier 1 buffer (~2.6km north from centroid: 1600m from boundary <= 2500m)
	latT1 := 39.7020 + (2600.0 / 111320.0)
	tier, dist, _ = engine.EvaluateLocationTier(latT1, 141.1520)
	if tier != "TIER_1_NOTICE" {
		t.Errorf("expected TIER_1_NOTICE at 2.6km, got %s (dist: %.1fm)", tier, dist)
	}

	// 4. Far away (>4.0km north from centroid: 3000m from boundary > 2500m)
	latSafe := 39.7020 + (4200.0 / 111320.0)
	tier, dist, _ = engine.EvaluateLocationTier(latSafe, 141.1520)
	if tier != "SAFE" {
		t.Errorf("expected SAFE at 4.2km, got %s (dist: %.1fm)", tier, dist)
	}
}

func BenchmarkEvaluateLocationTier(b *testing.B) {
	engine := NewEngine()
	core := model.ActiveGeofenceZone{
		Id:                "bench_zone",
		CausalExplanation: "Benchmark Zone",
		Tier:              "TIER_3_DANGER",
		TargetSpecies:     []string{"Asian Black Bear"},
		RadiusMeters:      1000.0,
		Centroid:          model.ActiveGeofencePoint{Latitude: 39.7020, Longitude: 141.1520},
	}
	engine.registerZoneWithGraduatedBuffers(core)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		engine.EvaluateLocationTier(39.7050, 141.1550)
	}
}

func BenchmarkCalculateSolarPhase(b *testing.B) {
	now := time.Now()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		CalculateSolarPhase(39.7020, 141.1520, now)
	}
}

