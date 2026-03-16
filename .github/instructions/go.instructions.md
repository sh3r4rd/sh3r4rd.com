---
applyTo: "infra/recruiter-dashboard/lambda-src/**"
---

# Go Lambda Instructions

## Package Structure

- Entry points live in `cmd/handler/main.go` (email-parser) or `main.go` (api-handler)
- All internal logic goes in `internal/` packages — these are unexported by Go convention
- email-parser packages: `handler`, `ssm`, `sanitizer`, `extractor`, `models`, `parser`, `db`, `errors`, `tagger`

## Testing

- Colocate test files as `*_test.go` alongside source files
- Use table-driven tests with `t.Run()` subtests
- Place test fixtures in `testdata/` directories
- Always run tests with `-race` flag: `go test -v -race ./...`

## Error Handling

- Use custom error types from `internal/errors/`
- Wrap errors with `%w` for unwrapping: `fmt.Errorf("context: %w", err)`
- Log errors at the handler level, not in utility packages

## AWS SDK v2 Patterns

- Define interfaces for AWS clients to enable testability (e.g., `S3Client interface`)
- Always pass `context.Context` as the first parameter
- Initialize AWS clients once on cold start in `main.go`, reuse across invocations

## Build Targets

- Compile for Lambda: `GOOS=linux GOARCH=arm64 CGO_ENABLED=0`
- Output binary must be named `bootstrap`
- Use Makefile targets: `make build-email-parser`, `make build-api-handler`

## Naming

- Follow Go standard: `camelCase` for unexported, `PascalCase` for exported
- Package names are lowercase, single-word where possible
- Test functions: `TestFunctionName_scenario`
