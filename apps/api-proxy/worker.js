// apps/api-proxy/worker.js
// Stable public entry for the StreamVerse Go API running behind a Cloudflare
// quick tunnel on an operator PC. Consoles point NUXT_INTERNAL_API_URL (and the
// web live-chat socket) at this worker; only API_UPSTREAM changes when the
// tunnel URL rotates, via `wrangler secret put API_UPSTREAM`.
export default {
  async fetch(request, env) {
    const upstream = (env.API_UPSTREAM || '').replace(/\/$/, '');
    if (!upstream) {
      return new Response('API_UPSTREAM not configured', { status: 503 });
    }
    if (new URL(request.url).pathname === '/healthz' && request.method === 'GET') {
      return Response.json({ status: 'UP', proxy: 'gvp-api-proxy' });
    }
    const url = new URL(request.url);
    const target = upstream + url.pathname + url.search;
    const headers = new Headers(request.headers);
    headers.set('Host', new URL(upstream).host);
    headers.set('X-Forwarded-Proto', 'https');
    headers.delete('cf-connecting-ip');
    return fetch(target, {
      method: request.method,
      headers,
      body: request.body,
      redirect: 'manual',
      duplex: 'half'
    });
  }
};
