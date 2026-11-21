"""Tests for the data-quality report, on synthetic data only (D-06: no real email in git)."""

import hashlib
import json
import random

import pytest

from recruiter_evals import data_quality as dq

# Synthetic values that must never appear in a report.
SECRETS = [
    "Jane", "Doe", "jane.doe@example.com", "ExampleCorp", "Staff Platform Engineer",
    "+15555550100", "Confidential subject line", "unique-body-marker-xyz", "msg-0001",
]


def word_count(text: str) -> int:
    """Deterministic stand-in for tiktoken (no network in tests)."""
    return len(text.split())


def fake_language(text: str) -> str:
    if sum(c.isalpha() for c in text) < dq.MIN_LETTERS_FOR_LANGUAGE:
        return "undetermined"
    return "es" if "hola" in text else "en"


def record(i: int, body: str, forwarded: bool = False) -> dict:
    return {
        "schema_version": 1,
        "parser_version": "1",
        "message_id": f"msg-{i:04d}",
        "raw_key": f"archive/raw/msg-{i:04d}",
        "headers": {"from": "Jane Doe <jane.doe@example.com>", "to": "me@example.org",
                    "subject": "Confidential subject line", "date": "2026-11-01T10:00:00Z"},
        "is_forwarded": forwarded,
        "body": body,
        "extraction": {"company": "ExampleCorp"},
        "extraction_source": "dynamodb",
    }


def row(i: int, **overrides) -> dq.Row:
    fields = {
        "first_name": "Jane", "last_name": "Doe", "recruiter_email": "jane.doe@example.com",
        "company": "ExampleCorp", "job_title": "Staff Platform Engineer", "phone": "+15555550100",
    }
    fields.update({k: v for k, v in overrides.items() if k in fields})
    return dq.Row(
        id=f"msg-{i:04d}",
        fields=fields,
        confidence=overrides.get("confidence", 0.9),
        dedup_key=overrides.get("dedup_key", f"key-{i}"),
    )


@pytest.fixture
def dataset():
    long_body = "unique-body-marker-xyz " + "word " * 600  # 601 words
    records = [
        record(1, long_body, forwarded=True),
        record(2, "short note about a role " * 10),  # 50 words
        record(3, "   \n "),  # empty
        record(4, "hola equipo, tenemos una vacante interesante para ti " * 3),
        record(5, "tiny"),  # near-empty, undetermined language
    ]
    unknown = {f: "Unknown" for f in dq.EXTRACTION_FIELDS}
    rows = [
        row(1),
        row(2, dedup_key="shared"),
        row(3, dedup_key="shared", confidence=0.4),
        row(4, **unknown, confidence=0.0, dedup_key=dq.PLACEHOLDER_DEDUP_KEY),
        row(6, **unknown, confidence=0.0, dedup_key=dq.PLACEHOLDER_DEDUP_KEY),  # body lost
        row(7, phone="", confidence=0.6),  # body lost
    ]
    return records, rows


def render(records, rows):
    return dq.build_report(records, rows, word_count, fake_language)


def test_placeholder_key_matches_the_lambda():
    # models.GenerateDedupKey("Unknown", "Unknown") lowercases, joins with "|", and
    # hex-encodes the first 16 bytes of SHA-256.
    assert dq.PLACEHOLDER_DEDUP_KEY == hashlib.sha256(b"unknown|unknown").digest()[:16].hex()


def test_report_contains_no_email_content(dataset):
    report = render(*dataset)
    for secret in SECRETS:
        assert secret not in report, f"report leaks {secret!r}"
    assert "@" not in report


def test_report_is_deterministic_and_order_independent(dataset):
    records, rows = dataset
    first = render(records, rows)
    shuffled_records, shuffled_rows = records[:], rows[:]
    random.Random(7).shuffle(shuffled_records)
    random.Random(7).shuffle(shuffled_rows)
    assert render(shuffled_records, shuffled_rows) == first
    assert render(records, rows) == first


def test_fingerprint_changes_with_the_data(dataset):
    records, rows = dataset
    before = render(records, rows)
    changed = [dict(records[0], body="different")] + records[1:]
    after = render(changed, rows)
    fp = lambda r: next(line for line in r.splitlines() if line.startswith("Fingerprints:"))
    assert fp(before) != fp(after)


def test_inputs_and_coverage(dataset):
    report = render(*dataset)
    assert "| Clean records | 5 |" in report
    assert "| DynamoDB extraction rows (all-time, excluding `STATS#`) | 6 |" in report
    assert "| Rows with a clean record (body still held) | 4 |" in report
    assert "| Rows without one (body lost before the archive existed) | 2 |" in report
    assert "| Clean records without a row | 1 |" in report  # msg-0005
    assert "`dynamodb` 5" in report


def test_bodies(dataset):
    report = render(*dataset)
    # word counts: 601, 50, 0, 24, 1
    assert "| 0 | 0 | 1 | 24 | 50 | 601 | 601 | 601 | 135 |" in report
    assert "| 0 | 1 | 20.0% |" in report  # empty bucket
    assert "| 1–24 | 2 | 40.0% |" in report
    assert "| 25–249 | 1 | 20.0% |" in report
    assert "| 500–999 | 1 | 20.0% |" in report
    assert "| Empty bodies (whitespace only) | 1 | 20.0% |" in report
    assert "| Near-empty bodies (1–24 tokens) | 2 | 40.0% |" in report
    assert "| Forwarded (`is_forwarded`) | 1 | 20.0% |" in report


def test_language_mix(dataset):
    report = render(*dataset)
    assert "| `en` | 2 | 40.0% |" in report
    assert "| `undetermined` | 2 | 40.0% |" in report
    assert "| `es` | 1 | 20.0% |" in report


def test_extractions(dataset):
    report = render(*dataset)
    assert "| `company` | 33.3% | 0.0% | 33.3% |" in report
    assert "| `phone` | 33.3% | 16.7% | 50.0% |" in report
    assert "Rows where every field is `Unknown` or empty: 2 (33.3%)." in report
    assert "**Confidence < 0.5: 3 of 6 rows (50.0%).**" in report
    assert "| 0 | 2 | 33.3% |" in report
    assert "| (0, 0.5) | 1 | 16.7% |" in report


def test_duplicates(dataset):
    report = render(*dataset)
    assert "| Keys shared by more than one row | 1 |" in report
    assert "| Rows in those groups | 2 (33.3%) |" in report
    assert "| Largest group | 2 |" in report
    assert "| Rows on the placeholder (`Unknown` email and title) key | 2 (33.3%) |" in report


def test_empty_inputs_render():
    report = render([], [])
    assert "| Clean records | 0 |" in report
    assert "n/a" in report


def test_percentile_nearest_rank():
    values = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
    assert dq.percentile(values, 50) == 5
    assert dq.percentile(values, 90) == 9
    assert dq.percentile(values, 95) == 10
    assert dq.percentile(values, 10) == 1
    assert dq.percentile([], 50) == 0


def test_load_clean_records_from_dir(tmp_path, dataset):
    records, _ = dataset
    for r in records:
        (tmp_path / f"{r['message_id']}.json").write_text(json.dumps(r), encoding="utf-8")
    (tmp_path / "notes.txt").write_text("ignored", encoding="utf-8")
    loaded = dq.load_clean_records_from_dir(tmp_path)
    assert [r["message_id"] for r in loaded] == sorted(r["message_id"] for r in records)


def test_real_language_detector_is_deterministic():
    detect = dq.make_language_detector()
    text = "We are hiring a senior engineer to build data pipelines for our platform team."
    assert detect(text) == detect(text) == "en"
    assert detect("ok") == "undetermined"
