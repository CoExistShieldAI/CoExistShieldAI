import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AdminApiService } from '../../services/admin-api.service';
import { AdminWildlifeIncident, ThreatLevel, WildlifeStatus } from '../../models/admin.model';

@Component({
  selector: 'app-wildlife',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="p-8 space-y-6 max-w-7xl mx-auto">
      <!-- Title & Action Bar -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 class="text-2xl font-bold tracking-tight text-white">Wildlife Encounter & Sighting Registry</h2>
          <p class="text-sm text-slate-400">Monitor reported apex carnivore and megafauna incursions in real-time.</p>
        </div>
        <button
          (click)="openAddModal()"
          class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 font-semibold text-xs text-white shadow transition-all self-start sm:self-auto">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
          + Log Wildlife Report
        </button>
      </div>

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
    </div>

    <!-- Modal: Log Wildlife Sighting -->
    @if (showAddModal()) {
      <div class="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 z-50">
        <div class="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-4">
          <div class="flex items-center justify-between border-b border-slate-800 pb-3">
            <h3 class="text-base font-bold text-white">Log Wildlife Sighting or Encounter</h3>
            <button (click)="showAddModal.set(false)" class="text-slate-400 hover:text-white">✕</button>
          </div>

          <form (ngSubmit)="saveIncident()" class="space-y-4 text-xs">
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-300 font-medium mb-1">Target Species</label>
                <select [(ngModel)]="newItem.species" name="species"
                  class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white">
                  <option value="Asian Elephant">Asian Elephant</option>
                  <option value="Sumatran Tiger">Sumatran Tiger</option>
                  <option value="Bengal Tiger">Bengal Tiger</option>
                  <option value="Brown Bear">Brown Bear</option>
                  <option value="Asian Black Bear">Asian Black Bear</option>
                </select>
              </div>
              <div>
                <label class="block text-slate-300 font-medium mb-1">Threat Level</label>
                <select [(ngModel)]="newItem.threat_level" name="threat_level"
                  class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white">
                  <option value="TIER_1_CAUTION">Tier 1 - Caution</option>
                  <option value="TIER_2_ALERT">Tier 2 - Alert</option>
                  <option value="TIER_3_CRITICAL">Tier 3 - Critical Danger</option>
                </select>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-300 font-medium mb-1">Latitude</label>
                <input type="number" step="0.0001" [(ngModel)]="itemLat" name="itemLat"
                  class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white">
              </div>
              <div>
                <label class="block text-slate-300 font-medium mb-1">Longitude</label>
                <input type="number" step="0.0001" [(ngModel)]="itemLng" name="itemLng"
                  class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white">
              </div>
            </div>

            <div>
              <label class="block text-slate-300 font-medium mb-1">Road / Region Description</label>
              <input type="text" [(ngModel)]="itemRoad" name="itemRoad" placeholder="e.g. Rural Route 3219 KM 14"
                class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white">
            </div>

            <div>
              <label class="block text-slate-300 font-medium mb-1">Reporter / Ranger Name</label>
              <input type="text" [(ngModel)]="newItem.reporter" name="reporter" placeholder="e.g. Officer Somchai (Field Unit)"
                class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white">
            </div>

            <div>
              <label class="block text-slate-300 font-medium mb-1">Description / Observations</label>
              <textarea rows="3" [(ngModel)]="newItem.description" name="description" placeholder="Observation details, herd size, direction of travel..."
                class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white"></textarea>
            </div>

            <div class="flex justify-end gap-3 pt-3 border-t border-slate-800">
              <button type="button" (click)="showAddModal.set(false)" class="px-4 py-2 rounded-lg bg-slate-800 text-slate-300">
                Cancel
              </button>
              <button type="submit" class="px-5 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 font-semibold text-white shadow">
                Save Incident
              </button>
            </div>
          </form>
        </div>
      </div>
    }
  `
})
export class WildlifeComponent implements OnInit {
  api = inject(AdminApiService);
  searchQuery = '';
  filterSpecies = '';
  filterTier = '';
  showAddModal = signal<boolean>(false);

  newItem: Partial<AdminWildlifeIncident> = {
    species: 'Asian Elephant',
    threat_level: 'TIER_3_CRITICAL',
    reporter: 'Admin Console',
    description: ''
  };
  itemLat = 12.865;
  itemLng = 99.655;
  itemRoad = 'Rural Highway 3219';

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

  openAddModal() {
    this.showAddModal.set(true);
  }

  saveIncident() {
    const payload: Partial<AdminWildlifeIncident> = {
      ...this.newItem,
      location: {
        lat: Number(this.itemLat),
        lng: Number(this.itemLng),
        road_name: this.itemRoad
      },
      status: 'VERIFIED'
    };
    this.api.createWildlife(payload).subscribe({
      next: () => {
        this.showAddModal.set(false);
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
