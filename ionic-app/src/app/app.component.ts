import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ShieldStateService } from './services/shield-state.service';
import { HudPanelComponent } from './components/hud-panel/hud-panel.component';
import { GisMapPanelComponent } from './components/gis-map-panel/gis-map-panel.component';
import { ReportPanelComponent } from './components/report-panel/report-panel.component';
import { ConsolePanelComponent } from './components/console-panel/console-panel.component';
import { AuthModalComponent } from './components/auth-modal/auth-modal.component';
import { AlertTier, UserRole } from './models/shield.model';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    CommonModule,
    HudPanelComponent,
    GisMapPanelComponent,
    ReportPanelComponent,
    ConsolePanelComponent,
    AuthModalComponent
  ],
  template: `
    <div class="fixed inset-0 w-full h-full flex flex-col bg-slate-950 text-slate-100 overflow-hidden select-none">
      <!-- 1. Interactive Demo & Judge Control Bar -->
      <header class="bg-slate-900 border-b border-slate-800 px-3 py-1.5 flex flex-wrap items-center justify-between gap-2 z-30 flex-shrink-0">
          <div class="flex items-center gap-2 overflow-x-auto">
            <button
              type="button"
              (click)="state.showDemoControls.set(!state.showDemoControls())"
              class="px-2.5 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-amber-400 text-[11px] font-extrabold whitespace-nowrap">
              🛠 DEMO {{ state.showDemoControls() ? '▲' : '▼' }}
            </button>

            @if (state.showDemoControls()) {
              <!-- Role Switcher -->
              <div class="flex items-center gap-1 bg-slate-950 px-2 py-0.5 rounded-lg border border-slate-800">
                <span class="text-[10px] font-bold text-slate-400 mr-1">ROLE:</span>
                @for (r of roles; track r) {
                  <button
                    type="button"
                    (click)="state.setRole(r)"
                    class="px-2 py-0.5 rounded text-[10px] font-bold transition-colors"
                    [ngClass]="
                      state.currentRole() === r
                        ? 'bg-sky-600 text-white'
                        : 'text-slate-400 hover:text-white'
                    ">
                    {{ r }}
                  </button>
                }
              </div>

              <!-- Hazard Tier Simulator -->
              <div class="flex items-center gap-1 bg-slate-950 px-2 py-0.5 rounded-lg border border-slate-800">
                <span class="text-[10px] font-bold text-slate-400 mr-1">TIER:</span>
                @for (t of tiers; track t.id) {
                  <button
                    type="button"
                    (click)="state.setTier(t.id)"
                    class="px-2 py-0.5 rounded text-[10px] font-bold transition-colors whitespace-nowrap"
                    [ngClass]="
                      state.currentTier() === t.id
                        ? t.activeClass
                        : 'text-slate-400 hover:text-white'
                    ">
                    {{ t.label }}
                  </button>
                }
              </div>
            }
          </div>

          <!-- Right Controls: Auth Session & Architecture Comparison Links -->
          <div class="flex items-center gap-2 text-[11px]">
            @if (state.authUser(); as user) {
              <button
                type="button"
                (click)="state.signOutUser()"
                class="px-2.5 py-0.5 rounded bg-slate-800 hover:bg-slate-700 text-amber-300 border border-slate-700 font-bold transition-colors">
                Sign Out ({{ user.email }})
              </button>
            } @else {
              <button
                type="button"
                (click)="state.showAuthModal.set(true)"
                class="px-2.5 py-0.5 rounded bg-sky-600 hover:bg-sky-500 text-white font-bold transition-colors">
                🔐 Sign In
              </button>
            }
            <div class="hidden xl:flex items-center gap-2">
              <span class="px-2 py-0.5 rounded bg-emerald-950/80 border border-emerald-700 text-emerald-300 font-semibold">
                ⚡ Angular + Ionic + Capacitor
              </span>
              <a
                href="https://corald-kmp.web.app/"
                target="_blank"
                rel="noopener noreferrer"
                class="px-2 py-0.5 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 font-medium transition-colors">
                🔗 KMP Edition (corald-kmp)
              </a>
              <a
                href="https://corald-admin.web.app/"
                target="_blank"
                rel="noopener noreferrer"
                class="px-2 py-0.5 rounded bg-slate-800 hover:bg-slate-700 text-sky-300 border border-slate-700 font-medium transition-colors">
                🔗 Admin Console
              </a>
            </div>
          </div>
        </header>

        <!-- 2. Main Responsive Workspace -->
        <!-- Wide Screen (lg+): Left Pane (HUD / Report / Console) + Right Pane (Google Maps) -->
        <main class="flex-1 min-h-0 flex flex-col lg:flex-row overflow-hidden relative">
          <!-- LEFT PANE: HUD / Report / Console (Always visible on lg+, visible on mobile when not on MAP tab) -->
          <section
            class="w-full flex-1 lg:flex-initial lg:w-[440px] xl:w-[470px] lg:flex-shrink-0 h-full min-h-0 flex flex-col border-r border-slate-800 bg-slate-950 z-10 overflow-hidden"
            [ngClass]="{
              'hidden lg:flex': state.mobileActiveView() === 'MAP',
              'flex': state.mobileActiveView() !== 'MAP'
            }">
            <!-- Desktop Left-Pane Mode Switcher -->
            <div class="hidden lg:flex items-center gap-1 p-2 bg-slate-900/80 border-b border-slate-800 flex-shrink-0">
              <button
                type="button"
                (click)="state.activeLeftTab.set('HUD')"
                class="flex-1 py-1.5 px-3 rounded-lg text-xs font-bold flex items-center justify-center gap-1.5 transition-all"
                [ngClass]="
                  state.activeLeftTab() === 'HUD'
                    ? 'bg-sky-600 text-white shadow'
                    : 'text-slate-400 hover:text-white hover:bg-slate-800'
                ">
                <span>🛡️</span>
                <span>Alert HUD</span>
              </button>
              <button
                type="button"
                (click)="state.activeLeftTab.set('REPORT')"
                class="flex-1 py-1.5 px-3 rounded-lg text-xs font-bold flex items-center justify-center gap-1.5 transition-all"
                [ngClass]="
                  state.activeLeftTab() === 'REPORT'
                    ? 'bg-sky-600 text-white shadow'
                    : 'text-slate-400 hover:text-white hover:bg-slate-800'
                ">
                <span>📸</span>
                <span>Field Report</span>
              </button>
              @if (state.currentRole() !== 'CITIZEN') {
                <button
                  type="button"
                  (click)="state.activeLeftTab.set('CONSOLE')"
                  class="flex-1 py-1.5 px-3 rounded-lg text-xs font-bold flex items-center justify-center gap-1.5 transition-all"
                  [ngClass]="
                    state.activeLeftTab() === 'CONSOLE'
                      ? 'bg-sky-600 text-white shadow'
                      : 'text-slate-400 hover:text-white hover:bg-slate-800'
                  ">
                  <span>⚙️</span>
                  <span>Console</span>
                </button>
              }
            </div>

            <!-- Active Left Panel Content -->
            <div class="flex-1 min-h-0 overflow-hidden">
              @switch (state.activeLeftTab()) {
                @case ('HUD') {
                  <app-hud-panel></app-hud-panel>
                }
                @case ('REPORT') {
                  <app-report-panel></app-report-panel>
                }
                @case ('CONSOLE') {
                  <app-console-panel></app-console-panel>
                }
              }
            </div>
          </section>

          <!-- RIGHT PANE: Full Interactive Google Maps GIS Canvas -->
          <section
            class="flex-1 h-full min-h-0 relative overflow-hidden"
            [ngClass]="{
              'hidden lg:block': state.mobileActiveView() !== 'MAP',
              'block': state.mobileActiveView() === 'MAP'
            }">
            <app-gis-map-panel class="block w-full h-full"></app-gis-map-panel>
          </section>
        </main>

        <!-- 3. Mobile Bottom Tab Bar (Visible only on screens < lg) -->
        <nav
          style="padding-bottom: max(0.375rem, env(safe-area-inset-bottom));"
          class="lg:hidden bg-slate-900/95 backdrop-blur border-t border-slate-700 px-2 pt-1.5 flex items-center justify-around z-40 flex-shrink-0 shadow-[0_-4px_16px_rgba(0,0,0,0.6)]">
          <button
            type="button"
            (click)="state.activeLeftTab.set('HUD'); state.mobileActiveView.set('SPLIT_OR_HUD')"
            class="flex-1 flex flex-col items-center py-1 px-2 rounded-xl text-[11px] font-bold transition-colors"
            [ngClass]="
              state.mobileActiveView() !== 'MAP' && state.activeLeftTab() === 'HUD'
                ? 'text-sky-400 bg-sky-950/50'
                : 'text-slate-400'
            ">
            <span class="text-base leading-none">🛡️</span>
            <span class="mt-0.5">Alert HUD</span>
          </button>

          <button
            type="button"
            (click)="state.mobileActiveView.set('MAP')"
            class="flex-1 flex flex-col items-center py-1 px-2 rounded-xl text-[11px] font-bold transition-colors"
            [ngClass]="
              state.mobileActiveView() === 'MAP'
                ? 'text-sky-400 bg-sky-950/50'
                : 'text-slate-400'
            ">
            <span class="text-base leading-none">🗺️</span>
            <span class="mt-0.5">GIS Map</span>
          </button>

          <button
            type="button"
            (click)="state.activeLeftTab.set('REPORT'); state.mobileActiveView.set('REPORT')"
            class="flex-1 flex flex-col items-center py-1 px-2 rounded-xl text-[11px] font-bold transition-colors"
            [ngClass]="
              state.mobileActiveView() !== 'MAP' && state.activeLeftTab() === 'REPORT'
                ? 'text-sky-400 bg-sky-950/50'
                : 'text-slate-400'
            ">
            <span class="text-base leading-none">📸</span>
            <span class="mt-0.5">Report</span>
          </button>

          @if (state.currentRole() !== 'CITIZEN') {
            <button
              type="button"
              (click)="state.activeLeftTab.set('CONSOLE'); state.mobileActiveView.set('CONSOLE')"
              class="flex-1 flex flex-col items-center py-1 px-2 rounded-xl text-[11px] font-bold transition-colors"
              [ngClass]="
                state.mobileActiveView() !== 'MAP' && state.activeLeftTab() === 'CONSOLE'
                  ? 'text-sky-400 bg-sky-950/50'
                  : 'text-slate-400'
              ">
              <span class="text-base leading-none">⚙️</span>
              <span class="mt-0.5">Console</span>
            </button>
          }
        </nav>

        <!-- 4. Firebase Auth Modal -->
        @if (state.showAuthModal()) {
          <app-auth-modal></app-auth-modal>
        }

        <!-- 5. Global Toast Notification Banner -->
        @if (state.toastMessage(); as msg) {
          <div class="fixed bottom-20 lg:bottom-6 left-1/2 -translate-x-1/2 z-50 px-4 py-2.5 rounded-xl bg-slate-900/95 backdrop-blur border border-sky-500/80 text-white text-xs font-bold shadow-2xl flex items-center gap-2 max-w-md">
            <span>{{ msg }}</span>
            <button
              type="button"
              (click)="state.toastMessage.set(null)"
              class="text-slate-400 hover:text-white ml-2">
              ✕
            </button>
          </div>
        }
    </div>
  `
})
export class AppComponent {
  state = inject(ShieldStateService);

  readonly roles: UserRole[] = ['CITIZEN', 'OFFICIAL', 'ADMIN'];

  readonly tiers: { id: AlertTier; label: string; activeClass: string }[] = [
    { id: 'SAFE', label: 'SAFE', activeClass: 'bg-emerald-600 text-white' },
    { id: 'TIER_1_NOTICE', label: 'T1 NOTICE', activeClass: 'bg-sky-600 text-white' },
    { id: 'TIER_2_WARNING', label: 'T2 WARNING', activeClass: 'bg-amber-600 text-white' },
    { id: 'TIER_3_DANGER', label: 'T3 DANGER', activeClass: 'bg-red-600 text-white' }
  ];
}
