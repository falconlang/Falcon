import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { createHash } from 'node:crypto';

const companionAssets = new Map();

function collectRequestBytes(req) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    req.on('data', chunk => chunks.push(chunk));
    req.on('end', () => resolve(Buffer.concat(chunks)));
    req.on('error', reject);
  });
}

function decodePathPart(value = '') {
  return value
    .split('/')
    .filter(Boolean)
    .map(part => decodeURIComponent(part))
    .join('/');
}

function assetKey(projectId, assetName) {
  return `${projectId}\0${assetName}`;
}

function sendJson(res, status, payload) {
  res.statusCode = status;
  res.setHeader('Content-Type', 'application/json; charset=utf-8');
  res.end(JSON.stringify(payload));
}

function companionAssetMiddleware(req, res, next) {
  const url = new URL(req.url, 'http://tensor.local');

  if (req.method === 'PUT' && url.pathname.startsWith('/__tensor_companion_assets/')) {
    const rest = url.pathname.slice('/__tensor_companion_assets/'.length);
    const slash = rest.indexOf('/');
    if (slash === -1) {
      sendJson(res, 400, { error: 'Missing asset path' });
      return;
    }
    const projectId = decodeURIComponent(rest.slice(0, slash));
    const assetName = decodePathPart(rest.slice(slash + 1));
    collectRequestBytes(req).then(bytes => {
      const etag = `"${createHash('sha1').update(bytes).digest('hex')}"`;
      companionAssets.set(assetKey(projectId, assetName), {
        bytes,
        etag,
        contentType: req.headers['content-type'] || 'application/octet-stream',
      });
      sendJson(res, 200, { ok: true, projectId, assetName, size: bytes.length, etag });
    }, error => sendJson(res, 500, { error: error?.message || String(error) }));
    return;
  }

  if (req.method === 'DELETE' && url.pathname.startsWith('/__tensor_companion_assets/')) {
    const projectId = decodeURIComponent(url.pathname.slice('/__tensor_companion_assets/'.length).replace(/\/+$/, ''));
    for (const key of Array.from(companionAssets.keys())) {
      if (key.startsWith(`${projectId}\0`)) companionAssets.delete(key);
    }
    sendJson(res, 200, { ok: true, projectId });
    return;
  }

  if (req.method === 'GET' && url.pathname.startsWith('/ode/download/file/')) {
    const rest = url.pathname.slice('/ode/download/file/'.length);
    const slash = rest.indexOf('/');
    if (slash === -1) {
      sendJson(res, 404, { error: 'Missing asset path' });
      return;
    }
    const projectId = decodeURIComponent(rest.slice(0, slash));
    const assetName = decodePathPart(rest.slice(slash + 1));
    const record = companionAssets.get(assetKey(projectId, assetName));
    if (!record) {
      sendJson(res, 404, { error: 'Asset not found' });
      return;
    }
    if (req.headers['if-none-match'] === record.etag) {
      res.statusCode = 304;
      res.end();
      return;
    }
    res.statusCode = 200;
    res.setHeader('Content-Type', record.contentType);
    res.setHeader('Content-Length', record.bytes.length);
    res.setHeader('ETag', record.etag);
    res.end(record.bytes);
    return;
  }

  next();
}

export default defineConfig({
  plugins: [
    svelte(),
    {
      name: 'tensor-companion-assets',
      configureServer(server) {
        server.middlewares.use(companionAssetMiddleware);
      },
      configurePreviewServer(server) {
        server.middlewares.use(companionAssetMiddleware);
      },
    },
  ],
});
