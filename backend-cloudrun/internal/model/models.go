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

type CausalAlertExplanationResponse struct {
	Explanation          string                 `json:"explanation"`
	HistoricalAnalogs    []IncidentMemoryRecord `json:"historicalAnalogs,omitempty"`
	ElevatedFromTier     string                 `json:"elevatedFromTier,omitempty"`
	EffectiveTier        string                 `json:"effectiveTier"`
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
	WindSpeedKmh     float64 `json:"windSpeedKmh"`
	WindDirectionDeg float64 `json:"windDirectionDeg"` // 0 = North, 90 = East, 180 = South, 270 = West (direction wind is blowing FROM)
	MoonIllumination float64 `json:"moonIllumination"` // 0.0 (New Moon) to 1.0 (Full Moon)
	PrecipitationMmH float64 `json:"precipitationMmH"`
}

type AttractantHub struct {
	Id           string  `json:"id"`
	Type         string  `json:"type"` // KAKI_TREE, ELEPHANT_CROP, ROADKILL, KILL_SITE
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	FruitLoad    string  `json:"fruitLoad,omitempty"` // SPARSE, MODERATE, ABUNDANT
	IsTop5Hazard bool    `json:"isTop5Hazard,omitempty"`
	HoursSinceDeath float64 `json:"hoursSinceDeath,omitempty"`
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

