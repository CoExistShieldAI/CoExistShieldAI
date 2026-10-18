import {
  EdgeVisionWeb,
  YoloRawCandidate,
  buildMunicipalDispatchGeoJson,
  computeIoU,
  computeSha256Hex,
  decodeYoloV8Tensor,
  estimateOpticalRange,
  evaluateKillSiteCachingHeuristic
} from './edge-vision.plugin';

describe('EdgeVisionPlugin (On-Device YOLOv8n / TFLite Vision & Forensic Heuristic)', () => {
  it('TC-01: should decode YOLOv8n tensor anchors and filter overlapping duplicates via NMS (<120ms)', async () => {
    const candidates: YoloRawCandidate[] = [
      {
        classId: 0,
        objectClass: 'BEAR',
        taxonomy: 'Ursus thibetanus',
        commonName: 'Asian Black Bear',
        confidence: 0.948,
        bbox: [0.22, 0.3, 0.55, 0.32]
      },
      {
        // High overlap (IoU > 0.45) duplicate anchor of same class -> should be suppressed by NMS
        classId: 0,
        objectClass: 'BEAR',
        taxonomy: 'Ursus thibetanus',
        commonName: 'Asian Black Bear',
        confidence: 0.85,
        bbox: [0.23, 0.31, 0.54, 0.31]
      },
      {
        // Low confidence (< 0.25) -> should be filtered out
        classId: 3,
        objectClass: 'WILD_BOAR',
        taxonomy: 'Sus scrofa',
        commonName: 'Wild Boar',
        confidence: 0.12,
        bbox: [0.1, 0.1, 0.2, 0.2]
      },
      {
        // Distinct class (PERSON_FACE for privacy masking) -> should be kept
        classId: 7,
        objectClass: 'PERSON_FACE',
        taxonomy: 'Privacy Region (Face)',
        commonName: 'Anonymized Bystander Face',
        confidence: 0.89,
        bbox: [0.05, 0.08, 0.12, 0.11]
      }
    ];

    const iou = computeIoU(candidates[0].bbox, candidates[1].bbox);
    expect(iou).toBeGreaterThan(0.45);

    const decoded = decodeYoloV8Tensor(candidates, 0.25, 0.45);
    expect(decoded.length).toBe(2);
    expect(decoded[0].taxonomy).toBe('Ursus thibetanus');
    expect(decoded[0].confidence).toBeCloseTo(0.948, 3);
    expect(decoded[1].objectClass).toBe('PERSON_FACE');

    const plugin = new EdgeVisionWeb();
    const res = await plugin.analyzeFrame({
      mode: 'KILL-SITE',
      speciesHint: 'BEAR',
      latitude: 39.7186,
      longitude: 140.1025,
      tiltDegrees: -6.4
    });

    expect(res.modelName).toBe('yolov8n_wildlife_int8.tflite');
    expect(res.inferenceLatencyMs).toBeLessThan(120);
    expect(res.primaryDetection.taxonomy).toBe('Ursus thibetanus');
  });

  it('TC-02: should estimate optical range and uncertainty from bounding box height, focal length, and tilt', () => {
    const estimate = estimateOpticalRange(0.3191, 'BEAR', 26, -6.4);
    expect(estimate.rangeMeters).toBeGreaterThanOrEqual(38);
    expect(estimate.rangeMeters).toBeLessThanOrEqual(46);
    expect(estimate.uncertaintyMeters).toBeGreaterThanOrEqual(3);
    expect(estimate.tiltDegrees).toBeCloseTo(-6.4, 1);
    expect(estimate.focalLengthMm).toBe(26);
  });

  it('TC-03: should trigger TIER_3_DANGER, 500m perimeter, ranger dispatch, and STEALTH_TACTICAL when carcass leaf_cover_score > 0.65', () => {
    const evalResult = evaluateKillSiteCachingHeuristic('CARCASS', 0.78);
    expect(evalResult.cachingDetected).toBeTrue();
    expect(evalResult.threatTier).toBe('TIER_3_DANGER');
    expect(evalResult.evacRadiusMeters).toBe(500);
    expect(evalResult.rangerDispatchRequired).toBeTrue();
    expect(evalResult.priority).toBe('HIGH');
    expect(evalResult.audioMode).toBe('STEALTH_TACTICAL');
  });

  it('TC-04: should not trigger caching alert when carcass leaf_cover_score <= 0.65 (boundary test)', () => {
    const boundaryResult = evaluateKillSiteCachingHeuristic('CARCASS', 0.65);
    expect(boundaryResult.cachingDetected).toBeFalse();
    expect(boundaryResult.threatTier).toBe('TIER_2_WARNING');
    expect(boundaryResult.evacRadiusMeters).toBe(250);
    expect(boundaryResult.audioMode).toBe('NORMAL');
  });

  it('TC-05: should apply privacy anonymization, generate 64-char SHA-256 digest, and format Municipal Dispatch GeoJSON', async () => {
    const hash = await computeSha256Hex('test-forensic-payload-2026');
    expect(hash.length).toBe(64);
    expect(/^[0-9a-f]{64}$/.test(hash)).toBeTrue();

    const geoJson = buildMunicipalDispatchGeoJson({
      incidentId: 'INC-892',
      tier: 'TIER_3_DANGER',
      taxonomy: 'Ursus thibetanus',
      confidence: 0.948,
      cachingDetected: true,
      leafCoverScore: 0.78,
      evacRadiusMeters: 500,
      rangerDispatchRequired: true,
      utcTimestamp: '2026-10-19T06:14:22Z',
      sha256: hash,
      latitude: 39.7186,
      longitude: 140.1025
    });

    expect(geoJson.type).toBe('Feature');
    expect(geoJson.geometry.type).toBe('Point');
    expect(geoJson.geometry.coordinates).toEqual([140.1025, 39.7186]);
    expect(geoJson.properties.incident_id).toBe('INC-892');
    expect(geoJson.properties.tier).toBe('DANGER');
    expect(geoJson.properties.caching_detected).toBeTrue();
    expect(geoJson.properties.evac_radius_m).toBe(500);
    expect(geoJson.properties.ranger_dispatch_required).toBeTrue();
  });
});
