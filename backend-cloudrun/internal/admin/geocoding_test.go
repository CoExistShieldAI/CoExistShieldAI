package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReverseGeocode_KnownLocations(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64
		lng      float64
		expected string
	}{
		{
			name:     "Kasugayama Nara",
			lat:      34.685,
			lng:      135.832,
			expected: "Kasugayama Primeval Forest, Nara",
		},
		{
			name:     "Morioka Iwate",
			lat:      39.702,
			lng:      141.152,
			expected: "Morioka Basin Route, Iwate",
		},
		{
			name:     "Kaeng Krachan Thailand",
			lat:      12.875,
			lng:      99.650,
			expected: "Rural Route 3219, Kaeng Krachan",
		},
		{
			name:     "Riau Sumatra",
			lat:      0.825,
			lng:      101.530,
			expected: "Riau Riparian Corridor, Sumatra",
		},
		{
			name:     "Unknown coordinate sector",
			lat:      10.0,
			lng:      20.0,
			expected: "Sector (Lat 10.0000, Lng 20.0000)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReverseGeocode(tt.lat, tt.lng)
			if result != tt.expected {
				t.Errorf("ReverseGeocode(%f, %f) = %q; expected %q", tt.lat, tt.lng, result, tt.expected)
			}
		})
	}
}

func TestEnhanceGeofenceNameWithLocation(t *testing.T) {
	// 1. Empty name should be populated with species and place
	name1 := EnhanceGeofenceNameWithLocation("", "Asian Black Bear", 34.685, 135.832)
	if !strings.Contains(name1, "Kasugayama Primeval Forest, Nara") {
		t.Errorf("Expected place in name, got %q", name1)
	}
	if !strings.Contains(name1, "Asian Black Bear") {
		t.Errorf("Expected species in name, got %q", name1)
	}

	// 2. Default "Manual Geofence" name should be replaced with place name
	name2 := EnhanceGeofenceNameWithLocation("Manual Geofence", "Asian Elephant", 12.875, 99.650)
	if !strings.Contains(name2, "Kaeng Krachan") {
		t.Errorf("Expected Kaeng Krachan in name, got %q", name2)
	}

	// 3. Custom base name without location tag should have location appended
	name3 := EnhanceGeofenceNameWithLocation("Thermal Perimeter Alpha", "Sumatran Tiger", 0.825, 101.530)
	if !strings.Contains(name3, "Riau Riparian Corridor, Sumatra") {
		t.Errorf("Expected Riau in name, got %q", name3)
	}

	// 4. Custom name already containing location in parentheses should not double-append
	name4 := EnhanceGeofenceNameWithLocation("Custom Zone (Kaeng Krachan Sector)", "Asian Elephant", 12.875, 99.650)
	if strings.Count(name4, "(") > 1 {
		t.Errorf("Did not expect duplicated parentheses, got %q", name4)
	}
}

func TestCreateGeofence_AutoFillsReverseGeocodedLocationName(t *testing.T) {
	store := NewStore()

	// Geofence with empty name
	g := store.CreateGeofence(AdminGeofence{
		Source:        "manual",
		GeometryType:  "circle",
		CircleCenter:  &Coordinate{Lat: 34.685, Lng: 135.832},
		ThreatLevel:   "TIER_2_ALERT",
		TargetSpecies: "Asian Black Bear",
	})

	if !strings.Contains(g.Name, "Kasugayama Primeval Forest, Nara") {
		t.Errorf("Expected reverse-geocoded place in g.Name, got %q", g.Name)
	}
}

func TestReverseGeocode_ExternalAPISuccessAndLRUCache(t *testing.T) {
	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		resp := googleGeocodeResponse{
			Status: "OK",
			Results: []struct {
				FormattedAddress  string `json:"formatted_address"`
				AddressComponents []struct {
					LongName  string   `json:"long_name"`
					ShortName string   `json:"short_name"`
					Types     []string `json:"types"`
				} `json:"address_components"`
			}{
				{
					FormattedAddress: "Champ de Mars, 5 Avenue Anatole France, 75007 Paris, France",
					AddressComponents: []struct {
						LongName  string   `json:"long_name"`
						ShortName string   `json:"short_name"`
						Types     []string `json:"types"`
					}{
						{LongName: "Avenue Anatole France", Types: []string{"route"}},
						{LongName: "Paris", Types: []string{"locality"}},
						{LongName: "Île-de-France", Types: []string{"administrative_area_level_1"}},
						{LongName: "France", Types: []string{"country"}},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	t.Setenv("GEOCODING_BASE_URL", ts.URL)
	t.Setenv("GEOCODING_API_KEY", "test-key-123")

	lat, lng := 48.8584, 2.2945
	result1 := ReverseGeocode(lat, lng)
	if result1 != "Avenue Anatole France, Paris" {
		t.Errorf("expected 'Avenue Anatole France, Paris', got %q", result1)
	}
	if requestCount != 1 {
		t.Errorf("expected 1 network request, got %d", requestCount)
	}

	// 2. Subsequent call should hit LRU cache and not trigger network request
	result2 := ReverseGeocode(lat, lng)
	if result2 != "Avenue Anatole France, Paris" {
		t.Errorf("expected cached 'Avenue Anatole France, Paris', got %q", result2)
	}
	if requestCount != 1 {
		t.Errorf("expected still 1 network request due to cache hit, got %d", requestCount)
	}
}

func TestReverseGeocode_APIFailureFallback(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	t.Setenv("GEOCODING_BASE_URL", ts.URL)
	t.Setenv("GEOCODING_API_KEY", "test-key-123")
	t.Setenv("ENV", "test")

	// Unknown coordinate in Pacific Ocean
	lat, lng := -20.1234, -150.5678
	result := ReverseGeocode(lat, lng)
	expectedFallback := "Sector (Lat -20.1234, Lng -150.5678)"
	if result != expectedFallback {
		t.Errorf("expected fallback %q, got %q", expectedFallback, result)
	}
}

func TestLRUGeocodeCache_Eviction(t *testing.T) {
	cache := NewLRUGeocodeCache(2)
	cache.Set("key1", "Place 1")
	cache.Set("key2", "Place 2")

	if v, ok := cache.Get("key1"); !ok || v != "Place 1" {
		t.Errorf("expected Place 1, got %v (ok: %t)", v, ok)
	}

	// Adding key3 should evict key2 (since key1 was recently accessed)
	cache.Set("key3", "Place 3")

	if _, ok := cache.Get("key2"); ok {
		t.Errorf("expected key2 to be evicted")
	}
	if v, ok := cache.Get("key1"); !ok || v != "Place 1" {
		t.Errorf("expected key1 to still exist")
	}
	if v, ok := cache.Get("key3"); !ok || v != "Place 3" {
		t.Errorf("expected key3 to exist")
	}
}
