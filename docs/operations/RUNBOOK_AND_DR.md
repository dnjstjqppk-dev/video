# vravra — Operations Runbook & Disaster Recovery (DR)

> **Current-hosting note (2026-11-20, code-audited):** §2–§4 target the **production K8s/RDS topology** and are
> NOT executable on the live stack. The actually running stack is operator-PC hosted (`gvp-api.exe :18092`,
> media worker `:8081`) behind a Cloudflare quick tunnel + `apps/api-proxy` worker — see `infra/local/README.md` and
> [docs/IMPLEMENTATION_STATUS.md](../IMPLEMENTATION_STATUS.md) §3. PITR/DLQ are unexercised (P1-20); outbox relay and
> NATS DLQ do not exist in code. Use §0 below for the live stack.

## 0. Live-stack procedures (operator PC, quick tunnel)

| Symptom | Action |
| :--- | :--- |
| Console `/api` dead | Check tunnel URL rotation: read the fresh `*.trycloudflare.com` URL and update the `API_UPSTREAM` secret, redeploy api-proxy. |
| API down (`:18092`) | Restart `gvp-api.exe` with the `infra/local` env (`GVP_ENV=test`, dev tokens, R2 → `:9002`, worker → `:8081`). In-memory videos are lost; PG-hydrated rows return. |
| Transcodes stuck QUEUED | Restart the media worker (`:8081`) with `FFMPEG_BIN`/`FFPROBE_BIN`/`PACKAGER_BIN`; re-submit. No DLQ exists — stuck jobs need manual re-queue. |
| S3 errors (`:9002`) | Restart `weed.exe server` (master 19333, volume 18083, s3 9002). If raft wedges after unclean shutdown, wipe `tools/data/seaweed` and restart (probe data only). |
| Full PC reboot | Start order: SeaweedFS → API → media worker → tunnel → verify `/healthz` + one `/v1/feed/home` through the proxy. Auto-start on boot is NOT registered. |

## 1. Recovery Objectives (RTO & RPO)

| System Tier | Components | Target RPO (Max Data Loss) | Target RTO (Max Recovery Time) | HA / Failover Mechanism |
| :--- | :--- | :--- | :--- | :--- |
| **Tier 0: Identity & Core Ledger** | PostgreSQL 18 Primary (`users`, `ledger_entries`, `settlements`, `outbox_events`) | **< 1 minute** (WAL archiving + synchronous Multi-AZ standby) | **< 15 minutes** | Automated Multi-AZ failover + PITR WAL replay |
| **Tier 1: Core Playback & API** | Go API (`gvp-api`), Valkey Cluster, Cloudflare R2 HLS (`media.brand.com`) | **0 seconds** (Stateless API + R2 11-nines durability) | **< 5 minutes** | Multi-replica K8s HPA + Argo CD instant rollback |
| **Tier 2: Async Media & Search** | NATS JetStream, `media-worker`, OpenSearch, ClickHouse | **< 5 minutes** (Reconstructible from PostgreSQL Outbox & R2 Raw) | **< 30 minutes** | Outbox replay + OpenSearch reindex job |

---

## 2. PostgreSQL 18 Point-In-Time Recovery (PITR) Runbook

### 2.1 Continuous Backup Architecture
- **Daily Base Snapshots**: Automated daily snapshots at `18:00 UTC` retained for **35 days** (`infra/terraform/main.tf`).
- **Continuous WAL Archiving**: Write-Ahead Logs (WAL) are streamed continuously to encrypted object storage every 60 seconds, enabling second-level Point-In-Time Recovery.

### 2.2 Step-by-Step PITR Restore Procedure
1. **Freeze Write Traffic**:
   - Put `gvp-api` into read-only maintenance mode or scale write workers (`gvp-scheduler`) to `0`:
     ```bash
     kubectl scale deployment/gvp-scheduler -n gvp-prod --replicas=0
     ```
2. **Provision Restored PostgreSQL 18 Instance to Target Timestamp**:
     ```bash
     aws rds restore-db-instance-to-point-in-time \
       --source-db-instance-identifier gvp-production-pg18-primary \
       --target-db-instance-identifier gvp-production-pg18-pitr-recovery \
       --restore-time "2026-10-24T07:55:00Z" \
       --db-instance-class db.r7g.2xlarge \
       --multi-az
     ```
3. **Verify Schema & Ledger Integrity on Restored Instance**:
   - Connect to the recovery endpoint and verify migration state and double-entry ledger balance:
     ```sql
     SELECT COUNT(*) FROM videos;
     SELECT SUM(debit_cents) - SUM(credit_cents) AS ledger_imbalance FROM ledger_entries;
     -- Must return 0
     ```
4. **Promote & Cut Over `DATABASE_URL` Secret**:
   - Update `gvp-backend-secrets` with the new primary endpoint and perform a rolling restart:
     ```bash
     kubectl rollout restart deployment/gvp-api deployment/gvp-scheduler -n gvp-prod
     kubectl rollout status deployment/gvp-api -n gvp-prod
     ```

---

## 3. Worker Dead-Letter Queue (DLQ) & Transactional Outbox Recovery

### Scenario A: `GvpMediaTranscodingQueueLagHigh` or Failed FFmpeg Jobs
1. **Inspect NATS JetStream DLQ (`MEDIA_JOBS_DLQ`)**:
   - Jobs failing 5 consecutive delivery attempts (e.g., corrupted container header or OOM on malformed 8K input) are routed to `MEDIA_JOBS_DLQ`.
2. **Triage Error Reason**:
   - If the error is `INVALID_CODEC_OR_CORRUPT_CONTAINER`, mark the video `FAILED` with user-facing diagnostic message in vravra studio.
   - If the error was transient (Spot node preemption during FFmpeg pass), replay messages from `MEDIA_JOBS_DLQ` back to `MEDIA_JOBS`.

### Scenario B: `GvpOutboxRelayStalled` (`outbox_events` Backlog > 500)
1. **Verify NATS JetStream Health**:
   ```bash
   kubectl exec -n gvp-prod deploy/gvp-api -- wget -qO- http://nats:8222/healthz
   ```
2. **Scale Outbox Relay Worker**:
   - Because `scheduler` uses `SELECT ... FOR UPDATE SKIP LOCKED`, multiple replicas can safely drain the backlog concurrently without duplicate publication:
   ```bash
   kubectl scale deployment/gvp-scheduler -n gvp-prod --replicas=3
   ```

---

## 4. Incident Severity & Escalation Matrix

| Severity | Trigger Condition | Initial Response SLA | Automated Action |
| :--- | :--- | :--- | :--- |
| **SEV-1 (Critical)** | API 5xx > 1.0%, Playback CDN outage, or Auth/Payment failure | **< 5 minutes** (PagerDuty On-Call SRE) | Argo CD auto-rollback if triggered within 15m of deployment |
| **SEV-2 (High)** | API p95 > 300ms for 5m, Live ingest failure, or Outbox lag > 500 | **< 15 minutes** | Kubernetes HPA scale-out up to 24 API pods |
| **SEV-3 (Moderate)** | VOD transcoding queue lag > 50 jobs or search indexing delay | **< 1 hour** | Spot media-worker pool autoscaling (`maxReplicas: 32`) |
