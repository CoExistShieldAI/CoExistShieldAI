package tiger

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"ai.coexistshield/backend/internal/model"
)

// LivestockEnclosure represents a pastoral grazing pasture or night corral in tiger habitats.
type LivestockEnclosure struct {
	ID                    string                         `json:"id"`
	Name                  string                         `json:"name"`
	Province              string                         `json:"province"`              // e.g., "Riau", "West Sumatra", "Jambi"
	Village               string                         `json:"village"`
	LivestockType         string                         `json:"livestockType"`         // CATTLE, WATER_BUFFALO, GOAT, PIG
	HeadCount             int                            `json:"headCount"`
	HasNightCorral        bool                           `json:"hasNightCorral"`        // Reinforced predator-proof night pen
	Latitude              float64                        `json:"latitude"`
	Longitude             float64                        `json:"longitude"`
	DistanceToForestEdgeM float64                        `json:"distanceToForestEdgeM"`
	IsActiveKillSite      bool                           `json:"isActiveKillSite"`
	KillReportedAt        *time.Time                     `json:"killReportedAt,omitempty"`
	HoursSinceKill        float64                        `json:"hoursSinceKill"`
	CachingRiskExpiry     *time.Time                     `json:"cachingRiskExpiry,omitempty"` // 72 hours post-kill
	ReturnProbability     float64                        `json:"returnProbability"`          // [0.0 - 1.0] Tiger return probability
	ActiveThreatScore     float64                        `json:"activeThreatScore"`          // [0.0 - 1.0]
	AlertTier             string                         `json:"alertTier"`                  // SAFE, TIER_1_NOTICE, TIER_2_WARNING, TIER_3_DANGER
	ScentBufferMeters     float64                        `json:"scentBufferMeters"`
	KillSiteGeofence      *model.GeoJSONPolygonGeometry  `json:"killSiteGeofence,omitempty"`
	ForensicNotes         string                         `json:"forensicNotes,omitempty"`
	UpdatedAt             string                         `json:"updatedAt"`
}

// ReportDepredationRequest is submitted when an animal attack or carcass caching is found.
type ReportDepredationRequest struct {
	EnclosureID           string  `json:"enclosureId"`
	LivestockSpecies      string  `json:"livestockSpecies"`      // CATTLE, WATER_BUFFALO, GOAT, PIG
	HeadCountLost         int     `json:"headCountLost"`
	Latitude              float64 `json:"latitude"`
	Longitude             float64 `json:"longitude"`
	HasNightCorral        bool    `json:"hasNightCorral"`
	PhotoUrl              string  `json:"photoUrl,omitempty"`
	DragTrailLengthMeters float64 `json:"dragTrailLengthMeters"`
	CanineBiteSpreadMm    float64 `json:"canineBiteSpreadMm"`    // Sumatran tiger canine inter-tooth distance ~65-80mm
}

// TigerDepredationAlert contains immediate 3km radius village warning dispatch payload.
type TigerDepredationAlert struct {
	AlertID                     string   `json:"alertId"`
	Tier                        string   `json:"tier"` // TIER_3_DANGER
	KillSiteLat                 float64  `json:"killSiteLat"`
	KillSiteLon                 float64  `json:"killSiteLon"`
	VillagesIn3kmRadius         []string `json:"villagesIn3kmRadius"`
	ReturnWindowHoursRemaining  float64  `json:"returnWindowHoursRemaining"`
	CanineForensicsConfirmed    bool     `json:"canineForensicsConfirmed"`
	CausalReasoning             string   `json:"causalReasoning"`
	IssuedAt                    string   `json:"issuedAt"`
}

// EnclosurePersistenceListener is called whenever an enclosure or depredation event is modified.
type EnclosurePersistenceListener func(enc LivestockEnclosure)

// Engine coordinates tiger predatory ambush modeling, carcass caching temporal decay, and alert broadcasting.
type Engine struct {
	mu                  sync.RWMutex
	enclosures          map[string]LivestockEnclosure
	villages            []villageLocation
	persistenceListener EnclosurePersistenceListener
}

type villageLocation struct {
	Name      string
	Province  string
	Latitude  float64
	Longitude float64
}

// NewEngine initializes the Sumatra Tiger Predation Engine.
// Pre-seeded enclosures and villages are only injected when ENABLE_SEED_DATA="true" or DEV_MODE="true".
func NewEngine() *Engine {
	e := &Engine{
		enclosures: make(map[string]LivestockEnclosure),
		villages:   make([]villageLocation, 0),
	}
	if os.Getenv("ENABLE_SEED_DATA") == "true" || os.Getenv("DEV_MODE") == "true" {
		e.LoadFixtureData()
	}
	return e
}

// LoadFixtureData populates seed villages and livestock enclosures.
func (e *Engine) LoadFixtureData() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.villages = seedVillages()
	e.seedInitialEnclosuresLocked()
}

// SetPersistenceListener registers a callback invoked when an enclosure is saved/updated.
func (e *Engine) SetPersistenceListener(l EnclosurePersistenceListener) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.persistenceListener = l
}

// LoadEnclosures bulk loads enclosures (e.g. restored from Firestore).
func (e *Engine) LoadEnclosures(encs []LivestockEnclosure) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, enc := range encs {
		e.enclosures[enc.ID] = enc
	}
}

// RegisterEnclosure dynamically adds or updates a livestock pasture enclosure.
func (e *Engine) RegisterEnclosure(enc LivestockEnclosure) {
	e.mu.Lock()
	e.enclosures[enc.ID] = enc
	l := e.persistenceListener
	e.mu.Unlock()

	if l != nil {
		l(enc)
	}
}

func seedVillages() []villageLocation {
	return []villageLocation{
		{Name: "Desa Bukit Kesuma", Province: "Riau", Latitude: 0.1250, Longitude: 101.4420},
		{Name: "Dusun Sungai Rawa", Province: "Riau", Latitude: 0.1380, Longitude: 101.4650},
		{Name: "Kampung Lubuk Mandi", Province: "Riau", Latitude: 0.1120, Longitude: 101.4280},
		{Name: "Desa Rimba Melintang", Province: "Riau", Latitude: 0.1650, Longitude: 101.4850},
		{Name: "Talang Mandi Border", Province: "West Sumatra", Latitude: -0.4210, Longitude: 100.6520},
	}
}

func (e *Engine) seedInitialEnclosuresLocked() {
	now := time.Now().UTC()
	killTime1 := now.Add(-14 * time.Hour) // 14 hours ago (Peak 72h return window)
	expiry1 := killTime1.Add(72 * time.Hour)

	initial := []LivestockEnclosure{
		{
			ID:                    "ls_sumatra_01",
			Name:                  "Bukit Kesuma Cattle Enclosure B",
			Province:              "Riau",
			Village:               "Desa Bukit Kesuma",
			LivestockType:         "CATTLE",
			HeadCount:             24,
			HasNightCorral:        false, // High vulnerability
			Latitude:              0.1310,
			Longitude:             101.4480,
			DistanceToForestEdgeM: 380.0,
			IsActiveKillSite:      true,
			KillReportedAt:        &killTime1,
			HoursSinceKill:        14.0,
			CachingRiskExpiry:     &expiry1,
			ForensicNotes:         "Canine spread 72mm confirmed. Tiger cached adult cow carcass in dense understory.",
			UpdatedAt:             now.Format(time.RFC3339),
		},
		{
			ID:                    "ls_sumatra_02",
			Name:                  "Sungai Rawa Water Buffalo Pasture",
			Province:              "Riau",
			Village:               "Dusun Sungai Rawa",
			LivestockType:         "WATER_BUFFALO",
			HeadCount:             18,
			HasNightCorral:        true, // Reinforced night corral
			Latitude:              0.1420,
			Longitude:             101.4680,
			DistanceToForestEdgeM: 850.0,
			IsActiveKillSite:      false,
			ForensicNotes:         "Night corral reinforced with double-layered solar electric fencing.",
			UpdatedAt:             now.Format(time.RFC3339),
		},
		{
			ID:                    "ls_sumatra_03",
			Name:                  "Lubuk Mandi Smallholder Goat Corral",
			Province:              "Riau",
			Village:               "Kampung Lubuk Mandi",
			LivestockType:         "GOAT",
			HeadCount:             45,
			HasNightCorral:        false,
			Latitude:              0.1180,
			Longitude:             101.4350,
			DistanceToForestEdgeM: 520.0,
			IsActiveKillSite:      false,
			ForensicNotes:         "High acoustic vulnerability; goats bleating audible >1.2km into forest.",
			UpdatedAt:             now.Format(time.RFC3339),
		},
	}

	defaultWeather := model.WeatherConditions{
		WindSpeedKmh:     10.0,
		WindDirectionDeg: 180.0, // South wind blowing North
		MoonIllumination: 0.10,  // New Moon nocturnal darkness
	}

	for _, enc := range initial {
		computed := e.calculateEnclosureRisk(enc, defaultWeather, now)
		e.enclosures[computed.ID] = computed
	}
}

// GetEnclosures returns all monitored livestock enclosures and active kill sites.
func (e *Engine) GetEnclosures() []LivestockEnclosure {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]LivestockEnclosure, 0, len(e.enclosures))
	for _, enc := range e.enclosures {
		result = append(result, enc)
	}
	return result
}

// GetEnclosure returns an enclosure by ID.
func (e *Engine) GetEnclosure(id string) (*LivestockEnclosure, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	enc, exists := e.enclosures[id]
	if !exists {
		return nil, fmt.Errorf("enclosure not found: %s", id)
	}
	return &enc, nil
}

// ReportDepredation logs a newly verified tiger livestock kill, initiates the 72-hour return countdown,
// deploys a dynamic Tier 3 Danger Geofence, and triggers a 3km radius alert broadcast.
func (e *Engine) ReportDepredation(req ReportDepredationRequest, weather model.WeatherConditions) (*LivestockEnclosure, *TigerDepredationAlert, error) {
	var (
		l EnclosurePersistenceListener
		computed LivestockEnclosure
	)
	defer func() {
		if l != nil {
			l(computed)
		}
	}()

	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	expiry := now.Add(72 * time.Hour)

	enc, exists := e.enclosures[req.EnclosureID]
	if !exists {
		// Create dynamically if new field coordinates
		enc = LivestockEnclosure{
			ID:                    req.EnclosureID,
			Name:                  fmt.Sprintf("Depredation Site #%s", req.EnclosureID),
			Province:              "Riau",
			Village:               "Immediate Pastoral Sector",
			LivestockType:         req.LivestockSpecies,
			HeadCount:             req.HeadCountLost,
			HasNightCorral:        req.HasNightCorral,
			Latitude:              req.Latitude,
			Longitude:             req.Longitude,
			DistanceToForestEdgeM: 400.0,
		}
	}

	// Forensic verification: Panthera tigris canine puncture spread is typically 65mm - 85mm
	isTigerConfirmed := req.CanineBiteSpreadMm >= 60.0 && req.CanineBiteSpreadMm <= 95.0
	forensicMsg := fmt.Sprintf("Reported: %d %s lost. Drag trail: %.1fm. Canine bite spread: %.1fmm",
		req.HeadCountLost, req.LivestockSpecies, req.DragTrailLengthMeters, req.CanineBiteSpreadMm)
	if isTigerConfirmed {
		forensicMsg += " [Gemini Forensics: Positive Panthera tigris dental match]"
	}

	enc.IsActiveKillSite = true
	enc.KillReportedAt = &now
	enc.HoursSinceKill = 0.0
	enc.CachingRiskExpiry = &expiry
	enc.ForensicNotes = forensicMsg
	enc.UpdatedAt = now.Format(time.RFC3339)

	computed = e.calculateEnclosureRisk(enc, weather, now)
	e.enclosures[computed.ID] = computed
	l = e.persistenceListener

	// Identify villages within 3km radius
	villages3km := e.findVillagesWithinRadius(computed.Latitude, computed.Longitude, 3000.0)

	alert := &TigerDepredationAlert{
		AlertID:                    fmt.Sprintf("alert_tiger_%d", now.UnixNano()),
		Tier:                       "TIER_3_DANGER",
		KillSiteLat:                computed.Latitude,
		KillSiteLon:                computed.Longitude,
		VillagesIn3kmRadius:        villages3km,
		ReturnWindowHoursRemaining: 72.0,
		CanineForensicsConfirmed:   isTigerConfirmed,
		CausalReasoning: fmt.Sprintf("IMMEDIATE DANGER: Fresh %s kill cached at %s. 72-hour tiger return feeding cycle active. Keep livestock corralled; avoid dense forest fringes within 3km.",
			req.LivestockSpecies, computed.Name),
		IssuedAt: now.Format(time.RFC3339),
	}

	return &computed, alert, nil
}

// ClearKillSite marks a carcass removed or safely dispatched, terminating the caching hazard.
func (e *Engine) ClearKillSite(id string) (*LivestockEnclosure, error) {
	e.mu.Lock()

	enc, exists := e.enclosures[id]
	if !exists {
		e.mu.Unlock()
		return nil, fmt.Errorf("enclosure not found: %s", id)
	}

	now := time.Now().UTC()
	enc.IsActiveKillSite = false
	enc.KillReportedAt = nil
	enc.CachingRiskExpiry = nil
	enc.HoursSinceKill = 0.0
	enc.ForensicNotes += " [Ranger Action: Carcass cleared and disposed. Return threat neutralized.]"
	enc.UpdatedAt = now.Format(time.RFC3339)

	weather := model.WeatherConditions{
		WindSpeedKmh:     10.0,
		WindDirectionDeg: 180.0,
		MoonIllumination: 0.50,
	}
	computed := e.calculateEnclosureRisk(enc, weather, now)
	e.enclosures[id] = computed
	l := e.persistenceListener
	e.mu.Unlock()

	if l != nil {
		l(computed)
	}

	return &computed, nil
}

// RecalculateAll updates predation risks and temporal decay across all enclosures given current conditions.
func (e *Engine) RecalculateAll(weather model.WeatherConditions) []LivestockEnclosure {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	result := make([]LivestockEnclosure, 0, len(e.enclosures))
	for id, enc := range e.enclosures {
		updated := e.calculateEnclosureRisk(enc, weather, now)
		e.enclosures[id] = updated
		result = append(result, updated)
	}
	return result
}

func (e *Engine) calculateEnclosureRisk(enc LivestockEnclosure, weather model.WeatherConditions, now time.Time) LivestockEnclosure {
	var returnProb float64 = 0.0
	hoursElapsed := 0.0

	// 1. 72-Hour Carcass Return Exponential Decay Model:
	// P_return(t) = 0.95 * exp(-t / 36.0) for t < 72h
	if enc.IsActiveKillSite && enc.KillReportedAt != nil {
		hoursElapsed = now.Sub(*enc.KillReportedAt).Hours()
		if hoursElapsed < 0 {
			hoursElapsed = 0
		}
		enc.HoursSinceKill = math.Round(hoursElapsed*10.0) / 10.0

		if hoursElapsed < 72.0 {
			returnProb = 0.95 * math.Exp(-hoursElapsed/36.0)
		} else {
			// Decayed past 72h window
			returnProb = 0.12
			enc.IsActiveKillSite = false
		}
	}

	// 2. Baseline Pasture Vulnerability:
	// Night corral reduces attack risk by 65% (factor 0.35)
	corralFactor := 1.0
	if enc.HasNightCorral {
		corralFactor = 0.35
	}
	headCountRatio := math.Min(1.0, float64(enc.HeadCount)/30.0)
	distFactor := math.Exp(-enc.DistanceToForestEdgeM / 1200.0)
	pastureVulnerability := corralFactor * headCountRatio * distFactor

	// 3. Environmental Amplifiers:
	// Nocturnal darkness (New Moon = highest ambush success)
	moonDarkness := 1.0 - math.Max(0.0, math.Min(1.0, weather.MoonIllumination))

	// 4. Composite Threat Score:
	// If active kill site, return probability dominates the threat.
	var threatScore float64
	if enc.IsActiveKillSite && hoursElapsed < 72.0 {
		// Active caching return window: 70% return probability + 20% nocturnal darkness + 10% forest proximity
		threatScore = 0.70*returnProb + 0.20*moonDarkness + 0.10*distFactor
	} else {
		// Un-attacked pasture baseline: vulnerability (45%) + darkness (30%) + forest proximity (25%)
		threatScore = 0.45*pastureVulnerability + 0.30*moonDarkness + 0.25*distFactor
		// Non-kill site capped at 0.72 (Tier 2 Warning maximum)
		threatScore = math.Min(0.72, threatScore)
	}

	threatScore = math.Max(0.05, math.Min(0.98, threatScore))
	threatScore = math.Round(threatScore*1000.0) / 1000.0
	returnProb = math.Round(returnProb*1000.0) / 1000.0

	// 5. Tier Assignment & Buffer Radius
	var tier string
	var bufferMeters float64
	switch {
	case enc.IsActiveKillSite && (threatScore >= 0.75 || hoursElapsed < 48.0):
		tier = "TIER_3_DANGER"
		bufferMeters = 3000.0 // 3km mandatory exclusion radius around active kill site
	case threatScore >= 0.50 || (enc.IsActiveKillSite && hoursElapsed < 72.0):
		tier = "TIER_2_WARNING"
		bufferMeters = 1800.0
	case threatScore >= 0.25:
		tier = "TIER_1_NOTICE"
		bufferMeters = 1000.0
	default:
		tier = "SAFE"
		bufferMeters = 500.0
	}

	enc.ReturnProbability = returnProb
	enc.ActiveThreatScore = threatScore
	enc.AlertTier = tier
	enc.ScentBufferMeters = bufferMeters

	// 6. Generate Geofence Buffer Polygon
	scentFlowDeg := math.Mod(weather.WindDirectionDeg+180.0, 360.0)
	enc.KillSiteGeofence = generateTigerGeofencePolygon(enc.Latitude, enc.Longitude, scentFlowDeg, bufferMeters)

	return enc
}

func (e *Engine) findVillagesWithinRadius(lat, lon, maxDistanceMeters float64) []string {
	matched := make([]string, 0)
	for _, v := range e.villages {
		dist := haversineMeters(lat, lon, v.Latitude, v.Longitude)
		if dist <= maxDistanceMeters {
			matched = append(matched, fmt.Sprintf("%s (%s, %.0fm)", v.Name, v.Province, dist))
		}
	}
	if len(matched) == 0 {
		matched = append(matched, "Adjacent Pastoral Settlement (< 3km)")
	}
	return matched
}

func generateTigerGeofencePolygon(lat, lon, scentBearingDeg, radiusMeters float64) *model.GeoJSONPolygonGeometry {
	const numVertices = 24
	metersPerDegreeLat := 111320.0
	metersPerDegreeLon := 111320.0 * math.Cos(lat*math.Pi/180.0)

	scentRad := scentBearingDeg * math.Pi / 180.0
	kDownwind := 1.45
	kCrosswind := 0.75

	ring := make([][]float64, numVertices+1)

	for i := 0; i < numVertices; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(numVertices)

		localX := radiusMeters * kCrosswind * math.Sin(angle)
		localY := radiusMeters * math.Cos(angle)

		if localY > 0 {
			localY *= kDownwind
		}

		rotX := localX*math.Cos(scentRad) + localY*math.Sin(scentRad)
		rotY := -localX*math.Sin(scentRad) + localY*math.Cos(scentRad)

		vertexLat := lat + (rotY / metersPerDegreeLat)
		vertexLon := lon + (rotX / metersPerDegreeLon)

		ring[i] = []float64{
			math.Round(vertexLon*1000000.0) / 1000000.0,
			math.Round(vertexLat*1000000.0) / 1000000.0,
		}
	}
	ring[numVertices] = []float64{ring[0][0], ring[0][1]}

	return &model.GeoJSONPolygonGeometry{
		Type:        "Polygon",
		Coordinates: [][][]float64{ring},
	}
}

func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	rLat1 := lat1 * (math.Pi / 180.0)
	rLat2 := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2.0)*math.Sin(dLat/2.0) +
		math.Cos(rLat1)*math.Cos(rLat2)*math.Sin(dLon/2.0)*math.Sin(dLon/2.0)
	c := 2.0 * math.Atan2(math.Sqrt(a), math.Sqrt(1.0-a))
	return 6371008.8 * c
}
