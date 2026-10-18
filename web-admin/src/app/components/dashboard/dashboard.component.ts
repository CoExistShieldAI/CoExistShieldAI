import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { AdminApiService } from '../../services/admin-api.service';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterLink],
  template: `
    <div class="p-8 space-y-8 max-w-7xl mx-auto">
      <!-- Title & Welcome -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h2 class="text-2xl font-bold tracking-tight text-white">Operations Command Dashboard</h2>
          <p class="text-sm text-slate-400">Holistic situational awareness across the JAPAC coexistence corridor.</p>
        </div>
        <div class="flex items-center gap-2">
          <a routerLink="/geofences" class="px-3.5 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold shadow transition-all">
            Open JAPAC Map Console
          </a>
          <a routerLink="/incidents/roadkill" class="px-3.5 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold border border-slate-700 transition-all">
            + Log Roadkill
          </a>
        </div>
      </div>

      <!-- KPI Summary Metric Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
        <!-- Card 1: Active Geofences -->
        <div class="p-5 rounded-2xl bg-slate-950/60 border border-slate-800 hover:border-slate-700 transition-all shadow-sm">
          <div class="flex items-center justify-between">
            <span class="text-xs font-medium text-slate-400">Active Geofences</span>
            <span class="p-2 rounded-xl bg-blue-500/10 text-blue-400">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7" />
              </svg>
            </span>
          </div>
          <div class="mt-4 flex items-baseline gap-2">
            <span class="text-3xl font-extrabold text-white">{{ stats()?.total_geofences || 0 }}</span>
            <span class="text-xs text-rose-400 font-semibold">{{ stats()?.tier3_geofences || 0 }} Critical</span>
          </div>
          <div class="mt-3 flex items-center gap-3 text-[11px] text-slate-500 border-t border-slate-800/80 pt-2.5">
            <span>AI: <strong class="text-slate-300">{{ stats()?.ai_geofences || 0 }}</strong></span>
            <span>Manual: <strong class="text-emerald-400">{{ stats()?.manual_geofences || 0 }}</strong></span>
          </div>
        </div>

        <!-- Card 2: Wildlife Sightings (24h) -->
        <div class="p-5 rounded-2xl bg-slate-950/60 border border-slate-800 hover:border-slate-700 transition-all shadow-sm">
          <div class="flex items-center justify-between">
            <span class="text-xs font-medium text-slate-400">Wildlife Sightings (24h)</span>
            <span class="p-2 rounded-xl bg-amber-500/10 text-amber-400">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </span>
          </div>
          <div class="mt-4 flex items-baseline gap-2">
            <span class="text-3xl font-extrabold text-white">{{ stats()?.total_wildlife_24h || 0 }}</span>
            <span class="text-xs text-slate-500">of {{ stats()?.total_wildlife_all || 0 }} all-time</span>
          </div>
          <div class="mt-3 flex items-center gap-2 text-[11px] text-slate-500 border-t border-slate-800/80 pt-2.5">
            <span class="text-amber-400">●</span> Active surveillance active
          </div>
        </div>

        <!-- Card 3: Roadkill Incidents & Predator Magnets -->
        <div class="p-5 rounded-2xl bg-slate-950/60 border border-slate-800 hover:border-slate-700 transition-all shadow-sm">
          <div class="flex items-center justify-between">
            <span class="text-xs font-medium text-slate-400">Pending Carcasses</span>
            <span class="p-2 rounded-xl bg-rose-500/10 text-rose-400">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </span>
          </div>
          <div class="mt-4 flex items-baseline gap-2">
            <span class="text-3xl font-extrabold text-rose-400">{{ stats()?.pending_roadkills || 0 }}</span>
            <span class="text-xs text-slate-500">/ {{ stats()?.total_roadkills || 0 }} total</span>
          </div>
          <div class="mt-3 text-[11px] text-slate-400 border-t border-slate-800/80 pt-2.5 flex items-center justify-between">
            <span>High Predator Magnet Risk</span>
            <a routerLink="/incidents/roadkill" class="text-rose-400 hover:underline">Dispatch</a>
          </div>
        </div>

        <!-- Card 4: Registered App Users -->
        <div class="p-5 rounded-2xl bg-slate-950/60 border border-slate-800 hover:border-slate-700 transition-all shadow-sm">
          <div class="flex items-center justify-between">
            <span class="text-xs font-medium text-slate-400">Protected Mobile Users</span>
            <span class="p-2 rounded-xl bg-emerald-500/10 text-emerald-400">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </span>
          </div>
          <div class="mt-4 flex items-baseline gap-2">
            <span class="text-3xl font-extrabold text-white">{{ stats()?.total_users || 0 | number }}</span>
            <span class="text-xs text-emerald-400 font-medium">+18% this month</span>
          </div>
          <div class="mt-3 text-[11px] text-slate-500 border-t border-slate-800/80 pt-2.5">
            Villagers, plantation staff & rangers
          </div>
        </div>

        <!-- Card 5: AI Met Stations -->
        <div class="p-5 rounded-2xl bg-slate-950/60 border border-slate-800 hover:border-slate-700 transition-all shadow-sm">
          <div class="flex items-center justify-between">
            <span class="text-xs font-medium text-slate-400">AI Met Sensor Nodes</span>
            <span class="p-2 rounded-xl bg-sky-500/10 text-sky-400">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 15a4 4 0 004 4h9a5 5 0 10-.1-9.999 5.002 5.002 0 00-9.78 2.096A4.001 4.001 0 003 15z" />
              </svg>
            </span>
          </div>
          <div class="mt-4 flex items-baseline gap-2">
            <span class="text-3xl font-extrabold text-sky-400">{{ stats()?.total_weather_stations || api.weather().length || 6 }}</span>
            <span class="text-xs text-sky-300 font-medium">Active JAPAC</span>
          </div>
          <div class="mt-3 flex items-center justify-between text-[11px] text-slate-500 border-t border-slate-800/80 pt-2.5">
            <span>Wind / Moon / Precip</span>
            <a routerLink="/weather" class="text-sky-400 hover:underline">Manage</a>
          </div>
        </div>
      </div>

      <!-- Threat Distribution & Recent Incidents Split -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Target Species Threat Distribution -->
        <div class="p-6 rounded-2xl bg-slate-950/60 border border-slate-800 space-y-5">
          <div class="flex items-center justify-between">
            <h3 class="text-base font-semibold text-white">Target Fauna Distribution</h3>
            <span class="text-xs text-slate-400">JAPAC Scope</span>
          </div>
          <div class="space-y-4">
            <div>
              <div class="flex justify-between text-xs text-slate-300 mb-1">
                <span class="font-medium">Asian Elephant (Sri Lanka / India / Thailand)</span>
                <span class="font-mono text-emerald-400">45%</span>
              </div>
              <div class="w-full h-2 bg-slate-800 rounded-full overflow-hidden">
                <div class="h-full bg-emerald-500 rounded-full" style="width: 45%"></div>
              </div>
            </div>

            <div>
              <div class="flex justify-between text-xs text-slate-300 mb-1">
                <span class="font-medium">Bengal & Sumatran Tiger (IN / ID / MY)</span>
                <span class="font-mono text-amber-400">35%</span>
              </div>
              <div class="w-full h-2 bg-slate-800 rounded-full overflow-hidden">
                <div class="h-full bg-amber-500 rounded-full" style="width: 35%"></div>
              </div>
            </div>

            <div>
              <div class="flex justify-between text-xs text-slate-300 mb-1">
                <span class="font-medium">Bears & Apex Wildlife (JP / LK / IN)</span>
                <span class="font-mono text-blue-400">20%</span>
              </div>
              <div class="w-full h-2 bg-slate-800 rounded-full overflow-hidden">
                <div class="h-full bg-blue-500 rounded-full" style="width: 20%"></div>
              </div>
            </div>
          </div>

          <div class="p-3.5 rounded-xl bg-slate-900 border border-slate-800 text-xs text-slate-400 leading-relaxed">
            <strong class="text-slate-200">Algorithmic Correlation:</strong> Carcass locations in Riau directly correlate with 72h tiger predation caching behavior.
          </div>
        </div>

        <!-- Real-time Wildlife Activity Stream -->
        <div class="lg:col-span-2 p-6 rounded-2xl bg-slate-950/60 border border-slate-800 space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-base font-semibold text-white">Live Field Reports</h3>
            <a routerLink="/incidents/wildlife" class="text-xs font-semibold text-emerald-400 hover:text-emerald-300">View All Sightings →</a>
          </div>

          <div class="divide-y divide-slate-800/80">
            @for (w of api.wildlife(); track w.id) {
              <div class="py-3.5 flex items-start justify-between gap-4">
                <div class="space-y-1">
                  <div class="flex items-center gap-2">
                    <span class="text-xs font-bold text-white">{{ w.species }}</span>
                    @if (w.threat_level === 'TIER_3_CRITICAL') {
                      <span class="text-[10px] uppercase font-bold px-2 py-0.5 rounded bg-rose-500/20 text-rose-400 border border-rose-500/30">Tier 3 Danger</span>
                    } @else if (w.threat_level === 'TIER_2_ALERT') {
                      <span class="text-[10px] uppercase font-bold px-2 py-0.5 rounded bg-amber-500/20 text-amber-400 border border-amber-500/30">Tier 2 Alert</span>
                    } @else {
                      <span class="text-[10px] uppercase font-bold px-2 py-0.5 rounded bg-yellow-500/20 text-yellow-400 border border-yellow-500/30">Tier 1 Notice</span>
                    }
                    <span class="text-[10px] font-mono text-slate-500">{{ w.status }}</span>
                  </div>
                  <p class="text-xs text-slate-300 line-clamp-1">{{ w.description }}</p>
                  <p class="text-[11px] text-slate-500">{{ w.location.road_name }} ({{ w.location.region }})</p>
                </div>
                <div class="text-right flex-shrink-0">
                  <span class="text-[11px] text-slate-400">{{ w.reported_at | date:'shortTime' }}</span>
                </div>
              </div>
            } @empty {
              <p class="text-xs text-slate-500 py-4 text-center">No recent sightings logged.</p>
            }
          </div>
        </div>
      </div>
    </div>
  `
})
export class DashboardComponent implements OnInit {
  api = inject(AdminApiService);
  stats = this.api.stats;

  ngOnInit() {
    this.api.getStats().subscribe();
    this.api.getWildlife().subscribe();
    this.api.getWeather().subscribe();
  }
}
