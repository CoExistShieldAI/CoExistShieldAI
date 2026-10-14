package elephant

import (
	"fmt"
	"math"
	"sync"
	"time"

	"ai.coexistshield/backend/internal/model"
)

// Crop Aromatic Volatility Constants (Volatile Organic Compound emission potential)
const (
	AromaJackfruit = 1.00 // Intensely aromatic volatile esters (detectable by elephants >3km downwind)
	AromaSugarcane = 0.85 // High-sucrose volatile spikes approaching harvest season
	AromaBanana    = 0.75 // Dense calorie attractant, continuous ripening clusters
	AromaMango     = 0.70 // Seasonal aromatic terpenes and sugars
	AromaDefault   = 0.50
)

// ElephantCropPlot represents an agricultural plot monitored for Human-Elephant Conflict (HEC).
type ElephantCropPlot struct {
	ID                    string                         `json:"id"`
	Name                  string                         `json:"name"`
	Province              string                         `json:"province"`
	ParkBoundary          string                         `json:"parkBoundary"`
	CropType              string                         `json:"cropType"`              // SUGARCANE, BANANA, JACKFRUIT, MANGO
	GrowthStage           string                         `json:"growthStage"`           // VEGETATIVE, RIPENING, READY_FOR_HARVEST, HARVESTED_CLEARED
	Latitude              float64                        `json:"latitude"`
	Longitude             float64                        `json:"longitude"`
	AreaHectares          float64                        `json:"areaHectares"`
	DistanceToForestEdgeM float64                        `json:"distanceToForestEdgeM"`
	ForestBearingDeg      float64                        `json:"forestBearingDeg"`      // Bearing from plot toward protected national park forest edge
	AromaVolatilityIndex  float64                        `json:"aromaVolatilityIndex"`  // [0.0 - 1.0]
	MaturityScore         float64                        `json:"maturityScore"`         // [0.05 - 1.0]
	ActiveRiskScore       float64                        `json:"activeRiskScore"`       // [0.0 - 1.0]
	AlertTier             string                         `json:"alertTier"`             // SAFE, TIER_1_NOTICE, TIER_2_WARNING, TIER_3_DANGER
	ScentBufferMeters     float64                        `json:"scentBufferMeters"`
	ScentPlumePolygon     *model.GeoJSONPolygonGeometry  `json:"scentPlumePolygon,omitempty"`
	UpdatedAt             string                         `json:"updatedAt"`
}

// Engine calculates dynamic Asian elephant crop-raiding probabilities and scent plume buffers.
type Engine struct {
	mu    sync.RWMutex
	plots map[string]ElephantCropPlot
}

// NewEngine initializes the Thailand Elephant HEC Engine with registered plantations near Kui Buri & Kaeng Krachan.
func NewEngine() *Engine {
	e := &Engine{
		plots: make(map[string]ElephantCropPlot),
	}
	e.seedInitialPlots()
	return e
}

func (e *Engine) seedInitialPlots() {
	now := time.Now().UTC().Format(time.RFC3339)

	initial := []ElephantCropPlot{
		{
			ID:                    "th_plot_kb_01",
			Name:                  "Kui Buri Valley Sugarcane Estate #1",
			Province:              "Prachuap Khiri Khan",
			ParkBoundary:          "Kui Buri National Park",
			CropType:              "SUGARCANE",
			GrowthStage:           "READY_FOR_HARVEST",
			Latitude:              12.0520,
			Longitude:             99.6210,
			AreaHectares:          14.5,
			DistanceToForestEdgeM: 650.0,
			ForestBearingDeg:      285.0, // Forest is West-Northwest
			AromaVolatilityIndex:  AromaSugarcane,
			MaturityScore:         1.00,
			UpdatedAt:             now,
		},
		{
			ID:                    "th_plot_kb_02",
			Name:                  "Khao Daeng Banana & Jackfruit Grove",
			Province:              "Prachuap Khiri Khan",
			ParkBoundary:          "Kui Buri National Park",
			CropType:              "JACKFRUIT",
			GrowthStage:           "RIPENING",
			Latitude:              12.0810,
			Longitude:             99.6450,
			AreaHectares:          8.2,
			DistanceToForestEdgeM: 1100.0,
			ForestBearingDeg:      270.0, // Forest is West
			AromaVolatilityIndex:  AromaJackfruit,
			MaturityScore:         0.85,
			UpdatedAt:             now,
		},
		{
			ID:                    "th_plot_kk_01",
			Name:                  "Kaeng Krachan Border Sweet Cane Plantation",
			Province:              "Phetchaburi",
			ParkBoundary:          "Kaeng Krachan National Park",
			CropType:              "SUGARCANE",
			GrowthStage:           "RIPENING",
			Latitude:              12.8640,
			Longitude:             99.5820,
			AreaHectares:          22.0,
			DistanceToForestEdgeM: 820.0,
			ForestBearingDeg:      260.0, // Forest is West-Southwest
			AromaVolatilityIndex:  AromaSugarcane,
			MaturityScore:         0.85,
			UpdatedAt:             now,
		},
		{
			ID:                    "th_plot_kk_02",
			Name:                  "Pa Deng Village Commercial Banana Cooperative",
			Province:              "Phetchaburi",
			ParkBoundary:          "Kaeng Krachan National Park",
			CropType:              "BANANA",
			GrowthStage:           "READY_FOR_HARVEST",
			Latitude:              12.7230,
			Longitude:             99.5110,
			AreaHectares:          12.0,
			DistanceToForestEdgeM: 450.0,
			ForestBearingDeg:      290.0,
			AromaVolatilityIndex:  AromaBanana,
			MaturityScore:         1.00,
			UpdatedAt:             now,
		},
		{
			ID:                    "th_plot_kb_03",
			Name:                  "Ban Yang Chum Mango Orchard (Cleared)",
			Province:              "Prachuap Khiri Khan",
			ParkBoundary:          "Kui Buri National Park",
			CropType:              "MANGO",
			GrowthStage:           "HARVESTED_CLEARED",
			Latitude:              12.1150,
			Longitude:             99.6890,
			AreaHectares:          9.0,
			DistanceToForestEdgeM: 2800.0,
			ForestBearingDeg:      265.0,
			AromaVolatilityIndex:  AromaMango,
			MaturityScore:         0.05,
			UpdatedAt:             now,
		},
	}

	// Default calm westerly wind for initial assessment
	defaultWeather := model.WeatherConditions{
		WindSpeedKmh:     14.0,
		WindDirectionDeg: 90.0, // Wind blowing FROM East (90) TOWARDS West (270) -> straight into national park forest!
		MoonIllumination: 0.15, // Near New Moon (high nocturnal raiding risk)
		PrecipitationMmH: 0.0,
	}

	for _, plot := range initial {
		computed := e.calculatePlotRisk(plot, defaultWeather)
		e.plots[computed.ID] = computed
	}
}

// GetRegisteredPlots returns all registered crop plots with their current risk ratings.
func (e *Engine) GetRegisteredPlots() []ElephantCropPlot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]ElephantCropPlot, 0, len(e.plots))
	for _, p := range e.plots {
		result = append(result, p)
	}
	return result
}

// GetPlot returns a single crop plot by ID.
func (e *Engine) GetPlot(id string) (*ElephantCropPlot, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	p, exists := e.plots[id]
	if !exists {
		return nil, fmt.Errorf("plot not found: %s", id)
	}
	return &p, nil
}

// UpdateHarvestStage updates the crop maturity stage (submitted by farmer or ranger) and recalculates risk.
func (e *Engine) UpdateHarvestStage(id string, stage string, weather *model.WeatherConditions) (*ElephantCropPlot, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	plot, exists := e.plots[id]
	if !exists {
		return nil, fmt.Errorf("plot not found: %s", id)
	}

	validStage := normalizeGrowthStage(stage)
	plot.GrowthStage = validStage
	plot.MaturityScore = getMaturityMultiplier(validStage)
	plot.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	w := model.WeatherConditions{
		WindSpeedKmh:     14.0,
		WindDirectionDeg: 90.0,
		MoonIllumination: 0.20,
	}
	if weather != nil {
		w = *weather
	}

	updated := e.calculatePlotRisk(plot, w)
	e.plots[id] = updated
	return &updated, nil
}

// RecalculateAllPlots updates risk scores and scent buffers across all plots given current meteorological streams.
func (e *Engine) RecalculateAllPlots(weather model.WeatherConditions) []ElephantCropPlot {
	e.mu.Lock()
	defer e.mu.Unlock()

	result := make([]ElephantCropPlot, 0, len(e.plots))
	for id, plot := range e.plots {
		updated := e.calculatePlotRisk(plot, weather)
		e.plots[id] = updated
		result = append(result, updated)
	}
	return result
}

func (e *Engine) calculatePlotRisk(plot ElephantCropPlot, weather model.WeatherConditions) ElephantCropPlot {
	// 1. Scent downwind flow direction: wind blows FROM windDirectionDeg TO (windDirectionDeg + 180)
	scentFlowDeg := math.Mod(weather.WindDirectionDeg+180.0, 360.0)

	// 2. Alignment delta between scent flow and vector towards national park forest edge
	deltaAngle := math.Abs(scentFlowDeg - plot.ForestBearingDeg)
	if deltaAngle > 180.0 {
		deltaAngle = 360.0 - deltaAngle
	}

	// Wind alignment factor [0.0 - 1.0]: 1.0 when scent blows directly into reserve
	rad := deltaAngle * (math.Pi / 180.0)
	windAlignFactor := math.Cos(rad / 2.0)
	windAlignFactor = windAlignFactor * windAlignFactor

	// 3. Distance decay toward forest edge (scale ~ 1500m)
	distFactor := math.Exp(-plot.DistanceToForestEdgeM / 1500.0)

	// 4. Lunar Nocturnal Darkness: Asian elephants raid primarily at night during dark moons
	moonDarkness := 1.0 - math.Max(0.0, math.Min(1.0, weather.MoonIllumination))

	// 5. Calculate Composite HEC Threat Score
	// Aroma + Maturity governs total odor emission; wind alignment transports that odor into the forest
	aromaStrength := plot.AromaVolatilityIndex * plot.MaturityScore
	windSpeedRatio := math.Min(2.0, math.Max(0.2, weather.WindSpeedKmh/12.0))

	score := 0.45*aromaStrength +
		0.25*(windAlignFactor*windSpeedRatio*aromaStrength) +
		0.20*(distFactor*(0.15+0.85*aromaStrength)) +
		0.10*(moonDarkness*aromaStrength)

	score = math.Max(0.05, math.Min(0.98, score))
	score = math.Round(score*1000.0) / 1000.0

	// 6. Assign Alert Tier
	var tier string
	var baseBufferMeters float64
	switch {
	case score >= 0.80:
		tier = "TIER_3_DANGER"
		baseBufferMeters = 2800.0
	case score >= 0.55:
		tier = "TIER_2_WARNING"
		baseBufferMeters = 1800.0
	case score >= 0.25:
		tier = "TIER_1_NOTICE"
		baseBufferMeters = 1000.0
	default:
		tier = "SAFE"
		baseBufferMeters = 400.0
	}

	// Expand buffer along downwind corridor toward forest
	effectiveBuffer := baseBufferMeters * (1.0 + 0.5*windAlignFactor)
	plot.ActiveRiskScore = score
	plot.AlertTier = tier
	plot.ScentBufferMeters = math.Round(effectiveBuffer)

	// 7. Generate Dynamic Scent-Plume Polygon Corridor towards forest
	plot.ScentPlumePolygon = generateCorridorPolygon(plot.Latitude, plot.Longitude, scentFlowDeg, effectiveBuffer)

	return plot
}

func generateCorridorPolygon(lat, lon, scentBearingDeg, bufferMeters float64) *model.GeoJSONPolygonGeometry {
	const numVertices = 24
	metersPerDegreeLat := 111320.0
	metersPerDegreeLon := 111320.0 * math.Cos(lat*math.Pi/180.0)

	scentRad := scentBearingDeg * math.Pi / 180.0

	// Elliptical deformation along scent flow bearing
	kDownwind := 1.6
	kCrosswind := 0.65

	ring := make([][]float64, numVertices+1)

	for i := 0; i < numVertices; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(numVertices)

		// Base local coordinates
		localX := bufferMeters * kCrosswind * math.Sin(angle)
		localY := bufferMeters * math.Cos(angle)

		// Elongate in forward scent direction
		if localY > 0 {
			localY *= kDownwind
		}

		// Rotate by scent flow azimuth
		rotX := localX*math.Cos(scentRad) + localY*math.Sin(scentRad)
		rotY := -localX*math.Sin(scentRad) + localY*math.Cos(scentRad)

		vertexLat := lat + (rotY / metersPerDegreeLat)
		vertexLon := lon + (rotX / metersPerDegreeLon)

		// RFC 7946: [lon, lat]
		ring[i] = []float64{
			math.Round(vertexLon*1000000.0) / 1000000.0,
			math.Round(vertexLat*1000000.0) / 1000000.0,
		}
	}
	// Close LinearRing
	ring[numVertices] = []float64{ring[0][0], ring[0][1]}

	return &model.GeoJSONPolygonGeometry{
		Type:        "Polygon",
		Coordinates: [][][]float64{ring},
	}
}

func normalizeGrowthStage(stage string) string {
	switch stage {
	case "READY_FOR_HARVEST", "ready_for_harvest", "ready":
		return "READY_FOR_HARVEST"
	case "RIPENING", "ripening", "ripe":
		return "RIPENING"
	case "VEGETATIVE", "vegetative", "growing":
		return "VEGETATIVE"
	case "HARVESTED_CLEARED", "harvested_cleared", "cleared", "harvested":
		return "HARVESTED_CLEARED"
	default:
		return "RIPENING"
	}
}

func getMaturityMultiplier(stage string) float64 {
	switch stage {
	case "READY_FOR_HARVEST":
		return 1.00
	case "RIPENING":
		return 0.85
	case "VEGETATIVE":
		return 0.20
	case "HARVESTED_CLEARED":
		return 0.05
	default:
		return 0.80
	}
}
