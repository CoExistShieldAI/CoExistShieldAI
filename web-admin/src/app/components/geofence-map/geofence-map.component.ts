import { Component, OnInit, inject, signal, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { GoogleMapsModule, MapInfoWindow, MapPolygon, MapCircle, MapMarker } from '@angular/google-maps';
import { AdminApiService } from '../../services/admin-api.service';
import { GeocodingService } from '../../services/geocoding.service';
import {
  AdminGeofence,
  AdminWildlifeIncident,
  AdminRoadkillRecord,
  AdminEnvironmentalData,
  ThreatLevel,
  GeometryType
} from '../../models/admin.model';

export type TimeFilterMode = '24h' | '7d' | '30d' | 'all' | 'custom';

export interface SelectedEntity {
  type: 'geofence' | 'wildlife' | 'roadkill' | 'environmental';
  data: any;
}

@Component({
  selector: 'app-geofence-map',
  standalone: true,
  imports: [CommonModule, FormsModule, GoogleMapsModule],
  template: `
    <div class="h-[calc(100vh-4rem)] flex flex-col bg-slate-950">
      <!-- Toolbar: Row 1 (Header, Active Totals, Quick Actions) -->
      <div class="bg-slate-950 border-b border-slate-800 px-6 py-3 flex flex-wrap items-center justify-between gap-4">
        <div class="flex items-center gap-3">
          <h2 class="text-lg font-bold text-white flex items-center gap-2">
            JAPAC Dynamic Geofence Console
            <span class="text-xs font-mono font-normal px-2.5 py-0.5 rounded-full bg-slate-800 text-slate-300 border border-slate-700">
              {{ api.geofences().length }} Geofences • {{ totalVisibleIncidents() }} Empirical Signals
            </span>
          </h2>
        </div>

        <!-- Controls: Add Manual Fence, Bulk Delete, Reset -->
        <div class="flex items-center gap-2.5">
          @if (isDrawingMode()) {
            <div class="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-purple-950 border border-purple-500 text-purple-200 text-xs font-semibold animate-pulse">
              <span>🎯 Map-Click Active: Click map to set coordinates</span>
              <button
                type="button"
                (click)="finishDrawing()"
                class="px-2 py-0.5 rounded bg-purple-600 hover:bg-purple-500 text-white font-bold text-[11px]">
                Done
              </button>
            </div>
          } @else {
            <button
              (click)="openAddModal()"
              class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-purple-600 hover:bg-purple-500 text-white shadow transition-all">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              + Create Manual Geofence
            </button>
          }

          <button
            (click)="confirmClearAllManual()"
            class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold bg-rose-950 hover:bg-rose-900 text-rose-200 border border-rose-800 transition-all">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            Clear Manual Fences
          </button>

          <!-- Quick Region Jump Dropdown -->
          <div class="flex items-center gap-1.5 bg-slate-900 px-2.5 py-1 rounded-lg border border-slate-700 text-xs">
            <span class="text-slate-400 text-[11px] font-medium">Focus Region:</span>
            <select
              (change)="onRegionFocusChange($event)"
              class="bg-transparent text-slate-200 text-xs focus:outline-none cursor-pointer">
              <option value="japac" class="bg-slate-900 text-white">🌏 JAPAC Pan-Regional</option>
              <option value="srilanka" class="bg-slate-900 text-white">🇱🇰 Sri Lanka (Minneriya / Udawalawe)</option>
              <option value="india" class="bg-slate-900 text-white">🇮🇳 India (Assam / Valparai)</option>
              <option value="thailand" class="bg-slate-900 text-white">🇹🇭 Thailand (Kaeng Krachan)</option>
              <option value="sumatra" class="bg-slate-900 text-white">🇮🇩 Indonesia (Sumatra / Bangka)</option>
              <option value="japan" class="bg-slate-900 text-white">🇯🇵 Japan (Tohoku / Hokkaido)</option>
            </select>
          </div>

          <button
            (click)="resetToJapacView()"
            class="px-2.5 py-1.5 rounded-lg text-xs bg-slate-900 hover:bg-slate-800 text-slate-300 border border-slate-700">
            Reset View
          </button>
        </div>
      </div>

      <!-- Toolbar: Row 2 (Layer Checkboxes & Temporal Time Filtering) -->
      <div class="bg-slate-900/80 border-b border-slate-800 px-6 py-2.5 flex flex-wrap items-center justify-between gap-4 text-xs">
        <!-- Layer Visibility Checkboxes -->
        <div class="flex items-center flex-wrap gap-2">
          <span class="text-slate-400 font-semibold mr-1 text-[11px] uppercase tracking-wider">Layers:</span>
          
          <!-- AI Geofences -->
          <label class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-slate-950 border border-slate-800 hover:border-slate-700 cursor-pointer transition-all">
            <input type="checkbox" [(ngModel)]="showAiGeofences" class="rounded bg-slate-800 border-slate-700 text-blue-500 focus:ring-0">
            <span class="w-2.5 h-2.5 rounded-full bg-blue-500"></span>
            <span class="text-slate-200">AI Fences</span>
            <span class="text-[10px] text-slate-500 font-mono">({{ aiCount() }})</span>
          </label>

          <!-- Manual Geofences -->
          <label class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-slate-950 border border-slate-800 hover:border-slate-700 cursor-pointer transition-all">
            <input type="checkbox" [(ngModel)]="showManualGeofences" class="rounded bg-slate-800 border-slate-700 text-purple-500 focus:ring-0">
            <span class="w-2.5 h-2.5 rounded-full bg-purple-500"></span>
            <span class="text-slate-200">Manual Fences</span>
            <span class="text-[10px] text-slate-500 font-mono">({{ manualCount() }})</span>
          </label>

          <!-- Wildlife Sightings -->
          <label class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-slate-950 border border-slate-800 hover:border-slate-700 cursor-pointer transition-all">
            <input type="checkbox" [(ngModel)]="showWildlife" class="rounded bg-slate-800 border-slate-700 text-cyan-500 focus:ring-0">
            <span class="w-2.5 h-2.5 rounded-full bg-cyan-400"></span>
            <span class="text-slate-200">Wildlife Sightings</span>
            <span class="text-[10px] font-mono" [class.text-cyan-400]="showWildlife">({{ filteredWildlife().length }})</span>
          </label>

          <!-- Roadkill Records -->
          <label class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-slate-950 border border-slate-800 hover:border-slate-700 cursor-pointer transition-all">
            <input type="checkbox" [(ngModel)]="showRoadkills" class="rounded bg-slate-800 border-slate-700 text-rose-500 focus:ring-0">
            <span class="w-2.5 h-2.5 rounded-full bg-rose-500" [class.animate-ping]="hasFreshRoadkill()"></span>
            <span class="text-slate-200">Roadkill Carcasses</span>
            <span class="text-[10px] font-mono" [class.text-rose-400]="showRoadkills">({{ filteredRoadkills().length }})</span>
            @if (hasFreshRoadkill()) {
              <span class="px-1 rounded bg-rose-950 text-rose-300 font-bold text-[9px] border border-rose-800">FRESH</span>
            }
          </label>

          <!-- Environmental Attractants -->
          <label class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-slate-950 border border-slate-800 hover:border-slate-700 cursor-pointer transition-all">
            <input type="checkbox" [(ngModel)]="showEnvironmental" class="rounded bg-slate-800 border-slate-700 text-emerald-500 focus:ring-0">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-500"></span>
            <span class="text-slate-200">Attractant Layers</span>
            <span class="text-[10px] font-mono" [class.text-emerald-400]="showEnvironmental">({{ filteredEnvironmental().length }})</span>
          </label>
        </div>

        <!-- Temporal / Date Filter & Secondary Filters -->
        <div class="flex items-center flex-wrap gap-2.5">
          <!-- Time Presets -->
          <div class="flex items-center p-0.5 rounded-lg bg-slate-950 border border-slate-800 text-[11px]">
            <button
              (click)="timeFilter.set('24h')"
              [class.bg-slate-800]="timeFilter() === '24h'"
              [class.text-white]="timeFilter() === '24h'"
              class="px-2.5 py-1 rounded text-slate-400 hover:text-white transition-colors">
              Past 24h
            </button>
            <button
              (click)="timeFilter.set('7d')"
              [class.bg-slate-800]="timeFilter() === '7d'"
              [class.text-white]="timeFilter() === '7d'"
              class="px-2.5 py-1 rounded text-slate-400 hover:text-white transition-colors">
              Past 7d
            </button>
            <button
              (click)="timeFilter.set('30d')"
              [class.bg-slate-800]="timeFilter() === '30d'"
              [class.text-white]="timeFilter() === '30d'"
              class="px-2.5 py-1 rounded text-slate-400 hover:text-white transition-colors">
              Past 30d
            </button>
            <button
              (click)="timeFilter.set('all')"
              [class.bg-slate-800]="timeFilter() === 'all'"
              [class.text-white]="timeFilter() === 'all'"
              class="px-2.5 py-1 rounded text-slate-400 hover:text-white transition-colors">
              All Time
            </button>
            <button
              (click)="timeFilter.set('custom')"
              [class.bg-emerald-950]="timeFilter() === 'custom'"
              [class.text-emerald-300]="timeFilter() === 'custom'"
              class="px-2.5 py-1 rounded text-slate-400 hover:text-white transition-colors">
              Custom Range
            </button>
          </div>

          <!-- Custom Range Inputs -->
          @if (timeFilter() === 'custom') {
            <div class="flex items-center gap-1 text-[11px] bg-slate-950 px-2 py-0.5 rounded-lg border border-slate-800">
              <input type="date" [(ngModel)]="customDateFrom" class="bg-slate-900 border border-slate-700 text-slate-200 px-1.5 py-0.5 rounded text-[10px]">
              <span class="text-slate-500">to</span>
              <input type="date" [(ngModel)]="customDateTo" class="bg-slate-900 border border-slate-700 text-slate-200 px-1.5 py-0.5 rounded text-[10px]">
            </div>
          }

          <!-- Secondary Filters (Species & Threat Tier) -->
          <select [(ngModel)]="filterSpecies" class="px-2 py-1 rounded-lg bg-slate-950 border border-slate-800 text-slate-300 text-[11px]">
            <option value="all">All Species</option>
            <option value="Asian Elephant">Asian Elephant</option>
            <option value="Sumatran Tiger">Sumatran Tiger</option>
            <option value="Bengal Tiger">Bengal Tiger</option>
            <option value="Malayan Tiger">Malayan Tiger</option>
            <option value="Brown Bear">Brown Bear</option>
            <option value="Asian Black Bear">Asian Black Bear</option>
            <option value="Sloth Bear">Sloth Bear</option>
            <option value="Sun Bear">Sun Bear</option>
            <option value="Indian Leopard">Indian Leopard</option>
            <option value="Saltwater Crocodile">Saltwater Crocodile</option>
            <option value="Wild Boar">Wild Boar</option>
          </select>

          <select [(ngModel)]="filterThreatTier" class="px-2 py-1 rounded-lg bg-slate-950 border border-slate-800 text-slate-300 text-[11px]">
            <option value="all">All Tiers</option>
            <option value="TIER_3_CRITICAL">Tier 3 (Critical)</option>
            <option value="TIER_2_ALERT">Tier 2 (Alert)</option>
            <option value="TIER_1_CAUTION">Tier 1 (Caution)</option>
          </select>
        </div>
      </div>

      <!-- Main Map & Overlay Panel -->
      <div class="flex-1 relative overflow-hidden">
        <!-- Google Map Canvas -->
        <google-map
          #map
          [height]="'100%'"
          [width]="'100%'"
          [center]="center"
          [zoom]="zoom"
          [options]="mapOptions"
          (mapClick)="onMapClick($event)">

          <!-- 1. AI & Manual Geofences: Polygons -->
          @for (geo of visibleGeofences(); track geo.id) {
            @if (geo.geometry_type === 'polygon' && geo.coordinates) {
              <map-polygon
                [paths]="geo.coordinates"
                [options]="getPolygonOptions(geo)"
                (polygonClick)="selectEntity('geofence', geo)">
              </map-polygon>
            }
          }

          <!-- 2. AI & Manual Geofences: Circles -->
          @for (geo of visibleGeofences(); track geo.id) {
            @if (geo.geometry_type === 'circle' && geo.circle_center) {
              <map-circle
                [center]="geo.circle_center"
                [radius]="geo.circle_radius_meters || 2000"
                [options]="getCircleOptions(geo)"
                (circleClick)="selectEntity('geofence', geo)">
              </map-circle>
            }
          }

          <!-- 3. Environmental Attractants: Circles -->
          @if (showEnvironmental) {
            @for (env of filteredEnvironmental(); track env.id) {
              @if (env.circle_center) {
                <map-circle
                  [center]="env.circle_center"
                  [radius]="env.radius_meters || 2000"
                  [options]="getEnvironmentalCircleOptions(env)"
                  (circleClick)="selectEntity('environmental', env)">
                </map-circle>
                <!-- Center marker for clickability -->
                <map-marker
                  [position]="env.circle_center"
                  [options]="getEnvironmentalCenterMarkerOptions(env)"
                  (mapClick)="selectEntity('environmental', env)">
                </map-marker>
              }
            }
          }

          <!-- 4. Wildlife Incidents: Map Markers -->
          @if (showWildlife) {
            @for (w of filteredWildlife(); track w.id) {
              <map-marker
                [position]="{ lat: w.location.lat, lng: w.location.lng }"
                [options]="getWildlifeMarkerOptions(w)"
                (mapClick)="selectEntity('wildlife', w)">
              </map-marker>
            }
          }

          <!-- 5. Roadkill Records: Map Markers -->
          @if (showRoadkills) {
            @for (rk of filteredRoadkills(); track rk.id) {
              <map-marker
                [position]="{ lat: rk.location.lat, lng: rk.location.lng }"
                [options]="getRoadkillMarkerOptions(rk)"
                (mapClick)="selectEntity('roadkill', rk)">
              </map-marker>
            }
          }

          <!-- 6. In-Progress Drawing Preview on Map -->
          @if (isDrawingMode()) {
            @if (newFence.geometry_type === 'circle') {
              <map-circle
                [center]="{ lat: circleLat, lng: circleLng }"
                [radius]="circleRadius"
                [options]="drawingCircleOptions">
              </map-circle>
              <map-marker
                [position]="{ lat: circleLat, lng: circleLng }"
                [options]="drawingCenterMarkerOptions">
              </map-marker>
            } @else if (newFence.geometry_type === 'polygon' && parsedPolygonCoords.length > 0) {
              <map-polygon
                [paths]="parsedPolygonCoords"
                [options]="drawingPolygonOptions">
              </map-polygon>
            }
          }
        </google-map>

        <!-- Map Drawing Mode Active Banner -->
        @if (isDrawingMode()) {
          <div class="absolute top-4 left-1/2 -translate-x-1/2 z-20 px-5 py-2.5 rounded-xl bg-purple-950/90 backdrop-blur border-2 border-purple-500 shadow-2xl flex items-center gap-4 text-xs">
            <span class="w-3 h-3 rounded-full bg-purple-400 animate-ping"></span>
            <div class="text-white font-medium">
              <strong class="text-purple-300">Click anywhere on the map</strong> to place center point or boundary vertices
              @if (newFence.geometry_type === 'circle') {
                <span class="text-slate-300 font-mono ml-2">(Current center: {{ circleLat }}, {{ circleLng }})</span>
              } @else {
                <span class="text-slate-300 font-mono ml-2">({{ parsedPolygonCoords.length }} vertices placed)</span>
              }
            </div>
            <button
              (click)="finishDrawing()"
              class="px-3 py-1 rounded-lg bg-purple-600 hover:bg-purple-500 text-white font-bold text-xs shadow">
              Confirm & Return
            </button>
          </div>
        }

        <!-- Legend Overlay in Corner -->
        <div class="absolute bottom-6 left-6 p-4 rounded-xl bg-slate-950/90 backdrop-blur border border-slate-800 shadow-xl space-y-2 z-10 text-xs">
          <div class="font-bold text-white text-[11px] uppercase tracking-wider">Map Overlay Legend</div>
          <div class="grid grid-cols-2 gap-x-4 gap-y-1.5 text-[11px]">
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded bg-rose-500/50 border border-rose-500"></span>
              <span class="text-slate-300">Tier 3 (Critical)</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full bg-cyan-400 border border-white"></span>
              <span class="text-slate-300">Wildlife Sighting</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded bg-amber-500/50 border border-amber-500"></span>
              <span class="text-slate-300">Tier 2 (Alert)</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full bg-rose-500 border border-white animate-pulse"></span>
              <span class="text-rose-300">Fresh Roadkill (Plume)</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded bg-purple-500/40 border border-purple-400"></span>
              <span class="text-purple-300">Manual Geofence</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="w-3 h-3 rounded-full bg-emerald-500/40 border border-emerald-400"></span>
              <span class="text-emerald-300">Attractant Zone</span>
            </div>
          </div>
        </div>

        <!-- Inspector Drawer & Geofence List (Right Side) -->
        <div class="absolute top-4 right-4 w-96 max-h-[calc(100%-2rem)] bg-slate-950/95 backdrop-blur border border-slate-800 rounded-2xl shadow-2xl p-4 flex flex-col z-10">
          <!-- Card 1: Selected Entity Inspector (If any item is clicked) -->
          @if (selectedEntity(); as ent) {
            <div class="p-3.5 rounded-xl border border-slate-700 bg-slate-900/90 shadow-lg space-y-3 mb-3">
              <div class="flex items-start justify-between gap-2 border-b border-slate-800 pb-2">
                <div>
                  <span [class]="getEntityBadgeClass(ent.type)" class="text-[9px] px-2 py-0.5 rounded-full font-bold uppercase tracking-wider font-mono">
                    {{ getEntityTypeName(ent.type) }}
                  </span>
                  <h4 class="text-sm font-bold text-white mt-1">
                    {{ getEntityTitle(ent) }}
                  </h4>
                </div>
                <button (click)="selectedEntity.set(null)" class="text-slate-400 hover:text-white p-1">✕</button>
              </div>

              <!-- Metadata Details -->
              <div class="text-xs space-y-1.5 text-slate-300">
                @if (ent.type === 'wildlife') {
                  <div class="flex justify-between"><span class="text-slate-500">Species:</span> <strong class="text-white">{{ ent.data.species }}</strong></div>
                  <div class="flex justify-between"><span class="text-slate-500">Threat Tier:</span> <span [class]="getBadgeClass(ent.data.threat_level)" class="px-1.5 py-0.5 rounded text-[10px] font-bold">{{ ent.data.threat_level }}</span></div>
                  <div class="flex justify-between"><span class="text-slate-500">Location:</span> <span>{{ ent.data.location.road_name || 'Forest Interface' }}</span></div>
                  <div class="flex justify-between"><span class="text-slate-500">Reported:</span> <span class="font-mono text-slate-400">{{ ent.data.reported_at | date:'yyyy-MM-dd HH:mm' }}</span></div>
                  @if (ent.data.description) {
                    <p class="text-[11px] text-slate-400 pt-1 border-t border-slate-800/60">{{ ent.data.description }}</p>
                  }
                } @else if (ent.type === 'roadkill') {
                  <div class="flex justify-between"><span class="text-slate-500">Carcass:</span> <strong class="text-white">{{ ent.data.species }} (~{{ ent.data.estimated_weight_kg }}kg)</strong></div>
                  <div class="flex justify-between">
                    <span class="text-slate-500">Decay Stage:</span>
                    <span class="font-bold text-[10px] px-1.5 py-0.5 rounded" [class.bg-rose-950]="ent.data.decay_stage === 'FRESH'" [class.text-rose-400]="ent.data.decay_stage === 'FRESH'">
                      {{ ent.data.decay_stage }}
                    </span>
                  </div>
                  <div class="flex justify-between"><span class="text-slate-500">Status:</span> <span class="font-semibold">{{ ent.data.status }}</span></div>
                  <div class="flex justify-between"><span class="text-slate-500">Scent Plume Risk:</span> <span class="text-rose-400 font-bold">{{ ent.data.decay_stage === 'FRESH' ? 'CRITICAL AMBUSH RISK' : 'MODERATE' }}</span></div>
                  <div class="flex justify-between"><span class="text-slate-500">Reported:</span> <span class="font-mono text-slate-400">{{ ent.data.reported_at | date:'yyyy-MM-dd HH:mm' }}</span></div>
                } @else if (ent.type === 'environmental') {
                  <div class="flex justify-between"><span class="text-slate-500">Attractant Type:</span> <strong class="text-white">{{ ent.data.category }}</strong></div>
                  <div class="flex justify-between"><span class="text-slate-500">Radius Buffer:</span> <span class="font-mono font-bold text-emerald-400">{{ ent.data.radius_meters }}m</span></div>
                  <div class="flex justify-between"><span class="text-slate-500">Risk Weight:</span> <span class="font-mono font-bold text-amber-400">{{ ent.data.risk_weight }}</span></div>
                  <div class="flex justify-between"><span class="text-slate-500">Harvest Window:</span> <span>{{ ent.data.active_season }}</span></div>
                } @else if (ent.type === 'geofence') {
                  <div class="flex justify-between"><span class="text-slate-500">Target Species:</span> <strong class="text-white">{{ ent.data.target_species }}</strong></div>
                  <div class="flex justify-between"><span class="text-slate-500">Source:</span> <span class="font-mono uppercase font-semibold">{{ ent.data.source }}</span></div>
                  <div class="flex justify-between"><span class="text-slate-500">Threat Level:</span> <span [class]="getBadgeClass(ent.data.threat_level)" class="px-1.5 py-0.5 rounded text-[10px] font-bold">{{ ent.data.threat_level }}</span></div>
                  <p class="text-[11px] text-slate-400 pt-1 border-t border-slate-800/60">{{ ent.data.reason }}</p>
                }
              </div>

              <!-- Attached Photos Preview (if any) -->
              @if (ent.data.photo_urls && ent.data.photo_urls.length > 0) {
                <div class="pt-2 border-t border-slate-800">
                  <div class="text-[10px] text-slate-400 font-semibold mb-1">Attached Field Evidence ({{ ent.data.photo_urls.length }}):</div>
                  <div class="flex items-center gap-1.5 overflow-x-auto pb-1">
                    @for (imgUrl of ent.data.photo_urls; track imgUrl) {
                      <a [href]="imgUrl" target="_blank" rel="noopener noreferrer" class="w-12 h-12 rounded-lg overflow-hidden border border-slate-700 hover:border-emerald-500 flex-shrink-0">
                        <img [src]="imgUrl" alt="Field Photo" class="w-full h-full object-cover">
                      </a>
                    }
                  </div>
                </div>
              }

              <!-- One-Click Geofence Generation Action -->
              @if (ent.type !== 'geofence') {
                <div class="pt-2 border-t border-slate-800">
                  <button
                    type="button"
                    (click)="createGeofenceFromEntity(ent)"
                    class="w-full py-2 px-3 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-xs shadow-md shadow-emerald-600/30 flex items-center justify-center gap-1.5 transition-all">
                    <span>📍</span>
                    <span>Create Geofence from this Signal</span>
                  </button>
                </div>
              } @else {
                <div class="pt-2 border-t border-slate-800 flex justify-end">
                  <button (click)="deleteSelectedGeofence(ent.data.id)" class="text-xs text-rose-400 hover:text-rose-200 font-semibold">
                    Delete Geofence
                  </button>
                </div>
              }
            </div>
          }

          <!-- Card 2: Active Geofences List -->
          <div class="flex items-center justify-between pb-2 border-b border-slate-800">
            <h3 class="text-xs font-bold text-white uppercase tracking-wider">Active Geofences</h3>
            <span class="text-[10px] text-slate-400">{{ visibleGeofences().length }} zones</span>
          </div>
          <div class="flex-1 overflow-y-auto space-y-2 mt-2 pr-1">
            @for (geo of visibleGeofences(); track geo.id) {
              <div
                (click)="selectAndPanTo(geo)"
                class="p-2.5 rounded-xl border border-slate-800 hover:border-slate-700 bg-slate-900/60 cursor-pointer transition-all">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-xs font-semibold text-white truncate max-w-[170px]">{{ geo.name }}</span>
                  <span [class]="getBadgeClass(geo.threat_level)" class="text-[9px] px-1.5 py-0.5 rounded font-bold">
                    {{ geo.threat_level.includes('TIER_3') ? 'T3' : geo.threat_level.includes('TIER_2') ? 'T2' : 'T1' }}
                  </span>
                </div>
                <div class="text-[11px] text-slate-400 flex items-center justify-between">
                  <span>{{ geo.target_species }}</span>
                  <span class="font-mono text-[10px]" [class.text-purple-400]="geo.source === 'manual'">
                    {{ geo.geometry_type | uppercase }} • {{ geo.source }}
                  </span>
                </div>
              </div>
            }
          </div>
        </div>
      </div>
    </div>

    <!-- Modal: Add Manual Geofence (Circle / Polygon with Map-Click Support) -->
    @if (showAddModal()) {
      <div class="fixed inset-0 bg-black/75 backdrop-blur-sm flex items-center justify-center p-4 z-50">
        <div class="bg-slate-900 border border-slate-800 rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-5">
          <div class="flex items-center justify-between border-b border-slate-800 pb-3">
            <h3 class="text-base font-bold text-white flex items-center gap-2">
              <span class="p-1.5 rounded-lg bg-purple-500/20 text-purple-400">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/></svg>
              </span>
              Create Manual Geofence
            </h3>
            <button (click)="showAddModal.set(false)" class="text-slate-400 hover:text-white">✕</button>
          </div>

          <form (ngSubmit)="saveManualGeofence()" class="space-y-4 text-xs">
            <div>
              <div class="flex items-center justify-between mb-1">
                <label class="block text-slate-300 font-medium">Geofence Name</label>
                @if (isResolvingLocation()) {
                  <span class="text-[10px] text-purple-400 animate-pulse flex items-center gap-1 font-mono">
                    <span>🔄</span> Resolving address...
                  </span>
                } @else if (resolvedLocationName()) {
                  <button type="button" (click)="applyResolvedLocationToName()"
                    class="text-[10px] text-purple-400 hover:text-purple-300 underline flex items-center gap-1 truncate max-w-[260px]"
                    title="Click to apply reverse-geocoded location name">
                    <span>📍 {{ resolvedLocationName() }}</span>
                  </button>
                }
              </div>
              <div class="relative">
                <input type="text" [(ngModel)]="newFence.name" name="name" required placeholder="e.g. Field Station 3 Thermal Perimeter"
                  class="w-full px-3 py-2 pr-20 rounded-lg bg-slate-800 border border-slate-700 text-white focus:outline-none focus:border-purple-500">
                <button type="button" (click)="resolveReverseGeocode()"
                  class="absolute right-1 top-1 bottom-1 px-2.5 rounded bg-purple-700/60 hover:bg-purple-600 text-purple-200 text-[10px] font-medium flex items-center gap-1">
                  <span>📍 Geocode</span>
                </button>
              </div>
              @if (resolvedLocationName()) {
                <p class="text-[10px] text-slate-400 mt-1 flex items-center gap-1">
                  <span class="text-slate-500">Reverse Geocoded:</span>
                  <span class="text-purple-300 font-medium">{{ resolvedLocationName() }}</span>
                </p>
              }
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-300 font-medium mb-1">Geometry Type</label>
                <select [(ngModel)]="newFence.geometry_type" name="geometry_type"
                  class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white focus:outline-none focus:border-purple-500">
                  <option value="circle">Circle (Center + Radius)</option>
                  <option value="polygon">Polygon (Multi-point Vertices)</option>
                </select>
              </div>
              <div>
                <label class="block text-slate-300 font-medium mb-1">Threat Tier</label>
                <select [(ngModel)]="newFence.threat_level" name="threat_level"
                  class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white focus:outline-none focus:border-purple-500">
                  <option value="TIER_1_CAUTION">Tier 1 - Caution</option>
                  <option value="TIER_2_ALERT">Tier 2 - Alert / Warning</option>
                  <option value="TIER_3_CRITICAL">Tier 3 - Critical Danger</option>
                </select>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-300 font-medium mb-1">Target Species</label>
                <select [(ngModel)]="newFence.target_species" name="target_species"
                  class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white focus:outline-none focus:border-purple-500">
                  <option value="Asian Elephant">Asian Elephant (Sri Lanka / India / Thailand / Sumatra)</option>
                  <option value="Sumatran Tiger">Sumatran Tiger (Indonesia - Sumatra)</option>
                  <option value="Bengal Tiger">Bengal Tiger (India / Bangladesh / Nepal)</option>
                  <option value="Malayan Tiger">Malayan Tiger (Peninsular Malaysia)</option>
                  <option value="Asian Black Bear">Asian Black Bear (Japan / India / Himalayas)</option>
                  <option value="Brown Bear">Brown Bear (Japan - Hokkaido / Himalayas)</option>
                  <option value="Sloth Bear">Sloth Bear (India / Sri Lanka / Nepal)</option>
                  <option value="Sun Bear">Sun Bear (Indonesia / Malaysia / Thailand)</option>
                  <option value="Indian Leopard">Indian Leopard (India / Sri Lanka / Nepal)</option>
                  <option value="Saltwater Crocodile">Saltwater Crocodile (Indonesia / Australia / Malaysia)</option>
                  <option value="Apex Carnivore">Apex Carnivore (Multi-predator)</option>
                  <option value="Other / Multi">Other / Multi-species</option>
                </select>
              </div>
              <div>
                <label class="block text-slate-300 font-medium mb-1">Interactive Map Placement</label>
                <button
                  type="button"
                  (click)="startInteractiveDrawing()"
                  class="w-full px-3 py-2 rounded-lg bg-purple-900/60 hover:bg-purple-800 text-purple-200 border border-purple-600 font-semibold flex items-center justify-center gap-1.5">
                  <span>🎯</span>
                  <span>Click on Map to Position</span>
                </button>
              </div>
            </div>

            <!-- Circle specific inputs -->
            @if (newFence.geometry_type === 'circle') {
              <div class="p-3.5 rounded-xl bg-slate-800/50 border border-slate-700 space-y-2">
                <div class="flex items-center justify-between font-semibold text-slate-200">
                  <span>Circle Center & Radius</span>
                  <span class="text-[10px] text-purple-400 font-normal">Click map to adjust center pin</span>
                </div>
                <div class="grid grid-cols-3 gap-2">
                  <div>
                    <label class="text-[10px] text-slate-400">Center Lat</label>
                    <input type="number" step="0.0001" [(ngModel)]="circleLat" name="circleLat" required
                      class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white font-mono">
                  </div>
                  <div>
                    <label class="text-[10px] text-slate-400">Center Lng</label>
                    <input type="number" step="0.0001" [(ngModel)]="circleLng" name="circleLng" required
                      class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white font-mono">
                  </div>
                  <div>
                    <label class="text-[10px] text-slate-400">Radius (Meters)</label>
                    <input type="number" step="100" min="200" [(ngModel)]="circleRadius" name="circleRadius" required
                      class="w-full px-2 py-1 rounded bg-slate-800 border border-slate-600 text-white font-mono font-bold text-emerald-400">
                  </div>
                </div>
              </div>
            } @else {
              <!-- Polygon specific coordinates input -->
              <div class="p-3.5 rounded-xl bg-slate-800/50 border border-slate-700 space-y-2">
                <div class="flex items-center justify-between font-semibold text-slate-200">
                  <span>Polygon Vertices (Lat, Lng)</span>
                  <button type="button" (click)="clearPolygonVertices()" class="text-[10px] text-rose-400 hover:text-rose-200">Clear</button>
                </div>
                <p class="text-[10px] text-slate-400">One coordinate pair per line or click on the map to add vertices</p>
                <textarea rows="3" [(ngModel)]="polygonCoordsText" name="polygonCoordsText"
                  class="w-full p-2 rounded bg-slate-800 border border-slate-600 text-white font-mono text-[11px]"></textarea>
              </div>
            }

            <div>
              <label class="block text-slate-300 font-medium mb-1">Operational Reason / Sighting Evidence</label>
              <textarea rows="2" [(ngModel)]="newFence.reason" name="reason" placeholder="e.g. Sighting cluster response along harvest interface"
                class="w-full px-3 py-2 rounded-lg bg-slate-800 border border-slate-700 text-white focus:outline-none focus:border-purple-500"></textarea>
            </div>

            <div class="flex justify-end gap-3 pt-3 border-t border-slate-800">
              <button type="button" (click)="showAddModal.set(false)" class="px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300">
                Cancel
              </button>
              <button type="submit" class="px-5 py-2 rounded-lg bg-purple-600 hover:bg-purple-500 font-semibold text-white shadow">
                Save & Deploy Geofence
              </button>
            </div>
          </form>
        </div>
      </div>
    }
  `
})
export class GeofenceMapComponent implements OnInit {
  api = inject(AdminApiService);
  geocodingService = inject(GeocodingService);
  @ViewChild('infoWindow') infoWindow?: MapInfoWindow;

  // Viewport Centering
  center: google.maps.LatLngLiteral = { lat: 15.0, lng: 105.0 };
  zoom = 4;

  mapOptions: google.maps.MapOptions = {
    mapTypeId: 'roadmap',
    backgroundColor: '#0f172a',
    disableDefaultUI: false,
    zoomControl: true,
    mapTypeControl: true,
    streetViewControl: false,
    fullscreenControl: true,
    styles: [
      { elementType: 'geometry', stylers: [{ color: '#1e293b' }] },
      { elementType: 'labels.text.stroke', stylers: [{ color: '#0f172a' }] },
      { elementType: 'labels.text.fill', stylers: [{ color: '#94a3b8' }] },
      { featureType: 'water', elementType: 'geometry', stylers: [{ color: '#0f172a' }] },
      { featureType: 'road', elementType: 'geometry', stylers: [{ color: '#334155' }] }
    ]
  };

  // 1. Layer Visibility Toggles (Feature #61)
  showAiGeofences = true;
  showManualGeofences = true;
  showWildlife = true;
  showRoadkills = true;
  showEnvironmental = true;

  // 2. Temporal & Date Filtering (Feature #62)
  timeFilter = signal<TimeFilterMode>('all');
  customDateFrom = signal<string>('');
  customDateTo = signal<string>('');

  // 3. Secondary Filters (Feature #63)
  filterSpecies = 'all';
  filterThreatTier = 'all';

  // 4. Inspector Selection & Modals (Feature #63)
  selectedEntity = signal<SelectedEntity | null>(null);
  showAddModal = signal<boolean>(false);
  isDrawingMode = signal<boolean>(false);

  // Reverse geocoded location status & cache
  resolvedLocationName = signal<string>('');
  isResolvingLocation = signal<boolean>(false);

  // New Geofence Model
  newFence: Partial<AdminGeofence> = {
    name: '',
    geometry_type: 'circle',
    threat_level: 'TIER_2_ALERT',
    target_species: 'Asian Elephant',
    reason: ''
  };
  circleLat = 12.875;
  circleLng = 99.650;
  circleRadius = 3000;
  polygonCoordsText = '12.875, 99.630\n12.890, 99.670\n12.860, 99.695\n12.845, 99.650';

  // Drawing preview options
  readonly drawingCircleOptions: google.maps.CircleOptions = {
    fillColor: '#c084fc',
    fillOpacity: 0.35,
    strokeColor: '#a855f7',
    strokeOpacity: 1.0,
    strokeWeight: 3
  };
  readonly drawingCenterMarkerOptions: google.maps.MarkerOptions = {
    draggable: true,
    title: 'Geofence Center'
  };
  readonly drawingPolygonOptions: google.maps.PolygonOptions = {
    fillColor: '#c084fc',
    fillOpacity: 0.35,
    strokeColor: '#a855f7',
    strokeOpacity: 1.0,
    strokeWeight: 3
  };

  ngOnInit() {
    this.api.getGeofences().subscribe();
    this.api.getWildlife().subscribe();
    this.api.getRoadkills(true).subscribe();
    this.api.getEnvironmental().subscribe();
  }

  // --- Filtering Computed Getters ---

  aiCount(): number {
    return this.api.geofences().filter(g => g.source === 'ai').length;
  }

  manualCount(): number {
    return this.api.geofences().filter(g => g.source === 'manual').length;
  }

  hasFreshRoadkill(): boolean {
    return this.filteredRoadkills().some(r => r.decay_stage === 'FRESH');
  }

  totalVisibleIncidents(): number {
    return this.filteredWildlife().length + this.filteredRoadkills().length + this.filteredEnvironmental().length;
  }

  visibleGeofences(): AdminGeofence[] {
    const filtered = this.api.geofences().filter(g => {
      if (g.source === 'ai' && !this.showAiGeofences) return false;
      if (g.source === 'manual' && !this.showManualGeofences) return false;

      // Species filter
      if (this.filterSpecies !== 'all') {
        if (!g.target_species.toLowerCase().includes(this.filterSpecies.toLowerCase())) {
          return false;
        }
      }
      // Threat tier filter
      if (this.filterThreatTier !== 'all') {
        if (g.threat_level !== this.filterThreatTier) {
          return false;
        }
      }
      return true;
    });

    const tierOrder: Record<string, number> = {
      'TIER_1_CAUTION': 1,
      'TIER_2_ALERT': 2,
      'TIER_3_CRITICAL': 3
    };
    return filtered.sort((a, b) => (tierOrder[a.threat_level] || 0) - (tierOrder[b.threat_level] || 0));
  }

  filteredWildlife(): AdminWildlifeIncident[] {
    if (!this.showWildlife) return [];
    return this.api.wildlife().filter(w => {
      // Time window filter
      if (!this.isWithinTimeWindow(w.reported_at)) return false;

      // Species filter
      if (this.filterSpecies !== 'all') {
        if (!w.species.toLowerCase().includes(this.filterSpecies.toLowerCase())) {
          return false;
        }
      }
      // Threat tier filter
      if (this.filterThreatTier !== 'all') {
        if (w.threat_level !== this.filterThreatTier) {
          return false;
        }
      }
      return true;
    });
  }

  filteredRoadkills(): AdminRoadkillRecord[] {
    if (!this.showRoadkills) return [];
    return this.api.roadkills().filter(rk => {
      // Time window filter
      if (!this.isWithinTimeWindow(rk.reported_at)) return false;

      // Species filter
      if (this.filterSpecies !== 'all') {
        if (!rk.species.toLowerCase().includes(this.filterSpecies.toLowerCase())) {
          return false;
        }
      }
      return true;
    });
  }

  filteredEnvironmental(): AdminEnvironmentalData[] {
    if (!this.showEnvironmental) return [];
    return this.api.environmental().filter(env => {
      // Time window filter
      if (!this.isWithinTimeWindow(env.last_updated)) return false;
      return true;
    });
  }

  isWithinTimeWindow(isoDateStr?: string): boolean {
    if (!isoDateStr) return true;
    const mode = this.timeFilter();
    if (mode === 'all') return true;

    const itemTime = new Date(isoDateStr).getTime();
    if (isNaN(itemTime)) return true;

    const now = Date.now();
    if (mode === '24h') {
      return now - itemTime <= 24 * 60 * 60 * 1000;
    }
    if (mode === '7d') {
      return now - itemTime <= 7 * 24 * 60 * 60 * 1000;
    }
    if (mode === '30d') {
      return now - itemTime <= 30 * 24 * 60 * 60 * 1000;
    }
    if (mode === 'custom') {
      if (this.customDateFrom()) {
        const from = new Date(this.customDateFrom()).getTime();
        if (!isNaN(from) && itemTime < from) return false;
      }
      if (this.customDateTo()) {
        const to = new Date(this.customDateTo()).getTime() + 24 * 60 * 60 * 1000;
        if (!isNaN(to) && itemTime > to) return false;
      }
      return true;
    }
    return true;
  }

  get parsedPolygonCoords(): google.maps.LatLngLiteral[] {
    const lines = this.polygonCoordsText.split('\n');
    return lines
      .map(line => {
        const parts = line.split(',');
        return {
          lat: parseFloat(parts[0]?.trim() || '0'),
          lng: parseFloat(parts[1]?.trim() || '0')
        };
      })
      .filter(c => !isNaN(c.lat) && !isNaN(c.lng) && c.lat !== 0);
  }

  // --- Map-Click & Drawing Mode ---

  onMapClick(event: google.maps.MapMouseEvent) {
    if (!this.isDrawingMode() || !event.latLng) return;

    const clickedLat = Number(event.latLng.lat().toFixed(5));
    const clickedLng = Number(event.latLng.lng().toFixed(5));

    if (this.newFence.geometry_type === 'circle') {
      this.circleLat = clickedLat;
      this.circleLng = clickedLng;
      this.resolveReverseGeocode(clickedLat, clickedLng, true);
    } else {
      const newLine = `${clickedLat}, ${clickedLng}`;
      if (!this.polygonCoordsText.trim()) {
        this.polygonCoordsText = newLine;
      } else {
        this.polygonCoordsText += `\n${newLine}`;
      }
      this.resolveReverseGeocode(clickedLat, clickedLng, true);
    }
  }

  startInteractiveDrawing() {
    this.showAddModal.set(false);
    this.isDrawingMode.set(true);
  }

  finishDrawing() {
    this.isDrawingMode.set(false);
    this.showAddModal.set(true);
    // When finishing drawing, trigger reverse geocode for polygon or circle
    if (this.newFence.geometry_type === 'circle') {
      this.resolveReverseGeocode(this.circleLat, this.circleLng, true);
    } else {
      const coords = this.parsedPolygonCoords;
      if (coords.length > 0) {
        this.resolveReverseGeocode(coords[0].lat, coords[0].lng, true);
      }
    }
  }

  clearPolygonVertices() {
    this.polygonCoordsText = '';
  }

  // --- Reverse Geocoding Helpers for Geofence Naming ---

  resolveReverseGeocode(lat?: number, lng?: number, updateName = true) {
    const targetLat = lat !== undefined ? lat : Number(this.circleLat);
    const targetLng = lng !== undefined ? lng : Number(this.circleLng);
    if (targetLat === 0 && targetLng === 0) return;

    this.isResolvingLocation.set(true);
    this.geocodingService.reverseGeocode(targetLat, targetLng).subscribe({
      next: (placeName) => {
        this.resolvedLocationName.set(placeName);
        this.isResolvingLocation.set(false);
        if (updateName) {
          this.applyResolvedLocationToName(placeName);
        }
      },
      error: () => this.isResolvingLocation.set(false)
    });
  }

  applyResolvedLocationToName(place?: string) {
    const loc = place || this.resolvedLocationName();
    if (!loc) return;
    const species = this.newFence.target_species || 'Wildlife';
    const isPoly = this.newFence.geometry_type === 'polygon';
    const prefix = isPoly ? 'Polygon Shield' : 'Buffer';
    this.newFence.name = `${prefix}: ${species} (${loc})`;
  }

  // --- Entity Selection & One-Click Geofence Creation ---

  selectEntity(type: 'geofence' | 'wildlife' | 'roadkill' | 'environmental', data: any) {
    this.selectedEntity.set({ type, data });
  }

  selectAndPanTo(geo: AdminGeofence) {
    this.selectedEntity.set({ type: 'geofence', data: geo });
    if (geo.geometry_type === 'circle' && geo.circle_center) {
      this.center = { lat: geo.circle_center.lat, lng: geo.circle_center.lng };
      this.zoom = 11;
    } else if (geo.coordinates && geo.coordinates.length > 0) {
      this.center = { lat: geo.coordinates[0].lat, lng: geo.coordinates[0].lng };
      this.zoom = 11;
    }
  }

  createGeofenceFromEntity(ent: SelectedEntity) {
    if (ent.type === 'wildlife') {
      const w: AdminWildlifeIncident = ent.data;
      this.circleLat = w.location.lat;
      this.circleLng = w.location.lng;
      this.circleRadius = w.threat_level === 'TIER_3_CRITICAL' ? 4000 : 3000;
      this.newFence = {
        name: `Buffer: ${w.species} (${w.location.road_name || 'Resolving location...'})`,
        geometry_type: 'circle',
        threat_level: w.threat_level,
        target_species: w.species,
        reason: `Targeted defense response for empirical sighting on ${w.reported_at?.slice(0, 10)}: ${w.description || 'Reported by ' + w.reporter}`
      };
      this.resolveReverseGeocode(w.location.lat, w.location.lng, true);
    } else if (ent.type === 'roadkill') {
      const rk: AdminRoadkillRecord = ent.data;
      this.circleLat = rk.location.lat;
      this.circleLng = rk.location.lng;
      this.circleRadius = rk.decay_stage === 'FRESH' ? 4500 : 2500;
      this.newFence = {
        name: `Scent-Plume Zone: ${rk.species} Carcass (${rk.location.road_name || 'Resolving location...'})`,
        geometry_type: 'circle',
        threat_level: rk.decay_stage === 'FRESH' ? 'TIER_3_CRITICAL' : 'TIER_2_ALERT',
        target_species: rk.species.includes('Boar') || rk.species.includes('Deer') ? 'Sumatran Tiger' : 'Apex Carnivore',
        reason: `Prevent roadside ambush risk attracted by ${rk.decay_stage} ${rk.species} (~${rk.estimated_weight_kg}kg) carcass scent plume`
      };
      this.resolveReverseGeocode(rk.location.lat, rk.location.lng, true);
    } else if (ent.type === 'environmental') {
      const env: AdminEnvironmentalData = ent.data;
      const lat = env.circle_center?.lat || 12.875;
      const lng = env.circle_center?.lng || 99.650;
      this.circleLat = lat;
      this.circleLng = lng;
      this.circleRadius = env.radius_meters || 3000;
      this.newFence = {
        name: `Attractant Defense: ${env.name}`,
        geometry_type: 'circle',
        threat_level: env.risk_weight >= 0.85 ? 'TIER_3_CRITICAL' : 'TIER_2_ALERT',
        target_species: env.category === 'CROP_PHENOLOGY' ? 'Asian Elephant' : env.category === 'PERSIMMON_KAKI' ? 'Asian Black Bear' : 'Sumatran Tiger',
        reason: `Ecological crop/pasture attractant protection during active window: ${env.active_season} (Risk weight ${env.risk_weight})`
      };
      this.isResolvingLocation.set(true);
      this.geocodingService.reverseGeocode(lat, lng).subscribe({
        next: (placeName) => {
          this.resolvedLocationName.set(placeName);
          this.newFence.name = `Attractant Defense: ${env.name} (${placeName})`;
          this.isResolvingLocation.set(false);
        },
        error: () => this.isResolvingLocation.set(false)
      });
    }

    this.showAddModal.set(true);
  }

  resetToJapacView() {
    this.center = { lat: 15.0, lng: 105.0 };
    this.zoom = 4;
  }

  onRegionFocusChange(event: Event) {
    const val = (event.target as HTMLSelectElement).value;
    switch (val) {
      case 'srilanka':
        this.center = { lat: 7.8731, lng: 80.7718 };
        this.zoom = 8;
        break;
      case 'india':
        this.center = { lat: 26.6, lng: 93.3 };
        this.zoom = 8;
        break;
      case 'thailand':
        this.center = { lat: 12.87, lng: 99.65 };
        this.zoom = 9;
        break;
      case 'sumatra':
        this.center = { lat: 0.5, lng: 101.5 };
        this.zoom = 8;
        break;
      case 'japan':
        this.center = { lat: 39.7, lng: 141.15 };
        this.zoom = 7;
        break;
      default:
        this.resetToJapacView();
        break;
    }
  }

  deleteSelectedGeofence(id: string) {
    if (confirm('Delete this geofence?')) {
      this.api.deleteGeofence(id).subscribe();
      this.selectedEntity.set(null);
    }
  }

  confirmClearAllManual() {
    if (confirm('Are you sure you want to PURGE ALL MANUAL GEOFENCES? AI-generated fences will remain intact.')) {
      this.api.clearAllManualGeofences().subscribe();
    }
  }

  openAddModal() {
    this.newFence = {
      name: '',
      geometry_type: 'circle',
      threat_level: 'TIER_2_ALERT',
      target_species: 'Asian Elephant',
      reason: ''
    };
    this.showAddModal.set(true);
    this.resolveReverseGeocode(this.circleLat, this.circleLng, true);
  }

  saveManualGeofence() {
    const toSave: Partial<AdminGeofence> = {
      ...this.newFence,
      source: 'manual',
      is_persistent: true
    };

    if (toSave.geometry_type === 'circle') {
      toSave.circle_center = { lat: Number(this.circleLat), lng: Number(this.circleLng) };
      toSave.circle_radius_meters = Number(this.circleRadius);
    } else {
      toSave.coordinates = this.parsedPolygonCoords;
    }

    // Ensure geofence name contains reverse-geocoded location name
    const lat = toSave.circle_center?.lat || (toSave.coordinates && toSave.coordinates[0]?.lat) || Number(this.circleLat);
    const lng = toSave.circle_center?.lng || (toSave.coordinates && toSave.coordinates[0]?.lng) || Number(this.circleLng);
    const place = this.resolvedLocationName() || this.geocodingService.resolveRegionalKnowledge(lat, lng);

    if (!toSave.name || toSave.name.trim() === '' || toSave.name === 'Manual Geofence') {
      toSave.name = `Buffer: ${toSave.target_species || 'Zone'} (${place})`;
    } else if (!toSave.name.includes(place) && !toSave.name.includes('(')) {
      toSave.name = `${toSave.name} (${place})`;
    }

    this.api.createGeofence(toSave).subscribe({
      next: () => {
        this.showAddModal.set(false);
      }
    });
  }

  // --- Marker & Layer Styling Helpers ---

  getWildlifeMarkerOptions(w: AdminWildlifeIncident): google.maps.MarkerOptions {
    const color = w.threat_level === 'TIER_3_CRITICAL' ? '#ef4444' : w.threat_level === 'TIER_2_ALERT' ? '#f97316' : '#eab308';
    return {
      title: `[${w.threat_level.replace('TIER_', 'T')}] ${w.species} (${w.location.road_name || 'Encounter'})`,
      icon: {
        path: google.maps.SymbolPath.CIRCLE,
        scale: w.threat_level === 'TIER_3_CRITICAL' ? 8 : 6,
        fillColor: color,
        fillOpacity: 0.95,
        strokeColor: '#ffffff',
        strokeWeight: 2
      }
    };
  }

  getRoadkillMarkerOptions(rk: AdminRoadkillRecord): google.maps.MarkerOptions {
    const isFresh = rk.decay_stage === 'FRESH';
    return {
      title: `Roadkill: ${rk.species} (~${rk.estimated_weight_kg}kg, ${rk.decay_stage})`,
      icon: {
        path: google.maps.SymbolPath.CIRCLE,
        scale: isFresh ? 9 : 6,
        fillColor: isFresh ? '#f43f5e' : '#fb923c',
        fillOpacity: 0.95,
        strokeColor: isFresh ? '#fda4af' : '#ffffff',
        strokeWeight: isFresh ? 3 : 1.5
      }
    };
  }

  getEnvironmentalCircleOptions(env: AdminEnvironmentalData): google.maps.CircleOptions {
    let strokeColor = '#10b981';
    let fillColor = '#059669';

    if (env.category === 'PERSIMMON_KAKI') {
      strokeColor = '#f59e0b';
      fillColor = '#d97706';
    } else if (env.category === 'LIVESTOCK_PASTURE') {
      strokeColor = '#8b5cf6';
      fillColor = '#7c3aed';
    } else if (env.category === 'WATER_SOURCE') {
      strokeColor = '#06b6d4';
      fillColor = '#0891b2';
    }

    return {
      fillColor,
      fillOpacity: 0.22,
      strokeColor,
      strokeOpacity: 0.85,
      strokeWeight: 2,
      clickable: true
    };
  }

  getEnvironmentalCenterMarkerOptions(env: AdminEnvironmentalData): google.maps.MarkerOptions {
    return {
      title: `Attractant: ${env.name} (${env.category})`,
      icon: {
        path: google.maps.SymbolPath.CIRCLE,
        scale: 4,
        fillColor: '#10b981',
        fillOpacity: 0.9,
        strokeColor: '#ffffff',
        strokeWeight: 1
      }
    };
  }

  getPolygonOptions(geo: AdminGeofence): google.maps.PolygonOptions {
    const isManual = geo.source === 'manual';
    const isBuffer = geo.id.includes('-tier2-buffer') || geo.id.includes('-tier1-buffer');
    let fillColor = '#facc15';
    let strokeColor = '#eab308';
    let zIndex = 10;
    let fillOpacity = isBuffer ? 0.10 : 0.16;

    if (geo.threat_level === 'TIER_3_CRITICAL') {
      fillColor = '#ef4444';
      strokeColor = '#dc2626';
      zIndex = 30;
      fillOpacity = 0.32;
    } else if (geo.threat_level === 'TIER_2_ALERT') {
      fillColor = '#f97316';
      strokeColor = '#ea580c';
      zIndex = 20;
      fillOpacity = isBuffer ? 0.16 : 0.22;
    }

    if (isManual) {
      strokeColor = '#a855f7';
    }

    return {
      fillColor,
      fillOpacity,
      strokeColor,
      strokeOpacity: isBuffer ? 0.70 : 0.90,
      strokeWeight: isManual ? 3 : (isBuffer ? 1.5 : 2.5),
      zIndex,
      clickable: true
    };
  }

  getCircleOptions(geo: AdminGeofence): google.maps.CircleOptions {
    const isManual = geo.source === 'manual';
    const isBuffer = geo.id.includes('-tier2-buffer') || geo.id.includes('-tier1-buffer');
    let fillColor = '#facc15';
    let strokeColor = '#eab308';
    let zIndex = 10;
    let fillOpacity = isBuffer ? 0.10 : 0.16;

    if (geo.threat_level === 'TIER_3_CRITICAL') {
      fillColor = '#ef4444';
      strokeColor = '#dc2626';
      zIndex = 30;
      fillOpacity = 0.32;
    } else if (geo.threat_level === 'TIER_2_ALERT') {
      fillColor = '#f97316';
      strokeColor = '#ea580c';
      zIndex = 20;
      fillOpacity = isBuffer ? 0.16 : 0.22;
    }

    if (isManual) {
      strokeColor = '#a855f7';
    }

    return {
      fillColor,
      fillOpacity,
      strokeColor,
      strokeOpacity: isBuffer ? 0.70 : 0.90,
      strokeWeight: isManual ? 3 : (isBuffer ? 1.5 : 2.5),
      zIndex,
      clickable: true
    };
  }

  getBadgeClass(tier: ThreatLevel): string {
    if (tier === 'TIER_3_CRITICAL') return 'bg-rose-950 text-rose-400 border border-rose-800';
    if (tier === 'TIER_2_ALERT') return 'bg-amber-950 text-amber-400 border border-amber-800';
    return 'bg-yellow-950 text-yellow-400 border border-yellow-800';
  }

  getEntityTypeName(type: string): string {
    if (type === 'wildlife') return 'Wildlife Sighting';
    if (type === 'roadkill') return 'Roadkill Carcass';
    if (type === 'environmental') return 'Attractant Zone';
    return 'Active Geofence';
  }

  getEntityTitle(ent: SelectedEntity): string {
    if (ent.type === 'wildlife') return `${ent.data.species} Encounter`;
    if (ent.type === 'roadkill') return `${ent.data.species} Carcass`;
    if (ent.type === 'environmental') return ent.data.name;
    return ent.data.name;
  }

  getEntityBadgeClass(type: string): string {
    if (type === 'wildlife') return 'bg-cyan-950 text-cyan-300 border border-cyan-800';
    if (type === 'roadkill') return 'bg-rose-950 text-rose-300 border border-rose-800';
    if (type === 'environmental') return 'bg-emerald-950 text-emerald-300 border border-emerald-800';
    return 'bg-purple-950 text-purple-300 border border-purple-800';
  }
}
