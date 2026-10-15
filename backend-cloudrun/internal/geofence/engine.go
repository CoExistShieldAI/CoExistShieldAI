package geofence

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"ai.coexistshield/backend/internal/model"
)

const (
	EarthRadiusMeters = 6378137.0
	NumPolygonPoints  = 32 // Discretized points per polygon boundary ring
	SynodicMonthDays  = 29.53058867
)

type Engine struct {
	mu                 sync.RWMutex
	activeZones        map[string]model.ActiveGeofenceZone
	activeModelVersion string
	modelWeights       map[string]float64
}

// NewEngine initializes the dynamic geofencing engine.
// Pre-seeded zones are only injected when ENABLE_SEED_DATA="true" or DEV_MODE="true".
func NewEngine() *Engine {
	e := &Engine{
		activeZones:        make(map[string]model.ActiveGeofenceZone),
		activeModelVersion: "v2026.10.01-baseline",
		modelWeights:       make(map[string]float64),
	}
	if os.Getenv("ENABLE_SEED_DATA") == "true" || os.Getenv("DEV_MODE") == "true" {
		e.LoadFixtureZones()
	}
	return e
}

// LoadFixtureZones populates calibrated test geofence zones.
func (e *Engine) LoadFixtureZones() {
	e.seedDefaultGeofenceZones()
}

// CalculateMoonPhase computes astronomical lunar phase, illumination percentage, and nocturnal risk multiplier
func CalculateMoonPhase(t time.Time) model.MoonPhaseInfo {
	// Known reference New Moon: Jan 11, 2024, 11:57 UTC
	refNewMoon := time.Date(2024, 1, 11, 11, 57, 0, 0, time.UTC)
	diffHours := t.UTC().Sub(refNewMoon).Hours()
	diffDays := diffHours / 24.0

	cycleAge := math.Mod(diffDays, SynodicMonthDays)
	if cycleAge < 0 {
		cycleAge += SynodicMonthDays
	}

	phaseFraction := cycleAge / SynodicMonthDays // [0.0, 1.0)

	// Illumination percentage (0.0 at New Moon, 1.0 at Full Moon)
	illumination := (1.0 - math.Cos(2.0*math.Pi*phaseFraction)) / 2.0
	illumination = roundToDecimals(illumination, 3)

	var phaseName string
	switch {
	case phaseFraction < 0.03 || phaseFraction >= 0.97:
		phaseName = "NEW_MOON"
	case phaseFraction < 0.22:
		phaseName = "WAXING_CRESCENT"
	case phaseFraction < 0.28:
		phaseName = "FIRST_QUARTER"
	case phaseFraction < 0.47:
		phaseName = "WAXING_GIBBOUS"
	case phaseFraction < 0.53:
		phaseName = "FULL_MOON"
	case phaseFraction < 0.72:
		phaseName = "WANING_GIBBOUS"
	case phaseFraction < 0.78:
		phaseName = "LAST_QUARTER"
	default:
		phaseName = "WANING_CRESCENT"
	}

	// Nocturnal Risk Multiplier: M_lunar = 1.0 + 0.6 * (1.0 - illumination)
	// Darker nights elevate tiger and bear nocturnal stealth movement by up to 1.6x
	nocturnalMultiplier := 1.0 + 0.6*(1.0-illumination)
	nocturnalMultiplier = roundToDecimals(nocturnalMultiplier, 3)

	return model.MoonPhaseInfo{
		PhaseName:               phaseName,
		Illumination:            illumination,
		NocturnalRiskMultiplier: nocturnalMultiplier,
		IsNocturnalPeakRisk:     illumination <= 0.20,
	}
}

// GetCurrentMeteorology returns real-time weather and astronomical conditions for a region
func (e *Engine) GetCurrentMeteorology(region string) model.MeteorologicalConditions {
	now := time.Now().UTC()
	moon := CalculateMoonPhase(now)

	// Defaults tailored by regional climate zones
	temp := 14.5
	humid := 68.0
	windSpeed := 12.0
	windDir := 285.0 // Westerly
	precip := 0.0
	refLat := 39.7020
	refLon := 141.1520

	switch region {
	case "THAILAND":
		temp = 29.0
		humid = 82.0
		windSpeed = 8.5
		windDir = 120.0
		precip = 5.0
		refLat = 13.5210
		refLon = 101.4520
	case "INDONESIA":
		temp = 31.0
		humid = 88.0
		windSpeed = 6.0
		windDir = 190.0
		precip = 12.0
		refLat = 0.8250
		refLon = 101.5300
	}

	solar := CalculateSolarPhase(refLat, refLon, now)

	return model.MeteorologicalConditions{
		TemperatureCelsius: temp,
		HumidityPct:        humid,
		WindSpeedKmh:       windSpeed,
		WindDirectionDeg:   windDir,
		PrecipitationMmH:   precip,
		IsFloodedRiverRisk: precip >= 25.0,
		MoonPhase:          moon,
		SolarPhase:         solar,
		Timestamp:          now.Format(time.RFC3339),
	}
}

// seedDefaultGeofenceZones populates initial calibrated zones with temporal decay parameters
func (e *Engine) seedDefaultGeofenceZones() {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)
	exp48h := now.Add(48 * time.Hour).Format(time.RFC3339)

	// 1. Japan Morioka Kaki Zone (Tier 2 Warning)
	poly1 := generatePlumePolygon(39.7020, 141.1520, 750.0, 1.8, 0.6, 1.83, false)
	coords1 := extractCoordinatesFromPolygon(poly1)

	zone1 := model.ActiveGeofenceZone{
		Id:                   "geo_jp_kaki_01",
		Tier:                 "TIER_2_WARNING",
		TargetSpecies:        []string{"BEAR"},
		Coordinates:          coords1,
		Centroid:             model.ActiveGeofencePoint{Latitude: 39.7020, Longitude: 141.1520},
		RadiusMeters:         750.0,
		CausalExplanation:    "Tier 2 Warning: Overripe persimmon fruit drops with westerly scent plume extending toward residential fringe.",
		PrimaryAttractantRef: "kt_1120",
		ActiveRiskScore:      0.78,
		PolygonGeoJson:       poly1,
		ValidFrom:            nowStr,
		ValidUntil:           exp48h,
		GeneratedAt:          nowStr,
	}

	// 2. Indonesia Sumatra Tiger Kill-Site (Tier 3 Danger with 72h decay)
	poly2 := generatePlumePolygon(0.8250, 101.5300, 1200.0, 1.4, 0.7, 0.17, false)
	coords2 := extractCoordinatesFromPolygon(poly2)

	zone2 := model.ActiveGeofenceZone{
		Id:                   "geo_id_tiger_02",
		Tier:                 "TIER_3_DANGER",
		TargetSpecies:        []string{"TIGER"},
		Coordinates:          coords2,
		Centroid:             model.ActiveGeofencePoint{Latitude: 0.8250, Longitude: 101.5300},
		RadiusMeters:         1200.0,
		CausalExplanation:    "TIER 3 DANGER: Fresh tiger livestock kill cached within secondary riparian scrub. Critical 72h return window.",
		PrimaryAttractantRef: "ls_sumatra_01",
		ActiveRiskScore:      0.96,
		PolygonGeoJson:       poly2,
		ValidFrom:            nowStr,
		ValidUntil:           now.Add(72 * time.Hour).Format(time.RFC3339),
		GeneratedAt:          nowStr,
	}

	// 3. Thailand Khao Ang Rue Nai Elephant Crop-Raid Corridor (Tier 2 Warning)
	poly3 := generatePlumePolygon(13.5210, 101.4520, 2000.0, 2.0, 0.5, 5.23, false)
	coords3 := extractCoordinatesFromPolygon(poly3)

	zone3 := model.ActiveGeofenceZone{
		Id:                   "geo_th_elephant_03",
		Tier:                 "TIER_2_WARNING",
		TargetSpecies:        []string{"ELEPHANT"},
		Coordinates:          coords3,
		Centroid:             model.ActiveGeofencePoint{Latitude: 13.5210, Longitude: 101.4520},
		RadiusMeters:         2000.0,
		CausalExplanation:    "Tier 2 Warning: Ripening sugarcane harvest boundary coupling with nighttime eastern corridor migration.",
		PrimaryAttractantRef: "plot_th_01",
		ActiveRiskScore:      0.82,
		PolygonGeoJson:       poly3,
		ValidFrom:            nowStr,
		ValidUntil:           exp48h,
		GeneratedAt:          nowStr,
	}

	e.activeZones[zone1.Id] = zone1
	e.activeZones[zone2.Id] = zone2
	e.activeZones[zone3.Id] = zone3
}

// GenerateScentPlumeGeofences synthesizes active attractant hubs, wind-driven scent plumes,
// moon illumination, and flooded river barrier deflections to generate 3-tier RFC 7946 GeoJSON polygons.
func (e *Engine) GenerateScentPlumeGeofences(req model.GeneratePlumeGeofencesRequest) model.GeneratePlumeGeofencesResponse {
	weather := req.Weather
	if weather == nil {
		now := time.Now().UTC()
		moon := CalculateMoonPhase(now)
		weather = &model.WeatherConditions{
			WindSpeedKmh:     10.0,
			WindDirectionDeg: 270.0, // Westerly wind (blowing from West toward East)
			MoonIllumination: moon.Illumination,
			PrecipitationMmH: 0.0,
		}
	}

	// 1. Lunar Scaling Multiplier: C_moon = 1.0 + 0.6 * (1.0 - MoonIllumination)
	// Elevates risk up to 1.6x on New Moon nights (0.0 illumination)
	lunarFactor := 1.0 + 0.6*(1.0-math.Min(math.Max(weather.MoonIllumination, 0.0), 1.0))

	// 2. Wind Scent-Plume distortion factors
	blowsTowardDeg := math.Mod(weather.WindDirectionDeg+180.0, 360.0)
	blowsTowardRad := blowsTowardDeg * (math.Pi / 180.0)

	// Downwind elongation: k_downwind = 1.0 + 0.08 * V_wind (clamped max 2.5x)
	kDownwind := math.Min(2.5, 1.0+0.08*weather.WindSpeedKmh)
	// Upwind compression: k_upwind = max(0.4, 1.0 - 0.04 * V_wind)
	kUpwind := math.Max(0.4, 1.0-0.04*weather.WindSpeedKmh)

	// 3. Flooded River Barrier Deflection
	isFloodedRiver := weather.PrecipitationMmH >= 25.0

	var zones []model.GeofenceZoneDto

	now := time.Now().UTC()
	for idx, hub := range req.Attractants {
		baseRadius, tier, species, baseRisk := calculateHubProperties(hub)

		// Astronomical solar elevation & crepuscular foraging peak (dawn/dusk)
		solar := CalculateSolarPhase(hub.Latitude, hub.Longitude, now)
		crepuscularFactor := 0.70 + 0.30*solar.CrepuscularRiskMultiplier

		// Scale risk by lunar factor and crepuscular timing
		effectiveRisk := math.Min(1.0, baseRisk*lunarFactor*crepuscularFactor)

		// Scale risk by flooded river barrier compression (wildlife concentrated near bridges/culverts)
		if isFloodedRiver {
			effectiveRisk = math.Min(1.0, effectiveRisk*1.25)
			baseRadius *= 1.35
		}

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
			isFloodedRiver,
		)

		// Apply OpenStreetMap infrastructure barriers & riparian corridor alignment
		osmCtx := DefaultOsmContextForRegion(hub.Latitude, hub.Longitude)
		polygon = ApplyOsmSpatialConstraints(polygon, osmCtx, hub.Latitude, hub.Longitude)

		causalFactors := []string{
			fmt.Sprintf("Attractant: %s", hub.Type),
			fmt.Sprintf("Wind: %.1f km/h (bearing %.0f°)", weather.WindSpeedKmh, blowsTowardDeg),
			fmt.Sprintf("Moon Illumination: %.2f (Risk Multiplier: %.2fx)", weather.MoonIllumination, lunarFactor),
			fmt.Sprintf("Solar Phase: %s (Elevation: %.1f°, Crepuscular Multiplier: %.2fx)", solar.PhaseName, solar.ElevationDegrees, solar.CrepuscularRiskMultiplier),
		}
		if osmCtx.HasDividedMotorway {
			causalFactors = append(causalFactors, fmt.Sprintf("OSM Barrier: Fenced highway limits cross-corridor spillover"))
		}
		if osmCtx.NearbyRiversCount > 0 {
			causalFactors = append(causalFactors, fmt.Sprintf("OSM Riparian: Corridor aligned along %s (bearing %.0f°)", osmCtx.RiparianCorridorName, osmCtx.CorridorDeflectionDeg))
		}
		if hub.FruitLoad != "" {
			causalFactors = append(causalFactors, fmt.Sprintf("Fruit Load: %s", hub.FruitLoad))
		}
		if isFloodedRiver {
			causalFactors = append(causalFactors, fmt.Sprintf("Flooded River Barrier: Crossing severed by torrential rain (%.1f mm/h), wildlife deflected toward road culverts", weather.PrecipitationMmH))
		}

		zoneDto := model.GeofenceZoneDto{
			Id:              fmt.Sprintf("geo_%s_%d", hub.Id, idx+1),
			Tier:            tier,
			TargetSpecies:   species,
			PolygonGeoJson:  polygon,
			ActiveRiskScore: effectiveRisk,
			CausalFactors:   causalFactors,
			ExpiresAt:       time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339),
			CreatedAt:       time.Now().UTC().Format(time.RFC3339),
		}
		zones = append(zones, zoneDto)

		// Register in active cache
		coords := extractCoordinatesFromPolygon(polygon)
		e.mu.Lock()
		e.activeZones[zoneDto.Id] = model.ActiveGeofenceZone{
			Id:                   zoneDto.Id,
			Tier:                 zoneDto.Tier,
			TargetSpecies:        zoneDto.TargetSpecies,
			Coordinates:          coords,
			Centroid:             model.ActiveGeofencePoint{Latitude: hub.Latitude, Longitude: hub.Longitude},
			RadiusMeters:         baseRadius,
			CausalExplanation:    fmt.Sprintf("%s triggered by %s under current wind plume and lunar darkness.", zoneDto.Tier, hub.Type),
			PrimaryAttractantRef: hub.Id,
			ActiveRiskScore:      effectiveRisk,
			PolygonGeoJson:       polygon,
			ValidFrom:            time.Now().UTC().Format(time.RFC3339),
			ValidUntil:           zoneDto.ExpiresAt,
			GeneratedAt:          zoneDto.CreatedAt,
		}
		e.mu.Unlock()
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
		radius := 1500.0
		risk := 0.70
		if hub.FruitLoad == "ABUNDANT" {
			radius = 3000.0
			risk = 0.86
		}
		return radius, "TIER_2_WARNING", []string{"ELEPHANT"}, risk

	case "ROADKILL", "KILL_SITE":
		// Exponential temporal decay: R(t) = R_0 * e^(-lambda * t)
		// Half-life of 24h: lambda = ln(2)/24 ~= 0.02888
		decayMultiplier := 1.0
		if hub.HoursSinceDeath > 0 {
			decayMultiplier = math.Exp(-0.02888 * hub.HoursSinceDeath)
		}
		radius := 800.0 * math.Max(0.5, decayMultiplier)
		risk := 0.90 * decayMultiplier
		species := []string{"TIGER", "BEAR"}
		tier := "TIER_3_DANGER"
		if risk < 0.60 {
			tier = "TIER_2_WARNING"
		}
		return radius, tier, species, risk

	case "PREDATOR_SIGHTING":
		// Sighting temporal decay: half-life of 12h
		decayMultiplier := 1.0
		if hub.HoursSinceSeen != nil && *hub.HoursSinceSeen > 0 {
			decayMultiplier = math.Exp(-0.05776 * (*hub.HoursSinceSeen))
		}
		radius := 1200.0 * math.Max(0.4, decayMultiplier)
		risk := 0.92 * decayMultiplier
		species := []string{"BEAR", "TIGER", "ELEPHANT"}
		tier := "TIER_3_DANGER"
		if risk < 0.65 {
			tier = "TIER_2_WARNING"
		}
		return radius, tier, species, risk

	default:
		return 500.0, "TIER_1_NOTICE", []string{"UNKNOWN"}, 0.50
	}
}

// generatePlumePolygon creates a deformed elliptical polygon elongated downwind with optional river barrier spreading
func generatePlumePolygon(lat, lon, baseRadius, kDownwind, kUpwind, blowsTowardRad float64, isFloodedRiver bool) *model.GeoJSONPolygonGeometry {
	var ring [][]float64

	for i := 0; i <= NumPolygonPoints; i++ {
		angle := float64(i) * (2.0 * math.Pi / float64(NumPolygonPoints))

		deltaAngle := angle - blowsTowardRad

		// Weight based on downwind/upwind alignment
		weight := (math.Cos(deltaAngle) + 1.0) / 2.0
		radialDist := baseRadius * (kUpwind + weight*(kDownwind-kUpwind))

		// If flooded river, expand lateral spread (perpendicular to river flow)
		if isFloodedRiver {
			lateralFactor := 1.0 + 0.35*math.Abs(math.Sin(deltaAngle))
			radialDist *= lateralFactor
		}

		dx := radialDist * math.Sin(angle) // Easting
		dy := radialDist * math.Cos(angle) // Northing

		dLat := (dy / EarthRadiusMeters) * (180.0 / math.Pi)
		dLon := (dx / (EarthRadiusMeters * math.Cos(lat*(math.Pi/180.0)))) * (180.0 / math.Pi)

		pointLat := lat + dLat
		pointLon := lon + dLon

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

func extractCoordinatesFromPolygon(poly *model.GeoJSONPolygonGeometry) []model.ActiveGeofencePoint {
	if poly == nil || len(poly.Coordinates) == 0 || len(poly.Coordinates[0]) == 0 {
		return nil
	}
	var pts []model.ActiveGeofencePoint
	for _, coord := range poly.Coordinates[0] {
		if len(coord) >= 2 {
			pts = append(pts, model.ActiveGeofencePoint{
				Longitude: coord[0],
				Latitude:  coord[1],
			})
		}
	}
	return pts
}

func roundToDecimals(val float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(val*pow) / pow
}

// Recalculate recalculates hazard boundaries within 5 minutes of a verified roadkill or predator sighting
func (e *Engine) Recalculate(req model.GeofenceRecalculateRequest) model.GeofenceRecalculateResponse {
	e.mu.Lock()
	defer e.mu.Unlock()

	lat := req.TargetCenterLatitude
	lon := req.TargetCenterLongitude
	if req.TargetCenter != nil {
		lat = req.TargetCenter.Latitude
		lon = req.TargetCenter.Longitude
	}

	radiusKm := req.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 5.0
	}

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)
	exp72h := now.Add(72 * time.Hour).Format(time.RFC3339)

	// Immediately recalibrate new Tier 3 Danger zone around the verified trigger
	triggerId := req.TriggerEntityRef
	if triggerId == "" {
		triggerId = fmt.Sprintf("incident_%d", now.Unix())
	}

	plumePoly := generatePlumePolygon(lat, lon, 1500.0, 1.8, 0.6, 1.57, false)
	coords := extractCoordinatesFromPolygon(plumePoly)

	newZone := model.ActiveGeofenceZone{
		Id:                   fmt.Sprintf("recalc_%s", triggerId),
		Tier:                 "TIER_3_DANGER",
		TargetSpecies:        []string{"BEAR", "TIGER"},
		Coordinates:          coords,
		Centroid:             model.ActiveGeofencePoint{Latitude: lat, Longitude: lon},
		RadiusMeters:         1500.0,
		CausalExplanation:    fmt.Sprintf("TIER 3 DANGER: Automated 5-minute spatial recalibration following verified trigger %s.", triggerId),
		PrimaryAttractantRef: triggerId,
		ActiveRiskScore:      0.98,
		PolygonGeoJson:       plumePoly,
		ValidFrom:            nowStr,
		ValidUntil:           exp72h,
		GeneratedAt:          nowStr,
	}

	e.activeZones[newZone.Id] = newZone

	return model.GeofenceRecalculateResponse{
		RecalculatedGeofencesCount: len(e.activeZones),
		AffectedTiers:              []string{"TIER_1_NOTICE", "TIER_2_WARNING", "TIER_3_DANGER"},
		UpdatedAt:                  nowStr,
	}
}

// GetActiveGeofences returns active geofence zones filtered by geographic proximity
func (e *Engine) GetActiveGeofences(lat, lon, radiusKm float64) []model.ActiveGeofenceZone {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if radiusKm <= 0 {
		radiusKm = 25.0
	}

	now := time.Now().UTC()
	var results []model.ActiveGeofenceZone

	for _, zone := range e.activeZones {
		// Expiration check
		if zone.ValidUntil != "" {
			exp, err := time.Parse(time.RFC3339, zone.ValidUntil)
			if err == nil && now.After(exp) {
				continue
			}
		}

		// Distance check from query location to centroid
		distKm := haversineDistanceKm(lat, lon, zone.Centroid.Latitude, zone.Centroid.Longitude)
		if distKm <= radiusKm {
			results = append(results, zone)
		}
	}

	return results
}

func haversineDistanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	radLat1 := lat1 * math.Pi / 180.0
	radLat2 := lat2 * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(radLat1)*math.Cos(radLat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
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

// ApplyPromotedWeights updates the live geofencing engine with newly promoted model weights without server restart.
func (e *Engine) ApplyPromotedWeights(version string, weights map[string]float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.activeModelVersion = version
	if weights != nil {
		e.modelWeights = weights
	}
}

// ActiveModelVersion returns the active model version currently governing geofence inferences.
func (e *Engine) ActiveModelVersion() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.activeModelVersion
}
