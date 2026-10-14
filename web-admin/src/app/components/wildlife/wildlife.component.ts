import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { GoogleMapsModule } from '@angular/google-maps';
import { AdminApiService } from '../../services/admin-api.service';
import { AdminWildlifeIncident, ThreatLevel, WildlifeStatus } from '../../models/admin.model';

@Component({
  selector: 'app-wildlife',
  standalone: true,
  imports: [CommonModule, FormsModule, GoogleMapsModule],
  template: `
    <div class="p-8 space-y-6 max-w-7xl mx-auto">
      <!-- Title & Tab Switcher (Matching Roadkill style) -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h2 class="text-2xl font-bold tracking-tight text-white">Wildlife Encounter & Sighting Registry</h2>
          <p class="text-sm text-slate-400">Monitor reported apex carnivore and megafauna incursions in real-time.</p>
        </div>

        <div class="flex items-center p-1 rounded-xl bg-slate-950 border border-slate-800 text-xs">
          <button
            (click)="activeTab.set('history')"
            [class.bg-slate-800]="activeTab() === 'history'"
            [class.text-white]="activeTab() === 'history'"
            class="px-4 py-2 rounded-lg font-semibold text-slate-400 hover:text-white transition-all">
            Sightings Registry
          </button>
          <button
            (click)="activeTab.set('entry')"
            [class.bg-emerald-600]="activeTab() === 'entry'"
            [class.text-white]="activeTab() === 'entry'"
            class="px-4 py-2 rounded-lg font-semibold text-slate-400 hover:text-white transition-all flex items-center gap-1.5">
            <span>+ Log Wildlife Sighting</span>
          </button>
        </div>
      </div>

      <!-- Tab 1: Sightings Registry (Table List & Filters) -->
      @if (activeTab() === 'history') {
        <!-- Filters & Search -->
        <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800 flex flex-wrap items-center gap-3 text-xs">
          <div class="flex-1 min-w-[200px]">
            <input
              type="text"
              [(ngModel)]="searchQuery"
              placeholder="Search by species, road name, or reporter..."
              class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white placeholder-slate-500 focus:outline-none focus:border-emerald-500">
          </div>
          <div>
            <select [(ngModel)]="filterSpecies" class="px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-slate-300">
              <option value="">All Species</option>
              <option value="Asian Elephant">Asian Elephant</option>
              <option value="Sumatran Tiger">Sumatran Tiger</option>
              <option value="Bengal Tiger">Bengal Tiger</option>
              <option value="Brown Bear">Brown Bear</option>
              <option value="Asian Black Bear">Asian Black Bear</option>
            </select>
          </div>
          <div>
            <select [(ngModel)]="filterTier" class="px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-slate-300">
              <option value="">All Threat Tiers</option>
              <option value="TIER_3_CRITICAL">Tier 3 (Critical)</option>
              <option value="TIER_2_ALERT">Tier 2 (Alert)</option>
              <option value="TIER_1_CAUTION">Tier 1 (Caution)</option>
            </select>
          </div>
        </div>

        <!-- Incidents Data Table -->
        <div class="rounded-2xl bg-slate-950/60 border border-slate-800 overflow-hidden shadow-sm">
          <div class="overflow-x-auto">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-900/80 text-slate-400 uppercase tracking-wider font-semibold border-b border-slate-800">
                <tr>
                  <th class="py-3 px-4">Reported At</th>
                  <th class="py-3 px-4">Species</th>
                  <th class="py-3 px-4">Location & Coordinates</th>
                  <th class="py-3 px-4">Threat Level</th>
                  <th class="py-3 px-4">Status</th>
                  <th class="py-3 px-4">Reporter / Source</th>
                  <th class="py-3 px-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-800/60">
                @for (item of filteredList(); track item.id) {
                  <tr class="hover:bg-slate-900/40 transition-colors">
                    <td class="py-3.5 px-4 text-slate-400 font-mono whitespace-nowrap">
                      {{ item.reported_at | date:'yyyy-MM-dd HH:mm' }}
                    </td>
                    <td class="py-3.5 px-4 font-semibold text-white whitespace-nowrap">
                      {{ item.species }}
                    </td>
                    <td class="py-3.5 px-4 text-slate-300">
                      <div class="font-medium">{{ item.location.road_name || 'Forest Interface' }}</div>
                      <div class="text-[11px] text-slate-500 font-mono">{{ item.location.lat | number:'1.3-3' }}, {{ item.location.lng | number:'1.3-3' }}</div>
                    </td>
                    <td class="py-3.5 px-4 whitespace-nowrap">
                      <span [class]="getBadgeClass(item.threat_level)" class="text-[10px] px-2 py-0.5 rounded font-bold uppercase">
                        {{ item.threat_level.replace('TIER_', 'T') }}
                      </span>
                    </td>
                    <td class="py-3.5 px-4 whitespace-nowrap">
                      <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-[10px] font-semibold"
                        [class]="getStatusClass(item.status)">
                        {{ item.status }}
                      </span>
                    </td>
                    <td class="py-3.5 px-4 text-slate-400">
                      {{ item.reporter }}
                    </td>
                    <td class="py-3.5 px-4 text-right whitespace-nowrap space-x-2">
                      @if (item.status === 'UNVERIFIED') {
                        <button (click)="verifyIncident(item.id)" class="px-2.5 py-1 rounded bg-emerald-600/30 text-emerald-300 hover:bg-emerald-600/50 font-medium">
                          Verify
                        </button>
                      }
                      <button (click)="deleteIncident(item.id)" class="px-2.5 py-1 rounded bg-rose-600/20 text-rose-300 hover:bg-rose-600/40 font-medium">
                        Delete
                      </button>
                    </td>
                  </tr>
                } @empty {
                  <tr>
                    <td colspan="7" class="py-8 text-center text-slate-500">No wildlife incidents match the active filters.</td>
                  </tr>
                }
              </tbody>
            </table>
          </div>
        </div>
      }

      <!-- Tab 2: New Sighting Registration (2-Column Form + Interactive Google Map) -->
      @if (activeTab() === 'entry') {
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 bg-slate-950/60 border border-slate-800 rounded-2xl p-6 shadow-xl">
          <!-- Left Column: Form (5 Cols) -->
          <form (ngSubmit)="saveIncident()" class="lg:col-span-5 space-y-4 text-xs">
            <div class="border-b border-slate-800 pb-3">
              <h3 class="text-base font-bold text-white">Log Wildlife Sighting or Encounter</h3>
              <p class="text-[11px] text-slate-400">Click on the map on the right to pinpoint precise sighting GPS coordinates.</p>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-300 font-medium mb-1">Target Species</label>
                <select [(ngModel)]="newItem.species" name="species" required
                  class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-emerald-500">
                  <option value="Asian Elephant">Asian Elephant (Thailand/India)</option>
                  <option value="Sumatran Tiger">Sumatran Tiger (Indonesia)</option>
                  <option value="Bengal Tiger">Bengal Tiger (India/Nepal)</option>
                  <option value="Brown Bear">Brown Bear (Japan - Hokkaido)</option>
                  <option value="Asian Black Bear">Asian Black Bear (Japan - Honshu)</option>
                </select>
              </div>
              <div>
                <label class="block text-slate-300 font-medium mb-1">Threat Level</label>
                <select [(ngModel)]="newItem.threat_level" name="threat_level" required
                  class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-emerald-500">
                  <option value="TIER_1_CAUTION">Tier 1 - Caution (Sighting)</option>
                  <option value="TIER_2_ALERT">Tier 2 - Alert (Crop/Livestock Risk)</option>
                  <option value="TIER_3_CRITICAL">Tier 3 - Critical Danger (Imminent Attack)</option>
                </select>
              </div>
            </div>

            <!-- Auto-populated from Map -->
            <div class="p-3.5 rounded-xl bg-slate-900/80 border border-slate-700/80 space-y-2">
              <div class="flex items-center justify-between font-semibold text-slate-200">
                <span>Selected Coordinates</span>
                <span class="text-[10px] text-emerald-400 font-normal">Click map to adjust pin</span>
              </div>
              <div class="grid grid-cols-2 gap-2">
                <div>
                  <label class="text-[10px] text-slate-400">Latitude</label>
                  <input type="number" step="0.00001" [(ngModel)]="pinLat" name="lat" required
                    class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white font-mono">
                </div>
                <div>
                  <label class="text-[10px] text-slate-400">Longitude</label>
                  <input type="number" step="0.00001" [(ngModel)]="pinLng" name="lng" required
                    class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white font-mono">
                </div>
              </div>
              <div>
                <label class="text-[10px] text-slate-400">Road / Sector Name</label>
                <input type="text" [(ngModel)]="itemRoad" name="roadName" placeholder="e.g. Rural Route 3219, KM 14"
                  class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white">
              </div>
            </div>

            <div>
              <label class="block text-slate-300 font-medium mb-1">Reporter / Field Unit</label>
              <input type="text" [(ngModel)]="newItem.reporter" name="reporter" placeholder="e.g. Officer Somchai (Kaeng Krachan Unit)"
                class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white placeholder-slate-500">
            </div>

            <div>
              <label class="block text-slate-300 font-medium mb-1">Description / Observations</label>
              <textarea rows="3" [(ngModel)]="newItem.description" name="description"
                placeholder="Observation details, herd size, direction of travel, animal behavior..."
                class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-emerald-500"></textarea>
            </div>

            <div class="flex items-center justify-between pt-4 border-t border-slate-800">
              <button type="button" (click)="activeTab.set('history')" class="px-4 py-2 rounded-lg bg-slate-800 text-slate-300">
                Cancel
              </button>
              <button type="submit" class="px-6 py-2.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 font-bold text-white shadow-lg shadow-emerald-600/20">
                Log Sighting & Deploy Zone
              </button>
            </div>
          </form>

          <!-- Right Column: Interactive Google Map (7 Cols) -->
          <div class="lg:col-span-7 h-[500px] rounded-xl overflow-hidden border border-slate-700 relative">
            <google-map
              [height]="'100%'"
              [width]="'100%'"
              [center]="mapCenter"
              [zoom]="mapZoom"
              (mapClick)="onMapClick($event)"
              [options]="mapOptions">
              <map-marker
                [position]="{ lat: pinLat, lng: pinLng }"
                [options]="markerOptions">
              </map-marker>
            </google-map>

            <div class="absolute top-3 left-3 bg-slate-950/80 backdrop-blur px-3 py-1.5 rounded-lg border border-slate-700 text-[11px] text-slate-300">
              📍 Click anywhere on map to position sighting pin
            </div>
          </div>
        </div>
      }
    </div>
  `
})
export class WildlifeComponent implements OnInit {
  api = inject(AdminApiService);
  activeTab = signal<'history' | 'entry'>('history');
  searchQuery = '';
  filterSpecies = '';
  filterTier = '';

  // Form Fields
  newItem: Partial<AdminWildlifeIncident> = {
    species: 'Asian Elephant',
    threat_level: 'TIER_3_CRITICAL',
    reporter: 'Field Ranger Patrol',
    description: ''
  };
  pinLat = 12.865;
  pinLng = 99.655;
  itemRoad = 'Rural Highway 3219, Sector 4';

  // Map settings
  mapCenter: google.maps.LatLngLiteral = { lat: 12.865, lng: 99.655 };
  mapZoom = 12;
  mapOptions: google.maps.MapOptions = {
    backgroundColor: '#0f172a',
    streetViewControl: false
  };
  markerOptions: google.maps.MarkerOptions = {
    draggable: true,
    title: 'Wildlife Sighting Location'
  };

  ngOnInit() {
    this.api.getWildlife().subscribe();
  }

  filteredList() {
    return this.api.wildlife().filter(item => {
      const q = this.searchQuery.toLowerCase();
      const matchQ = !q ||
        item.species.toLowerCase().includes(q) ||
        item.reporter.toLowerCase().includes(q) ||
        (item.location.road_name && item.location.road_name.toLowerCase().includes(q));
      const matchSpecies = !this.filterSpecies || item.species === this.filterSpecies;
      const matchTier = !this.filterTier || item.threat_level === this.filterTier;
      return matchQ && matchSpecies && matchTier;
    });
  }

  onMapClick(event: google.maps.MapMouseEvent) {
    if (event.latLng) {
      this.pinLat = Number(event.latLng.lat().toFixed(5));
      this.pinLng = Number(event.latLng.lng().toFixed(5));
      this.itemRoad = `Geocoded Spot: Lat ${this.pinLat}, Lng ${this.pinLng}`;
    }
  }

  saveIncident() {
    const payload: Partial<AdminWildlifeIncident> = {
      ...this.newItem,
      location: {
        lat: this.pinLat,
        lng: this.pinLng,
        road_name: this.itemRoad
      },
      status: 'VERIFIED'
    };
    this.api.createWildlife(payload).subscribe({
      next: () => {
        this.activeTab.set('history');
      }
    });
  }

  verifyIncident(id: string) {
    this.api.updateWildlifeStatus(id, 'VERIFIED').subscribe();
  }

  deleteIncident(id: string) {
    if (confirm('Delete this wildlife report?')) {
      this.api.deleteWildlife(id).subscribe();
    }
  }

  getBadgeClass(tier: ThreatLevel): string {
    if (tier === 'TIER_3_CRITICAL') return 'bg-rose-500/20 text-rose-400 border border-rose-500/30';
    if (tier === 'TIER_2_ALERT') return 'bg-amber-500/20 text-amber-400 border border-amber-500/30';
    return 'bg-yellow-500/20 text-yellow-400 border border-yellow-500/30';
  }

  getStatusClass(status: string): string {
    if (status === 'VERIFIED') return 'bg-emerald-500/20 text-emerald-400';
    if (status === 'UNVERIFIED') return 'bg-amber-500/20 text-amber-400';
    if (status === 'RESOLVED') return 'bg-blue-500/20 text-blue-400';
    return 'bg-slate-800 text-slate-400';
  }
}
