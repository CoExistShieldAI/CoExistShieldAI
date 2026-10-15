package ml

import (
	"testing"
)

func TestApplyH3SpatialDiffusion(t *testing.T) {
	lat := 39.7020
	lon := 141.1520
	baseProb := 0.95

	// Apply 1-ring spatial diffusion
	diffused1 := ApplyH3SpatialDiffusion(lat, lon, baseProb, 1)

	// Center cell + up to 6 neighbor cells = up to 7 cells
	if len(diffused1) < 2 {
		t.Fatalf("expected multiple diffused cells, got %d", len(diffused1))
	}

	centerIndex := GeoToH3Res8(lat, lon)
	if prob, ok := diffused1[centerIndex]; !ok || prob != baseProb {
		t.Errorf("expected center cell to preserve baseProb %f, got %f", baseProb, prob)
	}

	// Neighbor cells should have decayed probability (~0.606 * baseProb ~= 0.57-0.60)
	for idx, prob := range diffused1 {
		if idx != centerIndex {
			if prob >= baseProb || prob <= 0.0 {
				t.Errorf("expected neighbor probability to decay properly, got %f for %s", prob, idx)
			}
		}
	}

	// Apply 2-ring spatial diffusion
	diffused2 := ApplyH3SpatialDiffusion(lat, lon, baseProb, 2)
	if len(diffused2) <= len(diffused1) {
		t.Errorf("expected 2-ring diffusion to cover more cells than 1-ring, got %d vs %d",
			len(diffused2), len(diffused1))
	}
}
