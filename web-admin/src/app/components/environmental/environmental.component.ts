import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { GoogleMapsModule } from '@angular/google-maps';
import { AdminApiService } from '../../services/admin-api.service';
import { AdminEnvironmentalData, AttractantCategory } from '../../models/admin.model';

@Component({
  selector: 'app-environmental',
  standalone: true,
  imports: [CommonModule, FormsModule, GoogleMapsModule],
  template: `
    <div class="p-8 space-y-6 max-w-7xl mx-auto">
      <!-- Title & Tab Switcher (Matching Wildlife & Roadkill) -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h2 class="text-2xl font-bold tracking-tight text-white">Ecological & Crop Attractant Layers</h2>
          <p class="text-sm text-slate-400">Manage agricultural phenology, unmanaged fruit orchards, and livestock grazing attractants.</p>
        </div>

        <div class="flex items-center p-1 rounded-xl bg-slate-950 border border-slate-800 text-xs">
          <button
            (click)="activeTab.set('history')"
            [class.bg-slate-800]="activeTab() === 'history'"
            [class.text-white]="activeTab() === 'history'"
            class="px-4 py-2 rounded-lg font-semibold text-slate-400 hover:text-white transition-all">
            Attractant Layers Registry
          </button>
          <button
            (click)="activeTab.set('entry')"
            [class.bg-emerald-600]="activeTab() === 'entry'"
            [class.text-white]="activeTab() === 'entry'"
            class="px-4 py-2 rounded-lg font-semibold text-slate-400 hover:text-white transition-all flex items-center gap-1.5">
            <span>+ Log Attractant Layer</span>
          </button>
        </div>
      </div>

      <!-- Tab 1: Attractant Layers Registry -->
      @if (activeTab() === 'history') {
        <!-- Categories Overview Cards -->
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800">
            <div class="text-xs font-semibold text-emerald-400 uppercase tracking-wide">Thailand & APAC Core</div>
            <div class="text-sm font-bold text-white mt-1">Sugarcane & Fruit Phenology</div>
            <p class="text-[11px] text-slate-400 mt-1">High-sugar ripening attractants triggering Asian elephant crop-raiding corridors.</p>
          </div>

          <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800">
            <div class="text-xs font-semibold text-amber-400 uppercase tracking-wide">Indonesia & India Core</div>
            <div class="text-sm font-bold text-white mt-1">Livestock Corrals & Pastures</div>
            <p class="text-[11px] text-slate-400 mt-1">Unfenced night cattle/goat grazing areas attracting Sumatran & Bengal tigers.</p>
          </div>

          <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800">
            <div class="text-xs font-semibold text-blue-400 uppercase tracking-wide">Japan Validation Core</div>
            <div class="text-sm font-bold text-white mt-1">Abandoned Persimmon (Kaki)</div>
            <p class="text-[11px] text-slate-400 mt-1">Unharvested fruit clusters in aging rural valleys luring bears into residential zones.</p>
          </div>
        </div>

        <!-- Attractants Data Table -->
        <div class="rounded-2xl bg-slate-950/60 border border-slate-800 overflow-hidden shadow-sm">
          <div class="overflow-x-auto">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-900/80 text-slate-400 uppercase tracking-wider font-semibold border-b border-slate-800">
                <tr>
                  <th class="py-3 px-4">Category</th>
                  <th class="py-3 px-4">Attractant Layer Name</th>
                  <th class="py-3 px-4">Location Description</th>
                  <th class="py-3 px-4">Risk Weight</th>
                  <th class="py-3 px-4">Active Season / Harvest Window</th>
                  <th class="py-3 px-4">Last Updated</th>
                  <th class="py-3 px-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-800/60">
                @for (env of api.environmental(); track env.id) {
                  <tr class="hover:bg-slate-900/40 transition-colors">
                    <td class="py-3.5 px-4 whitespace-nowrap">
                      <span [class]="getCategoryClass(env.category)" class="text-[10px] font-bold px-2 py-0.5 rounded font-mono">
                        {{ env.category }}
                      </span>
                    </td>
                    <td class="py-3.5 px-4 font-bold text-white whitespace-nowrap">
                      {{ env.name }}
                    </td>
                    <td class="py-3.5 px-4 text-slate-300">
                      {{ env.location_desc }}
                      @if (env.circle_center) {
                        <div class="text-[10px] text-slate-500 font-mono">
                          {{ env.circle_center.lat | number:'1.3-3' }}, {{ env.circle_center.lng | number:'1.3-3' }} (r: {{ env.radius_meters }}m)
                        </div>
                      }
                    </td>
                    <td class="py-3.5 px-4 whitespace-nowrap font-mono font-bold text-amber-400">
                      {{ env.risk_weight | number:'1.2-2' }}
                    </td>
                    <td class="py-3.5 px-4 text-slate-300">
                      {{ env.active_season }}
                    </td>
                    <td class="py-3.5 px-4 text-slate-400 font-mono whitespace-nowrap">
                      {{ env.last_updated | date:'yyyy-MM-dd' }}
                    </td>
                    <td class="py-3.5 px-4 text-right whitespace-nowrap">
                      <button (click)="deleteEnvironmental(env.id)" class="px-2.5 py-1 rounded bg-rose-600/20 text-rose-300 hover:bg-rose-600/40 font-medium">
                        Delete
                      </button>
                    </td>
                  </tr>
                } @empty {
                  <tr>
                    <td colspan="7" class="py-8 text-center text-slate-500">No environmental attractant layers registered.</td>
                  </tr>
                }
              </tbody>
            </table>
          </div>
        </div>
      }

      <!-- Tab 2: New Attractant Layer Registration (2-Column Form + Interactive Google Map) -->
      @if (activeTab() === 'entry') {
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 bg-slate-950/60 border border-slate-800 rounded-2xl p-6 shadow-xl">
          <!-- Left Column: Form (5 Cols) -->
          <form (ngSubmit)="saveEnvironmental()" class="lg:col-span-5 space-y-4 text-xs">
            <div class="border-b border-slate-800 pb-3">
              <h3 class="text-base font-bold text-white">Register Ecological Risk / Attractant Zone</h3>
              <p class="text-[11px] text-slate-400">Click on the map on the right to pinpoint precise center coordinates.</p>
            </div>

            <div>
              <label class="block text-slate-300 font-medium mb-1">Attractant Category</label>
              <select [(ngModel)]="newItem.category" name="category" required
                class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-emerald-500">
                <option value="CROP_PHENOLOGY">CROP_PHENOLOGY (Sugarcane, Banana, Jackfruit)</option>
                <option value="PERSIMMON_KAKI">PERSIMMON_KAKI (Abandoned Persimmon Trees)</option>
                <option value="LIVESTOCK_PASTURE">LIVESTOCK_PASTURE (Cattle Grazing Corral)</option>
                <option value="WATER_SOURCE">WATER_SOURCE (Dry-season Watering Hole)</option>
              </select>
            </div>

            <div>
              <label class="block text-slate-300 font-medium mb-1">Layer / Zone Name</label>
              <input type="text" [(ngModel)]="newItem.name" name="name" required placeholder="e.g. Sector 5 Sugarcane Parcel"
                class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-emerald-500">
            </div>

            <!-- Auto-populated from Map -->
            <div class="p-3.5 rounded-xl bg-slate-900/80 border border-slate-700/80 space-y-2">
              <div class="flex items-center justify-between font-semibold text-slate-200">
                <span>Center Coordinates & Radius</span>
                <span class="text-[10px] text-emerald-400 font-normal">Click map to adjust center</span>
              </div>
              <div class="grid grid-cols-3 gap-2">
                <div>
                  <label class="text-[10px] text-slate-400">Center Lat</label>
                  <input type="number" step="0.00001" [(ngModel)]="itemLat" name="itemLat" required
                    class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white font-mono">
                </div>
                <div>
                  <label class="text-[10px] text-slate-400">Center Lng</label>
                  <input type="number" step="0.00001" [(ngModel)]="itemLng" name="itemLng" required
                    class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white font-mono">
                </div>
                <div>
                  <label class="text-[10px] text-slate-400">Radius (Meters)</label>
                  <input type="number" step="100" min="100" [(ngModel)]="itemRadius" name="itemRadius" required
                    class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white font-mono">
                </div>
              </div>
              <div>
                <label class="text-[10px] text-slate-400">Location Notes</label>
                <input type="text" [(ngModel)]="newItem.location_desc" name="location_desc" placeholder="e.g. Kaeng Krachan eastern perimeter border"
                  class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white">
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-300 font-medium mb-1">Risk Weight (0.1 - 1.0)</label>
                <input type="number" step="0.05" min="0.1" max="1.0" [(ngModel)]="newItem.risk_weight" name="risk_weight" required
                  class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-emerald-500 font-mono">
              </div>
              <div>
                <label class="block text-slate-300 font-medium mb-1">Active Season / Window</label>
                <input type="text" [(ngModel)]="newItem.active_season" name="active_season" placeholder="e.g. Oct - Jan (Harvest Peak)"
                  class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-emerald-500">
              </div>
            </div>

            <div class="flex items-center justify-between pt-4 border-t border-slate-800">
              <button type="button" (click)="activeTab.set('history')" class="px-4 py-2 rounded-lg bg-slate-800 text-slate-300">
                Cancel
              </button>
              <button type="submit" class="px-6 py-2.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 font-bold text-white shadow-lg shadow-emerald-600/20">
                Save Attractant Layer
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
              <!-- Center Marker -->
              <map-marker
                [position]="{ lat: itemLat, lng: itemLng }"
                [options]="markerOptions">
              </map-marker>
              <!-- Radius Circle Preview -->
              <map-circle
                [center]="{ lat: itemLat, lng: itemLng }"
                [radius]="itemRadius"
                [options]="circleOptions">
              </map-circle>
            </google-map>

            <div class="absolute top-3 left-3 bg-slate-950/80 backdrop-blur px-3 py-1.5 rounded-lg border border-slate-700 text-[11px] text-slate-300">
              📍 Click anywhere on map to set layer center & radius buffer
            </div>
          </div>
        </div>
      }
    </div>
  `
})
export class EnvironmentalComponent implements OnInit {
  api = inject(AdminApiService);
  activeTab = signal<'history' | 'entry'>('history');

  newItem: Partial<AdminEnvironmentalData> = {
    category: 'CROP_PHENOLOGY',
    name: 'Sector 5 Sugarcane Parcel',
    risk_weight: 0.85,
    active_season: 'October - January (Harvest Peak)',
    location_desc: 'Kaeng Krachan boundary agro-forest interface'
  };
  itemLat = 12.870;
  itemLng = 99.660;
  itemRadius = 2500;

  // Map settings
  mapCenter: google.maps.LatLngLiteral = { lat: 12.870, lng: 99.660 };
  mapZoom = 12;
  mapOptions: google.maps.MapOptions = {
    backgroundColor: '#0f172a',
    streetViewControl: false
  };
  markerOptions: google.maps.MarkerOptions = {
    draggable: true,
    title: 'Attractant Zone Center'
  };
  circleOptions: google.maps.CircleOptions = {
    fillColor: '#10b981',
    fillOpacity: 0.25,
    strokeColor: '#059669',
    strokeOpacity: 0.8,
    strokeWeight: 2
  };

  ngOnInit() {
    this.api.getEnvironmental().subscribe();
  }

  onMapClick(event: google.maps.MapMouseEvent) {
    if (event.latLng) {
      this.itemLat = Number(event.latLng.lat().toFixed(5));
      this.itemLng = Number(event.latLng.lng().toFixed(5));
      this.newItem.location_desc = `Geocoded Spot: Lat ${this.itemLat}, Lng ${this.itemLng}`;
    }
  }

  saveEnvironmental() {
    const payload: Partial<AdminEnvironmentalData> = {
      ...this.newItem,
      geometry_type: 'circle',
      circle_center: { lat: this.itemLat, lng: this.itemLng },
      radius_meters: Number(this.itemRadius)
    };
    this.api.createEnvironmental(payload).subscribe({
      next: () => {
        this.activeTab.set('history');
      }
    });
  }

  deleteEnvironmental(id: string) {
    if (confirm('Delete this environmental layer?')) {
      this.api.deleteEnvironmental(id).subscribe();
    }
  }

  getCategoryClass(cat: AttractantCategory): string {
    if (cat === 'CROP_PHENOLOGY') return 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30';
    if (cat === 'PERSIMMON_KAKI') return 'bg-blue-500/20 text-blue-400 border border-blue-500/30';
    if (cat === 'LIVESTOCK_PASTURE') return 'bg-amber-500/20 text-amber-400 border border-amber-500/30';
    return 'bg-purple-500/20 text-purple-400 border border-purple-500/30';
  }
}
