package geofence

import (
	"testing"

	"ai.coexistshield/backend/internal/model"
)

func TestApplyOsmSpatialConstraints(t *testing.T) {
	centerLat := 39.7020
	centerLon := 141.1520

	// Create base unconstrained circular polygon
	basePoly := generatePlumePolygon(centerLat, centerLon, 1000.0, 1.0, 1.0, 0.0, false)

	// 1. Divided motorway barrier test
	highwayCtx := &model.OsmSpatialContext{
		HasDividedMotorway:  true,
		DividedHighwayNames: []string{"Tohoku Expressway"},
	}

	constrainedPoly := ApplyOsmSpatialConstraints(basePoly, highwayCtx, centerLat, centerLon)
	if constrainedPoly == nil || len(constrainedPoly.Coordinates[0]) == 0 {
		t.Fatalf("expected valid polygon geometry")
	}

	// Verify that northern point (dLat > 0) is compressed compared to original
	origNorthLat := basePoly.Coordinates[0][0][1]
	constrainedNorthLat := constrainedPoly.Coordinates[0][0][1]

	origDist := origNorthLat - centerLat
	constrainedDist := constrainedNorthLat - centerLat

	if constrainedDist >= origDist {
		t.Errorf("expected motorway barrier to compress northern boundary: origDist=%f, constrainedDist=%f",
			origDist, constrainedDist)
	}

	// 2. Riparian river corridor alignment test
	riverCtx := &model.OsmSpatialContext{
		NearbyRiversCount:     1,
		RiparianCorridorName:  "Nakatsu River",
		CorridorDeflectionDeg: 45.0, // Northeast flow
	}

	alignedPoly := ApplyOsmSpatialConstraints(basePoly, riverCtx, centerLat, centerLon)
	if alignedPoly == nil {
		t.Fatalf("expected valid aligned polygon")
	}
}
