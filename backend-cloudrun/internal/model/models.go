package model

type ForensicTriageRequest struct {
	ImageBase64  string  `json:"imageBase64"`
	CategoryHint string  `json:"categoryHint,omitempty"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}

type TriageAttributes struct {
	FruitLoadLevel string  `json:"fruitLoadLevel,omitempty"`
	RipenessStage  string  `json:"ripenessStage,omitempty"`
	DecayState     *string `json:"decayState,omitempty"`
	TraumaType     *string `json:"traumaType,omitempty"`
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

type CausalAlertExplanationRequest struct {
	Tier              string                 `json:"tier"`
	TargetSpecies     string                 `json:"targetSpecies"`
	Weather           map[string]interface{} `json:"weather"`
	NearbyAttractants []map[string]interface{} `json:"nearbyAttractants"`
}

type CausalAlertExplanationResponse struct {
	Explanation string `json:"explanation"`
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
