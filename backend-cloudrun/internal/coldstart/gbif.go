package coldstart

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// GbifSearchResponse represents the JSON payload returned by GBIF Occurrence Search API.
type GbifSearchResponse struct {
	Offset       int `json:"offset"`
	Limit        int `json:"limit"`
	EndOfRecords bool `json:"endOfRecords"`
	Count        int `json:"count"`
	Results      []struct {
		Key              int64   `json:"key"`
		ScientificName   string  `json:"scientificName"`
		Species          string  `json:"species"`
		DecimalLatitude  float64 `json:"decimalLatitude"`
		DecimalLongitude float64 `json:"decimalLongitude"`
		EventDate        string  `json:"eventDate"`
		CountryCode      string  `json:"countryCode"`
		BasisOfRecord    string  `json:"basisOfRecord"`
	} `json:"results"`
}

// FetchGbifOccurrences fetches real historical wildlife occurrence records from the public GBIF API.
func FetchGbifOccurrences(ctx context.Context, client *http.Client, commonSpecies, scientificName, countryCode string, limit int) ([]GbifOccurrence, error) {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	if limit <= 0 {
		limit = 20
	}

	endpoint := "https://api.gbif.org/v1/occurrence/search"
	if custom := os.Getenv("GBIF_API_ENDPOINT"); custom != "" {
		endpoint = custom
	}

	q := url.Values{}
	q.Set("scientificName", scientificName)
	if countryCode != "" {
		q.Set("country", countryCode)
	}
	q.Set("hasCoordinate", "true")
	q.Set("limit", fmt.Sprintf("%d", limit))

	fullURL := fmt.Sprintf("%s?%s", endpoint, q.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build GBIF request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "CoExistShieldAI-ColdStart/1.0 (biodiversity-protection)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GBIF HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GBIF API returned HTTP %d", resp.StatusCode)
	}

	var searchResp GbifSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to parse GBIF JSON: %w", err)
	}

	occurrences := make([]GbifOccurrence, 0, len(searchResp.Results))
	for _, res := range searchResp.Results {
		if res.DecimalLatitude == 0 && res.DecimalLongitude == 0 {
			continue
		}

		eventDate := res.EventDate
		if len(eventDate) >= 10 {
			eventDate = eventDate[:10]
		}
		if eventDate == "" {
			eventDate = time.Now().UTC().Format("2006-01-02")
		}

		occurrences = append(occurrences, GbifOccurrence{
			OccurrenceID:   fmt.Sprintf("gbif_%d", res.Key),
			Species:        commonSpecies,
			ScientificName: res.ScientificName,
			CommonName:     commonSpecies,
			Latitude:       res.DecimalLatitude,
			Longitude:      res.DecimalLongitude,
			EventDate:      eventDate,
			CountryCode:    res.CountryCode,
			BasisOfRecord:  res.BasisOfRecord,
		})
	}

	return occurrences, nil
}
