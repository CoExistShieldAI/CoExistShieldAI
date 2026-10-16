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

  it('should accurately step simulated GPS location using D-Pad steering', async () => {
    const fixture = TestBed.createComponent(AppComponent);
    const app = fixture.componentInstance;
    const initialLoc = { latitude: 39.702, longitude: 141.152 };
    await app.state.updateDeviceLocation(initialLoc, 'Initial');

    app.state.setGpsStepSize(25);
    expect(app.state.gpsStepSizeMeters()).toBe(25);

    // Step North
    await app.state.stepSimulatedLocation('NORTH', 100);
    expect(app.state.userLocation().latitude).toBeGreaterThan(initialLoc.latitude);
    expect(app.state.userLocation().longitude).toBeCloseTo(initialLoc.longitude, 5);

    // Step South
    await app.state.stepSimulatedLocation('SOUTH', 100);
    expect(app.state.userLocation().latitude).toBeCloseTo(initialLoc.latitude, 5);

    // Step East
    await app.state.stepSimulatedLocation('EAST', 100);
    expect(app.state.userLocation().longitude).toBeGreaterThan(initialLoc.longitude);

    // Step West
    await app.state.stepSimulatedLocation('WEST', 100);
    expect(app.state.userLocation().longitude).toBeCloseTo(initialLoc.longitude, 5);
  });
});

