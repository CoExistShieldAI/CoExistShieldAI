package geofence

import (
	"fmt"
	"math"
	"time"

	"ai.coexistshield/backend/internal/model"
)

// GenerateForecastPlumeGeofences computes projected time-variant hazard zones
// across multi-hour future forecast steps, simulating scent plume rotation and corridor drift.
func (e *Engine) GenerateForecastPlumeGeofences(
	req model.ForecastPlumeGeofencesRequest,
	forecastPoints []model.HourlyForecastPoint,
) model.ForecastPlumeGeofencesResponse {

	now := time.Now().UTC()
	targetHours := req.TargetHours
	if len(targetHours) == 0 {
		targetHours = []int{1, 3, 6, 12}
	}

	// Index forecast points by offset hours
	forecastMap := make(map[int]model.HourlyForecastPoint)
	for _, fp := range forecastPoints {
		forecastMap[fp.OffsetHours] = fp
	}

	var steps []model.ForecastZoneStep

	for _, offsetH := range targetHours {
		futureTime := now.Add(time.Duration(offsetH) * time.Hour)

		// Get forecast weather or fallback to realistic regional baseline drift
		var windSpeed, windDir, precip float64
		if fp, ok := forecastMap[offsetH]; ok {
			windSpeed = fp.WindSpeedKmh
			windDir = fp.WindDirectionDeg
			precip = fp.PrecipitationMmH
		} else {
			// Deterministic gradual wind direction shift (e.g. +15 deg per hour) simulating weather front
			windSpeed = 12.0 + 3.0*math.Sin(float64(offsetH)*0.5)
			windDir = math.Mod(270.0+float64(offsetH)*15.0, 360.0)
			precip = 0.0
		}

		// Lunar and solar calculations for future time
		moon := CalculateMoonPhase(futureTime)
		lunarFactor := 1.0 + 0.6*(1.0-moon.Illumination)

		// Downwind plume distortion
		blowsTowardDeg := math.Mod(windDir+180.0, 360.0)
		blowsTowardRad := blowsTowardDeg * (math.Pi / 180.0)

		kDownwind := math.Min(2.5, 1.0+0.08*windSpeed)
		kUpwind := math.Max(0.4, 1.0-0.04*windSpeed)
		isFloodedRiver := precip >= 25.0

		for idx, hub := range req.Attractants {
			baseRadius, tier, species, baseRisk := calculateHubProperties(hub)

			solar := CalculateSolarPhase(hub.Latitude, hub.Longitude, futureTime)
			crepuscularFactor := 0.70 + 0.30*solar.CrepuscularRiskMultiplier

			effectiveRisk := math.Min(1.0, baseRisk*lunarFactor*crepuscularFactor)
			if isFloodedRiver {
				effectiveRisk = math.Min(1.0, effectiveRisk*1.25)
				baseRadius *= 1.35
			}

			if effectiveRisk >= 0.85 {
				tier = "TIER_3_DANGER"
			} else if effectiveRisk >= 0.60 && tier != "TIER_3_DANGER" {
				tier = "TIER_2_WARNING"
			}

			poly := generatePlumePolygon(
				hub.Latitude,
				hub.Longitude,
				baseRadius,
				kDownwind,
				kUpwind,
				blowsTowardRad,
				isFloodedRiver,
			)

			// Apply OpenStreetMap infrastructure barriers & riparian corridor alignment
			osmCtx := DefaultOsmContextForRegion(hub.Latitude, hub.Longitude)
			poly = ApplyOsmSpatialConstraints(poly, osmCtx, hub.Latitude, hub.Longitude)

			causalFactors := []string{
				fmt.Sprintf("Forecast Offset: +%dh (%s)", offsetH, futureTime.Format(time.RFC3339)),
				fmt.Sprintf("Projected Wind: %.1f km/h (bearing %.0f°)", windSpeed, blowsTowardDeg),
				fmt.Sprintf("Solar Phase: %s (Crepuscular Boost: %.2fx)", solar.PhaseName, solar.CrepuscularRiskMultiplier),
				fmt.Sprintf("Moon Illumination: %.2f", moon.Illumination),
			}
			if osmCtx.HasDividedMotorway {
				causalFactors = append(causalFactors, fmt.Sprintf("OSM Barrier: Highway limits cross-corridor spillover"))
			}

			zoneDto := model.GeofenceZoneDto{
				Id:              fmt.Sprintf("forecast_%s_+%dh_%d", hub.Id, offsetH, idx+1),
				Tier:            tier,
				TargetSpecies:   species,
				PolygonGeoJson:  poly,
				ActiveRiskScore: effectiveRisk,
				CausalFactors:   causalFactors,
				ExpiresAt:       futureTime.Add(24 * time.Hour).Format(time.RFC3339),
				CreatedAt:       now.Format(time.RFC3339),
			}

			steps = append(steps, model.ForecastZoneStep{
				OffsetHours:          offsetH,
				ForecastTime:         futureTime.Format(time.RFC3339),
				Zone:                 zoneDto,
				WindBearingDeg:       roundToDecimals(blowsTowardDeg, 1),
				WindSpeedKmh:         roundToDecimals(windSpeed, 1),
				CrepuscularRiskBoost: solar.CrepuscularRiskMultiplier,
				SolarPhase:           solar,
			})
		}
	}

	return model.ForecastPlumeGeofencesResponse{
		TimelineSteps: steps,
		TotalSteps:    len(steps),
		GeneratedAt:   now.Format(time.RFC3339),
	}
}
