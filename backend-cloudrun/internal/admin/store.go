package admin

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Store struct {
	mu            sync.RWMutex
	geofences     map[string]AdminGeofence
	wildlife      map[string]AdminWildlifeIncident
	roadkills     map[string]AdminRoadkillRecord
	environmental map[string]AdminEnvironmentalData
	userCount     int
}

func NewStore() *Store {
	s := &Store{
		geofences:     make(map[string]AdminGeofence),
		wildlife:      make(map[string]AdminWildlifeIncident),
		roadkills:     make(map[string]AdminRoadkillRecord),
		environmental: make(map[string]AdminEnvironmentalData),
		userCount:     1420,
	}
	s.seedDefaultData()
	return s
}

func (s *Store) seedDefaultData() {
	now := time.Now().UTC()

	// 1. Initial Geofences across JAPAC
	// Tier 3 Polygon - Kaeng Krachan (Thailand)
	geo1 := AdminGeofence{
		ID:           "geo-ai-thai-01",
		Name:         "Kaeng Krachan Eastern Buffer (HEC Risk)",
		Source:       "ai",
		IsPersistent: false,
		GeometryType: "polygon",
		Coordinates: []Coordinate{
			{Lat: 12.875, Lng: 99.630},
			{Lat: 12.890, Lng: 99.670},
			{Lat: 12.860, Lng: 99.695},
			{Lat: 12.845, Lng: 99.650},
		},
		ThreatLevel:   "TIER_3_CRITICAL",
		TargetSpecies: "Asian Elephant",
		CreatedAt:     now.Add(-2 * time.Hour),
		Reason:        "High sugarcane sugar maturation detected; herd movement corridor intersects Route 3219",
	}
	// Tier 2 Circle - Riau / Sumatra (Indonesia)
	centerSumatra := Coordinate{Lat: 0.520, Lng: 101.450}
	geo2 := AdminGeofence{
		ID:                 "geo-ai-sumatra-02",
		Name:               "Siak Palm Border Tiger Alert",
		Source:             "ai",
		IsPersistent:       false,
		GeometryType:       "circle",
		CircleCenter:       &centerSumatra,
		CircleRadiusMeters: 4500,
		ThreatLevel:        "TIER_2_ALERT",
		TargetSpecies:      "Sumatran Tiger",
		CreatedAt:          now.Add(-4 * time.Hour),
		Reason:             "Recent cattle depredation report & roadkill carcass scent plume drifting towards village",
	}
	// Tier 1 Polygon - Nagano (Japan)
	geo3 := AdminGeofence{
		ID:           "geo-ai-japan-03",
		Name:         "Nagano Valley Bear Foraging Zone",
		Source:       "ai",
		IsPersistent: false,
		GeometryType: "polygon",
		Coordinates: []Coordinate{
			{Lat: 36.650, Lng: 138.180},
			{Lat: 36.675, Lng: 138.220},
			{Lat: 36.640, Lng: 138.245},
			{Lat: 36.620, Lng: 138.200},
		},
		ThreatLevel:   "TIER_1_CAUTION",
		TargetSpecies: "Asian Black Bear",
		CreatedAt:     now.Add(-6 * time.Hour),
		Reason:        "Clustered abandoned persimmon (kaki) trees entering peak ripeness ahead of hibernation",
	}
	// Manual Geofence (Circle) - Debug / Ranger Station override
	centerManual := Coordinate{Lat: 12.820, Lng: 99.620}
	geoManual := AdminGeofence{
		ID:                 "geo-manual-debug-01",
		Name:               "Ranger Station 4 Patrol Perimeter",
		Source:             "manual",
		IsPersistent:       true,
		GeometryType:       "circle",
		CircleCenter:       &centerManual,
		CircleRadiusMeters: 2500,
		ThreatLevel:        "TIER_2_ALERT",
		TargetSpecies:      "Asian Elephant",
		CreatedAt:          now.Add(-1 * time.Hour),
		Reason:             "Manual field team debug perimeter: Night thermal camera deployment",
	}

	s.geofences[geo1.ID] = geo1
	s.geofences[geo2.ID] = geo2
	s.geofences[geo3.ID] = geo3
	s.geofences[geoManual.ID] = geoManual

	// 2. Initial Wildlife Sightings
	w1 := AdminWildlifeIncident{
		ID:      "wildlife-001",
		Species: "Asian Elephant",
		Location: IncidentLocation{
			Lat:      12.865,
			Lng:      99.655,
			RoadName: "Route 3219 KM 12",
			Region:   "Phetchaburi, Thailand",
		},
		ThreatLevel: "TIER_3_CRITICAL",
		Status:      "VERIFIED",
		Reporter:    "Ranger Somchai (Kaeng Krachan Unit)",
		Description: "A matriarch herd of 7 elephants observed crossing highway towards sugarcane plantations.",
		ReportedAt:  now.Add(-90 * time.Minute),
	}
	w2 := AdminWildlifeIncident{
		ID:      "wildlife-002",
		Species: "Sumatran Tiger",
		Location: IncidentLocation{
			Lat:      0.535,
			Lng:      101.465,
			RoadName: "Estate Road 4B",
			Region:   "Riau, Indonesia",
		},
		ThreatLevel: "TIER_3_CRITICAL",
		Status:      "UNVERIFIED",
		Reporter:    "Plantation Security Patrol",
		Description: "Fresh pugmarks and vocalization identified near boundary ditch.",
		ReportedAt:  now.Add(-3 * time.Hour),
	}
	w3 := AdminWildlifeIncident{
		ID:      "wildlife-003",
		Species: "Brown Bear",
		Location: IncidentLocation{
			Lat:      43.050,
			Lng:      141.340,
			RoadName: "Prefectural Road 82",
			Region:   "Sapporo Suburbs, Hokkaido, Japan",
		},
		ThreatLevel: "TIER_1_CAUTION",
		Status:      "VERIFIED",
		Reporter:    "Municipal Wildlife Control Unit",
		Description: "Single subadult bear observed feeding on wild berry patch near forestry road.",
		ReportedAt:  now.Add(-8 * time.Hour),
	}
	s.wildlife[w1.ID] = w1
	s.wildlife[w2.ID] = w2
	s.wildlife[w3.ID] = w3

	// 3. Initial Roadkill Records
	rk1 := AdminRoadkillRecord{
		ID:                "rk-2026-001",
		Species:           "Wild Boar",
		EstimatedWeightKg: 55,
		DecayStage:        "FRESH",
		Location: IncidentLocation{
			Lat:      0.512,
			Lng:      101.438,
			RoadName: "Trans-Sumatran Highway KM 84",
			Region:   "Riau, Indonesia",
		},
		PhotoURL:         "https://storage.googleapis.com/coexist-mock/rk-boar-01.jpg",
		ReliabilityScore: 0.95,
		IsExcluded:       false,
		Status:           "PENDING_REMOVAL",
		ReportedAt:       now.Add(-45 * time.Minute),
	}
	rk2 := AdminRoadkillRecord{
		ID:                "rk-2026-002",
		Species:           "Sambar Deer",
		EstimatedWeightKg: 120,
		DecayStage:        "MODERATE",
		Location: IncidentLocation{
			Lat:      12.842,
			Lng:      99.680,
			RoadName: "Highway 3219 KM 18",
			Region:   "Phetchaburi, Thailand",
		},
		PhotoURL:         "https://storage.googleapis.com/coexist-mock/rk-deer-02.jpg",
		ReliabilityScore: 0.88,
		IsExcluded:       false,
		Status:           "PENDING_REMOVAL",
		ReportedAt:       now.Add(-5 * time.Hour),
	}
	rk3 := AdminRoadkillRecord{
		ID:                "rk-2026-003",
		Species:           "Unknown Small Mammal",
		EstimatedWeightKg: 5,
		DecayStage:        "ADVANCED",
		Location: IncidentLocation{
			Lat:      36.635,
			Lng:      138.210,
			RoadName: "Local Route 406",
			Region:   "Nagano, Japan",
		},
		PhotoURL:         "",
		ReliabilityScore: 0.35, // Low credibility
		IsExcluded:       true, // Excluded from AI training
		Status:           "DISPOSED",
		ReportedAt:       now.Add(-24 * time.Hour),
	}
	s.roadkills[rk1.ID] = rk1
	s.roadkills[rk2.ID] = rk2
	s.roadkills[rk3.ID] = rk3

	// 4. Initial Environmental Attractants
	env1 := AdminEnvironmentalData{
		ID:           "env-crop-01",
		Category:     "CROP_PHENOLOGY",
		Name:         "Kaeng Krachan Perimeter Sugarcane Sector A",
		LocationDesc: "Eastern boundary agro-forest interface",
		GeometryType: "circle",
		CircleCenter: &Coordinate{Lat: 12.870, Lng: 99.660},
		RadiusMeters: 3000,
		RiskWeight:   0.85,
		ActiveSeason: "October - January (Harvest Peak)",
		LastUpdated:  now.Add(-12 * time.Hour),
	}
	env2 := AdminEnvironmentalData{
		ID:           "env-kaki-02",
		Category:     "PERSIMMON_KAKI",
		Name:         "Nagano Depopulated Valley Orchard Cluster",
		LocationDesc: "Abandoned residential hillside trees",
		GeometryType: "circle",
		CircleCenter: &Coordinate{Lat: 36.660, Lng: 138.190},
		RadiusMeters: 1800,
		RiskWeight:   0.90,
		ActiveSeason: "September - November (Pre-hibernation Hyperphagia)",
		LastUpdated:  now.Add(-1 * time.Hour),
	}
	env3 := AdminEnvironmentalData{
		ID:           "env-pasture-03",
		Category:     "LIVESTOCK_PASTURE",
		Name:         "Siak Fringe Communal Cattle Corral",
		LocationDesc: "Open night grazing pasture bordering oil palm",
		GeometryType: "circle",
		CircleCenter: &Coordinate{Lat: 0.510, Lng: 101.470},
		RadiusMeters: 2200,
		RiskWeight:   0.78,
		ActiveSeason: "Year-round",
		LastUpdated:  now.Add(-2 * time.Hour),
	}
	s.environmental[env1.ID] = env1
	s.environmental[env2.ID] = env2
	s.environmental[env3.ID] = env3
}

// --- Geofences CRUD ---

func (s *Store) ListGeofences(sourceFilter string) []AdminGeofence {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]AdminGeofence, 0, len(s.geofences))
	for _, g := range s.geofences {
		if sourceFilter == "ai" && g.Source != "ai" {
			continue
		}
		if sourceFilter == "manual" && g.Source != "manual" {
			continue
		}
		res = append(res, g)
	}
	return res
}

func (s *Store) CreateGeofence(g AdminGeofence) AdminGeofence {
	s.mu.Lock()
	defer s.mu.Unlock()

	if g.ID == "" {
		g.ID = fmt.Sprintf("geo-manual-%d", time.Now().UnixNano()%100000)
	}
	if g.Source == "" {
		g.Source = "manual"
	}
	if g.CreatedAt.IsZero() {
		g.CreatedAt = time.Now().UTC()
	}
	s.geofences[g.ID] = g
	return g
}

func (s *Store) DeleteGeofence(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.geofences[id]; !exists {
		return false
	}
	delete(s.geofences, id)
	return true
}

func (s *Store) ClearAllManualGeofences() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	deletedCount := 0
	for id, g := range s.geofences {
		if g.Source == "manual" {
			delete(s.geofences, id)
			deletedCount++
		}
	}
	return deletedCount
}

func (s *Store) RecalculateAiGeofences() (int, time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	// Update timestamp and slight jitter to simulate fresh AI inference without wiping manual fences
	aiCount := 0
	for id, g := range s.geofences {
		if g.Source == "ai" {
			g.CreatedAt = now
			s.geofences[id] = g
			aiCount++
		}
	}
	return aiCount, now
}

// --- Wildlife Incidents CRUD ---

func (s *Store) ListWildlife() []AdminWildlifeIncident {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]AdminWildlifeIncident, 0, len(s.wildlife))
	for _, w := range s.wildlife {
		res = append(res, w)
	}
	return res
}

func (s *Store) CreateWildlife(w AdminWildlifeIncident) AdminWildlifeIncident {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w.ID == "" {
		w.ID = fmt.Sprintf("wildlife-%d", time.Now().UnixNano()%100000)
	}
	if w.ReportedAt.IsZero() {
		w.ReportedAt = time.Now().UTC()
	}
	if w.Status == "" {
		w.Status = "UNVERIFIED"
	}
	s.wildlife[w.ID] = w
	return w
}

func (s *Store) DeleteWildlife(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.wildlife[id]; !exists {
		return false
	}
	delete(s.wildlife, id)
	return true
}

func (s *Store) UpdateWildlifeStatus(id string, status string) (AdminWildlifeIncident, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	w, exists := s.wildlife[id]
	if !exists {
		return AdminWildlifeIncident{}, false
	}
	w.Status = status
	s.wildlife[id] = w
	return w, true
}

// --- Roadkill Records CRUD ---

func (s *Store) ListRoadkills(includeExcluded bool) []AdminRoadkillRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]AdminRoadkillRecord, 0, len(s.roadkills))
	for _, rk := range s.roadkills {
		if !includeExcluded && rk.IsExcluded {
			continue
		}
		res = append(res, rk)
	}
	return res
}

func (s *Store) CreateRoadkill(rk AdminRoadkillRecord) AdminRoadkillRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rk.ID == "" {
		rk.ID = fmt.Sprintf("rk-%d", time.Now().UnixNano()%100000)
	}
	if rk.ReportedAt.IsZero() {
		rk.ReportedAt = time.Now().UTC()
	}
	if rk.Status == "" {
		rk.Status = "PENDING_REMOVAL"
	}
	if rk.ReliabilityScore == 0 {
		rk.ReliabilityScore = 0.90
	}
	if len(rk.PhotoURLs) > 0 && rk.PhotoURL == "" {
		rk.PhotoURL = rk.PhotoURLs[0]
	} else if rk.PhotoURL != "" && len(rk.PhotoURLs) == 0 {
		rk.PhotoURLs = []string{rk.PhotoURL}
	}
	s.roadkills[rk.ID] = rk
	return rk
}

func (s *Store) SetRoadkillExclusion(id string, isExcluded bool) (AdminRoadkillRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rk, exists := s.roadkills[id]
	if !exists {
		return AdminRoadkillRecord{}, false
	}
	rk.IsExcluded = isExcluded
	s.roadkills[id] = rk
	return rk, true
}

func (s *Store) SetRoadkillDisposed(id string) (AdminRoadkillRecord, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rk, exists := s.roadkills[id]
	if !exists {
		return AdminRoadkillRecord{}, false
	}
	rk.Status = "DISPOSED"
	s.roadkills[id] = rk
	return rk, true
}

// --- Environmental Data CRUD ---

func (s *Store) ListEnvironmental() []AdminEnvironmentalData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]AdminEnvironmentalData, 0, len(s.environmental))
	for _, env := range s.environmental {
		res = append(res, env)
	}
	return res
}

func (s *Store) CreateEnvironmental(env AdminEnvironmentalData) AdminEnvironmentalData {
	s.mu.Lock()
	defer s.mu.Unlock()

	if env.ID == "" {
		env.ID = fmt.Sprintf("env-%d", time.Now().UnixNano()%100000)
	}
	if env.LastUpdated.IsZero() {
		env.LastUpdated = time.Now().UTC()
	}
	s.environmental[env.ID] = env
	return env
}

func (s *Store) DeleteEnvironmental(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.environmental[id]; !exists {
		return false
	}
	delete(s.environmental, id)
	return true
}

// --- Dashboard Stats ---

func (s *Store) GetDashboardStats() DashboardStatsResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now().UTC()
	cutoff24h := now.Add(-24 * time.Hour)

	stats := DashboardStatsResponse{
		TotalGeofences:      len(s.geofences),
		TotalRoadkills:      len(s.roadkills),
		TotalWildlifeAll:    len(s.wildlife),
		TotalUsers:          s.userCount,
		SpeciesDistribution: make(map[string]int),
	}

	for _, g := range s.geofences {
		if g.Source == "ai" {
			stats.AiGeofences++
		} else {
			stats.ManualGeofences++
		}
		if g.ThreatLevel == "TIER_3_CRITICAL" {
			stats.Tier3Geofences++
		}
	}

	for _, w := range s.wildlife {
		if w.ReportedAt.After(cutoff24h) {
			stats.TotalWildlife24h++
		}
		stats.SpeciesDistribution[w.Species]++
	}

	for _, rk := range s.roadkills {
		if rk.Status == "PENDING_REMOVAL" {
			stats.PendingRoadkills++
		}
	}

	return stats
}

// --- Export to GeoJSON / CSV ---

func (s *Store) ExportDataset(dataset string, format string) ([]byte, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if format == "csv" {
		return s.exportCSV(dataset)
	}
	return s.exportGeoJSON(dataset)
}

func (s *Store) exportGeoJSON(dataset string) ([]byte, string, error) {
	features := []map[string]interface{}{}

	switch dataset {
	case "geofences":
		for _, g := range s.geofences {
			feat := map[string]interface{}{
				"type": "Feature",
				"properties": map[string]interface{}{
					"id":             g.ID,
					"name":           g.Name,
					"source":         g.Source,
					"threat_level":   g.ThreatLevel,
					"target_species": g.TargetSpecies,
					"is_persistent":  g.IsPersistent,
					"reason":         g.Reason,
				},
			}
			if g.GeometryType == "circle" && g.CircleCenter != nil {
				feat["geometry"] = map[string]interface{}{
					"type":        "Point",
					"coordinates": []float64{g.CircleCenter.Lng, g.CircleCenter.Lat},
				}
				feat["properties"].(map[string]interface{})["radius_meters"] = g.CircleRadiusMeters
			} else {
				coords := make([][]float64, 0, len(g.Coordinates))
				for _, pt := range g.Coordinates {
					coords = append(coords, []float64{pt.Lng, pt.Lat})
				}
				if len(coords) > 0 {
					coords = append(coords, coords[0]) // close ring
				}
				feat["geometry"] = map[string]interface{}{
					"type":        "Polygon",
					"coordinates": [][][]float64{coords},
				}
			}
			features = append(features, feat)
		}
	case "wildlife":
		for _, w := range s.wildlife {
			features = append(features, map[string]interface{}{
				"type": "Feature",
				"geometry": map[string]interface{}{
					"type":        "Point",
					"coordinates": []float64{w.Location.Lng, w.Location.Lat},
				},
				"properties": map[string]interface{}{
					"id":           w.ID,
					"species":      w.Species,
					"threat_level": w.ThreatLevel,
					"status":       w.Status,
					"reporter":     w.Reporter,
					"road_name":    w.Location.RoadName,
					"reported_at":  w.ReportedAt.Format(time.RFC3339),
				},
			})
		}
	case "roadkill":
		for _, rk := range s.roadkills {
			features = append(features, map[string]interface{}{
				"type": "Feature",
				"geometry": map[string]interface{}{
					"type":        "Point",
					"coordinates": []float64{rk.Location.Lng, rk.Location.Lat},
				},
				"properties": map[string]interface{}{
					"id":          rk.ID,
					"species":     rk.Species,
					"decay_stage": rk.DecayStage,
					"status":      rk.Status,
					"reliability": rk.ReliabilityScore,
					"is_excluded": rk.IsExcluded,
					"road_name":   rk.Location.RoadName,
					"reported_at": rk.ReportedAt.Format(time.RFC3339),
				},
			})
		}
	default:
		return nil, "", fmt.Errorf("unknown dataset: %s", dataset)
	}

	geoJson := map[string]interface{}{
		"type":     "FeatureCollection",
		"features": features,
	}
	data, err := json.MarshalIndent(geoJson, "", "  ")
	return data, "application/geo+json", err
}

func (s *Store) exportCSV(dataset string) ([]byte, string, error) {
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	switch dataset {
	case "geofences":
		_ = writer.Write([]string{"ID", "Name", "Source", "ThreatLevel", "TargetSpecies", "GeometryType", "IsPersistent", "CreatedAt"})
		for _, g := range s.geofences {
			_ = writer.Write([]string{
				g.ID, g.Name, g.Source, g.ThreatLevel, g.TargetSpecies, g.GeometryType,
				strconv.FormatBool(g.IsPersistent), g.CreatedAt.Format(time.RFC3339),
			})
		}
	case "wildlife":
		_ = writer.Write([]string{"ID", "Species", "ThreatLevel", "Status", "Reporter", "Latitude", "Longitude", "RoadName", "ReportedAt"})
		for _, w := range s.wildlife {
			_ = writer.Write([]string{
				w.ID, w.Species, w.ThreatLevel, w.Status, w.Reporter,
				strconv.FormatFloat(w.Location.Lat, 'f', 6, 64),
				strconv.FormatFloat(w.Location.Lng, 'f', 6, 64),
				w.Location.RoadName, w.ReportedAt.Format(time.RFC3339),
			})
		}
	case "roadkill":
		_ = writer.Write([]string{"ID", "Species", "WeightKg", "DecayStage", "Status", "Reliability", "IsExcluded", "Latitude", "Longitude", "RoadName", "ReportedAt"})
		for _, rk := range s.roadkills {
			_ = writer.Write([]string{
				rk.ID, rk.Species, strconv.FormatFloat(rk.EstimatedWeightKg, 'f', 1, 64),
				rk.DecayStage, rk.Status, strconv.FormatFloat(rk.ReliabilityScore, 'f', 2, 64),
				strconv.FormatBool(rk.IsExcluded),
				strconv.FormatFloat(rk.Location.Lat, 'f', 6, 64),
				strconv.FormatFloat(rk.Location.Lng, 'f', 6, 64),
				rk.Location.RoadName, rk.ReportedAt.Format(time.RFC3339),
			})
		}
	default:
		return nil, "", fmt.Errorf("unknown dataset: %s", dataset)
	}

	writer.Flush()
	return buf.Bytes(), "text/csv", nil
}
