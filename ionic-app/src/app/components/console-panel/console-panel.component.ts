import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ShieldStateService } from '../../services/shield-state.service';
import { AdminTab } from '../../models/shield.model';

@Component({
  selector: 'app-console-panel',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="h-full overflow-y-auto px-4 py-4 space-y-4 bg-slate-950 text-slate-100">
      <!-- Header -->
      <div class="flex items-center justify-between border-b border-slate-800 pb-3">
        <div class="flex items-center gap-2">
          <h2 class="text-base font-extrabold text-white">Decision-Support Console</h2>
          <span class="px-2 py-0.5 rounded bg-sky-950 text-sky-400 border border-sky-800 text-[10px] font-bold">
            {{ state.currentRole() }}
          </span>
        </div>
        <button
          type="button"
          (click)="state.activeLeftTab.set('HUD'); state.mobileActiveView.set('SPLIT_OR_HUD')"
          class="px-3 py-1.5 rounded-lg border border-slate-700 hover:bg-slate-800 text-xs font-semibold text-slate-300">
          HUD
        </button>
      </div>

      <!-- Sub-Tab Selector -->
      <div class="flex items-center gap-1 overflow-x-auto bg-slate-900 p-1 rounded-xl border border-slate-800">
        @for (t of tabs; track t.id) {
          <button
            type="button"
            (click)="state.adminActiveTab.set(t.id)"
            class="px-2.5 py-1.5 rounded-lg text-[11px] font-bold whitespace-nowrap flex-1 transition-colors"
            [ngClass]="
              state.adminActiveTab() === t.id
                ? 'bg-sky-600 text-white'
                : 'text-slate-400 hover:text-white'
            ">
            {{ t.label }}
          </button>
        }
      </div>

      @switch (state.adminActiveTab()) {
        @case ('JAPAN_MUNICIPAL') {
          <!-- 1. GCF Progress Card -->
          <div class="p-4 rounded-xl bg-sky-950/30 border border-sky-800/60 space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-sky-400">Hometown Tax GCF: Persimmon Clearance</span>
              <span class="px-2 py-0.5 rounded bg-emerald-950 text-emerald-400 text-[10px] font-bold">
                Morioka City (03201)
              </span>
            </div>
            <div class="space-y-1">
              <div class="flex justify-between text-xs font-bold">
                <span>Funds Raised: ¥11,850K / ¥15,000K</span>
                <span class="text-sky-400">79.0% (482 donors)</span>
              </div>
              <div class="w-full h-2 rounded-full bg-slate-800 overflow-hidden">
                <div class="h-full bg-sky-500 rounded-full" style="width: 79%"></div>
              </div>
            </div>
            <div class="grid grid-cols-3 gap-2 pt-1 text-center">
              <div class="p-2 rounded-lg bg-slate-950/70">
                <div class="text-[10px] text-slate-400">Trees Cleared</div>
                <div class="text-sm font-black text-white">{{ state.adminGcfFelledCount() }} / 420</div>
              </div>
              <div class="p-2 rounded-lg bg-slate-950/70">
                <div class="text-[10px] text-slate-400">Risk Reduction</div>
                <div class="text-sm font-black text-emerald-400">-{{ state.adminGcfRiskReductionPct() }}%</div>
              </div>
              <div class="p-2 rounded-lg bg-slate-950/70">
                <div class="text-[10px] text-slate-400">School Kids</div>
                <div class="text-sm font-black text-white">3,420</div>
              </div>
            </div>
          </div>

          <!-- 2. Top 5% Priority Kaki Trees -->
          <div class="space-y-2.5">
            <div class="text-xs font-bold text-white">Top 5% Priority Persimmon Clearance Roster</div>
            @for (tree of state.priorityKakiTrees(); track tree.id) {
              <div class="p-3.5 rounded-xl bg-slate-900 border border-slate-800 space-y-2">
                <div class="flex items-start justify-between gap-2">
                  <div>
                    <span class="px-1.5 py-0.5 rounded bg-sky-600 text-white text-[10px] font-bold mr-1.5">
                      #{{ tree.clearanceOrder }}
                    </span>
                    <span class="text-xs font-bold text-white">{{ tree.addressHint }}</span>
                  </div>
                  <span
                    class="px-2 py-0.5 rounded text-[10px] font-bold whitespace-nowrap"
                    [ngClass]="tree.isFelled ? 'bg-emerald-600 text-white' : 'bg-red-600 text-white'">
                    {{ tree.isFelled ? 'Cleared 0%' : 'Hazard ' + (tree.hazardScore * 100).toFixed(0) + '%' }}
                  </span>
                </div>
                <div class="text-[11px] text-amber-400 font-medium">
                  ⚠️ {{ tree.distanceToSchoolMeters }}m from school route • Fruit: {{ tree.fruitLoad }}
                </div>
                <div class="flex items-center justify-between pt-1">
                  <span class="text-[10px] font-mono text-slate-400">
                    {{ tree.ownerConsentStatus }} • {{ tree.permitId }}
                  </span>
                  <div class="flex items-center gap-1.5">
                    <button
                      type="button"
                      (click)="state.jumpToRegion('JAPAN_IWATE', { latitude: tree.lat, longitude: tree.lon })"
                      class="px-2.5 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-[11px] font-semibold text-sky-300">
                      📍 Map
                    </button>
                    @if (!tree.isFelled) {
                      <button
                        type="button"
                        (click)="state.markKakiTreeFelled(tree.id)"
                        class="px-3 py-1 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-[11px] font-bold text-white">
                        Execute Clearance
                      </button>
                    } @else {
                      <span class="text-[11px] font-bold text-emerald-400">✅ Executed</span>
                    }
                  </div>
                </div>
              </div>
            }
          </div>

          <!-- 3. AI Dispatch Recommendations -->
          <div class="space-y-2.5 pt-1">
            <div class="text-xs font-bold text-white">AI Dispatch Recommendations</div>
            @for (rec of state.dispatchRecommendations; track rec.location) {
              <div class="p-3.5 rounded-xl bg-slate-900 border border-slate-800 space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-extrabold text-amber-400">{{ rec.actionType }}</span>
                  <span class="px-2 py-0.5 rounded bg-red-950 text-red-400 border border-red-800 text-[10px] font-bold">
                    {{ rec.priority }}
                  </span>
                </div>
                <div class="text-xs font-semibold text-white">{{ rec.location }}</div>
                <p class="text-[11px] text-slate-400 leading-relaxed">{{ rec.justification }}</p>
                <button
                  type="button"
                  (click)="state.approveAndDispatchRecommendation(rec)"
                  class="w-full py-2 px-3 rounded-lg bg-sky-600 hover:bg-sky-500 text-[11px] font-bold text-white">
                  Approve & Dispatch Field Team
                </button>
              </div>
            }
          </div>
        }

        @case ('APAC_PLANTATION') {
          <!-- 1. RSPO & FSC ESG Wildlife Compliance -->
          <div class="p-4 rounded-xl bg-slate-900 border border-slate-800 space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-sky-400">RSPO & FSC ESG Wildlife Compliance</span>
              <span class="px-2 py-0.5 rounded bg-emerald-950 text-emerald-400 text-[10px] font-bold">
                {{ state.rspoAuditStatus() }}
              </span>
            </div>
            <div class="text-sm font-black text-white">
              184 Consecutive Days — Zero Human Fatalities across Concession
            </div>
            <div class="grid grid-cols-3 gap-2 text-center">
              <div class="p-2 rounded-lg bg-slate-950">
                <div class="text-[10px] text-slate-400">HCV Monitored</div>
                <div class="text-xs font-bold text-white">42,500 ha</div>
              </div>
              <div class="p-2 rounded-lg bg-slate-950">
                <div class="text-[10px] text-slate-400">Safe Passages</div>
                <div class="text-xs font-bold text-emerald-400">342 Herds</div>
              </div>
              <div class="p-2 rounded-lg bg-slate-950">
                <div class="text-[10px] text-slate-400">Compliance</div>
                <div class="text-xs font-bold text-emerald-400">100% Zero-Harm</div>
              </div>
            </div>
            <button
              type="button"
              (click)="state.exportRspoAuditPackage()"
              class="w-full py-2.5 px-3 rounded-xl bg-sky-600 hover:bg-sky-500 text-xs font-bold text-white">
              📄 Export 1-Click RSPO Audit Package (PDF / CSV)
            </button>
          </div>

          <!-- 2. Thailand Commercial Crop Buffer (Kui Buri / Kaeng Krachan) -->
          <div class="p-4 rounded-xl bg-slate-900 border border-slate-800 space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-white">🇹🇭 Thailand Commercial Crop Buffer</span>
              <span class="px-2 py-0.5 rounded bg-amber-950 text-amber-400 text-[10px] font-bold">
                {{ state.thaiCropShieldStatus() }}
              </span>
            </div>
            <p class="text-[11px] text-slate-400">
              Monitors commercial sugarcane, banana, and durian maturation along Kui Buri & Kaeng Krachan borders.
            </p>
            <div class="flex items-center justify-between text-xs">
              <span class="text-slate-400">Harvested Buffer Plots:</span>
              <span class="font-bold text-emerald-400">
                {{ state.thaiHarvestedPlotsCount() }} / {{ state.thaiTotalPlotsCount() }} Plots Cleared
              </span>
            </div>
            <div class="flex items-center gap-2">
              <button
                type="button"
                (click)="state.jumpToRegion('THAILAND_KUI_BURI')"
                class="px-3 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-[11px] font-bold text-sky-300">
                📍 View Kui Buri
              </button>
              <button
                type="button"
                (click)="state.updateThaiCropHarvest('th_plot_kb_01', 'HARVESTED')"
                class="flex-1 py-2 px-3 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-[11px] font-bold text-white">
                🌾 Mark Sugarcane Plot #KB-01 Harvested
              </button>
            </div>
          </div>

          <!-- 3. Sumatra Livestock & Tiger Caching (Riau / Bukit Kesuma) -->
          <div class="p-4 rounded-xl bg-slate-900 border border-slate-800 space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-white">🇮🇩 Sumatra Tiger Livestock Shield</span>
              <span
                class="px-2 py-0.5 rounded text-[10px] font-bold"
                [ngClass]="
                  state.sumatraActiveKillSitesCount() > 0
                    ? 'bg-red-950 text-red-400'
                    : 'bg-emerald-950 text-emerald-400'
                ">
                {{ state.sumatraActiveKillSitesCount() > 0 ? '72H RETURN HAZARD' : 'ALL CLEAR' }}
              </span>
            </div>
            <div class="flex items-center justify-between text-xs">
              <span class="text-slate-400">Protected Night Corrals: {{ state.sumatraProtectedCorralsCount() }}</span>
              <span class="font-bold text-red-400">Active Kill-Sites: {{ state.sumatraActiveKillSitesCount() }}</span>
            </div>
            <div class="flex items-center gap-2">
              <button
                type="button"
                (click)="state.jumpToRegion('INDONESIA_SUMATRA')"
                class="px-3 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-[11px] font-bold text-sky-300">
                📍 View Sumatra
              </button>
              <button
                type="button"
                (click)="state.markTigerCarcassCleared('ls_sumatra_01')"
                class="flex-1 py-2 px-3 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-[11px] font-bold text-white">
                🐅 Confirm Kill-Site Carcass Cleared
              </button>
            </div>
          </div>

          <!-- 4. Active Carcass & Kill-Site Dispatch Queue -->
          <div class="space-y-2">
            <div class="text-xs font-bold text-white">Active Carcass & Kill-Site Dispatch Queue</div>
            @for (item of state.pendingCarcasses(); track item.id) {
              <div class="p-3.5 rounded-xl bg-slate-900 border border-slate-800 space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-bold text-white">{{ item.species }} • {{ item.locationHint }}</span>
                  <span class="px-2 py-0.5 rounded bg-red-950 text-red-400 text-[10px] font-bold">
                    Threat {{ (item.threatScore * 100).toFixed(0) }}%
                  </span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-[11px] text-slate-400">{{ item.decayState }}</span>
                  <button
                    type="button"
                    [disabled]="item.isDispatched"
                    (click)="state.dispatchCarcassFromConsole(item.id)"
                    class="px-3 py-1 rounded-lg text-[11px] font-bold"
                    [ngClass]="item.isDispatched ? 'bg-slate-800 text-emerald-400' : 'bg-amber-600 hover:bg-amber-500 text-white'">
                    {{ item.isDispatched ? '✅ Dispatched' : '🚛 Dispatch Sanitation' }}
                  </button>
                </div>
              </div>
            }
          </div>
        }

        @case ('SYSTEM_AI') {
          <!-- 1. Autonomous Retraining Engine -->
          <div class="p-4 rounded-xl bg-slate-900 border border-slate-800 space-y-3">
            <div class="text-xs font-bold text-white">Autonomous Retraining Engine</div>
            <div class="grid grid-cols-3 gap-2 text-center">
              <div class="p-2 rounded-lg bg-slate-950">
                <div class="text-[10px] text-slate-400">Last Train</div>
                <div class="text-[11px] font-bold text-white">{{ state.adminRetrainTimestamp() }}</div>
              </div>
              <div class="p-2 rounded-lg bg-slate-950">
                <div class="text-[10px] text-slate-400">Performance</div>
                <div class="text-xs font-black text-emerald-400">AUC {{ state.adminModelAucScore() }}</div>
              </div>
              <div class="p-2 rounded-lg bg-slate-950">
                <div class="text-[10px] text-slate-400">Active Zones</div>
                <div class="text-xs font-bold text-white">{{ state.geofences().length }} Zones</div>
              </div>
            </div>
            <button
              type="button"
              [disabled]="state.adminIsRetraining()"
              (click)="state.triggerAiRetraining()"
              class="w-full py-2.5 px-3 rounded-xl bg-sky-600 hover:bg-sky-500 disabled:bg-slate-800 text-xs font-bold text-white">
              {{ state.adminIsRetraining() ? '🔄 Retraining on Cloud Run...' : '🤖 Force Retrain Models on Accumulated Incidents' }}
            </button>
          </div>

          <!-- 2. Unstructured Municipal Text & Social Media NLP -->
          <div class="p-4 rounded-xl bg-slate-900 border border-slate-800 space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-white">Unstructured Municipal Broadcast NLP</span>
              <span class="px-2 py-0.5 rounded bg-sky-950 text-sky-400 text-[10px] font-bold">
                {{ state.nlpExtractedSightingsCount() }} Extracted
              </span>
            </div>
            <p class="text-[11px] text-slate-400">
              Extracts structured coordinates, species, and threat tiers from municipal emergency emails & LINE alerts via Gemini 3.8 Flash.
            </p>
            <div class="p-2.5 rounded-lg bg-slate-950 border border-slate-800 text-[11px] text-sky-300 font-mono">
              {{ state.lastNlpExtractionResult() }}
            </div>
            <button
              type="button"
              (click)="
                state.testNlpIngestion(
                  '【緊急・盛岡市防災メール】本日午後4時15分頃、高松二丁目付近の柿の木周辺でツキノワグマ1頭（成獣）の目撃情報がありました。'
                )
              "
              class="w-full py-2 px-3 rounded-xl bg-slate-800 hover:bg-slate-700 border border-slate-700 text-xs font-bold text-sky-300">
              📡 Test NLP Extraction on Sample Municipal Alert
            </button>
          </div>

          <!-- 3. BigQuery GIS Streaming Data Lake -->
          <div class="p-4 rounded-xl bg-slate-900 border border-slate-800 space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-white">BigQuery GIS Streaming Data Lake</span>
              <span class="px-2 py-0.5 rounded bg-emerald-950 text-emerald-400 text-[10px] font-bold">
                ACTIVE_STREAMING
              </span>
            </div>
            <p class="text-[11px] text-slate-400">
              Uber H3 Res 8 (~460m edge) & S2 Level 13 spatial indexing • Replication Lag: 12.4s (&lt; 60s SLA)
            </p>
          </div>
        }

        @case ('COLD_START_NGO') {
          <!-- 1. Open Macro Baseline Pre-Training (GBIF + NASA FIRMS) -->
          <div class="p-4 rounded-xl bg-slate-900 border border-slate-800 space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-white">Open Macro Baseline Pre-Training</span>
              <span class="px-2 py-0.5 rounded bg-emerald-950 text-emerald-400 text-[10px] font-bold">
                {{ state.coldStartBootstrapped() ? 'DAY 1 ACTIVE' : 'PENDING' }}
              </span>
            </div>
            <p class="text-[11px] text-slate-400">
              Pre-trains spatial risk priors using GBIF historical occurrences and NASA FIRMS satellite thermal anomalies before local user crowdsourcing begins.
            </p>
            <div class="grid grid-cols-3 gap-2 text-center">
              <div class="p-2 rounded-lg bg-slate-950">
                <div class="text-[10px] text-slate-400">GBIF Records</div>
                <div class="text-xs font-bold text-white">{{ state.gbifOccurrencesCount() }}</div>
              </div>
              <div class="p-2 rounded-lg bg-slate-950">
                <div class="text-[10px] text-slate-400">NASA FIRMS</div>
                <div class="text-xs font-bold text-amber-400">{{ state.firmsAnomaliesCount() }}</div>
              </div>
              <div class="p-2 rounded-lg bg-slate-950">
                <div class="text-[10px] text-slate-400">Day 1 Zones</div>
                <div class="text-xs font-bold text-emerald-400">{{ state.coldStartDayOneGeofencesCount() }} Active</div>
              </div>
            </div>
            <div class="text-[11px] font-semibold text-sky-400">
              Corridors: {{ state.coldStartRolloutRegions().join(' • ') }}
            </div>
            <button
              type="button"
              (click)="state.triggerColdStartBootstrap('ALL')"
              class="w-full py-2.5 px-3 rounded-xl bg-sky-600 hover:bg-sky-500 text-xs font-bold text-white">
              🛰️ Re-Bootstrap Baseline Geofences (GBIF + FIRMS)
            </button>
          </div>

          <!-- 2. Conservation NGO Patrol Ingestion -->
          <div class="p-4 rounded-xl bg-slate-900 border border-slate-800 space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-white">Conservation NGO Patrol Ingestion</span>
              <span class="px-2 py-0.5 rounded bg-sky-950 text-sky-400 text-[10px] font-bold">
                &lt; 5-Min SLA
              </span>
            </div>
            <p class="text-[11px] text-slate-400">
              Allows park rangers and conservation NGOs to ingest offline field telemetry (SMART Conservation CSV, GeoJSON RFC 7946, GPX handheld logs).
            </p>
            <div class="flex items-center justify-between text-xs">
              <span class="text-slate-400">NGO Patrol Records:</span>
              <span class="font-bold text-white">{{ state.ngoPatrolRecordsCount() }} Tracks</span>
            </div>
            <div class="p-2.5 rounded-lg bg-slate-950 border border-slate-800 text-[11px] text-slate-300">
              {{ state.lastNgoImportMessage() }}
            </div>
            <button
              type="button"
              (click)="
                state.bulkImportNgoLogs(
                  'CSV',
                  'date,species,lat,lon,tier\n2026-10-14T06:00:00Z,BEAR,39.7020,141.1520,TIER_2_WARNING'
                )
              "
              class="w-full py-2.5 px-3 rounded-xl bg-amber-600 hover:bg-amber-500 text-xs font-bold text-white">
              📥 Bulk Ingest Offline Ranger Patrol Logs (5-Min SLA)
            </button>
          </div>
        }
      }
    </div>
  `
})
export class ConsolePanelComponent {
  state = inject(ShieldStateService);

  readonly tabs: { id: AdminTab; label: string }[] = [
    { id: 'JAPAN_MUNICIPAL', label: '🇯🇵 Japan GCF' },
    { id: 'APAC_PLANTATION', label: '🌴 APAC RSPO' },
    { id: 'SYSTEM_AI', label: '🤖 AI Health' },
    { id: 'COLD_START_NGO', label: '🌍 NGO Data' }
  ];
}
