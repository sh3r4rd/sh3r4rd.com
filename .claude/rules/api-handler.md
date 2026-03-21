---
paths:
  - "infra/recruiter-dashboard/lambda-src/api-handler/**"
---

# API Handler Lambda

Serves the recruiter dashboard REST API with anonymized responses. All PII is stripped before returning data.

## Endpoints

- `GET /recruiters` — List anonymized recruiter emails. Filters: `?company=X` (case-insensitive contains), `?month=YYYY-MM` (date-index GSI query)
- `GET /recruiters/{id}` — Single anonymized recruiter email by ID (queries partition key, returns most recent)
- `GET /stats` — Aggregate statistics: totalEmails, uniqueCompanies, byMonth, topJobTitles (top 10)
- `OPTIONS *` — CORS preflight (returns 204 No Content)

## Source Files

- `main.go` — Entry point, AWS config init, env var loading
- `handler.go` — Request routing, DynamoDB queries (Scan, Query by month via date-index GSI), response helpers
- `anonymizer.go` — `AnonymizedItem` struct, PII stripping, DynamoDB attribute helpers

## Anonymization

PII fields **never** included in responses: `recruiter_email`, `first_name`, `last_name`, `phone`, `s3_key`, `s3_bucket`, `dedup_key`. Output shape: `id`, `company`, `jobTitle`, `month` (coarsened to YYYY-MM), `recruiterLabel` ("Recruiter at {Company}"), `confidence`.

## Environment Variables

- `RECRUITER_TABLE` — DynamoDB table name
- `CORS_ALLOW_ORIGIN` — CORS allowed origin header value
- `DATE_INDEX_NAME` — Name of the date-index GSI

## Testing

```bash
cd infra/recruiter-dashboard/lambda-src/api-handler
RECRUITER_TABLE=test CORS_ALLOW_ORIGIN=http://localhost DATE_INDEX_NAME=date-index \
  AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=us-east-1 \
  go test -v -race ./...
```

## Patterns

- `DynamoDBAPI` interface for testability (GetItem, Query, Scan)
- Paginated scans with `ExclusiveStartKey` loop
- `ProjectionExpression` in stats scan to minimize RCU
- In-memory filtering and sorting after DynamoDB reads
