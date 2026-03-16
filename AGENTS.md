# AGENTS.md

Repository-specific guidance for Codex CLI (and compatible agents).
Preserve existing guidance in [CLAUDE.md](CLAUDE.md) and [.github/copilot-instructions.md](.github/copilot-instructions.md).

## Project Overview

Personal portfolio site (sh3r4rd.com): React 19, Vite 7, Tailwind CSS 3 frontend. Go 1.25 Lambda backend with DynamoDB, SES, and API Gateway. Terraform >= 1.5.0 infrastructure. Deployed to AWS S3/CloudFront via GitHub Actions on push to `main`.

## Commands

### Frontend

```bash
npm run dev          # Dev server
npm run build        # Production build → dist/
npm run lint         # ESLint
npm run preview      # Preview production build
make deploy bucket=<bucket-name>  # Build + sync to S3
```

### Backend (Go Lambda)

```bash
make build-lambdas       # Build all Lambda binaries (linux/arm64)
make build-email-parser  # Build email-parser only
make build-api-handler   # Build api-handler only

# Tests (run from email-parser directory)
cd infra/recruiter-dashboard/lambda-src/email-parser && go test -v -race ./...
cd infra/recruiter-dashboard/lambda-src/email-parser && go vet ./...
```

### Infrastructure (Terraform)

```bash
terraform -chdir=infra/recruiter-dashboard init
terraform -chdir=infra/recruiter-dashboard validate
terraform -chdir=infra/recruiter-dashboard fmt -check -recursive
terraform -chdir=infra/recruiter-dashboard plan
```

### All-in-one CI check

```bash
make ci  # lint + build + test-go + tf-fmt-check + tf-validate
```

## Project Structure

```
src/                         # React frontend (JSX only, no TypeScript)
├── App.jsx                  # Router entrypoint: /, /resume, *
├── pages/                   # Route-level page components (default exports)
├── components/layout/       # Header, Breadcrumbs, PageTracker (default exports)
├── components/sections/     # Content sections (default exports)
└── components/ui/           # Reusable primitives: Button, Card (named exports)

infra/recruiter-dashboard/
├── modules/                 # 7 Terraform modules: s3, dynamodb, lambda, iam,
│                            #   ses, api-gateway, monitoring
└── lambda-src/
    ├── email-parser/        # Go 1.25 — production
    │   ├── cmd/handler/     # Lambda entry point (main.go)
    │   ├── internal/        # 9 packages: handler, ssm, sanitizer, extractor,
    │   │                    #   models, parser, db, errors, tagger
    │   └── testdata/        # Test fixtures
    └── api-handler/         # Go 1.25 — stub for Phase 3
        └── main.go          # Returns 200 OK with CORS headers
```

## Code Style

### Frontend

Page components follow this layout:

```jsx
import Breadcrumbs from "../components/layout/Breadcrumbs";
import Header from "../components/layout/Header";

export default function MyPage() {
  return (
    <section className="max-w-4xl mx-auto space-y-12">
      <Breadcrumbs />
      <Header />
      {/* page content */}
    </section>
  );
}
```

- `.jsx` files only — never `.ts` or `.tsx`
- Tailwind utility classes only — no CSS modules, no inline `style`
- Always pair `dark:` variants: `text-gray-700 dark:text-gray-300`
- Icons from `lucide-react` only
- `components/ui/` uses named exports; all other components use default exports

### Backend (Go)

```go
// Table-driven test pattern
func TestParseRawEmail(t *testing.T) {
    tests := []struct {
        name    string
        input   []byte
        wantErr bool
    }{
        {"valid email", loadFixture(t, "valid.eml"), false},
        {"empty input", nil, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := parser.ParseRawEmail(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseRawEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

- `internal/` package pattern — all packages are unexported
- `bootstrap` binaries compiled for `linux/arm64` (`GOOS=linux GOARCH=arm64 CGO_ENABLED=0`)
- Colocated `_test.go` files with table-driven tests
- Test fixtures in `testdata/` directories
- Wrap errors with `%w`: `fmt.Errorf("context: %w", err)`
- Define interfaces for AWS clients (enables testability without mocks of concrete types)

### Git

- Conventional commits required: `feat:`, `fix:`, `docs:`, `chore:`, `ci:`, `refactor:`, `test:`
- Use `(infra)` scope for infrastructure: `feat(infra): add SES module`

## Boundaries

### ALWAYS

- Run validation commands before reporting work as complete (see Validation Checklist below)
- Follow conventional commit format
- Read existing files before modifying them
- Preserve dark mode support when changing UI
- For multi-file changes, explain the approach before editing

### ASK FIRST

- Adding npm or Go dependencies
- Creating new routes in `App.jsx`
- Modifying CI/CD workflows (`.github/workflows/`)
- Changing the API endpoint or payload structure (`POST https://api.sh3r4rd.com/requests`)
- Modifying DynamoDB schema or API Gateway configuration
- Modifying Terraform modules or adding new cloud resources
- Changing Lambda function signatures

### NEVER

- Create `.ts` or `.tsx` files
- Import or extend `src/App.css` or `src/assets/react.svg` (stale Vite scaffolding)
- Create `.go` files outside `infra/recruiter-dashboard/lambda-src/`
- Commit `.env` files, AWS credentials, API keys, or `terraform.tfvars`
- Add icon libraries other than `lucide-react`
- Add state management libraries (Redux, Zustand, etc.)
- Force push to any branch
- Skip validation steps without stating why

## Validation Checklist

Run the applicable checks before completing any task:

### Frontend changes

```bash
npm run lint   # Must exit 0
npm run build  # Must complete without errors
```

### Go Lambda changes

```bash
cd infra/recruiter-dashboard/lambda-src/email-parser
go vet ./...              # Must exit 0
go test -v -race ./...    # All tests must pass
```

```bash
make build-lambdas        # Must compile successfully (run from repo root)
```

### Infrastructure changes

```bash
terraform -chdir=infra/recruiter-dashboard fmt -check -recursive  # Must exit 0
terraform -chdir=infra/recruiter-dashboard validate               # Must be valid
```

### All changes

- Review your own diff for correctness before reporting completion
- Confirm no secrets or credentials in staged files
- If a validation step cannot run, state which step was skipped and why

## Environment Variables

### email-parser Lambda

- `RECRUITER_TABLE` — DynamoDB table name
- `EMAIL_BUCKET` — S3 bucket for incoming emails
- `S3_KEY_PREFIX` — S3 key prefix for emails
- `SSM_OPENAI_KEY_NAME` — Parameter Store path for OpenAI API key

### api-handler Lambda

- `CORS_ALLOW_ORIGIN` — CORS allowed origin header value
