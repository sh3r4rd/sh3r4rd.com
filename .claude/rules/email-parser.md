---
paths:
  - "infra/recruiter-dashboard/lambda-src/email-parser/**"
---

# Email Parser Lambda

Processes incoming recruiter emails received via SES. Parses raw email, extracts recruiter data via OpenAI, and writes structured records to DynamoDB.

## Pipeline Flow

SES receives email -> stores raw email in S3 -> triggers this Lambda -> `internal/parser` parses raw email -> `internal/extractor` calls OpenAI to extract recruiter data -> `internal/sanitizer` cleans/validates fields -> `internal/db` writes to DynamoDB -> `internal/tagger` tags S3 object with parse results

## Package Structure (`internal/`)

| Package | Purpose |
|---------|---------|
| `handler` | Lambda handler orchestration |
| `parser` | Raw email parsing (MIME) |
| `extractor` | OpenAI-based data extraction |
| `sanitizer` | Field cleaning and validation |
| `db` | DynamoDB write operations |
| `ssm` | Parameter Store client (OpenAI API key) |
| `tagger` | S3 object tagging with results |
| `models` | Shared data types (`email.go`, `extraction.go`) |
| `errors` | Custom error types |

## Entry Point

`cmd/handler/main.go` — Initializes AWS clients on cold start, wires dependencies.

## Environment Variables

- `RECRUITER_TABLE` — DynamoDB table name
- `EMAIL_BUCKET` — S3 bucket for incoming emails
- `S3_KEY_PREFIX` — S3 key prefix for emails
- `SSM_OPENAI_KEY_NAME` — Parameter Store path for OpenAI API key

## Testing

```bash
cd infra/recruiter-dashboard/lambda-src/email-parser
go test -v -race ./...
```

Test fixtures live in `testdata/` — never modify without explicit instruction.

## Patterns

- Interfaces for all AWS clients (S3, DynamoDB, SSM) for testability
- Custom error types in `internal/errors/` with `%w` wrapping
- Log errors at handler level, not in utility packages
- `context.Context` passed as first parameter throughout
