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

// Fallback: check local .env files if not provided in environment
if (!key) {
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
          if (k.trim() === 'GOOGLE_MAPS_API_KEY') {
            key = v.join('=').trim().replace(/^["']|["']$/g, '');
            break;
          }
        }
      }
      if (key) break;
    }
  }
}

if (key) {
  if (process.env.GITHUB_ACTIONS === 'true') {
    console.log(`::add-mask::${key}`);
  }
  console.log(`Injecting provided GOOGLE_MAPS_API_KEY into ${targetFile}...`);
} else {
  console.warn('::warning::GOOGLE_MAPS_API_KEY is empty or not set. Replacing placeholder with empty string.');
}

let content = fs.readFileSync(targetFile, 'utf8');
content = content.replace(/__GOOGLE_MAPS_API_KEY__/g, key);
fs.writeFileSync(targetFile, content, 'utf8');

console.log(`Successfully rendered Google Maps configuration in ${targetFile}`);
