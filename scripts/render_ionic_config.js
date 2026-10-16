#!/usr/bin/env node
const fs = require('fs');
const path = require('path');

const cwd = process.cwd();
const candidates = [
  process.argv[2] ? path.resolve(cwd, process.argv[2]) : null,
  path.resolve(cwd, 'ionic-app/dist/ionic-app/browser/index.html'),
  path.resolve(cwd, 'dist/ionic-app/browser/index.html'),
  path.resolve(__dirname, '../ionic-app/dist/ionic-app/browser/index.html')
].filter(Boolean);

let targetFile = candidates.find(p => fs.existsSync(p));

if (!targetFile) {
  console.error('Error: Could not locate ionic-app dist index.html. Did you run ng build first?');
  process.exit(1);
}

let mapsKey = (process.env.GOOGLE_MAPS_API_KEY || '').trim();
let firebaseKey = (process.env.COEXIST_FIREBASE_API_KEY || '').trim();
let backendUrl = (process.env.COEXIST_BACKEND_URL || '').trim();

const envPaths = [
  path.resolve(cwd, 'ionic-app/.env'),
  path.resolve(cwd, 'web-admin/.env'),
  path.resolve(cwd, '.env'),
  path.resolve(__dirname, '../ionic-app/.env'),
  path.resolve(__dirname, '../web-admin/.env'),
  path.resolve(__dirname, '../.env')
];

for (const p of envPaths) {
  if (fs.existsSync(p)) {
    const content = fs.readFileSync(p, 'utf8');
    for (const line of content.split('\n')) {
      const trimmed = line.trim();
      if (trimmed && !trimmed.startsWith('#') && trimmed.includes('=')) {
        const [k, ...v] = trimmed.split('=');
        const keyName = k.trim();
        const keyVal = v.join('=').trim().replace(/^["']|["']$/g, '');
        if (!mapsKey && keyName === 'GOOGLE_MAPS_API_KEY') {
          mapsKey = keyVal;
        }
        if (!firebaseKey && keyName === 'COEXIST_FIREBASE_API_KEY') {
          firebaseKey = keyVal;
        }
        if (!backendUrl && keyName === 'COEXIST_BACKEND_URL') {
          backendUrl = keyVal;
        }
      }
    }
  }
}

const DEFAULT_FIREBASE_KEY = 'AIzaSyDUaFvULclzZTv7FMgjXoUGWiWLBhn15Yo';
const DEFAULT_MAPS_KEY = 'AIzaSyBEhSnPw0oqxK_KZ0K4s8pZSBvVO5COg8o';

if (!firebaseKey) {
  console.log('Notice: COEXIST_FIREBASE_API_KEY not provided in environment; using project default key.');
  firebaseKey = DEFAULT_FIREBASE_KEY;
}

if (!mapsKey) {
  console.log('Notice: GOOGLE_MAPS_API_KEY not provided in environment; using default key.');
  mapsKey = DEFAULT_MAPS_KEY;
}

if (process.env.GITHUB_ACTIONS === 'true') {
  if (mapsKey) console.log(`::add-mask::${mapsKey}`);
  if (firebaseKey) console.log(`::add-mask::${firebaseKey}`);
}
console.log(`Injecting configuration into ${targetFile}...`);

let content = fs.readFileSync(targetFile, 'utf8');
content = content.replace(/__GOOGLE_MAPS_API_KEY__/g, mapsKey);
content = content.replace(/__COEXIST_FIREBASE_API_KEY__/g, firebaseKey);
fs.writeFileSync(targetFile, content, 'utf8');

const distDir = path.dirname(targetFile);
const configJsPath = path.join(distDir, 'config.js');
const configJsContent = `// Generated at build/deploy time (scripts/render_ionic_config.js)\nwindow.COEXIST_FIREBASE_API_KEY = ${JSON.stringify(firebaseKey)};\nwindow.GOOGLE_MAPS_API_KEY = ${JSON.stringify(mapsKey)};\nwindow.COEXIST_BACKEND_URL = ${JSON.stringify(backendUrl)};\n`;
fs.writeFileSync(configJsPath, configJsContent, 'utf8');

console.log(`Successfully rendered configuration in ${targetFile} and ${configJsPath}`);
