import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  appId: 'ai.coexistshield.ionic',
  appName: 'CoExist Shield AI',
  webDir: 'dist/ionic-app/browser',
  server: {
    androidScheme: 'https'
  },
  plugins: {
    LocalNotifications: {
      smallIcon: 'ic_stat_shield',
      iconColor: '#EF4444',
      sound: 'siren_alert.wav'
    }
  }
};

export default config;
