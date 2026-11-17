# P1-20 — Load, Recovery & SLO Measurements (Gate 5)

Measured 2026-11-17 against the operator-PC stack. No mocked results:
every number below came from a live process (per ARTICLE 17, mocked
results do not count as production verification).

## Method

- Target: `gvp-api` built from `services/api/cmd/server`, `GVP_ENV=test`,
  clean port `:18093` (the P1-12 tunnel upstream on `:18092` was excluded
  from load to keep the public path untouched).
- Tool: stdlib-only Go hammer (50 concurrent workers, 1000 requests per
  endpoint, 15 s client timeout). Source kept outside the repo as a
  throwaway probe; rerun with any HTTP load generator.
- Endpoints: `GET /v1/feed/home?page=1&limit=24`,
  `GET /v1/videos/00000000-0000-0000-0000-000000000001/playback`.

## Results

| Endpoint | rps | err | p50 | p95 | p99 | max |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| feed/home | 1822.7 | 0 | 23.8 ms | 49.2 ms | 84.3 ms | 90.2 ms |
| playback | 1039.0 | 0 | 41.2 ms | 103.7 ms | 158.6 ms | 169.2 ms |

Conditions: 50-way concurrency, loopback, single-API process, in-memory
seed store. The playback run distributed `X-Forwarded-For` across 250
client IPs (see Rate limiting); without distribution the global limiter
429s the burst by design.

## Rate limiting (verified live)

- Global sliding window: 1200 req/min per client IP
  (`services/api/cmd/server/main.go` → `security.NewRateLimiter`).
- A single-IP 1000-request burst tripped `429 RATE_LIMIT_EXCEEDED` with a
  `Retry-After` header; no 5xx, no error-log spam. Fail-closed as designed.

## Recovery (verified live)

- `Stop-Process` on the API (healthz → connection refused), then restart
  the same binary: **healthz 200 again after 1760 ms** (stopwatch from
  process spawn to first HTTP 200, 200 ms poll interval).
- Stateless API: no warmup, no cache rebuild, no data loss (seed store
  rehydrates from code; production PostgreSQL path is covered by
  `RUNBOOK_AND_DR.md` §2 PITR).

## Declared SLOs (operator-PC stack)

| Signal | Objective | Basis |
| :--- | :--- | :--- |
| Read p95 (feed/playback, loopback) | < 250 ms | measured p95 49–104 ms, 2.4x headroom |
| Read availability (unlimited client) | 99.9% / 5 min | 2000/2000 OK under 50-way burst |
| Process-crash RTO | < 30 s | measured 1.76 s |
| Rate-limit behavior | 429 + Retry-After, zero 5xx | observed live |

Production K8s/RDS targets remain in `RUNBOOK_AND_DR.md` §1; the table
above is the measured floor for the currently running stack.
