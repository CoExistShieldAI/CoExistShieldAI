import { TestBed } from '@angular/core/testing';
import { FoldableDeviceService, FoldPosture, FoldSimulationMode } from './foldable-device.service';

describe('FoldableDeviceService', () => {
  let service: FoldableDeviceService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [FoldableDeviceService]
    });
    service = TestBed.inject(FoldableDeviceService);
  });

  it('should be created and default to AUTO simulation mode', () => {
    expect(service).toBeTruthy();
    expect(service.simulationMode()).toBe('AUTO');
  });

  it('should transition correctly through cycleSimulationMode()', () => {
    expect(service.simulationMode()).toBe('AUTO');

    // 1st cycle: AUTO -> BOOK
    service.cycleSimulationMode();
    expect(service.simulationMode()).toBe('BOOK');
    expect(service.activePosture()).toBe('book');
    expect(service.isBookMode()).toBeTrue();
    expect(service.isDualScreen()).toBeTrue();
    expect(service.isTabletopMode()).toBeFalse();

    // 2nd cycle: BOOK -> TABLETOP
    service.cycleSimulationMode();
    expect(service.simulationMode()).toBe('TABLETOP');
    expect(service.activePosture()).toBe('tabletop');
    expect(service.isTabletopMode()).toBeTrue();
    expect(service.isDualScreen()).toBeTrue();
    expect(service.isBookMode()).toBeFalse();

    // 3rd cycle: TABLETOP -> SINGLE
    service.cycleSimulationMode();
    expect(service.simulationMode()).toBe('SINGLE');
    expect(service.activePosture()).toBe('single');
    expect(service.isDualScreen()).toBeFalse();
    expect(service.isBookMode()).toBeFalse();
    expect(service.isTabletopMode()).toBeFalse();

    // 4th cycle: SINGLE -> AUTO
    service.cycleSimulationMode();
    expect(service.simulationMode()).toBe('AUTO');
  });

  it('should allow direct setting of simulation mode', () => {
    service.setSimulationMode('BOOK');
    expect(service.activePosture()).toBe('book');
    expect(service.postureBadge().label).toContain('Book');

    service.setSimulationMode('TABLETOP');
    expect(service.activePosture()).toBe('tabletop');
    expect(service.postureBadge().label).toContain('Tabletop');

    service.setSimulationMode('SINGLE');
    expect(service.activePosture()).toBe('single');
    expect(service.postureBadge().label).toContain('Phone');
  });

  it('should evaluate hardware posture gracefully without crashing', () => {
    expect(() => service.evaluateHardwarePosture()).not.toThrow();
  });

  it('should toggle map maximization between dual split and full wide map', () => {
    expect(service.isMapMaximized()).toBeFalse();

    service.toggleMapMaximize();
    expect(service.isMapMaximized()).toBeTrue();

    service.toggleMapMaximize();
    expect(service.isMapMaximized()).toBeFalse();

    service.setMapMaximized(true);
    expect(service.isMapMaximized()).toBeTrue();
  });
});

