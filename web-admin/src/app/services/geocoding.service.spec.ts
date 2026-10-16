import { TestBed } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { GeocodingService } from './geocoding.service';

describe('GeocodingService', () => {
  let service: GeocodingService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [GeocodingService]
    });
    service = TestBed.inject(GeocodingService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('resolveRegionalKnowledge', () => {
    it('should resolve Kasugayama, Nara coordinates correctly', () => {
      const result = service.resolveRegionalKnowledge(34.685, 135.832);
      expect(result).toBe('Kasugayama Primeval Forest, Nara');
    });

    it('should resolve Rural Route 3219, Kaeng Krachan correctly', () => {
      const result = service.resolveRegionalKnowledge(12.875, 99.650);
      expect(result).toBe('Rural Route 3219, Kaeng Krachan');
    });

    it('should resolve Morioka Basin Route, Iwate correctly', () => {
      const result = service.resolveRegionalKnowledge(39.702, 141.152);
      expect(result).toBe('Morioka Basin Route, Iwate');
    });

    it('should resolve Riau Riparian Corridor, Sumatra correctly', () => {
      const result = service.resolveRegionalKnowledge(0.825, 101.530);
      expect(result).toBe('Riau Riparian Corridor, Sumatra');
    });

    it('should fallback to coordinate sector for uncataloged regions', () => {
      const result = service.resolveRegionalKnowledge(20.0, 50.0);
      expect(result).toBe('Sector (Lat 20.0000, Lng 50.0000)');
    });
  });

  describe('reverseGeocode fallback stream', () => {
    it('should emit regional knowledge when network fails or geocoder is absent', (done) => {
      service.reverseGeocode(34.685, 135.832).subscribe(placeName => {
        expect(placeName).toContain('Nara');
        done();
      });
    });
  });
});
