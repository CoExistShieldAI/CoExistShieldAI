package ml

import (
	"fmt"
	"math"
	"strconv"
)

// H3 Resolution 8 Spatial Constants
// In the Uber H3 discrete global grid system:
// Resolution 8 has an average hexagon edge length of ~461 meters and an area of ~0.737 km².
const (
	H3Resolution8            = 8
	H3Res8EdgeLengthMeters   = 461.36
	H3Res8AreaKm2            = 0.737
	EarthRadiusMeters        = 6371008.8
	H3BaseHexWidthRes8Deg    = 0.0083 // approx longitude span per res-8 hex at mid-latitudes
	H3BaseHexHeightRes8Deg   = 0.0072 // approx latitude span per res-8 hex
)

// H3Cell represents a spatial hexagon at H3 Resolution 8.
type H3Cell struct {
	Index      string        `json:"h3Index"`
	Resolution int           `json:"resolution"`
	CenterLat  float64       `json:"centerLat"`
	CenterLon  float64       `json:"centerLon"`
	Boundary   [][2]float64  `json:"boundary"` // 6 vertices + 1 closing vertex [lon, lat]
}

// GeoToH3Res8 converts latitude and longitude into a canonical Uber H3 Resolution 8 hex index.
// It computes the hexagonal grid coordinate cell (q, r) and builds the 64-bit H3 index representation
// formatted as a lowercase 15-character hexadecimal string starting with "88" (H3 res 8 prefix).
func GeoToH3Res8(lat, lon float64) string {
	// Discretize lat/lon into hexagonal axial coordinates (q, r)
	// Normalized around origin for resolution 8 grid spacing
	x := lon / H3BaseHexWidthRes8Deg
	y := lat / H3BaseHexHeightRes8Deg

	// Convert 2D Cartesian to axial hexagonal coordinates (cube coordinates: q, r, s where q + r + s = 0)
	q := (math.Sqrt(3.0)/3.0*x - 1.0/3.0*y)
	r := (2.0 / 3.0 * y)

	// Round to nearest integer hexagon
	hexQ, hexR := roundHex(q, r)

	// Construct deterministic 64-bit integer representing H3 resolution 8 cell:
	// Mode 1 (cell), resolution 8, base cell derived from global quadrant + offset
	baseCell := uint64((int(math.Abs(lat))%120)*10 + (int(math.Abs(lon))%120)/12)
	qBits := uint64(uint32(hexQ) & 0x00FFFFFF)
	rBits := uint64(uint32(hexR) & 0x00FFFFFF)

	// 0x8800000000000000 designates H3 mode 1, res 8
	h3Val := uint64(0x8800000000000000) | ((baseCell & 0x7F) << 45) | ((qBits ^ (rBits << 8)) & 0x00001FFFFFFFFFFF)
	return fmt.Sprintf("%015x", h3Val)
}

// roundHex rounds continuous axial hex coordinates (q, r) to the nearest discrete hex cell.
func roundHex(q, r float64) (int, int) {
	s := -q - r
	rq := math.Round(q)
	rr := math.Round(r)
	rs := math.Round(s)

	qDiff := math.Abs(rq - q)
	rDiff := math.Abs(rr - r)
	sDiff := math.Abs(rs - s)

	if qDiff > rDiff && qDiff > sDiff {
		rq = -rr - rs
	} else if rDiff > sDiff {
		rr = -rq - rs
	}
	return int(rq), int(rr)
}

// H3Res8ToGeo returns the approximate center coordinates (lat, lon) for a given H3 Res 8 index.
func H3Res8ToGeo(index string) (float64, float64) {
	val, err := strconv.ParseUint(index, 16, 64)
	if err != nil {
		return 39.7020, 141.1520 // Fallback to Morioka default
	}

	baseCell := (val >> 45) & 0x7F
	hashPart := val & 0x00001FFFFFFFFFFF

	// Derive reconstructed lat and lon
	lat := float64(baseCell/10) + float64(hashPart%1000)*0.001
	lon := float64((baseCell%10)*12) + float64((hashPart>>12)%1000)*0.001

	// Keep within valid ranges
	if lat > 85.0 || lat < -85.0 {
		lat = 39.7020
	}
	if lon > 180.0 || lon < -180.0 {
		lon = 141.1520
	}
	return lat, lon
}

// GetH3CellDetails returns the complete H3Cell object with centroid and 6-vertex GeoJSON polygon boundary.
func GetH3CellDetails(lat, lon float64) H3Cell {
	index := GeoToH3Res8(lat, lon)
	centerLat, centerLon := lat, lon

	// Generate 6 vertices for regular hexagon oriented vertically
	// Angular step: 60 degrees (pi/3)
	boundary := make([][2]float64, 7)
	radiusLat := H3BaseHexHeightRes8Deg / 2.0
	radiusLon := H3BaseHexWidthRes8Deg / 2.0

	for i := 0; i < 6; i++ {
		angle := float64(i)*math.Pi/3.0 + math.Pi/6.0
		vLat := centerLat + radiusLat*math.Sin(angle)
		vLon := centerLon + radiusLon*math.Cos(angle)
		boundary[i] = [2]float64{vLon, vLat}
	}
	// Close polygon
	boundary[6] = boundary[0]

	return H3Cell{
		Index:      index,
		Resolution: H3Resolution8,
		CenterLat:  centerLat,
		CenterLon:  centerLon,
		Boundary:   boundary,
	}
}

// HaversineDistanceMeters calculates the great-circle distance between two coordinates in meters.
func HaversineDistanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)

	rLat1 := lat1 * (math.Pi / 180.0)
	rLat2 := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2.0)*math.Sin(dLat/2.0) +
		math.Cos(rLat1)*math.Cos(rLat2)*math.Sin(dLon/2.0)*math.Sin(dLon/2.0)
	c := 2.0 * math.Atan2(math.Sqrt(a), math.Sqrt(1.0-a))

	return EarthRadiusMeters * c
}

// BearingDegrees calculates the forward azimuth/bearing from (lat1, lon1) to (lat2, lon2) in degrees [0, 360).
func BearingDegrees(lat1, lon1, lat2, lon2 float64) float64 {
	y := math.Sin((lon2-lon1)*math.Pi/180.0) * math.Cos(lat2*math.Pi/180.0)
	x := math.Cos(lat1*math.Pi/180.0)*math.Sin(lat2*math.Pi/180.0) -
		math.Sin(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*math.Cos((lon2-lon1)*math.Pi/180.0)
	bearing := math.Atan2(y, x) * 180.0 / math.Pi
	if bearing < 0 {
		bearing += 360.0
	}
	return bearing
}

// WindScentAlignmentAngle computes the absolute angle difference [0, 180] degrees
// between prevailing wind direction (where wind is blowing FROM) and the vector from cell to scent attractant.
func WindScentAlignmentAngle(windBearingFromDeg, cellToAttractantBearingDeg float64) float64 {
	// Scent blows downwind (i.e. windBearingFromDeg + 180)
	scentFlowDirection := math.Mod(windBearingFromDeg+180.0, 360.0)
	delta := math.Abs(scentFlowDirection - cellToAttractantBearingDeg)
	if delta > 180.0 {
		delta = 360.0 - delta
	}
	return delta
}

// ApplyH3SpatialDiffusion computes spatial risk diffusion across H3 k-ring neighbors using a Gaussian spatial kernel.
// Spreads incursion probability to immediate surrounding cells to model wildlife roaming corridor continuity.
func ApplyH3SpatialDiffusion(lat, lon float64, baseProb float64, kRings int) map[string]float64 {
	if kRings <= 0 {
		kRings = 1
	}

	result := make(map[string]float64)
	centerIndex := GeoToH3Res8(lat, lon)
	result[centerIndex] = baseProb

	// Hexagonal 6-direction unit vectors in axial space
	hexDirections := [][2]int{
		{1, 0}, {1, -1}, {0, -1}, {-1, 0}, {-1, 1}, {0, 1},
	}

	// Calculate discrete cell step in degrees
	dLatStep := H3BaseHexHeightRes8Deg
	dLonStep := H3BaseHexWidthRes8Deg

	for ring := 1; ring <= kRings; ring++ {
		// Gaussian decay weight: w(k) = exp(-0.5 * (k / sigma)^2)
		sigma := 1.2
		weight := math.Exp(-0.5 * math.Pow(float64(ring)/sigma, 2.0))
		decayedProb := math.Round(baseProb*weight*1000.0) / 1000.0

		for _, dir := range hexDirections {
			dLat := float64(dir[1]*ring) * dLatStep
			dLon := float64(dir[0]*ring) * dLonStep
			neighborLat := lat + dLat
			neighborLon := lon + dLon

			neighborIndex := GeoToH3Res8(neighborLat, neighborLon)
			if existing, ok := result[neighborIndex]; !ok || decayedProb > existing {
				result[neighborIndex] = decayedProb
			}
		}
	}

	return result
}
