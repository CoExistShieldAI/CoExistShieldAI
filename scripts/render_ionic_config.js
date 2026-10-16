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
      }
    }
  }
}

if (mapsKey) {
  if (process.env.GITHUB_ACTIONS === 'true') {
    console.log(`::add-mask::${mapsKey}`);
  }
  console.log(`Injecting provided GOOGLE_MAPS_API_KEY into ${targetFile}...`);
} else {
  console.warn('::warning::GOOGLE_MAPS_API_KEY is empty or not set. Replacing placeholder with empty string.');
}

let content = fs.readFileSync(targetFile, 'utf8');
content = content.replace(/__GOOGLE_MAPS_API_KEY__/g, mapsKey);
content = content.replace(/__COEXIST_FIREBASE_API_KEY__/g, firebaseKey);
fs.writeFileSync(targetFile, content, 'utf8');

const distDir = path.dirname(targetFile);
const configJsPath = path.join(distDir, 'config.js');
const configJsContent = `// Generated at build/deploy time (scripts/render_ionic_config.js)\nwindow.COEXIST_FIREBASE_API_KEY = ${JSON.stringify(firebaseKey)};\nwindow.GOOGLE_MAPS_API_KEY = ${JSON.stringify(mapsKey)};\n`;
fs.writeFileSync(configJsPath, configJsContent, 'utf8');

console.log(`Successfully rendered configuration in ${targetFile} and ${configJsPath}`);
