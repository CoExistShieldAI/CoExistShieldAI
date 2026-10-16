import { Component, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ShieldStateService } from '../../services/shield-state.service';
import {
  DecayState,
  PropertyContextTag,
  ReportCategory,
  WildlifeSpecies
} from '../../models/shield.model';

export type WizardStep = 1 | 2 | 3;

interface CategoryOption {
  id: ReportCategory;
  icon: string;
  label: string;
  subtitle: string;
  badgeText: string;
  isAlert: boolean;
}

@Component({
  selector: 'app-report-panel',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="h-full flex flex-col bg-slate-950 text-slate-100 overflow-hidden">
      <!-- Top Header & Wizard Stepper -->
      <div class="px-4 pt-3.5 pb-3 border-b border-slate-800 bg-slate-900/60 space-y-3 flex-shrink-0">
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-base font-extrabold text-white flex items-center gap-2">
              <span>📸 Field Incident Report</span>
              <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-sky-950 text-sky-300 border border-sky-700/60">
                Step {{ currentStep() }} of 3
              </span>
            </h2>
            <p class="text-[11px] text-slate-400">Guided Gemini 3.8 Flash Multimodal Verification</p>
          </div>
          <button
            type="button"
            (click)="state.activeLeftTab.set('HUD'); state.mobileActiveView.set('SPLIT_OR_HUD')"
            class="px-2.5 py-1.5 rounded-lg border border-slate-700 hover:bg-slate-800 text-xs font-semibold text-slate-300 transition-colors">
            ✕ Close
          </button>
        </div>

        <!-- Visual 3-Step Progress Bar -->
        <div class="grid grid-cols-3 gap-2">
          @for (st of steps; track st.step) {
            <button
              type="button"
              (click)="goToStep(st.step)"
              class="flex flex-col gap-1.5 text-left group focus:outline-none">
              <div class="flex items-center gap-1.5">
                <span
                  class="w-5 h-5 rounded-full text-[11px] font-extrabold flex items-center justify-center transition-all"
                  [ngClass]="
                    currentStep() === st.step
                      ? 'bg-sky-500 text-slate-950 ring-2 ring-sky-400/40'
                      : currentStep() > st.step
                      ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/50'
                      : 'bg-slate-800 text-slate-400 border border-slate-700'
                  ">
                  {{ currentStep() > st.step ? '✓' : st.step }}
                </span>
                <span
                  class="text-[11px] font-bold truncate transition-colors"
                  [ngClass]="
                    currentStep() === st.step
                      ? 'text-white'
                      : currentStep() > st.step
                      ? 'text-emerald-300'
                      : 'text-slate-400 group-hover:text-slate-200'
                  ">
                  {{ st.title }}
                </span>
              </div>
              <div
                class="h-1 w-full rounded-full transition-all"
                [ngClass]="
                  currentStep() === st.step
                    ? 'bg-sky-500'
                    : currentStep() > st.step
                    ? 'bg-emerald-500/70'
                    : 'bg-slate-800'
                "></div>
            </button>
          }
        </div>
      </div>

      <!-- Scrollable Step Content Area -->
      <div class="flex-1 overflow-y-auto px-4 py-4 space-y-4">
        <!-- Offline Pending Queue Banner (Always visible if items are queued) -->
        @if (state.offlineQueue().length > 0) {
          <div class="p-3 rounded-xl bg-amber-950/50 border border-amber-500/70 flex items-center justify-between gap-3">
            <div class="flex items-center gap-2.5">
              <span class="text-lg">📦</span>
              <div>
                <div class="text-xs font-bold text-amber-300">
                  Offline Queue: {{ state.offlineQueue().length }} pending report(s)
                </div>
                <div class="text-[10px] text-slate-300">
                  Stored locally • Auto-syncs when online
                </div>
              </div>
            </div>
            <button
              type="button"
              (click)="state.syncOfflineQueue()"
              class="px-2.5 py-1.5 rounded-lg bg-amber-500 hover:bg-amber-400 text-slate-950 font-bold text-[11px] whitespace-nowrap">
              Sync Now
            </button>
          </div>
        }

        <!-- ========================================================= -->
        <!-- STEP 1: SELECT INCIDENT CATEGORY                          -->
        <!-- ========================================================= -->
        @if (currentStep() === 1) {
          <div class="space-y-3">
            <div>
              <div class="text-xs font-extrabold uppercase tracking-wider text-sky-400">
                Step 1 • What did you observe?
              </div>
              <p class="text-xs text-slate-400 mt-0.5">
                Select the report type. This pre-configures the AI forensic model and required scene details.
              </p>
            </div>

            <div class="space-y-2.5">
              @for (cat of categories; track cat.id) {
                <button
                  type="button"
                  (click)="selectCategory(cat.id)"
                  class="w-full p-3.5 rounded-2xl border text-left transition-all flex items-start gap-3.5 group"
                  [ngClass]="
                    state.reportCategory() === cat.id
                      ? cat.isAlert
                        ? 'bg-red-950/50 border-red-500 shadow-lg shadow-red-950/40'
                        : 'bg-sky-950/50 border-sky-400 shadow-lg shadow-sky-950/40'
                      : 'bg-slate-900/70 border-slate-800 hover:border-slate-700 hover:bg-slate-900'
                  ">
                  <!-- Icon Box -->
                  <div
                    class="w-11 h-11 rounded-xl flex items-center justify-center text-2xl flex-shrink-0 border transition-transform group-hover:scale-105"
                    [ngClass]="
                      state.reportCategory() === cat.id
                        ? cat.isAlert
                          ? 'bg-red-900/50 border-red-500/60'
                          : 'bg-sky-900/50 border-sky-400/60'
                        : 'bg-slate-950 border-slate-800'
                    ">
                    {{ cat.icon }}
                  </div>

                  <!-- Content -->
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center justify-between gap-2">
                      <span class="text-sm font-extrabold text-white">{{ cat.label }}</span>
                      <span
                        class="px-2 py-0.5 rounded-md text-[10px] font-bold whitespace-nowrap border"
                        [ngClass]="
                          cat.isAlert
                            ? 'bg-red-950/90 text-red-300 border-red-700/70'
                            : 'bg-slate-950 text-slate-300 border-slate-800'
                        ">
                        {{ cat.badgeText }}
                      </span>
                    </div>
                    <p class="text-[11px] text-slate-300 mt-1 leading-relaxed">
                      {{ cat.subtitle }}
                    </p>
                  </div>

                  <!-- Radio Indicator -->
                  <div
                    class="w-5 h-5 rounded-full flex items-center justify-center text-xs font-bold mt-0.5 flex-shrink-0"
                    [ngClass]="
                      state.reportCategory() === cat.id
                        ? cat.isAlert
                          ? 'bg-red-500 text-white'
                          : 'bg-sky-500 text-slate-950'
                        : 'border border-slate-700 bg-slate-950 text-transparent'
                    ">
                    ✓
                  </div>
                </button>
              }
            </div>
          </div>
        }

        <!-- ========================================================= -->
        <!-- STEP 2: PHOTO, GPS & SCENE DETAILS                        -->
        <!-- ========================================================= -->
        @if (currentStep() === 2) {
          <div class="space-y-4">
            <!-- Selected Category Pill Bar -->
            <div class="p-2.5 rounded-xl bg-slate-900/90 border border-slate-800 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <span class="text-lg">{{ activeCategoryMeta().icon }}</span>
                <div>
                  <div class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Selected Category</div>
                  <div class="text-xs font-extrabold text-white">{{ activeCategoryMeta().label }}</div>
                </div>
              </div>
              <button
                type="button"
                (click)="goToStep(1)"
                class="px-2.5 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-[11px] font-semibold text-sky-300">
                Change
              </button>
            </div>

            <!-- 2A. Photo Evidence & EXIF GPS Lock Card -->
            <div class="p-4 rounded-2xl bg-slate-900 border border-slate-800 space-y-3">
              <div class="flex items-center justify-between">
                <div class="text-xs font-extrabold uppercase tracking-wider text-sky-400">
                  1. Field Photo & GPS Tag
                </div>
                <div class="flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-950/80 border border-emerald-700/60">
                  <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
                  <span class="text-[10px] font-bold text-emerald-300">EXIF GPS Locked</span>
                </div>
              </div>

              <!-- Visual Camera Preview Slot -->
              <div class="p-3 rounded-xl bg-slate-950 border border-slate-800/90 flex items-center justify-between gap-3">
                <div class="flex items-center gap-3">
                  <div class="w-12 h-12 rounded-xl bg-sky-950/60 border border-sky-700/50 flex items-center justify-center text-2xl">
                    📷
                  </div>
                  <div>
                    <div class="text-xs font-bold text-white">1 Field Evidence Photo Attached</div>
                    <div class="text-[11px] font-mono text-slate-400 mt-0.5">
                      📍 {{ state.userLocation().latitude.toFixed(4) }}, {{ state.userLocation().longitude.toFixed(4) }}
                    </div>
                  </div>
                </div>
                <div class="flex flex-col sm:flex-row gap-1.5">
                  <button
                    type="button"
                    (click)="state.showToast('📸 Camera reopened — New field evidence photo captured.')"
                    class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 border border-slate-700 text-[11px] font-bold text-slate-200 whitespace-nowrap">
                    📸 Retake
                  </button>
                  <button
                    type="button"
                    (click)="state.requestRealDeviceGps()"
                    class="px-2.5 py-1.5 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-700 text-[11px] font-semibold text-sky-300 whitespace-nowrap">
                    📍 GPS
                  </button>
                </div>
              </div>
            </div>

            <!-- 2B. Dynamic Context Options by Category -->
            <div class="p-4 rounded-2xl bg-slate-900 border border-slate-800 space-y-3">
              <div class="text-xs font-extrabold uppercase tracking-wider text-sky-400">
                2. Scene Context & Details
              </div>

              @switch (state.reportCategory()) {
                @case ('WILDLIFE_SIGHTING') {
                  <div class="space-y-2.5">
                    <div class="text-xs font-bold text-white">Select Wildlife Species Observed:</div>
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                      @for (sp of speciesOptions; track sp.id) {
                        <button
                          type="button"
                          (click)="state.reportSpecies.set(sp.id)"
                          class="p-3 rounded-xl border text-left flex items-center gap-2.5 transition-all"
                          [ngClass]="
                            state.reportSpecies() === sp.id
                              ? 'bg-sky-600/25 border-sky-400 text-white shadow'
                              : 'bg-slate-950 border-slate-800 text-slate-300 hover:border-slate-700'
                          ">
                          <span class="text-xl">{{ sp.icon }}</span>
                          <div class="flex-1 min-w-0">
                            <div class="text-xs font-bold truncate">{{ sp.label }}</div>
                            <div class="text-[10px] text-slate-400 truncate">{{ sp.regionHint }}</div>
                          </div>
                          @if (state.reportSpecies() === sp.id) {
                            <span class="text-sky-400 text-xs font-bold">✓</span>
                          }
                        </button>
                      }
                    </div>
                  </div>
                }

                @case ('PHYSICAL_TRACE') {
                  <div class="space-y-2.5">
                    <div class="text-xs font-bold text-white">Select Physical Trace Type:</div>
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                      @for (tr of traceOptions; track tr.id) {
                        <button
                          type="button"
                          (click)="state.reportTraceType.set(tr.id)"
                          class="p-3 rounded-xl border text-left flex items-center gap-2.5 transition-all"
                          [ngClass]="
                            state.reportTraceType() === tr.id
                              ? 'bg-sky-600/25 border-sky-400 text-white shadow'
                              : 'bg-slate-950 border-slate-800 text-slate-300 hover:border-slate-700'
                          ">
                          <span class="text-xl">{{ tr.icon }}</span>
                          <div class="flex-1 min-w-0">
                            <div class="text-xs font-bold truncate">{{ tr.label }}</div>
                            <div class="text-[10px] text-slate-400 truncate">{{ tr.hint }}</div>
                          </div>
                          @if (state.reportTraceType() === tr.id) {
                            <span class="text-sky-400 text-xs font-bold">✓</span>
                          }
                        </button>
                      }
                    </div>
                  </div>
                }

                @case ('ROADKILL') {
                  <div class="space-y-3">
                    <div class="text-xs font-bold text-white">1. Carcass Species:</div>
                    <div class="grid grid-cols-3 gap-2">
                      @for (sp of roadkillSpeciesOptions; track sp.id) {
                        <button
                          type="button"
                          (click)="state.reportSpecies.set(sp.id)"
                          class="p-2.5 rounded-xl border text-center transition-all"
                          [ngClass]="
                            state.reportSpecies() === sp.id
                              ? 'bg-sky-600/25 border-sky-400 text-white'
                              : 'bg-slate-950 border-slate-800 text-slate-400 hover:text-white'
                          ">
                          <div class="text-lg">{{ sp.icon }}</div>
                          <div class="text-[11px] font-bold mt-0.5">{{ sp.label }}</div>
                        </button>
                      }
                    </div>

                    <div class="text-xs font-bold text-white pt-1">2. Decay State (Olfactory Lure Index):</div>
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                      @for (dc of decayOptions; track dc.id) {
                        <button
                          type="button"
                          (click)="state.reportDecayState.set(dc.id)"
                          class="p-3 rounded-xl border text-left flex items-center justify-between transition-all"
                          [ngClass]="
                            state.reportDecayState() === dc.id
                              ? 'bg-amber-600/25 border-amber-400 text-white shadow'
                              : 'bg-slate-950 border-slate-800 text-slate-300 hover:border-slate-700'
                          ">
                          <div>
                            <div class="text-xs font-bold">{{ dc.label }}</div>
                            <div class="text-[10px] text-amber-300 font-mono mt-0.5">{{ dc.threat }}</div>
                          </div>
                          @if (state.reportDecayState() === dc.id) {
                            <span class="text-amber-400 text-xs font-bold">✓</span>
                          }
                        </button>
                      }
                    </div>

                    <div class="p-2.5 rounded-xl bg-red-950/40 border border-red-800/70 text-[11px] text-red-300">
                      ⚠️ Triggers municipal sanitation dispatch and activates an 800m predator lure geofence.
                    </div>
                  </div>
                }

                @case ('ATTRACTANT_FRUIT_TREE') {
                  <div class="space-y-3">
                    <div class="text-xs font-bold text-white">Property & Land Context (Kaki / Crop Map):</div>
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                      @for (pt of propertyTags; track pt.id) {
                        <button
                          type="button"
                          (click)="state.reportPropertyTag.set(pt.id)"
                          class="p-3 rounded-xl border text-left flex items-center gap-2.5 transition-all"
                          [ngClass]="
                            state.reportPropertyTag() === pt.id
                              ? 'bg-sky-600/25 border-sky-400 text-white shadow'
                              : 'bg-slate-950 border-slate-800 text-slate-300 hover:border-slate-700'
                          ">
                          <span class="text-lg">{{ pt.icon }}</span>
                          <span class="text-xs font-bold flex-1">{{ pt.label }}</span>
                          @if (state.reportPropertyTag() === pt.id) {
                            <span class="text-sky-400 text-xs font-bold">✓</span>
                          }
                        </button>
                      }
                    </div>

                    <button
                      type="button"
                      (click)="state.reportNearSchoolRoute.set(!state.reportNearSchoolRoute())"
                      class="w-full p-3 rounded-xl border flex items-center justify-between text-xs transition-all"
                      [ngClass]="
                        state.reportNearSchoolRoute()
                          ? 'bg-red-950/40 border-red-500/70 text-white'
                          : 'bg-slate-950 border-slate-800 text-slate-300'
                      ">
                      <div class="flex items-center gap-2.5 text-left">
                        <span class="text-base">🏫</span>
                        <div>
                          <div class="font-bold">Near School Route / Residential Area</div>
                          <div class="text-[10px] text-slate-400">Prioritizes tree in Top 5% Municipal GCF Queue</div>
                        </div>
                      </div>
                      <span
                        class="w-5 h-5 rounded-full flex items-center justify-center text-[11px] font-bold flex-shrink-0"
                        [ngClass]="state.reportNearSchoolRoute() ? 'bg-red-600 text-white' : 'bg-slate-800 text-slate-500'">
                        ✓
                      </span>
                    </button>
                  </div>
                }

                @case ('LIVESTOCK_DEPREDATION') {
                  <div class="space-y-3">
                    <div class="text-xs font-bold text-white">Target Livestock Attacked:</div>
                    <div class="grid grid-cols-2 gap-2">
                      @for (lv of livestockOptions; track lv.id) {
                        <button
                          type="button"
                          (click)="state.reportSpecies.set(lv.id)"
                          class="p-3 rounded-xl border text-left flex items-center gap-2.5 transition-all"
                          [ngClass]="
                            state.reportSpecies() === lv.id
                              ? 'bg-red-600/25 border-red-400 text-white shadow'
                              : 'bg-slate-950 border-slate-800 text-slate-300 hover:border-slate-700'
                          ">
                          <span class="text-xl">{{ lv.icon }}</span>
                          <span class="text-xs font-bold flex-1">{{ lv.label }}</span>
                          @if (state.reportSpecies() === lv.id) {
                            <span class="text-red-400 text-xs font-bold">✓</span>
                          }
                        </button>
                      }
                    </div>
                    <div class="p-3 rounded-xl bg-red-950/50 border border-red-800/80 space-y-1">
                      <div class="text-xs font-bold text-red-300">🚨 72-Hour Predator Caching Alert</div>
                      <p class="text-[11px] text-slate-300 leading-relaxed">
                        Submitting this report immediately establishes a Tier 3 Danger Buffer Zone and dispatches Estate Security & BBKSDA Rangers.
                      </p>
                    </div>
                  </div>
                }
              }
            </div>
          </div>
        }

        <!-- ========================================================= -->
        <!-- STEP 3: AI TRIAGE REVIEW & SUBMIT                         -->
        <!-- ========================================================= -->
        @if (currentStep() === 3) {
          <div class="space-y-4">
            <!-- 3A. Instant Gemini Multimodal Triage Card -->
            @if (state.geminiTriagePreview(); as triage) {
              <div class="p-4 rounded-2xl bg-gradient-to-br from-slate-900 via-slate-900 to-sky-950/40 border border-sky-500/40 space-y-3 shadow-lg">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-1.5">
                    <span class="text-xs font-extrabold text-sky-400">✨ Gemini 3.8 Flash AI Triage</span>
                  </div>
                  <span class="px-2.5 py-0.5 rounded-full bg-emerald-950 text-emerald-300 border border-emerald-700 text-[11px] font-extrabold">
                    {{ (triage.confidence * 100).toFixed(0) }}% Match
                  </span>
                </div>

                <!-- Confidence Meter Bar -->
                <div class="w-full h-1.5 rounded-full bg-slate-800 overflow-hidden">
                  <div
                    class="h-full bg-gradient-to-r from-sky-400 to-emerald-400 rounded-full"
                    [style.width.%]="triage.confidence * 100"></div>
                </div>

                <div>
                  <div class="text-sm font-extrabold text-white">{{ triage.detectedLabel }}</div>
                  @if (triage.commonName) {
                    <div class="text-[11px] font-semibold text-sky-300">{{ triage.commonName }}</div>
                  }
                </div>

                <p class="text-xs text-slate-300 leading-relaxed bg-slate-950/60 p-2.5 rounded-xl border border-slate-800/80">
                  {{ triage.notes }}
                </p>

                <div class="flex flex-wrap gap-1.5">
                  @for (tag of triage.forensicTags; track tag) {
                    <span class="px-2 py-0.5 rounded-md bg-slate-950 border border-slate-800 text-[10px] font-mono text-sky-300">
                      #{{ tag }}
                    </span>
                  }
                </div>
              </div>
            }

            <!-- 3B. Structured Report Summary Checklist -->
            <div class="p-4 rounded-2xl bg-slate-900 border border-slate-800 space-y-3">
              <div class="flex items-center justify-between">
                <span class="text-xs font-extrabold uppercase tracking-wider text-slate-300">
                  Report Summary
                </span>
                <button
                  type="button"
                  (click)="goToStep(2)"
                  class="text-[11px] font-semibold text-sky-400 hover:underline">
                  Edit Details
                </button>
              </div>

              <div class="grid grid-cols-2 gap-2.5 text-xs">
                <div class="p-2.5 rounded-xl bg-slate-950 border border-slate-800/80">
                  <div class="text-[10px] text-slate-400">Category</div>
                  <div class="font-bold text-white mt-0.5 truncate">
                    {{ activeCategoryMeta().icon }} {{ activeCategoryMeta().label }}
                  </div>
                </div>

                <div class="p-2.5 rounded-xl bg-slate-950 border border-slate-800/80">
                  <div class="text-[10px] text-slate-400">Scene Classification</div>
                  <div class="font-bold text-white mt-0.5 truncate">
                    {{ contextSummaryLabel() }}
                  </div>
                </div>

                <div class="p-2.5 rounded-xl bg-slate-950 border border-slate-800/80">
                  <div class="text-[10px] text-slate-400">GPS Coordinates</div>
                  <div class="font-mono font-bold text-emerald-400 mt-0.5 text-[11px]">
                    {{ state.userLocation().latitude.toFixed(4) }}, {{ state.userLocation().longitude.toFixed(4) }}
                  </div>
                </div>

                <div class="p-2.5 rounded-xl bg-slate-950 border border-slate-800/80">
                  <div class="text-[10px] text-slate-400">Field Evidence</div>
                  <div class="font-bold text-white mt-0.5">📷 1 Photo (EXIF)</div>
                </div>
              </div>
            </div>

            <!-- 3C. Offline Mode Simulator Toggle -->
            <div class="p-3 rounded-xl bg-slate-900/70 border border-slate-800 flex items-center justify-between gap-2">
              <div>
                <div class="text-xs font-bold text-slate-200">Mountain / Forest Offline Mode</div>
                <div class="text-[10px] text-slate-400">Queue locally in IndexedDB/Room when out of cell range</div>
              </div>
              <button
                type="button"
                (click)="state.toggleOfflineMode()"
                class="px-3 py-1.5 rounded-lg text-xs font-bold transition-colors whitespace-nowrap"
                [ngClass]="state.isOfflineMode() ? 'bg-amber-600 text-white' : 'bg-slate-800 text-slate-300'">
                {{ state.isOfflineMode() ? '📦 Offline Active' : '🌐 Online' }}
              </button>
            </div>
          </div>
        }
      </div>

      <!-- Sticky Bottom Wizard Navigation Footer -->
      <div class="p-3.5 border-t border-slate-800 bg-slate-900/90 flex items-center justify-between gap-2.5 flex-shrink-0">
        @if (currentStep() > 1) {
          <button
            type="button"
            (click)="prevStep()"
            class="px-4 py-3 rounded-xl border border-slate-700 bg-slate-800 hover:bg-slate-700 text-xs font-bold text-slate-200 transition-colors">
            ← Back
          </button>
        }

        @if (currentStep() < 3) {
          <button
            type="button"
            (click)="nextStep()"
            class="flex-1 py-3 px-4 rounded-xl bg-sky-600 hover:bg-sky-500 font-extrabold text-xs sm:text-sm text-white shadow-lg transition-all flex items-center justify-center gap-2">
            <span>
              {{
                currentStep() === 1
                  ? 'Next: Photo & Scene Details'
                  : 'Next: AI Triage & Review'
              }}
            </span>
            <span>→</span>
          </button>
        } @else {
          <button
            type="button"
            (click)="submitAndReturn()"
            class="flex-1 py-3 px-4 rounded-xl font-extrabold text-xs sm:text-sm text-white shadow-lg transition-transform active:scale-[0.99]"
            [ngClass]="
              state.isOfflineMode()
                ? 'bg-amber-600 hover:bg-amber-500'
                : state.reportCategory() === 'ROADKILL' || state.reportCategory() === 'LIVESTOCK_DEPREDATION'
                ? 'bg-red-600 hover:bg-red-500'
                : 'bg-emerald-600 hover:bg-emerald-500'
            ">
            {{
              state.isOfflineMode()
                ? '📦 Save to Offline Queue'
                : '✨ Submit Verified Report'
            }}
          </button>
        }
      </div>
    </div>
  `
})
export class ReportPanelComponent {
  state = inject(ShieldStateService);

  readonly currentStep = signal<WizardStep>(1);

  readonly steps: { step: WizardStep; title: string }[] = [
    { step: 1, title: 'Category' },
    { step: 2, title: 'Scene & Photo' },
    { step: 3, title: 'AI Review' }
  ];

  readonly categories: CategoryOption[] = [
    {
      id: 'ATTRACTANT_FRUIT_TREE',
      icon: '🍊',
      label: 'Fruit Tree / Crop Attractant',
      subtitle: 'Unharvested persimmon (kaki), sugarcane, or fruit orchard drawing wildlife.',
      badgeText: 'Attractant Map',
      isAlert: false
    },
    {
      id: 'WILDLIFE_SIGHTING',
      icon: '🐻',
      label: 'Live Wildlife Sighting',
      subtitle: 'Direct visual encounter with bear, tiger, elephant, boar, or deer.',
      badgeText: 'Encounter',
      isAlert: false
    },
    {
      id: 'PHYSICAL_TRACE',
      icon: '🐾',
      label: 'Physical Trace / Sign',
      subtitle: 'Footprints, claw marks on tree bark, scat/droppings, or crop damage.',
      badgeText: 'Forensic Sign',
      isAlert: false
    },
    {
      id: 'ROADKILL',
      icon: '🐗',
      label: 'Roadkill Carcass',
      subtitle: 'Roadside carcass acting as an olfactory predator magnet (800m lure zone).',
      badgeText: '⚠️ Predator Lure',
      isAlert: true
    },
    {
      id: 'LIVESTOCK_DEPREDATION',
      icon: '🐅',
      label: 'Livestock Attack',
      subtitle: 'Domestic herd predation triggering an immediate 72-hour Tiger Caching Buffer.',
      badgeText: '🚨 Tier 3 Danger',
      isAlert: true
    }
  ];

  readonly speciesOptions: { id: WildlifeSpecies; icon: string; label: string; regionHint: string }[] = [
    { id: 'BEAR', icon: '🐻', label: 'Asian Black Bear', regionHint: 'Japan • Forest Fringe' },
    { id: 'TIGER', icon: '🐅', label: 'Sumatran Tiger', regionHint: 'Indonesia • Corridor' },
    { id: 'ELEPHANT', icon: '🐘', label: 'Asian Elephant', regionHint: 'Thailand • Crop Border' },
    { id: 'WILD_BOAR', icon: '🐗', label: 'Wild Boar', regionHint: 'All Regions' },
    { id: 'DEER', icon: '🦌', label: 'Sika / Sambar Deer', regionHint: 'All Regions' }
  ];

  readonly roadkillSpeciesOptions: { id: WildlifeSpecies; icon: string; label: string }[] = [
    { id: 'WILD_BOAR', icon: '🐗', label: 'Wild Boar' },
    { id: 'DEER', icon: '🦌', label: 'Deer' },
    { id: 'BEAR', icon: '🐾', label: 'Other Mammal' }
  ];

  readonly livestockOptions: { id: WildlifeSpecies; icon: string; label: string }[] = [
    { id: 'TIGER', icon: '🐄', label: 'Cattle (Bovine)' },
    { id: 'BEAR', icon: '🐐', label: 'Goat / Sheep' }
  ];

  readonly traceOptions = [
    { id: 'CLAW_MARK_BARK_STRIPPING', icon: '🌲', label: 'Claw Marks (Bark)', hint: 'Territorial / climbing sign' },
    { id: 'PUGMARK_SPREAD', icon: '🐾', label: 'Pugmark / Footprint', hint: 'Track pad & stride' },
    { id: 'SCAT_VEGETATIVE_ACORN', icon: '🔍', label: 'Scat / Droppings', hint: 'Dietary forensic residue' },
    { id: 'CROP_TRAMPLING', icon: '🌾', label: 'Crop Trampling', hint: 'Foraging perimeter damage' }
  ];

  readonly decayOptions: { id: DecayState; label: string; threat: string }[] = [
    { id: 'FRESH', label: 'Fresh (0–6h)', threat: 'Lure Index: 0.60' },
    { id: 'BLOATED', label: 'Bloated (Peak Plume)', threat: 'Lure Index: 0.88 (High)' },
    { id: 'DECOMPOSING', label: 'Decomposing', threat: 'Lure Index: 0.75' },
    { id: 'SKELETAL', label: 'Skeletal / Dry', threat: 'Lure Index: 0.20' }
  ];

  readonly propertyTags: { id: PropertyContextTag; icon: string; label: string }[] = [
    { id: 'VACANT_HOUSE', icon: '🏚️', label: 'Vacant House (Akiya)' },
    { id: 'ROADSIDE_UNMANAGED', icon: '🛣️', label: 'Roadside Unmanaged' },
    { id: 'ACTIVE_FARM_BORDER', icon: '🚜', label: 'Farm Border' },
    { id: 'PUBLIC_LAND', icon: '🏞️', label: 'Public Land' }
  ];

  readonly activeCategoryMeta = computed<CategoryOption>(() => {
    const id = this.state.reportCategory();
    return this.categories.find((c) => c.id === id) || this.categories[0];
  });

  readonly contextSummaryLabel = computed<string>(() => {
    switch (this.state.reportCategory()) {
      case 'WILDLIFE_SIGHTING':
        return this.speciesOptions.find((s) => s.id === this.state.reportSpecies())?.label || 'Bear';
      case 'PHYSICAL_TRACE':
        return this.traceOptions.find((t) => t.id === this.state.reportTraceType())?.label || 'Claw Marks';
      case 'ROADKILL':
        return `${this.state.reportDecayState()} (${this.state.reportSpecies()})`;
      case 'ATTRACTANT_FRUIT_TREE': {
        const prop = this.propertyTags.find((p) => p.id === this.state.reportPropertyTag())?.label || 'Vacant House';
        return this.state.reportNearSchoolRoute() ? `${prop} • School Route` : prop;
      }
      case 'LIVESTOCK_DEPREDATION':
        return '72h Tiger Caching Buffer';
    }
  });

  selectCategory(catId: ReportCategory) {
    this.state.reportCategory.set(catId);
  }

  goToStep(step: WizardStep) {
    this.currentStep.set(step);
  }

  nextStep() {
    const curr = this.currentStep();
    if (curr < 3) {
      this.currentStep.set((curr + 1) as WizardStep);
    }
  }

  prevStep() {
    const curr = this.currentStep();
    if (curr > 1) {
      this.currentStep.set((curr - 1) as WizardStep);
    }
  }

  submitAndReturn() {
    this.state.submitIncidentReport();
    this.currentStep.set(1);
    if (!this.state.isOfflineMode()) {
      this.state.activeLeftTab.set('HUD');
    }
  }
}

