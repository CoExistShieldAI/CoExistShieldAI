import { Component, inject } from '@angular/core';
import { AdminApiService } from '../../services/admin-api.service';

@Component({
  selector: 'app-header',
  standalone: true,
  template: `
    <header class="h-16 bg-slate-900/90 backdrop-blur border-b border-slate-800 px-6 flex items-center justify-between sticky top-0 z-40">
      <!-- Theater Context -->
      <div class="flex items-center gap-3">
        <div class="px-2.5 py-1 rounded bg-slate-800 border border-slate-700 text-xs font-medium text-slate-300 flex items-center gap-2">
          <span class="w-2 h-2 rounded-full bg-blue-400"></span>
          <span>JAPAC Operational Theater</span>
        </div>
        <span class="text-xs text-slate-500 hidden md:inline">Covering Japan, Thailand, Indonesia, India & Nepal</span>
      </div>

      <!-- Actions & Profile -->
      <div class="flex items-center gap-4">
        <!-- AI Recalculation Trigger Button -->
        <button
          (click)="triggerAiRecalculation()"
          [disabled]="api.isRecalculating()"
          class="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-emerald-600 hover:bg-emerald-500 active:bg-emerald-700 text-white shadow-sm shadow-emerald-600/30 transition-all disabled:opacity-50 disabled:cursor-not-allowed">
          @if (api.isRecalculating()) {
            <svg class="animate-spin -ml-0.5 mr-1.5 h-3.5 w-3.5 text-white" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
            </svg>
            <span>Running AI Inference...</span>
          } @else {
            <svg class="w-3.5 h-3.5 text-emerald-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            <span>Run AI Geofence Engine</span>
          }
        </button>

        <!-- Admin Profile Pill -->
        <div class="flex items-center gap-2.5 pl-3 border-l border-slate-800">
          <div class="w-8 h-8 rounded-full bg-slate-800 border border-slate-700 flex items-center justify-center font-bold text-xs text-emerald-400">
            SA
          </div>
          <div class="text-left hidden lg:block">
            <div class="text-xs font-medium text-slate-200">System Admin</div>
            <div class="text-[10px] text-emerald-400 font-mono">ROLE: SUPERADMIN</div>
          </div>
        </div>
      </div>
    </header>
  `
})
export class HeaderComponent {
  api = inject(AdminApiService);

  triggerAiRecalculation() {
    this.api.recalculateAiGeofences().subscribe({
      next: (res) => {
        // Notification toast feedback
        console.log('AI Recalculation complete:', res);
      }
    });
  }
}
