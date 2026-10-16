package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	ColAdminWildlife      = "admin_wildlife"
	ColAdminGeofences     = "admin_geofences"
	ColAdminRoadkill      = "admin_roadkill"
	ColAdminEnvironmental = "admin_environmental"
	ColAdminWeather       = "admin_weather"
)

// FirestorePersistence manages persistent storage of admin datasets in Cloud Firestore via REST API.
type FirestorePersistence struct {
	projectID  string
	baseURL    string
	httpClient *http.Client
}

// NewFirestorePersistence initializes a Firestore client for admin data persistence.
func NewFirestorePersistence(projectID string) *FirestorePersistence {
	emulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST")
	var baseURL string
	if emulatorHost != "" {
		baseURL = fmt.Sprintf("http://%s/v1/projects/%s/databases/(default)/documents", emulatorHost, projectID)
	} else {
		baseURL = fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/(default)/documents", projectID)
	}

	return &FirestorePersistence{
		projectID: projectID,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Low-level Firestore REST helpers
type fsValue struct {
	StringValue *string  `json:"stringValue,omitempty"`
	DoubleValue *float64 `json:"doubleValue,omitempty"`
	BooleanValue *bool   `json:"booleanValue,omitempty"`
}

type fsDocument struct {
	Name       string             `json:"name"`
	Fields     map[string]fsValue `json:"fields"`
	CreateTime string             `json:"createTime,omitempty"`
	UpdateTime string             `json:"updateTime,omitempty"`
}

type fsListResponse struct {
	Documents []fsDocument `json:"documents"`
}

var (
	cachedAdminToken   string
	cachedAdminTokenAt time.Time
	tokenLock          sync.Mutex
)

func (p *FirestorePersistence) setAuth(req *http.Request) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") != "" || strings.Contains(p.baseURL, "127.0.0.1") || strings.Contains(p.baseURL, "localhost") {
		return
	}
	token, err := getAdminGCPToken()
	if err == nil && token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func getAdminGCPToken() (string, error) {
	tokenLock.Lock()
	defer tokenLock.Unlock()

	if cachedAdminToken != "" && time.Since(cachedAdminTokenAt) < 45*time.Minute {
		return cachedAdminToken, nil
	}

	if t := os.Getenv("GCP_ACCESS_TOKEN"); t != "" {
		cachedAdminToken = strings.TrimSpace(t)
		cachedAdminTokenAt = time.Now()
		return cachedAdminToken, nil
	}
	if t := os.Getenv("GCS_ACCESS_TOKEN"); t != "" {
		cachedAdminToken = strings.TrimSpace(t)
		cachedAdminTokenAt = time.Now()
		return cachedAdminToken, nil
	}
	if t := os.Getenv("FIREBASE_ACCESS_TOKEN"); t != "" {
		cachedAdminToken = strings.TrimSpace(t)
		cachedAdminTokenAt = time.Now()
		return cachedAdminToken, nil
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
					cachedAdminToken = tokenResp.AccessToken
					cachedAdminTokenAt = time.Now()
					return cachedAdminToken, nil
				}
			}
		}
	}

	// Local gcloud auth fallback
	if out, err := exec.Command("gcloud", "auth", "print-access-token").Output(); err == nil {
		tok := strings.TrimSpace(string(out))
		if tok != "" {
			cachedAdminToken = tok
			cachedAdminTokenAt = time.Now()
			return cachedAdminToken, nil
		}
	}

	return "", fmt.Errorf("no access token available")
}

func (p *FirestorePersistence) saveDocument(ctx context.Context, collection string, docID string, data interface{}, metaFields map[string]string) error {
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal %s/%s: %w", collection, docID, err)
	}
	payloadStr := string(payloadBytes)

	fields := make(map[string]fsValue)
	fields["payload"] = fsValue{StringValue: &payloadStr}
	nowStr := time.Now().UTC().Format(time.RFC3339)
	fields["updated_at"] = fsValue{StringValue: &nowStr}

	for k, v := range metaFields {
		val := v
		fields[k] = fsValue{StringValue: &val}
	}

	doc := fsDocument{Fields: fields}
	docBytes, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s/%s", p.baseURL, collection, docID)
	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(docBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	p.setAuth(req)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("firestore patch request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("firestore patch %s/%s returned %d: %s", collection, docID, resp.StatusCode, string(body))
	}
	return nil
}

func (p *FirestorePersistence) deleteDocument(ctx context.Context, collection string, docID string) error {
	url := fmt.Sprintf("%s/%s/%s", p.baseURL, collection, docID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	p.setAuth(req)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("firestore delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("firestore delete %s/%s returned %d: %s", collection, docID, resp.StatusCode, string(body))
	}
	return nil
}

func (p *FirestorePersistence) listDocuments(ctx context.Context, collection string) ([]fsDocument, error) {
	url := fmt.Sprintf("%s/%s", p.baseURL, collection)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	p.setAuth(req)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("firestore list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []fsDocument{}, nil
	}
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("firestore list %s returned %d: %s", collection, resp.StatusCode, string(body))
	}

	var listResp fsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}
	return listResp.Documents, nil
}

// --- Specific dataset persistence methods ---

// Wildlife Incidents
func (p *FirestorePersistence) SaveWildlife(ctx context.Context, w AdminWildlifeIncident) error {
	meta := map[string]string{
		"id":          w.ID,
		"species":     w.Species,
		"threat_level": w.ThreatLevel,
		"status":      w.Status,
	}
	return p.saveDocument(ctx, ColAdminWildlife, w.ID, w, meta)
}

func (p *FirestorePersistence) DeleteWildlife(ctx context.Context, id string) error {
	return p.deleteDocument(ctx, ColAdminWildlife, id)
}

func (p *FirestorePersistence) LoadWildlife(ctx context.Context) ([]AdminWildlifeIncident, error) {
	docs, err := p.listDocuments(ctx, ColAdminWildlife)
	if err != nil {
		return nil, err
	}
	results := make([]AdminWildlifeIncident, 0, len(docs))
	for _, doc := range docs {
		if v, ok := doc.Fields["payload"]; ok && v.StringValue != nil {
			var w AdminWildlifeIncident
			if err := json.Unmarshal([]byte(*v.StringValue), &w); err == nil {
				results = append(results, w)
			}
		}
	}
	return results, nil
}

// Geofences
func (p *FirestorePersistence) SaveGeofence(ctx context.Context, g AdminGeofence) error {
	meta := map[string]string{
		"id":           g.ID,
		"name":         g.Name,
		"source":       g.Source,
		"threat_level": g.ThreatLevel,
	}
	return p.saveDocument(ctx, ColAdminGeofences, g.ID, g, meta)
}

func (p *FirestorePersistence) DeleteGeofence(ctx context.Context, id string) error {
	return p.deleteDocument(ctx, ColAdminGeofences, id)
}

func (p *FirestorePersistence) LoadGeofences(ctx context.Context) ([]AdminGeofence, error) {
	docs, err := p.listDocuments(ctx, ColAdminGeofences)
	if err != nil {
		return nil, err
	}
	results := make([]AdminGeofence, 0, len(docs))
	for _, doc := range docs {
		if v, ok := doc.Fields["payload"]; ok && v.StringValue != nil {
			var g AdminGeofence
			if err := json.Unmarshal([]byte(*v.StringValue), &g); err == nil {
				results = append(results, g)
			}
		}
	}
	return results, nil
}

// Roadkill Records
func (p *FirestorePersistence) SaveRoadkill(ctx context.Context, rk AdminRoadkillRecord) error {
	meta := map[string]string{
		"id":      rk.ID,
		"species": rk.Species,
		"status":  rk.Status,
	}
	return p.saveDocument(ctx, ColAdminRoadkill, rk.ID, rk, meta)
}

func (p *FirestorePersistence) DeleteRoadkill(ctx context.Context, id string) error {
	return p.deleteDocument(ctx, ColAdminRoadkill, id)
}

func (p *FirestorePersistence) LoadRoadkills(ctx context.Context) ([]AdminRoadkillRecord, error) {
	docs, err := p.listDocuments(ctx, ColAdminRoadkill)
	if err != nil {
		return nil, err
	}
	results := make([]AdminRoadkillRecord, 0, len(docs))
	for _, doc := range docs {
		if v, ok := doc.Fields["payload"]; ok && v.StringValue != nil {
			var rk AdminRoadkillRecord
			if err := json.Unmarshal([]byte(*v.StringValue), &rk); err == nil {
				results = append(results, rk)
			}
		}
	}
	return results, nil
}

// Environmental Data
func (p *FirestorePersistence) SaveEnvironmental(ctx context.Context, env AdminEnvironmentalData) error {
	meta := map[string]string{
		"id":       env.ID,
		"category": env.Category,
		"name":     env.Name,
	}
	return p.saveDocument(ctx, ColAdminEnvironmental, env.ID, env, meta)
}

func (p *FirestorePersistence) DeleteEnvironmental(ctx context.Context, id string) error {
	return p.deleteDocument(ctx, ColAdminEnvironmental, id)
}

func (p *FirestorePersistence) LoadEnvironmental(ctx context.Context) ([]AdminEnvironmentalData, error) {
	docs, err := p.listDocuments(ctx, ColAdminEnvironmental)
	if err != nil {
		return nil, err
	}
	results := make([]AdminEnvironmentalData, 0, len(docs))
	for _, doc := range docs {
		if v, ok := doc.Fields["payload"]; ok && v.StringValue != nil {
			var env AdminEnvironmentalData
			if err := json.Unmarshal([]byte(*v.StringValue), &env); err == nil {
				results = append(results, env)
			}
		}
	}
	return results, nil
}

// Weather Data
func (p *FirestorePersistence) SaveWeather(ctx context.Context, w AdminWeatherData) error {
	meta := map[string]string{
		"id":           w.ID,
		"station_name": w.StationName,
		"region":       w.Region,
	}
	return p.saveDocument(ctx, ColAdminWeather, w.ID, w, meta)
}

func (p *FirestorePersistence) DeleteWeather(ctx context.Context, id string) error {
	return p.deleteDocument(ctx, ColAdminWeather, id)
}

func (p *FirestorePersistence) LoadWeather(ctx context.Context) ([]AdminWeatherData, error) {
	docs, err := p.listDocuments(ctx, ColAdminWeather)
	if err != nil {
		return nil, err
	}
	results := make([]AdminWeatherData, 0, len(docs))
	for _, doc := range docs {
		if v, ok := doc.Fields["payload"]; ok && v.StringValue != nil {
			var w AdminWeatherData
			if err := json.Unmarshal([]byte(*v.StringValue), &w); err == nil {
				results = append(results, w)
			}
		}
	}
	return results, nil
}

// RestoreOrSeed synchronizes between Firestore and the in-memory Store.
// If Firestore has existing data, it populates the Store with Firestore records (preserving user modifications across deployments).
// If Firestore is empty (first deployment), it writes the default seed records to Firestore so future instances load them.
func (p *FirestorePersistence) RestoreOrSeed(ctx context.Context, s *Store) error {
	log.Printf("[FirestorePersistence] Restoring admin datasets from project %s...", p.projectID)

	// 1. Wildlife
	wildlifeList, err := p.LoadWildlife(ctx)
	if err != nil {
		log.Printf("[FirestorePersistence] Warning: Failed to load wildlife from Firestore: %v", err)
	} else if len(wildlifeList) > 0 {
		s.mu.Lock()
		s.wildlife = make(map[string]AdminWildlifeIncident)
		for _, w := range wildlifeList {
			s.wildlife[w.ID] = w
		}
		s.mu.Unlock()
		log.Printf("[FirestorePersistence] Restored %d wildlife incidents from Firestore", len(wildlifeList))
	} else {
		// Seed default wildlife to Firestore
		s.mu.RLock()
		for _, w := range s.wildlife {
			_ = p.SaveWildlife(ctx, w)
		}
		s.mu.RUnlock()
		log.Printf("[FirestorePersistence] Seeded default wildlife incidents to Firestore")
	}

	// 2. Geofences
	geoList, err := p.LoadGeofences(ctx)
	if err != nil {
		log.Printf("[FirestorePersistence] Warning: Failed to load geofences from Firestore: %v", err)
	} else if len(geoList) > 0 {
		s.mu.Lock()
		s.geofences = make(map[string]AdminGeofence)
		for _, g := range geoList {
			s.geofences[g.ID] = g
		}
		s.mu.Unlock()
		log.Printf("[FirestorePersistence] Restored %d geofences from Firestore", len(geoList))
	} else {
		s.mu.RLock()
		for _, g := range s.geofences {
			_ = p.SaveGeofence(ctx, g)
		}
		s.mu.RUnlock()
		log.Printf("[FirestorePersistence] Seeded default geofences to Firestore")
	}

	// 3. Roadkill
	rkList, err := p.LoadRoadkills(ctx)
	if err != nil {
		log.Printf("[FirestorePersistence] Warning: Failed to load roadkills from Firestore: %v", err)
	} else if len(rkList) > 0 {
		s.mu.Lock()
		s.roadkills = make(map[string]AdminRoadkillRecord)
		for _, rk := range rkList {
			s.roadkills[rk.ID] = rk
		}
		s.mu.Unlock()
		log.Printf("[FirestorePersistence] Restored %d roadkill records from Firestore", len(rkList))
	} else {
		s.mu.RLock()
		for _, rk := range s.roadkills {
			_ = p.SaveRoadkill(ctx, rk)
		}
		s.mu.RUnlock()
		log.Printf("[FirestorePersistence] Seeded default roadkill records to Firestore")
	}

	// 4. Environmental
	envList, err := p.LoadEnvironmental(ctx)
	if err != nil {
		log.Printf("[FirestorePersistence] Warning: Failed to load environmental data from Firestore: %v", err)
	} else if len(envList) > 0 {
		s.mu.Lock()
		s.environmental = make(map[string]AdminEnvironmentalData)
		for _, env := range envList {
			s.environmental[env.ID] = env
		}
		s.mu.Unlock()
		log.Printf("[FirestorePersistence] Restored %d environmental features from Firestore", len(envList))
	} else {
		s.mu.RLock()
		for _, env := range s.environmental {
			_ = p.SaveEnvironmental(ctx, env)
		}
		s.mu.RUnlock()
		log.Printf("[FirestorePersistence] Seeded default environmental data to Firestore")
	}

	// 5. Weather
	weatherList, err := p.LoadWeather(ctx)
	if err != nil {
		log.Printf("[FirestorePersistence] Warning: Failed to load weather data from Firestore: %v", err)
	} else if len(weatherList) > 0 {
		s.mu.Lock()
		s.weather = make(map[string]AdminWeatherData)
		for _, w := range weatherList {
			s.weather[w.ID] = w
		}
		s.mu.Unlock()
		log.Printf("[FirestorePersistence] Restored %d weather stations from Firestore", len(weatherList))
	} else {
		s.mu.RLock()
		for _, w := range s.weather {
			_ = p.SaveWeather(ctx, w)
		}
		s.mu.RUnlock()
		log.Printf("[FirestorePersistence] Seeded default weather data to Firestore")
	}

	return nil
}
