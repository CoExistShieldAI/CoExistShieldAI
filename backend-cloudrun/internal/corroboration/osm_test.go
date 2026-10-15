package corroboration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchOsmGeospatialContextMock(t *testing.T) {
	mockResp := OverpassResponse{
		Elements: []struct {
			Type string            `json:"type"`
			Id   int64             `json:"id"`
			Lat  float64           `json:"lat"`
			Lon  float64           `json:"lon"`
			Tags map[string]string `json:"tags"`
		}{
			{
				Type: "way",
				Id:   1001,
				Tags: map[string]string{"highway": "motorway", "name": "Test Express Route"},
			},
			{
				Type: "way",
				Id:   1002,
				Tags: map[string]string{"waterway": "river", "name": "Test River"},
			},
			{
				Type: "way",
				Id:   1003,
				Tags: map[string]string{"landuse": "forest"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	t.Setenv("OSM_OVERPASS_API_ENDPOINT", server.URL)

	ctx := context.Background()
	osmCtx, err := FetchOsmGeospatialContext(ctx, server.Client(), 39.7020, 141.1520, 2000.0)
	if err != nil {
		t.Fatalf("unexpected error fetching OSM context: %v", err)
	}

	if !osmCtx.HasDividedMotorway {
		t.Errorf("expected HasDividedMotorway to be true")
	}
	if osmCtx.NearbyRiversCount != 1 {
		t.Errorf("expected 1 river, got %d", osmCtx.NearbyRiversCount)
	}
	if osmCtx.RiparianCorridorName != "Test River" {
		t.Errorf("expected 'Test River', got %s", osmCtx.RiparianCorridorName)
	}
	if osmCtx.ForestEdgeProximityM > 200.0 {
		t.Errorf("expected nearby forest proximity <= 200m, got %f", osmCtx.ForestEdgeProximityM)
	}
}
