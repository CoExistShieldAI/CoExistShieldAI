package model

type ForensicTriageRequest struct {
	ImageBase64  string  `json:"imageBase64"`
	CategoryHint string  `json:"categoryHint,omitempty"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

type TriageAttributes struct {
	FruitLoadLevel string   `json:"fruitLoadLevel,omitempty"`
	RipenessStage  string   `json:"ripenessStage,omitempty"`
	DecayState     *string  `json:"decayState,omitempty"`
	TraumaType     *string  `json:"traumaType,omitempty"`
	TraceType      *string  `json:"traceType,omitempty"`     // CLAW_MARK, PUGMARK, SCAT, RUB_TREE, FENCE_BREACH
	TraceAgeHours  *float64 `json:"traceAgeHours,omitempty"` // Estimated age in hours
}

type ForensicTriageResponse struct {
	DetectedSpecies        string           `json:"detectedSpecies"`
	CommonName             string           `json:"commonName,omitempty"`
	Confidence             float64          `json:"confidence"`
	Attributes             *TriageAttributes `json:"attributes,omitempty"`
	SuggestedTags          []string         `json:"suggestedTags"`
	PreliminaryThreatIndex float64          `json:"preliminaryThreatIndex"`
	Notes                  string           `json:"notes,omitempty"`
	Source                 string           `json:"source,omitempty"` // "gemini" or "mock" / "mock_fallback"
}

type TargetCenter struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type GeofenceRecalculateRequest struct {
	TriggerEntityRef     string        `json:"triggerEntityRef"`
	TargetCenter         *TargetCenter `json:"targetCenter,omitempty"`
	TargetCenterLatitude float64       `json:"targetCenterLatitude,omitempty"`
	TargetCenterLongitude float64      `json:"targetCenterLongitude,omitempty"`
	RadiusKm             float64       `json:"radiusKm"`
}

type GeofenceRecalculateResponse struct {
	RecalculatedGeofencesCount int      `json:"recalculatedGeofencesCount"`
	AffectedTiers              []string `json:"affectedTiers"`
	UpdatedAt                  string   `json:"updatedAt"`
}

type PriorityKakiTreeDto struct {
	TreeId                         string  `json:"treeId"`
	HazardScore                    float64 `json:"hazardScore"`
	DistanceToSchoolRouteMeters    int     `json:"distanceToSchoolRouteMeters"`
	EstimatedBearAttractantRadiusM int     `json:"estimatedBearAttractantRadiusM"`
	RecommendedClearanceOrder      int     `json:"recommendedClearanceOrder"`
}

type TopHazardsResponse struct {
	MunicipalityName             string                `json:"municipalityName"`
	TotalCatalogedTrees          int                   `json:"totalCatalogedTrees"`
	Top5PercentCutoffHazardScore float64               `json:"top5PercentCutoffHazardScore"`
	PriorityTrees                []PriorityKakiTreeDto `json:"priorityTrees"`
}

type IncidentMemoryRecord struct {
	MemoryId        string    `json:"memoryId"`
	Region          string    `json:"region"`   // JAPAN, THAILAND, INDONESIA
	Species         string    `json:"species"`  // BEAR, ELEPHANT, TIGER
	Category        string    `json:"category"` // SIGHTING, ATTACK, CROP_RAID, ROADKILL
	SeverityTier    string    `json:"severityTier"`
	Narrative       string    `json:"narrative"`
	Outcome         string    `json:"outcome"`
	EmbeddingVector []float64 `json:"embeddingVector,omitempty"` // 768-dim vector
	CosineSimilarity float64  `json:"cosineSimilarity,omitempty"`
}

type MemorySearchRequest struct {
	QueryText     string   `json:"queryText"`
	Region        string   `json:"region,omitempty"`
	Species       string   `json:"species,omitempty"`
	TopK          int      `json:"topK,omitempty"`
	MinSimilarity float64  `json:"minSimilarity,omitempty"` // default 0.85
}

type MemorySearchResponse struct {
	QueryVectorGenerated bool                   `json:"queryVectorGenerated"`
	Matches              []IncidentMemoryRecord `json:"matches"`
	ElevatedTier         string                 `json:"elevatedTier,omitempty"`
	ElevationReason      string                 `json:"elevationReason,omitempty"`
}

type CausalAlertExplanationRequest struct {
	Tier              string                 `json:"tier"`
	TargetSpecies     string                 `json:"targetSpecies"`
	Region            string                 `json:"region,omitempty"`
	Weather           map[string]interface{} `json:"weather"`
	NearbyAttractants []map[string]interface{} `json:"nearbyAttractants"`
	RetrieveMemory    bool                   `json:"retrieveMemory,omitempty"`
}

type DispatchRecommendation struct {
	ActionType      string `json:"actionType"`      // "FELL_PERSIMMON_TREE", "CLEAR_ROADKILL_CARCASS", "PATROL_CORRIDOR", "ALERT_SCHOOL_ROUTE"
	Priority        string `json:"priority"`        // "IMMEDIATE", "HIGH", "SCHEDULED"
	TargetEntityRef string `json:"targetEntityRef"` // e.g. "kt_1120", "rk_4412"
	Location        string `json:"location"`        // e.g. "Takamatsu 3-chome School Route (120m)"
	Justification   string `json:"justification"`   // e.g. "Overripe persimmon tree directly upwind of residential perimeter matching 3 historical bear intrusions."
}

type CausalAlertExplanationResponse struct {
	Explanation             string                   `json:"explanation"`
	HistoricalAnalogs       []IncidentMemoryRecord   `json:"historicalAnalogs,omitempty"`
	ElevatedFromTier        string                   `json:"elevatedFromTier,omitempty"`
	EffectiveTier           string                   `json:"effectiveTier"`
	DispatchRecommendations []DispatchRecommendation `json:"dispatchRecommendations,omitempty"`
}


// GeoJSON Types adhering to RFC 7946

type GeoJSONGeometry struct {
	Type        string        `json:"type"` // "Polygon" or "MultiPolygon"
	Coordinates [][][][]float64 `json:"coordinates,omitempty"` // for MultiPolygon: [polygon][ring][point][lon, lat]
}

type GeoJSONPolygonGeometry struct {
	Type        string        `json:"type"` // "Polygon"
	Coordinates [][][]float64 `json:"coordinates"` // [ring][point][lon, lat]
}

type GeofenceZoneDto struct {
	Id               string                  `json:"id"`
	Tier             string                  `json:"tier"` // SAFE, TIER_1_NOTICE, TIER_2_WARNING, TIER_3_DANGER
	TargetSpecies    []string                `json:"targetSpecies"`
	PolygonGeoJson   *GeoJSONPolygonGeometry `json:"polygonGeoJson"`
	ActiveRiskScore  float64                 `json:"activeRiskScore"`
	CausalFactors    []string                `json:"causalFactors"`
	ExpiresAt        string                  `json:"expiresAt"`
	CreatedAt        string                  `json:"createdAt"`
}

type WeatherConditions struct {
	WindSpeedKmh       float64  `json:"windSpeedKmh"`
	WindDirectionDeg   float64  `json:"windDirectionDeg"` // 0 = North, 90 = East, 180 = South, 270 = West (direction wind is blowing FROM)
	MoonIllumination   float64  `json:"moonIllumination"` // 0.0 (New Moon) to 1.0 (Full Moon)
	PrecipitationMmH   float64  `json:"precipitationMmH"`
	TemperatureCelsius *float64 `json:"temperatureCelsius,omitempty"`
	HumidityPct        *float64 `json:"humidityPct,omitempty"`
}

type MoonPhaseInfo struct {
	PhaseName               string  `json:"phaseName"`               // NEW_MOON, WAXING_CRESCENT, FIRST_QUARTER, WAXING_GIBBOUS, FULL_MOON, WANING_GIBBOUS, LAST_QUARTER, WANING_CRESCENT
	Illumination            float64 `json:"illumination"`            // 0.0 to 1.0
	NocturnalRiskMultiplier float64 `json:"nocturnalRiskMultiplier"` // 1.0 to 1.6x
	IsNocturnalPeakRisk     bool    `json:"isNocturnalPeakRisk"`     // True if illumination <= 0.20
}

type SolarPhaseInfo struct {
	PhaseName                string  `json:"phaseName"`                // NIGHT, ASTRONOMICAL_TWILIGHT, NAUTICAL_TWILIGHT, CIVIL_TWILIGHT, GOLDEN_HOUR, DAYLIGHT
	ElevationDegrees         float64 `json:"elevationDegrees"`         // Solar elevation angle above horizon in degrees
	CrepuscularRiskMultiplier float64 `json:"crepuscularRiskMultiplier"` // 0.75x (midday shelter) to 1.75x (dawn/dusk peak hunting/foraging)
	IsCrepuscularPeakRisk    bool    `json:"isCrepuscularPeakRisk"`    // True during civil twilight / golden hour (elevation between -6° and +6°)
}

type MeteorologicalConditions struct {
	TemperatureCelsius float64        `json:"temperatureCelsius"`
	HumidityPct        float64        `json:"humidityPct"`
	WindSpeedKmh       float64        `json:"windSpeedKmh"`
	WindDirectionDeg   float64        `json:"windDirectionDeg"`
	PrecipitationMmH   float64        `json:"precipitationMmH"`
	IsFloodedRiverRisk bool           `json:"isFloodedRiverRisk"` // True if PrecipitationMmH >= 25.0
	MoonPhase          MoonPhaseInfo  `json:"moonPhase"`
	SolarPhase         SolarPhaseInfo `json:"solarPhase"`
	Timestamp          string         `json:"timestamp"`
}

type AttractantHub struct {
	Id              string   `json:"id"`
	Type            string   `json:"type"` // KAKI_TREE, ELEPHANT_CROP, ROADKILL, KILL_SITE, PREDATOR_SIGHTING
	Latitude        float64  `json:"latitude"`
	Longitude       float64  `json:"longitude"`
	FruitLoad       string   `json:"fruitLoad,omitempty"` // SPARSE, MODERATE, ABUNDANT
	IsTop5Hazard    bool     `json:"isTop5Hazard,omitempty"`
	HoursSinceDeath float64  `json:"hoursSinceDeath,omitempty"`
	HoursSinceSeen  *float64 `json:"hoursSinceSeen,omitempty"`
}

type GeneratePlumeGeofencesRequest struct {
	Attractants []AttractantHub    `json:"attractants"`
	Weather     *WeatherConditions `json:"weather,omitempty"`
}

type GeneratePlumeGeofencesResponse struct {
	Zones     []GeofenceZoneDto `json:"zones"`
	Count     int               `json:"count"`
	CreatedAt string            `json:"createdAt"`
}

// Next-Gen Feature Models (Epic #64)

type HourlyForecastPoint struct {
	Timestamp        string  `json:"timestamp"`
	OffsetHours      int     `json:"offsetHours"`
	WindSpeedKmh     float64 `json:"windSpeedKmh"`
	WindDirectionDeg float64 `json:"windDirectionDeg"`
	PrecipitationMmH float64 `json:"precipitationMmH"`
	SolarPhase       string  `json:"solarPhase"`
}

type ForecastZoneStep struct {
	OffsetHours          int              `json:"offsetHours"`
	ForecastTime         string           `json:"forecastTime"`
	Zone                 GeofenceZoneDto  `json:"zone"`
	WindBearingDeg       float64          `json:"windBearingDeg"`
	WindSpeedKmh         float64          `json:"windSpeedKmh"`
	CrepuscularRiskBoost float64          `json:"crepuscularRiskBoost"`
	SolarPhase           SolarPhaseInfo   `json:"solarPhase"`
}

type ForecastPlumeGeofencesRequest struct {
	Attractants []AttractantHub `json:"attractants"`
	TargetHours []int           `json:"targetHours,omitempty"` // e.g. [1, 3, 6, 12] hours into future
	Region      string          `json:"region,omitempty"`
}

type ForecastPlumeGeofencesResponse struct {
	TimelineSteps []ForecastZoneStep `json:"timelineSteps"`
	TotalSteps    int                `json:"totalSteps"`
	GeneratedAt   string             `json:"generatedAt"`
}

// OsmSpatialContext models real-world barriers and natural corridors from OpenStreetMap
type OsmSpatialContext struct {
	NearbyHighwaysCount   int      `json:"nearbyHighwaysCount"`
	HasDividedMotorway    bool     `json:"hasDividedMotorway"`
	DividedHighwayNames   []string `json:"dividedHighwayNames,omitempty"`
	NearbyRiversCount     int      `json:"nearbyRiversCount"`
	RiparianCorridorName  string   `json:"riparianCorridorName,omitempty"`
	ForestEdgeProximityM  float64  `json:"forestEdgeProximityM"`
	CorridorDeflectionDeg float64  `json:"corridorDeflectionDeg"` // Angle of river flow guiding movement
}

type DeescalationResult struct {
	ZoneId            string  `json:"zoneId"`
	OriginalRadiusM   float64 `json:"originalRadiusM"`
	CurrentRadiusM    float64 `json:"currentRadiusM"`
	ContractionRatio  float64 `json:"contractionRatio"` // e.g. 0.75 (25% reduction)
	OriginalTier      string  `json:"originalTier"`
	CurrentTier       string  `json:"currentTier"`
	HoursSinceTrigger float64 `json:"hoursSinceTrigger"`
	Status            string  `json:"status"` // ACTIVE, CONTRACTED, DEMOTED, EXPIRED
}

type ActiveGeofencePoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type ActiveGeofenceZone struct {
	Id                   string                  `json:"id"`
	Tier                 string                  `json:"tier"` // SAFE, TIER_1_NOTICE, TIER_2_WARNING, TIER_3_DANGER
	TargetSpecies        []string                `json:"targetSpecies"`
	Coordinates          []ActiveGeofencePoint   `json:"coordinates"`
	Centroid             ActiveGeofencePoint     `json:"centroid"`
	RadiusMeters         float64                 `json:"radiusMeters"`
	CausalExplanation    string                  `json:"causalExplanation"`
	PrimaryAttractantRef string                  `json:"primaryAttractantRef,omitempty"`
	ActiveRiskScore      float64                 `json:"activeRiskScore"`
	PolygonGeoJson       *GeoJSONPolygonGeometry `json:"polygonGeoJson,omitempty"`
	ValidFrom            string                  `json:"validFrom"`
	ValidUntil           string                  `json:"validUntil"`
	GeneratedAt          string                  `json:"generatedAt"`
}

// Issue #10: Gemini 3.8 Flash Multimodal & NLP Incident Entity Models

type ExtractedIncidentEntity struct {
	EntityId            string   `json:"entityId"`
	SourceChannel       string   `json:"sourceChannel"` // MUNICIPAL_DISASTER_RADIO, X_POST, BLUESKY, CITIZEN_HOTLINE, REGIONAL_BOARD
	RawText             string   `json:"rawText"`
	Species             string   `json:"species"`        // BEAR, ELEPHANT, TIGER, WILD_BOAR, DEER, UNKNOWN
	ScientificName      string   `json:"scientificName"`
	CommonName          string   `json:"commonName"`
	IncidentType        string   `json:"incidentType"`   // LIVE_SIGHTING, PHYSICAL_TRACE, CROP_RAID, DEPREDATION, ROADKILL
	LocationDescription string   `json:"locationDescription"`
	EstimatedLatitude   float64  `json:"estimatedLatitude"`
	EstimatedLongitude  float64  `json:"estimatedLongitude"`
	IndividualCount     int      `json:"individualCount"`
	AttractantContext   string   `json:"attractantContext"`
	UrgencyLevel        string   `json:"urgencyLevel"` // LOW, MEDIUM, HIGH, CRITICAL
	Confidence          float64  `json:"confidence"`
	ExtractedTimestamp  string   `json:"extractedTimestamp"`
	CorroborationTags   []string `json:"corroborationTags"`
	Notes               string   `json:"notes"`
	InferenceProvider   string   `json:"inferenceProvider,omitempty"` // "gemini" or "mock" / "mock_fallback"
}

type UnstructuredNlpIngestRequest struct {
	RawText       string `json:"rawText"`
	SourceChannel string `json:"sourceChannel,omitempty"`
	RegionHint    string `json:"regionHint,omitempty"`
}

type UnstructuredNlpIngestResponse struct {
	Entity    ExtractedIncidentEntity `json:"entity"`
	ModelUsed string                  `json:"modelUsed"` // "gemini-3.8-flash"
	CreatedAt string                  `json:"createdAt"`
}

// Issue #4: Multi-Source Causal Environmental Data Ingestion Pipeline & Sensor Fusion Corroboration Engine

type CausalStreamType string

const (
	StreamAgriculturalHarvest CausalStreamType = "AGRICULTURAL_HARVEST"
	StreamMunicipalAttractant  CausalStreamType = "MUNICIPAL_ATTRACTANT"
	StreamSatelliteThermal    CausalStreamType = "SATELLITE_THERMAL"
	StreamSocialNLP           CausalStreamType = "SOCIAL_NLP_STREAM"
	StreamTrafficCongestion   CausalStreamType = "TRAFFIC_CONGESTION"
)

type CausalStreamItem struct {
	ItemId        string                 `json:"itemId"`
	StreamType    CausalStreamType       `json:"streamType"`
	Region        string                 `json:"region"` // JAPAN, THAILAND, INDONESIA
	Title         string                 `json:"title"`
	Description   string                 `json:"description"`
	Latitude      float64                `json:"latitude"`
	Longitude     float64                `json:"longitude"`
	RadiusMeters  float64                `json:"radiusMeters"`
	Confidence    float64                `json:"confidence"`
	TargetSpecies string                 `json:"targetSpecies"` // BEAR, ELEPHANT, TIGER, ALL
	Timestamp     string                 `json:"timestamp"`
	ExpiresAt     string                 `json:"expiresAt"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type CorroborationRequest struct {
	TargetLatitude  float64 `json:"targetLatitude"`
	TargetLongitude float64 `json:"targetLongitude"`
	RadiusKm        float64 `json:"radiusKm,omitempty"`    // default 5.0 km
	MaxAgeHours     float64 `json:"maxAgeHours,omitempty"` // default 12.0 hours
	Region          string  `json:"region,omitempty"`
}

type CorroboratedCluster struct {
	ClusterId              string             `json:"clusterId"`
	CenterLatitude         float64            `json:"centerLatitude"`
	CenterLongitude        float64            `json:"centerLongitude"`
	RadiusKm               float64            `json:"radiusKm"`
	TargetSpecies          string             `json:"targetSpecies"`
	CorroboratedSources    []CausalStreamType `json:"corroboratedSources"`
	IndependentSourceCount int                `json:"independentSourceCount"`
	CombinedConfidence     float64            `json:"combinedConfidence"` // 1.0 - prod(1 - c_i)
	ElevatedTier           string             `json:"elevatedTier"`       // SAFE, TIER_1_NOTICE, TIER_2_WARNING, TIER_3_DANGER
	AiJustification        string             `json:"aiJustification"`    // Generated by Gemini 3.8 Flash
	ContributingEvents     []CausalStreamItem `json:"contributingEvents"`
	ActionRecommendation   string             `json:"actionRecommendation"`
	EvaluatedAt            string             `json:"evaluatedAt"`
}

type CorroborationResponse struct {
	Clusters           []CorroboratedCluster `json:"clusters"`
	TotalActiveStreams int                   `json:"totalActiveStreams"`
	EvaluatedAt        string                `json:"evaluatedAt"`
	ModelUsed          string                `json:"modelUsed"` // "gemini-3.8-flash"
}

// Crowdsourced Incidents & Roadkill Mitigation Models

type GeoLocationDto struct {
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Altitude  *float64 `json:"altitude,omitempty"`
}

type GeminiTriageResultDto struct {
	DetectedLabel   string   `json:"detectedLabel,omitempty"`
	CommonName      string   `json:"commonName,omitempty"`
	Confidence      float64  `json:"confidence"`
	ForensicTags    []string `json:"forensicTags,omitempty"`
	DecayState      string   `json:"decayState,omitempty"`
	FruitLoadLevel  string   `json:"fruitLoadLevel,omitempty"`
	TraceType       string   `json:"traceType,omitempty"`
	TraumaType      string   `json:"traumaType,omitempty"`
	Notes           string   `json:"notes,omitempty"`
	ModelUsed       string   `json:"modelUsed,omitempty"`
	InferenceTimeMs int64    `json:"inferenceTimeMs,omitempty"`
}

type IncidentDto struct {
	Id                 string                 `json:"id"`
	Category           string                 `json:"category"` // WILDLIFE_SIGHTING, PHYSICAL_TRACE, ROADKILL, LIVESTOCK_DEPREDATION, ATTRACTANT_FRUIT_TREE
	Species            string                 `json:"species"`
	Location           GeoLocationDto         `json:"location"`
	Timestamp          string                 `json:"timestamp"`
	Region             string                 `json:"region"`
	Description        string                 `json:"description,omitempty"`
	PhotoUrls          []string               `json:"photoUrls,omitempty"`
	GeminiTriageResult *GeminiTriageResultDto `json:"geminiTriageResult,omitempty"`
	ReporterId         string                 `json:"reporterId"`
	ReporterRole       string                 `json:"reporterRole"`
	VerificationStatus string                 `json:"verificationStatus"` // PENDING, VERIFIED, REJECTED
}


type RoadkillRecordDto struct {
	Id                    string         `json:"id"`
	IncidentId            string         `json:"incidentId"`
	Species               string         `json:"species"`
	Location              GeoLocationDto `json:"location"`
	DecayState            string         `json:"decayState"` // FRESH, BLOATED, DECOMPOSING, SKELETAL
	AttractantThreatScore float64        `json:"attractantThreatScore"`
	ClearanceStatus       string         `json:"clearanceStatus"` // REPORTED, DISPATCHED, CLEARED
	AssignedOfficialId    string         `json:"assignedOfficialId,omitempty"`
	PhotoUrls             []string       `json:"photoUrls,omitempty"`
	ReportedAt            string         `json:"reportedAt"`
	ClearedAt             string         `json:"clearedAt,omitempty"`
}

type UpdateRoadkillStatusRequestDto struct {
	Status             string `json:"status"` // REPORTED, DISPATCHED, CLEARED
	AssignedOfficialId string `json:"assignedOfficialId,omitempty"`
}

type GeofenceEvent struct {
	Type         string               `json:"type"` // "CREATED", "UPDATED", "DELETED", "RECALCULATED", "CONNECTED"
	Action       string               `json:"action"` // "created", "updated", "deleted", "recalculated", "connected"
	GeofenceID   string               `json:"geofenceId,omitempty"`
	Location     *ActiveGeofencePoint `json:"location,omitempty"`
	RadiusMeters float64              `json:"radiusMeters,omitempty"`
	Tier         string               `json:"tier,omitempty"`
	Message      string               `json:"message,omitempty"`
	Timestamp    string               `json:"timestamp"`
}



