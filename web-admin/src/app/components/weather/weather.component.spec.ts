import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WeatherComponent } from './weather.component';
import { AdminApiService } from '../../services/admin-api.service';
import { provideHttpClient } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { of } from 'rxjs';
import { AdminWeatherData } from '../../models/admin.model';

describe('WeatherComponent', () => {
  let component: WeatherComponent;
  let fixture: ComponentFixture<WeatherComponent>;
  let api: AdminApiService;

  const mockWeatherData: AdminWeatherData[] = [
    {
      id: 'weather-thai-01',
      station_name: 'Kaeng Krachan Substation Alpha',
      region: 'Thailand (Kaeng Krachan)',
      latitude: 12.875,
      longitude: 99.630,
      temperature_celsius: 31.5,
      humidity_pct: 82.0,
      wind_speed_kmh: 14.0,
      wind_direction_deg: 90.0,
      precipitation_mm_h: 0.0,
      moon_illumination: 0.12,
      moon_phase: 'WANING_CRESCENT',
      is_flooded_river_risk: false,
      weather_condition: 'CLEAR',
      notes: 'Dry evening east wind',
      last_updated: new Date().toISOString()
    },
    {
      id: 'weather-india-05',
      station_name: 'Kaziranga Panbari Flood Gauge',
      region: 'India (Assam)',
      latitude: 26.590,
      longitude: 93.180,
      temperature_celsius: 27.5,
      humidity_pct: 94.0,
      wind_speed_kmh: 18.0,
      wind_direction_deg: 220.0,
      precipitation_mm_h: 32.0,
      moon_illumination: 0.88,
      moon_phase: 'WAXING_GIBBOUS',
      is_flooded_river_risk: true,
      weather_condition: 'HEAVY_RAIN',
      notes: 'Severe precipitation flood hazard',
      last_updated: new Date().toISOString()
    }
  ];

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [WeatherComponent],
      providers: [
        AdminApiService,
        provideHttpClient(),
        provideHttpClientTesting()
      ]
    }).compileComponents();

    api = TestBed.inject(AdminApiService);
    spyOn(api, 'getWeather').and.returnValue(of(mockWeatherData));
    spyOn(api, 'createWeather').and.callFake((item) => {
      const created: AdminWeatherData = {
        id: 'weather-new-99',
        station_name: item.station_name || '',
        region: item.region || '',
        latitude: item.latitude || 0,
        longitude: item.longitude || 0,
        temperature_celsius: item.temperature_celsius || 25,
        humidity_pct: item.humidity_pct || 60,
        wind_speed_kmh: item.wind_speed_kmh || 10,
        wind_direction_deg: item.wind_direction_deg || 90,
        precipitation_mm_h: item.precipitation_mm_h || 0,
        moon_illumination: item.moon_illumination || 0.5,
        moon_phase: item.moon_phase || 'FIRST_QUARTER',
        is_flooded_river_risk: (item.precipitation_mm_h || 0) >= 25,
        weather_condition: item.weather_condition || 'CLEAR',
        notes: item.notes || '',
        last_updated: new Date().toISOString()
      };
      api.weather.update(l => [created, ...l]);
      return of(created);
    });
    spyOn(api, 'updateWeather').and.callFake((id, item) => {
      let updated: AdminWeatherData;
      api.weather.update(list => list.map(w => {
        if (w.id === id) {
          updated = { ...w, ...item, last_updated: new Date().toISOString() };
          return updated;
        }
        return w;
      }));
      return of(updated!);
    });
    spyOn(api, 'deleteWeather').and.callFake((id) => {
      api.weather.update(l => l.filter(w => w.id !== id));
      return of({ status: 'deleted', id });
    });

    api.weather.set([...mockWeatherData]);

    fixture = TestBed.createComponent(WeatherComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create WeatherComponent', () => {
    expect(component).toBeTruthy();
  });

  it('should list all weather observations from signal', () => {
    expect(component.filteredWeather().length).toBe(2);
    expect(component.filteredWeather()[0].station_name).toBe('Kaeng Krachan Substation Alpha');
  });

  it('should filter weather records by region', () => {
    component.regionFilter = 'Thailand';
    expect(component.filteredWeather().length).toBe(1);
    expect(component.filteredWeather()[0].region).toContain('Thailand');

    component.regionFilter = 'India';
    expect(component.filteredWeather().length).toBe(1);
    expect(component.filteredWeather()[0].region).toContain('India');

    component.regionFilter = '';
    expect(component.filteredWeather().length).toBe(2);
  });

  it('should accurately detect flooded river hazards', () => {
    expect(component.countFloodedStations()).toBe(1);
  });

  it('should detect peak nocturnal risk under low moon illumination', () => {
    expect(component.countPeakNocturnalRisk()).toBe(1); // 0.12 <= 0.20
  });

  it('should apply quick presets correctly', () => {
    const preset = component.regionPresets.find(p => p.name.includes('Nagano'))!;
    expect(preset).toBeDefined();
    component.applyPreset(preset);
    expect(component.formData.region).toBe('Japan (Nagano / Chikuma River)');
    expect(component.formData.wind_speed_kmh).toBe(11.5);
    expect(component.formData.moon_phase).toBe('NEW_MOON');
  });

  it('should trigger live weather sync', () => {
    spyOn(api, 'syncLiveWeather').and.returnValue(of({
      status: 'success',
      updated_count: 20,
      error_count: 0,
      synced_at: new Date().toISOString(),
      source: 'Open-Meteo REST API'
    }));
    component.syncLiveWeather();
    expect(api.syncLiveWeather).toHaveBeenCalled();
    expect(component.syncMessage()).toContain('Successfully synchronized 20 observation stations');
  });

  it('should add a new observation node', () => {
    component.startNew();
    component.formData.station_name = 'New Test Sensor';
    component.formData.region = 'Japan (Nagano)';
    component.formData.precipitation_mm_h = 30;
    component.onPrecipChange();
    expect(component.formData.is_flooded_river_risk).toBeTrue();

    component.saveWeather();
    expect(api.createWeather).toHaveBeenCalled();
    expect(component.activeTab()).toBe('list');
    expect(api.weather().length).toBe(3);
  });

  it('should edit an existing observation node', () => {
    const itemToEdit = component.filteredWeather()[0];
    component.startEdit(itemToEdit);
    expect(component.isEditing()).toBeTrue();
    expect(component.activeTab()).toBe('form');
    expect(component.formData.station_name).toBe(itemToEdit.station_name);

    component.formData.temperature_celsius = 34.0;
    component.saveWeather();
    expect(api.updateWeather).toHaveBeenCalled();
    expect(component.activeTab()).toBe('list');
  });

  it('should delete an observation node after confirmation', () => {
    spyOn(window, 'confirm').and.returnValue(true);
    component.deleteWeather('weather-thai-01');
    expect(api.deleteWeather).toHaveBeenCalledWith('weather-thai-01');
    expect(api.weather().length).toBe(1);
  });
});
