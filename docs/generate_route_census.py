#!/usr/bin/env python3
"""Generate the API route census and synchronize the OpenAPI specification.

Inputs (all committed, reviewable):
  1. ROUTES below: chi.Walk dump of services/api (2026-11-16, 143 routes).
     Refresh: temporarily add a chi.Walk test in services/api/cmd/server,
     run `go test ./cmd/server/ -run TestZZDumpRoutes -v`, paste, delete.
  2. tests/openapi/gvp-v1.yaml: 23 hand-detailed paths + components.
     These are preserved verbatim as the detailed core of the spec.

Outputs (deterministic, do not hand-edit):
  - docs/api/ROUTE_CENSUS.md : every route with its auth guard + hole flags
  - docs/api/routes.json     : machine-readable census for contract tests
  - openapi/gvp-v1.yaml      : detailed core + generated inventory for the rest

Usage (repo root):
    python docs/generate_route_census.py
"""
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
FIXTURE = ROOT / "tests" / "openapi" / "gvp-v1.yaml"
CENSUS_MD = ROOT / "docs" / "api" / "ROUTE_CENSUS.md"
ROUTES_JSON = ROOT / "docs" / "api" / "routes.json"
OPENAPI = ROOT / "openapi" / "gvp-v1.yaml"

# ---------------------------------------------------------------------------
# 1. Verified route dump (chi.Walk, 2026-11-16). `/*/` is chi's mount artifact
#    for the users router mounted at "/": the real paths are `/v1/me...`.
# ---------------------------------------------------------------------------
ROUTES_RAW = """
GET /healthz
GET /metrics
GET /ready
GET /v1/admin/ads/delivery-diagnostics
POST /v1/admin/finance/batches/approve
GET /v1/admin/finance/settlements
GET /v1/admin/finance/summary
GET /v1/admin/moderation
POST /v1/admin/moderation/{id}/resolve
GET /v1/admin/policies/
PATCH /v1/admin/policies/
PATCH /v1/admin/policies/flags/{key}
GET /v1/admin/audit-logs
GET /v1/admin/dashboard
GET /v1/admin/users
PATCH /v1/admin/users/{id}/status
GET /v1/ads/admin/creatives
POST /v1/ads/admin/creatives/{id}/approve
POST /v1/ads/admin/creatives/{id}/reject
GET /v1/ads/admin/diagnostics
GET /v1/ads/advertiser/reports
GET /v1/ads/campaigns
POST /v1/ads/campaigns
GET /v1/ads/campaigns/{id}
PATCH /v1/ads/campaigns/{id}/status
POST /v1/ads/decision
POST /v1/ads/events
POST /v1/ads/preview
GET /v1/ads/reports
GET /v1/advertiser/reports
GET /v1/advertiser/admin/creatives
POST /v1/advertiser/admin/creatives/{id}/approve
POST /v1/advertiser/admin/creatives/{id}/reject
GET /v1/advertiser/admin/diagnostics
GET /v1/advertiser/advertiser/reports
GET /v1/advertiser/campaigns
POST /v1/advertiser/campaigns
GET /v1/advertiser/campaigns/{id}
PATCH /v1/advertiser/campaigns/{id}/status
POST /v1/advertiser/decision
POST /v1/advertiser/events
POST /v1/advertiser/preview
GET /v1/advertiser/reports
POST /v1/channels/
GET /v1/channels/{id}
PATCH /v1/channels/{id}
DELETE /v1/channels/{id}/subscribe
POST /v1/channels/{id}/subscribe
GET /v1/channels/{id}/videos
DELETE /v1/comments/{id}
POST /v1/comments/{id}/heart
POST /v1/comments/{id}/like
POST /v1/comments/{id}/pin
POST /v1/events/analytics
POST /v1/feed/feedback
GET /v1/feed/home
GET /v1/feed/shorts
GET /v1/feed/subscriptions
POST /v1/live-streams/
GET /v1/live-streams/{id}
GET /v1/live-streams/{id}/chat
POST /v1/live-streams/{id}/chat
GET /v1/live-streams/{id}/health
POST /v1/live-streams/{id}/polls
GET /v1/live-streams/{id}/polls/active
POST /v1/live-streams/{id}/polls/vote
POST /v1/live-streams/{id}/rotate-key
POST /v1/live-streams/{id}/start
POST /v1/live-streams/{id}/stop
POST /v1/live-streams/{id}/superchat
GET /v1/live-streams/{id}/superchats
GET /v1/live-streams/{id}/ws
POST /v1/moderation/reports
POST /v1/partners/campaigns
GET /v1/partners/campaigns
GET /v1/partners/rights/claims
POST /v1/partners/rights/claims
POST /v1/partners/rights/claims/{id}/dispute
POST /v1/partners/rights/claims/{id}/resolve
POST /v1/partners/rights/dmca-counter-notice
POST /v1/partners/rights/dmca-takedown
GET /v1/partners/settlements
GET /v1/partners/settlements/earnings-summary
POST /v1/partners/settlements/payout
GET /v1/partners/channels
GET /v1/partners/contracts
PATCH /v1/partners/contracts/{id}
GET /v1/partners/me
POST /v1/payments/checkout
POST /v1/payments/confirm
GET /v1/payments/transactions
POST /v1/payments/webhooks/{provider}
POST /v1/playlists/
PATCH /v1/playlists/{id}
GET /v1/playlists/{id}
POST /v1/playlists/{id}/items
GET /v1/search
GET /v1/search/suggestions
GET /v1/studio/earnings
GET /v1/studio/settlements
GET /v1/studio/analytics
GET /v1/studio/claims
GET /v1/studio/comments
GET /v1/studio/dashboard
GET /v1/studio/videos
POST /v1/studio/videos/{id}/edit-jobs
GET /v1/studio/videos/{id}/edit-jobs/{jobId}
GET /v1/subscriptions/
POST /v1/subscriptions/{channelId}
DELETE /v1/subscriptions/{channelId}
POST /v1/uploads/
PUT /v1/uploads/{id}/chunks/{partNumber}
POST /v1/uploads/{id}/complete
POST /v1/uploads/{id}/parts
GET /v1/uploads/{id}/status
POST /v1/videos/{id}/comments
GET /v1/videos/{id}/comments
GET /v1/videos/{id}
PATCH /v1/videos/{id}
DELETE /v1/videos/{id}
GET /v1/videos/{id}/captions
POST /v1/videos/{id}/captions
DELETE /v1/videos/{id}/captions/{capId}
POST /v1/videos/{id}/like
DELETE /v1/videos/{id}/like
GET /v1/videos/{id}/playback
GET /v1/*/me
PATCH /v1/*/me
GET /v1/*/me/blocked-users
POST /v1/*/me/blocked-users
DELETE /v1/*/me/blocked-users/{targetId}
POST /v1/*/me/delete-request
DELETE /v1/*/me/delete-request
GET /v1/*/me/history
DELETE /v1/*/me/history
GET /v1/*/me/mfa
POST /v1/*/me/mfa/disable
POST /v1/*/me/mfa/setup
POST /v1/*/me/mfa/verify
GET /v1/*/me/sessions
DELETE /v1/*/me/sessions/{id}
POST /v1/*/me/sessions/{id}/revoke
GET /v1/*/me/strikes
"""

# Routes that intentionally serve unauthenticated traffic (audited 2026-11-16).
PUBLIC = {
    ("GET", "/healthz"), ("GET", "/ready"), ("GET", "/metrics"),
    ("POST", "/v1/ads/decision"), ("POST", "/v1/ads/events"), ("POST", "/v1/ads/preview"),
    ("POST", "/v1/advertiser/decision"), ("POST", "/v1/advertiser/events"), ("POST", "/v1/advertiser/preview"),
    ("POST", "/v1/events/analytics"),
    ("GET", "/v1/feed/home"), ("GET", "/v1/feed/shorts"),
    ("GET", "/v1/feed/subscriptions"), ("POST", "/v1/feed/feedback"),
    ("GET", "/v1/search"), ("GET", "/v1/search/suggestions"),
    ("GET", "/v1/channels/{id}"), ("GET", "/v1/channels/{id}/videos"),
    ("GET", "/v1/videos/{id}"), ("GET", "/v1/videos/{id}/playback"),
    ("GET", "/v1/videos/{id}/comments"),
    ("GET", "/v1/live-streams/{id}"), ("GET", "/v1/live-streams/{id}/health"),
    ("GET", "/v1/live-streams/{id}/superchats"), ("GET", "/v1/live-streams/{id}/chat"),
    ("GET", "/v1/live-streams/{id}/ws"), ("GET", "/v1/live-streams/{id}/polls/active"),
    ("GET", "/v1/playlists/{id}"),
    ("POST", "/v1/payments/webhooks/{provider}"),
}

# (method, path-prefix) -> role guard, in match order (audited 2026-11-16).
ROLE_RULES = [
    ("/v1/admin/users", "super_admin, platform_admin (+support read on GET /users)"),
    ("/v1/admin/moderation", "super_admin, platform_admin, moderator, support_agent"),
    ("/v1/admin/finance", "super_admin, platform_admin, finance_admin"),
    ("/v1/admin/policies", "super_admin, platform_admin (+support read on GET)"),
    ("/v1/admin/ads/delivery-diagnostics", "super_admin, platform_admin"),
    ("/v1/admin/dashboard", "super_admin, platform_admin, finance_admin, moderator, support_agent"),
    ("/v1/admin/audit-logs", "super_admin, platform_admin, finance_admin, moderator, support_agent"),
    ("/v1/partners/rights/claims/{id}/resolve", "super_admin, platform_admin, moderator"),
]

# Known auth holes: public but must be guarded (P0 backlog, IMPLEMENTATION_STATUS).
HOLES = {
    "/v1/ads/campaigns", "/v1/ads/campaigns/{id}", "/v1/ads/campaigns/{id}/status",
    "/v1/ads/reports", "/v1/ads/advertiser/reports",
    "/v1/ads/admin/creatives", "/v1/ads/admin/creatives/{id}/approve",
    "/v1/ads/admin/creatives/{id}/reject", "/v1/ads/admin/diagnostics",
    "/v1/advertiser/campaigns", "/v1/advertiser/campaigns/{id}", "/v1/advertiser/campaigns/{id}/status",
    "/v1/advertiser/reports", "/v1/advertiser/advertiser/reports",
    "/v1/advertiser/admin/creatives", "/v1/advertiser/admin/creatives/{id}/approve",
    "/v1/advertiser/admin/creatives/{id}/reject", "/v1/advertiser/admin/diagnostics",
}

GROUP_ORDER = ["ops", "me", "channels", "videos", "comments", "uploads", "feed",
               "search", "live", "studio", "partners", "ads", "advertiser",
               "admin", "payments", "playlists", "subscriptions", "moderation", "events"]


def normalize(path: str) -> str:
    return path.replace("/v1/*/", "/v1/")


def auth_of(method: str, path: str) -> str:
    if (method, path) in PUBLIC:
        return "Public"
    for prefix, roles in ROLE_RULES:
        if path == prefix or path.startswith(prefix + "/") or path.startswith(prefix + "{"):
            return f"Role({roles})"
    if path.startswith("/v1/admin/"):
        return "Role(super_admin, platform_admin)"
    if path in HOLES:
        return "Public HOLE: guard required (P0)"
    return "Auth"


def group_of(path: str) -> str:
    if not path.startswith("/v1/"):
        return "ops"
    seg = path[len("/v1/"):].split("/")[0]
    return {"live-streams": "live"}.get(seg, seg)


VERB = {"GET": "Get", "POST": "Create", "PUT": "Put", "PATCH": "Update", "DELETE": "Delete"}


def summarize(method: str, path: str) -> str:
    if path in ("/healthz", "/ready", "/metrics"):
        return {"GET /healthz": "Liveness probe", "GET /ready": "Readiness probe",
                "GET /metrics": "Prometheus metrics"}[f"{method} {path}"]
    parts = [p for p in path.split("/") if p not in ("", "v1")]
    words: list[str] = []
    for part in parts:
        if part.startswith("{"):
            continue
        words.append(part.replace("-", " "))
    noun = " ".join(words).strip() or "resource"
    verb = {"GET": "Get", "POST": "Create", "PUT": "Upload",
            "PATCH": "Update", "DELETE": "Delete"}[method]
    text = f"{verb} {noun}".strip()
    if len(text) > 80:
        text = text[:77] + "..."
    return text


def operation_id(method: str, path: str) -> str:
    parts = []
    for part in path.split("/"):
        if not part or part in ("v1",):
            continue
        if part.startswith("{") and part.endswith("}"):
            inner = part[1:-1].replace("-", "_")
            camel = "".join(seg[:1].upper() + seg[1:] for seg in inner.split("_"))
            parts.append("By" + camel)
        else:
            camel = "".join(seg[:1].upper() + seg[1:] for seg in re.split(r"[-_]", part))
            parts.append(camel)
    return VERB[method] + "".join(parts)


def parse_routes():
    routes = []
    for line in ROUTES_RAW.strip().splitlines():
        method, raw = line.split(" ", 1)
        routes.append((method, normalize(raw.strip()), raw.strip()))
    return routes


def build_census(routes):
    rows = []
    seen = set()
    for method, path, raw in routes:
        dup = (method, path) in seen
        seen.add((method, path))
        note = ""
        if dup:
            note = "DUPLICATE registration (mount + explicit)"
        if path in HOLES:
            note = (note + "; " if note else "") + "P0: no auth guard"
        if raw != path:
            note = (note + "; " if note else "") + f"chi renders `{raw}`"
        if path.endswith("/") and len(path) > 1:
            note = (note + "; " if note else "") + "trailing-slash route"
        rows.append({"method": method, "path": path, "auth": auth_of(method, path),
                     "group": group_of(path), "note": note})
    return rows


def write_census_md(rows):
    by_group: dict[str, list] = {}
    for r in rows:
        by_group.setdefault(r["group"], []).append(r)
    out = ["# API Route Census (generated, chi.Walk verified)",
           "",
           "Source: `chi.Walk` dump of `services/api/cmd/server` (2026-11-16, "
           f"{len(rows)} registrations). Generator: `python docs/generate_route_census.py` — do not hand-edit.",
           "",
           "Auth legend: **Public** serves visitors · **Auth** needs a session · **Role(...)** needs the role ·",
           "**HOLE** is public but must be guarded (P0). `/v1/*/` in chi output is the users-router mount artifact",
           "for real paths `/v1/me...`.",
           ""]
    order = [g for g in GROUP_ORDER if g in by_group] + sorted(set(by_group) - set(GROUP_ORDER))
    for group in order:
        out += [f"## {group} ({len(by_group[group])})", "",
                "| Method | Path | Auth | Notes |", "| :--- | :--- | :--- | :--- |"]
        for r in by_group[group]:
            out.append(f"| `{r['method']}` | `{r['path']}` | {r['auth']} | {r['note']} |")
        out.append("")
    CENSUS_MD.write_text("\n".join(out), encoding="utf-8")
    print(f"census: {len(rows)} routes -> {CENSUS_MD.relative_to(ROOT)}")


def write_routes_json(rows):
    data = [{"method": r["method"], "path": r["path"], "auth": r["auth"]} for r in rows]
    ROUTES_JSON.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
    print(f"json: {len(data)} routes -> {ROUTES_JSON.relative_to(ROOT)}")


def split_fixture():
    """Split the 23-path fixture into (head, paths_block, components_block)."""
    lines = FIXTURE.read_text(encoding="utf-8").splitlines(keepends=True)
    paths_idx = next(i for i, l in enumerate(lines) if l.rstrip("\n") == "paths:")
    comp_idx = next(i for i, l in enumerate(lines) if l.rstrip("\n") == "components:")
    return "".join(lines[:paths_idx + 1]), "".join(lines[paths_idx + 1:comp_idx]), "".join(lines[comp_idx:])


def existing_paths(paths_block: str) -> set[str]:
    return set(re.findall(r"^  (/[^\s:]*):", paths_block, flags=re.M))


def openapi_operation(method: str, full_path: str, auth: str) -> list[str]:
    """Emit one generated operation (method block, without the path key)."""
    prefix = "/v1/"
    short = "/" + full_path[len(prefix):] if full_path.startswith(prefix) else full_path
    params = re.findall(r"\{(\w+)\}", short)
    tag = short.strip("/").split("/")[0] or "ops"
    op_id = operation_id(method, full_path)
    summary = summarize(method, full_path)
    status = {"GET": "200", "POST": "201", "PUT": "200", "PATCH": "200", "DELETE": "200"}[method]
    lines = [f"    {method.lower()}:",
             f"      tags: [{tag}]", f'      summary: "{summary}"',
             f"      operationId: {op_id}",
             "      x-census: {generated: true, detail: pending}",
             f'      x-auth: "{auth}"']
    if params:
        lines.append("      parameters:")
        for p in params:
            lines += [f"        - name: {p}", "          in: path", "          required: true",
                      "          schema:", "            type: string"]
    if method in ("POST", "PUT", "PATCH"):
        lines += ["      requestBody:", "        required: true",
                  "        content:", "          application/json:", "            schema:",
                  "              type: object"]
    lines += ["      responses:", f"        '{status}':", "          description: Success"]
    if auth.startswith("Role") or auth == "Auth":
        lines += ["        '401':", "          $ref: '#/components/responses/Unauthorized'"]
    if auth.startswith("Role"):
        lines += ["        '403':", "          $ref: '#/components/responses/Forbidden'"]
    if params or method in ("GET", "PATCH", "DELETE"):
        lines += ["        '404':", "          $ref: '#/components/responses/NotFound'"]
    return lines


def write_openapi(rows):
    head, paths_block, components_block = split_fixture()
    have = existing_paths(paths_block)
    # Normalize fixture servers/info header for the canonical spec.
    head = head.replace("title: StreamVerse (Global Video Platform) API",
                        "title: Global Video Platform API")
    generated: list[str] = []
    emitted: set[tuple[str, str]] = set()
    by_path: dict[str, list] = {}
    covered = 0
    for r in rows:
        method, full = r["method"], r["path"]
        if not full.startswith("/v1/"):
            continue  # ops probes stay out of the /v1 contract
        short = "/" + full[len("/v1/"):]
        if short in have:
            covered += 1
            continue
        if (method, short) in emitted:
            continue  # duplicate chi registration; census flags it, spec keeps one
        emitted.add((method, short))
        by_path.setdefault(short, []).append(r)
    for short, ops in by_path.items():
        generated.append(f"  {short}:")
        for r in ops:
            generated.extend(openapi_operation(r["method"], r["path"], r["auth"]))
    out = [head.rstrip("\n"),
           "  # --- hand-detailed core (from tests/openapi/gvp-v1.yaml) ---",
           paths_block.rstrip("\n"),
           "  # --- generated inventory (x-census.generated) ---",
           "\n".join(generated),
           components_block.rstrip("\n") + "\n"]
    OPENAPI.write_text("\n".join(out), encoding="utf-8")
    print(f"openapi: {covered} detailed + {len(generated)} lines generated -> {OPENAPI.relative_to(ROOT)}")


def main():
    routes = parse_routes()
    print(f"dump: {len(routes)} registrations")
    rows = build_census(routes)
    write_census_md(rows)
    write_routes_json(rows)
    write_openapi(rows)


if __name__ == "__main__":
    main()
