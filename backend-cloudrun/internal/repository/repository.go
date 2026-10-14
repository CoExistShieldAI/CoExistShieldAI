package repository

import (
	"context"

	"ai.coexistshield/backend/internal/model"
)

// IncidentRepository defines persistence methods for field incident reports.
type IncidentRepository interface {
	GetIncidents(ctx context.Context, limit int) ([]model.IncidentDto, error)
	GetIncidentByID(ctx context.Context, id string) (*model.IncidentDto, error)
	SaveIncident(ctx context.Context, incident model.IncidentDto) error
	UpdateIncidentVerification(ctx context.Context, id string, status string) error
}

// RoadkillRepository defines persistence methods for carcass hazards and attractants.
type RoadkillRepository interface {
	GetRoadkills(ctx context.Context, statusFilter string) ([]model.RoadkillRecordDto, error)
	GetRoadkillByID(ctx context.Context, id string) (*model.RoadkillRecordDto, error)
	SaveRoadkill(ctx context.Context, roadkill model.RoadkillRecordDto) error
	UpdateRoadkillStatus(ctx context.Context, id string, status string, notes string) (*model.RoadkillRecordDto, error)
}

// GeofenceRepository defines persistence methods for active dynamic danger tiers.
type GeofenceRepository interface {
	GetActiveGeofences(ctx context.Context) ([]model.ActiveGeofenceZone, error)
	SaveGeofences(ctx context.Context, geofences []model.ActiveGeofenceZone) error
}

// AttractantRepository defines persistence for kaki trees, crops, and livestock pastures.
type AttractantRepository interface {
	GetPriorityKakiTrees(ctx context.Context, municipalityCode string, limit int) ([]model.PriorityKakiTreeDto, error)
	RecordKakiAction(ctx context.Context, treeID string, actionType string, notes string) error
}
