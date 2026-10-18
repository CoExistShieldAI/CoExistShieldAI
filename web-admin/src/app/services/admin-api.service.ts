import { Injectable, inject, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, catchError, of, tap, throwError } from 'rxjs';
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
  // Dynamic baseUrl: points to Cloud Run in production, localhost in local development, or configured window.COEXIST_BACKEND_URL
  private baseUrl = typeof window !== 'undefined' && window.location.hostname === 'localhost'
    ? 'http://localhost:8080/api/v1/admin'
    : ((typeof window !== 'undefined' && (window as any).COEXIST_BACKEND_URL)
        ? `${(window as any).COEXIST_BACKEND_URL}/api/v1/admin`
        : 'https://coexistshield-backend-400266593079.asia-southeast1.run.app/api/v1/admin');

  // Reactive state signals
  readonly stats = signal<DashboardStats | null>(null);
  readonly geofences = signal<AdminGeofence[]>([]);
  readonly wildlife = signal<AdminWildlifeIncident[]>([]);
  readonly roadkills = signal<AdminRoadkillRecord[]>([]);
  readonly environmental = signal<AdminEnvironmentalData[]>([]);
  readonly weather = signal<AdminWeatherData[]>([]);
  readonly isRecalculating = signal<boolean>(false);
  readonly isOfflineFallback = signal<boolean>(false);
  readonly isDemoMode = signal<boolean>(typeof window !== 'undefined' && (localStorage.getItem('DEMO_MODE') === 'true' || window.location.search.includes('demo=true')));

  // --- Dashboard ---
  getStats(): Observable<DashboardStats> {
    return this.http.get<DashboardStats>(`${this.baseUrl}/dashboard/stats`).pipe(
      tap(data => {
        this.stats.set(data);
        this.isOfflineFallback.set(false);
      }),
      catchError(() => {
        this.isOfflineFallback.set(true);
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
            'Asian Elephant': 5,
            'Asian Black Bear': 2,
            'Brown Bear': 2,
            'Sumatran Tiger': 2,
            'Bengal Tiger': 1,
            'Sloth Bear': 1,
            'Indian Leopard': 1,
            'Saltwater Crocodile': 1
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
      tap(data => {
        this.geofences.set(data);
        this.isOfflineFallback.set(false);
      }),
      catchError(() => {
        this.isOfflineFallback.set(true);
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
            id: 'geo-ai-srilanka-06',
            name: 'Minneriya-Polonnaruwa Agricultural & Rail Corridor (HEC Hotspot)',
            source: 'ai',
            is_persistent: false,
            geometry_type: 'circle',
            circle_center: { lat: 8.045, lng: 80.932 },
            circle_radius_meters: 4200,
            threat_level: 'TIER_3_CRITICAL',
            target_species: 'Asian Elephant',
            created_at: new Date(Date.now() - 18000000).toISOString(),
            reason: 'Maha paddy harvest maturity coupling with elephant herd migration across Batticaloa railway line; extreme night clash risk'
          },
          {
            id: 'geo-ai-srilanka-07',
            name: 'Udawalawe-Sevenagala Sugar Basin Defense Buffer',
            source: 'ai',
            is_persistent: false,
            geometry_type: 'circle',
            circle_center: { lat: 6.418, lng: 80.942 },
            circle_radius_meters: 3500,
            threat_level: 'TIER_2_ALERT',
            target_species: 'Asian Elephant',
            created_at: new Date(Date.now() - 25200000).toISOString(),
            reason: 'High-sucrose volatile spikes from commercial sugarcane fields drawing breeding elephant herds across southern buffer'
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
      })
    );
  }

  deleteGeofence(id: string): Observable<any> {
    return this.http.delete(`${this.baseUrl}/geofences/${id}`).pipe(
      tap(() => {
        this.geofences.update(list => list.filter(g => g.id !== id));
      })
    );
  }

  clearAllManualGeofences(): Observable<any> {
    return this.http.delete(`${this.baseUrl}/geofences/manual/all`).pipe(
      tap(() => {
        this.geofences.update(list => list.filter(g => g.source !== 'manual'));
      })
    );
  }

  recalculateAiGeofences(): Observable<any> {
    if (this.isRecalculating()) {
      return of({ status: 'in_progress', message: 'Recalculation already running' });
    }
    this.isRecalculating.set(true);
    return this.http.post(`${this.baseUrl}/geofences/recalculate`, {}).pipe(
      tap(() => {
        this.isRecalculating.set(false);
        this.getGeofences().subscribe();
      }),
      catchError((err) => {
        this.isRecalculating.set(false);
        if (err?.status === 409) {
          return throwError(() => err);
        }
        // Fallback for offline development: simulate correlating incidents with attractants
        this.simulateLocalAiRecalculation();
        return of({ status: 'recalculated' });
      })
    );
  }

  private simulateLocalAiRecalculation() {
    const currentFences = [...this.geofences()];
    const wild = this.wildlife();
    const envs = this.environmental();

    for (const w of wild) {
      const isBear = (w.species || '').toLowerCase().includes('bear') || (w.species || '').includes('クマ');
      if (!isBear || !w.location?.lat || !w.location?.lng) continue;

      for (const e of envs) {
        const isKaki = e.category === 'PERSIMMON_KAKI' ||
          (e.name || '').toLowerCase().includes('kaki') ||
          (e.name || '').toLowerCase().includes('persimmon') ||
          (e.name || '').includes('柿');
        const envLat = e.circle_center?.lat || (e.coordinates && e.coordinates[0]?.lat);
        const envLng = e.circle_center?.lng || (e.coordinates && e.coordinates[0]?.lng);
        if (!isKaki || !envLat || !envLng) continue;

        const dLat = (w.location.lat - envLat);
        const dLng = (w.location.lng - envLng);
        const distKm = Math.sqrt(dLat * dLat + dLng * dLng) * 111.0;

        if (distKm <= 10.0) {
          const geoId = `geo-ai-corr-${w.id}-${e.id}`;
          if (!currentFences.some(f => f.id === geoId)) {
            const regionLabel = w.location.region || w.location.road_name || 'Nara';
            currentFences.unshift({
              id: geoId,
              name: `AI Dynamic Shield: ${regionLabel} Bear-Persimmon Hazard Zone`,
              source: 'ai',
              is_persistent: false,
              geometry_type: 'circle',
              circle_center: {
                lat: (w.location.lat + envLat) / 2.0,
                lng: (w.location.lng + envLng) / 2.0
              },
              circle_radius_meters: Math.min(4500, Math.max(1500, (distKm * 1000) / 2 + 1500)),
              threat_level: 'TIER_3_CRITICAL',
              target_species: w.species,
              created_at: new Date().toISOString(),
              reason: `AI Spatial Correlation: ${w.species} sighted within ${distKm.toFixed(1)}km of unharvested persimmon (kaki) attractant '${e.name}'. Hyperphagia incursion alert.`
            });
          }
        }
      }
    }
    this.geofences.set(currentFences);
  }

  // --- Wildlife Incidents ---
  getWildlife(): Observable<AdminWildlifeIncident[]> {
    return this.http.get<AdminWildlifeIncident[]>(`${this.baseUrl}/incidents/wildlife`).pipe(
      tap(data => {
        this.wildlife.set(data);
        this.isOfflineFallback.set(false);
      }),
      catchError(() => {
        this.isOfflineFallback.set(true);
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
          },
          {
            id: 'wildlife-011',
            species: 'Asian Elephant',
            location: {
              lat: 8.035,
              lng: 80.915,
              road_name: 'Habarana-Polonnaruwa Highway A11',
              region: 'Minneriya / Polonnaruwa, Sri Lanka'
            },
            threat_level: 'TIER_3_CRITICAL',
            status: 'VERIFIED',
            reporter: 'Sri Lanka DWC Elephant Response Unit',
            description: 'Matriarch herd of 11 Sri Lankan elephants crossing railway line and entering ripening paddy fields.',
            photo_urls: ['https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_srilanka_elephant.jpg'],
            reported_at: new Date(Date.now() - 7200000).toISOString()
          },
          {
            id: 'wildlife-012',
            species: 'Sloth Bear',
            location: {
              lat: 8.355,
              lng: 80.205,
              road_name: 'Wilpattu Buffer Access Track',
              region: 'Anuradhapura / Wilpattu, Sri Lanka'
            },
            threat_level: 'TIER_2_ALERT',
            status: 'VERIFIED',
            reporter: 'Community Forest Patrol',
            description: 'Adult sloth bear startled near termite mound along firewood collection path; defensive charge repelled without injury.',
            photo_urls: [],
            reported_at: new Date(Date.now() - 14400000).toISOString()
          },
          {
            id: 'wildlife-013',
            species: 'Saltwater Crocodile',
            location: {
              lat: -2.125,
              lng: 106.115,
              road_name: 'Batu Belubang Estuary Creek',
              region: 'Bangka Belitung, Indonesia'
            },
            threat_level: 'TIER_3_CRITICAL',
            status: 'VERIFIED',
            reporter: 'Marine & Fisheries Coastal Guard',
            description: '4.5m estuarine crocodile sighted stalking near community artisanal fishing pier and abandoned tin-mine pool.',
            photo_urls: [],
            reported_at: new Date(Date.now() - 25200000).toISOString()
          },
          {
            id: 'wildlife-014',
            species: 'Indian Leopard',
            location: {
              lat: 10.325,
              lng: 76.955,
              road_name: 'Waterfall Estate Sector 3',
              region: 'Valparai Plateau, Tamil Nadu, India'
            },
            threat_level: 'TIER_2_ALERT',
            status: 'VERIFIED',
            reporter: 'Nature Conservation Foundation Field Team',
            description: 'Adult leopard daytime resting in tea bushes within 50m of tea estate worker labour residential quarters.',
            photo_urls: [],
            reported_at: new Date(Date.now() - 39600000).toISOString()
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
      })
    );
  }

  deleteWildlife(id: string): Observable<any> {
    return this.http.delete(`${this.baseUrl}/incidents/wildlife/${id}`).pipe(
      tap(() => {
        this.wildlife.update(list => list.filter(w => w.id !== id));
      })
    );
  }

  updateWildlifeStatus(id: string, status: string): Observable<any> {
    return this.http.patch(`${this.baseUrl}/incidents/wildlife/${id}/verify`, { status }).pipe(
      tap(() => {
        this.wildlife.update(list =>
          list.map(w => w.id === id ? { ...w, status: status as any } : w)
        );
      })
    );
  }

  // --- Roadkill Records ---
  getRoadkills(includeExcluded: boolean = true): Observable<AdminRoadkillRecord[]> {
    return this.http.get<AdminRoadkillRecord[]>(`${this.baseUrl}/incidents/roadkill?include_excluded=${includeExcluded}`).pipe(
      tap(data => {
        this.roadkills.set(data);
        this.isOfflineFallback.set(false);
      }),
      catchError(() => {
        this.isOfflineFallback.set(true);
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
      })
    );
  }

  toggleRoadkillExclusion(id: string, isExcluded: boolean): Observable<any> {
    return this.http.patch(`${this.baseUrl}/incidents/roadkill/${id}/exclude`, { is_excluded: isExcluded }).pipe(
      tap(() => {
        this.roadkills.update(list =>
          list.map(rk => rk.id === id ? { ...rk, is_excluded: isExcluded } : rk)
        );
      })
    );
  }

  markRoadkillDisposed(id: string): Observable<any> {
    return this.http.patch(`${this.baseUrl}/incidents/roadkill/${id}/dispose`, {}).pipe(
      tap(() => {
        this.roadkills.update(list =>
          list.map(rk => rk.id === id ? { ...rk, status: 'DISPOSED' } : rk)
        );
      })
    );
  }

  // --- Environmental ---
  getEnvironmental(): Observable<AdminEnvironmentalData[]> {
    return this.http.get<AdminEnvironmentalData[]>(`${this.baseUrl}/environmental`).pipe(
      tap(data => {
        this.environmental.set(data);
        this.isOfflineFallback.set(false);
      }),
      catchError(() => {
        this.isOfflineFallback.set(true);
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
          },
          {
            id: 'env-crop-09',
            category: 'CROP_PHENOLOGY',
            name: 'Minneriya-Polonnaruwa Maha Paddy Agro-Belt',
            location_desc: 'Irrigated paddy fields along Batticaloa railway line and Habarana corridor',
            geometry_type: 'circle',
            circle_center: { lat: 8.045, lng: 80.932 },
            radius_meters: 4500,
            risk_weight: 0.94,
            active_season: 'December - March (Maha Harvest Peak Ripening)',
            photo_urls: ['https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_srilanka_paddy.jpg'],
            last_updated: new Date(Date.now() - 3600000).toISOString()
          },
          {
            id: 'env-crop-10',
            category: 'CROP_PHENOLOGY',
            name: 'Udawalawe Sevenagala Sugarcane Plantation',
            location_desc: 'Commercial sugarcane and banana cultivation adjacent to national park boundary',
            geometry_type: 'circle',
            circle_center: { lat: 6.418, lng: 80.942 },
            radius_meters: 3800,
            risk_weight: 0.89,
            active_season: 'Year-round (Intensive Harvesting Cycles)',
            photo_urls: ['https://storage.googleapis.com/corald-coexistshield-media/uploads/mock_sugarcane_field.jpg'],
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
      })
    );
  }

  deleteEnvironmental(id: string): Observable<any> {
    return this.http.delete(`${this.baseUrl}/environmental/${id}`).pipe(
      tap(() => {
        this.environmental.update(list => list.filter(e => e.id !== id));
      })
    );
  }

  // --- Weather & Meteorological Data ---
  getWeather(region?: string): Observable<AdminWeatherData[]> {
    const url = region ? `${this.baseUrl}/weather?region=${encodeURIComponent(region)}` : `${this.baseUrl}/weather`;
    return this.http.get<AdminWeatherData[]>(url).pipe(
      tap(data => {
        this.weather.set(data);
        this.isOfflineFallback.set(false);
      }),
      catchError(() => {
        this.isOfflineFallback.set(true);
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
            region: 'Japan (Hokkaido / Shiretoko)',
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
          },
          {
            id: 'weather-thai-02',
            station_name: 'Kui Buri Agro-Forest Met Tower',
            region: 'Thailand (Kui Buri / Prachuap)',
            latitude: 12.062,
            longitude: 99.782,
            temperature_celsius: 30.2,
            humidity_pct: 85.0,
            wind_speed_kmh: 11.0,
            wind_direction_deg: 80.0,
            precipitation_mm_h: 1.5,
            moon_illumination: 0.12,
            moon_phase: 'WANING_CRESCENT',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'Pineapple & jackfruit buffer; high-sugar crop volatility triggering breeding herd incursions',
            last_updated: new Date(Date.now() - 35 * 60000).toISOString()
          },
          {
            id: 'weather-thai-03',
            station_name: 'Khao Yai Dong Phayayen Gateway',
            region: 'Thailand (Khao Yai / Nakhon Ratchasima)',
            latitude: 14.439,
            longitude: 101.372,
            temperature_celsius: 27.8,
            humidity_pct: 76.0,
            wind_speed_kmh: 9.5,
            wind_direction_deg: 110.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.12,
            moon_phase: 'WANING_CRESCENT',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Highway 3077 tourist road corridor; bull elephant charging & roadblock incidents',
            last_updated: new Date(Date.now() - 50 * 60000).toISOString()
          },
          {
            id: 'weather-thai-04',
            station_name: 'Soi Dao Chanthaburi Fruit Station',
            region: 'Thailand (Chanthaburi / Soi Dao)',
            latitude: 12.980,
            longitude: 102.195,
            temperature_celsius: 29.0,
            humidity_pct: 88.0,
            wind_speed_kmh: 7.0,
            wind_direction_deg: 135.0,
            precipitation_mm_h: 4.0,
            moon_illumination: 0.12,
            moon_phase: 'WANING_CRESCENT',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'Durian & oil palm plantations; heavy elephant herd feeding with electric fence breach risks',
            last_updated: new Date(Date.now() - 65 * 60000).toISOString()
          },
          {
            id: 'weather-sumatra-06',
            station_name: 'Pelalawan Concession Outpost',
            region: 'Indonesia (Riau / Pelalawan)',
            latitude: 0.360,
            longitude: 101.850,
            temperature_celsius: 27.5,
            humidity_pct: 91.0,
            wind_speed_kmh: 5.5,
            wind_direction_deg: 190.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.02,
            moon_phase: 'NEW_MOON',
            is_flooded_river_risk: false,
            weather_condition: 'OVERCAST',
            notes: 'Pulpwood acacia plantation corridor; solitary tiger encounter site near communal pens',
            last_updated: new Date(Date.now() - 80 * 60000).toISOString()
          },
          {
            id: 'weather-sumatra-07',
            station_name: 'Bukit Tigapuluh Buffer Sensor',
            region: 'Indonesia (Jambi / Tebo)',
            latitude: -1.250,
            longitude: 102.580,
            temperature_celsius: 26.8,
            humidity_pct: 93.0,
            wind_speed_kmh: 4.0,
            wind_direction_deg: 210.0,
            precipitation_mm_h: 3.5,
            moon_illumination: 0.02,
            moon_phase: 'NEW_MOON',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'Rubber estate buffer; elephant crop raiding and tiger poaching wire snare conflicts',
            last_updated: new Date(Date.now() - 90 * 60000).toISOString()
          },
          {
            id: 'weather-sumatra-08',
            station_name: 'Leuser Ecosystem Aceh Border',
            region: 'Indonesia (Aceh / Subulussalam)',
            latitude: 2.650,
            longitude: 97.980,
            temperature_celsius: 25.5,
            humidity_pct: 90.0,
            wind_speed_kmh: 8.0,
            wind_direction_deg: 270.0,
            precipitation_mm_h: 6.0,
            moon_illumination: 0.02,
            moon_phase: 'NEW_MOON',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'Fragmented forest neck; elephant migration pathway intersecting village settlement',
            last_updated: new Date(Date.now() - 110 * 60000).toISOString()
          },
          {
            id: 'weather-india-10',
            station_name: 'Valparai Anamalai Tea Station',
            region: 'India (Tamil Nadu / Valparai)',
            latitude: 10.325,
            longitude: 76.955,
            temperature_celsius: 22.0,
            humidity_pct: 88.0,
            wind_speed_kmh: 12.0,
            wind_direction_deg: 250.0,
            precipitation_mm_h: 12.0,
            moon_illumination: 0.88,
            moon_phase: 'WAXING_GIBBOUS',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'Tea plantation plateau; resident elephant herd routes passing through worker labour lines',
            last_updated: new Date(Date.now() - 40 * 60000).toISOString()
          },
          {
            id: 'weather-india-11',
            station_name: 'Bandipur Wayanad NH-766 Sensor',
            region: 'India (Karnataka / Bandipur)',
            latitude: 11.665,
            longitude: 76.635,
            temperature_celsius: 24.0,
            humidity_pct: 75.0,
            wind_speed_kmh: 10.0,
            wind_direction_deg: 260.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.88,
            moon_phase: 'WAXING_GIBBOUS',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Interstate highway NH-766 night traffic corridor; high frequency deer & predator roadkill',
            last_updated: new Date(Date.now() - 60 * 60000).toISOString()
          },
          {
            id: 'weather-india-12',
            station_name: 'Sundarbans Coastal Met Post',
            region: 'India (West Bengal / Sundarbans)',
            latitude: 21.949,
            longitude: 88.899,
            temperature_celsius: 29.5,
            humidity_pct: 92.0,
            wind_speed_kmh: 15.0,
            wind_direction_deg: 160.0,
            precipitation_mm_h: 5.0,
            moon_illumination: 0.88,
            moon_phase: 'WAXING_GIBBOUS',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'Mangrove tidal channel; fatal tiger ambush risk on honey collectors and crab fishermen',
            last_updated: new Date(Date.now() - 75 * 60000).toISOString()
          },
          {
            id: 'weather-japan-15',
            station_name: 'Kazuno Hachimantai Bear Station',
            region: 'Japan (Akita / Kazuno)',
            latitude: 40.215,
            longitude: 140.785,
            temperature_celsius: 10.5,
            humidity_pct: 74.0,
            wind_speed_kmh: 9.0,
            wind_direction_deg: 300.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.15,
            moon_phase: 'WAXING_CRESCENT',
            is_flooded_river_risk: false,
            weather_condition: 'OVERCAST',
            notes: 'Beech mast failure epicentre; bamboo shoot foragers fatal attack site in mountain passes',
            last_updated: new Date(Date.now() - 120 * 60000).toISOString()
          },
          {
            id: 'weather-japan-16',
            station_name: 'Aizu Basin Foothill Sensor',
            region: 'Japan (Fukushima / Aizu)',
            latitude: 37.520,
            longitude: 139.930,
            temperature_celsius: 13.0,
            humidity_pct: 68.0,
            wind_speed_kmh: 7.5,
            wind_direction_deg: 280.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.08,
            moon_phase: 'NEW_MOON',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Onsen resort perimeter & soba fields; bear foraging along tourist hiking access routes',
            last_updated: new Date(Date.now() - 180 * 60000).toISOString()
          },
          {
            id: 'weather-japan-18',
            station_name: 'Sapporo Minami Hillside Sensor',
            region: 'Japan (Hokkaido / Sapporo)',
            latitude: 43.005,
            longitude: 141.310,
            temperature_celsius: 12.0,
            humidity_pct: 66.0,
            wind_speed_kmh: 8.5,
            wind_direction_deg: 330.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.50,
            moon_phase: 'FIRST_QUARTER',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Suburban foothills & school route; brown bear sightings near Toyohira river corridor',
            last_updated: new Date(Date.now() - 240 * 60000).toISOString()
          },
          {
            id: 'weather-japan-19',
            station_name: 'Eniwa Agriculture Sweetcorn Belt',
            region: 'Japan (Hokkaido / Eniwa)',
            latitude: 42.920,
            longitude: 141.580,
            temperature_celsius: 11.2,
            humidity_pct: 70.0,
            wind_speed_kmh: 10.0,
            wind_direction_deg: 340.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.50,
            moon_phase: 'FIRST_QUARTER',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'High-calorie sweetcorn & sugarbeet fields; massive pre-denning crop damage zones',
            last_updated: new Date(Date.now() - 300 * 60000).toISOString()
          },
          {
            id: 'weather-japan-20',
            station_name: 'Shumarinai Lake Basin Tower',
            region: 'Japan (Hokkaido / Horokanai)',
            latitude: 44.290,
            longitude: 142.170,
            temperature_celsius: 7.8,
            humidity_pct: 82.0,
            wind_speed_kmh: 12.0,
            wind_direction_deg: 10.0,
            precipitation_mm_h: 0.5,
            moon_illumination: 0.50,
            moon_phase: 'FIRST_QUARTER',
            is_flooded_river_risk: false,
            weather_condition: 'OVERCAST',
            notes: 'Wild lake peninsula; fatal angler predatory attack site & remote forestry frontier',
            last_updated: new Date(Date.now() - 300 * 60000).toISOString()
          },
          // --- ADJACENT & SIMILAR CONDITIONS EXPANSION (21 - 40) ---
          {
            id: 'weather-thai-21',
            station_name: 'Thap Lan Hwy 304 Wildlife Corridor',
            region: 'Thailand (Prachinburi / Thap Lan)',
            latitude: 14.215,
            longitude: 101.930,
            temperature_celsius: 28.5,
            humidity_pct: 79.0,
            wind_speed_kmh: 10.5,
            wind_direction_deg: 120.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.12,
            moon_phase: 'WANING_CRESCENT',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Adjacent to Khao Yai; Highway 304 wildlife ecoduct bridge corridor with frequent gaur & elephant vehicular hazard',
            last_updated: new Date(Date.now() - 30 * 60000).toISOString()
          },
          {
            id: 'weather-thai-22',
            station_name: 'Khao Ang Rue Nai Tha Takiap Outpost',
            region: 'Thailand (Chachoengsao / Tha Takiap)',
            latitude: 13.385,
            longitude: 101.950,
            temperature_celsius: 29.8,
            humidity_pct: 84.0,
            wind_speed_kmh: 8.5,
            wind_direction_deg: 140.0,
            precipitation_mm_h: 2.0,
            moon_illumination: 0.12,
            moon_phase: 'WANING_CRESCENT',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'Adjacent to Chanthaburi; Eastern Forest Complex core with severe elephant crop raiding across cassava & oil palm estates',
            last_updated: new Date(Date.now() - 40 * 60000).toISOString()
          },
          {
            id: 'weather-thai-23',
            station_name: 'Huai Yang Coastal Mountain Pass',
            region: 'Thailand (Prachuap / Thap Sakae)',
            latitude: 11.605,
            longitude: 99.660,
            temperature_celsius: 30.5,
            humidity_pct: 81.0,
            wind_speed_kmh: 13.0,
            wind_direction_deg: 95.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.12,
            moon_phase: 'WANING_CRESCENT',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Adjacent south of Kui Buri; narrow Isthmus corridor between Tenasserim mountain ridge and coconut coastal farms',
            last_updated: new Date(Date.now() - 45 * 60000).toISOString()
          },
          {
            id: 'weather-thai-24',
            station_name: 'Suan Phueng Phachi River Frontier',
            region: 'Thailand (Ratchaburi / Suan Phueng)',
            latitude: 13.545,
            longitude: 99.310,
            temperature_celsius: 27.2,
            humidity_pct: 83.0,
            wind_speed_kmh: 9.0,
            wind_direction_deg: 100.0,
            precipitation_mm_h: 1.0,
            moon_illumination: 0.12,
            moon_phase: 'WANING_CRESCENT',
            is_flooded_river_risk: false,
            weather_condition: 'OVERCAST',
            notes: 'Adjacent north of Kaeng Krachan; Myanmar border montane forest; bamboo shoot & pineapple raiding herds',
            last_updated: new Date(Date.now() - 55 * 60000).toISOString()
          },
          {
            id: 'weather-sumatra-25',
            station_name: 'Tesso Nilo North Buffer Station',
            region: 'Indonesia (Riau / Tesso Nilo)',
            latitude: 0.120,
            longitude: 101.620,
            temperature_celsius: 28.6,
            humidity_pct: 92.0,
            wind_speed_kmh: 5.0,
            wind_direction_deg: 175.0,
            precipitation_mm_h: 3.0,
            moon_illumination: 0.02,
            moon_phase: 'NEW_MOON',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'Adjacent to Pelalawan; fragmented lowland rain forest and encroached palm landscape with frequent elephant troop encounters',
            last_updated: new Date(Date.now() - 50 * 60000).toISOString()
          },
          {
            id: 'weather-sumatra-26',
            station_name: 'Giam Siak Peatland Bio-Observatory',
            region: 'Indonesia (Riau / Bengkalis)',
            latitude: 1.180,
            longitude: 101.880,
            temperature_celsius: 27.8,
            humidity_pct: 94.0,
            wind_speed_kmh: 6.5,
            wind_direction_deg: 165.0,
            precipitation_mm_h: 1.5,
            moon_illumination: 0.02,
            moon_phase: 'NEW_MOON',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'Adjacent north of Siak; peat swamp forest biosphere boundary; tiger livestock depredation around drainage canals',
            last_updated: new Date(Date.now() - 60 * 60000).toISOString()
          },
          {
            id: 'weather-sumatra-27',
            station_name: 'Langkat Leuser Foothill Monitor',
            region: 'Indonesia (North Sumatra / Langkat)',
            latitude: 3.580,
            longitude: 98.220,
            temperature_celsius: 26.2,
            humidity_pct: 88.0,
            wind_speed_kmh: 7.0,
            wind_direction_deg: 260.0,
            precipitation_mm_h: 4.5,
            moon_illumination: 0.02,
            moon_phase: 'NEW_MOON',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'Adjacent east slope of Gunung Leuser; palm concession border; orangutan and elephant trans-boundary herd routes',
            last_updated: new Date(Date.now() - 70 * 60000).toISOString()
          },
          {
            id: 'weather-sumatra-28',
            station_name: 'Peureulak River Conflict Zone',
            region: 'Indonesia (Aceh / Aceh Timur)',
            latitude: 4.850,
            longitude: 97.880,
            temperature_celsius: 26.8,
            humidity_pct: 87.0,
            wind_speed_kmh: 8.0,
            wind_direction_deg: 280.0,
            precipitation_mm_h: 5.0,
            moon_illumination: 0.02,
            moon_phase: 'NEW_MOON',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'Connected to Leuser Aceh corridor; highway & settlement boundary; frequent high-stress elephant encounters',
            last_updated: new Date(Date.now() - 80 * 60000).toISOString()
          },
          {
            id: 'weather-india-29',
            station_name: 'Karbi Anglong South Foothill Ridge',
            region: 'India (Assam / Karbi Anglong)',
            latitude: 26.510,
            longitude: 93.220,
            temperature_celsius: 26.5,
            humidity_pct: 91.0,
            wind_speed_kmh: 14.0,
            wind_direction_deg: 210.0,
            precipitation_mm_h: 26.0,
            moon_illumination: 0.88,
            moon_phase: 'WAXING_GIBBOUS',
            is_flooded_river_risk: true,
            weather_condition: 'HEAVY_RAIN',
            notes: 'Directly adjacent south of Kaziranga NH-715; critical animal escape corridor when Brahmaputra submerges national park grasslands',
            last_updated: new Date(Date.now() - 30 * 60000).toISOString()
          },
          {
            id: 'weather-india-30',
            station_name: 'Nagarhole Kabini Reservoir Edge',
            region: 'India (Karnataka / Nagarhole)',
            latitude: 11.950,
            longitude: 76.280,
            temperature_celsius: 25.0,
            humidity_pct: 78.0,
            wind_speed_kmh: 11.0,
            wind_direction_deg: 255.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.88,
            moon_phase: 'WAXING_GIBBOUS',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Adjacent west of Bandipur; high-density tiger & elephant concentration along Kabini backwaters and agricultural fringe',
            last_updated: new Date(Date.now() - 40 * 60000).toISOString()
          },
          {
            id: 'weather-india-31',
            station_name: 'Wayanad Sultan Bathery Agro-Forest',
            region: 'India (Kerala / Wayanad)',
            latitude: 11.660,
            longitude: 76.260,
            temperature_celsius: 23.5,
            humidity_pct: 86.0,
            wind_speed_kmh: 9.5,
            wind_direction_deg: 240.0,
            precipitation_mm_h: 8.0,
            moon_illumination: 0.88,
            moon_phase: 'WAXING_GIBBOUS',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'Adjacent south of Bandipur on Kerala side; coffee and pepper plantation matrix with persistent night elephant incursions',
            last_updated: new Date(Date.now() - 50 * 60000).toISOString()
          },
          {
            id: 'weather-india-32',
            station_name: 'Dooars Gorumara Tea Plains',
            region: 'India (West Bengal / Dooars)',
            latitude: 26.750,
            longitude: 88.790,
            temperature_celsius: 25.8,
            humidity_pct: 89.0,
            wind_speed_kmh: 12.0,
            wind_direction_deg: 150.0,
            precipitation_mm_h: 10.0,
            moon_illumination: 0.88,
            moon_phase: 'WAXING_GIBBOUS',
            is_flooded_river_risk: false,
            weather_condition: 'RAIN',
            notes: 'Similar tea-estate landscape to Valparai; railway corridor elephant collision hotspot and leopard day-bedding in tea bushes',
            last_updated: new Date(Date.now() - 60 * 60000).toISOString()
          },
          {
            id: 'weather-japan-33',
            station_name: 'Karuizawa Asama Foothill Sensor',
            region: 'Japan (Nagano / Karuizawa)',
            latitude: 36.350,
            longitude: 138.600,
            temperature_celsius: 12.5,
            humidity_pct: 69.0,
            wind_speed_kmh: 10.0,
            wind_direction_deg: 290.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.08,
            moon_phase: 'NEW_MOON',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Adjacent east of Nagano Chikuma; Mt Asama resort villas and urban perimeter; bear garbage and garden fruit attractants',
            last_updated: new Date(Date.now() - 60 * 60000).toISOString()
          },
          {
            id: 'weather-japan-34',
            station_name: 'Minamiuonuma Valley Base Station',
            region: 'Japan (Niigata / Minamiuonuma)',
            latitude: 37.060,
            longitude: 138.870,
            temperature_celsius: 13.8,
            humidity_pct: 73.0,
            wind_speed_kmh: 8.0,
            wind_direction_deg: 310.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.08,
            moon_phase: 'NEW_MOON',
            is_flooded_river_risk: false,
            weather_condition: 'OVERCAST',
            notes: 'Adjacent northeast of Nagano; heavy snowfall mountain basin; Kan-Etsu corridor and settlement fruit trees attraction',
            last_updated: new Date(Date.now() - 60 * 60000).toISOString()
          },
          {
            id: 'weather-japan-35',
            station_name: 'Hanamaki Kitakami River Corridor',
            region: 'Japan (Iwate / Hanamaki)',
            latitude: 39.390,
            longitude: 141.120,
            temperature_celsius: 12.2,
            humidity_pct: 69.0,
            wind_speed_kmh: 7.5,
            wind_direction_deg: 320.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.15,
            moon_phase: 'WAXING_CRESCENT',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Adjacent south of Morioka; Kitakami river riparian greenbelt providing direct bear travel corridor into suburban apple orchards',
            last_updated: new Date(Date.now() - 120 * 60000).toISOString()
          },
          {
            id: 'weather-japan-36',
            station_name: 'Sendai Aobayama Urban Forest Rim',
            region: 'Japan (Miyagi / Sendai)',
            latitude: 38.260,
            longitude: 140.840,
            temperature_celsius: 14.5,
            humidity_pct: 65.0,
            wind_speed_kmh: 9.0,
            wind_direction_deg: 290.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.15,
            moon_phase: 'WAXING_CRESCENT',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Similar large metropolitan hill margin; university campus and residential hilltops adjacent to Hirose River gorge',
            last_updated: new Date(Date.now() - 120 * 60000).toISOString()
          },
          {
            id: 'weather-japan-37',
            station_name: 'Kitaakita Moriyoshi Mountain Gate',
            region: 'Japan (Akita / Kitaakita)',
            latitude: 40.180,
            longitude: 140.370,
            temperature_celsius: 11.0,
            humidity_pct: 76.0,
            wind_speed_kmh: 8.5,
            wind_direction_deg: 295.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.15,
            moon_phase: 'WAXING_CRESCENT',
            is_flooded_river_risk: false,
            weather_condition: 'OVERCAST',
            notes: 'Adjacent west of Kazuno; Mt Moriyoshi beech zone; consecutive town centre intrusions and elderly pedestrian attacks',
            last_updated: new Date(Date.now() - 120 * 60000).toISOString()
          },
          {
            id: 'weather-japan-38',
            station_name: 'Yonezawa Foothill Orchard Tower',
            region: 'Japan (Yamagata / Yonezawa)',
            latitude: 37.910,
            longitude: 140.110,
            temperature_celsius: 13.5,
            humidity_pct: 67.0,
            wind_speed_kmh: 7.0,
            wind_direction_deg: 270.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.08,
            moon_phase: 'NEW_MOON',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Adjacent north of Aizu Basin; Ou mountain western foothills; La France pear & apple orchards vulnerable to hyperphagia foraging',
            last_updated: new Date(Date.now() - 180 * 60000).toISOString()
          },
          {
            id: 'weather-japan-39',
            station_name: 'Shari Utoro Coastal Salmon Stream',
            region: 'Japan (Hokkaido / Shari Utoro)',
            latitude: 44.070,
            longitude: 144.990,
            temperature_celsius: 10.0,
            humidity_pct: 75.0,
            wind_speed_kmh: 15.0,
            wind_direction_deg: 350.0,
            precipitation_mm_h: 0.5,
            moon_illumination: 0.50,
            moon_phase: 'FIRST_QUARTER',
            is_flooded_river_risk: false,
            weather_condition: 'OVERCAST',
            notes: 'Adjacent west coast of Shiretoko peninsula across pass from Rausu; salmon fishery and Oshinkoshin tourist trail encounters',
            last_updated: new Date(Date.now() - 180 * 60000).toISOString()
          },
          {
            id: 'weather-japan-40',
            station_name: 'Otaru Zenibako Seaside Hillside',
            region: 'Japan (Hokkaido / Otaru)',
            latitude: 43.140,
            longitude: 141.170,
            temperature_celsius: 11.5,
            humidity_pct: 68.0,
            wind_speed_kmh: 11.0,
            wind_direction_deg: 320.0,
            precipitation_mm_h: 0.0,
            moon_illumination: 0.50,
            moon_phase: 'FIRST_QUARTER',
            is_flooded_river_risk: false,
            weather_condition: 'CLEAR',
            notes: 'Adjacent west of Sapporo Teine/Minami mountains; steep coastal bluff residences with fruit trees and garbage foraging',
            last_updated: new Date(Date.now() - 240 * 60000).toISOString()
          }
        ];
        this.weather.set(fallback);
        return of(fallback);
      })
    );
  }

  syncLiveWeather(): Observable<{ status: string; updated_count: number; error_count: number; synced_at: string; source: string }> {
    return this.http.post<{ status: string; updated_count: number; error_count: number; synced_at: string; source: string }>(
      `${this.baseUrl}/weather/sync-live`,
      {}
    ).pipe(
      tap(() => {
        // Refresh local weather signals after successful live sync
        this.getWeather().subscribe();
      }),
      catchError(() => {
        // Fallback simulated sync: update timestamps and randomize wind slightly
        this.weather.update(list => list.map(item => ({
          ...item,
          wind_speed_kmh: +(item.wind_speed_kmh + (Math.random() * 2 - 1)).toFixed(1),
          temperature_celsius: +(item.temperature_celsius + (Math.random() * 0.8 - 0.4)).toFixed(1),
          last_updated: new Date().toISOString()
        })));
        return of({
          status: 'synced',
          updated_count: this.weather().length,
          error_count: 0,
          synced_at: new Date().toISOString(),
          source: 'Open-Meteo REST API (Simulated Fallback)'
        });
      })
    );
  }

  createWeather(item: Partial<AdminWeatherData>): Observable<AdminWeatherData> {
    return this.http.post<AdminWeatherData>(`${this.baseUrl}/weather`, item).pipe(
      tap(created => {
        this.weather.update(list => [created, ...list]);
      })
    );
  }

  updateWeather(id: string, item: Partial<AdminWeatherData>): Observable<AdminWeatherData> {
    return this.http.put<AdminWeatherData>(`${this.baseUrl}/weather/${id}`, item).pipe(
      tap(updated => {
        this.weather.update(list => list.map(w => w.id === id ? updated : w));
      })
    );
  }

  deleteWeather(id: string): Observable<any> {
    return this.http.delete(`${this.baseUrl}/weather/${id}`).pipe(
      tap(() => {
        this.weather.update(list => list.filter(w => w.id !== id));
      })
    );
  }

  // --- Operations ---
  sendEmergencyBroadcast(req: EmergencyBroadcastRequest): Observable<EmergencyBroadcastResponse> {
    return this.http.post<EmergencyBroadcastResponse>(`${this.baseUrl}/operations/broadcast`, req);
  }

  // --- Media Upload ---
  uploadImages(files: File[]): Observable<UploadResponse> {
    const formData = new FormData();
    files.forEach(f => formData.append('files', f, f.name));
    return this.http.post<UploadResponse>(`${this.baseUrl}/upload`, formData);
  }

  getExportUrl(dataset: string, format: 'geojson' | 'csv'): string {
    return `${this.baseUrl}/export/${dataset}?format=${format}`;
  }
}
