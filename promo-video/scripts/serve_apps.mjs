import http from 'node:http';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const ADMIN_DIR = path.resolve(__dirname, '../../web-admin/dist/web-admin/browser');
const IONIC_DIR = path.resolve(__dirname, '../../ionic-app/dist/ionic-app/browser');

const MIME_TYPES = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'application/javascript; charset=utf-8',
  '.mjs': 'application/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.svg': 'image/svg+xml',
  '.ico': 'image/x-icon',
  '.map': 'application/json',
};

function createStaticServer(rootDir, port, name) {
  const server = http.createServer((req, res) => {
    res.setHeader('Access-Control-Allow-Origin', '*');

    let reqPath = req.url?.split('?')[0] || '/';
    if (reqPath === '/') {
      reqPath = '/index.html';
    }

    let filePath = path.join(rootDir, reqPath);

    // SPA routing fallback: if file doesn't exist or doesn't have extension, serve index.html
    if (!fs.existsSync(filePath) || !path.extname(filePath)) {
      filePath = path.join(rootDir, 'index.html');
    }

    fs.stat(filePath, (err, stats) => {
      if (err || !stats.isFile()) {
        res.writeHead(404, { 'Content-Type': 'text/plain' });
        res.end('Not Found');
        return;
      }

      const ext = path.extname(filePath).toLowerCase();
      const contentType = MIME_TYPES[ext] || 'application/octet-stream';

      res.writeHead(200, {
        'Content-Type': contentType,
        'Content-Length': stats.size,
        'Cache-Control': 'no-cache',
      });

      fs.createReadStream(filePath).pipe(res);
    });
  });

  server.listen(port, '0.0.0.0', () => {
    console.log(`[${name}] Serving at http://localhost:${port}`);
  });

  return server;
}

export function startServers() {
  const adminServer = createStaticServer(ADMIN_DIR, 8086, 'Web-Admin');
  const ionicServer = createStaticServer(IONIC_DIR, 8087, 'Ionic-App');
  return { adminServer, ionicServer };
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  startServers();
}
