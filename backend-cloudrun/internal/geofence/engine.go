package geofence

import (
	"fmt"
	"math"
	"time"

	"ai.coexistshield/backend/internal/model"
)

const (
	EarthRadiusMeters = 6378137.0
	NumPolygonPoints  = 32 // Discretized points per polygon boundary ring
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

// GenerateScentPlumeGeofences synthesizes active attractant hubs, wind-driven scent plumes,
// and lunar darkness to generate 3-tier RFC 7946 GeoJSON polygons.
func (e *Engine) GenerateScentPlumeGeofences(req model.GeneratePlumeGeofencesRequest) model.GeneratePlumeGeofencesResponse {
	weather := req.Weather
	if weather == nil {
		weather = &model.WeatherConditions{
			WindSpeedKmh:     10.0,
			WindDirectionDeg: 270.0, // Westerly wind (blowing from West toward East)
			MoonIllumination: 0.20,
			PrecipitationMmH: 0.0,
		}
	}

	// 1. Lunar Scaling Multiplier: C_moon = 1.0 + 0.6 * (1.0 - MoonIllumination)
	// Elevates risk up to 1.6x on New Moon nights (0.0 illumination)
	lunarFactor := 1.0 + 0.6*(1.0-math.Min(math.Max(weather.MoonIllumination, 0.0), 1.0))

	// 2. Wind Scent-Plume distortion factors
	// Bearing theta_wind is the direction wind blows TOWARD: (WindDirectionDeg + 180) % 360
	blowsTowardDeg := math.Mod(weather.WindDirectionDeg+180.0, 360.0)
	blowsTowardRad := blowsTowardDeg * (math.Pi / 180.0)

	// Downwind elongation: k_downwind = 1.0 + 0.08 * V_wind (clamped max 2.5x)
	kDownwind := math.Min(2.5, 1.0+0.08*weather.WindSpeedKmh)
	// Upwind compression: k_upwind = max(0.4, 1.0 - 0.04 * V_wind)
	kUpwind := math.Max(0.4, 1.0-0.04*weather.WindSpeedKmh)

	var zones []model.GeofenceZoneDto

	for idx, hub := range req.Attractants {
		baseRadius, tier, species, baseRisk := calculateHubProperties(hub)

		// Scale risk by lunar factor
		effectiveRisk := math.Min(1.0, baseRisk*lunarFactor)
		if effectiveRisk >= 0.85 {
			tier = "TIER_3_DANGER"
		} else if effectiveRisk >= 0.60 && tier != "TIER_3_DANGER" {
			tier = "TIER_2_WARNING"
		}

		polygon := generatePlumePolygon(
			hub.Latitude,
			hub.Longitude,
			baseRadius,
			kDownwind,
			kUpwind,
			blowsTowardRad,
		)

		causalFactors := []string{
			fmt.Sprintf("Attractant: %s", hub.Type),
			fmt.Sprintf("Wind: %.1f km/h (bearing %.0f°)", weather.WindSpeedKmh, blowsTowardDeg),
			fmt.Sprintf("Moon Illumination: %.2f (Risk Multiplier: %.2fx)", weather.MoonIllumination, lunarFactor),
		}
		if hub.FruitLoad != "" {
			causalFactors = append(causalFactors, fmt.Sprintf("Fruit Load: %s", hub.FruitLoad))
		}

		zones = append(zones, model.GeofenceZoneDto{
			Id:              fmt.Sprintf("geo_%s_%d", hub.Id, idx+1),
			Tier:            tier,
			TargetSpecies:   species,
			PolygonGeoJson:  polygon,
			ActiveRiskScore: effectiveRisk,
			CausalFactors:   causalFactors,
			ExpiresAt:       time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339),
			CreatedAt:       time.Now().UTC().Format(time.RFC3339),
		})
	}

	return model.GeneratePlumeGeofencesResponse{
		Zones:     zones,
		Count:     len(zones),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func calculateHubProperties(hub model.AttractantHub) (float64, string, []string, float64) {
	switch hub.Type {
	case "KAKI_TREE":
		// Base: 500m, scaled to 1000m if ABUNDANT and Top 5% Hazard
		radius := 500.0
		risk := 0.65
		if hub.FruitLoad == "ABUNDANT" {
			radius = 750.0
			risk = 0.75
		}
		if hub.IsTop5Hazard {
			radius = 1000.0
			risk = 0.88
		}
		return radius, "TIER_2_WARNING", []string{"BEAR"}, risk

	case "ELEPHANT_CROP":
		// Base: 1500m, scaled to 3000m during harvest ripening
		radius := 1500.0
		risk := 0.70
		if hub.FruitLoad == "ABUNDANT" {
			radius = 3000.0
			risk = 0.86
		}
		return radius, "TIER_2_WARNING", []string{"ELEPHANT"}, risk

	case "ROADKILL", "KILL_SITE":
		// Base: 800m, temporal decay multiplier lambda(t) = e^(-t / 24)
		decayMultiplier := 1.0
		if hub.HoursSinceDeath > 0 {
			decayMultiplier = math.Exp(-hub.HoursSinceDeath / 24.0)
		}
		radius := 800.0 * math.Max(0.5, decayMultiplier)
		risk := 0.90 * decayMultiplier
		species := []string{"TIGER", "BEAR"}
		return radius, "TIER_3_DANGER", species, risk

	default:
		return 500.0, "TIER_1_NOTICE", []string{"UNKNOWN"}, 0.50
	}
}

// generatePlumePolygon creates a deformed elliptical polygon elongated in the downwind direction
func generatePlumePolygon(lat, lon, baseRadius, kDownwind, kUpwind, blowsTowardRad float64) *model.GeoJSONPolygonGeometry {
	var ring [][]float64

	// Discretize circle into NumPolygonPoints
	for i := 0; i <= NumPolygonPoints; i++ {
		angle := float64(i) * (2.0 * math.Pi / float64(NumPolygonPoints))

		// Project relative angle against the wind direction
		// deltaAngle = 0: directly downwind (elongated)
		// deltaAngle = pi: directly upwind (compressed)
		deltaAngle := angle - blowsTowardRad

		// Compute distance factor based on downwind/upwind alignment
		// cos(deltaAngle) in [-1, 1]. When cos == 1 (downwind), factor = kDownwind. When cos == -1 (upwind), factor = kUpwind.
		weight := (math.Cos(deltaAngle) + 1.0) / 2.0 // normalized to [0, 1]
		radialDist := baseRadius * (kUpwind + weight*(kDownwind-kUpwind))

		// Displace in meters (east/north)
		dx := radialDist * math.Sin(angle) // Easting
		dy := radialDist * math.Cos(angle) // Northing

		// Convert meters to delta latitude and longitude
		dLat := (dy / EarthRadiusMeters) * (180.0 / math.Pi)
		dLon := (dx / (EarthRadiusMeters * math.Cos(lat*(math.Pi/180.0)))) * (180.0 / math.Pi)

		pointLat := lat + dLat
		pointLon := lon + dLon

		// GeoJSON RFC 7946 coordinates format: [longitude, latitude]
		ring = append(ring, []float64{
			roundToDecimals(pointLon, 6),
			roundToDecimals(pointLat, 6),
		})
	}

	return &model.GeoJSONPolygonGeometry{
		Type:        "Polygon",
		Coordinates: [][][]float64{ring},
	}
}

func roundToDecimals(val float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(val*pow) / pow
}

func (e *Engine) Recalculate(req model.GeofenceRecalculateRequest) model.GeofenceRecalculateResponse {
	tiers := []string{"TIER_1_NOTICE", "TIER_2_WARNING"}
	if req.TriggerEntityRef != "" {
		tiers = append(tiers, "TIER_3_DANGER")
	}

	return model.GeofenceRecalculateResponse{
		RecalculatedGeofencesCount: 3,
		AffectedTiers:              tiers,
		UpdatedAt:                  time.Now().UTC().Format(time.RFC3339),
	}
}

func (e *Engine) GetTopHazardPersimmonTrees(municipalityCode string, limit int) model.TopHazardsResponse {
	name := "Morioka City"
	if municipalityCode == "03201" {
		name = "Morioka City (岩手県盛岡市)"
	}

	trees := []model.PriorityKakiTreeDto{
		{
			TreeId:                         "kt_1120",
			HazardScore:                    0.95,
			DistanceToSchoolRouteMeters:    120,
			EstimatedBearAttractantRadiusM: 800,
			RecommendedClearanceOrder:      1,
		},
		{
			TreeId:                         "kt_0842",
			HazardScore:                    0.91,
			DistanceToSchoolRouteMeters:    240,
			EstimatedBearAttractantRadiusM: 650,
			RecommendedClearanceOrder:      2,
		},
		{
			TreeId:                         "kt_0319",
			HazardScore:                    0.88,
			DistanceToSchoolRouteMeters:    310,
			EstimatedBearAttractantRadiusM: 500,
			RecommendedClearanceOrder:      3,
		},
	}

	if limit > 0 && limit < len(trees) {
		trees = trees[:limit]
	}

	return model.TopHazardsResponse{
		MunicipalityName:             name,
		TotalCatalogedTrees:          840,
		Top5PercentCutoffHazardScore: 0.89,
		PriorityTrees:                trees,
	}
}
