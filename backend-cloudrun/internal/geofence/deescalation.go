package geofence

import (
	"fmt"
	"time"

	"ai.coexistshield/backend/internal/model"
)

// DeescalateActiveZones evaluates all active hazard zones and applies time-based
// exponential spatial contraction and tier demotion in the absence of fresh verified incidents.
// This prevents alert fatigue while preserving heightened protection.
func (e *Engine) DeescalateActiveZones(now time.Time) []model.DeescalationResult {
	e.mu.Lock()
	defer e.mu.Unlock()

	if now.IsZero() {
		now = time.Now().UTC()
	}

	var results []model.DeescalationResult
	zonesToDelete := make([]string, 0)

	for id, zone := range e.activeZones {
		// Calculate elapsed hours since zone generation
		var hoursElapsed float64 = 0.0
		if zone.GeneratedAt != "" {
			tGen, err := time.Parse(time.RFC3339, zone.GeneratedAt)
			if err == nil {
				hoursElapsed = now.Sub(tGen).Hours()
			}
		}

		// Check explicit expiration
		isExpired := false
		if zone.ValidUntil != "" {
			tExp, err := time.Parse(time.RFC3339, zone.ValidUntil)
			if err == nil && now.After(tExp) {
				isExpired = true
			}
		}

		if isExpired || hoursElapsed >= 72.0 {
			zonesToDelete = append(zonesToDelete, id)
			results = append(results, model.DeescalationResult{
				ZoneId:            id,
				OriginalRadiusM:   zone.RadiusMeters,
				CurrentRadiusM:    0,
				ContractionRatio:  0.0,
				OriginalTier:      zone.Tier,
				CurrentTier:       "EXPIRED",
				HoursSinceTrigger: roundToDecimals(hoursElapsed, 1),
				Status:            "EXPIRED",
			})
			continue
		}

		// Calculate progressive de-escalation ratio and tier demotion
		var contractionRatio float64 = 1.0
		currentTier := zone.Tier
		status := "ACTIVE"

		switch {
		case hoursElapsed >= 48.0:
			// 48h elapsed: contract radius by 50% and demote tier
			contractionRatio = 0.50
			status = "DEMOTED"
			if zone.Tier == "TIER_3_DANGER" {
				currentTier = "TIER_2_WARNING"
			} else if zone.Tier == "TIER_2_WARNING" {
				currentTier = "TIER_1_NOTICE"
			}
		case hoursElapsed >= 24.0:
			// 24h elapsed: contract radius by 30% (ratio = 0.70)
			contractionRatio = 0.70
			status = "CONTRACTED"
			if zone.Tier == "TIER_3_DANGER" {
				currentTier = "TIER_2_WARNING"
				status = "DEMOTED"
			}
		case hoursElapsed >= 12.0:
			// 12h elapsed: initial contraction by 15% (ratio = 0.85)
			contractionRatio = 0.85
			status = "CONTRACTED"
		default:
			// Fresh zone (<12h): active at 100% capacity
			contractionRatio = 1.0
			status = "ACTIVE"
		}

		newRadius := zone.RadiusMeters * contractionRatio
		newPolygon := shrinkPolygon(zone.PolygonGeoJson, zone.Centroid.Latitude, zone.Centroid.Longitude, contractionRatio)
		newCoords := extractCoordinatesFromPolygon(newPolygon)

		// Update zone in engine
		zone.Tier = currentTier
		zone.RadiusMeters = newRadius
		zone.PolygonGeoJson = newPolygon
		zone.Coordinates = newCoords
		zone.ActiveRiskScore = roundToDecimals(zone.ActiveRiskScore*contractionRatio, 2)
		zone.CausalExplanation = fmt.Sprintf("%s (Progressive de-escalation active: %.0f%% of initial perimeter at +%.1fh)",
			zone.CausalExplanation, contractionRatio*100.0, hoursElapsed)
		e.activeZones[id] = zone

		results = append(results, model.DeescalationResult{
			ZoneId:            id,
			OriginalRadiusM:   roundToDecimals(zone.RadiusMeters/contractionRatio, 1),
			CurrentRadiusM:    roundToDecimals(newRadius, 1),
			ContractionRatio:  roundToDecimals(contractionRatio, 2),
			OriginalTier:      zone.Tier,
			CurrentTier:       currentTier,
			HoursSinceTrigger: roundToDecimals(hoursElapsed, 1),
			Status:            status,
		})
	}

	// Purge expired zones
	for _, id := range zonesToDelete {
		delete(e.activeZones, id)
	}

	return results
}

// shrinkPolygon scales polygon coordinates relative to its centroid by the specified ratio
func shrinkPolygon(poly *model.GeoJSONPolygonGeometry, cLat, cLon, ratio float64) *model.GeoJSONPolygonGeometry {
	if poly == nil || len(poly.Coordinates) == 0 || len(poly.Coordinates[0]) == 0 {
		return poly
	}

	ring := poly.Coordinates[0]
	newRing := make([][]float64, len(ring))

	for i, pt := range ring {
		if len(pt) < 2 {
			newRing[i] = pt
			continue
		}
		lon := pt[0]
		lat := pt[1]

		dLon := lon - cLon
		dLat := lat - cLat

		shrunkLon := cLon + dLon*ratio
		shrunkLat := cLat + dLat*ratio

		newRing[i] = []float64{
			roundToDecimals(shrunkLon, 6),
			roundToDecimals(shrunkLat, 6),
		}
	}

	return &model.GeoJSONPolygonGeometry{
		Type:        "Polygon",
		Coordinates: [][][]float64{newRing},
	}
}
