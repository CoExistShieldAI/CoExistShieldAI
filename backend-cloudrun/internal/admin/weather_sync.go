package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"ai.coexistshield/backend/internal/geofence"
)

// OpenMeteoWeatherResponse holds current weather returned from open-meteo.com API.
type OpenMeteoWeatherResponse struct {
	Current struct {
		Temperature2m      float64 `json:"temperature_2m"`
		RelativeHumidity2m float64 `json:"relative_humidity_2m"`
		Precipitation      float64 `json:"precipitation"`
		WindSpeed10m       float64 `json:"wind_speed_10m"`
		WindDirection10m   float64 `json:"wind_direction_10m"`
		WeatherCode        int     `json:"weather_code"`
	} `json:"current"`
}

// MapWmoWeatherCode maps WMO weather codes to our operational weather conditions.
func MapWmoWeatherCode(code int, precip float64) string {
	if precip >= 25.0 {
		return "HEAVY_RAIN"
	}
	switch code {
	case 0, 1:
		return "CLEAR"
	case 2, 3:
		return "OVERCAST"
	case 45, 48:
		return "FOG"
	case 51, 53, 55, 61, 63:
		return "RAIN"
	case 65, 82:
		return "HEAVY_RAIN"
	case 71, 73, 75, 77, 85, 86:
		return "RAIN" // snow / sleet grouped into precipitation
	case 95, 96, 99:
		return "STORM"
	default:
		if precip > 0 {
			return "RAIN"
		}
		return "CLEAR"
	}
}

// FetchLiveOpenMeteo fetches real-time meteorological observations for a specific coordinate.
func FetchLiveOpenMeteo(ctx context.Context, client *http.Client, lat, lon float64) (*OpenMeteoWeatherResponse, error) {
	if client == nil {
		client = &http.Client{Timeout: 7 * time.Second}
	}

	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,precipitation,wind_speed_10m,wind_direction_10m,weather_code",
		lat, lon,
	)

	// Allow custom mock testing endpoint
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
		return nil, fmt.Errorf("Open-Meteo returned HTTP status %d", resp.StatusCode)
	}

	var data OpenMeteoWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode Open-Meteo response: %w", err)
	}

	return &data, nil
}

// SyncAllWeatherFromOpenMeteo connects to Open-Meteo REST API and refreshes all registered stations.
func (s *Store) SyncAllWeatherFromOpenMeteo(ctx context.Context) (int, int, error) {
	s.mu.RLock()
	stationList := make([]AdminWeatherData, 0, len(s.weather))
	for _, w := range s.weather {
		stationList = append(stationList, w)
	}
	s.mu.RUnlock()

	client := &http.Client{Timeout: 6 * time.Second}
	now := time.Now().UTC()
	moonInfo := geofence.CalculateMoonPhase(now)

	successCount := 0
	errorCount := 0

	for _, st := range stationList {
		resp, err := FetchLiveOpenMeteo(ctx, client, st.Latitude, st.Longitude)
		if err != nil {
			log.Printf("[WeatherCollector] Warning: failed to fetch Open-Meteo for %s (%.3f, %.3f): %v", st.StationName, st.Latitude, st.Longitude, err)
			errorCount++
			continue
		}

		s.mu.Lock()
		if current, exists := s.weather[st.ID]; exists {
			current.TemperatureCelsius = resp.Current.Temperature2m
			current.HumidityPct = resp.Current.RelativeHumidity2m
			current.WindSpeedKmh = resp.Current.WindSpeed10m
			current.WindDirectionDeg = resp.Current.WindDirection10m
			current.PrecipitationMmH = resp.Current.Precipitation
			current.WeatherCondition = MapWmoWeatherCode(resp.Current.WeatherCode, resp.Current.Precipitation)
			current.IsFloodedRiverRisk = resp.Current.Precipitation >= 25.0
			current.MoonIllumination = moonInfo.Illumination
			current.MoonPhase = moonInfo.PhaseName
			current.LastUpdated = now
			s.weather[st.ID] = current
			successCount++
		}
		s.mu.Unlock()
	}

	log.Printf("[WeatherCollector] Live Open-Meteo sync finished. Updated: %d, Failed: %d (Moon: %s, Illum: %.2f)",
		successCount, errorCount, moonInfo.PhaseName, moonInfo.Illumination)

	return successCount, errorCount, nil
}

// StartPeriodicWeatherSync starts a background worker that fetches fresh weather every interval.
func (s *Store) StartPeriodicWeatherSync(ctx context.Context, interval time.Duration) {
	if interval < 1*time.Minute {
		interval = 15 * time.Minute
	}

	go func() {
		log.Printf("[WeatherCollector] Periodic Internet Weather Collector started (Interval: %v)", interval)
		// Initial sync shortly after boot
		time.Sleep(5 * time.Second)
		_, _, _ = s.SyncAllWeatherFromOpenMeteo(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Printf("[WeatherCollector] Weather Collector stopped.")
				return
			case <-ticker.C:
				_, _, _ = s.SyncAllWeatherFromOpenMeteo(ctx)
			}
		}
	}()
}
