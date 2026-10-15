package admin

import "time"

// Geometry coordinate pair
type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Location representation for incidents
type IncidentLocation struct {
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	RoadName string  `json:"road_name,omitempty"`
	Region   string  `json:"region,omitempty"`
}

// AdminGeofence represents a predictive or manual geofence
type AdminGeofence struct {
	ID                 string       `json:"id"`
	Name               string       `json:"name"`
	Source             string       `json:"source"` // "ai" or "manual"
	IsPersistent       bool         `json:"is_persistent"`
	GeometryType       string       `json:"geometry_type"` // "circle" or "polygon"
	Coordinates        []Coordinate `json:"coordinates,omitempty"`
	CircleCenter       *Coordinate  `json:"circle_center,omitempty"`
	CircleRadiusMeters float64      `json:"circle_radius_meters,omitempty"`
	ThreatLevel        string       `json:"threat_level"` // "TIER_1_CAUTION", "TIER_2_ALERT", "TIER_3_CRITICAL"
	TargetSpecies      string       `json:"target_species"`
	CreatedAt          time.Time    `json:"created_at"`
	ExpiresAt          *time.Time   `json:"expires_at,omitempty"`
	Reason             string       `json:"reason"`
}

// AdminWildlifeIncident represents a wildlife sighting or attack report
type AdminWildlifeIncident struct {
	ID          string           `json:"id"`
	Species     string           `json:"species"` // "Asian Elephant", "Sumatran Tiger", "Bengal Tiger", "Brown Bear", "Asian Black Bear"
	Location    IncidentLocation `json:"location"`
	ThreatLevel string           `json:"threat_level"` // "TIER_1_CAUTION", "TIER_2_ALERT", "TIER_3_CRITICAL"
	Status      string           `json:"status"`       // "UNVERIFIED", "VERIFIED", "RESOLVED", "FALSE_ALARM"
	Reporter    string           `json:"reporter"`
	Description string           `json:"description"`
	PhotoURLs   []string         `json:"photo_urls,omitempty"`
	ReportedAt  time.Time        `json:"reported_at"`
}

// AdminRoadkillRecord represents a roadkill incident record
type AdminRoadkillRecord struct {
	ID                string           `json:"id"`
	Species           string           `json:"species"`
	EstimatedWeightKg float64          `json:"estimated_weight_kg"`
	DecayStage        string           `json:"decay_stage"` // "FRESH", "MODERATE", "ADVANCED"
	Location          IncidentLocation `json:"location"`
	PhotoURL          string           `json:"photo_url,omitempty"`
	PhotoURLs         []string         `json:"photo_urls,omitempty"`
	ReliabilityScore  float64          `json:"reliability_score"` // 0.0 - 1.0
	IsExcluded        bool             `json:"is_excluded"`
	Status            string           `json:"status"` // "PENDING_REMOVAL", "DISPOSED"
	ReportedAt        time.Time        `json:"reported_at"`
}

// AdminEnvironmentalData represents an attractant or environmental hazard zone
type AdminEnvironmentalData struct {
	ID           string       `json:"id"`
	Category     string       `json:"category"` // "CROP_PHENOLOGY", "PERSIMMON_KAKI", "LIVESTOCK_PASTURE", "WATER_SOURCE"
	Name         string       `json:"name"`
	LocationDesc string       `json:"location_desc"`
	GeometryType string       `json:"geometry_type"` // "circle" or "polygon"
	Coordinates  []Coordinate `json:"coordinates,omitempty"`
	CircleCenter *Coordinate  `json:"circle_center,omitempty"`
	RadiusMeters float64      `json:"radius_meters,omitempty"`
	RiskWeight   float64      `json:"risk_weight"` // 0.1 - 1.0
	ActiveSeason string       `json:"active_season"`
	PhotoURLs    []string     `json:"photo_urls,omitempty"`
	LastUpdated  time.Time    `json:"last_updated"`
}

// UploadResponse represents the response with uploaded media URLs
type UploadResponse struct {
	URLs []string `json:"urls"`
}

// DashboardStatsResponse aggregates operational KPIs
type DashboardStatsResponse struct {
	TotalGeofences     int            `json:"total_geofences"`
	AiGeofences        int            `json:"ai_geofences"`
	ManualGeofences    int            `json:"manual_geofences"`
	Tier3Geofences     int            `json:"tier3_geofences"`
	TotalWildlife24h   int            `json:"total_wildlife_24h"`
	TotalWildlifeAll   int            `json:"total_wildlife_all"`
	PendingRoadkills   int            `json:"pending_roadkills"`
	TotalRoadkills      int            `json:"total_roadkills"`
	TotalWeatherStations int           `json:"total_weather_stations"`
	TotalUsers          int            `json:"total_users"`
	SpeciesDistribution map[string]int `json:"species_distribution"`
}

// EmergencyBroadcastRequest represents a push broadcast payload
type EmergencyBroadcastRequest struct {
	GeofenceID string  `json:"geofence_id,omitempty"`
	RadiusKm   float64 `json:"radius_km,omitempty"`
	CenterLat  float64 `json:"center_lat,omitempty"`
	CenterLng  float64 `json:"center_lng,omitempty"`
	Title      string  `json:"title"`
	Body       string  `json:"body"`
	Priority   string  `json:"priority"` // "HIGH", "URGENT"
}

// EmergencyBroadcastResponse represents the broadcast result
type EmergencyBroadcastResponse struct {
	BroadcastID        string    `json:"broadcast_id"`
	EstimatedRecipients int       `json:"estimated_recipients"`
	DispatchedAt       time.Time `json:"dispatched_at"`
	Status             string    `json:"status"`
}

// AdminWeatherData represents meteorological observations used by AI prediction engines
type AdminWeatherData struct {
	ID                 string    `json:"id"`
	StationName        string    `json:"station_name"`
	Region             string    `json:"region"`
	Latitude           float64   `json:"latitude"`
	Longitude          float64   `json:"longitude"`
	TemperatureCelsius float64   `json:"temperature_celsius"`
	HumidityPct        float64   `json:"humidity_pct"`
	WindSpeedKmh       float64   `json:"wind_speed_kmh"`
	WindDirectionDeg   float64   `json:"wind_direction_deg"` // 0=N, 90=E, 180=S, 270=W
	PrecipitationMmH   float64   `json:"precipitation_mm_h"`
	MoonIllumination   float64   `json:"moon_illumination"` // 0.0 - 1.0
	MoonPhase          string    `json:"moon_phase"`        // NEW_MOON, WAXING_CRESCENT, FULL_MOON, etc.
	IsFloodedRiverRisk bool      `json:"is_flooded_river_risk"`
	WeatherCondition   string    `json:"weather_condition"` // CLEAR, OVERCAST, RAIN, HEAVY_RAIN, STORM, FOG
	Notes              string    `json:"notes,omitempty"`
	LastUpdated        time.Time `json:"last_updated"`
}

