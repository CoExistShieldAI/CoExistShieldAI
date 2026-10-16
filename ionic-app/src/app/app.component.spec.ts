import { TestBed } from '@angular/core/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideIonicAngular } from '@ionic/angular';
import { AppComponent } from './app.component';
import { ReportPanelComponent } from './components/report-panel/report-panel.component';
import { haversineMeters, isPointInPolygon } from './plugins/geofence-bridge.plugin';

describe('AppComponent (CoExist Shield Ionic + Capacitor)', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AppComponent, ReportPanelComponent],
      providers: [provideHttpClient(), provideIonicAngular()]
    }).compileComponents();
  });

  it('should create the Ionic app with Split-Pane HUD & GIS Map', () => {
    const fixture = TestBed.createComponent(AppComponent);
    const app = fixture.componentInstance;
    expect(app).toBeTruthy();
    expect(app.state.geofences().length).toBeGreaterThanOrEqual(7);
  });

  it('should compute Haversine distance and point-in-polygon accurately for GeofenceBridge', () => {
    const dist = haversineMeters(
      { latitude: 39.702, longitude: 141.152 },
      { latitude: 39.702, longitude: 141.152 }
    );
    expect(dist).toBe(0);

    const poly = [
      { latitude: 39.69, longitude: 141.13 },
      { latitude: 39.74, longitude: 141.13 },
      { latitude: 39.74, longitude: 141.19 },
      { latitude: 39.69, longitude: 141.19 }
    ];
    expect(isPointInPolygon({ latitude: 39.71, longitude: 141.15 }, poly)).toBeTrue();
    expect(isPointInPolygon({ latitude: 35.0, longitude: 135.0 }, poly)).toBeFalse();
  });

  it('should navigate through the 3-step wizard in ReportPanelComponent', () => {
    const fixture = TestBed.createComponent(ReportPanelComponent);
    const reportPanel = fixture.componentInstance;
    fixture.detectChanges();

    expect(reportPanel.currentStep()).toBe(1);
    reportPanel.selectCategory('ROADKILL');
    expect(reportPanel.state.reportCategory()).toBe('ROADKILL');
    expect(reportPanel.activeCategoryMeta().isAlert).toBeTrue();

    reportPanel.nextStep();
    expect(reportPanel.currentStep()).toBe(2);

    reportPanel.nextStep();
    expect(reportPanel.currentStep()).toBe(3);

    reportPanel.prevStep();
    expect(reportPanel.currentStep()).toBe(2);

    reportPanel.goToStep(3);
    expect(reportPanel.currentStep()).toBe(3);
  });
});

