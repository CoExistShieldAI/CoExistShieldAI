import { test, expect } from '@playwright/test';

const BACKEND_URL = process.env.BACKEND_URL || 'http://localhost:8080';

test.describe('Live Backend API Integration & Core User Workflows', () => {
  // 1. Service Health & Discovery
  test('Backend service is healthy and responds on /health and /api/v1/health', async ({ request }) => {
    const res = await request.get(`${BACKEND_URL}/health`);
    expect(res.status()).toBe(200);

    const body = await res.json();
    expect(body.status).toBe('healthy');
    expect(body.service).toContain('coexistshield-backend-cloudrun');
    expect(body.time).toBeDefined();

    const apiV1Res = await request.get(`${BACKEND_URL}/api/v1/health`);
    expect(apiV1Res.status()).toBe(200);
  });

  // 2. Active Geofences Retrieval
  test('Active Geofences API returns defensive polygon zones', async ({ request }) => {
    const res = await request.get(`${BACKEND_URL}/api/v1/geofences/active`);
    expect(res.status()).toBe(200);

    const geofences = await res.json();
    expect(Array.isArray(geofences)).toBe(true);
    if (geofences.length > 0) {
      const first = geofences[0];
      expect(first).toHaveProperty('id');
      expect(first).toHaveProperty('tier');
      expect(first).toHaveProperty('targetSpecies');
      expect(first).toHaveProperty('centroid');
      expect(first).toHaveProperty('polygonGeoJson');
    }
  });

  // 3. Dynamic Geofence Recalculation
  test('Geofence dynamic recalculation triggers and updates boundaries', async ({ request }) => {
    const recalcPayload = {
      region: 'JAPAN',
      originLat: 39.702,
      originLon: 141.152,
      targetSpecies: ['BEAR'],
      windSpeedMps: 4.5,
      windDirectionDeg: 270.0,
      relativeHumidity: 65.0,
      temperatureCelsius: 14.2,
      attractantFactors: [
        { type: 'OVERRIPE_PERSIMMON', weight: 0.85, coordinates: { latitude: 39.702, longitude: 141.152 } }
      ]
    };

    const res = await request.post(`${BACKEND_URL}/api/v1/geofences/recalculate`, {
      data: recalcPayload,
    });
    expect(res.status()).toBe(200);

    const data = await res.json();
    expect(data).toHaveProperty('recalculatedGeofencesCount');
    expect(data).toHaveProperty('affectedTiers');
    expect(typeof data.recalculatedGeofencesCount).toBe('number');
    expect(Array.isArray(data.affectedTiers)).toBe(true);
  });

  // 4. Incident Reporting & Sighting Pipeline
  test('Incident sighting report flow: submission and retrieval', async ({ request }) => {
    const newIncident = {
      id: `inc_test_${Date.now()}`,
      species: 'BEAR',
      location: {
        latitude: 39.7035,
        longitude: 141.1542,
      },
      severity: 'HIGH',
      description: 'Bear spotted near abandoned persimmon grove',
      reportedAt: new Date().toISOString(),
      status: 'CONFIRMED',
      source: 'CITIZEN_MOBILE'
    };

    // Submit incident
    const postRes = await request.post(`${BACKEND_URL}/api/v1/incidents`, {
      data: newIncident,
    });
    expect([200, 201]).toContain(postRes.status());

    // Verify incident appears in list
    const getRes = await request.get(`${BACKEND_URL}/api/v1/incidents`);
    expect(getRes.status()).toBe(200);

    const incidents = await getRes.json();
    expect(Array.isArray(incidents)).toBe(true);
    const found = incidents.find((i: any) => i.id === newIncident.id);
    expect(found).toBeDefined();
    expect(found.species).toBe('BEAR');
  });

  // 5. Roadkill Registry
  test('Roadkill records retrieval and telemetry', async ({ request }) => {
    const res = await request.get(`${BACKEND_URL}/api/v1/roadkills`);
    expect(res.status()).toBe(200);

    const roadkills = await res.json();
    expect(Array.isArray(roadkills)).toBe(true);
  });

  // 6. Cold-Start and Conservation NGO Ingestion Status
  test('Coldstart status endpoint returns baseline readiness', async ({ request }) => {
    const res = await request.get(`${BACKEND_URL}/api/v1/coldstart/status`);
    expect(res.status()).toBe(200);

    const status = await res.json();
    expect(status).toHaveProperty('bootstrapped');
    expect(status).toHaveProperty('totalGbifOccurrences');
    expect(status).toHaveProperty('regionsCovered');
    expect(Array.isArray(status.regionsCovered)).toBe(true);
  });

  // 7. BigQuery GIS Replication Streaming Status
  test('BigQuery GIS replicator reports live status and SLA metrics', async ({ request }) => {
    const res = await request.get(`${BACKEND_URL}/api/v1/bigquery/sync-status`);
    expect(res.status()).toBe(200);

    const bqStatus = await res.json();
    expect(bqStatus).toHaveProperty('status');
    expect(bqStatus).toHaveProperty('replicationLagSeconds');
    expect(bqStatus).toHaveProperty('spatialIndexing');
    expect(typeof bqStatus.replicationLagSeconds).toBe('number');
  });

  // 8. Corroboration Engine: Ingest & Evaluation
  test('Multi-source corroboration engine accepts sensor anomaly stream', async ({ request }) => {
    const sensorStream = {
      source: 'NASA_FIRMS',
      eventType: 'THERMAL_ANOMALY',
      latitude: -0.589,
      longitude: 101.343,
      confidence: 0.92,
      rawPayload: { brightness: 342.5, scan: 1.1, track: 1.0 },
      timestamp: new Date().toISOString()
    };

    const res = await request.post(`${BACKEND_URL}/api/v1/corroboration/ingest`, {
      data: sensorStream,
    });
    expect([200, 201]).toContain(res.status());

    const evalRes = await request.get(`${BACKEND_URL}/api/v1/corroboration/streams`);
    expect(evalRes.status()).toBe(200);
    const streams = await evalRes.json();
    expect(Array.isArray(streams)).toBe(true);
  });

  // 9. ML Incursion Prediction with H3 Spatio-Temporal Indexing
  test('ML pipeline predicts incursion risk based on real features and H3 spatial index', async ({ request }) => {
    const predPayload = {
      latitude: 39.702,
      longitude: 141.152,
      distNearestPersimmonM: 120.0,
      distNearestCarcassM: 850.0,
      distNearestCaneFruitM: 500.0,
      windScentAlignmentDeg: 15.0,
      moonDarknessPct: 82.0,
      precipitationMm24h: 5.4,
      historicalIncidents30d: 3
    };

    const res = await request.post(`${BACKEND_URL}/api/v1/ml/predict`, {
      data: predPayload,
    });
    expect(res.status()).toBe(200);

    const pred = await res.json();
    expect(pred).toHaveProperty('h3Index');
    expect(pred).toHaveProperty('incursionProbability');
    expect(pred).toHaveProperty('alertTier');
    expect(typeof pred.incursionProbability).toBe('number');
    expect(pred).toHaveProperty('dominantFeature');
    expect(pred).toHaveProperty('featureContributions');
  });

  // 10. Full End-to-End browser verification with live backend running
  test('Web Client interacts with live backend environment', async ({ page }) => {
    const consoleErrors: string[] = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text());
      }
    });

    await page.goto('/', { waitUntil: 'load' });
    await page.waitForTimeout(2000);

    const canvas = page.locator('#ComposeTarget');
    await expect(canvas).toBeVisible();

    // Verify canvas rendered successfully with dimensions
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    expect(box!.width).toBeGreaterThan(100);
    expect(box!.height).toBeGreaterThan(100);

    // Verify backend is concurrently responsive
    const healthCheck = await page.request.get(`${BACKEND_URL}/health`);
    expect(healthCheck.status()).toBe(200);
  });
});
