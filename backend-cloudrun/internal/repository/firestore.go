package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"ai.coexistshield/backend/internal/model"
)

type FirestoreRepository struct {
	projectID  string
	baseURL    string
	httpClient *http.Client
}

func NewFirestoreRepository(projectID string) *FirestoreRepository {
	emulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST")
	var baseURL string
	if emulatorHost != "" {
		baseURL = fmt.Sprintf("http://%s/v1/projects/%s/databases/(default)/documents", emulatorHost, projectID)
	} else {
		baseURL = fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/(default)/documents", projectID)
	}

	return &FirestoreRepository{
		projectID: projectID,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Low-level Firestore REST helpers
type firestoreValue struct {
	StringValue  *string                 `json:"stringValue,omitempty"`
	DoubleValue  *float64                `json:"doubleValue,omitempty"`
	IntegerValue *string                 `json:"integerValue,omitempty"`
	BooleanValue *bool                   `json:"booleanValue,omitempty"`
	MapValue     *firestoreMapValue      `json:"mapValue,omitempty"`
	ArrayValue   *firestoreArrayValue    `json:"arrayValue,omitempty"`
	GeoPoint     *firestoreGeoPointValue `json:"geoPointValue,omitempty"`
}

type firestoreMapValue struct {
	Fields map[string]firestoreValue `json:"fields"`
}

type firestoreArrayValue struct {
	Values []firestoreValue `json:"values"`
}

type firestoreGeoPointValue struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type firestoreDocument struct {
	Name       string                    `json:"name"`
	Fields     map[string]firestoreValue `json:"fields"`
	CreateTime string                    `json:"createTime,omitempty"`
	UpdateTime string                    `json:"updateTime,omitempty"`
}

type firestoreListResponse struct {
	Documents []firestoreDocument `json:"documents"`
}

func strVal(s string) firestoreValue {
	return firestoreValue{StringValue: &s}
}

func dblVal(f float64) firestoreValue {
	return firestoreValue{DoubleValue: &f}
}

func getStr(fields map[string]firestoreValue, key string) string {
	if v, ok := fields[key]; ok && v.StringValue != nil {
		return *v.StringValue
	}
	return ""
}

func getDbl(fields map[string]firestoreValue, key string) float64 {
	if v, ok := fields[key]; ok && v.DoubleValue != nil {
		return *v.DoubleValue
	}
	return 0.0
}

func (f *FirestoreRepository) saveDoc(ctx context.Context, collection string, docID string, fields map[string]firestoreValue) error {
	doc := firestoreDocument{
		Fields: fields,
	}
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s/%s", f.baseURL, collection, docID)
	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("firestore request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("firestore save returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (f *FirestoreRepository) listDocs(ctx context.Context, collection string) ([]firestoreDocument, error) {
	url := fmt.Sprintf("%s/%s", f.baseURL, collection)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("firestore list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []firestoreDocument{}, nil
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("firestore list returned %d: %s", resp.StatusCode, string(body))
	}

	var listResp firestoreListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}
	return listResp.Documents, nil
}

func (f *FirestoreRepository) getDoc(ctx context.Context, collection string, docID string) (*firestoreDocument, error) {
	url := fmt.Sprintf("%s/%s/%s", f.baseURL, collection, docID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("document not found: %s/%s", collection, docID)
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("firestore get returned %d: %s", resp.StatusCode, string(body))
	}

	var doc firestoreDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// IncidentRepository methods
func (f *FirestoreRepository) GetIncidents(ctx context.Context, limit int) ([]model.IncidentDto, error) {
	docs, err := f.listDocs(ctx, "incidents")
	if err != nil {
		return nil, err
	}

	results := make([]model.IncidentDto, 0, len(docs))
	for _, doc := range docs {
		id := doc.Name[strings.LastIndex(doc.Name, "/")+1:]
		locVal := doc.Fields["location"]
		lat, lon := 0.0, 0.0
		if locVal.GeoPoint != nil {
			lat = locVal.GeoPoint.Latitude
			lon = locVal.GeoPoint.Longitude
		}

		results = append(results, model.IncidentDto{
			Id:                 id,
			Species:            getStr(doc.Fields, "species"),
			Category:           getStr(doc.Fields, "category"),
			VerificationStatus: getStr(doc.Fields, "verificationStatus"),
			Location: model.GeoLocationDto{
				Latitude:  lat,
				Longitude: lon,
			},
			Timestamp: getStr(doc.Fields, "timestamp"),
			Region:    getStr(doc.Fields, "region"),
		})
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (f *FirestoreRepository) GetIncidentByID(ctx context.Context, id string) (*model.IncidentDto, error) {
	doc, err := f.getDoc(ctx, "incidents", id)
	if err != nil {
		return nil, err
	}
	lat, lon := 0.0, 0.0
	if locVal := doc.Fields["location"]; locVal.GeoPoint != nil {
		lat = locVal.GeoPoint.Latitude
		lon = locVal.GeoPoint.Longitude
	}

	return &model.IncidentDto{
		Id:                 id,
		Species:            getStr(doc.Fields, "species"),
		Category:           getStr(doc.Fields, "category"),
		VerificationStatus: getStr(doc.Fields, "verificationStatus"),
		Location: model.GeoLocationDto{
			Latitude:  lat,
			Longitude: lon,
		},
		Timestamp: getStr(doc.Fields, "timestamp"),
		Region:    getStr(doc.Fields, "region"),
	}, nil
}

func (f *FirestoreRepository) SaveIncident(ctx context.Context, inc model.IncidentDto) error {
	id := inc.Id
	if id == "" {
		id = fmt.Sprintf("inc-%d", time.Now().UnixNano())
	}

	fields := map[string]firestoreValue{
		"species":            strVal(inc.Species),
		"category":           strVal(inc.Category),
		"verificationStatus": strVal(inc.VerificationStatus),
		"timestamp":          strVal(inc.Timestamp),
		"region":             strVal(inc.Region),
		"location": firestoreValue{
			GeoPoint: &firestoreGeoPointValue{
				Latitude:  inc.Location.Latitude,
				Longitude: inc.Location.Longitude,
			},
		},
	}
	return f.saveDoc(ctx, "incidents", id, fields)
}

func (f *FirestoreRepository) UpdateIncidentVerification(ctx context.Context, id string, status string) error {
	fields := map[string]firestoreValue{
		"verificationStatus": strVal(status),
	}
	return f.saveDoc(ctx, "incidents", id, fields)
}

// RoadkillRepository methods
func (f *FirestoreRepository) GetRoadkills(ctx context.Context, statusFilter string) ([]model.RoadkillRecordDto, error) {
	docs, err := f.listDocs(ctx, "roadkill")
	if err != nil {
		return nil, err
	}

	results := make([]model.RoadkillRecordDto, 0, len(docs))
	for _, doc := range docs {
		id := doc.Name[strings.LastIndex(doc.Name, "/")+1:]
		status := getStr(doc.Fields, "clearanceStatus")
		if statusFilter != "" && status != statusFilter {
			continue
		}

		lat, lon := 0.0, 0.0
		if locVal := doc.Fields["location"]; locVal.GeoPoint != nil {
			lat = locVal.GeoPoint.Latitude
			lon = locVal.GeoPoint.Longitude
		}

		results = append(results, model.RoadkillRecordDto{
			Id:                    id,
			IncidentId:            getStr(doc.Fields, "incidentId"),
			Species:               getStr(doc.Fields, "species"),
			DecayState:            getStr(doc.Fields, "decayState"),
			ClearanceStatus:       status,
			AttractantThreatScore: getDbl(doc.Fields, "attractantThreatScore"),
			Location: model.GeoLocationDto{
				Latitude:  lat,
				Longitude: lon,
			},
			ReportedAt: getStr(doc.Fields, "reportedAt"),
			ClearedAt:  getStr(doc.Fields, "clearedAt"),
		})
	}
	return results, nil
}

func (f *FirestoreRepository) GetRoadkillByID(ctx context.Context, id string) (*model.RoadkillRecordDto, error) {
	doc, err := f.getDoc(ctx, "roadkill", id)
	if err != nil {
		return nil, err
	}

	lat, lon := 0.0, 0.0
	if locVal := doc.Fields["location"]; locVal.GeoPoint != nil {
		lat = locVal.GeoPoint.Latitude
		lon = locVal.GeoPoint.Longitude
	}

	return &model.RoadkillRecordDto{
		Id:                    id,
		IncidentId:            getStr(doc.Fields, "incidentId"),
		Species:               getStr(doc.Fields, "species"),
		DecayState:            getStr(doc.Fields, "decayState"),
		ClearanceStatus:       getStr(doc.Fields, "clearanceStatus"),
		AttractantThreatScore: getDbl(doc.Fields, "attractantThreatScore"),
		Location: model.GeoLocationDto{
			Latitude:  lat,
			Longitude: lon,
		},
		ReportedAt: getStr(doc.Fields, "reportedAt"),
		ClearedAt:  getStr(doc.Fields, "clearedAt"),
	}, nil
}

func (f *FirestoreRepository) SaveRoadkill(ctx context.Context, rk model.RoadkillRecordDto) error {
	id := rk.Id
	if id == "" {
		id = fmt.Sprintf("rk-%d", time.Now().UnixNano())
	}

	fields := map[string]firestoreValue{
		"incidentId":            strVal(rk.IncidentId),
		"species":               strVal(rk.Species),
		"decayState":            strVal(rk.DecayState),
		"clearanceStatus":       strVal(rk.ClearanceStatus),
		"attractantThreatScore": dblVal(rk.AttractantThreatScore),
		"reportedAt":            strVal(rk.ReportedAt),
		"location": firestoreValue{
			GeoPoint: &firestoreGeoPointValue{
				Latitude:  rk.Location.Latitude,
				Longitude: rk.Location.Longitude,
			},
		},
	}
	return f.saveDoc(ctx, "roadkill", id, fields)
}

func (f *FirestoreRepository) UpdateRoadkillStatus(ctx context.Context, id string, status string, notes string) (*model.RoadkillRecordDto, error) {
	fields := map[string]firestoreValue{
		"clearanceStatus": strVal(status),
	}
	if status == "CLEARED" {
		fields["clearedAt"] = strVal(time.Now().UTC().Format(time.RFC3339))
	}
	if err := f.saveDoc(ctx, "roadkill", id, fields); err != nil {
		return nil, err
	}
	return f.GetRoadkillByID(ctx, id)
}

// GeofenceRepository methods
func (f *FirestoreRepository) GetActiveGeofences(ctx context.Context) ([]model.ActiveGeofenceZone, error) {
	docs, err := f.listDocs(ctx, "geofences")
	if err != nil {
		return nil, err
	}

	results := make([]model.ActiveGeofenceZone, 0, len(docs))
	for _, doc := range docs {
		id := doc.Name[strings.LastIndex(doc.Name, "/")+1:]
		results = append(results, model.ActiveGeofenceZone{
			Id:                id,
			Tier:              getStr(doc.Fields, "tier"),
			RadiusMeters:      getDbl(doc.Fields, "radiusMeters"),
			CausalExplanation: getStr(doc.Fields, "causalExplanation"),
			ActiveRiskScore:   getDbl(doc.Fields, "activeRiskScore"),
			ValidFrom:         getStr(doc.Fields, "validFrom"),
			ValidUntil:        getStr(doc.Fields, "validUntil"),
			GeneratedAt:       doc.CreateTime,
		})
	}
	return results, nil
}

func (f *FirestoreRepository) SaveGeofences(ctx context.Context, geofences []model.ActiveGeofenceZone) error {
	for _, gf := range geofences {
		id := gf.Id
		if id == "" {
			id = fmt.Sprintf("gf-%s-%d", gf.Tier, time.Now().UnixNano())
		}
		fields := map[string]firestoreValue{
			"tier":              strVal(gf.Tier),
			"radiusMeters":      dblVal(gf.RadiusMeters),
			"causalExplanation": strVal(gf.CausalExplanation),
			"activeRiskScore":   dblVal(gf.ActiveRiskScore),
			"validFrom":         strVal(gf.ValidFrom),
			"validUntil":        strVal(gf.ValidUntil),
		}
		if err := f.saveDoc(ctx, "geofences", id, fields); err != nil {
			return err
		}
	}
	return nil
}

// AttractantRepository methods
func (f *FirestoreRepository) GetPriorityKakiTrees(ctx context.Context, municipalityCode string, limit int) ([]model.PriorityKakiTreeDto, error) {
	docs, err := f.listDocs(ctx, "kaki_trees")
	if err != nil {
		return nil, err
	}

	results := make([]model.PriorityKakiTreeDto, 0, len(docs))
	for _, doc := range docs {
		id := doc.Name[strings.LastIndex(doc.Name, "/")+1:]
		results = append(results, model.PriorityKakiTreeDto{
			TreeId:                         id,
			HazardScore:                    getDbl(doc.Fields, "hazardScore"),
			DistanceToSchoolRouteMeters:    int(getDbl(doc.Fields, "distanceToSchoolRouteMeters")),
			EstimatedBearAttractantRadiusM: int(getDbl(doc.Fields, "estimatedBearAttractantRadiusM")),
		})
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (f *FirestoreRepository) RecordKakiAction(ctx context.Context, treeID string, actionType string, notes string) error {
	fields := map[string]firestoreValue{
		"managementStatus": strVal(actionType),
		"actionNotes":      strVal(notes),
		"updatedAt":        strVal(time.Now().UTC().Format(time.RFC3339)),
	}
	return f.saveDoc(ctx, "kaki_trees", treeID, fields)
}
