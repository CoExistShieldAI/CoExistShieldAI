package coldstart_test

import (
	"strings"
	"testing"
	"time"

	"ai.coexistshield/backend/internal/coldstart"
	"ai.coexistshield/backend/internal/geofence"
	"ai.coexistshield/backend/internal/memory"
	"ai.coexistshield/backend/internal/ml"
)

func setupTestEngine(t *testing.T) (*coldstart.Engine, *ml.RetrainingPipeline, *geofence.Engine) {
	memBank := memory.NewBank()
	geoEngine := geofence.NewEngine()
	mlPipeline := ml.NewRetrainingPipeline()
	engine := coldstart.NewEngine(mlPipeline, geoEngine, memBank)
	return engine, mlPipeline, geoEngine
}

func TestBootstrapDayOne_AllRegions(t *testing.T) {
	engine, _, geoEngine := setupTestEngine(t)

	status := engine.GetStatus()
	if status.Bootstrapped {
		t.Fatalf("expected Bootstrapped=false initially, got true")
	}
	if status.TotalGbifOccurrences == 0 {
		t.Fatalf("expected pre-seeded GBIF occurrences, got 0")
	}
	if status.TotalFirmsAnomalies == 0 {
		t.Fatalf("expected pre-seeded FIRMS anomalies, got 0")
	}

	// Run Bootstrap for ALL regions
	bootStatus := engine.BootstrapDayOne("ALL")
	if !bootStatus.Bootstrapped {
		t.Fatalf("expected Bootstrapped=true after BootstrapDayOne")
	}
	if bootStatus.LastBootstrapTimestamp == "" {
		t.Fatalf("expected LastBootstrapTimestamp to be set")
	}

	// Geofences should be generated in the geofence engine
	activeGeofences := geoEngine.GetActiveGeofences(39.7, 141.1, 5000.0)
	if len(activeGeofences) == 0 {
		t.Fatalf("expected Day 1 baseline geofences to be generated, got 0")
	}
}

func TestBootstrapDayOne_RegionFilter(t *testing.T) {
	engine, _, _ := setupTestEngine(t)

	// Bootstrap specifically for JAPAN
	status := engine.BootstrapDayOne("JAPAN")
	if !status.Bootstrapped {
		t.Fatalf("expected Bootstrapped=true")
	}
}

func TestImportPatrolLogs_CSV(t *testing.T) {
	engine, _, _ := setupTestEngine(t)

	csvData := `event_date,animal,lat,long,category,individuals,risk_tier,description
2026-10-14T06:00:00Z,BEAR,39.7020,141.1520,DIRECT_SIGHTING,2,TIER_3_DANGER,Mother bear with cub near corn field
2026-10-14T07:30:00Z,クマ,39.7150,141.1410,CLAW_MARK,1,TIER_2_WARNING,Tree claw markings
2026-10-14T08:00:00Z,ELEPHANT,13.5210,101.4520,CROP_DAMAGE,5,TIER_3_DANGER,Herd crossing road towards village
`

	resp, err := engine.ImportPatrolLogs(coldstart.NgoBulkImportRequest{
		Format:        "CSV",
		Organization:  "Japan Bear Patrol NGO",
		RangerStation: "Tohoku-Station-Alpha",
		Region:        "JAPAN",
		Payload:       csvData,
	})
	if err != nil {
		t.Fatalf("unexpected error importing CSV: %v", err)
	}

	if resp.ImportedCount != 3 {
		t.Fatalf("expected 3 records imported, got %d", resp.ImportedCount)
	}
	if resp.SpeciesBreakdown["BEAR"] != 2 {
		t.Fatalf("expected 2 BEAR records, got %d", resp.SpeciesBreakdown["BEAR"])
	}
	if resp.SpeciesBreakdown["ELEPHANT"] != 1 {
		t.Fatalf("expected 1 ELEPHANT record, got %d", resp.SpeciesBreakdown["ELEPHANT"])
	}
	if resp.GeofencesActivated == 0 {
		t.Fatalf("expected geofences to be activated for TIER_2 and TIER_3 records")
	}
}

func TestImportPatrolLogs_GeoJSON(t *testing.T) {
	engine, _, _ := setupTestEngine(t)

	geoJsonData := `{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "geometry": { "type": "Point", "coordinates": [101.5300, 0.8250] },
      "properties": {
        "species": "Sumatran Tiger",
        "observationType": "CAMERA_TRAP",
        "tier": "TIER_3_DANGER",
        "notes": "Camera trap image of adult female tiger"
      }
    },
    {
      "type": "Feature",
      "geometry": {
        "type": "LineString",
        "coordinates": [
          [101.5320, 0.8260],
          [101.5340, 0.8270]
        ]
      },
      "properties": {
        "species": "Tiger",
        "observationType": "PUGMARK_TRACK",
        "tier": "TIER_2_WARNING",
        "notes": "Pugmark trail heading towards palm oil boundary"
      }
    }
  ]
}`

	resp, err := engine.ImportPatrolLogs(coldstart.NgoBulkImportRequest{
		Format:        "GEOJSON",
		Organization:  "WWF Indonesia / Riau Wildlife Unit",
		RangerStation: "Bukit Tigapuluh Outpost",
		Region:        "INDONESIA",
		Payload:       geoJsonData,
	})
	if err != nil {
		t.Fatalf("unexpected error importing GeoJSON: %v", err)
	}

	if resp.ImportedCount != 2 {
		t.Fatalf("expected 2 records imported, got %d", resp.ImportedCount)
	}
	if resp.SpeciesBreakdown["TIGER"] != 2 {
		t.Fatalf("expected 2 TIGER records, got %d", resp.SpeciesBreakdown["TIGER"])
	}
}

func TestImportPatrolLogs_GPX(t *testing.T) {
	engine, _, _ := setupTestEngine(t)

	gpxData := `<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" creator="SMART Conservation Handheld">
  <wpt lat="13.5210" lon="101.4520">
    <name>Elephant Crop Raid</name>
    <desc>Danger: 4 bulls broken electric fence</desc>
    <time>2026-10-14T03:00:00Z</time>
  </wpt>
  <trk>
    <name>Elephant Tracking Patrol Line</name>
    <trkseg>
      <trkpt lat="13.5220" lon="101.4530">
        <time>2026-10-14T03:15:00Z</time>
      </trkpt>
      <trkpt lat="13.5230" lon="101.4540">
        <time>2026-10-14T03:30:00Z</time>
      </trkpt>
    </trkseg>
  </trk>
</gpx>`

	resp, err := engine.ImportPatrolLogs(coldstart.NgoBulkImportRequest{
		Format:        "GPX",
		Organization:  "Khao Ang Rue Nai Ranger Unit",
		RangerStation: "Eastern Forest Complex",
		Region:        "THAILAND",
		Payload:       gpxData,
	})
	if err != nil {
		t.Fatalf("unexpected error importing GPX: %v", err)
	}

	// 1 waypoint + 2 track points = 3 records
	if resp.ImportedCount != 3 {
		t.Fatalf("expected 3 records imported, got %d", resp.ImportedCount)
	}
	if resp.SpeciesBreakdown["ELEPHANT"] != 3 {
		t.Fatalf("expected 3 ELEPHANT records, got %d", resp.SpeciesBreakdown["ELEPHANT"])
	}
}

func TestImportPatrolLogs_AutoDetection(t *testing.T) {
	engine, _, _ := setupTestEngine(t)

	templates := engine.GetStandardTemplates()

	// 1. Auto-detect CSV
	respCsv, err := engine.ImportPatrolLogs(coldstart.NgoBulkImportRequest{
		Format:        "AUTO",
		Organization:  "Test Org",
		RangerStation: "Station 1",
		Payload:       templates["CSV"],
	})
	if err != nil || respCsv.FormatDetected != "CSV" {
		t.Fatalf("expected CSV detected, got %s (err: %v)", respCsv.FormatDetected, err)
	}

	// 2. Auto-detect GeoJSON
	respGeo, err := engine.ImportPatrolLogs(coldstart.NgoBulkImportRequest{
		Format:        "AUTO",
		Organization:  "Test Org",
		RangerStation: "Station 1",
		Payload:       templates["GEOJSON"],
	})
	if err != nil || respGeo.FormatDetected != "GEOJSON" {
		t.Fatalf("expected GEOJSON detected, got %s (err: %v)", respGeo.FormatDetected, err)
	}

	// 3. Auto-detect GPX
	respGpx, err := engine.ImportPatrolLogs(coldstart.NgoBulkImportRequest{
		Format:        "AUTO",
		Organization:  "Test Org",
		RangerStation: "Station 1",
		Payload:       templates["GPX"],
	})
	if err != nil || respGpx.FormatDetected != "GPX" {
		t.Fatalf("expected GPX detected, got %s (err: %v)", respGpx.FormatDetected, err)
	}
}

func TestImportPatrolLogs_SLAConstraint(t *testing.T) {
	engine, _, _ := setupTestEngine(t)

	// Build a larger CSV payload with 100 rows
	var sb strings.Builder
	sb.WriteString("date,species,lat,lon,type,tier\n")
	for i := 0; i < 100; i++ {
		sb.WriteString("2026-10-14T10:00:00Z,BEAR,39.7020,141.1520,DIRECT_SIGHTING,TIER_2_WARNING\n")
	}

	start := time.Now()
	resp, err := engine.ImportPatrolLogs(coldstart.NgoBulkImportRequest{
		Format:        "CSV",
		Organization:  "Large Scale Ingestion",
		RangerStation: "HQ",
		Payload:       sb.String(),
	})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if resp.ImportedCount != 100 {
		t.Fatalf("expected 100 imported, got %d", resp.ImportedCount)
	}

	// SLA is 5 minutes (300,000 ms); our execution should comfortably complete in < 5 seconds
	if elapsed > 5*time.Second {
		t.Fatalf("ingestion took %v, exceeded 5s threshold (SLA is 5m)", elapsed)
	}
}
