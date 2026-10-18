import { registerPlugin, WebPlugin, PluginListenerHandle } from '@capacitor/core';
import { AlertTier, WildlifeSpecies } from '../models/shield.model';

/**
 * Cross-platform Smartwatch Synchronization Plugin bridging Ionic/Angular
 * with Wear OS (Android) and Apple Watch (watchOS) smartwatches.
 *
 * - Wear OS: Google Play Services Wearable Data Layer API (DataClient & MessageClient)
 * - Apple Watch: Apple WatchConnectivity framework (WCSession updateApplicationContext & sendMessage)
 */
export interface WearSyncPlugin {
  syncHudState(options: {
    tier: AlertTier;
    distanceMeters: number | null;
    species: WildlifeSpecies[];
    isSiren: boolean;
  }): Promise<{ success: boolean; tier: string }>;

  sendImmediateAlert(options: {
    tier: AlertTier;
    distanceMeters: number | null;
    message: string;
  }): Promise<{ success: boolean; nodesCount: number }>;

  addListener(
    eventName: 'onWatchAction',
    listenerFunc: (event: { action: 'DISMISS' | 'SOS'; timestamp: number }) => void
  ): Promise<PluginListenerHandle>;
}

export class WearSyncWeb extends WebPlugin implements WearSyncPlugin {
  async syncHudState(options: {
    tier: AlertTier;
    distanceMeters: number | null;
    species: WildlifeSpecies[];
    isSiren: boolean;
  }): Promise<{ success: boolean; tier: string }> {
    // Web fallback logging
    return { success: true, tier: options.tier };
  }

  async sendImmediateAlert(options: {
    tier: AlertTier;
    distanceMeters: number | null;
    message: string;
  }): Promise<{ success: boolean; nodesCount: number }> {
    return { success: true, nodesCount: 0 };
  }
}

export const WearSync = registerPlugin<WearSyncPlugin>('WearSync', {
  web: () => new WearSyncWeb()
});

/**
 * Universal WatchSync alias for WearSync supporting both Wear OS and Apple Watch.
 */
export const WatchSync = WearSync;

