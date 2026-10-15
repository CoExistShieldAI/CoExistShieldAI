package geofence

import (
	"testing"
	"time"

	"ai.coexistshield/backend/internal/model"
)

func TestDeescalateActiveZones(t *testing.T) {
	engine := NewEngine()

	now := time.Now().UTC()
	poly := generatePlumePolygon(39.7020, 141.1520, 1000.0, 1.0, 1.0, 0.0, false)
	coords := extractCoordinatesFromPolygon(poly)

	// 1. Fresh zone (generated 2h ago)
	engine.activeZones["zone_fresh"] = model.ActiveGeofenceZone{
		Id:              "zone_fresh",
		Tier:            "TIER_3_DANGER",
		RadiusMeters:    1000.0,
		PolygonGeoJson:  poly,
		Coordinates:     coords,
		Centroid:        model.ActiveGeofencePoint{Latitude: 39.7020, Longitude: 141.1520},
		ActiveRiskScore: 0.95,
		GeneratedAt:     now.Add(-2 * time.Hour).Format(time.RFC3339),
		ValidUntil:      now.Add(70 * time.Hour).Format(time.RFC3339),
	}

	// 2. 15h old zone (should contract to 85%)
	engine.activeZones["zone_15h"] = model.ActiveGeofenceZone{
		Id:              "zone_15h",
		Tier:            "TIER_2_WARNING",
		RadiusMeters:    1000.0,
		PolygonGeoJson:  poly,
		Coordinates:     coords,
		Centroid:        model.ActiveGeofencePoint{Latitude: 39.7020, Longitude: 141.1520},
		ActiveRiskScore: 0.75,
		GeneratedAt:     now.Add(-15 * time.Hour).Format(time.RFC3339),
		ValidUntil:      now.Add(57 * time.Hour).Format(time.RFC3339),
	}

	// 3. 28h old zone (should contract to 70% and demote TIER_3 to TIER_2)
	engine.activeZones["zone_28h"] = model.ActiveGeofenceZone{
		Id:              "zone_28h",
		Tier:            "TIER_3_DANGER",
		RadiusMeters:    1000.0,
		PolygonGeoJson:  poly,
		Coordinates:     coords,
		Centroid:        model.ActiveGeofencePoint{Latitude: 39.7020, Longitude: 141.1520},
		ActiveRiskScore: 0.90,
		GeneratedAt:     now.Add(-28 * time.Hour).Format(time.RFC3339),
		ValidUntil:      now.Add(44 * time.Hour).Format(time.RFC3339),
	}

	// 4. 50h old zone (should contract to 50% and demote to TIER_1)
	engine.activeZones["zone_50h"] = model.ActiveGeofenceZone{
		Id:              "zone_50h",
		Tier:            "TIER_2_WARNING",
		RadiusMeters:    1000.0,
		PolygonGeoJson:  poly,
		Coordinates:     coords,
		Centroid:        model.ActiveGeofencePoint{Latitude: 39.7020, Longitude: 141.1520},
		ActiveRiskScore: 0.70,
		GeneratedAt:     now.Add(-50 * time.Hour).Format(time.RFC3339),
		ValidUntil:      now.Add(22 * time.Hour).Format(time.RFC3339),
	}

	// 5. Expired zone (75h old)
	engine.activeZones["zone_expired"] = model.ActiveGeofenceZone{
		Id:              "zone_expired",
		Tier:            "TIER_1_NOTICE",
		RadiusMeters:    1000.0,
		PolygonGeoJson:  poly,
		Coordinates:     coords,
		Centroid:        model.ActiveGeofencePoint{Latitude: 39.7020, Longitude: 141.1520},
		ActiveRiskScore: 0.50,
		GeneratedAt:     now.Add(-75 * time.Hour).Format(time.RFC3339),
		ValidUntil:      now.Add(-3 * time.Hour).Format(time.RFC3339),
	}

	results := engine.DeescalateActiveZones(now)

	if len(results) < 5 {
		t.Fatalf("expected 5 deescalation results, got %d", len(results))
	}

	resMap := make(map[string]model.DeescalationResult)
	for _, r := range results {
		resMap[r.ZoneId] = r
	}

	// Verify zone_fresh
	if resMap["zone_fresh"].Status != "ACTIVE" || resMap["zone_fresh"].ContractionRatio != 1.0 {
		t.Errorf("expected ACTIVE at 1.0 ratio for zone_fresh, got status=%s, ratio=%f",
			resMap["zone_fresh"].Status, resMap["zone_fresh"].ContractionRatio)
	}

	// Verify zone_15h
	if resMap["zone_15h"].Status != "CONTRACTED" || resMap["zone_15h"].ContractionRatio != 0.85 {
		t.Errorf("expected CONTRACTED at 0.85 ratio for zone_15h, got status=%s, ratio=%f",
			resMap["zone_15h"].Status, resMap["zone_15h"].ContractionRatio)
	}

	// Verify zone_28h
	if resMap["zone_28h"].Status != "DEMOTED" || resMap["zone_28h"].CurrentTier != "TIER_2_WARNING" {
		t.Errorf("expected DEMOTED to TIER_2_WARNING for zone_28h, got status=%s, tier=%s",
			resMap["zone_28h"].Status, resMap["zone_28h"].CurrentTier)
	}

	// Verify zone_50h
	if resMap["zone_50h"].Status != "DEMOTED" || resMap["zone_50h"].CurrentTier != "TIER_1_NOTICE" {
		t.Errorf("expected DEMOTED to TIER_1_NOTICE for zone_50h, got status=%s, tier=%s",
			resMap["zone_50h"].Status, resMap["zone_50h"].CurrentTier)
	}

	// Verify zone_expired is removed from activeZones
	if _, exists := engine.activeZones["zone_expired"]; exists {
		t.Errorf("expected zone_expired to be purged from engine.activeZones")
	}
}
