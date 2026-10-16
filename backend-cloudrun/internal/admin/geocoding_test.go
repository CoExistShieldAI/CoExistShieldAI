package admin

import (
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
