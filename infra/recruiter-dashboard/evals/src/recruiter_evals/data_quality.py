"""Data-quality report over the archive: aggregates only (#134, D-06).

Reads the archive/clean/ records (the #132 CleanRecord JSON written by the email-parser
Lambda and cmd/backfill) and the DynamoDB extraction rows, and writes a Markdown report
that contains **only counts, rates and distributions**: no email text, names, addresses,
companies, subjects or message ids.

The report is deterministic: the same records and rows always produce the same bytes.
It carries no timestamps; a fingerprint of the inputs identifies the data instead.

Usage (from infra/recruiter-dashboard/evals):

    uv run data-quality --clean-s3-bucket <bucket> --table <table>
    uv run data-quality --clean-dir <dir> --table <table>      # records from cmd/backfill --out-dir
"""

from __future__ import annotations

import argparse
import hashlib
import json
import math
import sys
from collections import Counter
from collections.abc import Callable, Iterable
from dataclasses import dataclass
from pathlib import Path

# Fields the extractor fills; "Unknown" is its placeholder for "not found".
EXTRACTION_FIELDS = ("first_name", "last_name", "recruiter_email", "company", "job_title", "phone")
UNKNOWN = "Unknown"

# The Lambda's dedup key for an extraction with an Unknown email and job title
# (models.GenerateDedupKey: first 16 bytes of SHA-256, hex). Every such row shares it,
# so it is reported separately from real duplicates.
PLACEHOLDER_DEDUP_KEY = hashlib.sha256(b"unknown|unknown").hexdigest()[:32]

# A body with fewer tokens than this (but not empty) counts as near-empty.
NEAR_EMPTY_TOKENS = 25

# Bodies with fewer letters than this are too short to classify a language.
MIN_LETTERS_FOR_LANGUAGE = 20

# Token-length buckets, aligned with week 3's truncation candidates (500 to 4,000).
TOKEN_BUCKETS = ((0, 0), (1, 24), (25, 249), (250, 499), (500, 999), (1000, 1999), (2000, 3999), (4000, None))

CONFIDENCE_BUCKETS = (
    ("0", lambda c: c == 0),
    ("(0, 0.5)", lambda c: 0 < c < 0.5),
    ("[0.5, 0.8)", lambda c: 0.5 <= c < 0.8),
    ("[0.8, 1]", lambda c: c >= 0.8),
)

TOKENIZER = "o200k_base"


@dataclass(frozen=True)
class Row:
    """The DynamoDB attributes the report needs (one extraction)."""

    id: str
    fields: dict[str, str]
    confidence: float
    dedup_key: str


# ---------------------------------------------------------------------------
# Inputs
# ---------------------------------------------------------------------------


def load_clean_records_from_dir(path: Path) -> list[dict]:
    """Loads every {id}.json record in a directory (cmd/backfill --out-dir output)."""
    return [json.loads(p.read_text(encoding="utf-8")) for p in sorted(path.glob("*.json"))]


def load_clean_records_from_s3(bucket: str, region: str, prefix: str = "archive/clean/") -> list[dict]:
    """Loads every record under archive/clean/ into memory (nothing is written to disk)."""
    import boto3

    s3 = boto3.client("s3", region_name=region)
    records = []
    for page in s3.get_paginator("list_objects_v2").paginate(Bucket=bucket, Prefix=prefix):
        for obj in page.get("Contents", []):
            if obj["Key"].endswith(".json"):
                body = s3.get_object(Bucket=bucket, Key=obj["Key"])["Body"].read()
                records.append(json.loads(body))
    return records


def load_dynamodb_rows(table: str, region: str) -> list[Row]:
    """Scans the extraction rows, skipping STATS# cache items. Only the attributes the
    report needs are projected; nothing is written to disk."""
    import boto3

    attrs = ("id", "confidence", "dedup_key", *EXTRACTION_FIELDS)
    names = {f"#a{i}": a for i, a in enumerate(attrs)}
    client = boto3.client("dynamodb", region_name=region)
    rows = []
    for page in client.get_paginator("scan").paginate(
        TableName=table,
        ProjectionExpression=", ".join(names),
        ExpressionAttributeNames=names,
        ConsistentRead=True,
    ):
        for item in page["Items"]:
            row_id = item.get("id", {}).get("S", "")
            if row_id.startswith("STATS#"):
                continue
            rows.append(
                Row(
                    id=row_id,
                    fields={f: item.get(f, {}).get("S", "") for f in EXTRACTION_FIELDS},
                    confidence=float(item.get("confidence", {}).get("N", "0")),
                    dedup_key=item.get("dedup_key", {}).get("S", ""),
                )
            )
    return rows


# ---------------------------------------------------------------------------
# Measures
# ---------------------------------------------------------------------------


def make_token_counter() -> Callable[[str], int]:
    """Counts tokens with OpenAI's o200k_base, the tokenizer of the current extractor
    model. Week 3 recalibrates with Bedrock Converse usage counts."""
    import tiktoken

    enc = tiktoken.get_encoding(TOKENIZER)
    return lambda text: len(enc.encode(text, disallowed_special=()))


def make_language_detector() -> Callable[[str], str]:
    """Returns an ISO 639-1 code, or "undetermined" for short or unclassifiable bodies.
    langdetect is seeded so its results are deterministic."""
    from langdetect import DetectorFactory, LangDetectException, detect

    DetectorFactory.seed = 0

    def detect_language(text: str) -> str:
        if sum(ch.isalpha() for ch in text) < MIN_LETTERS_FOR_LANGUAGE:
            return "undetermined"
        try:
            return detect(text)
        except LangDetectException:
            return "undetermined"

    return detect_language


def percentile(sorted_values: list[int], p: float) -> int:
    """Nearest-rank percentile (always an observed value, deterministic)."""
    if not sorted_values:
        return 0
    rank = max(1, math.ceil(p / 100 * len(sorted_values)))
    return sorted_values[rank - 1]


def pct(part: int, whole: int) -> str:
    return f"{part / whole * 100:.1f}%" if whole else "n/a"


def fingerprint(items: Iterable[str]) -> str:
    """Short SHA-256 over the canonical inputs, so a re-run on the same data is
    recognisable. A hash of the data, not the data."""
    h = hashlib.sha256()
    for item in sorted(items):
        h.update(item.encode("utf-8"))
        h.update(b"\n")
    return h.hexdigest()[:12]


def canonical(obj) -> str:
    return json.dumps(obj, sort_keys=True, ensure_ascii=False, separators=(",", ":"))


# ---------------------------------------------------------------------------
# Report
# ---------------------------------------------------------------------------


def build_report(
    records: list[dict],
    rows: list[Row],
    count_tokens: Callable[[str], int],
    detect_language: Callable[[str], str],
) -> str:
    """Renders the Markdown report. Input order doesn't matter."""
    records = sorted(records, key=lambda r: r.get("message_id", ""))
    rows = sorted(rows, key=lambda r: r.id)
    n_records, n_rows = len(records), len(rows)
    lines: list[str] = []
    out = lines.append

    out("# Data-quality report: recruiter email archive")
    out("")
    out("Aggregates only (D-06): this file contains no email text, names, addresses, companies,")
    out("subjects or message ids. Regenerate it from `infra/recruiter-dashboard/evals` with")
    out("`uv run data-quality --clean-s3-bucket <bucket> --table <table>`; the same data always")
    out("produces the same file. Spec: #134, PLAN.md § Week 1.")
    out("")

    # --- Inputs ---------------------------------------------------------------
    record_ids = {r.get("message_id", "") for r in records}
    row_ids = {r.id for r in rows}
    schema = Counter(f"schema {r.get('schema_version')} / parser {r.get('parser_version')}" for r in records)
    sources = Counter(r.get("extraction_source", "missing") for r in records)
    out("## Inputs")
    out("")
    out("| | Count |")
    out("|---|---|")
    out(f"| Clean records | {n_records} |")
    out(f"| DynamoDB extraction rows (all-time, excluding `STATS#`) | {n_rows} |")
    out(f"| Rows with a clean record (body still held) | {len(row_ids & record_ids)} |")
    out(f"| Rows without one (body lost before the archive existed) | {len(row_ids - record_ids)} |")
    out(f"| Clean records without a row | {len(record_ids - row_ids)} |")
    out("")
    out("Record versions: " + ", ".join(f"{k}: {v}" for k, v in sorted(schema.items())) + ".  ")
    out("Extraction source in the records: " + ", ".join(f"`{k}` {v}" for k, v in sorted(sources.items())) + ".  ")
    out("Clean records come from `archive/clean/`, or from `cmd/backfill --out-dir`, which writes the")
    out("same bytes; the records fingerprint is the same for both.  ")
    out(
        f"Fingerprints: records `{fingerprint(canonical(r) for r in records)}`, "
        f"rows `{fingerprint(canonical([r.id, r.fields, r.confidence, r.dedup_key]) for r in rows)}`."
    )
    out("")

    # --- Bodies -----------------------------------------------------------------
    bodies = [r.get("body", "") or "" for r in records]
    tokens = [count_tokens(b) for b in bodies]
    sorted_tokens = sorted(tokens)
    total_runes = sum(len(b) for b in bodies)
    total_tokens = sum(tokens)
    out("## Bodies")
    out("")
    out(f"Token counts use `{TOKENIZER}` (the current OpenAI extractor's tokenizer) over the cleaned")
    out("body the extractor sees. Percentiles are nearest-rank.")
    out("")
    out("| min | p10 | p25 | p50 | p75 | p90 | p95 | max | mean |")
    out("|---|---|---|---|---|---|---|---|---|")
    if tokens:
        stats = [percentile(sorted_tokens, p) for p in (10, 25, 50, 75, 90, 95)]
        mean = total_tokens / len(tokens)
        out(f"| {sorted_tokens[0]} | " + " | ".join(str(s) for s in stats) + f" | {sorted_tokens[-1]} | {mean:.0f} |")
    else:
        out("| n/a | n/a | n/a | n/a | n/a | n/a | n/a | n/a | n/a |")
    out("")
    out("| Tokens | Bodies | Share |")
    out("|---|---|---|")
    for lo, hi in TOKEN_BUCKETS:
        label = f"{lo:,}+" if hi is None else (f"{lo}" if lo == hi else f"{lo:,}–{hi:,}")
        count = sum(1 for t in tokens if t >= lo and (hi is None or t <= hi))
        out(f"| {label} | {count} | {pct(count, n_records)} |")
    out("")
    ratio = f"{total_runes / total_tokens:.2f}" if total_tokens else "n/a"
    out(f"Characters per token across all bodies: **{ratio}**. Week 3's `Truncate` cuts by runes")
    out("(`maxChars`), so this converts a token budget into a rune cap.")
    out("")
    empty = sum(1 for b in bodies if not b.strip())
    near_empty = sum(1 for b, t in zip(bodies, tokens) if b.strip() and t < NEAR_EMPTY_TOKENS)
    forwarded = sum(1 for r in records if r.get("is_forwarded"))
    out("| | Count | Share |")
    out("|---|---|---|")
    out(f"| Empty bodies (whitespace only) | {empty} | {pct(empty, n_records)} |")
    out(f"| Near-empty bodies (1–{NEAR_EMPTY_TOKENS - 1} tokens) | {near_empty} | {pct(near_empty, n_records)} |")
    out(f"| Forwarded (`is_forwarded`) | {forwarded} | {pct(forwarded, n_records)} |")
    out("")

    # --- Language ---------------------------------------------------------------
    languages = Counter(detect_language(b) for b in bodies)
    out("## Language mix")
    out("")
    out(f"`langdetect` (seeded) per body; bodies with fewer than {MIN_LETTERS_FOR_LANGUAGE} letters are")
    out("`undetermined`.")
    out("")
    out("| Language | Bodies | Share |")
    out("|---|---|---|")
    for lang, count in sorted(languages.items(), key=lambda kv: (-kv[1], kv[0])):
        out(f"| `{lang}` | {count} | {pct(count, n_records)} |")
    out("")

    # --- Extractions --------------------------------------------------------------
    out("## Extractions (DynamoDB, all-time)")
    out("")
    out("| Field | `Unknown` | Empty | Unknown or empty |")
    out("|---|---|---|---|")
    for f in EXTRACTION_FIELDS:
        unknown = sum(1 for r in rows if r.fields.get(f) == UNKNOWN)
        blank = sum(1 for r in rows if not (r.fields.get(f) or "").strip())
        out(f"| `{f}` | {pct(unknown, n_rows)} | {pct(blank, n_rows)} | {pct(unknown + blank, n_rows)} |")
    all_unknown = sum(1 for r in rows if all(r.fields.get(f) in (UNKNOWN, "") for f in EXTRACTION_FIELDS))
    out("")
    out(f"Rows where every field is `Unknown` or empty: {all_unknown} ({pct(all_unknown, n_rows)}).")
    out("")
    low = sum(1 for r in rows if r.confidence < 0.5)
    out(f"**Confidence < 0.5: {low} of {n_rows} rows ({pct(low, n_rows)}).**")
    out("")
    out("| Confidence | Rows | Share |")
    out("|---|---|---|")
    for label, test in CONFIDENCE_BUCKETS:
        count = sum(1 for r in rows if test(r.confidence))
        out(f"| {label} | {count} | {pct(count, n_rows)} |")
    out("")

    # --- Duplicates -----------------------------------------------------------------
    keyed = Counter(r.dedup_key for r in rows if r.dedup_key)
    placeholder_rows = keyed.pop(PLACEHOLDER_DEDUP_KEY, 0)
    dup_groups = {k: c for k, c in keyed.items() if c > 1}
    out("## Duplicates by `dedup_key`")
    out("")
    out("`dedup_key` hashes the extracted recruiter email and job title. Rows whose email and")
    out("title are both `Unknown` all share one placeholder key, so they're counted apart.")
    out("")
    out("| | Count |")
    out("|---|---|")
    out(f"| Distinct real keys | {len(keyed)} |")
    out(f"| Keys shared by more than one row | {len(dup_groups)} |")
    out(f"| Rows in those groups | {sum(dup_groups.values())} ({pct(sum(dup_groups.values()), n_rows)}) |")
    out(f"| Largest group | {max(dup_groups.values(), default=0)} |")
    out(f"| Rows on the placeholder (`Unknown` email and title) key | {placeholder_rows} ({pct(placeholder_rows, n_rows)}) |")
    out(f"| Rows with no `dedup_key` | {sum(1 for r in rows if not r.dedup_key)} |")
    out("")

    return "\n".join(lines)


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(prog="data-quality", description=__doc__.split("\n\n")[0])
    source = parser.add_mutually_exclusive_group(required=True)
    source.add_argument("--clean-s3-bucket", help="read archive/clean/ from this bucket (in memory)")
    source.add_argument("--clean-dir", type=Path, help="read {id}.json records from this local directory")
    parser.add_argument("--table", required=True, help="DynamoDB table with the extractions")
    parser.add_argument("--region", default="us-east-1", help="AWS region (default us-east-1)")
    parser.add_argument(
        "--out",
        type=Path,
        default=Path("reports/data-quality.md"),
        help="report path (default reports/data-quality.md)",
    )
    args = parser.parse_args(argv)
    if args.clean_dir is not None and not args.clean_dir.is_dir():
        parser.error(f"--clean-dir {args.clean_dir} is not a directory")
    return args


def main(argv: list[str] | None = None) -> int:
    args = parse_args(sys.argv[1:] if argv is None else argv)
    if args.clean_dir is not None:
        records = load_clean_records_from_dir(args.clean_dir)
    else:
        records = load_clean_records_from_s3(args.clean_s3_bucket, args.region)
    rows = load_dynamodb_rows(args.table, args.region)

    report = build_report(records, rows, make_token_counter(), make_language_detector())
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(report + "\n", encoding="utf-8")
    print(f"data-quality: {len(records)} records, {len(rows)} rows -> {args.out}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
