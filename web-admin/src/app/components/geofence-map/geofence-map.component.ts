import { Component, OnInit, inject, signal, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { GoogleMapsModule, MapInfoWindow, MapPolygon, MapCircle } from '@angular/google-maps';
import { AdminApiService } from '../../services/admin-api.service';
import { AdminGeofence, ThreatLevel, GeometryType } from '../../models/admin.model';

@Component({
  selector: 'app-geofence-map',
  standalone: true,
  imports: [CommonModule, FormsModule, GoogleMapsModule],
  template: `
    <div class="h-[calc(100vh-4rem)] flex flex-col">
      <!-- Toolbar & Actions -->
      <div class="bg-slate-950 border-b border-slate-800 px-6 py-3.5 flex flex-wrap items-center justify-between gap-4">
        <div class="flex items-center gap-3">
          <h2 class="text-lg font-bold text-white flex items-center gap-2">
            JAPAC Dynamic Geofence Console
            <span class="text-xs font-mono font-normal px-2 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700">
              {{ api.geofences().length }} Active Zones
            </span>
          </h2>
          <!-- Filter pills -->
          <div class="flex items-center gap-1.5 ml-4">
            <button (click)="filterSource.set('all')" [class.bg-slate-700]="filterSource() === 'all'" class="px-2.5 py-1 text-xs rounded text-slate-300 hover:bg-slate-800">All</button>
            <button (click)="filterSource.set('ai')" [class.bg-blue-600]="filterSource() === 'ai'" class="px-2.5 py-1 text-xs rounded text-slate-300 hover:bg-slate-800">AI Only</button>
            <button (click)="filterSource.set('manual')" [class.bg-purple-600]="filterSource() === 'manual'" class="px-2.5 py-1 text-xs rounded text-slate-300 hover:bg-slate-800">Manual Only</button>
          </div>
        </div>

        <!-- Controls: Add Manual Fence & Bulk Delete -->
        <div class="flex items-center gap-3">
          <button
            (click)="openAddModal()"
            class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-purple-600 hover:bg-purple-500 text-white shadow transition-all">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            + Create Manual Geofence
          </button>

          <button
            (click)="confirmClearAllManual()"
            class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-rose-900/60 hover:bg-rose-800 text-rose-200 border border-rose-700/60 transition-all">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            Clear All Manual Fences
          </button>

          <button
            (click)="resetToJapacView()"
            class="px-2.5 py-1.5 rounded-lg text-xs bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700">
            Reset JAPAC View
          </button>
        </div>
      </div>

      <!-- Main Map & Overlay Panel -->
      <div class="flex-1 relative overflow-hidden">
        <!-- Google Map Canvas -->
        <google-map
          #map
          [height]="'100%'"
          [width]="'100%'"
          [center]="center"
          [zoom]="zoom"
          [options]="mapOptions">

          <!-- Polygons -->
          @for (geo of visibleGeofences(); track geo.id) {
            @if (geo.geometry_type === 'polygon' && geo.coordinates) {
              <map-polygon
                [paths]="geo.coordinates"
                [options]="getPolygonOptions(geo)"
                (polygonClick)="onGeofenceClick(geo)">
              </map-polygon>
            }
          }

          <!-- Circles -->
          @for (geo of visibleGeofences(); track geo.id) {
            @if (geo.geometry_type === 'circle' && geo.circle_center) {
              <map-circle
                [center]="geo.circle_center"
                [radius]="geo.circle_radius_meters || 2000"
                [options]="getCircleOptions(geo)"
                (circleClick)="onGeofenceClick(geo)">
              </map-circle>
            }
          }

          <!-- Info Window -->
          <map-info-window #infoWindow>
            @if (selectedGeofence(); as g) {
              <div class="p-2 text-slate-900 max-w-xs space-y-1">
                <div class="flex items-center justify-between gap-2">
                  <strong class="text-xs font-bold">{{ g.name }}</strong>
                  <span [class]="getBadgeClass(g.threat_level)" class="text-[9px] px-1.5 py-0.5 rounded font-bold uppercase">
                    {{ g.threat_level.replace('TIER_', 'T') }}
                  </span>
                </div>
                <div class="text-[11px] text-slate-600">
                  Target: <strong>{{ g.target_species }}</strong>
                </div>
                <div class="text-[11px] text-slate-600">
                  Source: <span class="font-mono uppercase font-semibold" [class.text-purple-600]="g.source === 'manual'">{{ g.source }}</span>
                  @if (g.is_persistent) {
                    <span class="ml-1 text-emerald-600 font-bold">(Persistent)</span>
                  }
                </div>
                <p class="text-[11px] text-slate-700 pt-1 border-t border-slate-200">{{ g.reason }}</p>
                <div class="pt-2 flex justify-end">
                  <button (click)="deleteSelectedGeofence(g.id)" class="text-[10px] text-rose-600 hover:text-rose-800 font-bold">
                    Delete Geofence
                  </button>
                </div>
              </div>
            }
          </map-info-window>
        </google-map>

        <!-- Legend Overlay in Corner -->
        <div class="absolute bottom-6 left-6 p-4 rounded-xl bg-slate-950/90 backdrop-blur border border-slate-800 shadow-xl space-y-2.5 z-10 text-xs">
          <div class="font-bold text-white text-[11px] uppercase tracking-wider">Geofence Tiers & Sources</div>
          <div class="flex items-center gap-2">
            <span class="w-3.5 h-3.5 rounded bg-rose-500/50 border border-rose-500"></span>
            <span class="text-slate-300">Tier 3 (Critical Danger)</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="w-3.5 h-3.5 rounded bg-amber-500/50 border border-amber-500"></span>
            <span class="text-slate-300">Tier 2 (Warning / Alert)</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="w-3.5 h-3.5 rounded bg-yellow-400/50 border border-yellow-400"></span>
            <span class="text-slate-300">Tier 1 (Notice / Caution)</span>
          </div>
          <div class="flex items-center gap-2 pt-1 border-t border-slate-800">
            <span class="w-3.5 h-3.5 rounded border-2 border-purple-400 bg-purple-500/30"></span>
            <span class="text-purple-300 font-semibold">Manual Debug (Persistent)</span>
          </div>
        </div>

        <!-- Geofence List Drawer (Right Side) -->
        <div class="absolute top-4 right-4 w-80 max-h-[calc(100%-2rem)] bg-slate-950/90 backdrop-blur border border-slate-800 rounded-2xl shadow-2xl p-4 flex flex-col z-10">
          <div class="flex items-center justify-between pb-3 border-b border-slate-800">
            <h3 class="text-xs font-bold text-white uppercase tracking-wider">Zone Inspector</h3>
            <span class="text-[10px] text-slate-400">{{ visibleGeofences().length }} listed</span>
          </div>
          <div class="flex-1 overflow-y-auto space-y-2 mt-3 pr-1">
            @for (geo of visibleGeofences(); track geo.id) {
              <div
                (click)="selectAndPanTo(geo)"
                class="p-3 rounded-xl border border-slate-800/80 hover:border-slate-700 bg-slate-900/60 cursor-pointer transition-all">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-xs font-semibold text-white truncate max-w-[170px]">{{ geo.name }}</span>
                  <span [class]="getBadgeClass(geo.threat_level)" class="text-[9px] px-1.5 py-0.5 rounded font-bold">
                    {{ geo.threat_level.includes('TIER_3') ? 'T3' : geo.threat_level.includes('TIER_2') ? 'T2' : 'T1' }}
                  </span>
                </div>
                <div class="text-[11px] text-slate-400 flex items-center justify-between">
                  <span>{{ geo.target_species }}</span>
                  <span class="font-mono text-[10px]" [class.text-purple-400]="geo.source === 'manual'">
                    {{ geo.geometry_type | uppercase }} • {{ geo.source }}
                  </span>
                </div>
              </div>
            }
          </div>
        </div>
      </div>
    </div>

    <!-- Modal: Add Manual Geofence (Circle / Polygon) -->
    @if (showAddModal()) {
      <div class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 z-50">
        <div class="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-5">
          <div class="flex items-center justify-between border-b border-slate-800 pb-3">
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <span class="p-1.5 rounded-lg bg-purple-500/20 text-purple-400">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
              </span>
              Add Manual Geofence (Debug Override)
            </h3>
            <button (click)="showAddModal.set(false)" class="text-slate-400 hover:text-white">✕</button>
          </div>

          <form (ngSubmit)="saveManualGeofence()" class="space-y-4 text-xs">
            <div>
              <label class="block text-slate-300 font-medium mb-1">Geofence Name</label>
              <input type="text" [(ngModel)]="newFence.name" name="name" required placeholder="e.g. Field Station 3 Thermal Perimeter"
                class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white focus:outline-none focus:border-purple-500">
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-300 font-medium mb-1">Geometry Type</label>
                <select [(ngModel)]="newFence.geometry_type" name="geometry_type"
                  class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white focus:outline-none focus:border-purple-500">
                  <option value="circle">Circle (Center + Radius)</option>
                  <option value="polygon">Polygon (Multi-point Coordinates)</option>
                </select>
              </div>
              <div>
                <label class="block text-slate-300 font-medium mb-1">Threat Tier</label>
                <select [(ngModel)]="newFence.threat_level" name="threat_level"
                  class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white focus:outline-none focus:border-purple-500">
                  <option value="TIER_1_CAUTION">Tier 1 - Caution</option>
                  <option value="TIER_2_ALERT">Tier 2 - Alert / Warning</option>
                  <option value="TIER_3_CRITICAL">Tier 3 - Critical Danger</option>
                </select>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-300 font-medium mb-1">Target Species</label>
                <select [(ngModel)]="newFence.target_species" name="target_species"
                  class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white focus:outline-none focus:border-purple-500">
                  <option value="Asian Elephant">Asian Elephant (Thailand/India)</option>
                  <option value="Sumatran Tiger">Sumatran Tiger (Indonesia)</option>
                  <option value="Bengal Tiger">Bengal Tiger (India/Nepal)</option>
                  <option value="Brown Bear">Brown Bear (Japan - Hokkaido)</option>
                  <option value="Asian Black Bear">Asian Black Bear (Japan - Honshu)</option>
                  <option value="Other / Multi">Other / Multi-species</option>
                </select>
              </div>
              <div>
                <label class="block text-slate-300 font-medium mb-1">Persistence Policy</label>
                <div class="p-2 rounded-lg bg-slate-800/80 border border-slate-700 text-emerald-400 font-semibold flex items-center gap-1.5">
                  <span>✓</span> Protected from AI Auto-Overwrites
                </div>
              </div>
            </div>

            <!-- Circle specific inputs -->
            @if (newFence.geometry_type === 'circle') {
              <div class="p-3.5 rounded-xl bg-slate-800/50 border border-slate-700 space-y-2">
                <div class="font-semibold text-slate-200">Circle Center & Radius</div>
                <div class="grid grid-cols-3 gap-2">
                  <div>
                    <label class="text-[10px] text-slate-400">Center Lat</label>
                    <input type="number" step="0.0001" [(ngModel)]="circleLat" name="circleLat"
                      class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white">
                  </div>
                  <div>
                    <label class="text-[10px] text-slate-400">Center Lng</label>
                    <input type="number" step="0.0001" [(ngModel)]="circleLng" name="circleLng"
                      class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white">
                  </div>
                  <div>
                    <label class="text-[10px] text-slate-400">Radius (Meters)</label>
                    <input type="number" step="100" [(ngModel)]="circleRadius" name="circleRadius"
                      class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white">
                  </div>
                </div>
              </div>
            } @else {
              <!-- Polygon specific coordinates input -->
              <div class="p-3.5 rounded-xl bg-slate-800/50 border border-slate-700 space-y-2">
                <div class="font-semibold text-slate-200">Polygon Coordinate Vertices (Lat, Lng)</div>
                <p class="text-[10px] text-slate-400">One coordinate pair per line (e.g. 12.875, 99.630)</p>
                <textarea rows="3" [(ngModel)]="polygonCoordsText" name="polygonCoordsText"
                  class="w-full p-2 rounded bg-slate-800 border border-slate-600 text-white font-mono text-[11px]"></textarea>
              </div>
            }

            <div>
              <label class="block text-slate-300 font-medium mb-1">Operational Reason / Debug Notes</label>
              <textarea rows="2" [(ngModel)]="newFence.reason" name="reason" placeholder="e.g. Night camera sensor alert override"
                class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white focus:outline-none focus:border-purple-500"></textarea>
            </div>

            <div class="flex justify-end gap-3 pt-3 border-t border-slate-800">
              <button type="button" (click)="showAddModal.set(false)" class="px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300">
                Cancel
              </button>
              <button type="submit" class="px-5 py-2 rounded-lg bg-purple-600 hover:bg-purple-500 font-semibold text-white shadow">
                Save & Deploy Geofence
              </button>
            </div>
          </form>
        </div>
      </div>
    }
  `
})
export class GeofenceMapComponent implements OnInit {
  api = inject(AdminApiService);
  @ViewChild('infoWindow') infoWindow?: MapInfoWindow;

  // JAPAC Initial Centered Viewport: Lat 15.0 N, Lng 105.0 E (Covering Japan to Southeast Asia/India)
  center: google.maps.LatLngLiteral = { lat: 15.0, lng: 105.0 };
  zoom = 4;

  mapOptions: google.maps.MapOptions = {
    mapTypeId: 'roadmap',
    backgroundColor: '#0f172a',
    disableDefaultUI: false,
    zoomControl: true,
    mapTypeControl: true,
    streetViewControl: false,
    fullscreenControl: true,
    styles: [
      { elementType: 'geometry', stylers: [{ color: '#1e293b' }] },
      { elementType: 'labels.text.stroke', stylers: [{ color: '#0f172a' }] },
      { elementType: 'labels.text.fill', stylers: [{ color: '#94a3b8' }] },
      { featureType: 'water', elementType: 'geometry', stylers: [{ color: '#0f172a' }] },
      { featureType: 'road', elementType: 'geometry', stylers: [{ color: '#334155' }] }
    ]
  };

  filterSource = signal<'all' | 'ai' | 'manual'>('all');
  selectedGeofence = signal<AdminGeofence | null>(null);
  showAddModal = signal<boolean>(false);

  // New fence form fields
  newFence: Partial<AdminGeofence> = {
    name: '',
    geometry_type: 'circle',
    threat_level: 'TIER_2_ALERT',
    target_species: 'Asian Elephant',
    reason: ''
  };
  circleLat = 12.875;
  circleLng = 99.650;
  circleRadius = 3000;
  polygonCoordsText = '12.875, 99.630\n12.890, 99.670\n12.860, 99.695\n12.845, 99.650';

  ngOnInit() {
    this.api.getGeofences().subscribe();
  }

  visibleGeofences() {
    const list = this.api.geofences();
    const filter = this.filterSource();
    if (filter === 'all') return list;
    return list.filter(g => g.source === filter);
  }

  resetToJapacView() {
    this.center = { lat: 15.0, lng: 105.0 };
    this.zoom = 4;
  }

  selectAndPanTo(geo: AdminGeofence) {
    this.selectedGeofence.set(geo);
    if (geo.geometry_type === 'circle' && geo.circle_center) {
      this.center = { lat: geo.circle_center.lat, lng: geo.circle_center.lng };
      this.zoom = 11;
    } else if (geo.coordinates && geo.coordinates.length > 0) {
      this.center = { lat: geo.coordinates[0].lat, lng: geo.coordinates[0].lng };
      this.zoom = 11;
    }
  }

  onGeofenceClick(geo: AdminGeofence) {
    this.selectedGeofence.set(geo);
    // infoWindow open can be anchored
  }

  deleteSelectedGeofence(id: string) {
    if (confirm('Delete this geofence?')) {
      this.api.deleteGeofence(id).subscribe();
      this.selectedGeofence.set(null);
    }
  }

  confirmClearAllManual() {
    if (confirm('Are you sure you want to PURGE ALL MANUAL GEOFENCES? AI-generated fences will remain intact.')) {
      this.api.clearAllManualGeofences().subscribe();
    }
  }

  openAddModal() {
    this.newFence = {
      name: '',
      geometry_type: 'circle',
      threat_level: 'TIER_2_ALERT',
      target_species: 'Asian Elephant',
      reason: ''
    };
    this.showAddModal.set(true);
  }

  saveManualGeofence() {
    const toSave: Partial<AdminGeofence> = {
      ...this.newFence,
      source: 'manual',
      is_persistent: true
    };

    if (toSave.geometry_type === 'circle') {
      toSave.circle_center = { lat: Number(this.circleLat), lng: Number(this.circleLng) };
      toSave.circle_radius_meters = Number(this.circleRadius);
    } else {
      const lines = this.polygonCoordsText.split('\n');
      const coords = lines.map(line => {
        const parts = line.split(',');
        return { lat: parseFloat(parts[0]?.trim() || '0'), lng: parseFloat(parts[1]?.trim() || '0') };
      }).filter(c => !isNaN(c.lat) && !isNaN(c.lng) && c.lat !== 0);
      toSave.coordinates = coords;
    }

    this.api.createGeofence(toSave).subscribe({
      next: () => {
        this.showAddModal.set(false);
      }
    });
  }

  getPolygonOptions(geo: AdminGeofence): google.maps.PolygonOptions {
    const isManual = geo.source === 'manual';
    let fillColor = '#facc15';
    let strokeColor = '#eab308';

    if (geo.threat_level === 'TIER_3_CRITICAL') {
      fillColor = '#ef4444';
      strokeColor = '#dc2626';
    } else if (geo.threat_level === 'TIER_2_ALERT') {
      fillColor = '#f97316';
      strokeColor = '#ea580c';
    }

    if (isManual) {
      strokeColor = '#a855f7'; // Purple border for manual
    }

    return {
      fillColor,
      fillOpacity: 0.35,
      strokeColor,
      strokeOpacity: 0.9,
      strokeWeight: isManual ? 3 : 2,
      clickable: true
    };
  }

  getCircleOptions(geo: AdminGeofence): google.maps.CircleOptions {
    const isManual = geo.source === 'manual';
    let fillColor = '#facc15';
    let strokeColor = '#eab308';

    if (geo.threat_level === 'TIER_3_CRITICAL') {
      fillColor = '#ef4444';
      strokeColor = '#dc2626';
    } else if (geo.threat_level === 'TIER_2_ALERT') {
      fillColor = '#f97316';
      strokeColor = '#ea580c';
    }

    if (isManual) {
      strokeColor = '#a855f7';
    }

    return {
      fillColor,
      fillOpacity: 0.35,
      strokeColor,
      strokeOpacity: 0.9,
      strokeWeight: isManual ? 3 : 2,
      clickable: true
    };
  }

  getBadgeClass(tier: ThreatLevel): string {
    if (tier === 'TIER_3_CRITICAL') return 'bg-rose-500/20 text-rose-400 border border-rose-500/30';
    if (tier === 'TIER_2_ALERT') return 'bg-amber-500/20 text-amber-400 border border-amber-500/30';
    return 'bg-yellow-500/20 text-yellow-400 border border-yellow-500/30';
  }
}
