package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ai.coexistshield/backend/internal/model"
)

type MemoryRepository struct {
	mu        sync.RWMutex
	incidents map[string]model.IncidentDto
	roadkills map[string]model.RoadkillRecordDto
	geofences []model.ActiveGeofenceZone
	kakiTrees []model.PriorityKakiTreeDto
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		incidents: make(map[string]model.IncidentDto),
		roadkills: make(map[string]model.RoadkillRecordDto),
		geofences: make([]model.ActiveGeofenceZone, 0),
		kakiTrees: make([]model.PriorityKakiTreeDto, 0),
	}
}

// IncidentRepository implementation
func (m *MemoryRepository) GetIncidents(ctx context.Context, limit int) ([]model.IncidentDto, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]model.IncidentDto, 0, len(m.incidents))
	for _, inc := range m.incidents {
		res = append(res, inc)
		if limit > 0 && len(res) >= limit {
			break
		}
	}
	return res, nil
}

func (m *MemoryRepository) GetIncidentByID(ctx context.Context, id string) (*model.IncidentDto, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	inc, ok := m.incidents[id]
	if !ok {
		return nil, fmt.Errorf("incident not found: %s", id)
	}
	return &inc, nil
}

func (m *MemoryRepository) SaveIncident(ctx context.Context, incident model.IncidentDto) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if incident.Id == "" {
		incident.Id = fmt.Sprintf("inc-%d", time.Now().UnixNano())
	}
	m.incidents[incident.Id] = incident
	return nil
}

func (m *MemoryRepository) UpdateIncidentVerification(ctx context.Context, id string, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inc, ok := m.incidents[id]
	if !ok {
		return fmt.Errorf("incident not found: %s", id)
	}
	inc.VerificationStatus = status
	m.incidents[id] = inc
	return nil
}

// RoadkillRepository implementation
func (m *MemoryRepository) GetRoadkills(ctx context.Context, statusFilter string) ([]model.RoadkillRecordDto, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]model.RoadkillRecordDto, 0, len(m.roadkills))
	for _, rk := range m.roadkills {
		if statusFilter == "" || rk.ClearanceStatus == statusFilter {
			res = append(res, rk)
		}
	}
	return res, nil
}

func (m *MemoryRepository) GetRoadkillByID(ctx context.Context, id string) (*model.RoadkillRecordDto, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rk, ok := m.roadkills[id]
	if !ok {
		return nil, fmt.Errorf("roadkill not found: %s", id)
	}
	return &rk, nil
}

func (m *MemoryRepository) SaveRoadkill(ctx context.Context, roadkill model.RoadkillRecordDto) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if roadkill.Id == "" {
		roadkill.Id = fmt.Sprintf("rk-%d", time.Now().UnixNano())
	}
	m.roadkills[roadkill.Id] = roadkill
	return nil
}

func (m *MemoryRepository) UpdateRoadkillStatus(ctx context.Context, id string, status string, notes string) (*model.RoadkillRecordDto, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rk, ok := m.roadkills[id]
	if !ok {
		return nil, fmt.Errorf("roadkill not found: %s", id)
	}
	rk.ClearanceStatus = status
	if status == "CLEARED" {
		rk.ClearedAt = time.Now().UTC().Format(time.RFC3339)
	}
	m.roadkills[id] = rk
	return &rk, nil
}

// GeofenceRepository implementation
func (m *MemoryRepository) GetActiveGeofences(ctx context.Context) ([]model.ActiveGeofenceZone, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copied := make([]model.ActiveGeofenceZone, len(m.geofences))
	copy(copied, m.geofences)
	return copied, nil
}

func (m *MemoryRepository) SaveGeofences(ctx context.Context, geofences []model.ActiveGeofenceZone) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.geofences = make([]model.ActiveGeofenceZone, len(geofences))
	copy(m.geofences, geofences)
	return nil
}

// AttractantRepository implementation
func (m *MemoryRepository) GetPriorityKakiTrees(ctx context.Context, municipalityCode string, limit int) ([]model.PriorityKakiTreeDto, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]model.PriorityKakiTreeDto, len(m.kakiTrees))
	copy(res, m.kakiTrees)
	if limit > 0 && len(res) > limit {
		res = res[:limit]
	}
	return res, nil
}

func (m *MemoryRepository) RecordKakiAction(ctx context.Context, treeID string, actionType string, notes string) error {
	return nil
}

// SeedInitialData loads developer baseline mock dataset into in-memory store
func (m *MemoryRepository) SeedInitialData(incidents []model.IncidentDto, roadkills []model.RoadkillRecordDto, geofences []model.ActiveGeofenceZone) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, inc := range incidents {
		m.incidents[inc.Id] = inc
	}
	for _, rk := range roadkills {
		m.roadkills[rk.Id] = rk
	}
	m.geofences = append(m.geofences, geofences...)
}
