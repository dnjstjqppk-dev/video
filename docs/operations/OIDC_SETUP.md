# OIDC Issuer Provisioning (ZITADEL)

The API accepts RS256 JWTs from any OIDC-compliant issuer. Production refuses
to boot without one (`config.Validate` + startup discovery probe). The Nuxt
consoles use Authorization Code + PKCE against the same issuer.

## 1. Create the ZITADEL project

1. Console → Projects → New project named `vravra-production`.
2. Add an application per console surface:
   - `web` — redirect `https://<web>/auth/callback`, logout `https://<web>/`,
     PKCE required, auth method `none` (public client with secret for server exchange).
   - `studio`, `partners`, `admin` — same pattern with their own origins.
3. Grant the API project resource ID as audience (used for `OIDC_AUDIENCE`).
4. Copy each application's Client ID + Secret into the matching console env.

## 2. Wire the API

```env
GVP_ENV=production
OIDC_ISSUER=https://<tenant>.zitadel.cloud
OIDC_AUDIENCE=<api-project-resource-id>
OIDC_JWKS_URL=https://<tenant>.zitadel.cloud/oauth/v2/keys
JWT_SECRET=<unique 32+ char secret>
GVP_ALLOW_DEV_TOKENS=<unset>
```

(`JWT_SECRET` length is a recommendation — the code checks presence, not length.)

Rules enforced at boot:

- Any missing value → fatal with this document referenced.
- `auth.example.com` placeholder → fatal.
- Issuer discovery (`/.well-known/openid-configuration`) unreachable or
  issuer-mismatched → fatal in production, warning elsewhere.

## 3. Wire each Nuxt console

```env
NUXT_OIDC_ISSUER=https://<tenant>.zitadel.cloud
NUXT_OIDC_CLIENT_ID=<console application id>
NUXT_OIDC_CLIENT_SECRET=<console application secret>
```

(The callback is derived from the request origin; no redirect-URI variable is read.)

## 4. Verify

1. `go run ./cmd/server` (services/api) → log line
   `OIDC issuer discovery verified`.
2. Open each console → login → redirected to ZITADEL → back with session.
3. `curl -H "Authorization: Bearer <token>" $API/v1/me` returns the profile
   (the API reads the header only; browsers go through the console `/api`
   proxy, which injects it from the session cookie).
4. Revoke the grant in ZITADEL → already-issued tokens stay valid until `exp`
   (no revocation introspection); new logins stop working.

## 5. Local development

Leave `OIDC_ISSUER` unset and set `GVP_ALLOW_DEV_TOKENS=true` with
`GVP_ENV=development`. Dev tokens (`creator-token`, `viewer-token`,
`admin-token`, `advertiser-token`, or any exact user UUID) never work in
production: the middleware gates them on environment, and the dev-token
endpoint only serves localhost.

> **Current-deployment note (2026-11-20):** the live stack still runs `GVP_ENV=test` + `GVP_ALLOW_DEV_TOKENS=true`
> with no ZITADEL tenant, so demo login is open. Complete §1–§4 and switch to `GVP_ENV=production`
> before any real-user traffic (P0 backlog item 6 in [docs/IMPLEMENTATION_STATUS.md](../IMPLEMENTATION_STATUS.md)).
