# evals

Evaluation harness, datasets and aggregate reports for Recruiter Intelligence
(`docs/recruiter-intelligence/`). Python, per D-03; managed with [uv](https://docs.astral.sh/uv/).

**Real email never goes in git (D-06).** Golden datasets built from real emails live in the
private bucket under `s3://…/evals/` (SSE-KMS with the archive key). This directory holds code,
schemas, synthetic fixtures and **aggregate-only** reports.

## Setup

```bash
brew install uv          # once
cd infra/recruiter-dashboard/evals
uv sync                  # creates .venv from uv.lock (Python 3.13, downloaded by uv if needed)
uv run pytest            # synthetic tests only; no AWS access
```

`boto3` is installed with the `crt` extra so it can use the AWS CLI's `aws login` credentials.

## Data-quality report (#134)

`reports/data-quality.md` describes the archive before anything is evaluated on it: body
count, token-length distribution, empty/near-empty bodies, forwarded share, language mix,
duplicates by `dedup_key`, the `Unknown` rate per extraction field, and how often
`confidence < 0.5`.

```bash
# From archive/clean/ (normal case, after the backfill has run):
uv run data-quality --clean-s3-bucket sh3r4rd-recruiter-emails --table recruiter-emails

# From a local backfill output instead (cmd/backfill --out-dir; the same records):
uv run data-quality --clean-dir <dir> --table recruiter-emails
```

- **Read-only:** it reads `archive/clean/` and scans the DynamoDB table (only the attributes it
  needs) into memory, and writes nothing but the report.
- **Aggregates only:** the report has counts, rates and distributions, never text, names,
  addresses, companies, subjects or message ids. `tests/test_data_quality.py` asserts this
  on synthetic data. Before committing a regenerated report, read it.
- **Deterministic:** no timestamps; the same data produces the same bytes. The `Fingerprints:`
  line identifies the input data.
- **Tokens** use `o200k_base` (the current OpenAI extractor's tokenizer; `tiktoken` downloads
  it on first use). Week 3 recalibrates with Bedrock Converse usage counts.
- **Language** uses `langdetect`, seeded for determinism.
- A `--clean-dir` written by `cmd/backfill --out-dir` holds real email content: keep it private
  (`mkdir -m 700`) and delete it when the report is done.

## Layout

```
pyproject.toml, uv.lock, .python-version
src/recruiter_evals/       # package
  data_quality.py          # the report (CLI: data-quality)
tests/                     # pytest, synthetic data only
reports/                   # committed aggregate reports
```

CI doesn't run these tests yet; the eval workflow arrives in week 11 (PLAN.md), and changing
`.github/workflows/` is an ask-first action.
