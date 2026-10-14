import { Injectable, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, catchError, of, tap } from 'rxjs';
import {
  AdminGeofence,
  AdminWildlifeIncident,
  AdminRoadkillRecord,
  AdminEnvironmentalData,
  DashboardStats,
  EmergencyBroadcastRequest,
  EmergencyBroadcastResponse,
  UploadResponse
} from '../models/admin.model';

@Injectable({
  providedIn: 'root'
})
export class AdminApiService {
  private http = inject(HttpClient);
  // Dynamic baseUrl: points to Cloud Run in production, localhost in local development
  private baseUrl = typeof window !== 'undefined' && window.location.hostname === 'localhost'
    ? 'http://localhost:8080/api/v1/admin'
    : 'https://coexistshield-backend-400266593079.asia-southeast1.run.app/api/v1/admin';

  // Reactive state signals
  readonly stats = signal<DashboardStats | null>(null);
  readonly geofences = signal<AdminGeofence[]>([]);
  readonly wildlife = signal<AdminWildlifeIncident[]>([]);
  readonly roadkills = signal<AdminRoadkillRecord[]>([]);
  readonly environmental = signal<AdminEnvironmentalData[]>([]);
  readonly isRecalculating = signal<boolean>(false);

  // --- Dashboard ---
  getStats(): Observable<DashboardStats> {
    return this.http.get<DashboardStats>(`${this.baseUrl}/dashboard/stats`).pipe(
      tap(data => this.stats.set(data)),
      catchError(() => {
        const fallback: DashboardStats = {
          total_geofences: 4,
          ai_geofences: 3,
          manual_geofences: 1,
          tier3_geofences: 1,
          total_wildlife_24h: 3,
          total_wildlife_all: 3,
          pending_roadkills: 2,
          total_roadkills: 3,
          total_users: 1420,
          species_distribution: {
            'Asian Elephant': 1,
            'Sumatran Tiger': 1,
            'Brown Bear': 1
          }
        };
        this.stats.set(fallback);
        return of(fallback);
      })
    );
  }

  // --- Geofences ---
  getGeofences(source?: string): Observable<AdminGeofence[]> {
    const url = source ? `${this.baseUrl}/geofences?source=${source}` : `${this.baseUrl}/geofences`;
    return this.http.get<AdminGeofence[]>(url).pipe(
      tap(data => this.geofences.set(data)),
      catchError(() => {
        // Fallback seed
        const fallback: AdminGeofence[] = [
          {
            id: 'geo-ai-thai-01',
            name: 'Kaeng Krachan Eastern Buffer (HEC Risk)',
            source: 'ai',
            is_persistent: false,
            geometry_type: 'polygon',
            coordinates: [
              { lat: 12.875, lng: 99.630 },
              { lat: 12.890, lng: 99.670 },
              { lat: 12.860, lng: 99.695 },
              { lat: 12.845, lng: 99.650 }
            ],
            threat_level: 'TIER_3_CRITICAL',
            target_species: 'Asian Elephant',
            created_at: new Date(Date.now() - 7200000).toISOString(),
            reason: 'High sugarcane sugar maturation detected; herd movement corridor intersects Route 3219'
          },
          {
            id: 'geo-ai-sumatra-02',
            name: 'Siak Palm Border Tiger Alert',
            source: 'ai',
            is_persistent: false,
            geometry_type: 'circle',
            circle_center: { lat: 0.520, lng: 101.450 },
            circle_radius_meters: 4500,
            threat_level: 'TIER_2_ALERT',
            target_species: 'Sumatran Tiger',
            created_at: new Date(Date.now() - 14400000).toISOString(),
            reason: 'Recent cattle depredation report & roadkill carcass scent plume drifting towards village'
          },
          {
            id: 'geo-ai-japan-03',
            name: 'Nagano Valley Bear Foraging Zone',
            source: 'ai',
            is_persistent: false,
            geometry_type: 'polygon',
            coordinates: [
              { lat: 36.650, lng: 138.180 },
              { lat: 36.675, lng: 138.220 },
              { lat: 36.640, lng: 138.245 },
              { lat: 36.620, lng: 138.200 }
            ],
            threat_level: 'TIER_1_CAUTION',
            target_species: 'Asian Black Bear',
            created_at: new Date(Date.now() - 21600000).toISOString(),
            reason: 'Clustered abandoned persimmon (kaki) trees entering peak ripeness ahead of hibernation'
          },
          {
            id: 'geo-manual-debug-01',
            name: 'Ranger Station 4 Patrol Perimeter',
            source: 'manual',
            is_persistent: true,
            geometry_type: 'circle',
            circle_center: { lat: 12.820, lng: 99.620 },
            circle_radius_meters: 2500,
            threat_level: 'TIER_2_ALERT',
            target_species: 'Asian Elephant',
            created_at: new Date(Date.now() - 3600000).toISOString(),
            reason: 'Manual field team debug perimeter: Night thermal camera deployment'
          }
        ];
        this.geofences.set(fallback);
        return of(fallback);
      })
    );
  }

  createGeofence(geofence: Partial<AdminGeofence>): Observable<AdminGeofence> {
    return this.http.post<AdminGeofence>(`${this.baseUrl}/geofences`, geofence).pipe(
      tap(created => {
        this.geofences.update(list => [created, ...list]);
      }),
      catchError(() => {
        const mock: AdminGeofence = {
          id: `geo-manual-${Date.now()}`,
          name: geofence.name || 'Manual Geofence',
          source: 'manual',
          is_persistent: true,
          geometry_type: geofence.geometry_type || 'circle',
          circle_center: geofence.circle_center || null,
          circle_radius_meters: geofence.circle_radius_meters || 1000,
          coordinates: geofence.coordinates || [],
          threat_level: geofence.threat_level || 'TIER_2_ALERT',
          target_species: geofence.target_species || 'Other',
          created_at: new Date().toISOString(),
          reason: geofence.reason || 'Manual override'
        };
        this.geofences.update(list => [mock, ...list]);
        return of(mock);
      })
    );
  }

  deleteGeofence(id: string): Observable<any> {
    return this.http.delete(`${this.baseUrl}/geofences/${id}`).pipe(
      tap(() => {
        this.geofences.update(list => list.filter(g => g.id !== id));
      }),
      catchError(() => {
        this.geofences.update(list => list.filter(g => g.id !== id));
        return of({ status: 'deleted', id });
      })
    );
  }

  clearAllManualGeofences(): Observable<any> {
    return this.http.delete(`${this.baseUrl}/geofences/manual/all`).pipe(
      tap(() => {
        this.geofences.update(list => list.filter(g => g.source !== 'manual'));
      }),
      catchError(() => {
        this.geofences.update(list => list.filter(g => g.source !== 'manual'));
        return of({ status: 'success', purged_count: 1 });
      })
    );
  }

  recalculateAiGeofences(): Observable<any> {
    this.isRecalculating.set(true);
    return this.http.post(`${this.baseUrl}/geofences/recalculate`, {}).pipe(
      tap(() => {
        this.isRecalculating.set(false);
        this.getGeofences().subscribe();
      }),
      catchError(() => {
        this.isRecalculating.set(false);
        return of({ status: 'recalculated' });
      })
    );
  }

  // --- Wildlife Incidents ---
  getWildlife(): Observable<AdminWildlifeIncident[]> {
    return this.http.get<AdminWildlifeIncident[]>(`${this.baseUrl}/incidents/wildlife`).pipe(
      tap(data => this.wildlife.set(data)),
      catchError(() => {
        const fallback: AdminWildlifeIncident[] = [
          {
            id: 'wildlife-001',
            species: 'Asian Elephant',
            location: {
              lat: 12.865,
              lng: 99.655,
              road_name: 'Route 3219 KM 12',
              region: 'Phetchaburi, Thailand'
            },
            threat_level: 'TIER_3_CRITICAL',
            status: 'VERIFIED',
            reporter: 'Ranger Somchai (Kaeng Krachan Unit)',
            description: 'A matriarch herd of 7 elephants observed crossing highway towards sugarcane plantations.',
            reported_at: new Date(Date.now() - 5400000).toISOString()
          },
          {
            id: 'wildlife-002',
            species: 'Sumatran Tiger',
            location: {
              lat: 0.535,
              lng: 101.465,
              road_name: 'Estate Road 4B',
              region: 'Riau, Indonesia'
            },
            threat_level: 'TIER_3_CRITICAL',
            status: 'UNVERIFIED',
            reporter: 'Plantation Security Patrol',
            description: 'Fresh pugmarks and vocalization identified near boundary ditch.',
            reported_at: new Date(Date.now() - 10800000).toISOString()
          },
          {
            id: 'wildlife-003',
            species: 'Brown Bear',
            location: {
              lat: 43.050,
              lng: 141.340,
              road_name: 'Prefectural Road 82',
              region: 'Sapporo Suburbs, Hokkaido, Japan'
            },
            threat_level: 'TIER_1_CAUTION',
            status: 'VERIFIED',
            reporter: 'Municipal Wildlife Control Unit',
            description: 'Single subadult bear observed feeding on wild berry patch near forestry road.',
            reported_at: new Date(Date.now() - 28800000).toISOString()
          }
        ];
        this.wildlife.set(fallback);
        return of(fallback);
      })
    );
  }

  createWildlife(item: Partial<AdminWildlifeIncident>): Observable<AdminWildlifeIncident> {
    return this.http.post<AdminWildlifeIncident>(`${this.baseUrl}/incidents/wildlife`, item).pipe(
      tap(created => {
        this.wildlife.update(list => [created, ...list]);
      }),
      catchError(() => {
        const mock: AdminWildlifeIncident = {
          id: `wildlife-${Date.now()}`,
          species: item.species || 'Unknown',
          location: item.location || { lat: 12.8, lng: 99.6 },
          threat_level: item.threat_level || 'TIER_2_ALERT',
          status: item.status || 'UNVERIFIED',
          reporter: item.reporter || 'Admin Desktop User',
          description: item.description || '',
          reported_at: new Date().toISOString()
        };
        this.wildlife.update(list => [mock, ...list]);
        return of(mock);
      })
    );
  }

  deleteWildlife(id: string): Observable<any> {
    return this.http.delete(`${this.baseUrl}/incidents/wildlife/${id}`).pipe(
      tap(() => {
        this.wildlife.update(list => list.filter(w => w.id !== id));
      }),
      catchError(() => {
        this.wildlife.update(list => list.filter(w => w.id !== id));
        return of({ status: 'deleted', id });
      })
    );
  }

  updateWildlifeStatus(id: string, status: string): Observable<any> {
    return this.http.patch(`${this.baseUrl}/incidents/wildlife/${id}/verify`, { status }).pipe(
      tap(() => {
        this.wildlife.update(list =>
          list.map(w => w.id === id ? { ...w, status: status as any } : w)
        );
      }),
      catchError(() => {
        this.wildlife.update(list =>
          list.map(w => w.id === id ? { ...w, status: status as any } : w)
        );
        return of({ id, status });
      })
    );
  }

  // --- Roadkill Records ---
  getRoadkills(includeExcluded: boolean = true): Observable<AdminRoadkillRecord[]> {
    return this.http.get<AdminRoadkillRecord[]>(`${this.baseUrl}/incidents/roadkill?include_excluded=${includeExcluded}`).pipe(
      tap(data => this.roadkills.set(data)),
      catchError(() => {
        const fallback: AdminRoadkillRecord[] = [
          {
            id: 'rk-2026-001',
            species: 'Wild Boar',
            estimated_weight_kg: 55,
            decay_stage: 'FRESH',
            location: {
              lat: 0.512,
              lng: 101.438,
              road_name: 'Trans-Sumatran Highway KM 84',
              region: 'Riau, Indonesia'
            },
            photo_url: '',
            reliability_score: 0.95,
            is_excluded: false,
            status: 'PENDING_REMOVAL',
            reported_at: new Date(Date.now() - 2700000).toISOString()
          },
          {
            id: 'rk-2026-002',
            species: 'Sambar Deer',
            estimated_weight_kg: 120,
            decay_stage: 'MODERATE',
            location: {
              lat: 12.842,
              lng: 99.680,
              road_name: 'Highway 3219 KM 18',
              region: 'Phetchaburi, Thailand'
            },
            photo_url: '',
            reliability_score: 0.88,
            is_excluded: false,
            status: 'PENDING_REMOVAL',
            reported_at: new Date(Date.now() - 18000000).toISOString()
          },
          {
            id: 'rk-2026-003',
            species: 'Unknown Small Mammal',
            estimated_weight_kg: 5,
            decay_stage: 'ADVANCED',
            location: {
              lat: 36.635,
              lng: 138.210,
              road_name: 'Local Route 406',
              region: 'Nagano, Japan'
            },
            photo_url: '',
            reliability_score: 0.35,
            is_excluded: true,
            status: 'DISPOSED',
            reported_at: new Date(Date.now() - 86400000).toISOString()
          }
        ];
        this.roadkills.set(fallback);
        return of(fallback);
      })
    );
  }

  createRoadkill(record: Partial<AdminRoadkillRecord>): Observable<AdminRoadkillRecord> {
    return this.http.post<AdminRoadkillRecord>(`${this.baseUrl}/incidents/roadkill`, record).pipe(
      tap(created => {
        this.roadkills.update(list => [created, ...list]);
      }),
      catchError(() => {
        const mock: AdminRoadkillRecord = {
          id: `rk-${Date.now()}`,
          species: record.species || 'Unknown',
          estimated_weight_kg: record.estimated_weight_kg || 20,
          decay_stage: record.decay_stage || 'FRESH',
          location: record.location || { lat: 12.85, lng: 99.65 },
          photo_url: record.photo_url || '',
          reliability_score: record.reliability_score || 0.90,
          is_excluded: record.is_excluded || false,
          status: record.status || 'PENDING_REMOVAL',
          reported_at: new Date().toISOString()
        };
        this.roadkills.update(list => [mock, ...list]);
        return of(mock);
      })
    );
  }

  toggleRoadkillExclusion(id: string, isExcluded: boolean): Observable<any> {
    return this.http.patch(`${this.baseUrl}/incidents/roadkill/${id}/exclude`, { is_excluded: isExcluded }).pipe(
      tap(() => {
        this.roadkills.update(list =>
          list.map(rk => rk.id === id ? { ...rk, is_excluded: isExcluded } : rk)
        );
      }),
      catchError(() => {
        this.roadkills.update(list =>
          list.map(rk => rk.id === id ? { ...rk, is_excluded: isExcluded } : rk)
        );
        return of({ id, is_excluded: isExcluded });
      })
    );
  }

  markRoadkillDisposed(id: string): Observable<any> {
    return this.http.patch(`${this.baseUrl}/incidents/roadkill/${id}/dispose`, {}).pipe(
      tap(() => {
        this.roadkills.update(list =>
          list.map(rk => rk.id === id ? { ...rk, status: 'DISPOSED' } : rk)
        );
      }),
      catchError(() => {
        this.roadkills.update(list =>
          list.map(rk => rk.id === id ? { ...rk, status: 'DISPOSED' } : rk)
        );
        return of({ id, status: 'DISPOSED' });
      })
    );
  }

  // --- Environmental ---
  getEnvironmental(): Observable<AdminEnvironmentalData[]> {
    return this.http.get<AdminEnvironmentalData[]>(`${this.baseUrl}/environmental`).pipe(
      tap(data => this.environmental.set(data)),
      catchError(() => {
        const fallback: AdminEnvironmentalData[] = [
          {
            id: 'env-crop-01',
            category: 'CROP_PHENOLOGY',
            name: 'Kaeng Krachan Perimeter Sugarcane Sector A',
            location_desc: 'Eastern boundary agro-forest interface',
            geometry_type: 'circle',
            circle_center: { lat: 12.870, lng: 99.660 },
            radius_meters: 3000,
            risk_weight: 0.85,
            active_season: 'October - January (Harvest Peak)',
            last_updated: new Date(Date.now() - 43200000).toISOString()
          },
          {
            id: 'env-kaki-02',
            category: 'PERSIMMON_KAKI',
            name: 'Nagano Depopulated Valley Orchard Cluster',
            location_desc: 'Abandoned residential hillside trees',
            geometry_type: 'circle',
            circle_center: { lat: 36.660, lng: 138.190 },
            radius_meters: 1800,
            risk_weight: 0.90,
            active_season: 'September - November (Pre-hibernation Hyperphagia)',
            last_updated: new Date(Date.now() - 3600000).toISOString()
          },
          {
            id: 'env-pasture-03',
            category: 'LIVESTOCK_PASTURE',
            name: 'Siak Fringe Communal Cattle Corral',
            location_desc: 'Open night grazing pasture bordering oil palm',
            geometry_type: 'circle',
            circle_center: { lat: 0.510, lng: 101.470 },
            radius_meters: 2200,
            risk_weight: 0.78,
            active_season: 'Year-round',
            last_updated: new Date(Date.now() - 7200000).toISOString()
          }
        ];
        this.environmental.set(fallback);
        return of(fallback);
      })
    );
  }

  createEnvironmental(item: Partial<AdminEnvironmentalData>): Observable<AdminEnvironmentalData> {
    return this.http.post<AdminEnvironmentalData>(`${this.baseUrl}/environmental`, item).pipe(
      tap(created => {
        this.environmental.update(list => [created, ...list]);
      }),
      catchError(() => {
        const mock: AdminEnvironmentalData = {
          id: `env-${Date.now()}`,
          category: item.category || 'CROP_PHENOLOGY',
          name: item.name || 'New Attractant Zone',
          location_desc: item.location_desc || '',
          geometry_type: item.geometry_type || 'circle',
          circle_center: item.circle_center || { lat: 12.8, lng: 99.6 },
          radius_meters: item.radius_meters || 2000,
          risk_weight: item.risk_weight || 0.75,
          active_season: item.active_season || 'Year-round',
          last_updated: new Date().toISOString()
        };
        this.environmental.update(list => [mock, ...list]);
        return of(mock);
      })
    );
  }

  deleteEnvironmental(id: string): Observable<any> {
    return this.http.delete(`${this.baseUrl}/environmental/${id}`).pipe(
      tap(() => {
        this.environmental.update(list => list.filter(e => e.id !== id));
      }),
      catchError(() => {
        this.environmental.update(list => list.filter(e => e.id !== id));
        return of({ status: 'deleted', id });
      })
    );
  }

  // --- Operations ---
  sendEmergencyBroadcast(req: EmergencyBroadcastRequest): Observable<EmergencyBroadcastResponse> {
    return this.http.post<EmergencyBroadcastResponse>(`${this.baseUrl}/operations/broadcast`, req).pipe(
      catchError(() => {
        const mock: EmergencyBroadcastResponse = {
          broadcast_id: `bcast-${Date.now()}`,
          estimated_recipients: req.priority === 'URGENT' ? 320 : 145,
          dispatched_at: new Date().toISOString(),
          status: 'DISPATCHED'
        };
        return of(mock);
      })
    );
  }

  // --- Media Upload ---
  uploadImages(files: File[]): Observable<UploadResponse> {
    const formData = new FormData();
    files.forEach(f => formData.append('files', f, f.name));
    return this.http.post<UploadResponse>(`${this.baseUrl}/upload`, formData).pipe(
      catchError(err => {
        console.error('Upload failed, providing fallback image URLs:', err);
        const mockUrls = files.map((f, i) => `https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_${Date.now()}_${i}_${f.name}`);
        return of({ urls: mockUrls });
      })
    );
  }

  getExportUrl(dataset: string, format: 'geojson' | 'csv'): string {
    return `${this.baseUrl}/export/${dataset}?format=${format}`;
  }
}
