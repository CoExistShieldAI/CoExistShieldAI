import { Capacitor, WebPlugin, registerPlugin, PluginListenerHandle } from '@capacitor/core';
import { LocalNotifications } from '@capacitor/local-notifications';
import { Haptics, ImpactStyle } from '@capacitor/haptics';
import { AlertTier, GeofenceTransitionEvent, GeofenceZone, GeoLocation } from '../models/shield.model';

export interface GeofenceBridgePlugin {
  registerGeofences(options: { zones: GeofenceZone[] }): Promise<{ registeredCount: number; platform: string }>;
  unregisterGeofences(options: { zoneIds: string[] }): Promise<{ removedCount: number }>;
  getRegisteredZones(): Promise<{ zoneIds: string[]; platform: string }>;
  evaluateLocation(options: { latitude: number; longitude: number }): Promise<{
    insideZones: string[];
    closestZoneId: string | null;
    closestDistanceMeters: number | null;
    highestTier: AlertTier;
  }>;
  addListener(
    eventName: 'geofenceTransition',
    listenerFunc: (event: GeofenceTransitionEvent) => void
  ): Promise<PluginListenerHandle>;
}

/**
 * Web & Hybrid Simulation implementation of GeofenceBridgePlugin.
 * When running inside Android/iOS Capacitor with the native GeofenceBridgePlugin registered,
 * Capacitor automatically routes calls to Kotlin/Swift GeofencingClient/CLLocationManager.
 * On Web (or during Hackathon Judge Simulation), this WebPlugin executes exact Haversine +
 * Ray-Casting polygon geometry checks in real time.
 */
export class GeofenceBridgeWeb extends WebPlugin implements GeofenceBridgePlugin {
  private activeZones = new Map<string, GeofenceZone>();
  private currentlyInsideIds = new Set<string>();

  async registerGeofences(options: { zones: GeofenceZone[] }): Promise<{ registeredCount: number; platform: string }> {
    this.activeZones.clear();
    for (const zone of options.zones) {
      this.activeZones.set(zone.id, zone);
    }
    return {
      registeredCount: this.activeZones.size,
      platform: Capacitor.isNativePlatform() ? Capacitor.getPlatform() : 'web-simulation'
    };
  }

  async unregisterGeofences(options: { zoneIds: string[] }): Promise<{ removedCount: number }> {
    let removed = 0;
    for (const id of options.zoneIds) {
      if (this.activeZones.delete(id)) {
        removed++;
      }
      this.currentlyInsideIds.delete(id);
    }
    return { removedCount: removed };
  }

  async getRegisteredZones(): Promise<{ zoneIds: string[]; platform: string }> {
    return {
      zoneIds: Array.from(this.activeZones.keys()),
      platform: Capacitor.isNativePlatform() ? Capacitor.getPlatform() : 'web-simulation'
    };
  }

  async evaluateLocation(options: { latitude: number; longitude: number }): Promise<{
    insideZones: string[];
    closestZoneId: string | null;
    closestDistanceMeters: number | null;
    highestTier: AlertTier;
  }> {
    const point: GeoLocation = { latitude: options.latitude, longitude: options.longitude };
    const insideZones: string[] = [];
    let closestZoneId: string | null = null;
    let closestDistanceMeters: number | null = null;
    let highestTier: AlertTier = 'SAFE';

    const tierRank: Record<AlertTier, number> = {
      SAFE: 0,
      TIER_1_NOTICE: 1,
      TIER_2_WARNING: 2,
      TIER_3_DANGER: 3
    };

    for (const zone of this.activeZones.values()) {
      const centerDistMeters = haversineMeters(point, zone.centroid);
      const effectiveBoundaryDist = Math.max(0, centerDistMeters - (zone.radiusMeters || 1500));

      const isInsidePolygon =
        zone.coordinates && zone.coordinates.length >= 3
          ? isPointInPolygon(point, zone.coordinates)
          : false;
      const isInsideCircle = centerDistMeters <= (zone.radiusMeters || 1500);
      const isInside = isInsidePolygon || isInsideCircle;

      const distanceMetric = isInside ? Math.min(centerDistMeters, 180) : effectiveBoundaryDist;

      if (closestDistanceMeters === null || distanceMetric < closestDistanceMeters) {
        closestDistanceMeters = Math.round(distanceMetric);
        closestZoneId = zone.id;
      }

      if (isInside) {
        insideZones.push(zone.id);
        if (tierRank[zone.tier] > tierRank[highestTier]) {
          highestTier = zone.tier;
        }

        // Fire ENTER transition if newly entered
        if (!this.currentlyInsideIds.has(zone.id)) {
          this.currentlyInsideIds.add(zone.id);
          const transitionEvent: GeofenceTransitionEvent = {
            zoneId: zone.id,
            zoneName: zone.name,
            transitionType: 'ENTER',
            tier: zone.tier,
            distanceMeters: Math.round(distanceMetric),
            timestamp: new Date().toISOString(),
            platform: Capacitor.isNativePlatform() ? 'native' : 'web-simulation'
          };
          this.notifyListeners('geofenceTransition', transitionEvent);
          this.triggerDeviceFeedback(transitionEvent);
        }
      } else {
        // Proximity-based graduated tier evaluation for adjacent buffer areas:
        // Exiting or approaching a Tier 3 Critical zone doesn't abruptly flip between Danger and Safe.
        // Adjacent areas gracefully transition through Tier 2 Warning and Tier 1 Advisory halos.
        if (zone.tier === 'TIER_3_DANGER') {
          const t2Threshold = Math.max(1200, (zone.radiusMeters || 1500) * 0.75);
          const t1Threshold = Math.max(2500, (zone.radiusMeters || 1500) * 1.60);
          if (effectiveBoundaryDist <= t2Threshold) {
            if (tierRank['TIER_2_WARNING'] > tierRank[highestTier]) {
              highestTier = 'TIER_2_WARNING';
            }
          } else if (effectiveBoundaryDist <= t1Threshold) {
            if (tierRank['TIER_1_NOTICE'] > tierRank[highestTier]) {
              highestTier = 'TIER_1_NOTICE';
            }
          }
        } else if (zone.tier === 'TIER_2_WARNING') {
          const t1Threshold = Math.max(1500, (zone.radiusMeters || 1500) * 0.75);
          if (effectiveBoundaryDist <= t1Threshold) {
            if (tierRank['TIER_1_NOTICE'] > tierRank[highestTier]) {
              highestTier = 'TIER_1_NOTICE';
            }
          }
        }

        if (this.currentlyInsideIds.has(zone.id)) {
          this.currentlyInsideIds.delete(zone.id);

          // Graceful step-down tier on zone exit
          let exitTier: AlertTier = 'SAFE';
          if (zone.tier === 'TIER_3_DANGER') {
            const t2Threshold = Math.max(1200, (zone.radiusMeters || 1500) * 0.75);
            const t1Threshold = Math.max(2500, (zone.radiusMeters || 1500) * 1.60);
            if (effectiveBoundaryDist <= t2Threshold) {
              exitTier = 'TIER_2_WARNING';
            } else if (effectiveBoundaryDist <= t1Threshold) {
              exitTier = 'TIER_1_NOTICE';
            }
          } else if (zone.tier === 'TIER_2_WARNING') {
            const t1Threshold = Math.max(1500, (zone.radiusMeters || 1500) * 0.75);
            if (effectiveBoundaryDist <= t1Threshold) {
              exitTier = 'TIER_1_NOTICE';
            }
          }

          const exitEvent: GeofenceTransitionEvent = {
            zoneId: zone.id,
            zoneName: zone.name,
            transitionType: 'EXIT',
            tier: exitTier,
            distanceMeters: Math.round(distanceMetric),
            timestamp: new Date().toISOString(),
            platform: Capacitor.isNativePlatform() ? 'native' : 'web-simulation'
          };
          this.notifyListeners('geofenceTransition', exitEvent);
        }
      }
    }

    return {
      insideZones,
      closestZoneId,
      closestDistanceMeters,
      highestTier
    };
  }

  private async triggerDeviceFeedback(event: GeofenceTransitionEvent): Promise<void> {
    try {
      if (Capacitor.isNativePlatform()) {
        await Haptics.impact({
          style: event.tier === 'TIER_3_DANGER' ? ImpactStyle.Heavy : ImpactStyle.Medium
        });
        await LocalNotifications.schedule({
          notifications: [
            {
              id: Math.floor(Math.random() * 100000),
              title: `[${event.tier}] CoExist Shield Geofence Alert`,
              body: `Entered ${event.zoneName} (${event.distanceMeters}m proximity)`,
              schedule: { at: new Date(Date.now() + 100) }
            }
          ]
        });
      }
    } catch {
      // Ignore if notification/haptic permissions are not granted in browser
    }
  }
}

export function haversineMeters(a: GeoLocation, b: GeoLocation): number {
  const R = 6371000; // Earth radius in meters
  const dLat = ((b.latitude - a.latitude) * Math.PI) / 180;
  const dLon = ((b.longitude - a.longitude) * Math.PI) / 180;
  const lat1 = (a.latitude * Math.PI) / 180;
  const lat2 = (b.latitude * Math.PI) / 180;

  const sinDLat = Math.sin(dLat / 2);
  const sinDLon = Math.sin(dLon / 2);
  const h = sinDLat * sinDLat + Math.cos(lat1) * Math.cos(lat2) * sinDLon * sinDLon;
  return 2 * R * Math.atan2(Math.sqrt(h), Math.sqrt(1 - h));
}

export function isPointInPolygon(point: GeoLocation, polygon: GeoLocation[]): boolean {
  let inside = false;
  for (let i = 0, j = polygon.length - 1; i < polygon.length; j = i++) {
    const xi = polygon[i].latitude;
    const yi = polygon[i].longitude;
    const xj = polygon[j].latitude;
    const yj = polygon[j].longitude;

    const intersect =
      yi > point.longitude !== yj > point.longitude &&
      point.latitude < ((xj - xi) * (point.longitude - yi)) / (yj - yi + 1e-12) + xi;
    if (intersect) inside = !inside;
  }
  return inside;
}

export const GeofenceBridge = registerPlugin<GeofenceBridgePlugin>('GeofenceBridge', {
  web: () => new GeofenceBridgeWeb()
});
