package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"ai.coexistshield/backend/internal/model"
)

type FirestoreRepository struct {
	projectID  string
	baseURL    string
	httpClient *http.Client
}

var (
	cachedRepoToken   string
	cachedRepoTokenAt time.Time
	repoTokenLock     sync.Mutex
)

func getRepoGCPToken() (string, error) {
	repoTokenLock.Lock()
	defer repoTokenLock.Unlock()

	if cachedRepoToken != "" && time.Since(cachedRepoTokenAt) < 45*time.Minute {
		return cachedRepoToken, nil
	}

	if t := os.Getenv("GCP_ACCESS_TOKEN"); t != "" {
		cachedRepoToken = strings.TrimSpace(t)
		cachedRepoTokenAt = time.Now()
		return cachedRepoToken, nil
	}
	if t := os.Getenv("GCS_ACCESS_TOKEN"); t != "" {
		cachedRepoToken = strings.TrimSpace(t)
		cachedRepoTokenAt = time.Now()
		return cachedRepoToken, nil
	}
	if t := os.Getenv("FIREBASE_ACCESS_TOKEN"); t != "" {
		cachedRepoToken = strings.TrimSpace(t)
		cachedRepoTokenAt = time.Now()
		return cachedRepoToken, nil
	}

	// Cloud Run metadata server
	client := &http.Client{Timeout: 800 * time.Millisecond}
	req, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token", nil)
	if err == nil {
		req.Header.Set("Metadata-Flavor", "Google")
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var tokenResp struct {
					AccessToken string `json:"access_token"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err == nil && tokenResp.AccessToken != "" {
					cachedRepoToken = tokenResp.AccessToken
					cachedRepoTokenAt = time.Now()
					return cachedRepoToken, nil
				}
			}
		}
	}

	// Local gcloud auth fallback
	if out, err := exec.Command("gcloud", "auth", "print-access-token").Output(); err == nil {
		tok := strings.TrimSpace(string(out))
		if tok != "" {
			cachedRepoToken = tok
			cachedRepoTokenAt = time.Now()
			return cachedRepoToken, nil
		}
	}

	return "", fmt.Errorf("no access token available")
}

func (f *FirestoreRepository) setAuth(req *http.Request) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") != "" || strings.Contains(f.baseURL, "127.0.0.1") || strings.Contains(f.baseURL, "localhost") {
		return
	}
	token, err := getRepoGCPToken()
	if err == nil && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
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

func boolVal(b bool) firestoreValue {
	return firestoreValue{BooleanValue: &b}
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

func getBool(fields map[string]firestoreValue, key string) bool {
	if v, ok := fields[key]; ok && v.BooleanValue != nil {
		return *v.BooleanValue
	}
	return false
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
	f.setAuth(req)

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
	f.setAuth(req)

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
	f.setAuth(req)

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

func (f *FirestoreRepository) deleteDoc(ctx context.Context, collection string, docID string) error {
	url := fmt.Sprintf("%s/%s/%s", f.baseURL, collection, docID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	f.setAuth(req)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("firestore delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("firestore delete returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
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

// SaveCropPlot persists an elephant agricultural crop attractant plot to Firestore.
func (f *FirestoreRepository) SaveCropPlot(ctx context.Context, id, name, province, parkBoundary, cropType, stage, updatedAt string, lat, lon, area, dist, aroma, maturity, risk, scentBuf float64) error {
	fields := map[string]firestoreValue{
		"name":                  strVal(name),
		"province":              strVal(province),
		"parkBoundary":          strVal(parkBoundary),
		"cropType":              strVal(cropType),
		"growthStage":           strVal(stage),
		"latitude":              dblVal(lat),
		"longitude":             dblVal(lon),
		"areaHectares":          dblVal(area),
		"distanceToForestEdgeM": dblVal(dist),
		"aromaVolatilityIndex":  dblVal(aroma),
		"maturityScore":         dblVal(maturity),
		"activeRiskScore":       dblVal(risk),
		"scentBufferMeters":     dblVal(scentBuf),
		"updatedAt":             strVal(updatedAt),
	}
	return f.saveDoc(ctx, "crop_attractants", id, fields)
}

// ListCropPlots retrieves all persisted elephant crop plots from Firestore.
func (f *FirestoreRepository) ListCropPlots(ctx context.Context) ([]map[string]interface{}, error) {
	docs, err := f.listDocs(ctx, "crop_attractants")
	if err != nil {
		return nil, err
	}
	results := make([]map[string]interface{}, 0, len(docs))
	for _, doc := range docs {
		id := doc.Name[strings.LastIndex(doc.Name, "/")+1:]
		m := map[string]interface{}{
			"id":                    id,
			"name":                  getStr(doc.Fields, "name"),
			"province":              getStr(doc.Fields, "province"),
			"parkBoundary":          getStr(doc.Fields, "parkBoundary"),
			"cropType":              getStr(doc.Fields, "cropType"),
			"growthStage":           getStr(doc.Fields, "growthStage"),
			"latitude":              getDbl(doc.Fields, "latitude"),
			"longitude":             getDbl(doc.Fields, "longitude"),
			"areaHectares":          getDbl(doc.Fields, "areaHectares"),
			"distanceToForestEdgeM": getDbl(doc.Fields, "distanceToForestEdgeM"),
			"aromaVolatilityIndex":  getDbl(doc.Fields, "aromaVolatilityIndex"),
			"maturityScore":         getDbl(doc.Fields, "maturityScore"),
			"activeRiskScore":       getDbl(doc.Fields, "activeRiskScore"),
			"scentBufferMeters":     getDbl(doc.Fields, "scentBufferMeters"),
			"updatedAt":             getStr(doc.Fields, "updatedAt"),
		}
		results = append(results, m)
	}
	return results, nil
}

// SaveLivestockEnclosure persists a pastoral enclosure / tiger kill site to Firestore.
func (f *FirestoreRepository) SaveLivestockEnclosure(ctx context.Context, id, name, province, village, livestockType, forensicNotes, updatedAt string, headCount int, hasNightCorral, isActiveKill bool, lat, lon, dist, hoursSinceKill, returnProb, threatScore, scentBuf float64) error {
	fields := map[string]firestoreValue{
		"name":                  strVal(name),
		"province":              strVal(province),
		"village":               strVal(village),
		"livestockType":         strVal(livestockType),
		"headCount":             dblVal(float64(headCount)),
		"hasNightCorral":        boolVal(hasNightCorral),
		"isActiveKillSite":      boolVal(isActiveKill),
		"latitude":              dblVal(lat),
		"longitude":             dblVal(lon),
		"distanceToForestEdgeM": dblVal(dist),
		"hoursSinceKill":        dblVal(hoursSinceKill),
		"returnProbability":     dblVal(returnProb),
		"activeThreatScore":     dblVal(threatScore),
		"scentBufferMeters":     dblVal(scentBuf),
		"forensicNotes":         strVal(forensicNotes),
		"updatedAt":             strVal(updatedAt),
	}
	return f.saveDoc(ctx, "livestock_pastures", id, fields)
}

// ListLivestockEnclosures retrieves all persisted livestock pastures / kill sites from Firestore.
func (f *FirestoreRepository) ListLivestockEnclosures(ctx context.Context) ([]map[string]interface{}, error) {
	docs, err := f.listDocs(ctx, "livestock_pastures")
	if err != nil {
		return nil, err
	}
	results := make([]map[string]interface{}, 0, len(docs))
	for _, doc := range docs {
		id := doc.Name[strings.LastIndex(doc.Name, "/")+1:]
		m := map[string]interface{}{
			"id":                    id,
			"name":                  getStr(doc.Fields, "name"),
			"province":              getStr(doc.Fields, "province"),
			"village":               getStr(doc.Fields, "village"),
			"livestockType":         getStr(doc.Fields, "livestockType"),
			"headCount":             int(getDbl(doc.Fields, "headCount")),
			"hasNightCorral":        getBool(doc.Fields, "hasNightCorral"),
			"isActiveKillSite":      getBool(doc.Fields, "isActiveKillSite"),
			"latitude":              getDbl(doc.Fields, "latitude"),
			"longitude":             getDbl(doc.Fields, "longitude"),
			"distanceToForestEdgeM": getDbl(doc.Fields, "distanceToForestEdgeM"),
			"hoursSinceKill":        getDbl(doc.Fields, "hoursSinceKill"),
			"returnProbability":     getDbl(doc.Fields, "returnProbability"),
			"activeThreatScore":     getDbl(doc.Fields, "activeThreatScore"),
			"scentBufferMeters":     getDbl(doc.Fields, "scentBufferMeters"),
			"forensicNotes":         getStr(doc.Fields, "forensicNotes"),
			"updatedAt":             getStr(doc.Fields, "updatedAt"),
		}
		results = append(results, m)
	}
	return results, nil
}

// SaveMunicipalKakiTree persists a priority Kaki tree to Firestore.
func (f *FirestoreRepository) SaveMunicipalKakiTree(ctx context.Context, treeID, address, consentStatus, permitID, managementStatus, assignedSquad, fruitLoad string, lat, lon, hazardScore float64, isTop5 bool, schoolDist, bearRadius, clearanceOrder int) error {
	fields := map[string]firestoreValue{
		"address":                        strVal(address),
		"latitude":                       dblVal(lat),
		"longitude":                      dblVal(lon),
		"hazardScore":                    dblVal(hazardScore),
		"isTop5Percent":                  boolVal(isTop5),
		"distanceToSchoolRouteMeters":    dblVal(float64(schoolDist)),
		"estimatedBearAttractantRadiusM": dblVal(float64(bearRadius)),
		"ownerConsentStatus":             strVal(consentStatus),
		"permitId":                       strVal(permitID),
		"managementStatus":               strVal(managementStatus),
		"recommendedClearanceOrder":      dblVal(float64(clearanceOrder)),
		"assignedSquad":                  strVal(assignedSquad),
		"fruitLoad":                      strVal(fruitLoad),
		"updatedAt":                      strVal(time.Now().UTC().Format(time.RFC3339)),
	}
	return f.saveDoc(ctx, "kaki_trees", treeID, fields)
}

// ListMunicipalKakiTrees retrieves all registered municipal priority Kaki trees from Firestore.
func (f *FirestoreRepository) ListMunicipalKakiTrees(ctx context.Context) ([]map[string]interface{}, error) {
	docs, err := f.listDocs(ctx, "kaki_trees")
	if err != nil {
		return nil, err
	}
	results := make([]map[string]interface{}, 0, len(docs))
	for _, doc := range docs {
		id := doc.Name[strings.LastIndex(doc.Name, "/")+1:]
		m := map[string]interface{}{
			"treeId":                         id,
			"address":                        getStr(doc.Fields, "address"),
			"latitude":                       getDbl(doc.Fields, "latitude"),
			"longitude":                      getDbl(doc.Fields, "longitude"),
			"hazardScore":                    getDbl(doc.Fields, "hazardScore"),
			"isTop5Percent":                  getBool(doc.Fields, "isTop5Percent"),
			"distanceToSchoolRouteMeters":    int(getDbl(doc.Fields, "distanceToSchoolRouteMeters")),
			"estimatedBearAttractantRadiusM": int(getDbl(doc.Fields, "estimatedBearAttractantRadiusM")),
			"ownerConsentStatus":             getStr(doc.Fields, "ownerConsentStatus"),
			"permitId":                       getStr(doc.Fields, "permitId"),
			"managementStatus":               getStr(doc.Fields, "managementStatus"),
			"recommendedClearanceOrder":      int(getDbl(doc.Fields, "recommendedClearanceOrder")),
			"assignedSquad":                  getStr(doc.Fields, "assignedSquad"),
			"fruitLoad":                      getStr(doc.Fields, "fruitLoad"),
		}
		results = append(results, m)
	}
	return results, nil
}

// SaveFellingAuditLog records an official tree felling action log to Firestore.
func (f *FirestoreRepository) SaveFellingAuditLog(ctx context.Context, logID, treeID, permitID, officerID, action string, schoolDist int, riskDrop, biomass float64) error {
	fields := map[string]firestoreValue{
		"treeId":          strVal(treeID),
		"permitId":        strVal(permitID),
		"officerId":       strVal(officerID),
		"action":          strVal(action),
		"schoolDistanceM": dblVal(float64(schoolDist)),
		"riskDropPct":     dblVal(riskDrop),
		"biomassKg":       dblVal(biomass),
		"timestamp":       strVal(time.Now().UTC().Format(time.RFC3339)),
	}
	return f.saveDoc(ctx, "felling_audit_logs", logID, fields)
}

// SaveMLModelCheckpoint archives a versioned model checkpoint and updates the latest pointer in Firestore.
func (f *FirestoreRepository) SaveMLModelCheckpoint(ctx context.Context, modelJSON []byte, version, trainedAt string, aucRoc, precTop20 float64, sampleSize int) error {
	fields := map[string]firestoreValue{
		"version":        strVal(version),
		"trainedAt":      strVal(trainedAt),
		"aucRoc":         dblVal(aucRoc),
		"precisionTop20": dblVal(precTop20),
		"sampleSize":     dblVal(float64(sampleSize)),
		"modelJSON":      strVal(string(modelJSON)),
		"updatedAt":      strVal(time.Now().UTC().Format(time.RFC3339)),
	}
	// 1. Save versioned document
	if err := f.saveDoc(ctx, "ml_models", version, fields); err != nil {
		return err
	}
	// 2. Update 'latest' pointer
	return f.saveDoc(ctx, "ml_models", "latest", fields)
}

// GetLatestMLModel fetches the most recently promoted ML model checkpoint JSON from Firestore.
func (f *FirestoreRepository) GetLatestMLModel(ctx context.Context) ([]byte, error) {
	doc, err := f.getDoc(ctx, "ml_models", "latest")
	if err != nil {
		return nil, err
	}
	raw := getStr(doc.Fields, "modelJSON")
	if raw == "" {
		return nil, fmt.Errorf("empty modelJSON in ml_models/latest")
	}
	return []byte(raw), nil
}

// SaveMLRetrainResult logs an autonomous retraining run to Firestore.
func (f *FirestoreRepository) SaveMLRetrainResult(ctx context.Context, resultJSON []byte, runID, version, trigger string, promoted bool, trainedAt string) error {
	fields := map[string]firestoreValue{
		"runId":      strVal(runID),
		"version":    strVal(version),
		"trigger":    strVal(trigger),
		"promoted":   boolVal(promoted),
		"trainedAt":  strVal(trainedAt),
		"resultJSON": strVal(string(resultJSON)),
		"updatedAt":  strVal(time.Now().UTC().Format(time.RFC3339)),
	}
	return f.saveDoc(ctx, "ml_training_history", runID, fields)
}

// ListMLRetrainHistory retrieves past retraining execution logs from Firestore.
func (f *FirestoreRepository) ListMLRetrainHistory(ctx context.Context) ([]string, error) {
	docs, err := f.listDocs(ctx, "ml_training_history")
	if err != nil {
		return nil, err
	}
	results := make([]string, 0, len(docs))
	for _, doc := range docs {
		raw := getStr(doc.Fields, "resultJSON")
		if raw != "" {
			results = append(results, raw)
		}
	}
	return results, nil
}

// SaveMLTrainingSample persists an individual training sample to Firestore.
func (f *FirestoreRepository) SaveMLTrainingSample(ctx context.Context, sampleJSON []byte, sampleID string) error {
	fields := map[string]firestoreValue{
		"sampleId":   strVal(sampleID),
		"sampleJSON": strVal(string(sampleJSON)),
		"createdAt":  strVal(time.Now().UTC().Format(time.RFC3339)),
	}
	return f.saveDoc(ctx, "ml_training_samples", sampleID, fields)
}

// ListMLTrainingSamples loads training samples from Firestore.
func (f *FirestoreRepository) ListMLTrainingSamples(ctx context.Context) ([]string, error) {
	docs, err := f.listDocs(ctx, "ml_training_samples")
	if err != nil {
		return nil, err
	}
	samples := make([]string, 0, len(docs))
	for _, doc := range docs {
		raw := getStr(doc.Fields, "sampleJSON")
		if raw != "" {
			samples = append(samples, raw)
		}
	}
	return samples, nil
}
