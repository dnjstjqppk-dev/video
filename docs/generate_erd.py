#!/usr/bin/env python3
"""Regenerate per-phase ERD diagrams from the authoritative migrations.

Source of truth: db/migrations/00001..00014 (*.sql, goose format, Up section only).
Output: docs/architecture/erd/00001_*.md .. 00014_*.md (Mermaid erDiagram).

Usage (repo root):
    python docs/generate_erd.py

The script is deterministic: same migrations in -> byte-identical diagrams out.
It parses CREATE TABLE blocks, column names, and REFERENCES targets
(both inline `col ... REFERENCES t` and table-level FOREIGN KEY clauses).
"""
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
MIG_DIR = ROOT / "db" / "migrations"
OUT_DIR = ROOT / "docs" / "architecture" / "erd"

PHASE_TITLES = {
    "00001_create_accounts": "Identity, RBAC, Sessions, Organizations",
    "00002_create_content": "Catalog, Channels, Videos, Media Ingest",
    "00003_create_social": "Engagement: Reactions, Comments, History, Community",
    "00004_create_live": "Live Streams, Inputs, Chat, Recordings",
    "00005_create_ads": "Base Advertising: Campaigns, Creatives, Placements",
    "00006_create_payments_and_ledger": "Monetization, Payments, Double-Entry Ledger",
    "00007_create_copyright_and_moderation": "Rights, Copyright Claims, Moderation",
    "00008_create_operations_and_outbox": "Admin Ops, Audit, Feature Flags, Outbox",
    "00009_create_production_ad_engine": "Production Ad Engine: Slots, Flights, Delivery",
    "00010_forward_roles_and_payments": "P0-1 Forward Fill: Roles + Payment Columns",
    "00011_media_job_source": "Media Jobs: Source Reference",
    "00012_vravra_studio_completion": "Studio Completion: Hourly Views + Video Columns",
    "00013_video_drm": "Video DRM Protection Metadata",
    "00014_ai_originals": "AI Originals: Series + AI Review Columns",
}

TABLE_RE = re.compile(
    r"CREATE\s+TABLE\s+IF\s+NOT\s+EXISTS\s+(\w+)\s*\((.*?)\);",
    re.IGNORECASE | re.DOTALL,
)
# NOTE: [^,]*? must cross parens (VARCHAR(64) ...) so group(1) stays the
# column name; per-line matching keeps it from crossing into other columns.
INLINE_REF_RE = re.compile(r"(\w+)[^,]*?REFERENCES\s+(\w+)", re.IGNORECASE)
ALTER_RE = re.compile(
    r"ALTER\s+TABLE\s+(?:IF\s+EXISTS\s+)?(\w+)\s+ADD\s+COLUMN\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)",
    re.IGNORECASE,
)
DOWN_MARKER = "-- +goose Down"
TABLE_FK_RE = re.compile(
    r"FOREIGN\s+KEY\s*\((\w+)\)\s*REFERENCES\s+(\w+)", re.IGNORECASE
)
CLAUSE_START = ("CONSTRAINT", "PRIMARY", "FOREIGN", "UNIQUE", "CHECK", "KEY", "EXCLUDE")


def parse_migration(path: Path):
    """Return (tables, relations, alters) for one migration file.

    tables: {table_name: [column, ...]} in file order.
    relations: [(from_table, from_col, to_table), ...] deduplicated, in order.
    alters: {table_name: [added column, ...]} from ALTER TABLE ... ADD COLUMN.
    Only the Up section is parsed; Down is operational rollback, not schema.
    """
    sql = path.read_text(encoding="utf-8")
    if DOWN_MARKER in sql:
        sql = sql.split(DOWN_MARKER, 1)[0]
    alters: dict[str, list[str]] = {}
    for m in ALTER_RE.finditer(sql):
        alters.setdefault(m.group(1), []).append(m.group(2))
    tables: dict[str, list[str]] = {}
    relations: list[tuple[str, str, str]] = []
    seen_rel: set[tuple[str, str, str]] = set()

    for match in TABLE_RE.finditer(sql):
        table, body = match.group(1), match.group(2)
        cols: list[str] = []
        for raw_line in body.splitlines():
            line = raw_line.strip().rstrip(",").strip()
            if not line or line.startswith("--"):
                continue
            up = line.upper()
            if up.startswith(CLAUSE_START):
                for fk in TABLE_FK_RE.finditer(line):
                    rel = (table, fk.group(1), fk.group(2))
                    if rel not in seen_rel:
                        seen_rel.add(rel)
                        relations.append(rel)
                continue
            col = line.split()[0].strip('"')
            if re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", col):
                cols.append(col)
            for fk in INLINE_REF_RE.finditer(line):
                # Guard: the "column" captured must be this line's column.
                if fk.group(1) == col or fk.group(1).upper() in (
                    "KEY", "REFERENCES",
                ):
                    rel_col = col
                else:
                    rel_col = fk.group(1)
                rel = (table, rel_col, fk.group(2))
                if rel not in seen_rel:
                    seen_rel.add(rel)
                    relations.append(rel)
        tables[table] = cols
    return tables, relations, alters


def mermaid_entity(name: str, cols: list[str]) -> str:
    lines = [f"  {name.upper()} {{"]
    for col in cols:
        lines.append(f"    string {col}")
    lines.append("  }")
    return "\n".join(lines)


def render_phase(stem: str, sql_name: str, tables, relations, alters) -> str:
    title = PHASE_TITLES.get(stem, stem)
    out = [
        f"# {stem} — {title}",
        "",
        f"Source: `db/migrations/{sql_name}`. Tables: **{len(tables)}**. "
        f"Foreign keys: **{len(relations)}**.",
        "",
        "Regenerated by `python docs/generate_erd.py`. Do not hand-edit.",
        "",
    ]
    if alters:
        out += ["## Alters (added columns)", ""]
        for table, cols in alters.items():
            out.append(f"- `{table}`: {', '.join(f'`{c}`' for c in cols)}")
        out.append("")
    if not tables:
        out += ["_No new tables: data fill and/or column alters only._", ""]
        return "\n".join(out)
    out += ["## Tables", ""]
    for table, cols in tables.items():
        out.append(f"- `{table}` ({len(cols)} columns)")
    out += ["", "## Diagram", "", "```mermaid", "erDiagram"]
    for table, cols in tables.items():
        out.append(mermaid_entity(table, cols))
    for frm, col, to in relations:
        out.append(f'  {frm.upper()} ||--o{{ {to.upper()} : "{col}"')
    out += ["```", ""]
    return "\n".join(out)


def main() -> int:
    files = sorted(MIG_DIR.glob("00*.sql"))
    if len(files) != 14:
        print(f"expected 14 migrations, found {len(files)}", file=sys.stderr)
        return 1
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    total_tables = 0
    total_fks = 0
    for path in files:
        stem = path.stem
        tables, relations, alters = parse_migration(path)
        if not tables and not alters:
            print(f"no schema parsed from {path.name}", file=sys.stderr)
            return 1
        total_tables += len(tables)
        total_fks += len(relations)
        (OUT_DIR / f"{stem}.md").write_text(
            render_phase(stem, path.name, tables, relations, alters), encoding="utf-8"
        )
        print(f"{path.name}: {len(tables)} tables, {len(relations)} FKs, "
              f"{sum(len(v) for v in alters.values())} added cols")
    print(f"total: {total_tables} tables, {total_fks} FKs")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
