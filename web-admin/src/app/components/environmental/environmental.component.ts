import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { GoogleMapsModule } from '@angular/google-maps';
import { AdminApiService } from '../../services/admin-api.service';
import { AdminEnvironmentalData, AttractantCategory } from '../../models/admin.model';

interface CategoryPreset {
  category: AttractantCategory;
  title: string;
  badge: string;
  badgeColor: string;
  description: string;
  defaultRadius: number;
  defaultRisk: number;
  defaultSeason: string;
  defaultName: string;
  defaultDesc: string;
  mapCenter: google.maps.LatLngLiteral;
  mapZoom: number;
}

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
            (click)="startNewRegistration()"
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

      <!-- Tab 2: New Attractant Layer Registration -->
      @if (activeTab() === 'entry') {
        <!-- Step 1: Category Selection Cards (Poka-Yoke Flow) -->
        @if (!selectedCategory()) {
          <div class="bg-slate-950/60 border border-slate-800 rounded-2xl p-8 space-y-6 shadow-xl">
            <div class="text-center max-w-xl mx-auto space-y-2">
              <span class="px-3 py-1 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-semibold uppercase tracking-wider">
                Step 1 of 2: Select Ecological Context
              </span>
              <h3 class="text-xl font-bold text-white">Choose the Attractant Hazard Category</h3>
              <p class="text-xs text-slate-400">
                Each category pre-configures ecological scale parameters (default buffer radius, risk weights, and map focus).
              </p>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-5 mt-6">
              @for (preset of categoryPresets; track preset.category) {
                <div
                  (click)="selectCategory(preset.category)"
                  class="p-6 rounded-2xl border border-slate-800 bg-slate-900/60 hover:bg-slate-900 hover:border-emerald-500/60 transition-all cursor-pointer group space-y-3 relative overflow-hidden">
                  <div class="flex items-center justify-between">
                    <span [class]="preset.badgeColor" class="text-[10px] font-bold px-2 py-0.5 rounded font-mono uppercase">
                      {{ preset.badge }}
                    </span>
                    <span class="text-xs font-mono text-emerald-400 font-bold bg-slate-800 px-2 py-0.5 rounded">
                      Default: {{ preset.defaultRadius | number }}m
                    </span>
                  </div>

                  <h4 class="text-base font-bold text-white group-hover:text-emerald-400 transition-colors flex items-center gap-2">
                    {{ preset.title }}
                    <span class="text-slate-500 group-hover:translate-x-1 transition-transform">→</span>
                  </h4>

                  <p class="text-xs text-slate-400 leading-relaxed">
                    {{ preset.description }}
                  </p>

                  <div class="pt-2 flex items-center justify-between text-[11px] text-slate-500 border-t border-slate-800/80">
                    <span>Peak: <strong class="text-slate-300">{{ preset.defaultSeason }}</strong></span>
                    <span>Risk Weight: <strong class="text-amber-400">{{ preset.defaultRisk }}</strong></span>
                  </div>
                </div>
              }
            </div>
          </div>
        }

        <!-- Step 2: Form & Interactive Map (Rendered ONLY after category selected) -->
        @if (selectedCategory(); as activeCat) {
          <div class="space-y-4">
            <!-- Active Category Header & Reset Action -->
            <div class="p-4 rounded-xl bg-slate-950 border border-slate-800 flex items-center justify-between text-xs">
              <div class="flex items-center gap-3">
                <span class="text-slate-400">Selected Category:</span>
                <span [class]="getCategoryClass(activeCat)" class="text-xs font-bold px-2.5 py-1 rounded font-mono uppercase">
                  {{ activeCat }}
                </span>
                <span class="text-slate-500 hidden sm:inline">
                  (Default radius: <strong>{{ currentPreset?.defaultRadius }}m</strong> • Risk: <strong>{{ currentPreset?.defaultRisk }}</strong>)
                </span>
              </div>
              <button
                type="button"
                (click)="selectedCategory.set(null)"
                class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 font-medium border border-slate-700 transition-colors">
                ← Change Category
              </button>
            </div>

            <!-- 2-Column Registration Workspace -->
            <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 bg-slate-950/60 border border-slate-800 rounded-2xl p-6 shadow-xl">
              <!-- Left Column: Form (5 Cols) -->
              <form (ngSubmit)="saveEnvironmental()" class="lg:col-span-5 space-y-4 text-xs">
                <div class="border-b border-slate-800 pb-3">
                  <h3 class="text-base font-bold text-white">Configure Attractant Zone Parameters</h3>
                  <p class="text-[11px] text-slate-400">Click on the map on the right to position center coordinates.</p>
                </div>

                <div>
                  <label class="block text-slate-300 font-medium mb-1">Layer / Zone Name</label>
                  <input type="text" [(ngModel)]="newItem.name" name="name" required
                    class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-emerald-500">
                </div>

                <!-- Auto-populated from Map -->
                <div class="p-3.5 rounded-xl bg-slate-900/80 border border-slate-700/80 space-y-2">
                  <div class="flex items-center justify-between font-semibold text-slate-200">
                    <span>Center Coordinates & Buffer</span>
                    <span class="text-[10px] text-emerald-400 font-normal">Click map to set center</span>
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
                        class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white font-mono font-bold text-emerald-400">
                    </div>
                  </div>
                  <div>
                    <label class="text-[10px] text-slate-400">Location Notes</label>
                    <input type="text" [(ngModel)]="newItem.location_desc" name="location_desc"
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
                    <input type="text" [(ngModel)]="newItem.active_season" name="active_season" required
                      class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-emerald-500">
                  </div>
                </div>

                <div class="flex items-center justify-between pt-4 border-t border-slate-800">
                  <button type="button" (click)="selectedCategory.set(null)" class="px-4 py-2 rounded-lg bg-slate-800 text-slate-300">
                    Back to Categories
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
          </div>
        }
      }
    </div>
  `
})
export class EnvironmentalComponent implements OnInit {
  api = inject(AdminApiService);
  activeTab = signal<'history' | 'entry'>('history');
  selectedCategory = signal<AttractantCategory | null>(null);

  categoryPresets: CategoryPreset[] = [
    {
      category: 'CROP_PHENOLOGY',
      title: 'Sugarcane & Fruit Crop Ripening',
      badge: 'Thailand & APAC Core',
      badgeColor: 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30',
      description: 'High-sugar ripening attractants (sugarcane, bananas, jackfruit) triggering Asian elephant foraging incursions.',
      defaultRadius: 3000,
      defaultRisk: 0.85,
      defaultSeason: 'October - January (Harvest Peak)',
      defaultName: 'Sector Sugarcane Agro-Forest Interface',
      defaultDesc: 'Kaeng Krachan Eastern Agro-Forest Boundary',
      mapCenter: { lat: 12.870, lng: 99.660 },
      mapZoom: 12
    },
    {
      category: 'PERSIMMON_KAKI',
      title: 'Abandoned Persimmon (Kaki) Orchards',
      badge: 'Japan Validation Core',
      badgeColor: 'bg-blue-500/20 text-blue-400 border border-blue-500/30',
      description: 'Unharvested fruit clusters in aging rural valleys luring bears into residential communities ahead of hibernation.',
      defaultRadius: 1200,
      defaultRisk: 0.90,
      defaultSeason: 'September - November (Pre-hibernation)',
      defaultName: 'Nagano Valley Persimmon Hazard Cluster',
      defaultDesc: 'Abandoned residential hillside grove',
      mapCenter: { lat: 36.660, lng: 138.190 },
      mapZoom: 13
    },
    {
      category: 'LIVESTOCK_PASTURE',
      title: 'Livestock Grazing Corrals & Pastures',
      badge: 'Indonesia & India Core',
      badgeColor: 'bg-amber-500/20 text-amber-400 border border-amber-500/30',
      description: 'Unfenced night cattle/goat grazing areas attracting apex predators (Sumatran & Bengal tigers) to plantation borders.',
      defaultRadius: 2000,
      defaultRisk: 0.80,
      defaultSeason: 'Year-round (Nocturnal Ambush Risk)',
      defaultName: 'Siak Fringe Communal Cattle Pasture',
      defaultDesc: 'Open grazing pasture bordering oil palm blocks',
      mapCenter: { lat: 0.510, lng: 101.470 },
      mapZoom: 12
    },
    {
      category: 'WATER_SOURCE',
      title: 'Dry-Season Wildlife Watering Holes',
      badge: 'Multi-Theater Core',
      badgeColor: 'bg-purple-500/20 text-purple-400 border border-purple-500/30',
      description: 'Critical riparian corridors, irrigation canals, and isolated reservoirs during severe drought periods.',
      defaultRadius: 1000,
      defaultRisk: 0.70,
      defaultSeason: 'March - May (Dry Season Peak)',
      defaultName: 'Perimeter Irrigation Canal Watering Point',
      defaultDesc: 'Canal access point near village perimeter',
      mapCenter: { lat: 12.850, lng: 99.640 },
      mapZoom: 13
    }
  ];

  newItem: Partial<AdminEnvironmentalData> = {
    category: 'CROP_PHENOLOGY',
    name: '',
    risk_weight: 0.85,
    active_season: '',
    location_desc: ''
  };
  itemLat = 12.870;
  itemLng = 99.660;
  itemRadius = 3000;

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

  get currentPreset(): CategoryPreset | undefined {
    const cat = this.selectedCategory();
    return this.categoryPresets.find(p => p.category === cat);
  }

  ngOnInit() {
    this.api.getEnvironmental().subscribe();
  }

  startNewRegistration() {
    this.activeTab.set('entry');
    this.selectedCategory.set(null); // Always start at Step 1 (Poka-Yoke)
  }

  selectCategory(category: AttractantCategory) {
    this.selectedCategory.set(category);
    const preset = this.categoryPresets.find(p => p.category === category);
    if (preset) {
      this.newItem = {
        category,
        name: preset.defaultName,
        risk_weight: preset.defaultRisk,
        active_season: preset.defaultSeason,
        location_desc: preset.defaultDesc
      };
      this.itemRadius = preset.defaultRadius;
      this.itemLat = preset.mapCenter.lat;
      this.itemLng = preset.mapCenter.lng;
      this.mapCenter = { ...preset.mapCenter };
      this.mapZoom = preset.mapZoom;
    }
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
        this.selectedCategory.set(null);
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
