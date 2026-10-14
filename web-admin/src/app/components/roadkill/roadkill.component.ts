import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { GoogleMapsModule } from '@angular/google-maps';
import { AdminApiService } from '../../services/admin-api.service';
import { AdminRoadkillRecord, DecayStage, RoadkillStatus } from '../../models/admin.model';

@Component({
  selector: 'app-roadkill',
  standalone: true,
  imports: [CommonModule, FormsModule, GoogleMapsModule],
  template: `
    <div class="p-8 space-y-6 max-w-7xl mx-auto">
      <!-- Title & Tab Switcher -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h2 class="text-2xl font-bold tracking-tight text-white">Roadkill Surveillance & Disposal Console</h2>
          <p class="text-sm text-slate-400">Prevent carcass scent plumes from transforming roads into apex predator hunting corridors.</p>
        </div>

        <div class="flex items-center p-1 rounded-xl bg-slate-950 border border-slate-800 text-xs">
          <button
            (click)="activeTab.set('history')"
            [class.bg-slate-800]="activeTab() === 'history'"
            [class.text-white]="activeTab() === 'history'"
            class="px-4 py-2 rounded-lg font-semibold text-slate-400 hover:text-white transition-all">
            History & Data Cleansing
          </button>
          <button
            (click)="activeTab.set('entry')"
            [class.bg-emerald-600]="activeTab() === 'entry'"
            [class.text-white]="activeTab() === 'entry'"
            class="px-4 py-2 rounded-lg font-semibold text-slate-400 hover:text-white transition-all flex items-center gap-1.5">
            <span>+ Desktop Big-Screen Entry</span>
          </button>
        </div>
      </div>

      <!-- Tab 1: History & Data Cleansing -->
      @if (activeTab() === 'history') {
        <!-- Filter Bar -->
        <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800 flex flex-wrap items-center justify-between gap-4 text-xs">
          <div class="flex items-center gap-3">
            <label class="flex items-center gap-2 cursor-pointer text-slate-300">
              <input type="checkbox" [(ngModel)]="showExcluded" (change)="loadRoadkills()"
                class="rounded bg-slate-800 border-slate-700 text-emerald-500 focus:ring-0">
              <span>Show Low-Credibility / Excluded Reports</span>
            </label>
            <span class="text-slate-600">|</span>
            <span class="text-slate-400">
              Active Predator Magnets: <strong class="text-rose-400">{{ pendingCount() }} Pending Removal</strong>
            </span>
          </div>

          <div class="text-[11px] text-slate-500">
            * Excluding noise prevents corrupting AI scent-plume dispersion models.
          </div>
        </div>

        <!-- Roadkill Table -->
        <div class="rounded-2xl bg-slate-950/60 border border-slate-800 overflow-hidden shadow-sm">
          <div class="overflow-x-auto">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-900/80 text-slate-400 uppercase tracking-wider font-semibold border-b border-slate-800">
                <tr>
                  <th class="py-3 px-4">Reported</th>
                  <th class="py-3 px-4">Species & Decay Stage</th>
                  <th class="py-3 px-4">Location (Road & GPS)</th>
                  <th class="py-3 px-4">Reliability Score</th>
                  <th class="py-3 px-4">Carcass Status</th>
                  <th class="py-3 px-4">AI Ingestion</th>
                  <th class="py-3 px-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-800/60">
                @for (rk of api.roadkills(); track rk.id) {
                  <tr class="hover:bg-slate-900/40 transition-colors" [class.opacity-50]="rk.is_excluded">
                    <td class="py-3.5 px-4 text-slate-400 font-mono whitespace-nowrap">
                      {{ rk.reported_at | date:'MM-dd HH:mm' }}
                    </td>
                    <td class="py-3.5 px-4 text-white">
                      <div class="font-bold">{{ rk.species }}</div>
                      <div class="text-[11px] text-slate-400 flex items-center gap-1.5 mt-0.5">
                        <span>~{{ rk.estimated_weight_kg }}kg</span>
                        <span>•</span>
                        <span [class]="getDecayClass(rk.decay_stage)" class="font-semibold">
                          {{ rk.decay_stage }}
                        </span>
                      </div>
                    </td>
                    <td class="py-3.5 px-4 text-slate-300">
                      <div class="font-medium">{{ rk.location.road_name || 'Highway Corridor' }}</div>
                      <div class="text-[11px] text-slate-500 font-mono">{{ rk.location.lat | number:'1.3-3' }}, {{ rk.location.lng | number:'1.3-3' }}</div>
                    </td>
                    <td class="py-3.5 px-4 whitespace-nowrap">
                      <div class="flex items-center gap-2">
                        <div class="w-16 h-2 bg-slate-800 rounded-full overflow-hidden">
                          <div class="h-full rounded-full"
                            [style.width.%]="rk.reliability_score * 100"
                            [class.bg-emerald-500]="rk.reliability_score >= 0.8"
                            [class.bg-amber-500]="rk.reliability_score < 0.8 && rk.reliability_score >= 0.5"
                            [class.bg-rose-500]="rk.reliability_score < 0.5">
                          </div>
                        </div>
                        <span class="font-mono text-[11px]" [class.text-rose-400]="rk.reliability_score < 0.5">
                          {{ (rk.reliability_score * 100) | number:'1.0-0' }}%
                        </span>
                      </div>
                    </td>
                    <td class="py-3.5 px-4 whitespace-nowrap">
                      @if (rk.status === 'PENDING_REMOVAL') {
                        <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded bg-rose-500/20 text-rose-300 border border-rose-500/30 text-[10px] font-bold animate-pulse">
                          ⚠️ PENDING REMOVAL
                        </span>
                      } @else {
                        <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 text-[10px] font-semibold">
                          ✓ DISPOSED
                        </span>
                      }
                    </td>
                    <td class="py-3.5 px-4 whitespace-nowrap">
                      @if (rk.is_excluded) {
                        <span class="px-2 py-0.5 rounded bg-slate-800 text-slate-400 text-[10px] font-mono border border-slate-700">
                          EXCLUDED
                        </span>
                      } @else {
                        <span class="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 text-[10px] font-mono border border-blue-500/20">
                          ACTIVE IN AI
                        </span>
                      }
                    </td>
                    <td class="py-3.5 px-4 text-right whitespace-nowrap space-x-2">
                      @if (rk.status === 'PENDING_REMOVAL') {
                        <button (click)="markDisposed(rk.id)" class="px-2.5 py-1 rounded bg-emerald-600/30 text-emerald-300 hover:bg-emerald-600/50 font-medium">
                          Mark Disposed
                        </button>
                      }
                      <button (click)="toggleExclusion(rk.id, !rk.is_excluded)"
                        class="px-2.5 py-1 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 font-medium border border-slate-700">
                        {{ rk.is_excluded ? 'Include in AI' : 'Exclude (Noise)' }}
                      </button>
                    </td>
                  </tr>
                }
              </tbody>
            </table>
          </div>
        </div>
      }

      <!-- Tab 2: Desktop Big-Screen Entry -->
      @if (activeTab() === 'entry') {
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 bg-slate-950/60 border border-slate-800 rounded-2xl p-6 shadow-xl">
          <!-- Left Column: Form (5 Cols) -->
          <form (ngSubmit)="submitRoadkillEntry()" class="lg:col-span-5 space-y-4 text-xs">
            <div class="border-b border-slate-800 pb-3">
              <h3 class="text-base font-bold text-white">Log Carcass via Desktop Station</h3>
              <p class="text-[11px] text-slate-400">Click on the map on the right to pinpoint precise GPS location coordinates.</p>
            </div>

            <div>
              <label class="block text-slate-300 font-medium mb-1">Carcass Species Category</label>
              <select [(ngModel)]="newRecord.species" name="species" required
                class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-emerald-500">
                <option value="Wild Boar">Wild Boar (Sus scrofa) - High Tiger Magnet</option>
                <option value="Sambar Deer">Sambar Deer (Rusa unicolor) - High Tiger/Leopard Magnet</option>
                <option value="Macaque / Monkey">Macaque / Primate - Pathogen Vector</option>
                <option value="Stray Cattle / Goat">Domestic Livestock - Ambush Point</option>
                <option value="Raccoon Dog / Tanuki">Raccoon Dog / Small Mammal (Japan)</option>
                <option value="Other / Unknown">Other / Unidentified</option>
              </select>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-300 font-medium mb-1">Estimated Weight (kg)</label>
                <input type="number" [(ngModel)]="newRecord.estimated_weight_kg" name="weight" required min="1"
                  class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-emerald-500">
              </div>
              <div>
                <label class="block text-slate-300 font-medium mb-1">Decay Stage (Scent Risk)</label>
                <select [(ngModel)]="newRecord.decay_stage" name="decay" required
                  class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-emerald-500">
                  <option value="FRESH">Fresh (&lt;6 hours) - Critical Predator Attractant</option>
                  <option value="MODERATE">Moderate (6 - 24 hours) - High Plume</option>
                  <option value="ADVANCED">Advanced (&gt;24 hours) - Scavenger Zone</option>
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
                  <input type="number" step="0.00001" [(ngModel)]="pinLat" name="lat"
                    class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white font-mono">
                </div>
                <div>
                  <label class="text-[10px] text-slate-400">Longitude</label>
                  <input type="number" step="0.00001" [(ngModel)]="pinLng" name="lng"
                    class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white font-mono">
                </div>
              </div>
              <div>
                <label class="text-[10px] text-slate-400">Road / Sector Name</label>
                <input type="text" [(ngModel)]="roadName" name="roadName" placeholder="e.g. Route 3219 KM 16.5"
                  class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white">
              </div>
            </div>

            <!-- Photo Attachment -->
            <div>
              <label class="block text-slate-300 font-medium mb-1">Photo Attachment / Evidence URL</label>
              <input type="text" [(ngModel)]="newRecord.photo_url" name="photo" placeholder="https://storage.googleapis.com/..."
                class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white placeholder-slate-500">
            </div>

            <div class="flex items-center justify-between pt-4 border-t border-slate-800">
              <button type="button" (click)="activeTab.set('history')" class="px-4 py-2 rounded-lg bg-slate-800 text-slate-300">
                Cancel
              </button>
              <button type="submit" class="px-6 py-2.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 font-bold text-white shadow-lg shadow-emerald-600/20">
                Register Roadkill & Alert Rangers
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
              📍 Click anywhere on map to position carcass pin
            </div>
          </div>
        </div>
      }
    </div>
  `
})
export class RoadkillComponent implements OnInit {
  api = inject(AdminApiService);
  activeTab = signal<'history' | 'entry'>('history');
  showExcluded = true;

  // Desktop Entry Form
  newRecord: Partial<AdminRoadkillRecord> = {
    species: 'Wild Boar',
    estimated_weight_kg: 50,
    decay_stage: 'FRESH',
    photo_url: ''
  };
  pinLat = 12.842;
  pinLng = 99.680;
  roadName = 'Rural Route 3219, KM 14';

  mapCenter: google.maps.LatLngLiteral = { lat: 12.842, lng: 99.680 };
  mapZoom = 12;
  mapOptions: google.maps.MapOptions = {
    backgroundColor: '#0f172a',
    streetViewControl: false
  };
  markerOptions: google.maps.MarkerOptions = {
    draggable: true,
    title: 'Carcass Location'
  };

  ngOnInit() {
    this.loadRoadkills();
  }

  loadRoadkills() {
    this.api.getRoadkills(this.showExcluded).subscribe();
  }

  pendingCount(): number {
    return this.api.roadkills().filter(r => r.status === 'PENDING_REMOVAL').length;
  }

  toggleExclusion(id: string, isExcluded: boolean) {
    this.api.toggleRoadkillExclusion(id, isExcluded).subscribe();
  }

  markDisposed(id: string) {
    this.api.markRoadkillDisposed(id).subscribe();
  }

  onMapClick(event: google.maps.MapMouseEvent) {
    if (event.latLng) {
      this.pinLat = Number(event.latLng.lat().toFixed(5));
      this.pinLng = Number(event.latLng.lng().toFixed(5));
      this.roadName = `Geocoded Spot: Lat ${this.pinLat}, Lng ${this.pinLng}`;
    }
  }

  submitRoadkillEntry() {
    const payload: Partial<AdminRoadkillRecord> = {
      ...this.newRecord,
      location: {
        lat: this.pinLat,
        lng: this.pinLng,
        road_name: this.roadName
      },
      status: 'PENDING_REMOVAL',
      reliability_score: 0.95,
      is_excluded: false
    };

    this.api.createRoadkill(payload).subscribe({
      next: () => {
        this.activeTab.set('history');
      }
    });
  }

  getDecayClass(stage: DecayStage): string {
    if (stage === 'FRESH') return 'text-rose-400';
    if (stage === 'MODERATE') return 'text-amber-400';
    return 'text-slate-400';
  }
}
