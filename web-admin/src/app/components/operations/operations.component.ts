import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AdminApiService } from '../../services/admin-api.service';
import { EmergencyBroadcastRequest, EmergencyBroadcastResponse } from '../../models/admin.model';

@Component({
  selector: 'app-operations',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="p-8 space-y-8 max-w-7xl mx-auto">
      <div>
        <h2 class="text-2xl font-bold tracking-tight text-white">Operations & Emergency Tooling</h2>
        <p class="text-sm text-slate-400">High-priority field broadcast notifications and GIS data interoperability exports.</p>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <!-- Feature 3: Emergency Area Push Broadcast -->
        <div class="p-6 rounded-2xl bg-slate-950/60 border border-slate-800 space-y-5 shadow-sm">
          <div class="flex items-center gap-3 pb-3 border-b border-slate-800">
            <span class="p-2 rounded-xl bg-rose-500/20 text-rose-400">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
              </svg>
            </span>
            <div>
              <h3 class="text-base font-bold text-white">Emergency Area Push Broadcast</h3>
              <p class="text-xs text-slate-400">Dispatch instant siren/push warnings to mobile devices inside designated zones.</p>
            </div>
          </div>

          <form (ngSubmit)="sendBroadcast()" class="space-y-4 text-xs">
            <div>
              <label class="block text-slate-300 font-medium mb-1">Target Active Geofence Area</label>
              <select [(ngModel)]="broadcastReq.geofence_id" name="geofence"
                class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-rose-500">
                <option value="">-- All High-Risk Zones in Theater --</option>
                @for (g of api.geofences(); track g.id) {
                  <option [value]="g.id">{{ g.name }} ({{ g.threat_level }})</option>
                }
              </select>
            </div>

            <div>
              <label class="block text-slate-300 font-medium mb-1">Urgency Priority</label>
              <div class="grid grid-cols-2 gap-3">
                <label class="flex items-center gap-2 p-2.5 rounded-lg border border-slate-800 bg-slate-900 cursor-pointer">
                  <input type="radio" [(ngModel)]="broadcastReq.priority" name="priority" value="HIGH" class="text-rose-600">
                  <span class="text-slate-200 font-semibold">HIGH (Standard Push)</span>
                </label>
                <label class="flex items-center gap-2 p-2.5 rounded-lg border border-slate-800 bg-slate-900 cursor-pointer">
                  <input type="radio" [(ngModel)]="broadcastReq.priority" name="priority" value="URGENT" class="text-rose-600">
                  <span class="text-rose-400 font-semibold">URGENT (Audio Siren + Push)</span>
                </label>
              </div>
            </div>

            <div>
              <label class="block text-slate-300 font-medium mb-1">Alert Headline Title</label>
              <input type="text" [(ngModel)]="broadcastReq.title" name="title" required
                placeholder="e.g. CRITICAL ALERT: Asian Elephant Herd Approaching Sector 4"
                class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-rose-500 font-semibold">
            </div>

            <div>
              <label class="block text-slate-300 font-medium mb-1">Alert Body & Evacuation Instructions</label>
              <textarea rows="3" [(ngModel)]="broadcastReq.body" name="body" required
                placeholder="Immediate advisory for plantation staff and villagers: Vacate perimeter road 3219 and move toward designated shelters..."
                class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-rose-500"></textarea>
            </div>

            <button
              type="submit"
              [disabled]="isBroadcasting()"
              class="w-full py-2.5 rounded-lg bg-rose-600 hover:bg-rose-500 font-bold text-white shadow-lg shadow-rose-600/30 transition-all flex items-center justify-center gap-2">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z"/></svg>
              <span>{{ isBroadcasting() ? 'Dispatching Warning Net...' : 'Transmit Emergency Broadcast' }}</span>
            </button>
          </form>

          @if (lastBroadcastResult(); as res) {
            <div class="p-4 rounded-xl bg-slate-900 border border-emerald-500/30 text-xs space-y-1">
              <div class="text-emerald-400 font-bold flex items-center gap-1.5">
                <span>✓</span> Broadcast successfully transmitted
              </div>
              <div class="text-slate-300">
                Estimated Recipients Reached: <strong class="text-white">{{ res.estimated_recipients }} users</strong>
              </div>
              <div class="text-[10px] text-slate-500 font-mono">
                ID: {{ res.broadcast_id }} • {{ res.dispatched_at | date:'medium' }}
              </div>
            </div>
          }

          @if (broadcastError(); as err) {
            <div class="p-4 rounded-xl bg-rose-950/40 border border-rose-500/50 text-xs space-y-1.5 animate-fade-in">
              <div class="text-rose-400 font-bold flex items-center gap-1.5">
                <svg class="w-4 h-4 text-rose-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg>
                <span>Emergency Broadcast Dispatch Failed!</span>
              </div>
              <div class="text-slate-200">
                {{ err }}
              </div>
              <div class="text-[11px] text-rose-300 font-medium">
                CRITICAL WARNING: Sirens and push notifications were NOT delivered to field personnel or local communities.
              </div>
            </div>
          }
        </div>

        <!-- Feature 5: GIS & CSV Data Export -->
        <div class="p-6 rounded-2xl bg-slate-950/60 border border-slate-800 space-y-5 shadow-sm flex flex-col justify-between">
          <div>
            <div class="flex items-center gap-3 pb-3 border-b border-slate-800">
              <span class="p-2 rounded-xl bg-blue-500/20 text-blue-400">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
                </svg>
              </span>
              <div>
                <h3 class="text-base font-bold text-white">GIS & Analytics Data Export</h3>
                <p class="text-xs text-slate-400">Export spatial layers in open standard formats for QGIS, ArcGIS, and spreadsheets.</p>
              </div>
            </div>

            <div class="space-y-4 mt-5 text-xs">
              <div>
                <label class="block text-slate-300 font-medium mb-1">Target Dataset</label>
                <select [(ngModel)]="exportDataset"
                  class="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-700 text-white focus:outline-none focus:border-blue-500">
                  <option value="geofences">Dynamic Geofences (AI & Manual)</option>
                  <option value="wildlife">Wildlife Sighting & Encounter Reports</option>
                  <option value="roadkill">Roadkill Incidents & Removal Log</option>
                  <option value="weather">Meteorological Observations (AI Input Features)</option>
                </select>
              </div>

              <div>
                <label class="block text-slate-300 font-medium mb-1">Export Format</label>
                <div class="grid grid-cols-2 gap-3">
                  <label class="flex items-center gap-2 p-3 rounded-xl border border-slate-800 bg-slate-900 cursor-pointer"
                    [class.border-blue-500]="exportFormat === 'geojson'">
                    <input type="radio" [(ngModel)]="exportFormat" value="geojson" class="text-blue-500">
                    <div>
                      <div class="text-slate-200 font-bold">GeoJSON (RFC 7946)</div>
                      <div class="text-[10px] text-slate-400">Full spatial polygons & coordinates</div>
                    </div>
                  </label>

                  <label class="flex items-center gap-2 p-3 rounded-xl border border-slate-800 bg-slate-900 cursor-pointer"
                    [class.border-blue-500]="exportFormat === 'csv'">
                    <input type="radio" [(ngModel)]="exportFormat" value="csv" class="text-blue-500">
                    <div>
                      <div class="text-slate-200 font-bold">CSV Spreadsheet</div>
                      <div class="text-[10px] text-slate-400">Tabular format for Excel / BI tools</div>
                    </div>
                  </label>
                </div>
              </div>

              <div class="p-3.5 rounded-xl bg-slate-900 border border-slate-800 text-[11px] text-slate-400 leading-relaxed">
                <strong class="text-slate-200">Municipal & Plantation DX:</strong>
                Exported files are ready to ingest into municipal dispatch systems in Japan or RSPO compliance reports in Southeast Asia.
              </div>
            </div>
          </div>

          <a
            [href]="getDownloadLink()"
            target="_blank"
            class="w-full py-2.5 rounded-lg bg-blue-600 hover:bg-blue-500 font-bold text-white shadow-lg shadow-blue-600/30 transition-all flex items-center justify-center gap-2 text-xs">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/></svg>
            <span>Download {{ exportDataset | uppercase }} (.{{ exportFormat }})</span>
          </a>
        </div>
      </div>
    </div>
  `
})
export class OperationsComponent implements OnInit {
  api = inject(AdminApiService);

  broadcastReq: EmergencyBroadcastRequest = {
    title: 'CRITICAL ALERT: Asian Elephant Herd Approaching Sector 4',
    body: 'Immediate notice for plantation staff and villagers: Vacate Route 3219 immediately.',
    priority: 'URGENT'
  };
  isBroadcasting = signal<boolean>(false);
  lastBroadcastResult = signal<EmergencyBroadcastResponse | null>(null);
  broadcastError = signal<string | null>(null);

  exportDataset = 'geofences';
  exportFormat: 'geojson' | 'csv' = 'geojson';

  ngOnInit() {
    this.api.getGeofences().subscribe();
  }

  sendBroadcast() {
    this.isBroadcasting.set(true);
    this.lastBroadcastResult.set(null);
    this.broadcastError.set(null);
    this.api.sendEmergencyBroadcast(this.broadcastReq).subscribe({
      next: (res) => {
        this.isBroadcasting.set(false);
        this.lastBroadcastResult.set(res);
      },
      error: (err) => {
        this.isBroadcasting.set(false);
        const detail = err?.error?.message || err?.message || 'Server connection error';
        this.broadcastError.set(`Broadcast failed (${err?.status || '500'}): ${detail}`);
      }
    });
  }

  getDownloadLink(): string {
    return this.api.getExportUrl(this.exportDataset, this.exportFormat);
  }
}
