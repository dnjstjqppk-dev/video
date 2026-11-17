#!/usr/bin/env python3
"""Regenerate the canonical schema snapshot from the authoritative migrations.

Source of truth: db/migrations/00001..00010 (*.sql, goose format).
Outputs (deterministic, do not hand-edit):
  - db/schema/schema.sql    : concatenated Up sections with per-file banners
  - db/schema/inventory.csv : phase,table,column,type,primary_key rows

Usage (repo root):
    python docs/generate_schema.py

The inventory parses CREATE TABLE blocks (column names, types, inline and
table-level PRIMARY KEY) plus ALTER TABLE ... ADD COLUMN clauses, so forward
migrations like 00010 are reflected too.
"""
import csv
import io
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
MIG_DIR = ROOT / "db" / "migrations"
SCHEMA_SQL = ROOT / "db" / "schema" / "schema.sql"
INVENTORY_CSV = ROOT / "db" / "schema" / "inventory.csv"

TABLE_RE = re.compile(
    r"CREATE\s+TABLE\s+IF\s+NOT\s+EXISTS\s+(\w+)\s*\((.*?)\);",
    re.IGNORECASE | re.DOTALL,
)
ALTER_RE = re.compile(
    r"ALTER\s+TABLE\s+(\w+)\s+ADD\s+COLUMN\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)\s+([A-Za-z0-9_\[\]() ]+?)(?:,|;|\s+(?:NOT\s+NULL|NULL|DEFAULT|REFERENCES|UNIQUE|CHECK|CONSTRAINT)\b)",
    re.IGNORECASE,
)
TABLE_PK_RE = re.compile(r"PRIMARY\s+KEY\s*\(([^)]+)\)", re.IGNORECASE)
CLAUSE_START = ("CONSTRAINT", "PRIMARY", "FOREIGN", "UNIQUE", "CHECK", "KEY", "EXCLUDE")
TYPE_WORDS = {
    "UUID", "VARCHAR", "CHAR", "CHARACTER", "TEXT", "BOOLEAN", "BOOL",
    "TIMESTAMPTZ", "TIMESTAMP", "DATE", "TIME", "INTERVAL",
    "BIGINT", "INT", "INTEGER", "SMALLINT", "NUMERIC", "DECIMAL", "REAL",
    "DOUBLE", "PRECISION", "INET", "JSONB", "JSON", "BYTEA", "SERIAL",
}
TYPE_STOP = {
    "NOT", "NULL", "DEFAULT", "REFERENCES", "PRIMARY", "UNIQUE", "CHECK",
    "CONSTRAINT", "COLLATE",
}


def up_section(text: str) -> str:
    """Return the goose Up part (everything before '-- +goose Down')."""
    head, _, _ = text.partition("-- +goose Down")
    return head.strip() + "\n"


def split_columns(body: str) -> list[str]:
    """Split a CREATE TABLE body on top-level commas (parens-aware).

    Line comments are stripped first: a comma inside `-- ...` must not
    split a column or glue a comment fragment onto the next column.
    """
    cleaned = "\n".join(line.split("--", 1)[0] for line in body.splitlines())
    parts, depth, current = [], 0, []
    for ch in cleaned:
        if ch == "(":
            depth += 1
        elif ch == ")":
            depth = max(0, depth - 1)
        if ch == "," and depth == 0:
            parts.append("".join(current))
            current = []
        else:
            current.append(ch)
    tail = "".join(current).strip()
    if tail:
        parts.append(tail)
    return parts


def parse_column(part: str):
    """Return (name, type, inline_pk) or None for table-level clauses."""
    # Strip trailing line comments.
    part = re.sub(r"--[^\n]*", "", part).strip()
    if not part:
        return None
    tokens = part.split()
    if len(tokens) < 2 or not re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*", tokens[0]):
        return None
    # Whole-word clause check, except when the second token is clearly a
    # column type (e.g. the `key VARCHAR(255)` column of idempotency_keys).
    if tokens[0].upper() in CLAUSE_START:
        second = re.split(r"\(", tokens[1], 1)[0].upper()
        if second not in TYPE_WORDS:
            return None
    name = tokens[0].strip('"')
    type_tokens = []
    for tok in tokens[1:]:
        if tok.upper().rstrip(",") in TYPE_STOP:
            break
        type_tokens.append(tok)
    col_type = " ".join(type_tokens).rstrip(",").strip() or "UNKNOWN"
    inline_pk = bool(re.search(r"PRIMARY\s+KEY", part, re.IGNORECASE))
    return name, col_type, inline_pk


def parse_migration(path: Path):
    """Return ([(table, [(col, type, pk)])], [(table, col, type)]) incl. ALTERs."""
    sql = up_section(path.read_text(encoding="utf-8"))
    tables: list[tuple[str, list[tuple[str, str, bool]]]] = []
    for match in TABLE_RE.finditer(sql):
        table, body = match.group(1), match.group(2)
        table_pks: set[str] = set()
        for pk_match in TABLE_PK_RE.finditer(body):
            for col in pk_match.group(1).split(","):
                table_pks.add(col.strip().strip('"'))
        cols = []
        for part in split_columns(body):
            parsed = parse_column(part)
            if parsed is None:
                continue
            name, col_type, inline_pk = parsed
            cols.append((name, col_type, inline_pk or name in table_pks))
        tables.append((table, cols))
    alters = []
    for match in ALTER_RE.finditer(sql):
        alters.append((match.group(1), match.group(2), match.group(3).strip()))
    return tables, alters


def main() -> int:
    files = sorted(MIG_DIR.glob("*.sql"))
    if not files:
        print("no migrations found", file=sys.stderr)
        return 1

    schema_parts = [
        "-- Canonical PostgreSQL schema snapshot (generated, do not hand-edit).",
        "-- Source of truth: db/migrations/*.sql (goose Up sections).",
        "-- Regenerate: python docs/generate_schema.py",
        "",
    ]
    rows = [("phase", "table", "column", "type", "primary_key")]
    total_tables = 0
    for path in files:
        stem = path.stem
        tables, alters = parse_migration(path)
        schema_parts.append(f"-- {'=' * 10} {path.name} (Up) {'=' * 10}")
        schema_parts.append(up_section(path.read_text(encoding="utf-8")))
        seen: set[tuple[str, str]] = set()
        for table, cols in tables:
            total_tables += 1
            for name, col_type, pk in cols:
                rows.append((stem, table, name, col_type, "True" if pk else "False"))
                seen.add((table, name))
        for table, col, col_type in alters:
            if (table, col) not in seen:
                rows.append((stem, table, col, col_type, "False"))
                seen.add((table, col))
        print(f"{path.name}: {len(tables)} tables, {len(alters)} added columns")

    SCHEMA_SQL.write_text("\n".join(schema_parts), encoding="utf-8")
    buf = io.StringIO()
    writer = csv.writer(buf, lineterminator="\n")
    writer.writerows(rows)
    INVENTORY_CSV.write_text(buf.getvalue(), encoding="utf-8")
    print(f"total: {total_tables} tables, {len(rows) - 1} columns -> db/schema/")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
