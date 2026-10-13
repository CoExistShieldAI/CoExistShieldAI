package geofence

import (
	"time"

	"ai.coexistshield/backend/internal/model"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Recalculate(req model.GeofenceRecalculateRequest) model.GeofenceRecalculateResponse {
	tiers := []string{"TIER_1_NOTICE", "TIER_2_WARNING"}
	if req.TriggerEntityRef != "" {
		tiers = append(tiers, "TIER_3_DANGER")
	}

	return model.GeofenceRecalculateResponse{
		RecalculatedGeofencesCount: 3,
		AffectedTiers:              tiers,
		UpdatedAt:                  time.Now().UTC().Format(time.RFC3339),
	}
}

func (e *Engine) GetTopHazardPersimmonTrees(municipalityCode string, limit int) model.TopHazardsResponse {
	name := "Morioka City"
	if municipalityCode == "03201" {
		name = "Morioka City (岩手県盛岡市)"
	}

	trees := []model.PriorityKakiTreeDto{
		{
			TreeId:                         "kt_1120",
			HazardScore:                    0.95,
			DistanceToSchoolRouteMeters:    120,
			EstimatedBearAttractantRadiusM: 800,
			RecommendedClearanceOrder:      1,
		},
		{
			TreeId:                         "kt_0842",
			HazardScore:                    0.91,
			DistanceToSchoolRouteMeters:    240,
			EstimatedBearAttractantRadiusM: 650,
			RecommendedClearanceOrder:      2,
		},
		{
			TreeId:                         "kt_0319",
			HazardScore:                    0.88,
			DistanceToSchoolRouteMeters:    310,
			EstimatedBearAttractantRadiusM: 500,
			RecommendedClearanceOrder:      3,
		},
	}

	if limit > 0 && limit < len(trees) {
		trees = trees[:limit]
	}

	return model.TopHazardsResponse{
		MunicipalityName:             name,
		TotalCatalogedTrees:          840,
		Top5PercentCutoffHazardScore: 0.89,
		PriorityTrees:                trees,
	}
}
