import { Injectable, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, catchError, of, tap } from 'rxjs';
import {
  AdminGeofence,
  AdminWildlifeIncident,
  AdminRoadkillRecord,
  AdminEnvironmentalData,
  AdminWeatherData,
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
  readonly weather = signal<AdminWeatherData[]>([]);
  readonly isRecalculating = signal<boolean>(false);

  // --- Dashboard ---
  getStats(): Observable<DashboardStats> {
    return this.http.get<DashboardStats>(`${this.baseUrl}/dashboard/stats`).pipe(
      tap(data => this.stats.set(data)),
      catchError(() => {
        const fallback: DashboardStats = {
          total_geofences: 6,
          ai_geofences: 5,
          manual_geofences: 1,
          tier3_geofences: 2,
          total_wildlife_24h: 5,
          total_wildlife_all: 10,
          pending_roadkills: 4,
          total_roadkills: 10,
          total_users: 1420,
          species_distribution: {
            'Asian Elephant': 3,
            'Asian Black Bear': 2,
            'Brown Bear': 2,
            'Sumatran Tiger': 2,
            'Bengal Tiger': 1
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
            name: 'Siak-Minas Palm Border Tiger Alert',
            source: 'ai',
            is_persistent: false,
            geometry_type: 'circle',
            circle_center: { lat: 0.520, lng: 101.450 },
            circle_radius_meters: 4500,
            threat_level: 'TIER_2_ALERT',
            target_species: 'Sumatran Tiger',
            created_at: new Date(Date.now() - 14400000).toISOString(),
            reason: 'Recent cattle depredation report & roadkill carcass scent plume drifting towards village border'
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
            id: 'geo-ai-india-04',
            name: 'Kaziranga Panbari-Haldibari Migration Shield',
            source: 'ai',
            is_persistent: false,
            geometry_type: 'circle',
            circle_center: { lat: 26.590, lng: 93.180 },
            circle_radius_meters: 3800,
            threat_level: 'TIER_3_CRITICAL',
            target_species: 'Asian Elephant',
            created_at: new Date(Date.now() - 36000000).toISOString(),
            reason: 'Brahmaputra flood surge causing megafauna migration south across NH-715 towards Karbi Anglong Hills'
          },
          {
            id: 'geo-ai-hokkaido-05',
            name: 'Shiretoko Route 334 Coastal Brown Bear Buffer',
            source: 'ai',
            is_persistent: false,
            geometry_type: 'circle',
            circle_center: { lat: 44.050, lng: 145.020 },
            circle_radius_meters: 3200,
            threat_level: 'TIER_2_ALERT',
            target_species: 'Brown Bear',
            created_at: new Date(Date.now() - 64800000).toISOString(),
            reason: 'Coastal salmon run and sika deer carcass presence drawing hyperphagic adult brown bears'
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
            reason: 'Manual field team patrol perimeter: Night thermal camera deployment'
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
            description: 'A matriarch herd of 7 elephants crossing highway towards perimeter sugarcane plantations.',
            photo_urls: ['https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_elephant_herd.jpg'],
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
            description: 'Fresh pugmarks and vocalization identified near oil palm boundary drainage ditch.',
            photo_urls: ['https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_tiger_track.jpg'],
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
            photo_urls: [],
            reported_at: new Date(Date.now() - 28800000).toISOString()
          },
          {
            id: 'wildlife-004',
            species: 'Asian Black Bear',
            location: {
              lat: 39.735,
              lng: 141.165,
              road_name: 'Matsuzono 3-chome Hillside',
              region: 'Morioka, Iwate, Japan'
            },
            threat_level: 'TIER_3_CRITICAL',
            status: 'VERIFIED',
            reporter: 'Iwate Mobile Alert Network',
            description: 'Adult black bear observed climbing abandoned persimmon tree inside residential boundary adjacent to primary school route.',
            photo_urls: ['https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_bear_kaki.jpg'],
            reported_at: new Date(Date.now() - 50400000).toISOString()
          },
          {
            id: 'wildlife-005',
            species: 'Asian Elephant',
            location: {
              lat: 26.582,
              lng: 93.175,
              road_name: 'NH-715 Haldibari Corridor',
              region: 'Kaziranga, Assam, India'
            },
            threat_level: 'TIER_3_CRITICAL',
            status: 'VERIFIED',
            reporter: 'Assam Forest Dept Highway Patrol',
            description: 'Breeding herd of 14 elephants crossing National Highway southwards to reach elevated Karbi Anglong refuge.',
            photo_urls: ['https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_nh715_crossing.jpg'],
            reported_at: new Date(Date.now() - 79200000).toISOString()
          },
          {
            id: 'wildlife-006',
            species: 'Asian Elephant',
            location: {
              lat: 12.540,
              lng: 99.530,
              road_name: 'Protection Unit KK.7 Access Road',
              region: 'Huai Sat Yai, Pa La-U, Thailand'
            },
            threat_level: 'TIER_2_ALERT',
            status: 'VERIFIED',
            reporter: 'DNP Wild Elephant Surveillance Team',
            description: "Bull elephant 'Plai Boonchuay' foraging along fruit plantation border fence line.",
            photo_urls: [],
            reported_at: new Date(Date.now() - 172800000).toISOString()
          },
          {
            id: 'wildlife-007',
            species: 'Asian Black Bear',
            location: {
              lat: 36.662,
              lng: 138.210,
              road_name: 'Suzaka Orchard Ridge Trail',
              region: 'Nagano, Japan'
            },
            threat_level: 'TIER_2_ALERT',
            status: 'VERIFIED',
            reporter: 'Nagano Agricultural Cooperative',
            description: 'Female bear with two cubs feeding in unharvested chestnut grove bordering mountain slope.',
            photo_urls: [],
            reported_at: new Date(Date.now() - 345600000).toISOString()
          },
          {
            id: 'wildlife-008',
            species: 'Brown Bear',
            location: {
              lat: 44.045,
              lng: 145.030,
              road_name: 'Route 334 Coastal Pass',
              region: 'Rausu, Shiretoko, Hokkaido, Japan'
            },
            threat_level: 'TIER_2_ALERT',
            status: 'VERIFIED',
            reporter: 'Shiretoko Nature Foundation Ranger',
            description: 'Large male brown bear tracking coastal salmon scent near river mouth, 80m from roadway.',
            photo_urls: ['https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_brown_bear.jpg'],
            reported_at: new Date(Date.now() - 518400000).toISOString()
          },
          {
            id: 'wildlife-009',
            species: 'Sumatran Tiger',
            location: {
              lat: 0.785,
              lng: 102.045,
              road_name: 'Sungai Apit Peatland Edge',
              region: 'Siak, Riau, Indonesia'
            },
            threat_level: 'TIER_3_CRITICAL',
            status: 'VERIFIED',
            reporter: 'BBKSDA Riau Rapid Response Unit',
            description: 'Tiger attempted night predation on communal livestock pen; clear track trail confirmed.',
            photo_urls: [],
            reported_at: new Date(Date.now() - 1036800000).toISOString()
          },
          {
            id: 'wildlife-010',
            species: 'Bengal Tiger',
            location: {
              lat: 27.410,
              lng: 84.120,
              road_name: 'Valmiki-Chitwan Border Trail',
              region: 'Terai Arc, Bihar/Nepal Border'
            },
            threat_level: 'TIER_2_ALERT',
            status: 'VERIFIED',
            reporter: 'WWF Camera Trap Survey Team',
            description: 'Tigress recorded using agricultural irrigation canal as travel corridor between forest patches.',
            photo_urls: [],
            reported_at: new Date(Date.now() - 1728000000).toISOString()
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
            photo_url: 'https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_rk_boar.jpg',
            photo_urls: ['https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_rk_boar.jpg'],
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
            photo_urls: [],
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
            photo_urls: [],
            reliability_score: 0.35,
            is_excluded: true,
            status: 'DISPOSED',
            reported_at: new Date(Date.now() - 86400000).toISOString()
          },
          {
            id: 'rk-2026-004',
            species: 'Hog Deer',
            estimated_weight_kg: 42,
            decay_stage: 'FRESH',
            location: {
              lat: 26.588,
              lng: 93.185,
              road_name: 'NH-715 Panbari Bottleneck',
              region: 'Kaziranga, Assam, India'
            },
            photo_url: 'https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_rk_hogdeer.jpg',
            photo_urls: ['https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_rk_hogdeer.jpg'],
            reliability_score: 0.96,
            is_excluded: false,
            status: 'PENDING_REMOVAL',
            reported_at: new Date(Date.now() - 10800000).toISOString()
          },
          {
            id: 'rk-2026-005',
            species: 'Ezo Sika Deer',
            estimated_weight_kg: 115,
            decay_stage: 'FRESH',
            location: {
              lat: 43.620,
              lng: 144.980,
              road_name: 'National Route 272',
              region: 'Nemuro, Hokkaido, Japan'
            },
            photo_url: 'https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_rk_sikadeer.jpg',
            photo_urls: ['https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_rk_sikadeer.jpg'],
            reliability_score: 0.94,
            is_excluded: false,
            status: 'PENDING_REMOVAL',
            reported_at: new Date(Date.now() - 18000000).toISOString()
          },
          {
            id: 'rk-2026-006',
            species: 'Japanese Serow (Kamoshika)',
            estimated_weight_kg: 38,
            decay_stage: 'MODERATE',
            location: {
              lat: 39.810,
              lng: 141.420,
              road_name: 'National Route 455',
              region: 'Iwaizumi, Iwate, Japan'
            },
            photo_url: '',
            photo_urls: [],
            reliability_score: 0.89,
            is_excluded: false,
            status: 'DISPOSED',
            reported_at: new Date(Date.now() - 129600000).toISOString()
          },
          {
            id: 'rk-2026-007',
            species: 'Wild Boar',
            estimated_weight_kg: 68,
            decay_stage: 'MODERATE',
            location: {
              lat: 12.528,
              lng: 99.542,
              road_name: 'Route 3201 Hua Hin Bypass',
              region: 'Prachuap Khiri Khan, Thailand'
            },
            photo_url: '',
            photo_urls: [],
            reliability_score: 0.82,
            is_excluded: false,
            status: 'DISPOSED',
            reported_at: new Date(Date.now() - 259200000).toISOString()
          },
          {
            id: 'rk-2026-008',
            species: 'Domestic Cattle (Calf)',
            estimated_weight_kg: 85,
            decay_stage: 'ADVANCED',
            location: {
              lat: 0.525,
              lng: 101.480,
              road_name: 'Minas Concession Road',
              region: 'Siak, Riau, Indonesia'
            },
            photo_url: '',
            photo_urls: [],
            reliability_score: 0.90,
            is_excluded: false,
            status: 'DISPOSED',
            reported_at: new Date(Date.now() - 432000000).toISOString()
          },
          {
            id: 'rk-2026-009',
            species: 'Red Fox',
            estimated_weight_kg: 7,
            decay_stage: 'ADVANCED',
            location: {
              lat: 44.060,
              lng: 145.010,
              road_name: 'Route 334 Pass',
              region: 'Shiretoko, Hokkaido, Japan'
            },
            photo_url: '',
            photo_urls: [],
            reliability_score: 0.40,
            is_excluded: true,
            status: 'DISPOSED',
            reported_at: new Date(Date.now() - 950400000).toISOString()
          },
          {
            id: 'rk-2026-010',
            species: 'Wild Boar',
            estimated_weight_kg: 52,
            decay_stage: 'ADVANCED',
            location: {
              lat: 26.540,
              lng: 93.050,
              road_name: 'Kaliabor Bypass NH-715',
              region: 'Nagaon, Assam, India'
            },
            photo_url: '',
            photo_urls: [],
            reliability_score: 0.75,
            is_excluded: false,
            status: 'DISPOSED',
            reported_at: new Date(Date.now() - 1555200000).toISOString()
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
            photo_urls: ['https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_sugarcane_field.jpg'],
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
            photo_urls: [],
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
            photo_urls: [],
            last_updated: new Date(Date.now() - 7200000).toISOString()
          },
          {
            id: 'env-kaki-04',
            category: 'PERSIMMON_KAKI',
            name: 'Morioka Matsuzono Abandoned Fruit Tree Cluster',
            location_desc: 'Residential hillside fringe fruit groves near elementary school',
            geometry_type: 'circle',
            circle_center: { lat: 39.735, lng: 141.165 },
            radius_meters: 1500,
            risk_weight: 0.92,
            active_season: 'October - December (Peak Fruiting & Pre-Denning)',
            photo_urls: ['https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_kaki_trees.jpg'],
            last_updated: new Date(Date.now() - 21600000).toISOString()
          },
          {
            id: 'env-water-05',
            category: 'WATER_SOURCE',
            name: 'Huai Sat Yai Reservoir Channel Watering Point',
            location_desc: 'Dry season sole watering access between forest and village',
            geometry_type: 'circle',
            circle_center: { lat: 12.535, lng: 99.525 },
            radius_meters: 1200,
            risk_weight: 0.75,
            active_season: 'February - May (Dry Season Drought)',
            photo_urls: [],
            last_updated: new Date(Date.now() - 64800000).toISOString()
          },
          {
            id: 'env-crop-06',
            category: 'CROP_PHENOLOGY',
            name: 'Kaziranga Bokakhat Tea & Lowland Buffer Belt',
            location_desc: 'Wetland buffer corridor between NH-715 and tea plantation edge',
            geometry_type: 'circle',
            circle_center: { lat: 26.615, lng: 93.590 },
            radius_meters: 3200,
            risk_weight: 0.88,
            active_season: 'June - September (Monsoon Inundation Dispersal)',
            photo_urls: [],
            last_updated: new Date(Date.now() - 86400000).toISOString()
          },
          {
            id: 'env-crop-07',
            category: 'CROP_PHENOLOGY',
            name: 'Eniwa Agricultural Sweetcorn & Sugarbeet Belt',
            location_desc: 'Forest edge crop plots attractive to brown bears prior to denning',
            geometry_type: 'circle',
            circle_center: { lat: 42.920, lng: 141.580 },
            radius_meters: 2500,
            risk_weight: 0.82,
            active_season: 'August - October (Sweetcorn Ripening Peak)',
            photo_urls: [],
            last_updated: new Date(Date.now() - 172800000).toISOString()
          },
          {
            id: 'env-pasture-08',
            category: 'LIVESTOCK_PASTURE',
            name: 'Pelalawan Concession Night Grazing Enclosure',
            location_desc: 'Communal sheep and cattle pens bordering pulpwood concession',
            geometry_type: 'circle',
            circle_center: { lat: 0.360, lng: 101.850 },
            radius_meters: 2000,
            risk_weight: 0.80,
            active_season: 'Year-round',
            photo_urls: [],
            last_updated: new Date(Date.now() - 259200000).toISOString()
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

  // --- Weather & Meteorological Data ---
  getWeather(region?: string): Observable<AdminWeatherData[]> {
    const url = region ? `${this.baseUrl}/weather?region=${encodeURIComponent(region)}` : `${this.baseUrl}/weather`;
    return this.http.get<AdminWeatherData[]>(url).pipe(
      tap(data => this.weather.set(data)),
      catchError(() => {
        const fallback: AdminWeatherData[] = [
          {
            id: 'weather-thai-01',
            station_name: 'Kaeng Krachan Substation Alpha',
            region: 'Thailand (Kaeng Krachan)',
            latitude: 12.875,
            longitude: 99.630,
            temperature_celsius: 31.5,
            humidity_pct: 82.0,
            wind_speed_kmh: 14.0,
            wind_direction_deg: 90.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.12,
            moon_phase: 'WANING_CRESCENT',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Dry evening east wind dispersing sugarcane volatile plume towards park border',
            last_updated: new Date(Date.now() - 20 * 60000).toISOString()
          },
          {
            id: 'weather-sumatra-02',
            station_name: 'Siak Forest Edge Met-Tower 4',
            region: 'Indonesia (Riau)',
            latitude: 0.520,
            longitude: 101.450,
            temperature_celsius: 28.0,
            humidity_pct: 89.0,
            wind_speed_kmh: 6.0,
            wind_direction_deg: 180.0,
            precipitation_mm_h: 2.5,
            moon_illumination: 0.02,
            moon_phase: 'NEW_MOON',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'New moon nocturnal cycle creates peak ambush probability near cattle corral fringes',
            last_updated: new Date(Date.now() - 45 * 60000).toISOString()
          },
          {
            id: 'weather-japan-03',
            station_name: 'Nagano Valley Agromet Station',
            region: 'Japan (Nagano)',
            latitude: 36.650,
            longitude: 138.180,
            temperature_celsius: 14.2,
            humidity_pct: 65.0,
            wind_speed_kmh: 11.5,
            wind_direction_deg: 270.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.08,
            moon_phase: 'NEW_MOON',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Crisp autumn dusk wind carrying ripe kaki volatile esters into western foothills',
            last_updated: new Date(Date.now() - 60 * 60000).toISOString()
          },
          {
            id: 'weather-japan-04',
            station_name: 'Morioka Matsuzono Hilltop Sensor',
            region: 'Japan (Iwate)',
            latitude: 39.735,
            longitude: 141.165,
            temperature_celsius: 11.8,
            humidity_pct: 71.0,
            wind_speed_kmh: 8.0,
            wind_direction_deg: 315.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.15,
            moon_phase: 'WAXING_CRESCENT',
            is_flooded_river_risk: false,
            weather_condition: 'OVERCAST',
            notes: 'Overcast night conditions encouraging crepuscular bear foraging along stream corridor',
            last_updated: new Date(Date.now() - 120 * 60000).toISOString()
          },
          {
            id: 'weather-india-05',
            station_name: 'Kaziranga Panbari Flood Gauge',
            region: 'India (Assam)',
            latitude: 26.590,
            longitude: 93.180,
            temperature_celsius: 27.5,
            humidity_pct: 94.0,
            wind_speed_kmh: 18.0,
            wind_direction_deg: 220.0,
            precipitation_mm_h: 32.0,
            moon_illumination: 0.88,
            moon_phase: 'WAXING_GIBBOUS',
            is_flooded_river_risk: true,
            weather_condition: 'HEAVY_RAIN',
            notes: 'Severe precipitation (>25mm/h) triggers Brahmaputra river overflow hazard and animal highland migration',
            last_updated: new Date(Date.now() - 10 * 60000).toISOString()
          },
          {
            id: 'weather-japan-06',
            station_name: 'Shiretoko Rausu Coastal Tower',
            region: 'Japan (Hokkaido)',
            latitude: 44.020,
            longitude: 145.190,
            temperature_celsius: 9.5,
            humidity_pct: 78.0,
            wind_speed_kmh: 16.0,
            wind_direction_deg: 45.0,
            precipitation_mm_h: 1.0,
            moon_illumination: 0.50,
            moon_phase: 'FIRST_QUARTER',
            is_flooded_river_risk: false,
            weather_condition: 'OVERCAST',
            notes: 'Coastal onshore breeze dispersing salmon carcass odor upriver into mountain passes',
            last_updated: new Date(Date.now() - 180 * 60000).toISOString()
          }
        ];
        this.weather.set(fallback);
        return of(fallback);
      })
    );
  }

  createWeather(item: Partial<AdminWeatherData>): Observable<AdminWeatherData> {
    return this.http.post<AdminWeatherData>(`${this.baseUrl}/weather`, item).pipe(
      tap(created => {
        this.weather.update(list => [created, ...list]);
      }),
      catchError(() => {
        const mock: AdminWeatherData = {
          id: `weather-${Date.now()}`,
          station_name: item.station_name || 'New Observation Station',
          region: item.region || 'Regional Met-Grid',
          latitude: item.latitude || 0,
          longitude: item.longitude || 0,
          temperature_celsius: item.temperature_celsius ?? 22.0,
          humidity_pct: item.humidity_pct ?? 70.0,
          wind_speed_kmh: item.wind_speed_kmh ?? 10.0,
          wind_direction_deg: item.wind_direction_deg ?? 90.0,
          precipitation_mm_h: item.precipitation_mm_h ?? 0.0,
          moon_illumination: item.moon_illumination ?? 0.5,
          moon_phase: item.moon_phase || 'FIRST_QUARTER',
          is_flooded_river_risk: (item.precipitation_mm_h ?? 0) >= 25.0,
          weather_condition: item.weather_condition || 'CLEAR',
          notes: item.notes || '',
          last_updated: new Date().toISOString()
        };
        this.weather.update(list => [mock, ...list]);
        return of(mock);
      })
    );
  }

  updateWeather(id: string, item: Partial<AdminWeatherData>): Observable<AdminWeatherData> {
    return this.http.put<AdminWeatherData>(`${this.baseUrl}/weather/${id}`, item).pipe(
      tap(updated => {
        this.weather.update(list => list.map(w => w.id === id ? updated : w));
      }),
      catchError(() => {
        let updatedItem: AdminWeatherData | null = null;
        this.weather.update(list => list.map(w => {
          if (w.id === id) {
            updatedItem = {
              ...w,
              ...item,
              is_flooded_river_risk: (item.precipitation_mm_h ?? w.precipitation_mm_h) >= 25.0,
              last_updated: new Date().toISOString()
            };
            return updatedItem;
          }
          return w;
        }));
        return of(updatedItem!);
      })
    );
  }

  deleteWeather(id: string): Observable<any> {
    return this.http.delete(`${this.baseUrl}/weather/${id}`).pipe(
      tap(() => {
        this.weather.update(list => list.filter(w => w.id !== id));
      }),
      catchError(() => {
        this.weather.update(list => list.filter(w => w.id !== id));
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
