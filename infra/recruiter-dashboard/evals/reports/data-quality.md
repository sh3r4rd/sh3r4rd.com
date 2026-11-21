# Data-quality report: recruiter email archive

Aggregates only (D-06): this file contains no email text, names, addresses, companies,
subjects or message ids. Regenerate it from `infra/recruiter-dashboard/evals` with
`uv run data-quality --clean-s3-bucket <bucket> --table <table>`; the same data always
produces the same file. Spec: #134, PLAN.md § Week 1.

## Inputs

| | Count |
|---|---|
| Clean records | 25 |
| DynamoDB extraction rows (all-time, excluding `STATS#`) | 92 |
| Rows with a clean record (body still held) | 25 |
| Rows without one (body lost before the archive existed) | 67 |
| Clean records without a row | 0 |

Record versions: schema 1 / parser 1: 25.  
Extraction source in the records: `dynamodb` 25.  
Clean records come from `archive/clean/`, or from `cmd/backfill --out-dir`, which writes the
same bytes; the records fingerprint is the same for both.  
Fingerprints: records `6939c45dc0a4`, rows `b3664d1b5fb7`.

## Bodies

Token counts use `o200k_base` (the current OpenAI extractor's tokenizer) over the cleaned
body the extractor sees. Percentiles are nearest-rank.

| min | p10 | p25 | p50 | p75 | p90 | p95 | max | mean |
|---|---|---|---|---|---|---|---|---|
| 189 | 260 | 278 | 322 | 346 | 459 | 525 | 583 | 336 |

| Tokens | Bodies | Share |
|---|---|---|
| 0 | 0 | 0.0% |
| 1–24 | 0 | 0.0% |
| 25–249 | 1 | 4.0% |
| 250–499 | 22 | 88.0% |
| 500–999 | 2 | 8.0% |
| 1,000–1,999 | 0 | 0.0% |
| 2,000–3,999 | 0 | 0.0% |
| 4,000+ | 0 | 0.0% |

Characters per token across all bodies: **4.28**. Week 3's `Truncate` cuts by runes
(`maxChars`), so this converts a token budget into a rune cap.

| | Count | Share |
|---|---|---|
| Empty bodies (whitespace only) | 0 | 0.0% |
| Near-empty bodies (1–24 tokens) | 0 | 0.0% |
| Forwarded (`is_forwarded`) | 24 | 96.0% |

## Language mix

`langdetect` (seeded) per body; bodies with fewer than 20 letters are
`undetermined`.

| Language | Bodies | Share |
|---|---|---|
| `en` | 25 | 100.0% |

## Extractions (DynamoDB, all-time)

| Field | `Unknown` | Empty | Unknown or empty |
|---|---|---|---|
| `first_name` | 0.0% | 0.0% | 0.0% |
| `last_name` | 0.0% | 0.0% | 0.0% |
| `recruiter_email` | 0.0% | 0.0% | 0.0% |
| `company` | 0.0% | 0.0% | 0.0% |
| `job_title` | 9.8% | 0.0% | 9.8% |
| `phone` | 67.4% | 0.0% | 67.4% |

Rows where every field is `Unknown` or empty: 0 (0.0%).

**Confidence < 0.5: 0 of 92 rows (0.0%).**

| Confidence | Rows | Share |
|---|---|---|
| 0 | 0 | 0.0% |
| (0, 0.5) | 0 | 0.0% |
| [0.5, 0.8) | 7 | 7.6% |
| [0.8, 1] | 85 | 92.4% |

## Duplicates by `dedup_key`

`dedup_key` hashes the extracted recruiter email and job title. Rows whose email and
title are both `Unknown` all share one placeholder key, so they're counted apart.

| | Count |
|---|---|
| Distinct real keys | 89 |
| Keys shared by more than one row | 3 |
| Rows in those groups | 6 (6.5%) |
| Largest group | 2 |
| Rows on the placeholder (`Unknown` email and title) key | 0 (0.0%) |
| Rows with no `dedup_key` | 0 |

