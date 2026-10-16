#!/usr/bin/env node
const fs = require('fs');
const path = require('path');

const cwd = process.cwd();
const candidates = [
  process.argv[2] ? path.resolve(cwd, process.argv[2]) : null,
  path.resolve(cwd, 'web-admin/dist/web-admin/browser/index.html'),
  path.resolve(cwd, 'dist/web-admin/browser/index.html'),
  path.resolve(__dirname, '../web-admin/dist/web-admin/browser/index.html')
].filter(Boolean);

let targetFile = candidates.find(p => fs.existsSync(p));

if (!targetFile) {
  console.error('Error: Could not locate web-admin dist index.html. Did you run ng build first?');
  process.exit(1);
}

let key = (process.env.GOOGLE_MAPS_API_KEY || '').trim();
let backendUrl = (process.env.COEXIST_BACKEND_URL || '').trim();

// Fallback: check local .env files if not provided in environment
const envPaths = [
  path.resolve(cwd, 'web-admin/.env'),
  path.resolve(cwd, '.env'),
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
        if (!key && keyName === 'GOOGLE_MAPS_API_KEY') {
          key = keyVal;
        }
        if (!backendUrl && keyName === 'COEXIST_BACKEND_URL') {
          backendUrl = keyVal;
        }
      }
    }
  }
}

const DEFAULT_MAPS_KEY = 'AIzaSyBEhSnPw0oqxK_KZ0K4s8pZSBvVO5COg8o';

if (!key) {
  console.log('Notice: GOOGLE_MAPS_API_KEY not provided in environment; using default key.');
  key = DEFAULT_MAPS_KEY;
}

if (process.env.GITHUB_ACTIONS === 'true') {
  if (key) console.log(`::add-mask::${key}`);
}
console.log(`Injecting provided GOOGLE_MAPS_API_KEY into ${targetFile}...`);

let content = fs.readFileSync(targetFile, 'utf8');
if (content.includes('__GOOGLE_MAPS_API_KEY__')) {
  content = content.replace(/__GOOGLE_MAPS_API_KEY__/g, key);
} else {
  content = content.replace(/maps\.googleapis\.com\/maps\/api\/js\?key=[^&"']*/g, `maps.googleapis.com/maps/api/js?key=${key}`);
  content = content.replace(/window\.GOOGLE_MAPS_API_KEY\s*=\s*(?:'[^']*'|"[^"]*"|`[^`]*`|__GOOGLE_MAPS_API_KEY__)/g, `window.GOOGLE_MAPS_API_KEY = ${JSON.stringify(key)}`);
}
fs.writeFileSync(targetFile, content, 'utf8');

const distDir = path.dirname(targetFile);
const configJsPath = path.join(distDir, 'config.js');
const configJsContent = `// Generated at build/deploy time (scripts/render_admin_config.js)\nwindow.GOOGLE_MAPS_API_KEY = ${JSON.stringify(key)};\nwindow.COEXIST_BACKEND_URL = ${JSON.stringify(backendUrl)};\n`;
fs.writeFileSync(configJsPath, configJsContent, 'utf8');

console.log(`Successfully rendered configuration in ${targetFile} and ${configJsPath}`);

console.log(`Successfully rendered Google Maps configuration in ${targetFile}`);
