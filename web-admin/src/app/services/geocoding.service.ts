import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, from, of } from 'rxjs';
import { catchError, map, timeout } from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class GeocodingService {
  private http = inject(HttpClient);

  /**
   * Reverse geocodes latitude & longitude into a human-readable location or road name.
   * Priority pipeline:
   * 1. Google Maps JavaScript API Geocoder (if SDK is loaded)
   * 2. OpenStreetMap Nominatim reverse geocoding (with short timeout)
   * 3. Comprehensive JAPAC regional GIS boundary knowledge base
   */
  reverseGeocode(lat: number, lng: number): Observable<string> {
    if (this.isGoogleMapsGeocoderAvailable()) {
      return this.reverseGeocodeWithGoogle(lat, lng).pipe(
        catchError(() => this.fallbackReverseGeocode(lat, lng))
      );
    }
    return this.fallbackReverseGeocode(lat, lng);
  }

  private isGoogleMapsGeocoderAvailable(): boolean {
    return typeof google !== 'undefined' &&
           typeof google.maps !== 'undefined' &&
           typeof google.maps.Geocoder === 'function';
  }

  private reverseGeocodeWithGoogle(lat: number, lng: number): Observable<string> {
    return new Observable<string>(observer => {
      try {
        const geocoder = new google.maps.Geocoder();
        geocoder.geocode({ location: { lat, lng } }, (results, status) => {
          if (status === google.maps.GeocoderStatus.OK && results && results.length > 0) {
            const extracted = this.extractPlaceNameFromGoogleResults(results);
            observer.next(extracted);
            observer.complete();
          } else {
            observer.error(new Error(`Google Geocoder status: ${status}`));
          }
        });
      } catch (err) {
        observer.error(err);
      }
    }).pipe(timeout(3000));
  }

  private extractPlaceNameFromGoogleResults(results: google.maps.GeocoderResult[]): string {
    const first = results[0];
    const components = first.address_components || [];

    // Check for high-relevance POI, park, natural feature, or establishment
    const poi = components.find(c =>
      c.types.includes('point_of_interest') ||
      c.types.includes('natural_feature') ||
      c.types.includes('park') ||
      c.types.includes('establishment')
    );

    const route = components.find(c => c.types.includes('route'));
    const locality = components.find(c =>
      c.types.includes('locality') ||
      c.types.includes('sublocality') ||
      c.types.includes('administrative_area_level_2')
    );

    if (poi && poi.long_name) {
      if (locality && locality.long_name !== poi.long_name) {
        return `${poi.long_name}, ${locality.long_name}`;
      }
      return poi.long_name;
    }

    if (route && locality) {
      return `${route.long_name}, ${locality.long_name}`;
    } else if (route) {
      return route.long_name;
    } else if (locality) {
      return locality.long_name;
    }

    const parts = (first.formatted_address || '').split(',');
    if (parts.length >= 2) {
      return `${parts[0].trim()}, ${parts[1].trim()}`;
    }
    const geo = latLngFromGoogle(first);
    return parts[0]?.trim() || this.resolveRegionalKnowledge(geo.lat, geo.lng);
  }

  private fallbackReverseGeocode(lat: number, lng: number): Observable<string> {
    const url = `https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lng}&zoom=14&addressdetails=1`;
    return this.http.get<any>(url).pipe(
      timeout(2000),
      map(res => {
        if (res && (res.name || res.display_name)) {
          const addr = res.address || {};
          const primary = res.name || addr.road || addr.suburb || addr.village || addr.town || addr.city;
          const secondary = addr.city || addr.county || addr.state;
          if (primary && secondary && primary !== secondary) {
            return `${primary}, ${secondary}`;
          }
          if (primary) return primary;
          const parts = (res.display_name || '').split(',');
          return parts.slice(0, 2).map((p: string) => p.trim()).join(', ');
        }
        return this.resolveRegionalKnowledge(lat, lng);
      }),
      catchError(() => of(this.resolveRegionalKnowledge(lat, lng)))
    );
  }

  /**
   * Deterministic regional GIS reverse geocode knowledge base for CoExist Shield operational territories.
   */
  resolveRegionalKnowledge(lat: number, lng: number): string {
    // 1. Nara, Japan (Asian Black Bear & Persimmon Orchards)
    if (lat >= 34.3 && lat <= 34.85 && lng >= 135.6 && lng <= 136.1) {
      if (lat >= 34.66 && lat <= 34.72 && lng >= 135.81 && lng <= 135.89) {
        return 'Kasugayama Primeval Forest, Nara';
      }
      if (lat >= 34.35 && lat <= 34.52) {
        return 'Yoshino Mountain Foothills, Nara';
      }
      return 'Nara Forest Fringe, Nara';
    }

    // 2. Nagano, Japan (Alps Bear Corridor)
    if (lat >= 35.8 && lat <= 37.0 && lng >= 137.4 && lng <= 138.8) {
      if (lat >= 36.15 && lat <= 36.35 && lng >= 137.85 && lng <= 138.1) {
        return 'Matsumoto Highlands, Nagano';
      }
      return 'Nagano Alpine Sector';
    }

    // 3. Iwate / Morioka, Japan (Persimmon Attractant & Bear Interface)
    if (lat >= 39.4 && lat <= 40.2 && lng >= 140.8 && lng <= 141.8) {
      if (lat >= 39.65 && lat <= 39.75 && lng >= 141.10 && lng <= 141.22) {
        return 'Morioka Basin Route, Iwate';
      }
      return 'Iwate Foothills Corridor';
    }

    // 4. Hokkaido, Japan (Brown Bear Habitats)
    if (lat >= 41.5 && lat <= 45.6 && lng >= 139.5 && lng <= 145.9) {
      if (lat >= 43.8 && lat <= 44.5 && lng >= 144.8 && lng <= 145.4) {
        return 'Shiretoko Peninsula Coast, Hokkaido';
      }
      return 'Hokkaido Wilderness Perimeter';
    }

    // 5. Thailand - Kaeng Krachan & Phetchaburi (Elephant Corridor & Route 3219)
    if (lat >= 12.5 && lat <= 13.5 && lng >= 99.1 && lng <= 100.2) {
      if (lat >= 12.80 && lat <= 12.92 && lng >= 99.60 && lng <= 99.72) {
        return 'Rural Route 3219, Kaeng Krachan';
      }
      return 'Kaeng Krachan Eastern Buffer';
    }

    // 6. Thailand - Khao Ang Rue Nai & Chachoengsao (Elephant Crop Raiding)
    if (lat >= 13.2 && lat <= 13.8 && lng >= 101.2 && lng <= 101.9) {
      return 'Khao Ang Rue Nai Corridor';
    }

    // 7. Indonesia - Riau & Sumatra (Sumatran Tiger Reserve)
    if (lat >= -1.0 && lat <= 2.5 && lng >= 100.0 && lng <= 103.5) {
      if (lat >= 0.7 && lat <= 1.0 && lng >= 101.3 && lng <= 101.8) {
        return 'Riau Riparian Corridor, Sumatra';
      }
      return 'Sumatra Tiger Reserve Perimeter';
    }

    // 8. India - Kaziranga & Assam (Elephant & Bengal Tiger)
    if (lat >= 26.3 && lat <= 26.9 && lng >= 92.8 && lng <= 93.8) {
      return 'Kaziranga Buffer Zone, Assam';
    }

    // 9. India - Valparai / Anamalai (Elephant & Leopard in Tea Plantations)
    if (lat >= 10.2 && lat <= 10.5 && lng >= 76.8 && lng <= 77.15) {
      return 'Valparai Tea Plateau Corridor, Tamil Nadu';
    }

    // 10. Sri Lanka - Major Human-Elephant & Sloth Bear Conflict Corridors
    if (lat >= 5.8 && lat <= 9.9 && lng >= 79.5 && lng <= 82.0) {
      if (lat >= 7.8 && lat <= 8.3 && lng >= 80.7 && lng <= 81.3) {
        return 'Minneriya-Polonnaruwa Paddy Corridor, Sri Lanka';
      }
      if (lat >= 6.3 && lat <= 6.6 && lng >= 80.7 && lng <= 81.1) {
        return 'Udawalawe-Sevenagala Sugar Basin, Sri Lanka';
      }
      if (lat >= 6.3 && lat <= 6.85 && lng >= 81.1 && lng <= 81.65) {
        return 'Yala-Pelwatte Wildlife Buffer, Sri Lanka';
      }
      if (lat >= 8.2 && lat <= 8.7 && lng >= 80.0 && lng <= 80.6) {
        return 'Wilpattu-Anuradhapura Forest Fringe, Sri Lanka';
      }
      return 'Sri Lanka Dry-Zone Wildlife Corridor';
    }

    // 11. Indonesia - Bangka Belitung Islands (Saltwater Crocodile Mining Estuary Hotspot)
    if (lat >= -3.3 && lat <= -1.4 && lng >= 105.0 && lng <= 107.2) {
      return 'Bangka-Belitung Estuary Crocodile Zone, Indonesia';
    }

    return `Sector (Lat ${lat.toFixed(4)}, Lng ${lng.toFixed(4)})`;
  }
}

function latLngFromGoogle(r: google.maps.GeocoderResult): { lat: number; lng: number } {
  return {
    lat: r.geometry?.location ? r.geometry.location.lat() : 0,
    lng: r.geometry?.location ? r.geometry.location.lng() : 0
  };
}
