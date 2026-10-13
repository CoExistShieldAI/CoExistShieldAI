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
