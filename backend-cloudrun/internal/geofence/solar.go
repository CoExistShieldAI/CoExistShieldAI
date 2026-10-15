package geofence

import (
	"math"
	"time"

	"ai.coexistshield/backend/internal/model"
)

// CalculateSolarPhase computes the solar elevation angle and crepuscular risk multiplier
// using standard NOAA astronomical equations for any coordinate and timestamp.
func CalculateSolarPhase(lat, lon float64, t time.Time) model.SolarPhaseInfo {
	tUTC := t.UTC()

	// Day of year
	dayOfYear := float64(tUTC.YearDay())

	// Fractional year angle B in radians
	bRad := (2.0 * math.Pi / 365.0) * (dayOfYear - 81.0)

	// Equation of time E in minutes
	eMin := 9.87*math.Sin(2.0*bRad) - 7.53*math.Cos(bRad) - 1.5*math.Sin(bRad)

	// Solar declination delta in radians
	// Approximated: 23.44° * sin(B)
	declinationRad := (23.44 * math.Pi / 180.0) * math.Sin(bRad)

	// Time in hours (UTC fractional hours)
	utcHours := float64(tUTC.Hour()) + float64(tUTC.Minute())/60.0 + float64(tUTC.Second())/3600.0

	// Local Solar Time (LST) in hours
	lstHours := utcHours + (lon / 15.0) + (eMin / 60.0)
	// Normalize to [0, 24)
	lstHours = math.Mod(lstHours, 24.0)
	if lstHours < 0 {
		lstHours += 24.0
	}

	// Solar Hour Angle H in degrees (-180° to +180°)
	hourAngleDeg := (lstHours - 12.0) * 15.0
	hourAngleRad := hourAngleDeg * (math.Pi / 180.0)

	latRad := lat * (math.Pi / 180.0)

	// Solar elevation angle: sin(alpha) = sin(lat)*sin(delta) + cos(lat)*cos(delta)*cos(H)
	sinAlpha := math.Sin(latRad)*math.Sin(declinationRad) + math.Cos(latRad)*math.Cos(declinationRad)*math.Cos(hourAngleRad)
	sinAlpha = math.Max(-1.0, math.Min(1.0, sinAlpha))
	elevationRad := math.Asin(sinAlpha)
	elevationDeg := elevationRad * (180.0 / math.Pi)
	elevationDeg = roundToDecimals(elevationDeg, 2)

	// Classify phase name
	var phaseName string
	var crepuscularMultiplier float64
	var isCrepuscularPeak bool

	switch {
	case elevationDeg < -18.0:
		phaseName = "NIGHT"
		crepuscularMultiplier = 1.05 // Baseline nighttime movement
		isCrepuscularPeak = false
	case elevationDeg < -12.0:
		phaseName = "ASTRONOMICAL_TWILIGHT"
		crepuscularMultiplier = 1.20
		isCrepuscularPeak = false
	case elevationDeg < -6.0:
		phaseName = "NAUTICAL_TWILIGHT"
		crepuscularMultiplier = 1.45
		isCrepuscularPeak = true
	case elevationDeg < 0.0:
		phaseName = "CIVIL_TWILIGHT"
		// Peak stealth foraging/ambush right before dawn or after dusk
		crepuscularMultiplier = 1.70
		isCrepuscularPeak = true
	case elevationDeg < 6.0:
		phaseName = "GOLDEN_HOUR"
		// Immediate post-sunrise / pre-sunset crepuscular window
		crepuscularMultiplier = 1.55
		isCrepuscularPeak = true
	case elevationDeg < 25.0:
		phaseName = "DAYLIGHT"
		crepuscularMultiplier = 0.95
		isCrepuscularPeak = false
	default:
		phaseName = "DAYLIGHT"
		// High solar noon: animals bed down or shelter in deep canopy
		crepuscularMultiplier = 0.75
		isCrepuscularPeak = false
	}

	return model.SolarPhaseInfo{
		PhaseName:                 phaseName,
		ElevationDegrees:          elevationDeg,
		CrepuscularRiskMultiplier: roundToDecimals(crepuscularMultiplier, 3),
		IsCrepuscularPeakRisk:     isCrepuscularPeak,
	}
}
