# Local backend hosting (operator PC + Cloudflare Tunnel)

The Go API cannot run on Workers, so this PC hosts it and a Cloudflare quick
tunnel publishes it. A tiny `api-proxy` worker gives consoles one stable URL.

## Topology

```text
browser -> console workers (web/studio/partners/admin)
           NUXT_INTERNAL_API_URL=https://global-video-platform-api.<sub>.workers.dev
        -> api-proxy worker (API_UPSTREAM secret = quick-tunnel URL)
        -> cloudflared quick tunnel (*.trycloudflare.com)
        -> gvp-api.exe :18092 + gvp-media-worker.exe :8081 (this PC)
```

## Files

- `Start-GvpApi.ps1` — starts API + media worker (idempotent; logs under
  `%LOCALAPPDATA%\gvp\logs`).
- `Start-GvpTunnel.ps1` — starts the tunnel, waits for its URL, updates the
  api-proxy `API_UPSTREAM` secret. Reads the Cloudflare API token from
  `%APPDATA%\gvp\cf_api_token.txt` (operator-created, never in the repo).
- `Register-GvpTasks.ps1` — registers `GVP-API` + `GVP-Tunnel` scheduled tasks
  (ONLOGON, current user) and starts both immediately.

## Ports

API `:18092`, media worker `:8081` (`MEDIA_WORKER_URL` default). Port 8080 on
this machine belongs to another occupant and is intentionally avoided.

## Tunnel URL rotation

Quick-tunnel URLs change whenever cloudflared restarts. Re-run
`Start-GvpTunnel.ps1` (or let the scheduled task do it at logon); it rewrites
the api-proxy secret, so consoles need no redeploy. For a fixed hostname, add
a custom domain to the Cloudflare account and switch to a named tunnel.

## Production hardening (before real users)

- Provision OIDC and set `GVP_ENV=production` (dev tokens are currently on).
- Set real `DATABASE_URL`, payment keys, R2 credentials, `SENTRY_DSN`
  (see the mission report `.env` guide).
- Move off quick tunnels to a named tunnel or a hosted container.
