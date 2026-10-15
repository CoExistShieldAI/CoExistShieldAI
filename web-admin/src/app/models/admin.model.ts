export interface Coordinate {
  lat: number;
  lng: number;
}

export interface IncidentLocation {
  lat: number;
  lng: number;
  road_name?: string;
  region?: string;
}

export type ThreatLevel = 'TIER_1_CAUTION' | 'TIER_2_ALERT' | 'TIER_3_CRITICAL';
export type GeofenceSource = 'ai' | 'manual';
export type GeometryType = 'circle' | 'polygon';

export interface AdminGeofence {
  id: string;
  name: string;
  source: GeofenceSource;
  is_persistent: boolean;
  geometry_type: GeometryType;
  coordinates?: Coordinate[];
  circle_center?: Coordinate | null;
  circle_radius_meters?: number;
  threat_level: ThreatLevel;
  target_species: string;
  created_at: string;
  expires_at?: string | null;
  reason: string;
}

export type WildlifeStatus = 'UNVERIFIED' | 'VERIFIED' | 'RESOLVED' | 'FALSE_ALARM';

export interface AdminWildlifeIncident {
  id: string;
  species: string;
  location: IncidentLocation;
  threat_level: ThreatLevel;
  status: WildlifeStatus;
  reporter: string;
  description: string;
  photo_urls?: string[];
  reported_at: string;
}

export type DecayStage = 'FRESH' | 'MODERATE' | 'ADVANCED';
export type RoadkillStatus = 'PENDING_REMOVAL' | 'DISPOSED';

export interface AdminRoadkillRecord {
  id: string;
  species: string;
  estimated_weight_kg: number;
  decay_stage: DecayStage;
  location: IncidentLocation;
  photo_url?: string;
  photo_urls?: string[];
  reliability_score: number;
  is_excluded: boolean;
  status: RoadkillStatus;
  reported_at: string;
}

export type AttractantCategory = 'CROP_PHENOLOGY' | 'PERSIMMON_KAKI' | 'LIVESTOCK_PASTURE' | 'WATER_SOURCE';

export interface AdminEnvironmentalData {
  id: string;
  category: AttractantCategory;
  name: string;
  location_desc: string;
  geometry_type: GeometryType;
  coordinates?: Coordinate[];
  circle_center?: Coordinate | null;
  radius_meters?: number;
  risk_weight: number;
  active_season: string;
  photo_urls?: string[];
  last_updated: string;
}

export interface UploadResponse {
  urls: string[];
}

export interface DashboardStats {
  total_geofences: number;
  ai_geofences: number;
  manual_geofences: number;
  tier3_geofences: number;
  total_wildlife_24h: number;
  total_wildlife_all: number;
  pending_roadkills: number;
  total_roadkills: number;
  total_weather_stations?: number;
  total_users: number;
  species_distribution: Record<string, number>;
}

export type WeatherCondition = 'CLEAR' | 'OVERCAST' | 'RAIN' | 'HEAVY_RAIN' | 'STORM' | 'FOG';

export type MoonPhase =
  | 'NEW_MOON'
  | 'WAXING_CRESCENT'
  | 'FIRST_QUARTER'
  | 'WAXING_GIBBOUS'
  | 'FULL_MOON'
  | 'WANING_GIBBOUS'
  | 'LAST_QUARTER'
  | 'WANING_CRESCENT';

export interface AdminWeatherData {
  id: string;
  station_name: string;
  region: string;
  latitude: number;
  longitude: number;
  temperature_celsius: number;
  humidity_pct: number;
  wind_speed_kmh: number;
  wind_direction_deg: number;
  precipitation_mm_h: number;
  moon_illumination: number;
  moon_phase: MoonPhase | string;
  is_flooded_river_risk: boolean;
  weather_condition: WeatherCondition | string;
  notes?: string;
  last_updated: string;
}

export interface EmergencyBroadcastRequest {
  geofence_id?: string;
  radius_km?: number;
  center_lat?: number;
  center_lng?: number;
  title: string;
  body: string;
  priority: 'HIGH' | 'URGENT';
}

export interface EmergencyBroadcastResponse {
  broadcast_id: string;
  estimated_recipients: number;
  dispatched_at: string;
  status: string;
}

