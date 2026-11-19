# vravra Deployment — Consoles, API Tunnel and Rollback

## 1. Pipeline

Pushes to `master` trigger `.github/workflows/cloudflare-deploy.yml`, which
builds and deploys the four consoles in a matrix (`web`, `studio`, `partners`,
`admin`) with the Cloudflare Nitro preset:

```bash
pnpm install --frozen-lockfile
pnpm --filter @gvp/<app> build
# wrangler deploy .output/server/index.mjs --assets .output/public
```

The `apps/api-proxy` worker is deployed manually (see its README) and holds
the `API_UPSTREAM` secret so console configs never embed tunnel URLs.

## 2. Runtime topology

```text
users -> console workers (*.workers.dev, /api) -> api-proxy worker
      -> cloudflared quick tunnel -> Go API :18092 + media worker :8081
```

- Port `8080` is reserved on the tunnel host; the API listens on `:18092`
  (see `infra/local/README.md`).
- Tunnels rotate host names on restart — only `API_UPSTREAM` needs updating.

## 3. Environment and secrets

- Console public URLs: `NUXT_PUBLIC_{SITE,STUDIO,PARTNERS,ADMIN}_URL`.
- API: `NUXT_INTERNAL_API_URL` (server-side), OIDC issuer/client/secret.
- Production requires `GVP_ENV=production`, dev tokens disabled, and a real
  ZITADEL tenant — full gate list in `docs/operations/PRODUCTION_CHECKLIST.md`.

## 4. Verify and roll back

```bash
# Health (via proxy)
curl -s https://<console-host>/api/v1/version
# Local production parity
pnpm --filter @gvp/web build && npx wrangler dev .output/server/index.mjs --assets .output/public
```

Roll back with `wrangler rollback` on the affected worker, or redeploy the
previous green commit from the Actions tab. Keep `CHANGELOG.md` current so
each deploy maps to a release entry.
