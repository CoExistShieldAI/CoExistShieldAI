package corroboration

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"ai.coexistshield/backend/internal/gemini"
	"ai.coexistshield/backend/internal/model"
)

type Engine struct {
	mu           sync.RWMutex
	geminiClient *gemini.Client
	streamItems  map[string]model.CausalStreamItem
}

func NewEngine(geminiClient *gemini.Client) *Engine {
	e := &Engine{
		geminiClient: geminiClient,
		streamItems:  make(map[string]model.CausalStreamItem),
	}
	e.seedDefaultCausalStreams()
	return e
}

// seedDefaultCausalStreams initializes realistic multi-source environmental feeds across Japan, Thailand, and Indonesia
func (e *Engine) seedDefaultCausalStreams() {
	now := time.Now().UTC()
	exp24h := now.Add(24 * time.Hour).Format(time.RFC3339)
	nowStr := now.Format(time.RFC3339)

	defaults := []model.CausalStreamItem{
		// 1. Agricultural Harvest Stream (Japan: Kaki Persimmons)
		{
			ItemId:        "stream-agri-jp-01",
			StreamType:    model.StreamAgriculturalHarvest,
			Region:        "JAPAN",
			Title:         "Tohoku Autumn Persimmon Peak Ripening Index",
			Description:   "Overripe unharvested kaki fruit dropping across abandoned hillside orchards in Morioka fringe.",
			Latitude:      39.7025,
			Longitude:     141.1528,
			RadiusMeters:  2000,
			Confidence:    0.88,
			TargetSpecies: "BEAR",
			Timestamp:     nowStr,
			ExpiresAt:     exp24h,
			Metadata: map[string]interface{}{
				"crop":           "Diospyros kaki",
				"ripenessStage":  "OVERRIPE_FALLING",
				"calorieDensity": "HIGH",
			},
		},
		// 2. Municipal Attractant Stream (Japan: Garbage Collection)
		{
			ItemId:        "stream-muni-jp-02",
			StreamType:    model.StreamMunicipalAttractant,
			Region:        "JAPAN",
			Title:         "Municipal Burnable Waste Collection Day (Morning Olfactory Peak)",
			Description:   "Matsugaeda residential district scheduled morning waste placement; curbside attractant exposure.",
			Latitude:      39.7040,
			Longitude:     141.1540,
			RadiusMeters:  1200,
			Confidence:    0.82,
			TargetSpecies: "BEAR",
			Timestamp:     nowStr,
			ExpiresAt:     exp24h,
			Metadata: map[string]interface{}{
				"category":    "BURNABLE_FOOD_WASTE",
				"timeWindow":  "06:00-09:00",
				"curbsideBags": 140,
			},
		},
		// 3. Satellite Thermal / Wildfire Stream (Thailand: Forest Hotspots)
		{
			ItemId:        "stream-firms-th-03",
			StreamType:    model.StreamSatelliteThermal,
			Region:        "THAILAND",
			Title:         "NASA FIRMS / VIIRS Active Thermal Fire Anomaly",
			Description:   "Moderate forest fire hotspot detected inside Khao Ang Rue Nai Sanctuary; compresses wildlife corridor outward.",
			Latitude:      13.5180,
			Longitude:     101.4490,
			RadiusMeters:  3500,
			Confidence:    0.86,
			TargetSpecies: "ELEPHANT",
			Timestamp:     nowStr,
			ExpiresAt:     exp24h,
			Metadata: map[string]interface{}{
				"sensor":      "VIIRS_NOAA20",
				"brightnessK": 342.5,
				"frpMegaWatt": 28.4,
			},
		},
		// 4. Agricultural Harvest Stream (Thailand: Sugarcane Harvest)
		{
			ItemId:        "stream-agri-th-04",
			StreamType:    model.StreamAgriculturalHarvest,
			Region:        "THAILAND",
			Title:         "Nong Khla Sugarcane Harvesting & Crushing Cycle",
			Description:   "High-brix sugarcane stalks cut and piled at plantation edge; massive olfactory lure for elephant herds.",
			Latitude:      13.5220,
			Longitude:     101.4550,
			RadiusMeters:  2500,
			Confidence:    0.91,
			TargetSpecies: "ELEPHANT",
			Timestamp:     nowStr,
			ExpiresAt:     exp24h,
			Metadata: map[string]interface{}{
				"crop":        "Saccharum officinarum",
				"harvestBrix": 18.5,
				"openPiles":   true,
			},
		},
		// 5. Traffic Congestion & Roadkill Stream (Indonesia: Siak Highway Corridor)
		{
			ItemId:        "stream-traffic-id-05",
			StreamType:    model.StreamTrafficCongestion,
			Region:        "INDONESIA",
			Title:         "Siak Highway Commercial Truck Flow & Animal Crossing Severance",
			Description:   "High nighttime palm oil tanker frequency along Jalan Lintas Siak creates barrier effect and roadside strike hazards.",
			Latitude:      0.8260,
			Longitude:     101.5315,
			RadiusMeters:  3000,
			Confidence:    0.79,
			TargetSpecies: "TIGER",
			Timestamp:     nowStr,
			ExpiresAt:     exp24h,
			Metadata: map[string]interface{}{
				"heavyVehiclePerHour": 64,
				"recentAnimalStrikes": 2,
			},
		},
	}

	for _, item := range defaults {
		e.streamItems[item.ItemId] = item
	}
}

// IngestStreamItem registers a new causal event from external crawler/sensor
func (e *Engine) IngestStreamItem(item model.CausalStreamItem) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if item.ItemId == "" {
		item.ItemId = fmt.Sprintf("stream-%d", time.Now().UnixNano())
	}
	if item.Timestamp == "" {
		item.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	if item.ExpiresAt == "" {
		item.ExpiresAt = time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	}
	e.streamItems[item.ItemId] = item
}

// GetActiveStreamItems returns all active unexpired stream items filtered by region and stream type
func (e *Engine) GetActiveStreamItems(region string, streamType model.CausalStreamType) []model.CausalStreamItem {
	e.mu.RLock()
	defer e.mu.RUnlock()

	now := time.Now().UTC()
	var result []model.CausalStreamItem

	for _, item := range e.streamItems {
		if item.ExpiresAt != "" {
			exp, err := time.Parse(time.RFC3339, item.ExpiresAt)
			if err == nil && now.After(exp) {
				continue
			}
		}

		if region != "" && item.Region != region {
			continue
		}
		if streamType != "" && item.StreamType != streamType {
			continue
		}

		result = append(result, item)
	}

	return result
}

// Corroborate evaluates multi-source sensor fusion across all active streams within a spatial radius.
// Computes Bayesian combined confidence: C_corroborated = 1.0 - \prod (1 - C_i)
func (e *Engine) Corroborate(ctx context.Context, req model.CorroborationRequest) (*model.CorroborationResponse, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	radiusKm := req.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 5.0
	}
	maxAgeHours := req.MaxAgeHours
	if maxAgeHours <= 0 {
		maxAgeHours = 24.0
	}

	now := time.Now().UTC()

	// Gather events within spatial and temporal proximity
	var nearbyEvents []model.CausalStreamItem
	for _, item := range e.streamItems {
		if req.Region != "" && item.Region != req.Region {
			continue
		}

		// Temporal check
		if item.Timestamp != "" {
			t, err := time.Parse(time.RFC3339, item.Timestamp)
			if err == nil && now.Sub(t).Hours() > maxAgeHours {
				continue
			}
		}

		distKm := haversineDistanceKm(req.TargetLatitude, req.TargetLongitude, item.Latitude, item.Longitude)
		if distKm <= radiusKm {
			nearbyEvents = append(nearbyEvents, item)
		}
	}

	totalActive := len(e.streamItems)

	// If no nearby events, return empty response
	if len(nearbyEvents) == 0 {
		return &model.CorroborationResponse{
			Clusters:           []model.CorroboratedCluster{},
			TotalActiveStreams: totalActive,
			EvaluatedAt:        now.Format(time.RFC3339),
			ModelUsed:          gemini.GeminiModelVersion,
		}, nil
	}

	// Group by independent stream sources
	streamMaxConfidence := make(map[model.CausalStreamType]float64)
	speciesCount := make(map[string]int)

	for _, ev := range nearbyEvents {
		if ev.Confidence > streamMaxConfidence[ev.StreamType] {
			streamMaxConfidence[ev.StreamType] = ev.Confidence
		}
		if ev.TargetSpecies != "" && ev.TargetSpecies != "ALL" {
			speciesCount[ev.TargetSpecies]++
		}
	}

	// Determine dominant species
	dominantSpecies := "BEAR"
	maxSpCount := 0
	for sp, count := range speciesCount {
		if count > maxSpCount {
			maxSpCount = count
			dominantSpecies = sp
		}
	}

	// Calculate Multi-Source Bayesian Independent Confidence Fusion:
	// C_corroborated = 1.0 - \prod (1 - C_i)
	unconfidenceProduct := 1.0
	var distinctSources []model.CausalStreamType
	for st, conf := range streamMaxConfidence {
		distinctSources = append(distinctSources, st)
		unconfidenceProduct *= (1.0 - conf)
	}

	combinedConf := 1.0 - unconfidenceProduct
	if combinedConf > 0.999 {
		combinedConf = 0.999
	}
	// Round to 3 decimal places
	combinedConf = math.Round(combinedConf*1000) / 1000

	// Determine dynamic elevated tier based on independent source corroboration
	independentCount := len(distinctSources)
	elevatedTier := "SAFE"
	actionRecommendation := "Standard ecological monitoring. No emergency response required."

	if independentCount >= 3 {
		elevatedTier = "TIER_3_DANGER"
		actionRecommendation = fmt.Sprintf("CRITICAL ACTION: Triangulated multi-source alert (%d sources). Immediate dispatch of ranger deterrent squad and automated acoustic barrier activation for %s defense.", independentCount, dominantSpecies)
	} else if independentCount == 2 {
		elevatedTier = "TIER_2_WARNING"
		actionRecommendation = fmt.Sprintf("ELEVATED ALERT: Cross-sensor corroboration verified across %s and %s. Broadcast warning to local farmers and secure residential attractants.", distinctSources[0], distinctSources[1])
	} else if independentCount == 1 {
		if combinedConf >= 0.85 {
			elevatedTier = "TIER_1_NOTICE"
			actionRecommendation = "Advisory notice: Single causal stream anomaly detected. Autonomous surveillance crawler frequency elevated."
		} else {
			elevatedTier = "SAFE"
			actionRecommendation = "Low-threat baseline signal. Routine monitoring continues."
		}
	}

	// Generate Explainable AI Justification using Gemini 3.8 Flash
	var aiJustification string
	if e.geminiClient != nil {
		explanationReq := model.CausalAlertExplanationRequest{
			Tier:          elevatedTier,
			TargetSpecies: dominantSpecies,
			Region:        req.Region,
			Weather: map[string]interface{}{
				"corroboratedSourcesCount": independentCount,
				"combinedConfidence":       combinedConf,
			},
			NearbyAttractants: []map[string]interface{}{
				{"independentSources": distinctSources, "nearbyEventsCount": len(nearbyEvents)},
			},
		}
		justification, err := e.geminiClient.GenerateExplainableAlert(ctx, explanationReq, nil)
		if err == nil && justification != "" {
			aiJustification = justification
		}
	}

	if aiJustification == "" {
		aiJustification = fmt.Sprintf("%s: Corroborated %d independent environmental streams (combined confidence: %.1f%%) indicating imminent %s intrusion risk.", elevatedTier, independentCount, combinedConf*100, dominantSpecies)
	}

	cluster := model.CorroboratedCluster{
		ClusterId:              fmt.Sprintf("cluster-%s-%d", dominantSpecies, now.Unix()%100000),
		CenterLatitude:         req.TargetLatitude,
		CenterLongitude:        req.TargetLongitude,
		RadiusKm:               radiusKm,
		TargetSpecies:          dominantSpecies,
		CorroboratedSources:    distinctSources,
		IndependentSourceCount: independentCount,
		CombinedConfidence:     combinedConf,
		ElevatedTier:           elevatedTier,
		AiJustification:        aiJustification,
		ContributingEvents:     nearbyEvents,
		ActionRecommendation:   actionRecommendation,
		EvaluatedAt:            now.Format(time.RFC3339),
	}

	return &model.CorroborationResponse{
		Clusters:           []model.CorroboratedCluster{cluster},
		TotalActiveStreams: totalActive,
		EvaluatedAt:        now.Format(time.RFC3339),
		ModelUsed:          gemini.GeminiModelVersion,
	}, nil
}

// RunIngestionCycle simulates a scheduled autonomous polling cycle across all 5 causal streams
func (e *Engine) RunIngestionCycle() int {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)
	exp24h := now.Add(24 * time.Hour).Format(time.RFC3339)

	// Ingest updated dynamic readings
	cycleItems := []model.CausalStreamItem{
		{
			ItemId:        fmt.Sprintf("cycle-firms-%d", now.Unix()),
			StreamType:    model.StreamSatelliteThermal,
			Region:        "JAPAN",
			Title:         "Controlled Agricultural Slash Burn Anomaly",
			Description:   "Hillside agricultural burn detected near forest boundary, driving wildlife downhill.",
			Latitude:      39.7080,
			Longitude:     141.1560,
			RadiusMeters:  1800,
			Confidence:    0.84,
			TargetSpecies: "BEAR",
			Timestamp:     nowStr,
			ExpiresAt:     exp24h,
		},
		{
			ItemId:        fmt.Sprintf("cycle-nlp-%d", now.Unix()),
			StreamType:    model.StreamSocialNLP,
			Region:        "JAPAN",
			Title:         "Municipal Radio Broadcast NLP Incident",
			Description:   "Automated Gemini 3.8 Flash extraction from Matsugaeda municipal audio broadcast.",
			Latitude:      39.7030,
			Longitude:     141.1535,
			RadiusMeters:  1000,
			Confidence:    0.95,
			TargetSpecies: "BEAR",
			Timestamp:     nowStr,
			ExpiresAt:     exp24h,
			Metadata: map[string]interface{}{
				"model": GeminiModelVersion(),
			},
		},
	}

	for _, item := range cycleItems {
		e.IngestStreamItem(item)
	}

	return len(cycleItems)
}

func GeminiModelVersion() string {
	return gemini.GeminiModelVersion
}

// haversineDistanceKm calculates great-circle distance between two GPS coordinates
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
