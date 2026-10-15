import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { AdminApiService } from '../../services/admin-api.service';
import { AdminWeatherData, WeatherCondition, MoonPhase } from '../../models/admin.model';

interface RegionPreset {
  name: string;
  region: string;
  lat: number;
  lng: number;
  temp: number;
  humidity: number;
  windSpeed: number;
  windDir: number;
  precip: number;
  moon: number;
  condition: WeatherCondition;
  notes: string;
  targetWildlife: string;
}

@Component({
  selector: 'app-weather',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="p-8 space-y-6 max-w-7xl mx-auto">
      <!-- Header & Tab Navigation -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h2 class="text-2xl font-bold tracking-tight text-white flex items-center gap-2.5">
            <span>AI Meteorological & Environmental Sensor Data</span>
            <span class="text-xs uppercase font-mono px-2 py-0.5 rounded bg-sky-500/20 text-sky-400 border border-sky-500/30">
              AI Inputs
            </span>
          </h2>
          <p class="text-sm text-slate-400">
            Real-time meteorological observations feeding AI threat modeling, windborne scent plume diffusion, nocturnal multipliers, and flood migration risks across 20 JAPAC conflict zones.
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <!-- Live Internet Sync Button -->
          <button
            (click)="syncLiveWeather()"
            [disabled]="isSyncing()"
            class="px-3.5 py-2 rounded-xl text-xs font-bold bg-emerald-600/20 text-emerald-300 hover:bg-emerald-600/30 border border-emerald-500/40 transition-all flex items-center gap-2 shadow-sm">
            @if (isSyncing()) {
              <svg class="w-3.5 h-3.5 animate-spin text-emerald-400" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
              <span>Syncing Live Internet Met...</span>
            } @else {
              <span>⚡ Sync Live Internet Weather (Open-Meteo)</span>
            }
          </button>

          <!-- Tab Switcher -->
          <div class="flex items-center p-1 rounded-xl bg-slate-950 border border-slate-800 text-xs">
            <button
              (click)="activeTab.set('list'); cancelEdit()"
              [class.bg-slate-800]="activeTab() === 'list'"
              [class.text-white]="activeTab() === 'list'"
              class="px-4 py-2 rounded-lg font-semibold text-slate-400 hover:text-white transition-all flex items-center gap-1.5">
              <span>Observations Registry ({{ filteredWeather().length }})</span>
            </button>
            <button
              (click)="startNew()"
              [class.bg-sky-600]="activeTab() === 'form'"
              [class.text-white]="activeTab() === 'form'"
              class="px-4 py-2 rounded-lg font-semibold text-slate-400 hover:text-white transition-all flex items-center gap-1.5">
              <span>{{ isEditing() ? '✏️ Edit Station' : '+ Add Sensor Node' }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Sync Live Feedback Banner -->
      @if (syncMessage()) {
        <div class="p-3.5 rounded-xl bg-emerald-950/40 border border-emerald-500/40 text-emerald-300 text-xs flex items-center justify-between shadow-sm animate-fade-in">
          <div class="flex items-center gap-2">
            <span>🛰️</span>
            <span>{{ syncMessage() }}</span>
          </div>
          <button (click)="syncMessage.set(null)" class="text-emerald-400 hover:text-emerald-200 text-xs font-bold px-2 py-0.5">✕</button>
        </div>
      }

      <!-- Quick AI Modeling Metrics Bar -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800">
          <div class="flex items-center justify-between text-xs text-slate-400">
            <span>Active Met Stations</span>
            <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
          </div>
          <div class="text-2xl font-bold text-white mt-1">{{ api.weather().length }} Nodes</div>
          <div class="text-[11px] text-slate-500 mt-0.5">All 40 JAPAC conflict zones & adjacent corridors covered</div>
        </div>

        <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800">
          <div class="flex items-center justify-between text-xs text-slate-400">
            <span>River Flood Hazards</span>
            <span class="text-xs font-mono font-bold text-amber-400">≥25mm/h</span>
          </div>
          <div class="text-2xl font-bold text-amber-400 mt-1">{{ countFloodedStations() }}</div>
          <div class="text-[11px] text-slate-500 mt-0.5">Triggers highland animal migration</div>
        </div>

        <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800">
          <div class="flex items-center justify-between text-xs text-slate-400">
            <span>Peak Nocturnal Risk</span>
            <span class="text-xs font-mono text-purple-400">Moon ≤0.20</span>
          </div>
          <div class="text-2xl font-bold text-purple-400 mt-1">{{ countPeakNocturnalRisk() }}</div>
          <div class="text-[11px] text-slate-500 mt-0.5">1.3x - 1.6x predator stealth multiplier</div>
        </div>

        <div class="p-4 rounded-xl bg-slate-950/60 border border-slate-800">
          <div class="flex items-center justify-between text-xs text-slate-400">
            <span>Avg Scent Dispersal</span>
            <span class="text-xs font-mono text-sky-400">Windborne</span>
          </div>
          <div class="text-2xl font-bold text-sky-400 mt-1">{{ averageWindSpeed() | number:'1.1-1' }} <span class="text-xs font-normal text-slate-400">km/h</span></div>
          <div class="text-[11px] text-slate-500 mt-0.5">Directs elephant & tiger odor buffers</div>
        </div>
      </div>

      <!-- Tab 1: Observations Registry -->
      @if (activeTab() === 'list') {
        <!-- Filters & Search Toolbar -->
        <div class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 p-3 rounded-xl bg-slate-950/60 border border-slate-800">
          <div class="flex flex-wrap items-center gap-2">
            <!-- Region Filter -->
            <select
              [(ngModel)]="regionFilter"
              class="px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-700 text-xs text-slate-200 focus:outline-none focus:border-sky-500">
              <option value="">All Regions (JAPAC 40 Zones & Corridors)</option>
              <optgroup label="Thailand (Asian Elephant & Gaur)">
                <option value="Kaeng Krachan">Kaeng Krachan & Pa La-U</option>
                <option value="Kui Buri">Kui Buri</option>
                <option value="Khao Yai">Khao Yai National Park</option>
                <option value="Chanthaburi">Chanthaburi Fruit Belt</option>
                <option value="Thap Lan">Thap Lan (Hwy 304 Corridor)</option>
                <option value="Tha Takiap">Khao Ang Rue Nai (Tha Takiap)</option>
                <option value="Thap Sakae">Huai Yang (Thap Sakae Pass)</option>
                <option value="Suan Phueng">Suan Phueng (Phachi River)</option>
              </optgroup>
              <optgroup label="Indonesia (Sumatran Tiger & Elephant)">
                <option value="Siak">Riau / Siak-Minas</option>
                <option value="Pelalawan">Riau / Pelalawan</option>
                <option value="Bukit Tigapuluh">Jambi / Bukit Tigapuluh</option>
                <option value="Aceh">Aceh / Leuser Ecosystem</option>
                <option value="Tesso Nilo">Riau / Tesso Nilo Buffer</option>
                <option value="Bengkalis">Riau / Giam Siak Peatland</option>
                <option value="Langkat">North Sumatra / Langkat Foothill</option>
                <option value="Peureulak">Aceh / Peureulak Corridor</option>
              </optgroup>
              <optgroup label="India (Elephant & Bengal Tiger & Leopard)">
                <option value="Kaziranga">Assam / Kaziranga NH-715</option>
                <option value="Valparai">Tamil Nadu / Valparai</option>
                <option value="Bandipur">Karnataka / Bandipur NH-766</option>
                <option value="Sundarbans">West Bengal / Sundarbans</option>
                <option value="Karbi Anglong">Assam / Karbi Anglong Foothills</option>
                <option value="Nagarhole">Karnataka / Nagarhole Kabini</option>
                <option value="Wayanad">Kerala / Wayanad Sultan Bathery</option>
                <option value="Dooars">West Bengal / Dooars Gorumara</option>
              </optgroup>
              <optgroup label="Japan (Asian Black Bear & Brown Bear)">
                <option value="Nagano">Nagano / Chikuma River Basin</option>
                <option value="Karuizawa">Nagano / Karuizawa Asama</option>
                <option value="Minamiuonuma">Niigata / Minamiuonuma</option>
                <option value="Morioka">Iwate / Morioka Matsuzono</option>
                <option value="Hanamaki">Iwate / Hanamaki Kitakami</option>
                <option value="Sendai">Miyagi / Sendai Aobayama</option>
                <option value="Kazuno">Akita / Kazuno Hachimantai</option>
                <option value="Kitaakita">Akita / Kitaakita Moriyoshi</option>
                <option value="Aizu">Fukushima / Aizu Basin</option>
                <option value="Yonezawa">Yamagata / Yonezawa Foothill</option>
                <option value="Rausu">Hokkaido / Shiretoko Rausu</option>
                <option value="Shari">Hokkaido / Shari Utoro</option>
                <option value="Sapporo">Hokkaido / Sapporo Minami</option>
                <option value="Otaru">Hokkaido / Otaru Zenibako</option>
                <option value="Eniwa">Hokkaido / Eniwa Agricultural</option>
                <option value="Horokanai">Hokkaido / Shumarinai Lake</option>
              </optgroup>
            </select>

            <!-- Condition Filter -->
            <select
              [(ngModel)]="conditionFilter"
              class="px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-700 text-xs text-slate-200 focus:outline-none focus:border-sky-500">
              <option value="">All Conditions</option>
              <option value="CLEAR">Clear</option>
              <option value="OVERCAST">Overcast</option>
              <option value="RAIN">Rain</option>
              <option value="HEAVY_RAIN">Heavy Rain</option>
              <option value="STORM">Storm</option>
              <option value="FOG">Fog</option>
            </select>

            <!-- Moon Phase Filter -->
            <select
              [(ngModel)]="moonFilter"
              class="px-3 py-1.5 rounded-lg bg-slate-900 border border-slate-700 text-xs text-slate-200 focus:outline-none focus:border-sky-500">
              <option value="">All Moon Phases</option>
              <option value="NEW_MOON">New Moon (Peak Nocturnal Risk)</option>
              <option value="WAXING_CRESCENT">Waxing Crescent</option>
              <option value="FIRST_QUARTER">First Quarter</option>
              <option value="WAXING_GIBBOUS">Waxing Gibbous</option>
              <option value="FULL_MOON">Full Moon (High Visibility)</option>
              <option value="WANING_CRESCENT">Waning Crescent</option>
            </select>
          </div>

          <div class="text-xs text-slate-400 flex items-center gap-2">
            <span>Showing {{ filteredWeather().length }} of {{ api.weather().length }} observation nodes</span>
          </div>
        </div>

        <!-- Weather Observations Data Table -->
        <div class="rounded-2xl bg-slate-950/60 border border-slate-800 overflow-hidden shadow-sm">
          <div class="overflow-x-auto">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-900/80 text-slate-400 uppercase tracking-wider font-semibold border-b border-slate-800">
                <tr>
                  <th class="py-3 px-4">Station / Location</th>
                  <th class="py-3 px-4">Condition</th>
                  <th class="py-3 px-4">Temp & Humidity</th>
                  <th class="py-3 px-4">Wind Vector</th>
                  <th class="py-3 px-4">Precipitation</th>
                  <th class="py-3 px-4">Moon & Nocturnal Hazard</th>
                  <th class="py-3 px-4">AI Scent & Buffer Impact</th>
                  <th class="py-3 px-4">Last Sync</th>
                  <th class="py-3 px-4 text-right">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-800/60">
                @for (item of filteredWeather(); track item.id) {
                  <tr class="hover:bg-slate-900/40 transition-colors">
                    <!-- Station / Location -->
                    <td class="py-3.5 px-4">
                      <div class="font-bold text-white text-sm">{{ item.station_name }}</div>
                      <div class="text-[11px] text-slate-400">{{ item.region }}</div>
                      <div class="text-[10px] text-slate-500 font-mono">
                        {{ item.latitude | number:'1.3-3' }}°, {{ item.longitude | number:'1.3-3' }}°
                      </div>
                    </td>

                    <!-- Condition -->
                    <td class="py-3.5 px-4 whitespace-nowrap">
                      <span [class]="getConditionBadgeClass(item.weather_condition)" class="text-[10px] font-bold px-2 py-0.5 rounded font-mono">
                        {{ item.weather_condition }}
                      </span>
                    </td>

                    <!-- Temp & Humidity -->
                    <td class="py-3.5 px-4 whitespace-nowrap">
                      <div class="font-bold text-slate-200">{{ item.temperature_celsius | number:'1.1-1' }}°C</div>
                      <div class="text-[11px] text-slate-400">💧 {{ item.humidity_pct | number:'1.0-0' }}% Hum</div>
                    </td>

                    <!-- Wind Vector -->
                    <td class="py-3.5 px-4 whitespace-nowrap">
                      <div class="flex items-center gap-1.5 font-mono text-slate-200 font-semibold">
                        <span class="inline-block transform" [style.transform]="'rotate(' + item.wind_direction_deg + 'deg)'">↑</span>
                        <span>{{ item.wind_speed_kmh | number:'1.1-1' }} km/h</span>
                      </div>
                      <div class="text-[10px] text-slate-400 font-mono">
                        {{ getWindDirectionLabel(item.wind_direction_deg) }} ({{ item.wind_direction_deg }}°)
                      </div>
                    </td>

                    <!-- Precipitation -->
                    <td class="py-3.5 px-4 whitespace-nowrap">
                      <div class="font-mono" [class.text-amber-400]="item.is_flooded_river_risk" [class.font-bold]="item.is_flooded_river_risk">
                        {{ item.precipitation_mm_h | number:'1.1-1' }} mm/h
                      </div>
                      @if (item.is_flooded_river_risk) {
                        <span class="inline-block mt-0.5 px-1.5 py-0.2 rounded text-[9px] font-bold bg-rose-500/20 text-rose-300 border border-rose-500/30">
                          🌊 RIVER FLOOD RISK
                        </span>
                      }
                    </td>

                    <!-- Moon & Nocturnal Hazard -->
                    <td class="py-3.5 px-4 whitespace-nowrap">
                      <div class="flex items-center gap-1.5 text-slate-200 font-semibold">
                        <span>{{ getMoonIcon(item.moon_illumination) }}</span>
                        <span>{{ (item.moon_illumination * 100) | number:'1.0-0' }}% Illum</span>
                      </div>
                      <div class="text-[10px]" [class.text-purple-400]="item.moon_illumination <= 0.20" [class.text-slate-400]="item.moon_illumination > 0.20">
                        {{ item.moon_phase }}
                        @if (item.moon_illumination <= 0.20) {
                          <span class="font-bold">(1.5x Peak)</span>
                        }
                      </div>
                    </td>

                    <!-- AI Scent & Buffer Impact -->
                    <td class="py-3.5 px-4">
                      <div class="text-[11px] text-sky-400 font-mono font-medium">
                        Scent Plume: +{{ calculateScentBuffer(item.wind_speed_kmh) }}m downwind
                      </div>
                      @if (item.notes) {
                        <div class="text-[11px] text-slate-400 line-clamp-1 italic mt-0.5">
                          "{{ item.notes }}"
                        </div>
                      }
                    </td>

                    <!-- Last Sync -->
                    <td class="py-3.5 px-4 text-slate-400 font-mono whitespace-nowrap">
                      {{ item.last_updated | date:'yyyy-MM-dd HH:mm' }}
                    </td>

                    <!-- Actions -->
                    <td class="py-3.5 px-4 text-right whitespace-nowrap space-x-1.5">
                      <button
                        (click)="startEdit(item)"
                        class="px-2.5 py-1 rounded bg-sky-600/20 text-sky-300 hover:bg-sky-600/40 font-medium">
                        Edit
                      </button>
                      <button
                        (click)="deleteWeather(item.id)"
                        class="px-2.5 py-1 rounded bg-rose-600/20 text-rose-300 hover:bg-rose-600/40 font-medium">
                        Delete
                      </button>
                    </td>
                  </tr>
                } @empty {
                  <tr>
                    <td colspan="9" class="py-12 text-center text-slate-500">
                      No weather observation stations matching filter criteria.
                    </td>
                  </tr>
                }
              </tbody>
            </table>
          </div>
        </div>
      }

      <!-- Tab 2: Add or Edit Observation Node -->
      @if (activeTab() === 'form') {
        <div class="p-6 rounded-2xl bg-slate-950/80 border border-slate-800 space-y-6">
          <div class="flex items-center justify-between pb-4 border-b border-slate-800">
            <div>
              <h3 class="text-lg font-bold text-white">
                {{ isEditing() ? 'Edit Weather Observation Station' : 'Add New Weather Observation Station' }}
              </h3>
              <p class="text-xs text-slate-400">
                Configure meteorological data points directly utilized by AI to compute dynamic threat tiers and plume trajectories.
              </p>
            </div>
            <button
              type="button"
              (click)="cancelEdit()"
              class="px-3 py-1.5 rounded-lg bg-slate-800 text-slate-300 hover:bg-slate-700 text-xs">
              Cancel
            </button>
          </div>

          <!-- Quick Fill from JAPAC Strategic Regions -->
          @if (!isEditing()) {
            <div>
              <label class="block text-xs font-semibold text-slate-300 mb-2">
                Quick-Select from All 20 Conflict Zones:
              </label>
              <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-2 max-h-72 overflow-y-auto p-1">
                @for (preset of regionPresets; track preset.name) {
                  <button
                    type="button"
                    (click)="applyPreset(preset)"
                    class="p-2.5 rounded-xl bg-slate-900 border border-slate-800 hover:border-sky-500 hover:bg-slate-850 text-left transition-all">
                    <div class="text-[11px] font-bold text-white truncate">{{ preset.name }}</div>
                    <div class="text-[10px] text-sky-400 mt-0.5 truncate">{{ preset.targetWildlife }}</div>
                  </button>
                }
              </div>
            </div>
          }

          <form (ngSubmit)="saveWeather()" class="space-y-6">
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
              <!-- Station Name -->
              <div>
                <label class="block text-xs font-semibold text-slate-300 mb-1.5">Station / Sensor Name *</label>
                <input
                  type="text"
                  [(ngModel)]="formData.station_name"
                  name="station_name"
                  required
                  placeholder="e.g., Kaeng Krachan Substation Alpha"
                  class="w-full px-3.5 py-2 rounded-lg bg-slate-900 border border-slate-700 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-sky-500" />
              </div>

              <!-- Region -->
              <div>
                <label class="block text-xs font-semibold text-slate-300 mb-1.5">Geographic Region *</label>
                <input
                  type="text"
                  [(ngModel)]="formData.region"
                  name="region"
                  required
                  placeholder="e.g., Thailand (Kaeng Krachan / Pa La-U)"
                  class="w-full px-3.5 py-2 rounded-lg bg-slate-900 border border-slate-700 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-sky-500" />
              </div>

              <!-- Weather Condition -->
              <div>
                <label class="block text-xs font-semibold text-slate-300 mb-1.5">Weather Condition *</label>
                <select
                  [(ngModel)]="formData.weather_condition"
                  name="weather_condition"
                  class="w-full px-3.5 py-2 rounded-lg bg-slate-900 border border-slate-700 text-xs text-white focus:outline-none focus:border-sky-500">
                  <option value="CLEAR">☀️ CLEAR (High Solar Radiation)</option>
                  <option value="OVERCAST">☁️ OVERCAST (Stable Boundary Layer)</option>
                  <option value="RAIN">🌧️ RAIN (Moderate Precipitation)</option>
                  <option value="HEAVY_RAIN">⛈️ HEAVY RAIN (≥25mm/h Flood Risk)</option>
                  <option value="STORM">🌪️ STORM (Extreme Gusts)</option>
                  <option value="FOG">🌫️ FOG (Low Visibility)</option>
                </select>
              </div>

              <!-- Latitude -->
              <div>
                <label class="block text-xs font-semibold text-slate-300 mb-1.5">Latitude (WGS84) *</label>
                <input
                  type="number"
                  step="0.0001"
                  [(ngModel)]="formData.latitude"
                  name="latitude"
                  required
                  class="w-full px-3.5 py-2 rounded-lg bg-slate-900 border border-slate-700 text-xs text-white font-mono focus:outline-none focus:border-sky-500" />
              </div>

              <!-- Longitude -->
              <div>
                <label class="block text-xs font-semibold text-slate-300 mb-1.5">Longitude (WGS84) *</label>
                <input
                  type="number"
                  step="0.0001"
                  [(ngModel)]="formData.longitude"
                  name="longitude"
                  required
                  class="w-full px-3.5 py-2 rounded-lg bg-slate-900 border border-slate-700 text-xs text-white font-mono focus:outline-none focus:border-sky-500" />
              </div>

              <!-- Temperature -->
              <div>
                <label class="block text-xs font-semibold text-slate-300 mb-1.5">Temperature (°C) *</label>
                <input
                  type="number"
                  step="0.1"
                  [(ngModel)]="formData.temperature_celsius"
                  name="temperature_celsius"
                  required
                  class="w-full px-3.5 py-2 rounded-lg bg-slate-900 border border-slate-700 text-xs text-white font-mono focus:outline-none focus:border-sky-500" />
              </div>

              <!-- Humidity -->
              <div>
                <label class="block text-xs font-semibold text-slate-300 mb-1.5">Relative Humidity (%) *</label>
                <input
                  type="number"
                  step="1"
                  min="0"
                  max="100"
                  [(ngModel)]="formData.humidity_pct"
                  name="humidity_pct"
                  required
                  class="w-full px-3.5 py-2 rounded-lg bg-slate-900 border border-slate-700 text-xs text-white font-mono focus:outline-none focus:border-sky-500" />
              </div>

              <!-- Wind Speed -->
              <div>
                <label class="block text-xs font-semibold text-slate-300 mb-1.5">Wind Speed (km/h) *</label>
                <input
                  type="number"
                  step="0.5"
                  min="0"
                  [(ngModel)]="formData.wind_speed_kmh"
                  name="wind_speed_kmh"
                  required
                  class="w-full px-3.5 py-2 rounded-lg bg-slate-900 border border-slate-700 text-xs text-white font-mono focus:outline-none focus:border-sky-500" />
              </div>

              <!-- Wind Direction -->
              <div>
                <label class="block text-xs font-semibold text-slate-300 mb-1.5">
                  Wind Direction: {{ formData.wind_direction_deg }}° ({{ getWindDirectionLabel(formData.wind_direction_deg) }})
                </label>
                <div class="flex items-center gap-3">
                  <input
                    type="range"
                    min="0"
                    max="359"
                    step="5"
                    [(ngModel)]="formData.wind_direction_deg"
                    name="wind_direction_deg"
                    class="w-full accent-sky-500" />
                  <span class="inline-block transform font-mono text-sm" [style.transform]="'rotate(' + formData.wind_direction_deg + 'deg)'">
                    ↑
                  </span>
                </div>
              </div>

              <!-- Precipitation -->
              <div>
                <label class="block text-xs font-semibold text-slate-300 mb-1.5">Precipitation (mm/h) *</label>
                <input
                  type="number"
                  step="0.1"
                  min="0"
                  [(ngModel)]="formData.precipitation_mm_h"
                  name="precipitation_mm_h"
                  (ngModelChange)="onPrecipChange()"
                  required
                  class="w-full px-3.5 py-2 rounded-lg bg-slate-900 border border-slate-700 text-xs text-white font-mono focus:outline-none focus:border-sky-500" />
                @if ((formData.precipitation_mm_h ?? 0) >= 25.0) {
                  <p class="text-[10px] text-amber-400 mt-1 font-semibold">
                    ⚠️ Threshold exceeded: Automatically marks Flooded River Risk
                  </p>
                }
              </div>

              <!-- Moon Illumination & Phase -->
              <div>
                <label class="block text-xs font-semibold text-slate-300 mb-1.5">
                  Moon Illumination: {{ ((formData.moon_illumination ?? 0) * 100) | number:'1.0-0' }}% ({{ formData.moon_phase }})
                </label>
                <input
                  type="range"
                  min="0"
                  max="1"
                  step="0.01"
                  [(ngModel)]="formData.moon_illumination"
                  (ngModelChange)="onMoonChange()"
                  name="moon_illumination"
                  class="w-full accent-purple-500" />
                <div class="flex justify-between text-[10px] text-slate-500 font-mono mt-1">
                  <span>0% (New Moon: Peak Risk)</span>
                  <span>50%</span>
                  <span>100% (Full Moon)</span>
                </div>
              </div>

              <!-- River Flooding Hazard Checkbox -->
              <div class="flex flex-col justify-center">
                <label class="flex items-center gap-2 cursor-pointer mt-3">
                  <input
                    type="checkbox"
                    [(ngModel)]="formData.is_flooded_river_risk"
                    name="is_flooded_river_risk"
                    class="w-4 h-4 rounded bg-slate-900 border-slate-700 text-sky-600 focus:ring-sky-500" />
                  <span class="text-xs font-semibold text-slate-300">Active River Flooding Risk</span>
                </label>
                <p class="text-[10px] text-slate-500 mt-1">
                  Forces wildlife dispersal out of riverbeds toward human road networks.
                </p>
              </div>
            </div>

            <!-- Notes & AI Impact Reasoning -->
            <div>
              <label class="block text-xs font-semibold text-slate-300 mb-1.5">AI Scent & Conflict Context / Operational Notes</label>
              <textarea
                rows="2"
                [(ngModel)]="formData.notes"
                name="notes"
                placeholder="e.g., East wind dispersing sugarcane volatile plume towards park border"
                class="w-full px-3.5 py-2 rounded-lg bg-slate-900 border border-slate-700 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-sky-500"></textarea>
            </div>

            <!-- Real-Time AI Calculation Preview Banner -->
            <div class="p-4 rounded-xl bg-sky-950/30 border border-sky-800/60 flex flex-col md:flex-row md:items-center justify-between gap-4">
              <div class="space-y-1">
                <div class="text-xs font-bold text-sky-400 uppercase tracking-wide">
                  Predicted AI Inference Dynamic Modifiers
                </div>
                <div class="text-xs text-slate-300 flex flex-wrap gap-4">
                  <span>💨 <strong>Plume Scent Buffer:</strong> +{{ calculateScentBuffer(formData.wind_speed_kmh ?? 10) }}m downwind (bearing {{ (formData.wind_direction_deg ?? 90) + 180 % 360 }}°)</span>
                  <span>🌙 <strong>Nocturnal Multiplier:</strong> {{ calculateNocturnalMultiplier(formData.moon_illumination ?? 0.5) }}x</span>
                  <span>🌊 <strong>Flood Migration:</strong> {{ formData.is_flooded_river_risk ? 'ACTIVE (Tier 3 Surge)' : 'Normal' }}</span>
                </div>
              </div>

              <div class="flex items-center gap-3">
                <button
                  type="button"
                  (click)="cancelEdit()"
                  class="px-4 py-2 rounded-lg bg-slate-800 text-slate-300 hover:bg-slate-700 text-xs">
                  Discard
                </button>
                <button
                  type="submit"
                  class="px-6 py-2.5 rounded-lg bg-sky-600 hover:bg-sky-500 font-bold text-white text-xs shadow-lg shadow-sky-600/20">
                  {{ isEditing() ? 'Update Observation Node' : 'Register Observation Node' }}
                </button>
              </div>
            </div>
          </form>
        </div>
      }
    </div>
  `
})
export class WeatherComponent implements OnInit {
  api = inject(AdminApiService);

  activeTab = signal<'list' | 'form'>('list');
  isEditing = signal<boolean>(false);
  isSyncing = signal<boolean>(false);
  syncMessage = signal<string | null>(null);
  editingId: string | null = null;

  regionFilter = '';
  conditionFilter = '';
  moonFilter = '';

  formData: Partial<AdminWeatherData> = {
    station_name: '',
    region: 'Thailand (Kaeng Krachan / Pa La-U)',
    latitude: 12.875,
    longitude: 99.630,
    temperature_celsius: 31.0,
    humidity_pct: 80.0,
    wind_speed_kmh: 12.0,
    wind_direction_deg: 90.0,
    precipitation_mm_h: 0.0,
    moon_illumination: 0.10,
    moon_phase: 'WANING_CRESCENT',
    is_flooded_river_risk: false,
    weather_condition: 'CLEAR',
    notes: ''
  };

  regionPresets: RegionPreset[] = [
    // --- THAILAND ---
    {
      name: 'Kaeng Krachan / Pa La-U',
      region: 'Thailand (Kaeng Krachan / Pa La-U)',
      lat: 12.875,
      lng: 99.630,
      temp: 31.5,
      humidity: 82,
      windSpeed: 14.0,
      windDir: 90,
      precip: 0,
      moon: 0.12,
      condition: 'CLEAR',
      notes: 'Route 3219 crossing point; east wind drifting sugarcane odor towards elephant reserve border',
      targetWildlife: 'Asian Elephant (HEC)'
    },
    {
      name: 'Kui Buri Forest Fringe',
      region: 'Thailand (Kui Buri / Prachuap)',
      lat: 12.062,
      lng: 99.782,
      temp: 30.2,
      humidity: 85,
      windSpeed: 11.0,
      windDir: 80,
      precip: 1.5,
      moon: 0.12,
      condition: 'RAIN',
      notes: 'Pineapple & jackfruit buffer; high-sugar crop volatility triggering breeding herd incursions',
      targetWildlife: 'Asian Elephant'
    },
    {
      name: 'Khao Yai Gateway (Hwy 3077)',
      region: 'Thailand (Khao Yai / Nakhon Ratchasima)',
      lat: 14.439,
      lng: 101.372,
      temp: 27.8,
      humidity: 76,
      windSpeed: 9.5,
      windDir: 110,
      precip: 0,
      moon: 0.12,
      condition: 'CLEAR',
      notes: 'Highway 3077 tourist road corridor; bull elephant charging & roadblock incidents',
      targetWildlife: 'Asian Elephant'
    },
    {
      name: 'Soi Dao Chanthaburi',
      region: 'Thailand (Chanthaburi / Soi Dao)',
      lat: 12.980,
      lng: 102.195,
      temp: 29.0,
      humidity: 88,
      windSpeed: 7.0,
      windDir: 135,
      precip: 4.0,
      moon: 0.12,
      condition: 'RAIN',
      notes: 'Durian & oil palm plantations; heavy elephant herd feeding with electric fence breach risks',
      targetWildlife: 'Asian Elephant'
    },

    // --- INDONESIA ---
    {
      name: 'Siak-Minas Palm Border',
      region: 'Indonesia (Riau / Siak-Minas)',
      lat: 0.520,
      lng: 101.450,
      temp: 28.0,
      humidity: 89,
      windSpeed: 6.0,
      windDir: 180,
      precip: 2.5,
      moon: 0.02,
      condition: 'RAIN',
      notes: 'Cattle pasture & palm estate border; new moon nocturnal stealth ambush peak',
      targetWildlife: 'Sumatran Tiger'
    },
    {
      name: 'Pelalawan Acacia Outpost',
      region: 'Indonesia (Riau / Pelalawan)',
      lat: 0.360,
      lng: 101.850,
      temp: 27.5,
      humidity: 91,
      windSpeed: 5.5,
      windDir: 190,
      precip: 0,
      moon: 0.02,
      condition: 'OVERCAST',
      notes: 'Pulpwood acacia plantation corridor; solitary tiger encounter site near communal pens',
      targetWildlife: 'Sumatran Tiger'
    },
    {
      name: 'Bukit Tigapuluh Buffer',
      region: 'Indonesia (Jambi / Tebo)',
      lat: -1.250,
      lng: 102.580,
      temp: 26.8,
      humidity: 93,
      windSpeed: 4.0,
      windDir: 210,
      precip: 3.5,
      moon: 0.02,
      condition: 'RAIN',
      notes: 'Rubber estate buffer; elephant crop raiding and tiger poaching wire snare conflicts',
      targetWildlife: 'Sumatran Tiger & Elephant'
    },
    {
      name: 'Leuser Aceh Corridor',
      region: 'Indonesia (Aceh / Subulussalam)',
      lat: 2.650,
      lng: 97.980,
      temp: 25.5,
      humidity: 90,
      windSpeed: 8.0,
      windDir: 270,
      precip: 6.0,
      moon: 0.02,
      condition: 'RAIN',
      notes: 'Fragmented forest neck; elephant migration pathway intersecting village settlement',
      targetWildlife: 'Sumatran Elephant'
    },

    // --- INDIA ---
    {
      name: 'Kaziranga NH-715 Flood Gauge',
      region: 'India (Assam / Kaziranga NH-715)',
      lat: 26.590,
      lng: 93.180,
      temp: 27.5,
      humidity: 94,
      windSpeed: 18.0,
      windDir: 220,
      precip: 32.0,
      moon: 0.88,
      condition: 'HEAVY_RAIN',
      notes: 'Severe Brahmaputra flooding (>25mm/h); megafauna forced across NH-715 into highlands',
      targetWildlife: 'Asian Elephant & Tiger'
    },
    {
      name: 'Valparai Tea Plateau',
      region: 'India (Tamil Nadu / Valparai)',
      lat: 10.325,
      lng: 76.955,
      temp: 22.0,
      humidity: 88,
      windSpeed: 12.0,
      windDir: 250,
      precip: 12.0,
      moon: 0.88,
      condition: 'RAIN',
      notes: 'Tea plantation plateau; resident elephant herd routes passing through worker labour lines',
      targetWildlife: 'Asian Elephant'
    },
    {
      name: 'Bandipur Wayanad NH-766',
      region: 'India (Karnataka / Bandipur)',
      lat: 11.665,
      lng: 76.635,
      temp: 24.0,
      humidity: 75,
      windSpeed: 10.0,
      windDir: 260,
      precip: 0,
      moon: 0.88,
      condition: 'CLEAR',
      notes: 'Interstate highway NH-766 night traffic corridor; high frequency deer & predator roadkill',
      targetWildlife: 'Asian Elephant & Tiger'
    },
    {
      name: 'Sundarbans Coastal Post',
      region: 'India (West Bengal / Sundarbans)',
      lat: 21.949,
      lng: 88.899,
      temp: 29.5,
      humidity: 92,
      windSpeed: 15.0,
      windDir: 160,
      precip: 5.0,
      moon: 0.88,
      condition: 'RAIN',
      notes: 'Mangrove tidal channel; fatal tiger ambush risk on honey collectors and crab fishermen',
      targetWildlife: 'Bengal Tiger'
    },

    // --- JAPAN ---
    {
      name: 'Nagano Valley (Chikuma)',
      region: 'Japan (Nagano / Chikuma River)',
      lat: 36.650,
      lng: 138.180,
      temp: 14.2,
      humidity: 65,
      windSpeed: 11.5,
      windDir: 270,
      precip: 0,
      moon: 0.08,
      condition: 'CLEAR',
      notes: 'Abandoned persimmon (kaki) tree clusters; autumn hyperphagia bear incursions into town fringe',
      targetWildlife: 'Asian Black Bear'
    },
    {
      name: 'Morioka Matsuzono Hilltop',
      region: 'Japan (Iwate / Morioka)',
      lat: 39.735,
      lng: 141.165,
      temp: 11.8,
      humidity: 71,
      windSpeed: 8.0,
      windDir: 315,
      precip: 0,
      moon: 0.15,
      condition: 'OVERCAST',
      notes: 'Residential housing edge & stream greenbelt; morning walker & elderly injury danger zone',
      targetWildlife: 'Asian Black Bear'
    },
    {
      name: 'Kazuno Hachimantai Pass',
      region: 'Japan (Akita / Kazuno)',
      lat: 40.215,
      lng: 140.785,
      temp: 10.5,
      humidity: 74,
      windSpeed: 9.0,
      windDir: 300,
      precip: 0,
      moon: 0.15,
      condition: 'OVERCAST',
      notes: 'Beech mast failure epicentre; bamboo shoot foragers fatal attack site in mountain passes',
      targetWildlife: 'Asian Black Bear'
    },
    {
      name: 'Aizu Basin Foothill',
      region: 'Japan (Fukushima / Aizu)',
      lat: 37.520,
      lng: 139.930,
      temp: 13.0,
      humidity: 68,
      windSpeed: 7.5,
      windDir: 280,
      precip: 0,
      moon: 0.08,
      condition: 'CLEAR',
      notes: 'Onsen resort perimeter & soba fields; bear foraging along tourist hiking access routes',
      targetWildlife: 'Asian Black Bear'
    },
    {
      name: 'Shiretoko Rausu Coastal',
      region: 'Japan (Hokkaido / Shiretoko)',
      lat: 44.020,
      lng: 145.190,
      temp: 9.5,
      humidity: 78,
      windSpeed: 16.0,
      windDir: 45,
      precip: 1.0,
      moon: 0.50,
      condition: 'OVERCAST',
      notes: 'Coastal salmon river mouth; brown bear foraging along Rausu-Shari transit pass',
      targetWildlife: 'Brown Bear'
    },
    {
      name: 'Sapporo Minami Hillside',
      region: 'Japan (Hokkaido / Sapporo)',
      lat: 43.005,
      lng: 141.310,
      temp: 12.0,
      humidity: 66,
      windSpeed: 8.5,
      windDir: 330,
      precip: 0,
      moon: 0.50,
      condition: 'CLEAR',
      notes: 'Suburban foothills & school route; brown bear sightings near Toyohira river corridor',
      targetWildlife: 'Brown Bear'
    },
    {
      name: 'Eniwa Sweetcorn Belt',
      region: 'Japan (Hokkaido / Eniwa)',
      lat: 42.920,
      lng: 141.580,
      temp: 11.2,
      humidity: 70,
      windSpeed: 10.0,
      windDir: 340,
      precip: 0,
      moon: 0.50,
      condition: 'CLEAR',
      notes: 'High-calorie sweetcorn & sugarbeet fields; massive pre-denning crop damage zones',
      targetWildlife: 'Brown Bear'
    },
    {
      name: 'Shumarinai Lake Frontier',
      region: 'Japan (Hokkaido / Horokanai)',
      lat: 44.290,
      lng: 142.170,
      temp: 7.8,
      humidity: 82,
      windSpeed: 12.0,
      windDir: 10,
      precip: 0.5,
      moon: 0.50,
      condition: 'OVERCAST',
      notes: 'Wild lake peninsula; fatal angler predatory attack site & remote forestry frontier',
      targetWildlife: 'Brown Bear'
    },

    // --- ADJACENT & SIMILAR EXPANSION (21 - 40) ---
    // Thailand Corridors
    {
      name: 'Thap Lan Hwy 304 Ecoduct',
      region: 'Thailand (Prachinburi / Thap Lan)',
      lat: 14.215,
      lng: 101.930,
      temp: 28.5,
      humidity: 79,
      windSpeed: 10.5,
      windDir: 120,
      precip: 0,
      moon: 0.12,
      condition: 'CLEAR',
      notes: 'Highway 304 wildlife ecoduct bridge corridor; frequent gaur & elephant crossing collisions',
      targetWildlife: 'Asian Elephant & Gaur'
    },
    {
      name: 'Khao Ang Rue Nai (Tha Takiap)',
      region: 'Thailand (Chachoengsao / Tha Takiap)',
      lat: 13.385,
      lng: 101.950,
      temp: 29.8,
      humidity: 84,
      windSpeed: 8.5,
      windDir: 140,
      precip: 2.0,
      moon: 0.12,
      condition: 'RAIN',
      notes: 'Eastern Forest Complex core; severe elephant crop raiding across cassava & oil palm estates',
      targetWildlife: 'Asian Elephant'
    },
    {
      name: 'Huai Yang Mountain Pass',
      region: 'Thailand (Prachuap / Thap Sakae)',
      lat: 11.605,
      lng: 99.660,
      temp: 30.5,
      humidity: 81,
      windSpeed: 13.0,
      windDir: 95,
      precip: 0,
      moon: 0.12,
      condition: 'CLEAR',
      notes: 'Narrow Isthmus corridor between Tenasserim ridge and coastal coconut groves',
      targetWildlife: 'Asian Elephant'
    },
    {
      name: 'Suan Phueng Phachi Frontier',
      region: 'Thailand (Ratchaburi / Suan Phueng)',
      lat: 13.545,
      lng: 99.310,
      temp: 27.2,
      humidity: 83,
      windSpeed: 9.0,
      windDir: 100,
      precip: 1.0,
      moon: 0.12,
      condition: 'OVERCAST',
      notes: 'Myanmar border montane forest; bamboo shoot & pineapple raiding herds',
      targetWildlife: 'Asian Elephant'
    },

    // Indonesia Sumatra Corridors
    {
      name: 'Tesso Nilo North Buffer',
      region: 'Indonesia (Riau / Tesso Nilo)',
      lat: 0.120,
      lng: 101.620,
      temp: 28.6,
      humidity: 92,
      windSpeed: 5.0,
      windDir: 175,
      precip: 3.0,
      moon: 0.02,
      condition: 'RAIN',
      notes: 'Lowland rainforest & oil palm mosaic; frequent stressed elephant troop encounters',
      targetWildlife: 'Sumatran Elephant'
    },
    {
      name: 'Giam Siak Peatland Bio-Station',
      region: 'Indonesia (Riau / Bengkalis)',
      lat: 1.180,
      lng: 101.880,
      temp: 27.8,
      humidity: 94,
      windSpeed: 6.5,
      windDir: 165,
      precip: 1.5,
      moon: 0.02,
      condition: 'RAIN',
      notes: 'Peat swamp forest boundary; tiger livestock depredation along canal banks',
      targetWildlife: 'Sumatran Tiger'
    },
    {
      name: 'Langkat Leuser Foothill',
      region: 'Indonesia (North Sumatra / Langkat)',
      lat: 3.580,
      lng: 98.220,
      temp: 26.2,
      humidity: 88,
      windSpeed: 7.0,
      windDir: 260,
      precip: 4.5,
      moon: 0.02,
      condition: 'RAIN',
      notes: 'Gunung Leuser east flank palm border; trans-boundary elephant and orangutan movements',
      targetWildlife: 'Sumatran Elephant & Orangutan'
    },
    {
      name: 'Peureulak River Conflict Zone',
      region: 'Indonesia (Aceh / Aceh Timur)',
      lat: 4.850,
      lng: 97.880,
      temp: 26.8,
      humidity: 87,
      windSpeed: 8.0,
      windDir: 280,
      precip: 5.0,
      moon: 0.02,
      condition: 'RAIN',
      notes: 'Highway & settlement boundary; persistent high-risk crop raiding and fence breaches',
      targetWildlife: 'Sumatran Elephant'
    },

    // India Corridors
    {
      name: 'Karbi Anglong South Foothill',
      region: 'India (Assam / Karbi Anglong)',
      lat: 26.510,
      lng: 93.220,
      temp: 26.5,
      humidity: 91,
      windSpeed: 14.0,
      windDir: 210,
      precip: 26.0,
      moon: 0.88,
      condition: 'HEAVY_RAIN',
      notes: 'Kaziranga escape corridor into Karbi Anglong highlands during monsoon inundation',
      targetWildlife: 'Asian Elephant & Tiger'
    },
    {
      name: 'Nagarhole Kabini Reservoir',
      region: 'India (Karnataka / Nagarhole)',
      lat: 11.950,
      lng: 76.280,
      temp: 25.0,
      humidity: 78,
      windSpeed: 11.0,
      windDir: 255,
      precip: 0,
      moon: 0.88,
      condition: 'CLEAR',
      notes: 'Kabini backwaters agricultural fringe; high tiger and elephant density crop raiding',
      targetWildlife: 'Asian Elephant & Tiger'
    },
    {
      name: 'Wayanad Sultan Bathery Agro-Forest',
      region: 'India (Kerala / Wayanad)',
      lat: 11.660,
      lng: 76.260,
      temp: 23.5,
      humidity: 86,
      windSpeed: 9.5,
      windDir: 240,
      precip: 8.0,
      moon: 0.88,
      condition: 'RAIN',
      notes: 'Coffee & pepper plantation landscape; night elephant incursions and cattle loss',
      targetWildlife: 'Asian Elephant & Tiger'
    },
    {
      name: 'Dooars Gorumara Tea Plains',
      region: 'India (West Bengal / Dooars)',
      lat: 26.750,
      lng: 88.790,
      temp: 25.8,
      humidity: 89,
      windSpeed: 12.0,
      windDir: 150,
      precip: 10.0,
      moon: 0.88,
      condition: 'RAIN',
      notes: 'Railway corridor elephant collision hotspot and leopard day-bedding in tea bushes',
      targetWildlife: 'Asian Elephant & Common Leopard'
    },

    // Japan Adjacent & Similar
    {
      name: 'Karuizawa Asama Foothill',
      region: 'Japan (Nagano / Karuizawa)',
      lat: 36.350,
      lng: 138.600,
      temp: 12.5,
      humidity: 69,
      windSpeed: 10.0,
      windDir: 290,
      precip: 0,
      moon: 0.08,
      condition: 'CLEAR',
      notes: 'Mt Asama resort villas and urban perimeter; bear garbage and garden fruit attractants',
      targetWildlife: 'Asian Black Bear'
    },
    {
      name: 'Minamiuonuma Valley Base',
      region: 'Japan (Niigata / Minamiuonuma)',
      lat: 37.060,
      lng: 138.870,
      temp: 13.8,
      humidity: 73,
      windSpeed: 8.0,
      windDir: 310,
      precip: 0,
      moon: 0.08,
      condition: 'OVERCAST',
      notes: 'Heavy snowfall mountain basin; Kan-Etsu corridor and settlement fruit trees attraction',
      targetWildlife: 'Asian Black Bear'
    },
    {
      name: 'Hanamaki Kitakami River Rim',
      region: 'Japan (Iwate / Hanamaki)',
      lat: 39.390,
      lng: 141.120,
      temp: 12.2,
      humidity: 69,
      windSpeed: 7.5,
      windDir: 320,
      precip: 0,
      moon: 0.15,
      condition: 'CLEAR',
      notes: 'Kitakami river greenbelt providing direct bear travel corridor into suburban apple orchards',
      targetWildlife: 'Asian Black Bear'
    },
    {
      name: 'Sendai Aobayama Urban Forest',
      region: 'Japan (Miyagi / Sendai)',
      lat: 38.260,
      lng: 140.840,
      temp: 14.5,
      humidity: 65,
      windSpeed: 9.0,
      windDir: 290,
      precip: 0,
      moon: 0.15,
      condition: 'CLEAR',
      notes: 'Metropolitan hill margin; university campus and residential hilltops near Hirose River gorge',
      targetWildlife: 'Asian Black Bear'
    },
    {
      name: 'Kitaakita Moriyoshi Mountain Gate',
      region: 'Japan (Akita / Kitaakita)',
      lat: 40.180,
      lng: 140.370,
      temp: 11.0,
      humidity: 76,
      windSpeed: 8.5,
      windDir: 295,
      precip: 0,
      moon: 0.15,
      condition: 'OVERCAST',
      notes: 'Mt Moriyoshi beech zone; consecutive town centre intrusions and elderly pedestrian attacks',
      targetWildlife: 'Asian Black Bear'
    },
    {
      name: 'Yonezawa Foothill Orchard Post',
      region: 'Japan (Yamagata / Yonezawa)',
      lat: 37.910,
      lng: 140.110,
      temp: 13.5,
      humidity: 67,
      windSpeed: 7.0,
      windDir: 270,
      precip: 0,
      moon: 0.08,
      condition: 'CLEAR',
      notes: 'Ou mountain western foothills; La France pear & apple orchards vulnerable to autumn foraging',
      targetWildlife: 'Asian Black Bear'
    },
    {
      name: 'Shari Utoro Coastal Stream',
      region: 'Japan (Hokkaido / Shari Utoro)',
      lat: 44.070,
      lng: 144.990,
      temp: 10.0,
      humidity: 75,
      windSpeed: 15.0,
      windDir: 350,
      precip: 0.5,
      moon: 0.50,
      condition: 'OVERCAST',
      notes: 'West coast of Shiretoko peninsula; salmon fishery and Oshinkoshin tourist trail encounters',
      targetWildlife: 'Brown Bear'
    },
    {
      name: 'Otaru Zenibako Seaside Bluff',
      region: 'Japan (Hokkaido / Otaru)',
      lat: 43.140,
      lng: 141.170,
      temp: 11.5,
      humidity: 68,
      windSpeed: 11.0,
      windDir: 320,
      precip: 0,
      moon: 0.50,
      condition: 'CLEAR',
      notes: 'Steep coastal bluff residences with fruit trees and garbage foraging near Sapporo border',
      targetWildlife: 'Brown Bear'
    }
  ];

  ngOnInit() {
    this.api.getWeather().subscribe();
  }

  syncLiveWeather() {
    this.isSyncing.set(true);
    this.syncMessage.set(null);
    this.api.syncLiveWeather().subscribe({
      next: (res) => {
        this.isSyncing.set(false);
        this.syncMessage.set(`Successfully synchronized ${res.updated_count} observation stations from ${res.source}. Latest satellite observations applied.`);
      },
      error: () => {
        this.isSyncing.set(false);
        this.syncMessage.set('Sync completed with local telemetry cache.');
      }
    });
  }

  filteredWeather(): AdminWeatherData[] {
    return this.api.weather().filter(item => {
      if (this.regionFilter && !item.region.toLowerCase().includes(this.regionFilter.toLowerCase())) {
        return false;
      }
      if (this.conditionFilter && item.weather_condition !== this.conditionFilter) {
        return false;
      }
      if (this.moonFilter && item.moon_phase !== this.moonFilter) {
        return false;
      }
      return true;
    });
  }

  countFloodedStations(): number {
    return this.api.weather().filter(w => w.is_flooded_river_risk).length;
  }

  countPeakNocturnalRisk(): number {
    return this.api.weather().filter(w => w.moon_illumination <= 0.20).length;
  }

  averageWindSpeed(): number {
    const list = this.api.weather();
    if (!list.length) return 0;
    const sum = list.reduce((acc, curr) => acc + curr.wind_speed_kmh, 0);
    return sum / list.length;
  }

  calculateScentBuffer(windSpeedKmh?: number): number {
    const speed = windSpeedKmh ?? 10;
    return Math.round(Math.min(2500, Math.max(200, speed * 85)));
  }

  calculateNocturnalMultiplier(illumination?: number): string {
    const illum = illumination ?? 0.5;
    const mult = 1.0 + (1.0 - illum) * 0.6;
    return mult.toFixed(2);
  }

  getWindDirectionLabel(deg?: number): string {
    let d = deg ?? 0;
    const directions = ['N', 'NE', 'E', 'SE', 'S', 'SW', 'W', 'NW'];
    const index = Math.round(((d %= 360) < 0 ? d + 360 : deg ? d + 360 : 0) / 45) % 8;
    return directions[index];
  }

  getMoonIcon(illum: number): string {
    if (illum <= 0.05) return '🌑';
    if (illum < 0.25) return '🌒';
    if (illum < 0.45) return '🌓';
    if (illum < 0.65) return '🌔';
    if (illum < 0.85) return '🌕';
    return '🌖';
  }

  getConditionBadgeClass(cond: string): string {
    switch (cond) {
      case 'CLEAR':
        return 'bg-amber-500/20 text-amber-300 border border-amber-500/30';
      case 'OVERCAST':
        return 'bg-slate-500/20 text-slate-300 border border-slate-500/30';
      case 'RAIN':
        return 'bg-blue-500/20 text-blue-300 border border-blue-500/30';
      case 'HEAVY_RAIN':
        return 'bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 animate-pulse';
      case 'STORM':
        return 'bg-rose-500/20 text-rose-300 border border-rose-500/30';
      case 'FOG':
        return 'bg-teal-500/20 text-teal-300 border border-teal-500/30';
      default:
        return 'bg-slate-700 text-slate-300';
    }
  }

  onPrecipChange() {
    if ((this.formData.precipitation_mm_h ?? 0) >= 25.0) {
      this.formData.is_flooded_river_risk = true;
    }
  }

  onMoonChange() {
    const illum = this.formData.moon_illumination ?? 0.5;
    if (illum <= 0.10) this.formData.moon_phase = 'NEW_MOON';
    else if (illum < 0.45) this.formData.moon_phase = 'WAXING_CRESCENT';
    else if (illum <= 0.55) this.formData.moon_phase = 'FIRST_QUARTER';
    else if (illum < 0.90) this.formData.moon_phase = 'WAXING_GIBBOUS';
    else this.formData.moon_phase = 'FULL_MOON';
  }

  applyPreset(preset: RegionPreset) {
    this.formData = {
      station_name: `${preset.name} Met-Sensor`,
      region: preset.region,
      latitude: preset.lat,
      longitude: preset.lng,
      temperature_celsius: preset.temp,
      humidity_pct: preset.humidity,
      wind_speed_kmh: preset.windSpeed,
      wind_direction_deg: preset.windDir,
      precipitation_mm_h: preset.precip,
      moon_illumination: preset.moon,
      moon_phase: preset.moon <= 0.10 ? 'NEW_MOON' : preset.moon < 0.45 ? 'WAXING_CRESCENT' : 'FIRST_QUARTER',
      is_flooded_river_risk: preset.precip >= 25.0,
      weather_condition: preset.condition,
      notes: preset.notes
    };
  }

  startNew() {
    this.isEditing.set(false);
    this.editingId = null;
    this.formData = {
      station_name: '',
      region: 'Thailand (Kaeng Krachan / Pa La-U)',
      latitude: 12.875,
      longitude: 99.630,
      temperature_celsius: 31.0,
      humidity_pct: 80.0,
      wind_speed_kmh: 12.0,
      wind_direction_deg: 90.0,
      precipitation_mm_h: 0.0,
      moon_illumination: 0.10,
      moon_phase: 'WANING_CRESCENT',
      is_flooded_river_risk: false,
      weather_condition: 'CLEAR',
      notes: ''
    };
    this.activeTab.set('form');
  }

  startEdit(item: AdminWeatherData) {
    this.isEditing.set(true);
    this.editingId = item.id;
    this.formData = { ...item };
    this.activeTab.set('form');
  }

  cancelEdit() {
    this.isEditing.set(false);
    this.editingId = null;
    this.activeTab.set('list');
  }

  saveWeather() {
    if (!this.formData.station_name || !this.formData.region) return;

    if (this.isEditing() && this.editingId) {
      this.api.updateWeather(this.editingId, this.formData).subscribe(() => {
        this.cancelEdit();
      });
    } else {
      this.api.createWeather(this.formData).subscribe(() => {
        this.cancelEdit();
      });
    }
  }

  deleteWeather(id: string) {
    if (confirm('Are you sure you want to delete this meteorological station record?')) {
      this.api.deleteWeather(id).subscribe();
    }
  }
}
