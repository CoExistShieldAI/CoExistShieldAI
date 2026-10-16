import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ShieldStateService } from '../../services/shield-state.service';
import { HudIntelSection, WildlifeSpecies } from '../../models/shield.model';

@Component({
  selector: 'app-hud-panel',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="h-full overflow-y-auto px-4 py-4 space-y-4 bg-slate-950 text-slate-100">
      <!-- Top Status Header -->
      <div class="flex items-center justify-between bg-slate-900/90 border border-slate-800 rounded-xl px-3.5 py-2.5">
        <div class="flex items-center gap-2">
          <span class="text-base font-extrabold tracking-tight text-white">CoExist Shield AI</span>
          <span
            class="text-[10px] font-bold px-2 py-0.5 rounded-full uppercase tracking-wider"
            [ngClass]="{
              'bg-slate-800 text-slate-300 border border-slate-700': state.currentRole() === 'CITIZEN',
              'bg-blue-950 text-blue-300 border border-blue-700': state.currentRole() === 'OFFICIAL',
              'bg-pink-950 text-pink-300 border border-pink-700': state.currentRole() === 'ADMIN'
            }">
            {{ state.currentRole() }}
          </span>
        </div>

        <button
          type="button"
          (click)="state.toggleOfflineMode()"
          class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-slate-950 border border-slate-800 hover:border-slate-700 text-xs font-medium transition-colors">
          <span
            class="w-2 h-2 rounded-full"
            [ngClass]="state.isOfflineMode() ? 'bg-amber-400' : 'bg-emerald-400 animate-pulse'">
          </span>
          <span class="text-slate-300">{{ state.isOfflineMode() ? 'Offline (Cached)' : 'Live' }}</span>
        </button>
      </div>

      <!-- Capacitor Native Geofence Bridge Status Strip -->
      <div class="flex items-center justify-between px-3 py-1.5 rounded-lg bg-slate-900/60 border border-slate-800/80 text-[11px] text-slate-400">
        <div class="flex items-center gap-1.5">
          <span class="text-sky-400 font-semibold">⚡ GeofenceBridge:</span>
          <span class="font-mono text-slate-300">{{ state.nativeBridgePlatform() }}</span>
        </div>
        <div class="font-mono text-[10px] text-emerald-400">
          {{ state.registeredQuotaCount() }}/{{ state.maxGeofenceQuota }} OS Zones Active
        </div>
      </div>

      <!-- 1. Primary Graphical Status Display (Tactical Threat Orb & Proximity Gauge) -->
      @if (state.hudPresentation(); as hud) {
        <div class="flex flex-col items-center py-2">
          <div class="relative w-56 h-56 flex items-center justify-center">
            <!-- Ambient Radial Glow -->
            <div
              class="absolute inset-3 rounded-full blur-xl opacity-30 transition-all duration-500"
              [style.backgroundColor]="hud.tierColorHex">
            </div>

            <!-- SVG Tactical Gauge & Radar Ring -->
            <svg class="w-full h-full -rotate-90" viewBox="0 0 220 220">
              <!-- Outer Track -->
              <circle
                cx="110"
                cy="110"
                r="94"
                fill="none"
                [attr.stroke]="hud.tierColorHex"
                stroke-opacity="0.18"
                stroke-width="10" />

              <!-- Dynamic Proximity Arc -->
              <circle
                cx="110"
                cy="110"
                r="94"
                fill="none"
                [attr.stroke]="hud.tierColorHex"
                stroke-width="10"
                stroke-linecap="round"
                stroke-dasharray="590"
                [attr.stroke-dashoffset]="getArcOffset(hud.closestDistanceMeters, hud.tier)"
                class="transition-all duration-500" />

              <!-- Inner Dashed Tactical Ring -->
              <circle
                cx="110"
                cy="110"
                r="78"
                fill="none"
                [attr.stroke]="hud.tierColorHex"
                stroke-opacity="0.35"
                stroke-width="2"
                stroke-dasharray="4 8" />
            </svg>

            <!-- Central Information Orb -->
            <div class="absolute inset-0 flex flex-col items-center justify-center text-center px-6">
              <div
                class="text-3xl mb-0.5"
                [class.animate-bounce]="hud.tier === 'TIER_3_DANGER'">
                {{ hud.tierIcon }}
              </div>
              <div
                class="text-sm font-black tracking-wide uppercase"
                [style.color]="hud.tierColorHex">
                {{ hud.tierLabel }}
              </div>

              @if (hud.closestDistanceMeters !== null) {
                <div class="text-2xl font-black text-white mt-0.5">
                  {{ hud.closestDistanceMeters }}m
                </div>
                <div class="text-[9px] font-bold tracking-widest text-slate-400 uppercase">
                  ESTIMATED PROXIMITY
                </div>
              } @else {
                <div class="text-lg font-extrabold text-white mt-1">
                  ALL SECURE
                </div>
                <div class="text-[9px] font-bold tracking-widest text-slate-400 uppercase">
                  CONTINUOUS SENSING
                </div>
              }

              @if (hud.targetSpecies.length > 0) {
                <div
                  class="text-[11px] font-bold mt-1 truncate max-w-[150px]"
                  [style.color]="hud.tierColorHex">
                  {{ hud.targetSpecies.join(' • ') }}
                </div>
              }
            </div>
          </div>

          <p class="text-xs font-medium text-slate-300 text-center mt-2 px-4">
            {{ hud.tierDescription }}
          </p>
        </div>

        <!-- 2. High-Priority Alert & Direct Map Link Card -->
        <div
          class="rounded-2xl p-4 bg-slate-900/90 border space-y-3 shadow-lg transition-colors"
          [style.borderColor]="hud.tierColorHex + '66'">
          @if (hud.tier !== 'SAFE') {
            <button
              type="button"
              (click)="focusHazardOnMap()"
              class="w-full py-3 px-4 rounded-xl font-bold text-xs text-white shadow-md flex items-center justify-center gap-2 transition-transform active:scale-[0.99]"
              [style.backgroundColor]="hud.tierColorHex">
              <span>📍</span>
              <span>
                Focus Hazard Location on Map
                @if (hud.closestDistanceMeters !== null) {
                  ({{ hud.closestDistanceMeters }}m ahead)
                }
              </span>
            </button>
          }

          @if (hud.tier === 'TIER_3_DANGER') {
            <button
              type="button"
              (click)="state.triggerEmergencySos()"
              class="w-full py-2.5 px-4 rounded-xl border-2 border-red-500 bg-red-950/50 hover:bg-red-900/60 text-red-300 font-black text-xs tracking-wide uppercase transition-colors">
              🚨 BROADCAST EMERGENCY SOS DISPATCH
            </button>
          }

          <div class="flex items-center justify-between text-[11px] pt-0.5">
            <span class="text-slate-400 font-semibold">
              {{ hud.tier === 'SAFE' ? 'Autonomous Sensor Perimeter: Nominal' : hud.hazardLabel }}
            </span>
            <span class="font-mono text-sky-400">
              {{ state.userLocation().latitude.toFixed(3) }}, {{ state.userLocation().longitude.toFixed(3) }}
            </span>
          </div>
        </div>

        <!-- 3. Collapsible Intel Control Bar -->
        <div class="space-y-2.5">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="text-xs font-bold text-slate-200">Additional Intelligence</span>
              <span class="text-[10px] px-2 py-0.5 rounded-full bg-slate-800 text-slate-400 font-medium">
                {{ showAllDetails() ? 'Expanded' : 'Compact' }}
              </span>
            </div>
            <button
              type="button"
              (click)="toggleExpandAll()"
              class="px-2.5 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-[11px] font-bold transition-colors">
              {{ showAllDetails() ? 'Hide Details ▲' : 'View Details ▼' }}
            </button>
          </div>

          <!-- Quick Category Filter Chips -->
          <div class="flex items-center gap-1.5 overflow-x-auto pb-1">
            @for (sec of availableSections(); track sec.id) {
              <button
                type="button"
                (click)="selectFilter(sec.id)"
                class="flex items-center gap-1 px-2.5 py-1.5 rounded-xl text-[11px] font-semibold whitespace-nowrap border transition-all"
                [ngClass]="
                  activeFilter() === sec.id
                    ? 'bg-sky-600 border-sky-400 text-white'
                    : 'bg-slate-900 border-slate-800 text-slate-300 hover:border-slate-700'
                ">
                <span>{{ sec.icon }}</span>
                <span>{{ sec.label }}</span>
              </button>
            }
          </div>
        </div>

        <!-- 4. Expandable Intelligence Cards -->
        @if (showAllDetails() || activeFilter() !== null) {
          <div class="space-y-3">
            <!-- Siren & Haptic Deterrent Card -->
            @if (
              (activeFilter() === null || activeFilter() === 'DETERRENT_SIREN') &&
              (hud.tier === 'TIER_3_DANGER' || state.isSirenPlaying())
            ) {
              <div class="p-3.5 rounded-xl bg-red-950/40 border-2 border-red-500/80 space-y-2.5">
                <div class="flex items-center justify-between gap-2">
                  <div class="flex items-center gap-2">
                    <span class="text-lg">🚨</span>
                    <div>
                      <div class="text-xs font-black text-red-400 uppercase">Acoustic & Haptic Alert</div>
                      <div class="text-[11px] text-slate-300">
                        {{ state.isSirenPlaying() ? 'Emitting 95dB Deterrent Siren' : 'Acoustic Deterrent Armed' }}
                      </div>
                    </div>
                  </div>
                  <button
                    type="button"
                    (click)="state.toggleSiren()"
                    class="px-3 py-1.5 rounded-lg text-[11px] font-bold text-white"
                    [ngClass]="state.isSirenPlaying() ? 'bg-slate-700 hover:bg-slate-600' : 'bg-red-600 hover:bg-red-500'">
                    {{ state.isSirenPlaying() ? 'MUTE SIREN' : 'TRIGGER SIREN' }}
                  </button>
                </div>
                <div class="flex items-center justify-between px-3 py-1.5 rounded-lg bg-slate-950/70 text-[11px]">
                  <span class="font-semibold text-slate-200">📳 Capacitor Haptic Pulse</span>
                  <span class="font-bold text-red-400">
                    {{ hud.tier === 'TIER_3_DANGER' ? 'VIBRATING (Rapid Pulse)' : 'ARMED' }}
                  </span>
                </div>
              </div>
            }

            <!-- Detour Suggestion Card -->
            @if (
              (activeFilter() === null || activeFilter() === 'DETOUR') &&
              hud.suggestedDetour
            ) {
              <div class="p-3.5 rounded-xl bg-slate-900 border border-slate-800 space-y-2.5">
                <div class="flex items-start gap-2.5">
                  <span class="text-lg">🔀</span>
                  <div>
                    <div class="text-xs font-bold text-sky-400">Recommended Safety Detour Route</div>
                    <p class="text-xs text-slate-200 mt-0.5 leading-relaxed">{{ hud.suggestedDetour }}</p>
                  </div>
                </div>
                <button
                  type="button"
                  (click)="focusHazardOnMap()"
                  class="w-full py-1.5 px-3 rounded-lg border border-slate-700 hover:bg-slate-800 text-[11px] font-semibold text-slate-200 transition-colors">
                  🗺️ View Detour Corridor on Map
                </button>
              </div>
            }

            <!-- AI Causal Explanation Card -->
            @if (activeFilter() === null || activeFilter() === 'AI_CAUSAL') {
              <div class="p-3.5 rounded-xl bg-slate-900/90 border border-slate-800 space-y-1.5">
                <div class="flex items-center gap-1.5 text-[11px]">
                  <span class="font-bold text-sky-400">🤖 AI Causal Reasoning</span>
                  <span class="text-slate-500">• Powered by Gemini 3.8 Flash</span>
                </div>
                <p class="text-xs text-slate-200 leading-relaxed">
                  {{ hud.causalExplanation }}
                </p>
              </div>
            }

            <!-- Species Safety Protocol Card -->
            @if (
              (activeFilter() === null || activeFilter() === 'SAFETY_PROTOCOL') &&
              hud.targetSpecies.length > 0
            ) {
              <div class="p-3.5 rounded-xl bg-slate-900 border border-slate-800 space-y-2">
                <div class="text-xs font-bold text-white flex items-center gap-1.5">
                  <span>🛡️</span>
                  <span>Species Safety & Deterrent Protocol</span>
                </div>
                @for (sp of hud.targetSpecies; track sp) {
                  <div class="p-2.5 rounded-lg bg-slate-950 border border-slate-800/80 space-y-1">
                    <div class="text-[11px] font-bold text-sky-400">{{ getSpeciesTitle(sp) }}</div>
                    <div class="text-[11px] text-slate-300 whitespace-pre-line leading-relaxed">{{ getSpeciesProtocol(sp) }}</div>
                  </div>
                }
              </div>
            }

            <!-- Municipal Action Tasks -->
            @if (
              (activeFilter() === null || activeFilter() === 'MUNICIPAL_TASKS') &&
              (state.currentRole() === 'OFFICIAL' || state.currentRole() === 'ADMIN')
            ) {
              <div class="p-3.5 rounded-xl bg-slate-900 border border-slate-800 space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-bold text-sky-400">🎯 Prioritized Municipal Action Tasks</span>
                  <span class="text-[9px] font-bold px-2 py-0.5 rounded bg-red-950 text-red-400 border border-red-800">
                    AGENT DISPATCH
                  </span>
                </div>
                @for (rec of state.dispatchRecommendations; track rec.location) {
                  <div class="p-2.5 rounded-lg bg-slate-950 border border-slate-800 space-y-1">
                    <div class="flex items-center justify-between">
                      <span class="text-[11px] font-bold text-white">[{{ rec.actionType }}] {{ rec.location }}</span>
                      <span class="text-[9px] font-bold px-1.5 py-0.5 rounded bg-red-600 text-white">{{ rec.priority }}</span>
                    </div>
                    <p class="text-[11px] text-slate-400">{{ rec.justification }}</p>
                  </div>
                }
              </div>
            }
          </div>
        }
      }

      <!-- 5. Primary Actions Footer -->
      <div class="pt-2 space-y-2">
        <div class="grid grid-cols-2 gap-2">
          <button
            type="button"
            (click)="state.activeLeftTab.set('REPORT'); state.mobileActiveView.set('REPORT')"
            class="py-2.5 px-3 rounded-xl bg-slate-900 hover:bg-slate-800 border border-slate-700 text-xs font-bold text-slate-200 flex items-center justify-center gap-1.5 transition-colors">
            <span>📸</span>
            <span>Camera Report</span>
          </button>

          @if (state.currentRole() === 'OFFICIAL' || state.currentRole() === 'ADMIN') {
            <button
              type="button"
              (click)="state.activeLeftTab.set('CONSOLE'); state.mobileActiveView.set('CONSOLE')"
              class="py-2.5 px-3 rounded-xl bg-sky-950 hover:bg-sky-900 border border-sky-700 text-xs font-bold text-sky-200 flex items-center justify-center gap-1.5 transition-colors">
              <span>⚙️</span>
              <span>Official Console</span>
            </button>
          } @else {
            <button
              type="button"
              (click)="state.requestRealDeviceGps()"
              class="py-2.5 px-3 rounded-xl bg-slate-900 hover:bg-slate-800 border border-slate-700 text-xs font-bold text-sky-300 flex items-center justify-center gap-1.5 transition-colors">
              <span>📡</span>
              <span>Sync Device GPS</span>
            </button>
          }
        </div>
      </div>
    </div>
  `
})
export class HudPanelComponent {
  state = inject(ShieldStateService);

  showAllDetails = signal<boolean>(true);
  activeFilter = signal<HudIntelSection | null>(null);

  availableSections() {
    const hud = this.state.hudPresentation();
    const role = this.state.currentRole();
    const list: { id: HudIntelSection; label: string; icon: string }[] = [
      { id: 'AI_CAUSAL', label: 'AI Analysis', icon: '🤖' }
    ];
    if (hud.suggestedDetour) {
      list.push({ id: 'DETOUR', label: 'Detour Route', icon: '🔀' });
    }
    if (hud.targetSpecies.length > 0) {
      list.push({ id: 'SAFETY_PROTOCOL', label: 'Safety Rules', icon: '🛡️' });
    }
    if (role === 'OFFICIAL' || role === 'ADMIN') {
      list.push({ id: 'MUNICIPAL_TASKS', label: 'Tasks', icon: '🎯' });
    }
    if (hud.tier === 'TIER_3_DANGER' || this.state.isSirenPlaying()) {
      list.push({ id: 'DETERRENT_SIREN', label: 'Siren & Haptic', icon: '🚨' });
    }
    return list;
  }

  toggleExpandAll() {
    const next = !this.showAllDetails();
    this.showAllDetails.set(next);
    if (!next) {
      this.activeFilter.set(null);
    }
  }

  selectFilter(sec: HudIntelSection) {
    if (this.activeFilter() === sec) {
      this.activeFilter.set(null);
    } else {
      this.activeFilter.set(sec);
      this.showAllDetails.set(true);
    }
  }

  getArcOffset(distanceMeters: number | null, tier: string): number {
    const circumference = 590;
    if (tier === 'SAFE') return 0;
    if (distanceMeters === null) return circumference * 0.4;
    const clamped = Math.max(50, Math.min(2000, distanceMeters));
    const ratio = (2000 - clamped) / 1950;
    const fillRatio = Math.max(0.2, Math.min(0.92, ratio));
    return circumference * (1 - fillRatio);
  }

  focusHazardOnMap() {
    const hud = this.state.hudPresentation();
    if (hud.hazardRegion && hud.hazardLocation) {
      this.state.jumpToRegion(hud.hazardRegion, hud.hazardLocation, hud.hazardEntityId);
    }
    this.state.mobileActiveView.set('MAP');
  }

  getSpeciesTitle(sp: WildlifeSpecies): string {
    switch (sp) {
      case 'BEAR':
        return 'Asian Black Bear (Ursus thibetanus)';
      case 'TIGER':
        return 'Sumatran Tiger (Panthera tigris sumatrae)';
      case 'ELEPHANT':
        return 'Asian Elephant (Elephas maximus)';
      default:
        return sp;
    }
  }

  getSpeciesProtocol(sp: WildlifeSpecies): string {
    switch (sp) {
      case 'BEAR':
        return '1. Make presence known with voice / bear bell.\n2. Do NOT run or climb trees; stand ground.\n3. Back away slowly facing animal.\n4. Arm capsaicin spray.';
      case 'TIGER':
        return '1. Maintain upright posture; never crouch.\n2. Do NOT turn your back or run.\n3. Keep continuous eye contact.\n4. Retreat slowly to vehicle or night corral.';
      case 'ELEPHANT':
        return '1. Dim headlights and do NOT sound vehicle horns.\n2. Maintain minimum 100m standoff buffer.\n3. Do NOT block migration trails to water or crop buffers.';
      default:
        return 'Maintain safe distance and avoid confrontation.';
    }
  }
}
