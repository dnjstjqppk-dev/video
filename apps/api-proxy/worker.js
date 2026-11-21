// apps/api-proxy/worker.js
// Stable public entry for the StreamVerse Go API running behind a Cloudflare
// quick tunnel on an operator PC. Consoles point NUXT_INTERNAL_API_URL (and the
// web live-chat socket) at this worker; only API_UPSTREAM changes when the
// tunnel URL rotates, via `wrangler secret put API_UPSTREAM`.
const CORS_HEADERS = {
  'Access-Control-Allow-Origin': '*',
  'Access-Control-Allow-Methods': 'GET, HEAD, OPTIONS',
  'Access-Control-Allow-Headers': 'Range, Content-Type, Authorization',
  'Access-Control-Expose-Headers': 'Content-Length, Content-Range, Accept-Ranges',
  'Access-Control-Max-Age': '86400',
};

export default {
  async fetch(request, env) {
    if (request.method === 'OPTIONS') {
      return new Response(null, { status: 204, headers: CORS_HEADERS });
    }
    const upstream = (env.API_UPSTREAM || '').replace(/\/$/, '');
    if (!upstream) {
      return new Response('API_UPSTREAM not configured', { status: 503 });
    }
    if (new URL(request.url).pathname === '/healthz' && request.method === 'GET') {
      return Response.json({ status: 'UP', proxy: 'gvp-api-proxy' }, { headers: CORS_HEADERS });
    }
    const url = new URL(request.url);
    const target = upstream + url.pathname + url.search;
    // WebSocket upgrade passthrough (live chat): forward the request as-is
    // so the 101 Switching Protocols handshake reaches the API hub.
    if (request.headers.get('Upgrade')?.toLowerCase() === 'websocket') {
      return fetch(new Request(target, request));
    }
    const headers = new Headers(request.headers);
    headers.set('Host', new URL(upstream).host);
    headers.set('X-Forwarded-Proto', 'https');
    headers.delete('cf-connecting-ip');
    const upstreamResp = await fetch(target, {
      method: request.method,
      headers,
      body: request.body,
      redirect: 'manual',
      duplex: 'half'
    });
    // Re-wrap to attach CORS for direct browser media fetches (proxied
    // streams pass through untouched; Range in/out is preserved).
    const respHeaders = new Headers(upstreamResp.headers);
    for (const [k, v] of Object.entries(CORS_HEADERS)) {
      respHeaders.set(k, v);
    }
    return new Response(upstreamResp.body, {
      status: upstreamResp.status,
      statusText: upstreamResp.statusText,
      headers: respHeaders,
    });
  }
};
