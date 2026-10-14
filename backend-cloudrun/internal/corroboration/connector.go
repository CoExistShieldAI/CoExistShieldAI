package corroboration

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"ai.coexistshield/backend/internal/model"
)

// OpenMeteoCurrentResponse represents current weather output from open-meteo.com API.
type OpenMeteoCurrentResponse struct {
	Current struct {
		Time               string  `json:"time"`
		Temperature2m      float64 `json:"temperature_2m"`
		RelativeHumidity2m float64 `json:"relative_humidity_2m"`
		Precipitation      float64 `json:"precipitation"`
		WindSpeed10m       float64 `json:"wind_speed_10m"`
		WindDirection10m   float64 `json:"wind_direction_10m"`
	} `json:"current"`
}

// FetchOpenMeteoWeather fetches live meteorological observations from the free Open-Meteo REST API.
func FetchOpenMeteoWeather(ctx context.Context, client *http.Client, lat, lon float64) (*model.WeatherConditions, error) {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,precipitation,wind_speed_10m,wind_direction_10m",
		lat, lon)

	// Allow custom endpoint for mock testing
	if customEndpoint := os.Getenv("OPEN_METEO_API_ENDPOINT"); customEndpoint != "" {
		url = fmt.Sprintf("%s?latitude=%.4f&longitude=%.4f", customEndpoint, lat, lon)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Open-Meteo request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Open-Meteo HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Open-Meteo returned HTTP %d", resp.StatusCode)
	}

	var data OpenMeteoCurrentResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse Open-Meteo response: %w", err)
	}

	return &model.WeatherConditions{
		WindSpeedKmh:     data.Current.WindSpeed10m,
		WindDirectionDeg: data.Current.WindDirection10m,
		PrecipitationMmH: data.Current.Precipitation,
		MoonIllumination: 0.5, // Will be enriched with astronomical calculations
	}, nil
}

// FetchFirmsThermalAnomalies fetches live satellite thermal fire anomalies from NASA FIRMS REST API.
// Requires FIRMS_MAP_KEY environment variable.
func FetchFirmsThermalAnomalies(ctx context.Context, client *http.Client, countryCode string, mapKey string) ([]model.CausalStreamItem, error) {
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}

	if mapKey == "" {
		mapKey = os.Getenv("FIRMS_MAP_KEY")
	}
	if mapKey == "" {
		// Key not set; return empty list without failing
		return nil, nil
	}

	url := fmt.Sprintf("https://firms.modaps.eosdis.nasa.gov/api/country/csv/%s/VIIRS_SNPP_NRT/%s/1", mapKey, countryCode)
	if customEndpoint := os.Getenv("FIRMS_API_ENDPOINT"); customEndpoint != "" {
		url = fmt.Sprintf("%s?country=%s", customEndpoint, countryCode)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build FIRMS request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("FIRMS HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FIRMS API returned HTTP %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read FIRMS response body: %w", err)
	}

	reader := csv.NewReader(strings.NewReader(string(bodyBytes)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse FIRMS CSV: %w", err)
	}

	if len(records) <= 1 {
		return nil, nil // Header only or empty
	}

	now := time.Now().UTC()
	exp24h := now.Add(24 * time.Hour).Format(time.RFC3339)
	var items []model.CausalStreamItem

	// Parse CSV rows: country_id,latitude,longitude,bright_ti4,scan,track,acq_date,acq_time,satellite,instrument,confidence,version,bright_ti5,frp,daynight
	for idx, row := range records[1:] {
		if len(row) < 4 {
			continue
		}
		lat, err1 := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		lon, err2 := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
		if err1 != nil || err2 != nil {
			continue
		}

		brightK, _ := strconv.ParseFloat(strings.TrimSpace(row[3]), 64)

		item := model.CausalStreamItem{
			ItemId:        fmt.Sprintf("firms-%s-%d-%d", countryCode, now.Unix(), idx),
			StreamType:    model.StreamSatelliteThermal,
			Region:        countryCodeToRegion(countryCode),
			Title:         fmt.Sprintf("NASA FIRMS Active Thermal Anomaly (%.1f K)", brightK),
			Description:   "Satellite-detected thermal fire/burn hotspot compressing wildlife movement corridor.",
			Latitude:      lat,
			Longitude:     lon,
			RadiusMeters:  3000,
			Confidence:    0.85,
			TargetSpecies: regionToSpecies(countryCodeToRegion(countryCode)),
			Timestamp:     now.Format(time.RFC3339),
			ExpiresAt:     exp24h,
			Metadata: map[string]interface{}{
				"sensor":      "VIIRS_SNPP",
				"brightnessK": brightK,
				"country":     countryCode,
			},
		}
		items = append(items, item)
	}

	return items, nil
}

func countryCodeToRegion(code string) string {
	switch strings.ToUpper(code) {
	case "JPN", "JP":
		return "JAPAN"
	case "THA", "TH":
		return "THAILAND"
	case "IDN", "ID":
		return "INDONESIA"
	default:
		return "JAPAN"
	}
}

func regionToSpecies(region string) string {
	switch region {
	case "JAPAN":
		return "BEAR"
	case "THAILAND":
		return "ELEPHANT"
	case "INDONESIA":
		return "TIGER"
	default:
		return "BEAR"
	}
}
