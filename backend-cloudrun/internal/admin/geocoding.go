package admin

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// LRUGeocodeCache provides an in-memory thread-safe LRU cache for resolved coordinates.
type geocodeCacheEntry struct {
	key       string
	placeName string
}

type LRUGeocodeCache struct {
	mu        sync.Mutex
	capacity  int
	items     map[string]*list.Element
	evictList *list.List
}

func NewLRUGeocodeCache(capacity int) *LRUGeocodeCache {
	return &LRUGeocodeCache{
		capacity:  capacity,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
}

func (c *LRUGeocodeCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		return elem.Value.(*geocodeCacheEntry).placeName, true
	}
	return "", false
}

func (c *LRUGeocodeCache) Set(key, placeName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.evictList.MoveToFront(elem)
		elem.Value.(*geocodeCacheEntry).placeName = placeName
		return
	}
	if c.evictList.Len() >= c.capacity {
		oldest := c.evictList.Back()
		if oldest != nil {
			c.evictList.Remove(oldest)
			delete(c.items, oldest.Value.(*geocodeCacheEntry).key)
		}
	}
	entry := &geocodeCacheEntry{key: key, placeName: placeName}
	elem := c.evictList.PushFront(entry)
	c.items[key] = elem
}

func (c *LRUGeocodeCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*list.Element)
	c.evictList.Init()
}

var (
	geocodingHTTPClient = &http.Client{Timeout: 2500 * time.Millisecond}
	globalGeocodeCache  = NewLRUGeocodeCache(1000)
)

// ReverseGeocode resolves geographic coordinates (lat, lng) to a descriptive, human-readable place name.
func ReverseGeocode(lat, lng float64) string {
	return ReverseGeocodeWithContext(context.Background(), lat, lng)
}

// ReverseGeocodeWithContext resolves geographic coordinates with context timeout and fallback hierarchy.
func ReverseGeocodeWithContext(ctx context.Context, lat, lng float64) string {
	cacheKey := fmt.Sprintf("%.4f,%.4f", lat, lng)
	if cached, ok := globalGeocodeCache.Get(cacheKey); ok {
		return cached
	}

	// 1. High-priority known operational conservation sectors (Fast zero-cost heuristic)
	if heuristic := heuristicReverseGeocode(lat, lng); heuristic != "" {
		globalGeocodeCache.Set(cacheKey, heuristic)
		return heuristic
	}

	// 2. Google Maps Geocoding API if key configured or custom base URL provided
	apiKey := os.Getenv("GEOCODING_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_MAPS_API_KEY")
	}
	baseURL := os.Getenv("GEOCODING_BASE_URL")

	if apiKey != "" || strings.HasPrefix(baseURL, "http") {
		if place, err := queryGoogleGeocoding(ctx, lat, lng, apiKey, baseURL); err == nil && place != "" {
			globalGeocodeCache.Set(cacheKey, place)
			return place
		}
	}

	// 3. OpenStreetMap Nominatim API (secondary public geocoder) if not explicitly disabled
	if os.Getenv("DISABLE_EXTERNAL_GEOCODING") != "true" && os.Getenv("ENV") != "test" {
		if place, err := queryNominatimGeocoding(ctx, lat, lng); err == nil && place != "" {
			globalGeocodeCache.Set(cacheKey, place)
			return place
		}
	}

	// 4. Fallback formatted sector coordinate string
	fallback := fmt.Sprintf("Sector (Lat %.4f, Lng %.4f)", lat, lng)
	globalGeocodeCache.Set(cacheKey, fallback)
	return fallback
}

// heuristicReverseGeocode checks the 8 designated high-risk regional wildlife sectors.
func heuristicReverseGeocode(lat, lng float64) string {
	// 1. Nara, Japan (Asian Black Bear & Persimmon Attractant)
	if lat >= 34.30 && lat <= 34.85 && lng >= 135.60 && lng <= 136.15 {
		if lat >= 34.66 && lat <= 34.72 && lng >= 135.80 && lng <= 135.88 {
			return "Kasugayama Primeval Forest, Nara"
		}
		if lat >= 34.35 && lat <= 34.55 {
			return "Yoshino Mountain Foothills, Nara"
		}
		return "Nara Forest Fringe, Nara"
	}

	// 2. Nagano, Japan (Alps Bear Corridor)
	if lat >= 35.80 && lat <= 37.00 && lng >= 137.40 && lng <= 138.80 {
		if lat >= 36.15 && lat <= 36.35 && lng >= 137.85 && lng <= 138.10 {
			return "Matsumoto Highlands, Nagano"
		}
		return "Nagano Alpine Sector"
	}

	// 3. Iwate / Morioka, Japan (Persimmon Attractant & Bear Corridors)
	if lat >= 39.40 && lat <= 40.20 && lng >= 140.80 && lng <= 141.80 {
		if lat >= 39.65 && lat <= 39.75 && lng >= 141.10 && lng <= 141.22 {
			return "Morioka Basin Route, Iwate"
		}
		return "Iwate Foothills Corridor"
	}

	// 4. Hokkaido, Japan (Brown Bear Habitats)
	if lat >= 41.50 && lat <= 45.60 && lng >= 139.50 && lng <= 145.90 {
		if lat >= 43.80 && lat <= 44.50 && lng >= 144.80 && lng <= 145.40 {
			return "Shiretoko Peninsula Coast, Hokkaido"
		}
		return "Hokkaido Wilderness Perimeter"
	}

	// 5. Thailand - Kaeng Krachan & Phetchaburi (Asian Elephant Corridor & Route 3219)
	if lat >= 12.50 && lat <= 13.50 && lng >= 99.10 && lng <= 100.20 {
		if lat >= 12.80 && lat <= 12.92 && lng >= 99.60 && lng <= 99.72 {
			return "Rural Route 3219, Kaeng Krachan"
		}
		return "Kaeng Krachan Eastern Buffer"
	}

	// 6. Thailand - Khao Ang Rue Nai & Chachoengsao (Elephant Crop Raiding)
	if lat >= 13.20 && lat <= 13.80 && lng >= 101.20 && lng <= 101.90 {
		return "Khao Ang Rue Nai Corridor"
	}

	// 7. Indonesia - Riau & Sumatra (Sumatran Tiger Reserve)
	if lat >= -1.00 && lat <= 2.50 && lng >= 100.00 && lng <= 103.50 {
		if lat >= 0.70 && lat <= 1.00 && lng >= 101.30 && lng <= 101.80 {
			return "Riau Riparian Corridor, Sumatra"
		}
		return "Sumatra Tiger Reserve Perimeter"
	}

	// 8. India - Kaziranga & Assam (Elephant & Bengal Tiger)
	if lat >= 26.30 && lat <= 26.90 && lng >= 92.80 && lng <= 93.80 {
		return "Kaziranga Buffer Zone, Assam"
	}

	return ""
}

type googleGeocodeResponse struct {
	Status  string `json:"status"`
	Results []struct {
		FormattedAddress string `json:"formatted_address"`
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
	} `json:"results"`
}

func queryGoogleGeocoding(ctx context.Context, lat, lng float64, apiKey, customBaseURL string) (string, error) {
	apiEndpoint := "https://maps.googleapis.com/maps/api/geocode/json"
	if customBaseURL != "" {
		apiEndpoint = customBaseURL
	}

	var reqURL string
	if apiKey != "" {
		reqURL = fmt.Sprintf("%s?latlng=%.6f,%.6f&key=%s&language=en", apiEndpoint, lat, lng, apiKey)
	} else {
		reqURL = fmt.Sprintf("%s?latlng=%.6f,%.6f&language=en", apiEndpoint, lat, lng)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := geocodingHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google geocoding returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsed googleGeocodeResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}

	if (parsed.Status != "OK" && parsed.Status != "") || len(parsed.Results) == 0 {
		return "", fmt.Errorf("google geocoding status: %s", parsed.Status)
	}

	res := parsed.Results[0]
	var route, locality, adminArea, country string
	for _, c := range res.AddressComponents {
		for _, t := range c.Types {
			switch t {
			case "route":
				route = c.LongName
			case "locality", "postal_town", "sublocality":
				if locality == "" {
					locality = c.LongName
				}
			case "administrative_area_level_1":
				adminArea = c.LongName
			case "country":
				country = c.LongName
			}
		}
	}

	if route != "" && locality != "" {
		return fmt.Sprintf("%s, %s", route, locality), nil
	}
	if locality != "" && adminArea != "" {
		return fmt.Sprintf("%s, %s", locality, adminArea), nil
	}
	if res.FormattedAddress != "" {
		parts := strings.Split(res.FormattedAddress, ", ")
		if len(parts) > 3 {
			return strings.Join(parts[:3], ", "), nil
		}
		return res.FormattedAddress, nil
	}
	if country != "" {
		return country, nil
	}

	return "", fmt.Errorf("no address details extracted")
}

type nominatimResponse struct {
	DisplayName string `json:"display_name"`
	Address     struct {
		Road    string `json:"road"`
		Village string `json:"village"`
		Suburb  string `json:"suburb"`
		City    string `json:"city"`
		County  string `json:"county"`
		State   string `json:"state"`
		Country string `json:"country"`
	} `json:"address"`
}

func queryNominatimGeocoding(ctx context.Context, lat, lng float64) (string, error) {
	reqURL := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=json&lat=%.6f&lon=%.6f&zoom=14&addressdetails=1", lat, lng)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "CoExistShieldAI-ReverseGeocoder/1.0 (coexist-shield-ai@example.org)")

	resp, err := geocodingHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nominatim returned HTTP %d", resp.StatusCode)
	}

	var parsed nominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}

	loc := parsed.Address.City
	if loc == "" {
		loc = parsed.Address.Village
	}
	if loc == "" {
		loc = parsed.Address.Suburb
	}

	if parsed.Address.Road != "" && loc != "" {
		return fmt.Sprintf("%s, %s", parsed.Address.Road, loc), nil
	}
	if loc != "" && parsed.Address.State != "" {
		return fmt.Sprintf("%s, %s", loc, parsed.Address.State), nil
	}
	if parsed.DisplayName != "" {
		parts := strings.Split(parsed.DisplayName, ", ")
		if len(parts) > 3 {
			return strings.Join(parts[:3], ", "), nil
		}
		return parsed.DisplayName, nil
	}

	return "", fmt.Errorf("no address details extracted from nominatim")
}

// EnhanceGeofenceNameWithLocation ensures a geofence name includes a reverse-geocoded place name.
func EnhanceGeofenceNameWithLocation(baseName string, species string, lat, lng float64) string {
	place := ReverseGeocode(lat, lng)
	trimmed := strings.TrimSpace(baseName)

	if trimmed == "" || trimmed == "Manual Geofence" {
		if species != "" {
			return fmt.Sprintf("Buffer: %s (%s)", species, place)
		}
		return fmt.Sprintf("Zone: %s", place)
	}

	// If place name is already present, return as is
	if strings.Contains(trimmed, place) {
		return trimmed
	}

	// If it doesn't already have a location tag in parentheses, append the place name
	if !strings.Contains(trimmed, "(") {
		return fmt.Sprintf("%s (%s)", trimmed, place)
	}

	return trimmed
}
