export type AlertTier = 'SAFE' | 'TIER_1_NOTICE' | 'TIER_2_WARNING' | 'TIER_3_DANGER';

export type UserRole = 'CITIZEN' | 'OFFICIAL' | 'ADMIN';

export type WildlifeSpecies = 'BEAR' | 'TIGER' | 'ELEPHANT' | 'WILD_BOAR' | 'DEER' | 'UNKNOWN';

export type ReportCategory =
  | 'WILDLIFE_SIGHTING'
  | 'PHYSICAL_TRACE'
  | 'ROADKILL'
  | 'ATTRACTANT_FRUIT_TREE'
  | 'LIVESTOCK_DEPREDATION';

export type DecayState = 'FRESH' | 'BLOATED' | 'DECOMPOSING' | 'SKELETAL';

export type ClearanceStatus = 'REPORTED' | 'DISPATCHED' | 'CLEARED';

export type PropertyContextTag =
  | 'VACANT_HOUSE'
  | 'ROADSIDE_UNMANAGED'
  | 'ACTIVE_FARM_BORDER'
  | 'PUBLIC_LAND'
  | 'UNKNOWN';

export type MapRegionId = 'JAPAN_IWATE' | 'THAILAND_KUI_BURI' | 'INDONESIA_SUMATRA';

export type MapLayerType =
  | 'GEOFENCES'
  | 'ROADKILL'
  | 'PERSIMMON_TREES'
  | 'ELEPHANT_CROPS'
  | 'LIVESTOCK_PREDATION';

export type ActiveLeftTab = 'HUD' | 'REPORT' | 'CONSOLE';

export type HudIntelSection =
  | 'AI_CAUSAL'
  | 'DETOUR'
  | 'SAFETY_PROTOCOL'
  | 'MUNICIPAL_TASKS'
  | 'DETERRENT_SIREN';

export type AdminTab = 'JAPAN_MUNICIPAL' | 'APAC_PLANTATION' | 'SYSTEM_AI' | 'COLD_START_NGO';

export interface GeoLocation {
  latitude: number;
  longitude: number;
}

export interface MapRegionConfig {
  id: MapRegionId;
  label: string;
  shortLabel: string;
  centerLat: number;
  centerLon: number;
  defaultZoom: number;
}

export interface GeofenceZone {
  id: string;
  name: string;
  tier: AlertTier;
  targetSpecies: WildlifeSpecies[];
  coordinates: GeoLocation[];
  centroid: GeoLocation;
  radiusMeters: number;
  causalExplanation: string;
  source: 'ai' | 'manual';
  validFrom: string;
  validUntil: string;
  generatedAt: string;
}

export interface RoadkillRecord {
  id: string;
  incidentId: string;
  species: WildlifeSpecies;
  location: GeoLocation;
  decayState: DecayState;
  attractantThreatScore: number;
  clearanceStatus: ClearanceStatus;
  assignedOfficialId?: string;
  reportedAt: string;
  clearedAt?: string;
}

export interface AttractantMarkerItem {
  id: string;
  lat: number;
  lon: number;
  iconEmoji: string;
  title: string;
  subtitle: string;
  layerType: MapLayerType;
  threatTier: AlertTier;
  geminiBadge: string;
  region: MapRegionId;
  radiusMeters?: number;
  isResolved?: boolean;
}

export interface SelectedEntityDetails {
  id: string;
  title: string;
  subtitle: string;
  speciesOrType: string;
  threatTier: AlertTier;
  geminiBadgeText?: string;
  canMarkResolved: boolean;
  isRoadkill: boolean;
  roadkillStatus?: ClearanceStatus;
  isGeofence: boolean;
  radiusMeters?: number;
  latitude: number;
  longitude: number;
}

export interface GeminiTriageResult {
  detectedLabel: string;
  commonName: string;
  confidence: number;
  forensicTags: string[];
  fruitLoadLevel?: string;
  traceType?: string;
  traumaType?: string;
  notes: string;
}

export interface IncidentReport {
  id: string;
  category: ReportCategory;
  species: WildlifeSpecies;
  location: GeoLocation;
  timestamp: string;
  region: string;
  description: string;
  reporterId: string;
  reporterRole: UserRole;
  photoBase64?: string;
  sha256?: string;
  privacyMaskApplied?: boolean;
  cachingDetected?: boolean;
  leafCoverScore?: number;
  evacRadiusMeters?: number;
  opticalRangeMeters?: number;
  taxonomy?: string;
  geoJsonFeature?: MunicipalDispatchGeoJsonFeature;
}

export type EdgeCameraMode = 'SIGHTING' | 'KILL-SITE' | 'ATTRACTANT';

export type EdgeObjectClass =
  | 'BEAR'
  | 'TIGER'
  | 'ELEPHANT'
  | 'WILD_BOAR'
  | 'DEER'
  | 'CARCASS'
  | 'ATTRACTANT'
  | 'PERSON_FACE'
  | 'VEHICLE_PLATE';

export interface EdgeDetectionBox {
  classId: number;
  objectClass: EdgeObjectClass;
  taxonomy: string;
  commonName: string;
  confidence: number;
  /** Normalized [x, y, width, height] in 0.0 .. 1.0 */
  bbox: [number, number, number, number];
}

export interface OpticalRangeEstimate {
  rangeMeters: number;
  uncertaintyMeters: number;
  tiltDegrees: number;
  focalLengthMm: number;
}

export interface KillSiteCachingEvaluation {
  objectClass: EdgeObjectClass;
  leafCoverScore: number;
  cachingDetected: boolean;
  threatTier: AlertTier;
  evacRadiusMeters: number;
  rangerDispatchRequired: boolean;
  priority: 'IMMEDIATE' | 'HIGH' | 'STANDARD';
  audioMode: 'STEALTH_TACTICAL' | 'NORMAL';
}

export interface MunicipalDispatchGeoJsonFeature {
  type: 'Feature';
  geometry: {
    type: 'Point';
    coordinates: [number, number]; // [longitude, latitude]
  };
  properties: {
    incident_id: string;
    tier: string;
    taxonomy: string;
    confidence: number;
    caching_detected: boolean;
    leaf_cover_score: number;
    evac_radius_m: number;
    ranger_dispatch_required: boolean;
    offline_capture_time: string;
    sha256: string;
  };
}

export interface EdgeVisionAnalysisResult {
  modelName: string;
  delegate: string;
  inferenceLatencyMs: number;
  primaryDetection: EdgeDetectionBox;
  detections: EdgeDetectionBox[];
  opticalRange: OpticalRangeEstimate;
  cachingEvaluation: KillSiteCachingEvaluation;
  privacyMaskApplied: boolean;
  blurredRegionsCount: number;
  sha256: string;
  evidenceFilename: string;
  altitudeMeters: number;
  utcTimestamp: string;
  anonymizedImageBase64?: string;
  geoJsonPayload: MunicipalDispatchGeoJsonFeature;
}

export interface PriorityKakiItem {
  id: string;
  addressHint: string;
  hazardScore: number;
  isNearSchool: boolean;
  distanceToSchoolMeters: number;
  ownerConsentStatus: 'GRANTED' | 'ORDINANCE_CLEARANCE' | 'PENDING_CONSENT';
  permitId: string;
  clearanceOrder: number;
  fruitLoad: string;
  isFelled: boolean;
  lat: number;
  lon: number;
}

export interface CarcassDispatchItem {
  id: string;
  locationHint: string;
  species: string;
  decayState: string;
  threatScore: number;
  isDispatched: boolean;
}

export interface DispatchRecommendation {
  actionType: string;
  location: string;
  priority: 'IMMEDIATE' | 'HIGH' | 'SCHEDULED';
  justification: string;
}

export interface AuthUser {
  uid: string;
  email: string;
  role: UserRole;
  idToken: string;
  refreshToken: string;
}

export interface GeofenceTransitionEvent {
  zoneId: string;
  zoneName: string;
  transitionType: 'ENTER' | 'DWELL' | 'EXIT';
  tier: AlertTier;
  distanceMeters: number;
  timestamp: string;
  platform: 'native' | 'web-simulation';
}
