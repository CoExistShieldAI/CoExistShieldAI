package admin

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Store struct {
	mu              sync.RWMutex
	geofences       map[string]AdminGeofence
	wildlife        map[string]AdminWildlifeIncident
	roadkills       map[string]AdminRoadkillRecord
	environmental   map[string]AdminEnvironmentalData
	weather         map[string]AdminWeatherData
	userCount       int
	isRecalculating bool
	persistence     *FirestorePersistence
	persistWg       sync.WaitGroup
}

// WaitForPersistence blocks until all background Firestore persistence operations complete.
func (s *Store) WaitForPersistence() {
	s.persistWg.Wait()
}

func NewStore() *Store {
	s := &Store{
		geofences:     make(map[string]AdminGeofence),
		wildlife:      make(map[string]AdminWildlifeIncident),
		roadkills:     make(map[string]AdminRoadkillRecord),
		environmental: make(map[string]AdminEnvironmentalData),
		weather:       make(map[string]AdminWeatherData),
		userCount:     1420,
	}
	s.seedDefaultData()
	return s
}

func (s *Store) SetPersistence(p *FirestorePersistence) {
	s.mu.Lock()
	s.persistence = p
	s.mu.Unlock()
}

func (s *Store) GetPersistence() *FirestorePersistence {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.persistence
}

func (s *Store) InitPersistence(ctx context.Context, projectID string) error {
	p := NewFirestorePersistence(projectID)
	s.SetPersistence(p)
	return p.RestoreOrSeed(ctx, s)
}

func (s *Store) persistGeofence(g AdminGeofence) {
	s.mu.RLock()
	p := s.persistence
	s.mu.RUnlock()
	if p != nil {
		s.persistWg.Add(1)
		go func() {
			defer s.persistWg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = p.SaveGeofence(ctx, g)
		}()
	}
}

func (s *Store) persistDeleteGeofence(id string) {
	s.mu.RLock()
	p := s.persistence
	s.mu.RUnlock()
	if p != nil {
		s.persistWg.Add(1)
		go func() {
			defer s.persistWg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = p.DeleteGeofence(ctx, id)
		}()
	}
}

func (s *Store) persistWildlife(w AdminWildlifeIncident) {
	s.mu.RLock()
	p := s.persistence
	s.mu.RUnlock()
	if p != nil {
		s.persistWg.Add(1)
		go func() {
			defer s.persistWg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = p.SaveWildlife(ctx, w)
		}()
	}
}

func (s *Store) persistDeleteWildlife(id string) {
	s.mu.RLock()
	p := s.persistence
	s.mu.RUnlock()
	if p != nil {
		s.persistWg.Add(1)
		go func() {
			defer s.persistWg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = p.DeleteWildlife(ctx, id)
		}()
	}
}

func (s *Store) persistRoadkill(rk AdminRoadkillRecord) {
	s.mu.RLock()
	p := s.persistence
	s.mu.RUnlock()
	if p != nil {
		s.persistWg.Add(1)
		go func() {
			defer s.persistWg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = p.SaveRoadkill(ctx, rk)
		}()
	}
}

func (s *Store) persistDeleteRoadkill(id string) {
	s.mu.RLock()
	p := s.persistence
	s.mu.RUnlock()
	if p != nil {
		s.persistWg.Add(1)
		go func() {
			defer s.persistWg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = p.DeleteRoadkill(ctx, id)
		}()
	}
}

func (s *Store) persistEnvironmental(env AdminEnvironmentalData) {
	s.mu.RLock()
	p := s.persistence
	s.mu.RUnlock()
	if p != nil {
		s.persistWg.Add(1)
		go func() {
			defer s.persistWg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = p.SaveEnvironmental(ctx, env)
		}()
	}
}

func (s *Store) persistDeleteEnvironmental(id string) {
	s.mu.RLock()
	p := s.persistence
	s.mu.RUnlock()
	if p != nil {
		s.persistWg.Add(1)
		go func() {
			defer s.persistWg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = p.DeleteEnvironmental(ctx, id)
		}()
	}
}

func (s *Store) persistWeather(w AdminWeatherData) {
	s.mu.RLock()
	p := s.persistence
	s.mu.RUnlock()
	if p != nil {
		s.persistWg.Add(1)
		go func() {
			defer s.persistWg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = p.SaveWeather(ctx, w)
		}()
	}
}

func (s *Store) persistDeleteWeather(id string) {
	s.mu.RLock()
	p := s.persistence
	s.mu.RUnlock()
	if p != nil {
		s.persistWg.Add(1)
		go func() {
			defer s.persistWg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = p.DeleteWeather(ctx, id)
		}()
	}
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

	s.upsertGraduatedBuffersLocked(geo1, now)
	s.upsertGraduatedBuffersLocked(geo2, now)
	s.upsertGraduatedBuffersLocked(geo3, now)
	s.upsertGraduatedBuffersLocked(geo4, now)
	s.upsertGraduatedBuffersLocked(geo5, now)
	s.upsertGraduatedBuffersLocked(geoManual, now)

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

	// 5. Initial Weather & Meteorological Observation Stations for AI (Covering All 20 JAPAC Conflict Zones)
	weatherList := []AdminWeatherData{
		// --- THAILAND (Asian Elephant HEC Corridors) ---
		{
			ID:                 "weather-thai-01",
			StationName:        "Kaeng Krachan Substation Alpha",
			Region:             "Thailand (Kaeng Krachan / Pa La-U)",
			Latitude:           12.875,
			Longitude:          99.630,
			TemperatureCelsius: 31.5,
			HumidityPct:        82.0,
			WindSpeedKmh:       14.0,
			WindDirectionDeg:   90.0, // East wind
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.12,
			MoonPhase:          "WANING_CRESCENT",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Route 3219 crossing point; east wind drifting sugarcane odor towards elephant reserve border",
			LastUpdated:        now.Add(-20 * time.Minute),
		},
		{
			ID:                 "weather-thai-02",
			StationName:        "Kui Buri Agro-Forest Met Tower",
			Region:             "Thailand (Kui Buri / Prachuap)",
			Latitude:           12.062,
			Longitude:          99.782,
			TemperatureCelsius: 30.2,
			HumidityPct:        85.0,
			WindSpeedKmh:       11.0,
			WindDirectionDeg:   80.0,
			PrecipitationMmH:   1.5,
			MoonIllumination:   0.12,
			MoonPhase:          "WANING_CRESCENT",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Pineapple & jackfruit buffer; high-sugar crop volatility triggering breeding herd incursions",
			LastUpdated:        now.Add(-35 * time.Minute),
		},
		{
			ID:                 "weather-thai-03",
			StationName:        "Khao Yai Dong Phayayen Gateway",
			Region:             "Thailand (Khao Yai / Nakhon Ratchasima)",
			Latitude:           14.439,
			Longitude:          101.372,
			TemperatureCelsius: 27.8,
			HumidityPct:        76.0,
			WindSpeedKmh:       9.5,
			WindDirectionDeg:   110.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.12,
			MoonPhase:          "WANING_CRESCENT",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Highway 3077 tourist road corridor; bull elephant charging & roadblock incidents",
			LastUpdated:        now.Add(-50 * time.Minute),
		},
		{
			ID:                 "weather-thai-04",
			StationName:        "Soi Dao Chanthaburi Fruit Station",
			Region:             "Thailand (Chanthaburi / Soi Dao)",
			Latitude:           12.980,
			Longitude:          102.195,
			TemperatureCelsius: 29.0,
			HumidityPct:        88.0,
			WindSpeedKmh:       7.0,
			WindDirectionDeg:   135.0,
			PrecipitationMmH:   4.0,
			MoonIllumination:   0.12,
			MoonPhase:          "WANING_CRESCENT",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Durian & oil palm plantations; heavy elephant herd feeding with electric fence breach risks",
			LastUpdated:        now.Add(-65 * time.Minute),
		},

		// --- INDONESIA (Sumatran Tiger & Elephant Depredation Zones) ---
		{
			ID:                 "weather-sumatra-05",
			StationName:        "Siak Forest Edge Met-Tower 4",
			Region:             "Indonesia (Riau / Siak-Minas)",
			Latitude:           0.520,
			Longitude:          101.450,
			TemperatureCelsius: 28.0,
			HumidityPct:        89.0,
			WindSpeedKmh:       6.0,
			WindDirectionDeg:   180.0,
			PrecipitationMmH:   2.5,
			MoonIllumination:   0.02,
			MoonPhase:          "NEW_MOON",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Cattle pasture & palm estate border; new moon nocturnal stealth ambush peak",
			LastUpdated:        now.Add(-45 * time.Minute),
		},
		{
			ID:                 "weather-sumatra-06",
			StationName:        "Pelalawan Concession Outpost",
			Region:             "Indonesia (Riau / Pelalawan)",
			Latitude:           0.360,
			Longitude:          101.850,
			TemperatureCelsius: 27.5,
			HumidityPct:        91.0,
			WindSpeedKmh:       5.5,
			WindDirectionDeg:   190.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.02,
			MoonPhase:          "NEW_MOON",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "OVERCAST",
			Notes:              "Pulpwood acacia plantation corridor; solitary tiger encounter site near communal pens",
			LastUpdated:        now.Add(-80 * time.Minute),
		},
		{
			ID:                 "weather-sumatra-07",
			StationName:        "Bukit Tigapuluh Buffer Sensor",
			Region:             "Indonesia (Jambi / Tebo)",
			Latitude:           -1.250,
			Longitude:          102.580,
			TemperatureCelsius: 26.8,
			HumidityPct:        93.0,
			WindSpeedKmh:       4.0,
			WindDirectionDeg:   210.0,
			PrecipitationMmH:   3.5,
			MoonIllumination:   0.02,
			MoonPhase:          "NEW_MOON",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Rubber estate buffer; elephant crop raiding and tiger poaching wire snare conflicts",
			LastUpdated:        now.Add(-90 * time.Minute),
		},
		{
			ID:                 "weather-sumatra-08",
			StationName:        "Leuser Ecosystem Aceh Border",
			Region:             "Indonesia (Aceh / Subulussalam)",
			Latitude:           2.650,
			Longitude:          97.980,
			TemperatureCelsius: 25.5,
			HumidityPct:        90.0,
			WindSpeedKmh:       8.0,
			WindDirectionDeg:   270.0,
			PrecipitationMmH:   6.0,
			MoonIllumination:   0.02,
			MoonPhase:          "NEW_MOON",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Fragmented forest neck; elephant migration pathway intersecting village settlement",
			LastUpdated:        now.Add(-110 * time.Minute),
		},

		// --- INDIA (Elephant Migration & Bengal Tiger River Corridor) ---
		{
			ID:                 "weather-india-09",
			StationName:        "Kaziranga Panbari Flood Gauge",
			Region:             "India (Assam / Kaziranga NH-715)",
			Latitude:           26.590,
			Longitude:          93.180,
			TemperatureCelsius: 27.5,
			HumidityPct:        94.0,
			WindSpeedKmh:       18.0,
			WindDirectionDeg:   220.0,
			PrecipitationMmH:   32.0,
			MoonIllumination:   0.88,
			MoonPhase:          "WAXING_GIBBOUS",
			IsFloodedRiverRisk: true,
			WeatherCondition:   "HEAVY_RAIN",
			Notes:              "Severe Brahmaputra flooding (>25mm/h); megafauna forced across NH-715 into Karbi Anglong highlands",
			LastUpdated:        now.Add(-10 * time.Minute),
		},
		{
			ID:                 "weather-india-10",
			StationName:        "Valparai Anamalai Tea Station",
			Region:             "India (Tamil Nadu / Valparai)",
			Latitude:           10.325,
			Longitude:          76.955,
			TemperatureCelsius: 22.0,
			HumidityPct:        88.0,
			WindSpeedKmh:       12.0,
			WindDirectionDeg:   250.0,
			PrecipitationMmH:   12.0,
			MoonIllumination:   0.88,
			MoonPhase:          "WAXING_GIBBOUS",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Tea plantation plateau; resident elephant herd routes passing through worker labour lines",
			LastUpdated:        now.Add(-40 * time.Minute),
		},
		{
			ID:                 "weather-india-11",
			StationName:        "Bandipur Wayanad NH-766 Sensor",
			Region:             "India (Karnataka / Bandipur)",
			Latitude:           11.665,
			Longitude:          76.635,
			TemperatureCelsius: 24.0,
			HumidityPct:        75.0,
			WindSpeedKmh:       10.0,
			WindDirectionDeg:   260.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.88,
			MoonPhase:          "WAXING_GIBBOUS",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Interstate highway NH-766 night traffic corridor; high frequency deer & predator roadkill",
			LastUpdated:        now.Add(-60 * time.Minute),
		},
		{
			ID:                 "weather-india-12",
			StationName:        "Sundarbans Coastal Met Post",
			Region:             "India (West Bengal / Sundarbans)",
			Latitude:           21.949,
			Longitude:          88.899,
			TemperatureCelsius: 29.5,
			HumidityPct:        92.0,
			WindSpeedKmh:       15.0,
			WindDirectionDeg:   160.0,
			PrecipitationMmH:   5.0,
			MoonIllumination:   0.88,
			MoonPhase:          "WAXING_GIBBOUS",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Mangrove tidal channel; fatal tiger ambush risk on honey collectors and crab fishermen",
			LastUpdated:        now.Add(-75 * time.Minute),
		},

		// --- JAPAN (Asian Black Bear & Brown Bear Human Conflict Zones) ---
		{
			ID:                 "weather-japan-13",
			StationName:        "Nagano Valley Agromet Station",
			Region:             "Japan (Nagano / Chikuma River)",
			Latitude:           36.650,
			Longitude:          138.180,
			TemperatureCelsius: 14.2,
			HumidityPct:        65.0,
			WindSpeedKmh:       11.5,
			WindDirectionDeg:   270.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.08,
			MoonPhase:          "NEW_MOON",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Abandoned persimmon (kaki) tree clusters; autumn hyperphagia bear incursions into town fringe",
			LastUpdated:        now.Add(-1 * time.Hour),
		},
		{
			ID:                 "weather-japan-14",
			StationName:        "Morioka Matsuzono Hilltop Sensor",
			Region:             "Japan (Iwate / Morioka)",
			Latitude:           39.735,
			Longitude:          141.165,
			TemperatureCelsius: 11.8,
			HumidityPct:        71.0,
			WindSpeedKmh:       8.0,
			WindDirectionDeg:   315.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.15,
			MoonPhase:          "WAXING_CRESCENT",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "OVERCAST",
			Notes:              "Residential housing edge & stream greenbelt; morning walker & elderly injury danger zone",
			LastUpdated:        now.Add(-2 * time.Hour),
		},
		{
			ID:                 "weather-japan-15",
			StationName:        "Kazuno Hachimantai Bear Station",
			Region:             "Japan (Akita / Kazuno)",
			Latitude:           40.215,
			Longitude:          140.785,
			TemperatureCelsius: 10.5,
			HumidityPct:        74.0,
			WindSpeedKmh:       9.0,
			WindDirectionDeg:   300.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.15,
			MoonPhase:          "WAXING_CRESCENT",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "OVERCAST",
			Notes:              "Beech mast failure epicentre; bamboo shoot foragers fatal attack site in mountain passes",
			LastUpdated:        now.Add(-2 * time.Hour),
		},
		{
			ID:                 "weather-japan-16",
			StationName:        "Aizu Basin Foothill Sensor",
			Region:             "Japan (Fukushima / Aizu)",
			Latitude:           37.520,
			Longitude:          139.930,
			TemperatureCelsius: 13.0,
			HumidityPct:        68.0,
			WindSpeedKmh:       7.5,
			WindDirectionDeg:   280.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.08,
			MoonPhase:          "NEW_MOON",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Onsen resort perimeter & soba fields; bear foraging along tourist hiking access routes",
			LastUpdated:        now.Add(-3 * time.Hour),
		},
		{
			ID:                 "weather-japan-17",
			StationName:        "Shiretoko Rausu Coastal Tower",
			Region:             "Japan (Hokkaido / Shiretoko)",
			Latitude:           44.020,
			Longitude:          145.190,
			TemperatureCelsius: 9.5,
			HumidityPct:        78.0,
			WindSpeedKmh:       16.0,
			WindDirectionDeg:   45.0,
			PrecipitationMmH:   1.0,
			MoonIllumination:   0.50,
			MoonPhase:          "FIRST_QUARTER",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "OVERCAST",
			Notes:              "Coastal salmon river mouth; brown bear foraging along Rausu-Shari transit pass",
			LastUpdated:        now.Add(-3 * time.Hour),
		},
		{
			ID:                 "weather-japan-18",
			StationName:        "Sapporo Minami Hillside Sensor",
			Region:             "Japan (Hokkaido / Sapporo)",
			Latitude:           43.005,
			Longitude:          141.310,
			TemperatureCelsius: 12.0,
			HumidityPct:        66.0,
			WindSpeedKmh:       8.5,
			WindDirectionDeg:   330.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.50,
			MoonPhase:          "FIRST_QUARTER",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Suburban foothills & school route; brown bear sightings near Toyohira river corridor",
			LastUpdated:        now.Add(-4 * time.Hour),
		},
		{
			ID:                 "weather-japan-19",
			StationName:        "Eniwa Agriculture Sweetcorn Belt",
			Region:             "Japan (Hokkaido / Eniwa)",
			Latitude:           42.920,
			Longitude:          141.580,
			TemperatureCelsius: 11.2,
			HumidityPct:        70.0,
			WindSpeedKmh:       10.0,
			WindDirectionDeg:   340.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.50,
			MoonPhase:          "FIRST_QUARTER",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "High-calorie sweetcorn & sugarbeet fields; massive pre-denning crop damage zones",
			LastUpdated:        now.Add(-5 * time.Hour),
		},
		{
			ID:                 "weather-japan-20",
			StationName:        "Shumarinai Lake Basin Tower",
			Region:             "Japan (Hokkaido / Horokanai)",
			Latitude:           44.290,
			Longitude:          142.170,
			TemperatureCelsius: 7.8,
			HumidityPct:        82.0,
			WindSpeedKmh:       12.0,
			WindDirectionDeg:   10.0,
			PrecipitationMmH:   0.5,
			MoonIllumination:   0.50,
			MoonPhase:          "FIRST_QUARTER",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "OVERCAST",
			Notes:              "Wild lake peninsula; fatal angler predatory attack site & remote forestry frontier",
			LastUpdated:        now.Add(-5 * time.Hour),
		},

		// --- ADJACENT & SIMILAR CONDITIONS EXPANSION (JAPAC Expansion 21 - 40) ---
		// Thailand Adjacent Corridors (Dong Phayayen East, Eastern 5 Provinces, Tenasserim South & North)
		{
			ID:                 "weather-thai-21",
			StationName:        "Thap Lan Hwy 304 Wildlife Corridor",
			Region:             "Thailand (Prachinburi / Thap Lan)",
			Latitude:           14.215,
			Longitude:          101.930,
			TemperatureCelsius: 28.5,
			HumidityPct:        79.0,
			WindSpeedKmh:       10.5,
			WindDirectionDeg:   120.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.12,
			MoonPhase:          "WANING_CRESCENT",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Adjacent to Khao Yai; Highway 304 wildlife ecoduct bridge corridor with frequent gaur & elephant vehicular hazard",
			LastUpdated:        now.Add(-30 * time.Minute),
		},
		{
			ID:                 "weather-thai-22",
			StationName:        "Khao Ang Rue Nai Tha Takiap Outpost",
			Region:             "Thailand (Chachoengsao / Tha Takiap)",
			Latitude:           13.385,
			Longitude:          101.950,
			TemperatureCelsius: 29.8,
			HumidityPct:        84.0,
			WindSpeedKmh:       8.5,
			WindDirectionDeg:   140.0,
			PrecipitationMmH:   2.0,
			MoonIllumination:   0.12,
			MoonPhase:          "WANING_CRESCENT",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Adjacent to Chanthaburi; Eastern Forest Complex core with severe elephant crop raiding across cassava & oil palm estates",
			LastUpdated:        now.Add(-40 * time.Minute),
		},
		{
			ID:                 "weather-thai-23",
			StationName:        "Huai Yang Coastal Mountain Pass",
			Region:             "Thailand (Prachuap / Thap Sakae)",
			Latitude:           11.605,
			Longitude:          99.660,
			TemperatureCelsius: 30.5,
			HumidityPct:        81.0,
			WindSpeedKmh:       13.0,
			WindDirectionDeg:   95.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.12,
			MoonPhase:          "WANING_CRESCENT",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Adjacent south of Kui Buri; narrow Isthmus corridor between Tenasserim mountain ridge and coconut coastal farms",
			LastUpdated:        now.Add(-45 * time.Minute),
		},
		{
			ID:                 "weather-thai-24",
			StationName:        "Suan Phueng Phachi River Frontier",
			Region:             "Thailand (Ratchaburi / Suan Phueng)",
			Latitude:           13.545,
			Longitude:          99.310,
			TemperatureCelsius: 27.2,
			HumidityPct:        83.0,
			WindSpeedKmh:       9.0,
			WindDirectionDeg:   100.0,
			PrecipitationMmH:   1.0,
			MoonIllumination:   0.12,
			MoonPhase:          "WANING_CRESCENT",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "OVERCAST",
			Notes:              "Adjacent north of Kaeng Krachan; Myanmar border montane forest; bamboo shoot & pineapple raiding herds",
			LastUpdated:        now.Add(-55 * time.Minute),
		},

		// Indonesia Sumatra Adjacent Corridors (Tesso Nilo Buffer, Giam Siak Peatland, Leuser East Langkat, East Aceh)
		{
			ID:                 "weather-sumatra-25",
			StationName:        "Tesso Nilo North Buffer Station",
			Region:             "Indonesia (Riau / Tesso Nilo)",
			Latitude:           0.120,
			Longitude:          101.620,
			TemperatureCelsius: 28.6,
			HumidityPct:        92.0,
			WindSpeedKmh:       5.0,
			WindDirectionDeg:   175.0,
			PrecipitationMmH:   3.0,
			MoonIllumination:   0.02,
			MoonPhase:          "NEW_MOON",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Adjacent to Pelalawan; fragmented lowland rain forest and encroached palm landscape with frequent elephant troop encounters",
			LastUpdated:        now.Add(-50 * time.Minute),
		},
		{
			ID:                 "weather-sumatra-26",
			StationName:        "Giam Siak Peatland Bio-Observatory",
			Region:             "Indonesia (Riau / Bengkalis)",
			Latitude:           1.180,
			Longitude:          101.880,
			TemperatureCelsius: 27.8,
			HumidityPct:        94.0,
			WindSpeedKmh:       6.5,
			WindDirectionDeg:   165.0,
			PrecipitationMmH:   1.5,
			MoonIllumination:   0.02,
			MoonPhase:          "NEW_MOON",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Adjacent north of Siak; peat swamp forest biosphere boundary; tiger livestock depredation around drainage canals",
			LastUpdated:        now.Add(-60 * time.Minute),
		},
		{
			ID:                 "weather-sumatra-27",
			StationName:        "Langkat Leuser Foothill Monitor",
			Region:             "Indonesia (North Sumatra / Langkat)",
			Latitude:           3.580,
			Longitude:          98.220,
			TemperatureCelsius: 26.2,
			HumidityPct:        88.0,
			WindSpeedKmh:       7.0,
			WindDirectionDeg:   260.0,
			PrecipitationMmH:   4.5,
			MoonIllumination:   0.02,
			MoonPhase:          "NEW_MOON",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Adjacent east slope of Gunung Leuser; palm concession border; orangutan and elephant trans-boundary herd routes",
			LastUpdated:        now.Add(-70 * time.Minute),
		},
		{
			ID:                 "weather-sumatra-28",
			StationName:        "Peureulak River Conflict Zone",
			Region:             "Indonesia (Aceh / Aceh Timur)",
			Latitude:           4.850,
			Longitude:          97.880,
			TemperatureCelsius: 26.8,
			HumidityPct:        87.0,
			WindSpeedKmh:       8.0,
			WindDirectionDeg:   280.0,
			PrecipitationMmH:   5.0,
			MoonIllumination:   0.02,
			MoonPhase:          "NEW_MOON",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Connected to Leuser Aceh corridor; highway & settlement boundary; frequent high-stress elephant encounters",
			LastUpdated:        now.Add(-80 * time.Minute),
		},

		// India Adjacent Corridors (Kaziranga South Karbi Foothills, Nagarhole Kabini, Wayanad Coffee Belt, Dooars Tea)
		{
			ID:                 "weather-india-29",
			StationName:        "Karbi Anglong South Foothill Ridge",
			Region:             "India (Assam / Karbi Anglong)",
			Latitude:           26.510,
			Longitude:          93.220,
			TemperatureCelsius: 26.5,
			HumidityPct:        91.0,
			WindSpeedKmh:       14.0,
			WindDirectionDeg:   210.0,
			PrecipitationMmH:   26.0,
			MoonIllumination:   0.88,
			MoonPhase:          "WAXING_GIBBOUS",
			IsFloodedRiverRisk: true,
			WeatherCondition:   "HEAVY_RAIN",
			Notes:              "Directly adjacent south of Kaziranga NH-715; critical animal escape corridor when Brahmaputra submerges national park grasslands",
			LastUpdated:        now.Add(-30 * time.Minute),
		},
		{
			ID:                 "weather-india-30",
			StationName:        "Nagarhole Kabini Reservoir Edge",
			Region:             "India (Karnataka / Nagarhole)",
			Latitude:           11.950,
			Longitude:          76.280,
			TemperatureCelsius: 25.0,
			HumidityPct:        78.0,
			WindSpeedKmh:       11.0,
			WindDirectionDeg:   255.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.88,
			MoonPhase:          "WAXING_GIBBOUS",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Adjacent west of Bandipur; high-density tiger & elephant concentration along Kabini backwaters and agricultural fringe",
			LastUpdated:        now.Add(-40 * time.Minute),
		},
		{
			ID:                 "weather-india-31",
			StationName:        "Wayanad Sultan Bathery Agro-Forest",
			Region:             "India (Kerala / Wayanad)",
			Latitude:           11.660,
			Longitude:          76.260,
			TemperatureCelsius: 23.5,
			HumidityPct:        86.0,
			WindSpeedKmh:       9.5,
			WindDirectionDeg:   240.0,
			PrecipitationMmH:   8.0,
			MoonIllumination:   0.88,
			MoonPhase:          "WAXING_GIBBOUS",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Adjacent south of Bandipur on Kerala side; coffee and pepper plantation matrix with persistent night elephant incursions",
			LastUpdated:        now.Add(-50 * time.Minute),
		},
		{
			ID:                 "weather-india-32",
			StationName:        "Dooars Gorumara Tea Plains",
			Region:             "India (West Bengal / Dooars)",
			Latitude:           26.750,
			Longitude:          88.790,
			TemperatureCelsius: 25.8,
			HumidityPct:        89.0,
			WindSpeedKmh:       12.0,
			WindDirectionDeg:   150.0,
			PrecipitationMmH:   10.0,
			MoonIllumination:   0.88,
			MoonPhase:          "WAXING_GIBBOUS",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "RAIN",
			Notes:              "Similar tea-estate landscape to Valparai; railway corridor elephant collision hotspot and leopard day-bedding in tea bushes",
			LastUpdated:        now.Add(-60 * time.Minute),
		},

		// Japan Adjacent & Similar Topography (Asama Foothill Karuizawa, Minamiuonuma, Hanamaki Kitakami, Sendai Hills, Kitaakita Moriyoshi, Yonezawa Basin, Shari Utoro, Otaru Coast)
		{
			ID:                 "weather-japan-33",
			StationName:        "Karuizawa Asama Foothill Sensor",
			Region:             "Japan (Nagano / Karuizawa)",
			Latitude:           36.350,
			Longitude:          138.600,
			TemperatureCelsius: 12.5,
			HumidityPct:        69.0,
			WindSpeedKmh:       10.0,
			WindDirectionDeg:   290.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.08,
			MoonPhase:          "NEW_MOON",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Adjacent east of Nagano Chikuma; Mt Asama resort villas and urban perimeter; bear garbage and garden fruit attractants",
			LastUpdated:        now.Add(-1 * time.Hour),
		},
		{
			ID:                 "weather-japan-34",
			StationName:        "Minamiuonuma Valley Base Station",
			Region:             "Japan (Niigata / Minamiuonuma)",
			Latitude:           37.060,
			Longitude:          138.870,
			TemperatureCelsius: 13.8,
			HumidityPct:        73.0,
			WindSpeedKmh:       8.0,
			WindDirectionDeg:   310.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.08,
			MoonPhase:          "NEW_MOON",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "OVERCAST",
			Notes:              "Adjacent northeast of Nagano; heavy snowfall mountain basin; Kan-Etsu corridor and settlement fruit trees attraction",
			LastUpdated:        now.Add(-1 * time.Hour),
		},
		{
			ID:                 "weather-japan-35",
			StationName:        "Hanamaki Kitakami River Corridor",
			Region:             "Japan (Iwate / Hanamaki)",
			Latitude:           39.390,
			Longitude:          141.120,
			TemperatureCelsius: 12.2,
			HumidityPct:        69.0,
			WindSpeedKmh:       7.5,
			WindDirectionDeg:   320.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.15,
			MoonPhase:          "WAXING_CRESCENT",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Adjacent south of Morioka; Kitakami river riparian greenbelt providing direct bear travel corridor into suburban apple orchards",
			LastUpdated:        now.Add(-2 * time.Hour),
		},
		{
			ID:                 "weather-japan-36",
			StationName:        "Sendai Aobayama Urban Forest Rim",
			Region:             "Japan (Miyagi / Sendai)",
			Latitude:           38.260,
			Longitude:          140.840,
			TemperatureCelsius: 14.5,
			HumidityPct:        65.0,
			WindSpeedKmh:       9.0,
			WindDirectionDeg:   290.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.15,
			MoonPhase:          "WAXING_CRESCENT",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Similar large metropolitan hill margin; university campus and residential hilltops adjacent to Hirose River gorge",
			LastUpdated:        now.Add(-2 * time.Hour),
		},
		{
			ID:                 "weather-japan-37",
			StationName:        "Kitaakita Moriyoshi Mountain Gate",
			Region:             "Japan (Akita / Kitaakita)",
			Latitude:           40.180,
			Longitude:          140.370,
			TemperatureCelsius: 11.0,
			HumidityPct:        76.0,
			WindSpeedKmh:       8.5,
			WindDirectionDeg:   295.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.15,
			MoonPhase:          "WAXING_CRESCENT",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "OVERCAST",
			Notes:              "Adjacent west of Kazuno; Mt Moriyoshi beech zone; consecutive town centre intrusions and elderly pedestrian attacks",
			LastUpdated:        now.Add(-2 * time.Hour),
		},
		{
			ID:                 "weather-japan-38",
			StationName:        "Yonezawa Foothill Orchard Tower",
			Region:             "Japan (Yamagata / Yonezawa)",
			Latitude:           37.910,
			Longitude:          140.110,
			TemperatureCelsius: 13.5,
			HumidityPct:        67.0,
			WindSpeedKmh:       7.0,
			WindDirectionDeg:   270.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.08,
			MoonPhase:          "NEW_MOON",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Adjacent north of Aizu Basin; Ou mountain western foothills; La France pear & apple orchards vulnerable to hyperphagia foraging",
			LastUpdated:        now.Add(-3 * time.Hour),
		},
		{
			ID:                 "weather-japan-39",
			StationName:        "Shari Utoro Coastal Salmon Stream",
			Region:             "Japan (Hokkaido / Shari Utoro)",
			Latitude:           44.070,
			Longitude:          144.990,
			TemperatureCelsius: 10.0,
			HumidityPct:        75.0,
			WindSpeedKmh:       15.0,
			WindDirectionDeg:   350.0,
			PrecipitationMmH:   0.5,
			MoonIllumination:   0.50,
			MoonPhase:          "FIRST_QUARTER",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "OVERCAST",
			Notes:              "Adjacent west coast of Shiretoko peninsula across pass from Rausu; salmon fishery and Oshinkoshin tourist trail encounters",
			LastUpdated:        now.Add(-3 * time.Hour),
		},
		{
			ID:                 "weather-japan-40",
			StationName:        "Otaru Zenibako Seaside Hillside",
			Region:             "Japan (Hokkaido / Otaru)",
			Latitude:           43.140,
			Longitude:          141.170,
			TemperatureCelsius: 11.5,
			HumidityPct:        68.0,
			WindSpeedKmh:       11.0,
			WindDirectionDeg:   320.0,
			PrecipitationMmH:   0.0,
			MoonIllumination:   0.50,
			MoonPhase:          "FIRST_QUARTER",
			IsFloodedRiverRisk: false,
			WeatherCondition:   "CLEAR",
			Notes:              "Adjacent west of Sapporo Teine/Minami mountains; steep coastal bluff residences with fruit trees and garbage foraging",
			LastUpdated:        now.Add(-4 * time.Hour),
		},
	}
	for _, w := range weatherList {
		s.weather[w.ID] = w
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

func scalePolygonCoords(coords []Coordinate, factor float64) []Coordinate {
	if len(coords) == 0 {
		return nil
	}
	var sumLat, sumLng float64
	for _, c := range coords {
		sumLat += c.Lat
		sumLng += c.Lng
	}
	cLat := sumLat / float64(len(coords))
	cLng := sumLng / float64(len(coords))
	res := make([]Coordinate, len(coords))
	for i, c := range coords {
		res[i] = Coordinate{
			Lat: cLat + (c.Lat-cLat)*factor,
			Lng: cLng + (c.Lng-cLng)*factor,
		}
	}
	return res
}

func (s *Store) upsertGraduatedBuffersLocked(core AdminGeofence, now time.Time) {
	if strings.HasSuffix(core.ID, "-tier2-buffer") || strings.HasSuffix(core.ID, "-tier1-buffer") {
		return
	}

	if core.ThreatLevel == "TIER_3_CRITICAL" || core.ThreatLevel == "TIER_3_DANGER" {
		// Tier 2 Buffer (1.75x)
		tier2ID := core.ID + "-tier2-buffer"
		t2 := AdminGeofence{
			ID:            tier2ID,
			Name:          core.Name + " (Tier 2 Warning Buffer)",
			Source:        core.Source,
			IsPersistent:  core.IsPersistent,
			GeometryType:  core.GeometryType,
			ThreatLevel:   "TIER_2_ALERT",
			TargetSpecies: core.TargetSpecies,
			CreatedAt:     now,
			Reason:        fmt.Sprintf("Graduated Tier 2 warning buffer surrounding active hazard zone (%s)", core.ID),
		}
		if core.GeometryType == "circle" && core.CircleCenter != nil {
			center := *core.CircleCenter
			t2.CircleCenter = &center
			t2.CircleRadiusMeters = core.CircleRadiusMeters * 1.75
		} else if len(core.Coordinates) > 0 {
			t2.Coordinates = scalePolygonCoords(core.Coordinates, 1.75)
		}
		s.geofences[tier2ID] = t2

		// Tier 1 Buffer (2.60x)
		tier1ID := core.ID + "-tier1-buffer"
		t1 := AdminGeofence{
			ID:            tier1ID,
			Name:          core.Name + " (Tier 1 Advisory Buffer)",
			Source:        core.Source,
			IsPersistent:  core.IsPersistent,
			GeometryType:  core.GeometryType,
			ThreatLevel:   "TIER_1_CAUTION",
			TargetSpecies: core.TargetSpecies,
			CreatedAt:     now,
			Reason:        fmt.Sprintf("Graduated Tier 1 advisory buffer surrounding active hazard zone (%s)", core.ID),
		}
		if core.GeometryType == "circle" && core.CircleCenter != nil {
			center := *core.CircleCenter
			t1.CircleCenter = &center
			t1.CircleRadiusMeters = core.CircleRadiusMeters * 2.60
		} else if len(core.Coordinates) > 0 {
			t1.Coordinates = scalePolygonCoords(core.Coordinates, 2.60)
		}
		s.geofences[tier1ID] = t1

	} else if core.ThreatLevel == "TIER_2_ALERT" || core.ThreatLevel == "TIER_2_WARNING" {
		// Tier 1 Buffer (1.75x)
		tier1ID := core.ID + "-tier1-buffer"
		t1 := AdminGeofence{
			ID:            tier1ID,
			Name:          core.Name + " (Tier 1 Advisory Buffer)",
			Source:        core.Source,
			IsPersistent:  core.IsPersistent,
			GeometryType:  core.GeometryType,
			ThreatLevel:   "TIER_1_CAUTION",
			TargetSpecies: core.TargetSpecies,
			CreatedAt:     now,
			Reason:        fmt.Sprintf("Graduated Tier 1 advisory buffer surrounding active hazard zone (%s)", core.ID),
		}
		if core.GeometryType == "circle" && core.CircleCenter != nil {
			center := *core.CircleCenter
			t1.CircleCenter = &center
			t1.CircleRadiusMeters = core.CircleRadiusMeters * 1.75
		} else if len(core.Coordinates) > 0 {
			t1.Coordinates = scalePolygonCoords(core.Coordinates, 1.75)
		}
		s.geofences[tier1ID] = t1
	}
}

func (s *Store) CreateGeofence(g AdminGeofence) AdminGeofence {
	s.mu.Lock()
	if g.ID == "" {
		g.ID = fmt.Sprintf("geo-manual-%d", time.Now().UnixNano()%100000)
	}
	if g.Source == "" {
		g.Source = "manual"
	}
	if g.CreatedAt.IsZero() {
		g.CreatedAt = time.Now().UTC()
	}

	var lat, lng float64
	if g.CircleCenter != nil {
		lat = g.CircleCenter.Lat
		lng = g.CircleCenter.Lng
	} else if len(g.Coordinates) > 0 {
		lat = g.Coordinates[0].Lat
		lng = g.Coordinates[0].Lng
	}
	g.Name = EnhanceGeofenceNameWithLocation(g.Name, g.TargetSpecies, lat, lng)

	s.geofences[g.ID] = g
	s.upsertGraduatedBuffersLocked(g, g.CreatedAt)
	s.mu.Unlock()

	s.persistGeofence(g)
	return g
}

func (s *Store) DeleteGeofence(id string) bool {
	s.mu.Lock()
	if _, exists := s.geofences[id]; !exists {
		s.mu.Unlock()
		return false
	}
	delete(s.geofences, id)
	t2ID := id + "-tier2-buffer"
	t1ID := id + "-tier1-buffer"
	hasT2 := false
	hasT1 := false
	if _, exists := s.geofences[t2ID]; exists {
		delete(s.geofences, t2ID)
		hasT2 = true
	}
	if _, exists := s.geofences[t1ID]; exists {
		delete(s.geofences, t1ID)
		hasT1 = true
	}
	s.mu.Unlock()

	s.persistDeleteGeofence(id)
	if hasT2 {
		s.persistDeleteGeofence(t2ID)
	}
	if hasT1 {
		s.persistDeleteGeofence(t1ID)
	}
	return true
}

func (s *Store) ClearAllManualGeofences() int {
	s.mu.Lock()
	deletedCount := 0
	toDelete := make([]string, 0)
	for id, g := range s.geofences {
		if g.Source == "manual" {
			delete(s.geofences, id)
			deletedCount++
			toDelete = append(toDelete, id)
		}
	}
	s.mu.Unlock()

	for _, id := range toDelete {
		s.persistDeleteGeofence(id)
	}
	return deletedCount
}

func haversineDistanceKm(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(math.Pi/180.0))*math.Cos(lat2*(math.Pi/180.0))*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return 6371.0 * c
}

func extractRegionLabel(region, road, desc, fallback string) string {
	candidates := []string{region, road, desc}
	for _, c := range candidates {
		lower := strings.ToLower(c)
		if strings.Contains(lower, "nara") || strings.Contains(c, "奈良") {
			return "Nara"
		}
		if strings.Contains(lower, "nagano") || strings.Contains(c, "長野") {
			return "Nagano"
		}
		if strings.Contains(lower, "iwate") || strings.Contains(c, "岩手") || strings.Contains(lower, "morioka") || strings.Contains(c, "盛岡") {
			return "Morioka/Iwate"
		}
		if strings.Contains(lower, "hokkaido") || strings.Contains(c, "北海道") || strings.Contains(lower, "shiretoko") {
			return "Hokkaido"
		}
		if strings.Contains(lower, "thai") || strings.Contains(lower, "phetchaburi") {
			return "Thailand"
		}
		if strings.Contains(lower, "sumatra") || strings.Contains(lower, "riau") {
			return "Sumatra"
		}
	}
	if region != "" {
		parts := strings.Split(region, ",")
		return strings.TrimSpace(parts[0])
	}
	return fallback
}

func getEnvCenterLat(env AdminEnvironmentalData) float64 {
	if env.CircleCenter != nil && env.CircleCenter.Lat != 0 {
		return env.CircleCenter.Lat
	}
	if len(env.Coordinates) > 0 {
		var sum float64
		for _, c := range env.Coordinates {
			sum += c.Lat
		}
		return sum / float64(len(env.Coordinates))
	}
	return 0
}

func getEnvCenterLng(env AdminEnvironmentalData) float64 {
	if env.CircleCenter != nil && env.CircleCenter.Lng != 0 {
		return env.CircleCenter.Lng
	}
	if len(env.Coordinates) > 0 {
		var sum float64
		for _, c := range env.Coordinates {
			sum += c.Lng
		}
		return sum / float64(len(env.Coordinates))
	}
	return 0
}

func (s *Store) RecalculateAiGeofences() (int, time.Time, bool) {
	s.mu.Lock()
	if s.isRecalculating {
		s.mu.Unlock()
		return 0, time.Time{}, true // busy / conflict
	}
	s.isRecalculating = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.isRecalculating = false
		s.mu.Unlock()
	}()

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()

	// 1. Refresh existing baseline AI geofences
	for id, g := range s.geofences {
		if g.Source == "ai" {
			g.CreatedAt = now
			s.geofences[id] = g
		}
	}

	// 2. Correlate Wildlife Incidents with Environmental Attractants & Generate Dynamic AI Geofences
	correlatedWildlife := make(map[string]bool)

	// Scan Bear, Elephant, Tiger sightings against attractants
	for wID, w := range s.wildlife {
		if w.Status == "RESOLVED" || w.Status == "FALSE_ALARM" {
			continue
		}
		if w.Location.Lat == 0 && w.Location.Lng == 0 {
			continue
		}

		isBear := strings.Contains(strings.ToLower(w.Species), "bear") || strings.Contains(w.Species, "クマ")
		isElephant := strings.Contains(strings.ToLower(w.Species), "elephant") || strings.Contains(w.Species, "ゾウ")
		isTiger := strings.Contains(strings.ToLower(w.Species), "tiger") || strings.Contains(w.Species, "トラ")

		for eID, env := range s.environmental {
			envLat := getEnvCenterLat(env)
			envLng := getEnvCenterLng(env)
			if envLat == 0 && envLng == 0 {
				continue
			}

			distKm := haversineDistanceKm(w.Location.Lat, w.Location.Lng, envLat, envLng)

			// Rule 1: Bear + Persimmon (Kaki) within 10 km
			isKaki := env.Category == "PERSIMMON_KAKI" ||
				strings.Contains(strings.ToLower(env.Name), "kaki") ||
				strings.Contains(strings.ToLower(env.Name), "persimmon") ||
				strings.Contains(strings.ToLower(env.LocationDesc), "kaki") ||
				strings.Contains(env.Name, "柿") ||
				strings.Contains(env.LocationDesc, "柿")

			if isBear && isKaki && distKm <= 10.0 {
				correlatedWildlife[wID] = true
				geoID := fmt.Sprintf("geo-ai-corr-%s-%s", wID, eID)
				centerLat := (w.Location.Lat + envLat) / 2.0
				centerLng := (w.Location.Lng + envLng) / 2.0
				radiusM := (distKm*1000.0)/2.0 + 1500.0
				if radiusM < 1500.0 {
					radiusM = 1500.0
				}
				if radiusM > 4500.0 {
					radiusM = 4500.0
				}

				regionLabel := extractRegionLabel(w.Location.Region, w.Location.RoadName, w.Description, "Nara")

				centerCoord := Coordinate{Lat: centerLat, Lng: centerLng}
				s.geofences[geoID] = AdminGeofence{
					ID:                 geoID,
					Name:               fmt.Sprintf("AI Dynamic Shield: %s Bear-Persimmon Hazard Zone", regionLabel),
					Source:             "ai",
					IsPersistent:       false,
					GeometryType:       "circle",
					CircleCenter:       &centerCoord,
					CircleRadiusMeters: radiusM,
					ThreatLevel:        "TIER_3_CRITICAL",
					TargetSpecies:      w.Species,
					CreatedAt:          now,
					Reason:             fmt.Sprintf("AI Spatial Correlation: %s sighted within %.1fkm of unharvested persimmon (kaki) attractant '%s'. Autumn hyperphagia incursion alert.", w.Species, distKm, env.Name),
				}
			}

			// Rule 2: Elephant + Crop Phenology (Sugarcane/Fruit) within 10 km
			isCrop := env.Category == "CROP_PHENOLOGY" ||
				strings.Contains(strings.ToLower(env.Name), "sugar") ||
				strings.Contains(strings.ToLower(env.Name), "crop") ||
				strings.Contains(strings.ToLower(env.Name), "farm")
			if isElephant && isCrop && distKm <= 10.0 {
				correlatedWildlife[wID] = true
				geoID := fmt.Sprintf("geo-ai-corr-%s-%s", wID, eID)
				centerCoord := Coordinate{Lat: (w.Location.Lat + envLat) / 2.0, Lng: (w.Location.Lng + envLng) / 2.0}
				radiusM := (distKm*1000.0)/2.0 + 2000.0
				if radiusM > 5000.0 {
					radiusM = 5000.0
				}
				s.geofences[geoID] = AdminGeofence{
					ID:                 geoID,
					Name:               fmt.Sprintf("AI Dynamic Shield: %s Crop-Raiding Corridor", env.Name),
					Source:             "ai",
					IsPersistent:       false,
					GeometryType:       "circle",
					CircleCenter:       &centerCoord,
					CircleRadiusMeters: radiusM,
					ThreatLevel:        "TIER_3_CRITICAL",
					TargetSpecies:      w.Species,
					CreatedAt:          now,
					Reason:             fmt.Sprintf("AI Spatial Correlation: %s breeding herd moving towards ripening crop cluster '%s' (%.1fkm).", w.Species, env.Name, distKm),
				}
			}

			// Rule 3: Tiger + Livestock Pasture within 10 km
			isPasture := env.Category == "LIVESTOCK_PASTURE" ||
				strings.Contains(strings.ToLower(env.Name), "pasture") ||
				strings.Contains(strings.ToLower(env.Name), "livestock") ||
				strings.Contains(strings.ToLower(env.Name), "cattle")
			if isTiger && isPasture && distKm <= 10.0 {
				correlatedWildlife[wID] = true
				geoID := fmt.Sprintf("geo-ai-corr-%s-%s", wID, eID)
				centerCoord := Coordinate{Lat: (w.Location.Lat + envLat) / 2.0, Lng: (w.Location.Lng + envLng) / 2.0}
				radiusM := (distKm*1000.0)/2.0 + 2500.0
				if radiusM > 5500.0 {
					radiusM = 5500.0
				}
				s.geofences[geoID] = AdminGeofence{
					ID:                 geoID,
					Name:               fmt.Sprintf("AI Dynamic Shield: %s Depredation Perimeter", env.Name),
					Source:             "ai",
					IsPersistent:       false,
					GeometryType:       "circle",
					CircleCenter:       &centerCoord,
					CircleRadiusMeters: radiusM,
					ThreatLevel:        "TIER_3_CRITICAL",
					TargetSpecies:      w.Species,
					CreatedAt:          now,
					Reason:             fmt.Sprintf("AI Spatial Correlation: %s detected within %.1fkm of unfenced livestock '%s'. Nighttime predatory ambush alert.", w.Species, distKm, env.Name),
				}
			}
		}
	}

	// 3. Standalone verified sightings without matched attractant get dynamic alert zones
	for wID, w := range s.wildlife {
		if correlatedWildlife[wID] || w.Status == "RESOLVED" || w.Status == "FALSE_ALARM" {
			continue
		}
		if w.Location.Lat == 0 && w.Location.Lng == 0 {
			continue
		}
		// Sighting within last 7 days gets active geofence
		if now.Sub(w.ReportedAt) < 7*24*time.Hour || w.ReportedAt.IsZero() {
			geoID := fmt.Sprintf("geo-ai-sighting-%s", wID)
			centerCoord := Coordinate{Lat: w.Location.Lat, Lng: w.Location.Lng}
			tier := "TIER_2_ALERT"
			if w.ThreatLevel == "TIER_3_CRITICAL" {
				tier = "TIER_3_CRITICAL"
			}
			locationName := w.Location.RoadName
			if locationName == "" || strings.HasPrefix(locationName, "Geocoded Spot") {
				locationName = w.Location.Region
			}
			if locationName == "" {
				locationName = ReverseGeocode(w.Location.Lat, w.Location.Lng)
			}
			s.geofences[geoID] = AdminGeofence{
				ID:                 geoID,
				Name:               fmt.Sprintf("AI Alert: %s Sighting Buffer (%s)", w.Species, locationName),
				Source:             "ai",
				IsPersistent:       false,
				GeometryType:       "circle",
				CircleCenter:       &centerCoord,
				CircleRadiusMeters: 1800.0,
				ThreatLevel:        tier,
				TargetSpecies:      w.Species,
				CreatedAt:          now,
				Reason:             fmt.Sprintf("Active verified sighting reported by %s: %s", w.Reporter, w.Description),
			}
		}
	}

	// 4. Roadkills pending removal generate carcass scent plume alert zones
	for rkID, rk := range s.roadkills {
		if rk.Status == "PENDING_REMOVAL" && !rk.IsExcluded && (rk.Location.Lat != 0 || rk.Location.Lng != 0) {
			geoID := fmt.Sprintf("geo-ai-roadkill-%s", rkID)
			centerCoord := Coordinate{Lat: rk.Location.Lat, Lng: rk.Location.Lng}
			roadName := rk.Location.RoadName
			if roadName == "" || strings.HasPrefix(roadName, "Geocoded Spot") {
				roadName = rk.Location.Region
			}
			if roadName == "" {
				roadName = ReverseGeocode(rk.Location.Lat, rk.Location.Lng)
			}
			s.geofences[geoID] = AdminGeofence{
				ID:                 geoID,
				Name:               fmt.Sprintf("AI Carcass Scent Buffer: %s (%s)", rk.Species, roadName),
				Source:             "ai",
				IsPersistent:       false,
				GeometryType:       "circle",
				CircleCenter:       &centerCoord,
				CircleRadiusMeters: 2000.0,
				ThreatLevel:        "TIER_2_ALERT",
				TargetSpecies:      "Apex Predator / Scavenger",
				CreatedAt:          now,
				Reason:             fmt.Sprintf("Uncollected %s carcass (%s decay, %.0fkg). Predicted apex predator attraction radius.", rk.Species, rk.DecayStage, rk.EstimatedWeightKg),
			}
		}
	}

	// 4.5. Generate graduated concentric buffers (Tier 2 / Tier 1) for AI hazard zones
	var aiCoreFences []AdminGeofence
	for _, g := range s.geofences {
		if g.Source == "ai" && !strings.HasSuffix(g.ID, "-tier2-buffer") && !strings.HasSuffix(g.ID, "-tier1-buffer") {
			aiCoreFences = append(aiCoreFences, g)
		}
	}
	for _, g := range aiCoreFences {
		s.upsertGraduatedBuffersLocked(g, now)
	}

	// 5. Count total AI geofences
	aiCount := 0
	for _, g := range s.geofences {
		if g.Source == "ai" {
			aiCount++
		}
	}

	return aiCount, now, false
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
	s.mu.Unlock()

	s.persistWildlife(w)
	return w
}

func (s *Store) DeleteWildlife(id string) bool {
	s.mu.Lock()
	if _, exists := s.wildlife[id]; !exists {
		s.mu.Unlock()
		return false
	}
	delete(s.wildlife, id)
	s.mu.Unlock()

	s.persistDeleteWildlife(id)
	return true
}

func (s *Store) UpdateWildlifeStatus(id string, status string) (AdminWildlifeIncident, bool) {
	s.mu.Lock()
	w, exists := s.wildlife[id]
	if !exists {
		s.mu.Unlock()
		return AdminWildlifeIncident{}, false
	}
	w.Status = status
	s.wildlife[id] = w
	s.mu.Unlock()

	s.persistWildlife(w)
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
	s.mu.Unlock()

	s.persistRoadkill(rk)
	return rk
}

func (s *Store) SetRoadkillExclusion(id string, isExcluded bool) (AdminRoadkillRecord, bool) {
	s.mu.Lock()
	rk, exists := s.roadkills[id]
	if !exists {
		s.mu.Unlock()
		return AdminRoadkillRecord{}, false
	}
	rk.IsExcluded = isExcluded
	s.roadkills[id] = rk
	s.mu.Unlock()

	s.persistRoadkill(rk)
	return rk, true
}

func (s *Store) SetRoadkillDisposed(id string) (AdminRoadkillRecord, bool) {
	s.mu.Lock()
	rk, exists := s.roadkills[id]
	if !exists {
		s.mu.Unlock()
		return AdminRoadkillRecord{}, false
	}
	rk.Status = "DISPOSED"
	s.roadkills[id] = rk
	s.mu.Unlock()

	s.persistRoadkill(rk)
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
	if env.ID == "" {
		env.ID = fmt.Sprintf("env-%d", time.Now().UnixNano()%100000)
	}
	if env.LastUpdated.IsZero() {
		env.LastUpdated = time.Now().UTC()
	}
	s.environmental[env.ID] = env
	s.mu.Unlock()

	s.persistEnvironmental(env)
	return env
}

func (s *Store) DeleteEnvironmental(id string) bool {
	s.mu.Lock()
	if _, exists := s.environmental[id]; !exists {
		s.mu.Unlock()
		return false
	}
	delete(s.environmental, id)
	s.mu.Unlock()

	s.persistDeleteEnvironmental(id)
	return true
}

// --- Weather CRUD ---

func (s *Store) ListWeather(regionFilter string) []AdminWeatherData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]AdminWeatherData, 0, len(s.weather))
	for _, w := range s.weather {
		if regionFilter != "" && !strings.Contains(strings.ToLower(w.Region), strings.ToLower(regionFilter)) {
			continue
		}
		res = append(res, w)
	}
	return res
}

func (s *Store) GetWeather(id string) (AdminWeatherData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w, exists := s.weather[id]
	return w, exists
}

func (s *Store) CreateWeather(w AdminWeatherData) AdminWeatherData {
	s.mu.Lock()
	if w.ID == "" {
		w.ID = fmt.Sprintf("weather-%d", time.Now().UnixNano())
	}
	if w.LastUpdated.IsZero() {
		w.LastUpdated = time.Now().UTC()
	}
	if w.PrecipitationMmH >= 25.0 {
		w.IsFloodedRiverRisk = true
	}
	if w.MoonPhase == "" {
		if w.MoonIllumination <= 0.05 {
			w.MoonPhase = "NEW_MOON"
		} else if w.MoonIllumination < 0.45 {
			w.MoonPhase = "WAXING_CRESCENT"
		} else if w.MoonIllumination <= 0.55 {
			w.MoonPhase = "FIRST_QUARTER"
		} else if w.MoonIllumination < 0.95 {
			w.MoonPhase = "WAXING_GIBBOUS"
		} else {
			w.MoonPhase = "FULL_MOON"
		}
	}
	s.weather[w.ID] = w
	s.mu.Unlock()

	s.persistWeather(w)
	return w
}

func (s *Store) UpdateWeather(id string, w AdminWeatherData) (AdminWeatherData, bool) {
	s.mu.Lock()
	if _, exists := s.weather[id]; !exists {
		s.mu.Unlock()
		return AdminWeatherData{}, false
	}
	w.ID = id
	w.LastUpdated = time.Now().UTC()
	if w.PrecipitationMmH >= 25.0 {
		w.IsFloodedRiverRisk = true
	}
	s.weather[id] = w
	s.mu.Unlock()

	s.persistWeather(w)
	return w, true
}

func (s *Store) DeleteWeather(id string) bool {
	s.mu.Lock()
	if _, exists := s.weather[id]; !exists {
		s.mu.Unlock()
		return false
	}
	delete(s.weather, id)
	s.mu.Unlock()

	s.persistDeleteWeather(id)
	return true
}

// --- Dashboard Stats ---

func (s *Store) GetDashboardStats() DashboardStatsResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now().UTC()
	cutoff24h := now.Add(-24 * time.Hour)

	stats := DashboardStatsResponse{
		TotalGeofences:       len(s.geofences),
		TotalRoadkills:       len(s.roadkills),
		TotalWildlifeAll:     len(s.wildlife),
		TotalWeatherStations: len(s.weather),
		TotalUsers:           s.userCount,
		SpeciesDistribution:  make(map[string]int),
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
	case "weather":
		for _, w := range s.weather {
			features = append(features, map[string]interface{}{
				"type": "Feature",
				"geometry": map[string]interface{}{
					"type":        "Point",
					"coordinates": []float64{w.Longitude, w.Latitude},
				},
				"properties": map[string]interface{}{
					"id":                    w.ID,
					"station_name":          w.StationName,
					"region":                w.Region,
					"temperature_celsius":   w.TemperatureCelsius,
					"humidity_pct":          w.HumidityPct,
					"wind_speed_kmh":        w.WindSpeedKmh,
					"wind_direction_deg":    w.WindDirectionDeg,
					"precipitation_mm_h":    w.PrecipitationMmH,
					"moon_illumination":     w.MoonIllumination,
					"moon_phase":            w.MoonPhase,
					"is_flooded_river_risk": w.IsFloodedRiverRisk,
					"weather_condition":     w.WeatherCondition,
					"notes":                 w.Notes,
					"last_updated":          w.LastUpdated.Format(time.RFC3339),
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
	case "weather":
		_ = writer.Write([]string{"ID", "StationName", "Region", "Latitude", "Longitude", "TemperatureCelsius", "HumidityPct", "WindSpeedKmh", "WindDirectionDeg", "PrecipitationMmH", "MoonIllumination", "MoonPhase", "FloodedRiverRisk", "Condition", "LastUpdated"})
		for _, w := range s.weather {
			_ = writer.Write([]string{
				w.ID, w.StationName, w.Region,
				strconv.FormatFloat(w.Latitude, 'f', 6, 64),
				strconv.FormatFloat(w.Longitude, 'f', 6, 64),
				strconv.FormatFloat(w.TemperatureCelsius, 'f', 1, 64),
				strconv.FormatFloat(w.HumidityPct, 'f', 1, 64),
				strconv.FormatFloat(w.WindSpeedKmh, 'f', 1, 64),
				strconv.FormatFloat(w.WindDirectionDeg, 'f', 1, 64),
				strconv.FormatFloat(w.PrecipitationMmH, 'f', 1, 64),
				strconv.FormatFloat(w.MoonIllumination, 'f', 2, 64),
				w.MoonPhase,
				strconv.FormatBool(w.IsFloodedRiverRisk),
				w.WeatherCondition,
				w.LastUpdated.Format(time.RFC3339),
			})
		}
	default:
		return nil, "", fmt.Errorf("unknown dataset: %s", dataset)
	}

	writer.Flush()
	return buf.Bytes(), "text/csv", nil
}
