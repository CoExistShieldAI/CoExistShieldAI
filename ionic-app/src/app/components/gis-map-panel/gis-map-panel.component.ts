import { Component, inject, signal, computed, OnInit, HostListener } from '@angular/core';
import { CommonModule } from '@angular/common';
import { GoogleMapsModule } from '@angular/google-maps';
import { MAP_REGIONS, ShieldStateService } from '../../services/shield-state.service';
import { FoldableDeviceService } from '../../services/foldable-device.service';
import {
  AlertTier,
  AttractantMarkerItem,
  GeofenceZone,
  MapLayerType,
  RoadkillRecord
} from '../../models/shield.model';

interface RenderedGeofenceOverlay {
  zone: GeofenceZone;
  hasPolygon: boolean;
  paths: google.maps.LatLngLiteral[];
  polygonOptions: google.maps.PolygonOptions;
  center: google.maps.LatLngLiteral;
  circleOptions: google.maps.CircleOptions;
  markerOptions: google.maps.MarkerOptions;
}

interface RenderedRoadkillOverlay {
  record: RoadkillRecord;
  position: google.maps.LatLngLiteral;
  markerOptions: google.maps.MarkerOptions;
}

interface RenderedAttractantOverlay {
  item: AttractantMarkerItem;
  position: google.maps.LatLngLiteral;
  circleOptions: google.maps.CircleOptions;
  markerOptions: google.maps.MarkerOptions;
}

@Component({
  selector: 'app-gis-map-panel',
  standalone: true,
  imports: [CommonModule, GoogleMapsModule],
  styles: [
    `
      :host {
        display: block;
        width: 100%;
        height: 100%;
        position: relative;
      }
    `
  ],
  template: `
    <div class="relative w-full h-full min-h-[420px] bg-slate-950 overflow-hidden flex flex-col">
      <!-- Top Floating Overlay: Region Selector, Layer Toggles & GPS Simulator -->
      <div class="absolute top-3 left-3 right-3 z-20 flex flex-col gap-2 pointer-events-none">
        <!-- Row 1: Region Quick-Jump Tabs & GPS Simulation Toggle -->
        <div class="flex flex-wrap items-center justify-between gap-2 pointer-events-auto">
          <div class="flex items-center gap-1 overflow-x-auto bg-slate-950/90 backdrop-blur border border-slate-800 p-1 rounded-xl shadow-xl">
            @for (reg of regions; track reg.id) {
              <button
                type="button"
                (click)="state.jumpToRegion(reg.id)"
                class="px-2.5 py-1.5 rounded-lg text-xs font-bold whitespace-nowrap transition-all"
                [ngClass]="
                  state.selectedRegionId() === reg.id && state.mapZoom() >= 11
                    ? 'bg-sky-600 text-white shadow'
                    : 'text-slate-300 hover:bg-slate-800/80'
                ">
                {{ reg.label }}
              </button>
            }
            <button
              type="button"
              (click)="state.setNationwideOverview('JAPAN')"
              class="px-2.5 py-1.5 rounded-lg text-[11px] font-bold whitespace-nowrap text-sky-300 hover:bg-slate-800/80 border-l border-slate-800">
              🗾 Japan Wide
            </button>
            <button
              type="button"
              (click)="state.setNationwideOverview('APAC')"
              class="px-2.5 py-1.5 rounded-lg text-[11px] font-bold whitespace-nowrap text-emerald-300 hover:bg-slate-800/80">
              🌏 APAC Wide
            </button>
          </div>

          <!-- Interactive GPS Simulation Control (Judge & Field Demo Feature) -->
          <div class="flex items-center gap-1.5 bg-slate-950/90 backdrop-blur border border-slate-800 p-1 rounded-xl shadow-xl">
            <button
              type="button"
              (click)="state.isGpsSimulateClickMode.set(!state.isGpsSimulateClickMode())"
              class="px-2.5 py-1.5 rounded-lg text-[11px] font-bold flex items-center gap-1.5 transition-all"
              [ngClass]="
                state.isGpsSimulateClickMode()
                  ? 'bg-emerald-600 text-white'
                  : 'bg-slate-900 text-slate-400 hover:text-white'
              "
              title="When active, clicking anywhere on Google Maps moves your simulated device GPS and evaluates geofences">
              <span>🎯</span>
              <span>{{ state.isGpsSimulateClickMode() ? 'Click Map = Move GPS (ON)' : 'Click Map = Move GPS (OFF)' }}</span>
            </button>
            <button
              type="button"
              (click)="state.requestRealDeviceGps()"
              class="px-2.5 py-1.5 rounded-lg bg-slate-900 hover:bg-slate-800 text-sky-400 border border-slate-700 text-[11px] font-bold"
              title="Use real device GPS">
              📡 Live GPS
            </button>
          </div>

          <!-- Wide Map / Split View Toggle (Foldable & Desktop) -->
          <div class="flex items-center gap-1.5 bg-slate-950/90 backdrop-blur border border-slate-800 p-1 rounded-xl shadow-xl">
            <button
              type="button"
              (click)="foldable.toggleMapMaximize()"
              class="px-2.5 py-1.5 rounded-lg text-[11px] font-bold flex items-center gap-1.5 transition-all"
              [ngClass]="
                foldable.isMapMaximized()
                  ? 'bg-sky-600 text-white shadow ring-1 ring-sky-400'
                  : 'bg-slate-900 text-sky-400 hover:text-white'
              "
              [title]="foldable.isMapMaximized() ? 'Switch to Dual Screen Split View (左右2画面表示に戻す)' : 'Expand Map to Full Width (マップ全画面・広域表示)'">
              <span>{{ foldable.isMapMaximized() ? '◧' : '⛶' }}</span>
              <span>{{ foldable.isMapMaximized() ? 'Split View (◧)' : 'Wide Map (⛶)' }}</span>
            </button>
          </div>
        </div>

        <!-- Row 2: Layer Filter Chips -->
        <div class="flex items-center gap-1.5 overflow-x-auto pointer-events-auto py-0.5">
          @for (layer of layerOptions; track layer.id) {
            <button
              type="button"
              (click)="state.toggleMapLayer(layer.id)"
              class="px-2.5 py-1 rounded-full text-[11px] font-semibold whitespace-nowrap border backdrop-blur transition-all flex items-center gap-1.5 shadow"
              [ngClass]="
                state.activeLayers().has(layer.id)
                  ? 'bg-slate-900/95 border-sky-500 text-white'
                  : 'bg-slate-950/75 border-slate-800 text-slate-500'
              ">
              <span>{{ layer.emoji }}</span>
              <span>{{ layer.label }}</span>
            </button>
          }
        </div>

        <!-- Row 3: Proximity Roadkill Warning Banner (< 800m to decomposing carcass) -->
        @if (!warningBannerDismissed() && activeRoadkillAlert(); as rkAlert) {
          <div
            (click)="selectRoadkill(rkAlert)"
            class="pointer-events-auto cursor-pointer p-3 rounded-xl bg-amber-950/95 backdrop-blur border-2 border-amber-500 shadow-2xl flex items-center justify-between gap-3">
            <div class="flex items-center gap-2.5">
              <span class="text-xl animate-pulse">🐗</span>
              <div>
                <div class="text-xs font-black text-amber-300 uppercase tracking-wide">
                  PROXIMITY ALERT: Decomposing {{ rkAlert.species }} Carcass (#{{ rkAlert.id }})
                </div>
                <div class="text-[11px] text-amber-100">
                  Olfactory Lure Score {{ rkAlert.attractantThreatScore }} • Predator ambush risk within 800m radius
                </div>
              </div>
            </div>
            <button
              type="button"
              (click)="$event.stopPropagation(); warningBannerDismissed.set(true)"
              class="text-amber-300 hover:text-white px-2 py-1 text-xs font-bold">
              ✕
            </button>
          </div>
        }
      </div>

      <!-- Floating Restore Split View Pill when Map is Maximized -->
      @if (foldable.isMapMaximized()) {
        <div class="absolute top-28 left-4 z-30 pointer-events-auto">
          <button
            type="button"
            (click)="foldable.setMapMaximized(false)"
            class="px-3.5 py-2 rounded-xl bg-sky-600/95 hover:bg-sky-500 text-white text-xs font-bold shadow-2xl flex items-center gap-2 border border-sky-400 transition-all hover:scale-105 backdrop-blur">
            <span>◧</span>
            <span>Restore Split View (左右2画面に戻す)</span>
          </button>
        </div>
      }


      <!-- Main Google Map Canvas -->
      <div class="flex-1 w-full h-full min-h-[420px] relative overflow-hidden">
        @if (googleMapsReady()) {
          <div class="absolute inset-0 w-full h-full">
            <google-map
              height="100%"
              width="100%"
              [center]="state.mapCenter()"
              [zoom]="state.mapZoom()"
              [options]="mapOptions"
              (mapClick)="onGoogleMapClick($event)">

              <!-- 1. Geofence Polygons & Radius Circles -->
              @if (state.activeLayers().has('GEOFENCES')) {
                @for (gf of renderedGeofences(); track gf.zone.id) {
                  @if (gf.hasPolygon) {
                    <map-polygon
                      [paths]="gf.paths"
                      [options]="gf.polygonOptions"
                      (polygonClick)="onPolygonClick($event, gf.zone)">
                    </map-polygon>
                  }
                  <map-circle
                    [center]="gf.center"
                    [radius]="gf.zone.radiusMeters"
                    [options]="gf.circleOptions"
                    (circleClick)="onCircleClick($event, gf.zone)">
                  </map-circle>
                  <map-marker
                    [position]="gf.center"
                    [options]="gf.markerOptions"
                    (mapClick)="selectGeofence(gf.zone)">
                  </map-marker>
                }
              }

              <!-- 2. Roadkill Carcass Markers & 800m Scent Plume Circles -->
              @if (state.activeLayers().has('ROADKILL')) {
                @for (rk of renderedRoadkills(); track rk.record.id) {
                  @if (rk.record.clearanceStatus !== 'CLEARED') {
                    <map-circle
                      [center]="rk.position"
                      [radius]="800"
                      [options]="roadkillCircleOptions"
                      (circleClick)="selectRoadkill(rk.record)">
                    </map-circle>
                  }
                  <map-marker
                    [position]="rk.position"
                    [options]="rk.markerOptions"
                    (mapClick)="selectRoadkill(rk.record)">
                  </map-marker>
                }
              }

              <!-- 3. Regional Attractants (Kaki Trees, Elephant Crops, Tiger Kill-Sites) -->
              @for (att of renderedAttractants(); track att.item.id) {
                @if (att.item.radiusMeters) {
                  <map-circle
                    [center]="att.position"
                    [radius]="att.item.radiusMeters"
                    [options]="att.circleOptions"
                    (circleClick)="selectAttractant(att.item)">
                  </map-circle>
                }
                <map-marker
                  [position]="att.position"
                  [options]="att.markerOptions"
                  (mapClick)="selectAttractant(att.item)">
                </map-marker>
              }

              <!-- 4. Current / Simulated User Device GPS Marker (Draggable for instant boundary tests) -->
              <map-circle
                [center]="renderedUserGps()"
                [radius]="250"
                [options]="userGpsHaloOptions">
              </map-circle>
              <map-marker
                [position]="renderedUserGps()"
                [options]="userGpsMarkerOptions()"
                (mapDragend)="onUserGpsDragEnd($event)">
              </map-marker>
            </google-map>
          </div>
        } @else {
          <!-- Interactive Fallback Tactical Map Canvas if Google Maps JS SDK hasn't loaded yet -->
          <div class="w-full h-full bg-slate-900 flex flex-col items-center justify-center p-6 text-center space-y-4">
            <div class="text-4xl">🗺️</div>
            <div class="text-base font-bold text-white">Loading Google Maps GIS Canvas...</div>
            <p class="text-xs text-slate-400 max-w-md">
              Initializing Google Maps JavaScript SDK. Use the quick region buttons below to test spatial geofence transitions:
            </p>
            <div class="flex flex-wrap justify-center gap-2">
              @for (gf of state.geofences(); track gf.id) {
                <button
                  type="button"
                  (click)="simulateEnterZone(gf)"
                  class="px-3 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 border border-slate-700 text-xs font-bold text-sky-300">
                  📍 Enter {{ gf.name }} ({{ gf.tier }})
                </button>
              }
            </div>
          </div>
        }
      </div>

      <!-- Bottom Left: Legend & Simulated GPS Coordinates Pill -->
      <div class="absolute bottom-4 left-4 z-10 pointer-events-auto hidden sm:flex items-center gap-3 px-3.5 py-2 rounded-xl bg-slate-950/90 backdrop-blur border border-slate-800 text-[11px] shadow-xl">
        <div class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-red-500"></span>
          <span class="text-slate-300">Tier 3 Danger</span>
        </div>
        <div class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-amber-500"></span>
          <span class="text-slate-300">Tier 2 Warning</span>
        </div>
        <div class="flex items-center gap-1.5">
          <span class="w-2.5 h-2.5 rounded-full bg-sky-400"></span>
          <span class="text-slate-300">Tier 1 Notice</span>
        </div>
        <div class="border-l border-slate-800 pl-3 font-mono text-emerald-400">
          GPS: {{ state.userLocation().latitude.toFixed(4) }}, {{ state.userLocation().longitude.toFixed(4) }}
        </div>
      </div>

      <!-- Interactive GPS Simulation D-Pad / Steering Controller -->
      @if (state.isGpsSimulateClickMode()) {
        <div
          class="absolute right-4 z-40 pointer-events-auto flex flex-col items-end gap-2 transition-all duration-200"
          [ngClass]="state.selectedEntity() ? 'bottom-4 md:bottom-[280px]' : 'bottom-4'">
          @if (state.isDpadVisible()) {
            <div class="bg-slate-950/95 backdrop-blur-md border border-slate-700/80 rounded-2xl p-3 shadow-2xl flex flex-col items-center gap-2 select-none w-44 transition-all">
              <!-- Header -->
              <div class="w-full flex items-center justify-between border-b border-slate-800/80 pb-1.5 text-[11px]">
                <div class="flex items-center gap-1.5 font-bold text-sky-400">
                  <span>🎮</span>
                  <span>GPS Steering</span>
                </div>
                <button
                  type="button"
                  (click)="state.isDpadVisible.set(false)"
                  class="text-slate-400 hover:text-white px-1.5 py-0.5 rounded text-xs font-bold leading-none">
                  ✕
                </button>
              </div>

              <!-- Step Size Selector -->
              <div class="w-full flex items-center justify-between gap-1 bg-slate-900/90 p-1 rounded-lg text-[10px] font-bold">
                @for (step of [10, 50, 200]; track step) {
                  <button
                    type="button"
                    (click)="state.setGpsStepSize(step)"
                    class="flex-1 py-0.5 rounded text-center transition-all"
                    [ngClass]="state.gpsStepSizeMeters() === step ? 'bg-sky-600 text-white shadow' : 'text-slate-400 hover:text-slate-200'">
                    {{ step }}m
                  </button>
                }
              </div>

              <!-- Directional Buttons -->
              <div class="grid grid-cols-3 gap-1.5 my-1">
                <div></div>
                <button
                  type="button"
                  (click)="state.stepSimulatedLocation('NORTH')"
                  class="w-10 h-10 rounded-xl bg-slate-800 hover:bg-sky-600 active:scale-95 text-slate-100 font-black text-sm flex items-center justify-center border border-slate-700 hover:border-sky-500 shadow transition-all"
                  title="Step North (ArrowUp / W)">
                  ▲
                </button>
                <div></div>

                <button
                  type="button"
                  (click)="state.stepSimulatedLocation('WEST')"
                  class="w-10 h-10 rounded-xl bg-slate-800 hover:bg-sky-600 active:scale-95 text-slate-100 font-black text-sm flex items-center justify-center border border-slate-700 hover:border-sky-500 shadow transition-all"
                  title="Step West (ArrowLeft / A)">
                  ◀
                </button>
                <button
                  type="button"
                  (click)="recenterToUserGps()"
                  class="w-10 h-10 rounded-xl bg-slate-900 hover:bg-slate-800 active:scale-95 text-sky-400 text-xs flex flex-col items-center justify-center border border-slate-800 shadow transition-all"
                  title="Recenter Map View to GPS">
                  🎯
                </button>
                <button
                  type="button"
                  (click)="state.stepSimulatedLocation('EAST')"
                  class="w-10 h-10 rounded-xl bg-slate-800 hover:bg-sky-600 active:scale-95 text-slate-100 font-black text-sm flex items-center justify-center border border-slate-700 hover:border-sky-500 shadow transition-all"
                  title="Step East (ArrowRight / D)">
                  ▶
                </button>

                <div></div>
                <button
                  type="button"
                  (click)="state.stepSimulatedLocation('SOUTH')"
                  class="w-10 h-10 rounded-xl bg-slate-800 hover:bg-sky-600 active:scale-95 text-slate-100 font-black text-sm flex items-center justify-center border border-slate-700 hover:border-sky-500 shadow transition-all"
                  title="Step South (ArrowDown / S)">
                  ▼
                </button>
                <div></div>
              </div>

              <!-- Hotkey Hint -->
              <div class="text-[9px] text-slate-400 font-mono tracking-tight text-center">
                WASD / ↑↓←→ keys supported
              </div>
            </div>
          } @else {
            <!-- Minimized Trigger Pill -->
            <button
              type="button"
              (click)="state.isDpadVisible.set(true)"
              class="px-3 py-2 rounded-xl bg-slate-950/95 hover:bg-sky-950 border border-sky-500/70 text-white shadow-2xl flex items-center gap-1.5 text-xs font-bold transition-all hover:scale-105"
              title="Open GPS Steering D-Pad">
              <span>🎮</span>
              <span class="text-sky-300">D-Pad</span>
            </button>
          }
        </div>
      }

      <!-- Selected Entity Details Bottom Card Overlay -->
      @if (state.selectedEntity(); as entity) {
        <div class="absolute bottom-4 left-4 right-4 md:left-auto md:right-4 md:w-[440px] z-30 p-4 rounded-2xl bg-slate-950/95 backdrop-blur border border-slate-700 shadow-2xl space-y-3">
          <div class="flex items-start justify-between gap-2">
            <div>
              <div class="flex items-center gap-2">
                <span
                  class="text-[10px] font-black px-2 py-0.5 rounded uppercase"
                  [ngClass]="{
                    'bg-red-950 text-red-400 border border-red-800': entity.threatTier === 'TIER_3_DANGER',
                    'bg-amber-950 text-amber-400 border border-amber-800': entity.threatTier === 'TIER_2_WARNING',
                    'bg-sky-950 text-sky-400 border border-sky-800': entity.threatTier === 'TIER_1_NOTICE',
                    'bg-emerald-950 text-emerald-400 border border-emerald-800': entity.threatTier === 'SAFE'
                  }">
                  {{ entity.threatTier }}
                </span>
                <span class="text-[11px] font-mono text-slate-400">#{{ entity.id }}</span>
              </div>
              <h4 class="text-sm font-extrabold text-white mt-1">{{ entity.title }}</h4>
            </div>
            <button
              type="button"
              (click)="state.selectedEntity.set(null)"
              class="text-slate-400 hover:text-white px-2 py-0.5 text-sm font-bold">
              ✕
            </button>
          </div>

          <p class="text-xs text-slate-300 leading-relaxed">{{ entity.subtitle }}</p>

          @if (entity.geminiBadgeText) {
            <div class="px-3 py-1.5 rounded-lg bg-sky-950/60 border border-sky-800/70 text-[11px] font-semibold text-sky-300">
              ✨ {{ entity.geminiBadgeText }}
            </div>
          }

          <!-- Action Buttons: Simulate GPS Here & Official Mitigation Actions -->
          <div class="flex flex-wrap items-center gap-2 pt-1">
            <button
              type="button"
              (click)="state.updateDeviceLocation({ latitude: entity.latitude, longitude: entity.longitude }, 'Entity Pin GPS')"
              class="px-3 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 border border-slate-600 text-xs font-bold text-sky-300 flex-1">
              📍 Simulate GPS Here
            </button>

            @if (entity.isRoadkill && entity.roadkillStatus !== 'CLEARED') {
              @if (entity.roadkillStatus === 'REPORTED') {
                <button
                  type="button"
                  (click)="state.updateRoadkillStatus(entity.id, 'DISPATCHED')"
                  class="px-3 py-2 rounded-xl bg-amber-600 hover:bg-amber-500 text-xs font-bold text-white flex-1">
                  🚛 Dispatch Sanitation
                </button>
              }
              <button
                type="button"
                (click)="state.updateRoadkillStatus(entity.id, 'CLEARED')"
                class="px-3 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-xs font-bold text-white flex-1">
                ✅ Mark Carcass Cleared
              </button>
            } @else if (
              entity.canMarkResolved &&
              (state.currentRole() === 'OFFICIAL' || state.currentRole() === 'ADMIN')
            ) {
              <button
                type="button"
                (click)="state.resolveAttractantEntity(entity.id)"
                class="px-3 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-xs font-bold text-white flex-1">
                ✅ Execute Official Clearance
              </button>
            }
          </div>
        </div>
      }
    </div>
  `
})
export class GisMapPanelComponent implements OnInit {
  state = inject(ShieldStateService);
  foldable = inject(FoldableDeviceService);

  readonly regions = Object.values(MAP_REGIONS);

  readonly layerOptions: { id: MapLayerType; label: string; emoji: string }[] = [
    { id: 'GEOFENCES', label: 'Dynamic Geofences', emoji: '🛡️' },
    { id: 'ROADKILL', label: 'Roadkill', emoji: '🐗' },
    { id: 'PERSIMMON_TREES', label: 'Kaki Trees (JP)', emoji: '🍊' },
    { id: 'ELEPHANT_CROPS', label: 'Elephant Crops (TH)', emoji: '🌾' },
    { id: 'LIVESTOCK_PREDATION', label: 'Tiger Kills (ID)', emoji: '🐅' }
  ];

  warningBannerDismissed = signal<boolean>(false);
  googleMapsReady = signal<boolean>(false);
  private iconCache = new Map<string, string>();

  readonly mapOptions: google.maps.MapOptions = {
    mapTypeId: 'roadmap',
    backgroundColor: '#0f172a',
    disableDefaultUI: false,
    clickableIcons: false,
    zoomControl: true,
    mapTypeControl: true,
    streetViewControl: false,
    fullscreenControl: true,
    styles: [
      { elementType: 'geometry', stylers: [{ color: '#1e293b' }] },
      { elementType: 'labels.text.stroke', stylers: [{ color: '#0f172a' }] },
      { elementType: 'labels.text.fill', stylers: [{ color: '#94a3b8' }] },
      { featureType: 'water', elementType: 'geometry', stylers: [{ color: '#090d16' }] },
      { featureType: 'road', elementType: 'geometry', stylers: [{ color: '#334155' }] },
      { featureType: 'poi.park', elementType: 'geometry', stylers: [{ color: '#14261c' }] },
      {
        featureType: 'poi',
        elementType: 'all',
        stylers: [{ saturation: -100 }, { lightness: -40 }]
      },
      {
        featureType: 'poi',
        elementType: 'labels.icon',
        stylers: [{ saturation: -100 }, { lightness: -45 }]
      },
      {
        featureType: 'poi',
        elementType: 'labels.text.fill',
        stylers: [{ color: '#475569' }]
      },
      {
        featureType: 'poi',
        elementType: 'labels.text.stroke',
        stylers: [{ color: '#0f172a' }]
      },
      {
        featureType: 'transit',
        elementType: 'all',
        stylers: [{ saturation: -100 }, { lightness: -40 }]
      },
      {
        featureType: 'transit',
        elementType: 'labels.icon',
        stylers: [{ saturation: -100 }, { lightness: -45 }]
      },
      {
        featureType: 'transit',
        elementType: 'labels.text.fill',
        stylers: [{ color: '#475569' }]
      }
    ]
  };

  readonly roadkillCircleOptions: google.maps.CircleOptions = {
    fillColor: '#f59e0b',
    fillOpacity: 0.18,
    strokeColor: '#f59e0b',
    strokeOpacity: 0.85,
    strokeWeight: 2
  };

  readonly userGpsHaloOptions: google.maps.CircleOptions = {
    fillColor: '#38bdf8',
    fillOpacity: 0.25,
    strokeColor: '#38bdf8',
    strokeOpacity: 0.9,
    strokeWeight: 2,
    clickable: false
  };

  readonly userGpsMarkerOptions = computed<google.maps.MarkerOptions>(() => ({
    title: 'Your Device GPS Location (Drag to move)',
    draggable: this.state.isGpsSimulateClickMode(),
    icon: this.buildBadgeSvgIcon('📍', '#0284c7', '#38bdf8', 'YOU')
  }));

  readonly renderedUserGps = computed<google.maps.LatLngLiteral>(() => {
    const loc = this.state.userLocation();
    return { lat: loc.latitude, lng: loc.longitude };
  });

  readonly renderedGeofences = computed<RenderedGeofenceOverlay[]>(() => {
    const tierOrder: Record<AlertTier, number> = {
      SAFE: 0,
      TIER_1_NOTICE: 1,
      TIER_2_WARNING: 2,
      TIER_3_DANGER: 3
    };
    // Sort so outermost buffer zones (Tier 1) render first (base layer),
    // and core danger zones (Tier 3) render on top with higher zIndex and interactive priority.
    const sorted = [...this.state.geofences()].sort(
      (a, b) => (tierOrder[a.tier] || 0) - (tierOrder[b.tier] || 0)
    );

    return sorted.map((gf) => {
      const color = this.getTierColor(gf.tier);
      const hasPolygon = Array.isArray(gf.coordinates) && gf.coordinates.length >= 3;
      const emoji = gf.targetSpecies.includes('ELEPHANT')
        ? '🐘'
        : gf.targetSpecies.includes('TIGER')
        ? '🐅'
        : gf.targetSpecies.includes('BEAR')
        ? '🐻'
        : '🛡️';
      const shortTier = gf.tier
        .replace('TIER_', 'T')
        .replace('_DANGER', '')
        .replace('_WARNING', '')
        .replace('_NOTICE', '');

      const zIndex = gf.tier === 'TIER_3_DANGER' ? 30 : gf.tier === 'TIER_2_WARNING' ? 20 : 10;
      const fillOpacity =
        gf.tier === 'TIER_3_DANGER'
          ? (hasPolygon ? 0.28 : 0.26)
          : gf.tier === 'TIER_2_WARNING'
          ? (hasPolygon ? 0.18 : 0.16)
          : (hasPolygon ? 0.10 : 0.08);

      return {
        zone: gf,
        hasPolygon,
        paths: hasPolygon
          ? gf.coordinates.map((c) => ({ lat: c.latitude, lng: c.longitude }))
          : [],
        polygonOptions: {
          fillColor: color,
          fillOpacity,
          strokeColor: color,
          strokeOpacity: gf.tier === 'TIER_3_DANGER' ? 0.95 : 0.75,
          strokeWeight: gf.tier === 'TIER_3_DANGER' ? 2.5 : 1.8,
          zIndex
        },
        center: { lat: gf.centroid.latitude, lng: gf.centroid.longitude },
        circleOptions: {
          fillColor: color,
          fillOpacity,
          strokeColor: color,
          strokeOpacity: gf.tier === 'TIER_3_DANGER' ? 0.90 : 0.70,
          strokeWeight: gf.tier === 'TIER_3_DANGER' ? 2.5 : 1.5,
          zIndex
        },
        markerOptions: {
          title: gf.name,
          zIndex: zIndex + 5,
          icon: this.buildBadgeSvgIcon(emoji, '#0f172a', color, shortTier)
        }
      };
    });
  });

  readonly renderedRoadkills = computed<RenderedRoadkillOverlay[]>(() => {
    return this.state.roadkills().map((rk) => {
      const stroke = rk.clearanceStatus === 'CLEARED' ? '#10b981' : '#f59e0b';
      return {
        record: rk,
        position: { lat: rk.location.latitude, lng: rk.location.longitude },
        markerOptions: {
          title: `Roadkill #${rk.id} (${rk.species})`,
          icon: this.buildBadgeSvgIcon('🐗', '#1e1b4b', stroke, 'RK')
        }
      };
    });
  });

  readonly renderedAttractants = computed<RenderedAttractantOverlay[]>(() => {
    const layers = this.state.activeLayers();
    return this.state
      .attractantMarkers()
      .filter((m) => !m.isResolved && layers.has(m.layerType))
      .map((att) => {
        const color = this.getTierColor(att.threatTier);
        return {
          item: att,
          position: { lat: att.lat, lng: att.lon },
          circleOptions: {
            fillColor: color,
            fillOpacity: 0.16,
            strokeColor: color,
            strokeOpacity: 0.8,
            strokeWeight: 1.5
          },
          markerOptions: {
            title: att.title,
            icon: this.buildBadgeSvgIcon(att.iconEmoji, '#0f172a', color)
          }
        };
      });
  });

  readonly activeRoadkillAlert = computed(() => {
    if (!this.state.activeLayers().has('ROADKILL')) return null;
    return (
      this.state
        .roadkills()
        .find((rk) => rk.clearanceStatus !== 'CLEARED' && rk.attractantThreatScore >= 0.7) || null
    );
  });

  ngOnInit() {
    this.ensureGoogleMapsLoaded();
  }

  private async ensureGoogleMapsLoaded() {
    if (typeof window === 'undefined') return;

    const g = (window as any).google;
    if (typeof g?.maps?.Map === 'function') {
      this.googleMapsReady.set(true);
      return;
    }

    if (typeof g?.maps?.importLibrary === 'function') {
      try {
        await g.maps.importLibrary('maps');
        this.googleMapsReady.set(true);
        return;
      } catch {
        // Fall through to polling
      }
    }

    const interval = setInterval(async () => {
      const gm = (window as any).google?.maps;
      if (typeof gm?.Map === 'function') {
        this.googleMapsReady.set(true);
        clearInterval(interval);
      } else if (typeof gm?.importLibrary === 'function') {
        try {
          await gm.importLibrary('maps');
          if (typeof (window as any).google?.maps?.Map === 'function') {
            this.googleMapsReady.set(true);
            clearInterval(interval);
          }
        } catch {
          // Wait for next tick
        }
      }
    }, 150);
    setTimeout(() => clearInterval(interval), 15000);
  }

  getTierColor(tier: AlertTier): string {
    switch (tier) {
      case 'TIER_3_DANGER':
        return '#ef4444';
      case 'TIER_2_WARNING':
        return '#f59e0b';
      case 'TIER_1_NOTICE':
        return '#38bdf8';
      default:
        return '#10b981';
    }
  }

  private buildBadgeSvgIcon(emoji: string, bgHex: string, strokeHex: string, tag?: string): string {
    const key = `${emoji}_${bgHex}_${strokeHex}_${tag || ''}`;
    if (this.iconCache.has(key)) {
      return this.iconCache.get(key)!;
    }
    const width = tag ? 68 : 40;
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="40" viewBox="0 0 ${width} 40">
      <rect x="2" y="4" width="${width - 4}" height="30" rx="15" fill="${bgHex}" stroke="${strokeHex}" stroke-width="2.5"/>
      <text x="${tag ? 18 : 20}" y="24" font-size="16" text-anchor="middle">${emoji}</text>
      ${
        tag
          ? `<text x="44" y="23" font-family="Inter, sans-serif" font-size="10" font-weight="800" fill="#ffffff" text-anchor="middle">${tag}</text>`
          : ''
      }
    </svg>`;
    const url = `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`;
    this.iconCache.set(key, url);
    return url;
  }

  onGoogleMapClick(event: google.maps.MapMouseEvent) {
    if (!event.latLng) return;
    const lat = event.latLng.lat();
    const lng = event.latLng.lng();
    if (this.state.isGpsSimulateClickMode()) {
      this.state.updateDeviceLocation({ latitude: lat, longitude: lng }, 'Map Click GPS');
    }
  }

  simulateEnterZone(gf: GeofenceZone) {
    this.state.updateDeviceLocation(gf.centroid, 'Zone Teleport');
    this.selectGeofence(gf);
  }

  selectGeofence(gf: GeofenceZone) {
    this.state.selectedEntity.set({
      id: gf.id,
      title: gf.name,
      subtitle: `${gf.causalExplanation} • Radius ${gf.radiusMeters}m`,
      speciesOrType: gf.targetSpecies.join(', '),
      threatTier: gf.tier,
      geminiBadgeText: `Autonomous AI Barrier • Target: ${gf.targetSpecies.join(', ')}`,
      canMarkResolved: false,
      isRoadkill: false,
      isGeofence: true,
      radiusMeters: gf.radiusMeters,
      latitude: gf.centroid.latitude,
      longitude: gf.centroid.longitude
    });
  }

  selectRoadkill(rk: RoadkillRecord) {
    this.state.selectedEntity.set({
      id: rk.id,
      title: `Decomposing ${rk.species} Carcass (#${rk.id})`,
      subtitle: `Active Apex Predator Lure • Status: ${rk.clearanceStatus}`,
      speciesOrType: rk.species,
      threatTier: rk.clearanceStatus === 'CLEARED' ? 'SAFE' : 'TIER_2_WARNING',
      geminiBadgeText: `Gemini Forensics: ${rk.decayState} • Threat Score ${rk.attractantThreatScore}`,
      canMarkResolved: true,
      isRoadkill: true,
      roadkillStatus: rk.clearanceStatus,
      isGeofence: false,
      latitude: rk.location.latitude,
      longitude: rk.location.longitude
    });
  }

  selectAttractant(att: AttractantMarkerItem) {
    this.state.selectedEntity.set({
      id: att.id,
      title: att.title,
      subtitle: att.subtitle,
      speciesOrType: att.layerType,
      threatTier: att.threatTier,
      geminiBadgeText: att.geminiBadge,
      canMarkResolved: true,
      isRoadkill: false,
      isGeofence: false,
      radiusMeters: att.radiusMeters,
      latitude: att.lat,
      longitude: att.lon
    });
  }

  onPolygonClick(event: google.maps.PolyMouseEvent, zone: GeofenceZone) {
    if (this.state.isGpsSimulateClickMode() && event.latLng) {
      const lat = event.latLng.lat();
      const lng = event.latLng.lng();
      this.state.updateDeviceLocation({ latitude: lat, longitude: lng }, 'Geofence Border Click');
    }
    this.selectGeofence(zone);
  }

  onCircleClick(event: google.maps.MapMouseEvent, zone: GeofenceZone) {
    if (this.state.isGpsSimulateClickMode() && event.latLng) {
      const lat = event.latLng.lat();
      const lng = event.latLng.lng();
      this.state.updateDeviceLocation({ latitude: lat, longitude: lng }, 'Geofence Border Click');
    }
    this.selectGeofence(zone);
  }

  onUserGpsDragEnd(event: google.maps.MapMouseEvent) {
    if (event.latLng) {
      const lat = event.latLng.lat();
      const lng = event.latLng.lng();
      this.state.updateDeviceLocation({ latitude: lat, longitude: lng }, 'GPS Drag & Drop');
    }
  }

  recenterToUserGps() {
    const loc = this.state.userLocation();
    this.state.mapCenter.set({ lat: loc.latitude, lng: loc.longitude });
    this.state.showToast(`🎯 Centered view to GPS (${loc.latitude.toFixed(4)}, ${loc.longitude.toFixed(4)})`);
  }

  @HostListener('window:keydown', ['$event'])
  handleKeyboardEvent(event: KeyboardEvent) {
    // Ignore keystrokes if typing inside an interactive form control
    const target = event.target as HTMLElement | null;
    if (
      target &&
      (target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.tagName === 'SELECT' ||
        target.isContentEditable)
    ) {
      return;
    }

    if (!this.state.isGpsSimulateClickMode()) {
      return;
    }

    let direction: 'NORTH' | 'SOUTH' | 'EAST' | 'WEST' | null = null;
    switch (event.code) {
      case 'ArrowUp':
      case 'KeyW':
        direction = 'NORTH';
        break;
      case 'ArrowDown':
      case 'KeyS':
        direction = 'SOUTH';
        break;
      case 'ArrowLeft':
      case 'KeyA':
        direction = 'WEST';
        break;
      case 'ArrowRight':
      case 'KeyD':
        direction = 'EAST';
        break;
    }

    if (direction) {
      event.preventDefault();
      const stepMeters = event.shiftKey
        ? this.state.gpsStepSizeMeters() * 3
        : this.state.gpsStepSizeMeters();
      this.state.stepSimulatedLocation(direction, stepMeters);
    }
  }
}

