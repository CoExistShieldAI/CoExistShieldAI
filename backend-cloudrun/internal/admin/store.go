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

	// 1. Initial Geofences across JAPAC (Thailand, Sumatra, India, Japan Honshu, Hokkaido)
	// Tier 3 Polygon - Kaeng Krachan (Thailand)
	geo1 := AdminGeofence{
		ID:           "geo-ai-thai-01",
		Name:         "Kaeng Krachan Eastern Buffer (HEC Corridor)",
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
		Reason:        "High sugarcane maturation detected; breeding herd movement intersects Route 3219 and Pa La-U access",
	}
	// Tier 2 Circle - Riau / Sumatra (Indonesia)
	centerSumatra := Coordinate{Lat: 0.520, Lng: 101.450}
	geo2 := AdminGeofence{
		ID:                 "geo-ai-sumatra-02",
		Name:               "Siak-Minas Palm Border Tiger Alert",
		Source:             "ai",
		IsPersistent:       false,
		GeometryType:       "circle",
		CircleCenter:       &centerSumatra,
		CircleRadiusMeters: 4500,
		ThreatLevel:        "TIER_2_ALERT",
		TargetSpecies:      "Sumatran Tiger",
		CreatedAt:          now.Add(-4 * time.Hour),
		Reason:             "Recent cattle depredation report & roadkill carcass scent plume drifting towards village border",
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
	// Tier 3 Circle - Kaziranga NH-715 Corridor (India)
	centerIndia := Coordinate{Lat: 26.590, Lng: 93.180}
	geo4 := AdminGeofence{
		ID:                 "geo-ai-india-04",
		Name:               "Kaziranga Panbari-Haldibari Migration Shield",
		Source:             "ai",
		IsPersistent:       false,
		GeometryType:       "circle",
		CircleCenter:       &centerIndia,
		CircleRadiusMeters: 3800,
		ThreatLevel:        "TIER_3_CRITICAL",
		TargetSpecies:      "Asian Elephant",
		CreatedAt:          now.Add(-10 * time.Hour),
		Reason:             "Brahmaputra flood surge causing megafauna migration south across NH-715 towards Karbi Anglong Hills",
	}
	// Tier 2 Circle - Shiretoko Coastal Pass (Hokkaido, Japan)
	centerHokkaido := Coordinate{Lat: 44.050, Lng: 145.020}
	geo5 := AdminGeofence{
		ID:                 "geo-ai-hokkaido-05",
		Name:               "Shiretoko Route 334 Coastal Brown Bear Buffer",
		Source:             "ai",
		IsPersistent:       false,
		GeometryType:       "circle",
		CircleCenter:       &centerHokkaido,
		CircleRadiusMeters: 3200,
		ThreatLevel:        "TIER_2_ALERT",
		TargetSpecies:      "Brown Bear",
		CreatedAt:          now.Add(-18 * time.Hour),
		Reason:             "Coastal salmon run and sika deer carcass presence drawing hyperphagic adult brown bears",
	}
	// Manual Geofence (Circle) - Ranger Station 4 Patrol
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
		Reason:             "Manual field team patrol perimeter: Night thermal camera deployment",
	}

	s.geofences[geo1.ID] = geo1
	s.geofences[geo2.ID] = geo2
	s.geofences[geo3.ID] = geo3
	s.geofences[geo4.ID] = geo4
	s.geofences[geo5.ID] = geo5
	s.geofences[geoManual.ID] = geoManual

	// 2. Wildlife Sightings (Empirical Web Data across Thailand, Sumatra, India, Japan)
	wList := []AdminWildlifeIncident{
		{
			ID:          "wildlife-001",
			Species:     "Asian Elephant",
			Location:    IncidentLocation{Lat: 12.865, Lng: 99.655, RoadName: "Route 3219 KM 12", Region: "Phetchaburi, Thailand"},
			ThreatLevel: "TIER_3_CRITICAL",
			Status:      "VERIFIED",
			Reporter:    "Ranger Somchai (Kaeng Krachan Unit)",
			Description: "A matriarch herd of 7 elephants crossing highway towards perimeter sugarcane plantations.",
			PhotoURLs:   []string{"https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_elephant_herd.jpg"},
			ReportedAt:  now.Add(-90 * time.Minute),
		},
		{
			ID:          "wildlife-002",
			Species:     "Sumatran Tiger",
			Location:    IncidentLocation{Lat: 0.535, Lng: 101.465, RoadName: "Estate Road 4B", Region: "Riau, Indonesia"},
			ThreatLevel: "TIER_3_CRITICAL",
			Status:      "UNVERIFIED",
			Reporter:    "Plantation Security Patrol",
			Description: "Fresh pugmarks and vocalization identified near oil palm boundary drainage ditch.",
			PhotoURLs:   []string{"https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_tiger_track.jpg"},
			ReportedAt:  now.Add(-3 * time.Hour),
		},
		{
			ID:          "wildlife-003",
			Species:     "Brown Bear",
			Location:    IncidentLocation{Lat: 43.050, Lng: 141.340, RoadName: "Prefectural Road 82", Region: "Sapporo Suburbs, Hokkaido, Japan"},
			ThreatLevel: "TIER_1_CAUTION",
			Status:      "VERIFIED",
			Reporter:    "Municipal Wildlife Control Unit",
			Description: "Single subadult bear observed feeding on wild berry patch near forestry road.",
			PhotoURLs:   []string{},
			ReportedAt:  now.Add(-8 * time.Hour),
		},
		{
			ID:          "wildlife-004",
			Species:     "Asian Black Bear",
			Location:    IncidentLocation{Lat: 39.735, Lng: 141.165, RoadName: "Matsuzono 3-chome Hillside", Region: "Morioka, Iwate, Japan"},
			ThreatLevel: "TIER_3_CRITICAL",
			Status:      "VERIFIED",
			Reporter:    "Iwate Mobile Alert Network",
			Description: "Adult black bear observed climbing abandoned persimmon tree inside residential boundary adjacent to primary school route.",
			PhotoURLs:   []string{"https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_bear_kaki.jpg"},
			ReportedAt:  now.Add(-14 * time.Hour),
		},
		{
			ID:          "wildlife-005",
			Species:     "Asian Elephant",
			Location:    IncidentLocation{Lat: 26.582, Lng: 93.175, RoadName: "NH-715 Haldibari Corridor", Region: "Kaziranga, Assam, India"},
			ThreatLevel: "TIER_3_CRITICAL",
			Status:      "VERIFIED",
			Reporter:    "Assam Forest Dept Highway Patrol",
			Description: "Breeding herd of 14 elephants crossing National Highway southwards to reach elevated Karbi Anglong refuge.",
			PhotoURLs:   []string{"https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_nh715_crossing.jpg"},
			ReportedAt:  now.Add(-22 * time.Hour),
		},
		{
			ID:          "wildlife-006",
			Species:     "Asian Elephant",
			Location:    IncidentLocation{Lat: 12.540, Lng: 99.530, RoadName: "Protection Unit KK.7 Access Road", Region: "Huai Sat Yai, Pa La-U, Thailand"},
			ThreatLevel: "TIER_2_ALERT",
			Status:      "VERIFIED",
			Reporter:    "DNP Wild Elephant Surveillance Team",
			Description: "Bull elephant 'Plai Boonchuay' foraging along fruit plantation border fence line.",
			PhotoURLs:   []string{},
			ReportedAt:  now.Add(-48 * time.Hour), // 2 days ago
		},
		{
			ID:          "wildlife-007",
			Species:     "Asian Black Bear",
			Location:    IncidentLocation{Lat: 36.662, Lng: 138.210, RoadName: "Suzaka Orchard Ridge Trail", Region: "Nagano, Japan"},
			ThreatLevel: "TIER_2_ALERT",
			Status:      "VERIFIED",
			Reporter:    "Nagano Agricultural Cooperative",
			Description: "Female bear with two cubs feeding in unharvested chestnut grove bordering mountain slope.",
			PhotoURLs:   []string{},
			ReportedAt:  now.Add(-4 * 24 * time.Hour), // 4 days ago
		},
		{
			ID:          "wildlife-008",
			Species:     "Brown Bear",
			Location:    IncidentLocation{Lat: 44.045, Lng: 145.030, RoadName: "Route 334 Coastal Pass", Region: "Rausu, Shiretoko, Hokkaido, Japan"},
			ThreatLevel: "TIER_2_ALERT",
			Status:      "VERIFIED",
			Reporter:    "Shiretoko Nature Foundation Ranger",
			Description: "Large male brown bear tracking coastal salmon scent near river mouth, 80m from roadway.",
			PhotoURLs:   []string{"https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_brown_bear.jpg"},
			ReportedAt:  now.Add(-6 * 24 * time.Hour), // 6 days ago
		},
		{
			ID:          "wildlife-009",
			Species:     "Sumatran Tiger",
			Location:    IncidentLocation{Lat: 0.785, Lng: 102.045, RoadName: "Sungai Apit Peatland Edge", Region: "Siak, Riau, Indonesia"},
			ThreatLevel: "TIER_3_CRITICAL",
			Status:      "VERIFIED",
			Reporter:    "BBKSDA Riau Rapid Response Unit",
			Description: "Tiger attempted night predation on communal livestock pen; clear track trail confirmed.",
			PhotoURLs:   []string{},
			ReportedAt:  now.Add(-12 * 24 * time.Hour), // 12 days ago
		},
		{
			ID:          "wildlife-010",
			Species:     "Bengal Tiger",
			Location:    IncidentLocation{Lat: 27.410, Lng: 84.120, RoadName: "Valmiki-Chitwan Border Trail", Region: "Terai Arc, Bihar/Nepal Border"},
			ThreatLevel: "TIER_2_ALERT",
			Status:      "VERIFIED",
			Reporter:    "WWF Camera Trap Survey Team",
			Description: "Tigress recorded using agricultural irrigation canal as travel corridor between forest patches.",
			PhotoURLs:   []string{},
			ReportedAt:  now.Add(-20 * 24 * time.Hour), // 20 days ago
		},
	}
	for _, w := range wList {
		s.wildlife[w.ID] = w
	}

	// 3. Roadkill Records (Empirical Carcasses, Decay Stages & Scent Plume Magnets)
	rkList := []AdminRoadkillRecord{
		{
			ID:                "rk-2026-001",
			Species:           "Wild Boar",
			EstimatedWeightKg: 55,
			DecayStage:        "FRESH",
			Location:          IncidentLocation{Lat: 0.512, Lng: 101.438, RoadName: "Trans-Sumatran Highway KM 84", Region: "Riau, Indonesia"},
			PhotoURL:          "https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_rk_boar.jpg",
			PhotoURLs:         []string{"https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_rk_boar.jpg"},
			ReliabilityScore:  0.95,
			IsExcluded:        false,
			Status:            "PENDING_REMOVAL",
			ReportedAt:        now.Add(-45 * time.Minute),
		},
		{
			ID:                "rk-2026-002",
			Species:           "Sambar Deer",
			EstimatedWeightKg: 120,
			DecayStage:        "MODERATE",
			Location:          IncidentLocation{Lat: 12.842, Lng: 99.680, RoadName: "Highway 3219 KM 18", Region: "Phetchaburi, Thailand"},
			PhotoURL:          "",
			PhotoURLs:         []string{},
			ReliabilityScore:  0.88,
			IsExcluded:        false,
			Status:            "PENDING_REMOVAL",
			ReportedAt:        now.Add(-5 * time.Hour),
		},
		{
			ID:                "rk-2026-003",
			Species:           "Unknown Small Mammal",
			EstimatedWeightKg: 5,
			DecayStage:        "ADVANCED",
			Location:          IncidentLocation{Lat: 36.635, Lng: 138.210, RoadName: "Local Route 406", Region: "Nagano, Japan"},
			PhotoURL:          "",
			ReliabilityScore:  0.35,
			IsExcluded:        true,
			Status:            "DISPOSED",
			ReportedAt:        now.Add(-24 * time.Hour),
		},
		{
			ID:                "rk-2026-004",
			Species:           "Hog Deer",
			EstimatedWeightKg: 42,
			DecayStage:        "FRESH",
			Location:          IncidentLocation{Lat: 26.588, Lng: 93.185, RoadName: "NH-715 Panbari Bottleneck", Region: "Kaziranga, Assam, India"},
			PhotoURL:          "https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_rk_hogdeer.jpg",
			PhotoURLs:         []string{"https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_rk_hogdeer.jpg"},
			ReliabilityScore:  0.96,
			IsExcluded:        false,
			Status:            "PENDING_REMOVAL",
			ReportedAt:        now.Add(-3 * time.Hour),
		},
		{
			ID:                "rk-2026-005",
			Species:           "Ezo Sika Deer",
			EstimatedWeightKg: 115,
			DecayStage:        "FRESH",
			Location:          IncidentLocation{Lat: 43.620, Lng: 144.980, RoadName: "National Route 272", Region: "Nemuro, Hokkaido, Japan"},
			PhotoURL:          "https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_rk_sikadeer.jpg",
			PhotoURLs:         []string{"https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_rk_sikadeer.jpg"},
			ReliabilityScore:  0.94,
			IsExcluded:        false,
			Status:            "PENDING_REMOVAL",
			ReportedAt:        now.Add(-5 * time.Hour),
		},
		{
			ID:                "rk-2026-006",
			Species:           "Japanese Serow (Kamoshika)",
			EstimatedWeightKg: 38,
			DecayStage:        "MODERATE",
			Location:          IncidentLocation{Lat: 39.810, Lng: 141.420, RoadName: "National Route 455", Region: "Iwaizumi, Iwate, Japan"},
			PhotoURL:          "",
			ReliabilityScore:  0.89,
			IsExcluded:        false,
			Status:            "DISPOSED",
			ReportedAt:        now.Add(-36 * time.Hour),
		},
		{
			ID:                "rk-2026-007",
			Species:           "Wild Boar",
			EstimatedWeightKg: 68,
			DecayStage:        "MODERATE",
			Location:          IncidentLocation{Lat: 12.528, Lng: 99.542, RoadName: "Route 3201 Hua Hin Bypass", Region: "Prachuap Khiri Khan, Thailand"},
			PhotoURL:          "",
			ReliabilityScore:  0.82,
			IsExcluded:        false,
			Status:            "DISPOSED",
			ReportedAt:        now.Add(-3 * 24 * time.Hour),
		},
		{
			ID:                "rk-2026-008",
			Species:           "Domestic Cattle (Calf)",
			EstimatedWeightKg: 85,
			DecayStage:        "ADVANCED",
			Location:          IncidentLocation{Lat: 0.525, Lng: 101.480, RoadName: "Minas Concession Road", Region: "Siak, Riau, Indonesia"},
			PhotoURL:          "",
			ReliabilityScore:  0.90,
			IsExcluded:        false,
			Status:            "DISPOSED",
			ReportedAt:        now.Add(-5 * 24 * time.Hour),
		},
		{
			ID:                "rk-2026-009",
			Species:           "Red Fox",
			EstimatedWeightKg: 7,
			DecayStage:        "ADVANCED",
			Location:          IncidentLocation{Lat: 44.060, Lng: 145.010, RoadName: "Route 334 Pass", Region: "Shiretoko, Hokkaido, Japan"},
			PhotoURL:          "",
			ReliabilityScore:  0.40,
			IsExcluded:        true,
			Status:            "DISPOSED",
			ReportedAt:        now.Add(-11 * 24 * time.Hour),
		},
		{
			ID:                "rk-2026-010",
			Species:           "Wild Boar",
			EstimatedWeightKg: 52,
			DecayStage:        "ADVANCED",
			Location:          IncidentLocation{Lat: 26.540, Lng: 93.050, RoadName: "Kaliabor Bypass NH-715", Region: "Nagaon, Assam, India"},
			PhotoURL:          "",
			ReliabilityScore:  0.75,
			IsExcluded:        false,
			Status:            "DISPOSED",
			ReportedAt:        now.Add(-18 * 24 * time.Hour),
		},
	}
	for _, rk := range rkList {
		s.roadkills[rk.ID] = rk
	}

	// 4. Initial Environmental Attractants (Crops, Persimmons, Pastures, Water Holes)
	envList := []AdminEnvironmentalData{
		{
			ID:           "env-crop-01",
			Category:     "CROP_PHENOLOGY",
			Name:         "Kaeng Krachan Perimeter Sugarcane Sector A",
			LocationDesc: "Eastern boundary agro-forest interface",
			GeometryType: "circle",
			CircleCenter: &Coordinate{Lat: 12.870, Lng: 99.660},
			RadiusMeters: 3000,
			RiskWeight:   0.85,
			ActiveSeason: "October - January (Harvest Peak)",
			PhotoURLs:    []string{"https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_sugarcane_field.jpg"},
			LastUpdated:  now.Add(-12 * time.Hour),
		},
		{
			ID:           "env-kaki-02",
			Category:     "PERSIMMON_KAKI",
			Name:         "Nagano Depopulated Valley Orchard Cluster",
			LocationDesc: "Abandoned residential hillside trees",
			GeometryType: "circle",
			CircleCenter: &Coordinate{Lat: 36.660, Lng: 138.190},
			RadiusMeters: 1800,
			RiskWeight:   0.90,
			ActiveSeason: "September - November (Pre-hibernation Hyperphagia)",
			PhotoURLs:    []string{},
			LastUpdated:  now.Add(-1 * time.Hour),
		},
		{
			ID:           "env-pasture-03",
			Category:     "LIVESTOCK_PASTURE",
			Name:         "Siak Fringe Communal Cattle Corral",
			LocationDesc: "Open night grazing pasture bordering oil palm",
			GeometryType: "circle",
			CircleCenter: &Coordinate{Lat: 0.510, Lng: 101.470},
			RadiusMeters: 2200,
			RiskWeight:   0.78,
			ActiveSeason: "Year-round",
			PhotoURLs:    []string{},
			LastUpdated:  now.Add(-2 * time.Hour),
		},
		{
			ID:           "env-kaki-04",
			Category:     "PERSIMMON_KAKI",
			Name:         "Morioka Matsuzono Abandoned Fruit Tree Cluster",
			LocationDesc: "Residential hillside fringe fruit groves near elementary school",
			GeometryType: "circle",
			CircleCenter: &Coordinate{Lat: 39.735, Lng: 141.165},
			RadiusMeters: 1500,
			RiskWeight:   0.92,
			ActiveSeason: "October - December (Peak Fruiting & Pre-Denning)",
			PhotoURLs:    []string{"https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_kaki_trees.jpg"},
			LastUpdated:  now.Add(-6 * time.Hour),
		},
		{
			ID:           "env-water-05",
			Category:     "WATER_SOURCE",
			Name:         "Huai Sat Yai Reservoir Channel Watering Point",
			LocationDesc: "Dry season sole watering access between forest and village",
			GeometryType: "circle",
			CircleCenter: &Coordinate{Lat: 12.535, Lng: 99.525},
			RadiusMeters: 1200,
			RiskWeight:   0.75,
			ActiveSeason: "February - May (Dry Season Drought)",
			PhotoURLs:    []string{},
			LastUpdated:  now.Add(-18 * time.Hour),
		},
		{
			ID:           "env-crop-06",
			Category:     "CROP_PHENOLOGY",
			Name:         "Kaziranga Bokakhat Tea & Lowland Buffer Belt",
			LocationDesc: "Wetland buffer corridor between NH-715 and tea plantation edge",
			GeometryType: "circle",
			CircleCenter: &Coordinate{Lat: 26.615, Lng: 93.590},
			RadiusMeters: 3200,
			RiskWeight:   0.88,
			ActiveSeason: "June - September (Monsoon Inundation Dispersal)",
			PhotoURLs:    []string{},
			LastUpdated:  now.Add(-24 * time.Hour),
		},
		{
			ID:           "env-crop-07",
			Category:     "CROP_PHENOLOGY",
			Name:         "Eniwa Agricultural Sweetcorn & Sugarbeet Belt",
			LocationDesc: "Forest edge crop plots attractive to brown bears prior to denning",
			GeometryType: "circle",
			CircleCenter: &Coordinate{Lat: 42.920, Lng: 141.580},
			RadiusMeters: 2500,
			RiskWeight:   0.82,
			ActiveSeason: "August - October (Sweetcorn Ripening Peak)",
			PhotoURLs:    []string{},
			LastUpdated:  now.Add(-48 * time.Hour),
		},
		{
			ID:           "env-pasture-08",
			Category:     "LIVESTOCK_PASTURE",
			Name:         "Pelalawan Concession Night Grazing Enclosure",
			LocationDesc: "Communal sheep and cattle pens bordering pulpwood concession",
			GeometryType: "circle",
			CircleCenter: &Coordinate{Lat: 0.360, Lng: 101.850},
			RadiusMeters: 2000,
			RiskWeight:   0.80,
			ActiveSeason: "Year-round",
			PhotoURLs:    []string{},
			LastUpdated:  now.Add(-72 * time.Hour),
		},
	}
	for _, env := range envList {
		s.environmental[env.ID] = env
	}
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
