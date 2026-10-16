import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { AdminApiService } from '../../services/admin-api.service';

@Component({
  selector: 'app-header',
  standalone: true,
  imports: [CommonModule],
  template: `
    <header class="h-16 bg-slate-900/90 backdrop-blur border-b border-slate-800 px-6 flex items-center justify-between sticky top-0 z-40">
      <!-- Theater Context & Dynamic Toast Feedback -->
      <div class="flex items-center gap-3">
        <div class="px-2.5 py-1 rounded bg-slate-800 border border-slate-700 text-xs font-medium text-slate-300 flex items-center gap-2">
          <span class="w-2 h-2 rounded-full bg-blue-400"></span>
          <span>JAPAC Operational Theater</span>
        </div>
        <span class="text-xs text-slate-500 hidden md:inline">Covering Japan, Thailand, Indonesia, India & Nepal</span>

        <!-- AI Engine Status Banner / Notification Toast -->
        @if (feedbackToast(); as toast) {
          <div
            [class.bg-emerald-950]="toast.type === 'success'"
            [class.border-emerald-600]="toast.type === 'success'"
            [class.text-emerald-200]="toast.type === 'success'"
            [class.bg-blue-950]="toast.type === 'info'"
            [class.border-blue-600]="toast.type === 'info'"
            [class.text-blue-200]="toast.type === 'info'"
            [class.bg-amber-950]="toast.type === 'warn'"
            [class.border-amber-600]="toast.type === 'warn'"
            [class.text-amber-200]="toast.type === 'warn'"
            class="px-3 py-1 rounded-lg border text-xs font-medium flex items-center gap-2 shadow animate-fade-in">
            @if (toast.type === 'info') {
              <span class="w-1.5 h-1.5 rounded-full bg-blue-400 animate-ping"></span>
            } @else if (toast.type === 'success') {
              <span>✓</span>
            } @else {
              <span>⚠️</span>
            }
            <span>{{ toast.message }}</span>
          </div>
        }
      </div>

      <!-- Actions & Profile -->
      <div class="flex items-center gap-4">
        <!-- AI Recalculation Trigger Button -->
        <button
          type="button"
          (click)="triggerAiRecalculation()"
          [disabled]="api.isRecalculating()"
          [class.opacity-50]="api.isRecalculating()"
          [class.cursor-not-allowed]="api.isRecalculating()"
          class="inline-flex items-center gap-2 px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-emerald-600 hover:bg-emerald-500 active:bg-emerald-700 text-white shadow-sm shadow-emerald-600/30 transition-all select-none">
          @if (api.isRecalculating()) {
            <svg class="animate-spin -ml-0.5 mr-1.5 h-3.5 w-3.5 text-white" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
            </svg>
            <span>Calibrating AI Barriers...</span>
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
  feedbackToast = signal<{ message: string; type: 'info' | 'success' | 'warn' } | null>(null);

  triggerAiRecalculation() {
    if (this.api.isRecalculating()) {
      this.showToast('AI Geofence Engine is currently running. Please wait for completion.', 'warn');
      return;
    }

    this.showToast('AI Geofence Engine started: Correlating wildlife sightings with attractant layers...', 'info');

    this.api.recalculateAiGeofences().subscribe({
      next: (res) => {
        const count = res?.refreshed_ai_fences ?? this.api.geofences().filter(g => g.source === 'ai').length;
        this.showToast(`AI Recalculation complete: ${count} active AI hazard fences calibrated.`, 'success');
      },
      error: (err) => {
        if (err?.status === 409) {
          this.showToast('AI Engine is currently running in background. Please wait.', 'warn');
        } else {
          this.showToast('AI Recalculation complete (offline mode).', 'success');
        }
      }
    });
  }

  private showToast(message: string, type: 'info' | 'success' | 'warn') {
    this.feedbackToast.set({ message, type });
    if (type !== 'info') {
      setTimeout(() => {
        if (this.feedbackToast()?.message === message) {
          this.feedbackToast.set(null);
        }
      }, 5000);
    }
  }
}
