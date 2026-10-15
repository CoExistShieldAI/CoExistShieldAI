import { Routes } from '@angular/router';
import { DashboardComponent } from './components/dashboard/dashboard.component';
import { GeofenceMapComponent } from './components/geofence-map/geofence-map.component';
import { WildlifeComponent } from './components/wildlife/wildlife.component';
import { RoadkillComponent } from './components/roadkill/roadkill.component';
import { EnvironmentalComponent } from './components/environmental/environmental.component';
import { OperationsComponent } from './components/operations/operations.component';
import { WeatherComponent } from './components/weather/weather.component';

export const routes: Routes = [
  { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
  { path: 'dashboard', component: DashboardComponent },
  { path: 'geofences', component: GeofenceMapComponent },
  { path: 'incidents/wildlife', component: WildlifeComponent },
  { path: 'incidents/roadkill', component: RoadkillComponent },
  { path: 'environmental', component: EnvironmentalComponent },
  { path: 'weather', component: WeatherComponent },
  { path: 'operations', component: OperationsComponent },
  { path: '**', redirectTo: 'dashboard' }
];
