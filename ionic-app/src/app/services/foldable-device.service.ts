import { Injectable, signal, computed, inject, DestroyRef } from '@angular/core';

export type FoldPosture = 'single' | 'book' | 'tabletop';
export type FoldSimulationMode = 'AUTO' | 'BOOK' | 'TABLETOP' | 'SINGLE';

export interface ViewportSegment {
  x: number;
  y: number;
  width: number;
  height: number;
}

@Injectable({
  providedIn: 'root'
})
export class FoldableDeviceService {
  private destroyRef = inject(DestroyRef);

  // Manual or automatic simulation state
  readonly simulationMode = signal<FoldSimulationMode>('AUTO');

  // Hardware detected posture
  readonly detectedPosture = signal<FoldPosture>('single');

  // Detected physical viewport segments if supported by browser/device
  readonly detectedSegments = signal<ViewportSegment[]>([]);

  // Computed effective posture considering manual simulation override
  readonly activePosture = computed<FoldPosture>(() => {
    const sim = this.simulationMode();
    if (sim === 'BOOK') return 'book';
    if (sim === 'TABLETOP') return 'tabletop';
    if (sim === 'SINGLE') return 'single';
    return this.detectedPosture();
  });

  readonly isDualScreen = computed<boolean>(() => this.activePosture() !== 'single');
  readonly isBookMode = computed<boolean>(() => this.activePosture() === 'book');
  readonly isTabletopMode = computed<boolean>(() => this.activePosture() === 'tabletop');

  readonly postureBadge = computed<{ label: string; icon: string; cssClass: string }>(() => {
    const posture = this.activePosture();
    const sim = this.simulationMode();
    const isSimulated = sim !== 'AUTO';

    switch (posture) {
      case 'book':
        return {
          label: isSimulated ? 'Book (Simulated)' : 'Pixel Fold (Book)',
          icon: '📖',
          cssClass: 'bg-indigo-950/80 border-indigo-500/80 text-indigo-300'
        };
      case 'tabletop':
        return {
          label: isSimulated ? 'Tabletop (Simulated)' : 'Tabletop Mode',
          icon: '📐',
          cssClass: 'bg-amber-950/80 border-amber-500/80 text-amber-300'
        };
      case 'single':
      default:
        return {
          label: isSimulated ? 'Phone (Simulated)' : 'Phone / Cover',
          icon: '📱',
          cssClass: 'bg-slate-800 border-slate-700 text-slate-300'
        };
    }
  });

  private horizontalMq: MediaQueryList | null = null;
  private verticalMq: MediaQueryList | null = null;
  private resizeListener: (() => void) | null = null;

  constructor() {
    this.initFoldableListeners();
  }

  private initFoldableListeners() {
    if (typeof window === 'undefined') return;

    try {
      this.horizontalMq = window.matchMedia('(horizontal-viewport-segments: 2)');
      this.verticalMq = window.matchMedia('(vertical-viewport-segments: 2)');

      const handleMqChange = () => this.evaluateHardwarePosture();

      if (this.horizontalMq.addEventListener) {
        this.horizontalMq.addEventListener('change', handleMqChange);
        this.verticalMq?.addEventListener('change', handleMqChange);
      } else if (this.horizontalMq.addListener) {
        // Fallback for older Safari/browsers
        this.horizontalMq.addListener(handleMqChange);
        this.verticalMq?.addListener(handleMqChange);
      }

      this.resizeListener = () => this.evaluateHardwarePosture();
      window.addEventListener('resize', this.resizeListener);

      this.destroyRef.onDestroy(() => {
        if (this.horizontalMq?.removeEventListener) {
          this.horizontalMq.removeEventListener('change', handleMqChange);
          this.verticalMq?.removeEventListener('change', handleMqChange);
        }
        if (this.resizeListener) {
          window.removeEventListener('resize', this.resizeListener);
        }
      });
    } catch {
      // In case media matchers throw or in test runner
    }

    this.evaluateHardwarePosture();
  }

  /**
   * Evaluates hardware screen dimensions and Web Foldable Media Queries.
   */
  evaluateHardwarePosture() {
    if (typeof window === 'undefined') return;

    // 1. Check Web Foldable Viewport Segments API
    if (this.horizontalMq?.matches) {
      this.detectedPosture.set('book');
      this.readVisualViewportSegments();
      return;
    }
    if (this.verticalMq?.matches) {
      this.detectedPosture.set('tabletop');
      this.readVisualViewportSegments();
      return;
    }

    // 2. Check window.visualViewport.segments if available (Dual Screen / Surface Duo / Foldables)
    const visualViewport = (window as any).visualViewport;
    if (visualViewport && Array.isArray(visualViewport.segments) && visualViewport.segments.length === 2) {
      const seg1 = visualViewport.segments[0];
      const seg2 = visualViewport.segments[1];
      if (seg1.top === seg2.top) {
        this.detectedPosture.set('book');
      } else {
        this.detectedPosture.set('tabletop');
      }
      this.detectedSegments.set(visualViewport.segments);
      return;
    }

    // 3. Fallback heuristic: Detect wide aspect ratio foldables like unfolded Pixel Fold or Surface Duo
    const width = window.innerWidth;
    const height = window.innerHeight;
    const aspect = width / (height || 1);

    // Unfolded Pixel Fold has inner display around 6:5 aspect ratio (aspect ~ 1.2 or 0.83 in portrait)
    // Surface Duo dual-portrait is ~ 1.35 aspect ratio with width > 1000px
    if (width >= 800 && width <= 1400 && aspect >= 0.85 && aspect <= 1.4) {
      // If window is wide with squarish aspect ratio characteristic of an unfolded book device
      // (only if not a regular desktop; checked on touch/mobile platforms)
      const isTouch = 'ontouchstart' in window || navigator.maxTouchPoints > 0;
      if (isTouch) {
        this.detectedPosture.set('book');
        return;
      }
    }

    this.detectedPosture.set('single');
  }

  private readVisualViewportSegments() {
    const visualViewport = (window as any).visualViewport;
    if (visualViewport && Array.isArray(visualViewport.segments)) {
      this.detectedSegments.set(visualViewport.segments);
    }
  }

  /**
   * Sets manual simulation mode for judge / evaluator demos.
   */
  setSimulationMode(mode: FoldSimulationMode) {
    this.simulationMode.set(mode);
  }

  /**
   * Cycle simulation mode: AUTO -> BOOK -> TABLETOP -> SINGLE -> AUTO
   */
  cycleSimulationMode() {
    const current = this.simulationMode();
    switch (current) {
      case 'AUTO':
        this.simulationMode.set('BOOK');
        break;
      case 'BOOK':
        this.simulationMode.set('TABLETOP');
        break;
      case 'TABLETOP':
        this.simulationMode.set('SINGLE');
        break;
      case 'SINGLE':
      default:
        this.simulationMode.set('AUTO');
        break;
    }
  }
}
