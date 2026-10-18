import { chromium } from '../../web-e2e/node_modules/playwright/index.mjs';
import path from 'node:path';
import fs from 'node:fs';
import { fileURLToPath } from 'node:url';
import { startServers } from './serve_apps.mjs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const RECORDING_DIR = path.resolve(__dirname, '../public/screen-recordings');
if (!fs.existsSync(RECORDING_DIR)) {
  fs.mkdirSync(RECORDING_DIR, { recursive: true });
}

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function main() {
  console.log('🚀 Starting web-admin (8086) & ionic-app (8087) servers...');
  const { adminServer, ionicServer } = startServers();

  await sleep(2000);

  console.log('🎥 Launching Playwright browser...');
  const browser = await chromium.launch({ headless: true });

  try {
    // ==========================================
    // 1. RECORD WEB-ADMIN (1920x1080 Full HD)
    // ==========================================
    console.log('📹 [1/2] Recording Web-Admin Operations Portal...');
    const adminContext = await browser.newContext({
      viewport: { width: 1920, height: 1080 },
      recordVideo: {
        dir: RECORDING_DIR,
        size: { width: 1920, height: 1080 },
      },
    });

    const adminPage = await adminContext.newPage();
    await adminPage.goto('http://localhost:8086/dashboard', { waitUntil: 'networkidle', timeout: 30000 });
    await sleep(2500);

    // Smooth scroll through dashboard metrics
    console.log('  -> Scrolling dashboard metrics & live field reports...');
    await adminPage.mouse.wheel(0, 350);
    await sleep(2000);
    await adminPage.mouse.wheel(0, -350);
    await sleep(2000);

    // Hover over Run AI Geofence Engine button
    await adminPage.hover('button:has-text("Run AI Geofence Engine")');
    await sleep(1500);

    // Navigate to Geofences Console
    console.log('  -> Navigating to JAPAC Geofences Console...');
    await adminPage.click('a[href="/geofences"]');
    await adminPage.waitForURL('**/geofences', { timeout: 10000 });
    await sleep(3000);

    // Toggle Layers on Geofence Console
    console.log('  -> Toggling Roadkill and Attractant Layers...');
    const checkboxes = await adminPage.locator('input[type="checkbox"]');
    const count = await checkboxes.count();
    if (count > 0) {
      await checkboxes.nth(1).click();
      await sleep(1000);
      await checkboxes.nth(1).click();
      await sleep(1500);
    }

    // Click through active geofence items in sidebar
    const geofenceCards = await adminPage.locator('.cursor-pointer, [class*="ACTIVE GEOFENCES"] ~ div');
    if ((await geofenceCards.count()) > 0) {
      await geofenceCards.first().click().catch(() => {});
      await sleep(2000);
    }

    await sleep(2000);
    const adminVideo = adminPage.video();
    await adminPage.close();
    await adminContext.close();

    if (adminVideo) {
      const rawAdmin = await adminVideo.path();
      const targetAdmin = path.join(RECORDING_DIR, 'admin_demo.mp4');
      const { execSync } = await import('node:child_process');
      execSync(`ffmpeg -y -i "${rawAdmin}" -c:v libx264 -pix_fmt yuv420p "${targetAdmin}"`);
      console.log(`✅ Web-Admin video saved to: ${targetAdmin}`);
    }

    // ==========================================
    // 2. RECORD IONIC-APP (Mobile 412x892)
    // ==========================================
    console.log('📱 [2/2] Recording Ionic Mobile App (Worker & Citizen HUD)...');
    const mobileContext = await browser.newContext({
      viewport: { width: 412, height: 892 },
      isMobile: true,
      hasTouch: true,
      recordVideo: {
        dir: RECORDING_DIR,
        size: { width: 412, height: 892 },
      },
    });

    const mobilePage = await mobileContext.newPage();
    await mobilePage.goto('http://localhost:8087', { waitUntil: 'networkidle', timeout: 30000 });
    await sleep(2500);

    // Examine Alert HUD
    console.log('  -> Interacting with Alert HUD & Proximity Warning...');
    await mobilePage.mouse.wheel(0, 200);
    await sleep(2000);
    await mobilePage.mouse.wheel(0, -200);
    await sleep(2000);

    // Click "Focus Hazard Location on Map"
    const focusBtn = mobilePage.locator('button:has-text("Focus Hazard Location")');
    if (await focusBtn.isVisible().catch(() => false)) {
      await focusBtn.click();
      await sleep(2000);
    }

    // Switch to GIS Map tab
    console.log('  -> Switching to GIS Map tab...');
    const gisTab = mobilePage.locator('button:has-text("GIS Map"), [role="tab"]:has-text("GIS")');
    if (await gisTab.isVisible().catch(() => false)) {
      await gisTab.click();
      await sleep(3000);
    }

    // Switch to Report tab
    console.log('  -> Switching to Report tab...');
    const reportTab = mobilePage.locator('button:has-text("Report"), [role="tab"]:has-text("Report")');
    if (await reportTab.isVisible().catch(() => false)) {
      await reportTab.click();
      await sleep(2500);
    }

    // Switch back to Alert HUD
    const hudTab = mobilePage.locator('button:has-text("Alert HUD"), [role="tab"]:has-text("HUD")');
    if (await hudTab.isVisible().catch(() => false)) {
      await hudTab.click();
      await sleep(2500);
    }

    await sleep(2000);
    const mobileVideo = mobilePage.video();
    await mobilePage.close();
    await mobileContext.close();

    if (mobileVideo) {
      const rawMobile = await mobileVideo.path();
      const targetMobile = path.join(RECORDING_DIR, 'ionic_mobile_demo.mp4');
      const { execSync } = await import('node:child_process');
      execSync(`ffmpeg -y -i "${rawMobile}" -c:v libx264 -pix_fmt yuv420p "${targetMobile}"`);
      console.log(`✅ Ionic Mobile video saved to: ${targetMobile}`);
    }

  } catch (err) {
    console.error('Error during recording:', err);
  } finally {
    await browser.close();
    adminServer.close();
    ionicServer.close();
    console.log('🎉 All screen recordings completed successfully!');
  }
}

main();
