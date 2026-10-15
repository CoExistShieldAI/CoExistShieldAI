package geofence

import (
	"math"

	"ai.coexistshield/backend/internal/model"
)

// ApplyOsmSpatialConstraints adjusts raw scent plume polygon boundaries based on
// OpenStreetMap physical barriers (highways severing corridors) and natural movement pathways (rivers).
func ApplyOsmSpatialConstraints(
	poly *model.GeoJSONPolygonGeometry,
	osmCtx *model.OsmSpatialContext,
	centerLat, centerLon float64,
) *model.GeoJSONPolygonGeometry {
	if poly == nil || osmCtx == nil || len(poly.Coordinates) == 0 || len(poly.Coordinates[0]) == 0 {
		return poly
	}

	ring := poly.Coordinates[0]
	modifiedRing := make([][]float64, len(ring))

	// Corridor angle in radians
	corridorRad := osmCtx.CorridorDeflectionDeg * (math.Pi / 180.0)

	for i, coord := range ring {
		if len(coord) < 2 {
			modifiedRing[i] = coord
			continue
		}

		lon := coord[0]
		lat := coord[1]

		dLat := lat - centerLat
		dLon := lon - centerLon

		// Bearing from center to polygon vertex
		vertexAngle := math.Atan2(dLon, dLat) // radians from North

		// 1. Highway Barrier Clipping:
		// If divided motorway is present, compress polygon northward/crossing boundary by up to 35%
		barrierCompression := 1.0
		if osmCtx.HasDividedMotorway {
			// Assume divided motorway runs roughly east-west nearby, severing movement northward
			if dLat > 0 {
				barrierCompression = 0.65 // Suppress spillover across fenced highway
			}
		}

		// 2. Riparian Corridor Guiding:
		// If riverway present, stretch polygon along corridor alignment angle
		corridorStretch := 1.0
		if osmCtx.NearbyRiversCount > 0 && osmCtx.CorridorDeflectionDeg > 0 {
			angleDiff := math.Abs(vertexAngle - corridorRad)
			if angleDiff < math.Pi/4.0 || angleDiff > 7.0*math.Pi/4.0 {
				// Points aligned with river direction stretch by up to 25%
				corridorStretch = 1.25
			}
		}

		newLat := centerLat + dLat*barrierCompression*corridorStretch
		newLon := centerLon + dLon*barrierCompression*corridorStretch

		modifiedRing[i] = []float64{
			roundToDecimals(newLon, 6),
			roundToDecimals(newLat, 6),
		}
	}

	return &model.GeoJSONPolygonGeometry{
		Type:        "Polygon",
		Coordinates: [][][]float64{modifiedRing},
	}
}

// DefaultOsmContextForRegion returns calibrated regional baseline OSM context
// for offline testing or when Overpass is rate-limited.
func DefaultOsmContextForRegion(lat, lon float64) *model.OsmSpatialContext {
	// Tohoku Morioka area default: Route 4 bypass / Kitakami river corridor
	if lat > 35.0 && lat < 45.0 && lon > 135.0 && lon < 145.0 {
		return &model.OsmSpatialContext{
			NearbyHighwaysCount:   2,
			HasDividedMotorway:    true,
			DividedHighwayNames:   []string{"Tohoku Expressway", "Morioka Bypass"},
			NearbyRiversCount:     1,
			RiparianCorridorName:  "Nakatsu River Riparian Corridor",
			ForestEdgeProximityM:  220.0,
			CorridorDeflectionDeg: 45.0,
		}
	}
	// Thailand default: Highway 304 & Khao Ang Rue Nai corridor
	if lat > 10.0 && lat < 18.0 && lon > 98.0 && lon < 105.0 {
		return &model.OsmSpatialContext{
			NearbyHighwaysCount:   1,
			HasDividedMotorway:    true,
			DividedHighwayNames:   []string{"Highway 304 Wildlife Corridor"},
			NearbyRiversCount:     1,
			RiparianCorridorName:  "Khlong Rabom Tributary",
			ForestEdgeProximityM:  120.0,
			CorridorDeflectionDeg: 70.0,
		}
	}
	return &model.OsmSpatialContext{
		NearbyHighwaysCount:   0,
		HasDividedMotorway:    false,
		NearbyRiversCount:     0,
		ForestEdgeProximityM:  500.0,
		CorridorDeflectionDeg: 0.0,
	}
}
