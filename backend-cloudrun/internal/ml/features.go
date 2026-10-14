package ml

import (
	"math"
	"math/rand"
	"time"
)

// DefaultKnownAttractants returns baseline ground-truth attractant locations across Japan, Thailand, and Indonesia.
func DefaultKnownAttractants() (kaki []AttractantFeaturePoint, carcasses []AttractantFeaturePoint, crops []AttractantFeaturePoint) {
	kaki = []AttractantFeaturePoint{
		{Type: "KAKI", Latitude: 39.7020, Longitude: 141.1520, Weight: 0.95},
		{Type: "KAKI", Latitude: 39.7045, Longitude: 141.1490, Weight: 0.91},
		{Type: "KAKI", Latitude: 39.7120, Longitude: 141.1550, Weight: 0.88},
	}
	carcasses = []AttractantFeaturePoint{
		{Type: "CARCASS", Latitude: 39.7120, Longitude: 141.1480, Weight: 0.90},
		{Type: "CARCASS", Latitude: 39.6410, Longitude: 141.9520, Weight: 0.80},
	}
	crops = []AttractantFeaturePoint{
		{Type: "CROP", Latitude: 13.5210, Longitude: 101.4520, Weight: 0.85},
		{Type: "CROP", Latitude: 0.8250, Longitude: 101.5300, Weight: 0.90},
	}
	return kaki, carcasses, crops
}

// ExtractFeatureVector calculates spatial distances, wind scent alignment, moon phase, and environmental features
// for any geographic coordinate and observation timestamp.
func ExtractFeatureVector(lat, lon float64, timestamp time.Time,
	kakiList, carcassList, cropList []AttractantFeaturePoint) FeatureVector {

	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}

	// 1. Uber H3 Resolution 8 Spatial Indexing
	h3Index := GeoToH3Res8(lat, lon)

	// 2. Spatial Joins: Min Distance to Unmanaged Persimmons
	minKakiDist := 9999.0
	for _, k := range kakiList {
		d := HaversineDistanceMeters(lat, lon, k.Latitude, k.Longitude)
		if d < minKakiDist {
			minKakiDist = d
		}
	}

	// 3. Spatial Joins: Min Distance to Decomposing Carcasses
	minCarcassDist := 9999.0
	nearestCarcassLat, nearestCarcassLon := lat, lon
	for _, c := range carcassList {
		d := HaversineDistanceMeters(lat, lon, c.Latitude, c.Longitude)
		if d < minCarcassDist {
			minCarcassDist = d
			nearestCarcassLat = c.Latitude
			nearestCarcassLon = c.Longitude
		}
	}

	// 4. Spatial Joins: Min Distance to High-Sugar / Cane Plantations
	minCropDist := 9999.0
	for _, cr := range cropList {
		d := HaversineDistanceMeters(lat, lon, cr.Latitude, cr.Longitude)
		if d < minCropDist {
			minCropDist = d
		}
	}

	// 5. Environmental & Weather Feature Enrichment
	bearingToAttractant := BearingDegrees(lat, lon, nearestCarcassLat, nearestCarcassLon)
	prevailingWindFromDeg := 285.0 // Westerly wind
	windScentAngle := WindScentAlignmentAngle(prevailingWindFromDeg, bearingToAttractant)

	// Moon darkness based on day of month (approximate lunar cycle)
	dayOfMonth := float64(timestamp.Day())
	moonDarkness := math.Max(10.0, math.Min(95.0, 50.0+45.0*math.Sin(dayOfMonth*0.21)))

	return FeatureVector{
		H3Index:                h3Index,
		DistNearestPersimmonM:  math.Round(minKakiDist*10.0) / 10.0,
		DistNearestCarcassM:    math.Round(minCarcassDist*10.0) / 10.0,
		DistNearestCaneFruitM:  math.Round(minCropDist*10.0) / 10.0,
		WindScentAlignmentDeg:  math.Round(windScentAngle*10.0) / 10.0,
		MoonDarknessPct:        math.Round(moonDarkness*10.0) / 10.0,
		PrecipitationMm24h:      2.5,
		HistoricalIncidents30d: 3.0,
	}
}

// GenerateNegativeSamples creates balanced negative training samples (IsIncursion = 0)
// drawn from safe non-incursion buffer zones well away from human settlements and attractants.
func GenerateNegativeSamples(count int, centerLat, centerLon float64,
	kakiList, carcassList, cropList []AttractantFeaturePoint) []TrainingSample {

	samples := make([]TrainingSample, 0, count)
	r := rand.New(rand.NewSource(42))

	for i := 0; i < count; i++ {
		// Sample coordinates 3km - 10km away in deep forest or distant safe buffer
		angle := r.Float64() * 2 * math.Pi
		distKm := 3.0 + r.Float64()*7.0
		// Approx degrees conversion
		dLat := (distKm / 111.0) * math.Cos(angle)
		dLon := (distKm / (111.0 * math.Cos(centerLat*math.Pi/180.0))) * math.Sin(angle)

		sampleLat := centerLat + dLat
		sampleLon := centerLon + dLon

		features := ExtractFeatureVector(sampleLat, sampleLon, time.Now().UTC().AddDate(0, 0, -i),
			kakiList, carcassList, cropList)

		// Negative control features: low incident history, safe weather / rain damping
		features.HistoricalIncidents30d = 0.0
		features.PrecipitationMm24h = 15.0 + r.Float64()*10.0 // rainy, animals shelter
		features.MoonDarknessPct = 20.0 + r.Float64()*20.0     // bright moon

		samples = append(samples, TrainingSample{
			Features:    features,
			IsIncursion: 0,
		})
	}

	return samples
}
