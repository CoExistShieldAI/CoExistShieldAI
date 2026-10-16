package admin

import (
	"fmt"
	"strings"
)

// ReverseGeocode resolves geographic coordinates (lat, lng) to a descriptive, human-readable place name.
func ReverseGeocode(lat, lng float64) string {
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

	return fmt.Sprintf("Sector (Lat %.4f, Lng %.4f)", lat, lng)
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
