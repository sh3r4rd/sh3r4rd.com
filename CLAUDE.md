# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Personal portfolio website (sh3r4rd.com) built with React, Tailwind CSS, and Vite. Deployed to AWS S3/CloudFront via GitHub Actions on push to `main`.

- **GitHub repository:** `sh3r4rd/sh3r4rd.com` — always use this for `gh` CLI commands (e.g., `gh pr view`, `gh api`)

## Commands

### Frontend

- **Dev server:** `npm run dev` (or `make server`)
- **Build:** `npm run build` (or `make build`)
- **Lint:** `npm run lint`
- **Deploy (manual):** `make deploy bucket=<bucket-name>` (builds then syncs to S3)
- **Preview production build:** `npm run preview`

### Backend (Go Lambda)

- **Build all Lambdas:** `make build-lambdas`
- **Build email parser:** `make build-email-parser`
- **Build API handler:** `make build-api-handler`
- **Run email-parser tests:** `cd infra/recruiter-dashboard/lambda-src/email-parser && go test ./...`
- **Run email-parser tests (verbose + race):** `cd infra/recruiter-dashboard/lambda-src/email-parser && go test -v -race ./...`
- **Run api-handler tests:** `cd infra/recruiter-dashboard/lambda-src/api-handler && RECRUITER_TABLE=test CORS_ALLOW_ORIGIN=http://localhost DATE_INDEX_NAME=date-index AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test AWS_REGION=us-east-1 go test -v -race ./...`
- **Run all CI checks locally:** `make ci`

### Infrastructure (Terraform)

All commands run from `infra/recruiter-dashboard/`:

- **Init:** `terraform -chdir=infra/recruiter-dashboard init`
- **Validate:** `terraform -chdir=infra/recruiter-dashboard validate`
- **Format:** `terraform -chdir=infra/recruiter-dashboard fmt -recursive`
- **Plan:** `terraform -chdir=infra/recruiter-dashboard plan`

## Architecture

### Frontend

- **React 19** with **react-router-dom v7** for client-side routing (BrowserRouter)
- **Vite 7** as build tool with React plugin
- **Tailwind CSS 3** with typography plugin, PostCSS/Autoprefixer
- **ESLint 9** flat config with react-hooks and react-refresh plugins

### Infrastructure (`infra/recruiter-dashboard/`)

- **Terraform >= 1.5.0** with **AWS provider >= 5.31.0**
- Directory structure: root module + 7 modules: `modules/s3`, `modules/dynamodb`, `modules/lambda`, `modules/iam`, `modules/ses`, `modules/api-gateway`, `modules/monitoring`
- Local state (single-developer project)
- AWS region: us-east-1 (required for SES email receiving)
- DynamoDB: provisioned billing mode (15/15 RCU/WCU, under 25 free tier)
- S3: AES-256 encryption, 30-day lifecycle, full public access block
- Default tags applied via provider `default_tags` block

### Backend Lambda Functions

- **email-parser** (`infra/recruiter-dashboard/lambda-src/email-parser/`) — Go 1.25, production-ready
  - Entry point: `cmd/handler/main.go`
  - 9 internal packages: `handler`, `ssm`, `sanitizer`, `extractor`, `models`, `parser`, `db`, `errors`, `tagger`
  - Env vars: `RECRUITER_TABLE`, `EMAIL_BUCKET`, `S3_KEY_PREFIX`, `SSM_OPENAI_KEY_NAME`
  - Test fixtures in `testdata/`
- **api-handler** (`infra/recruiter-dashboard/lambda-src/api-handler/`) — Go 1.25, serves recruiter dashboard REST API
  - 3 source files: `main.go` (entry point), `handler.go` (routing + DynamoDB queries), `anonymizer.go` (PII stripping)
  - 3 endpoints: `GET /recruiters` (list, filterable by `?company=` or `?month=YYYY-MM`), `GET /recruiters/{id}`, `GET /stats`
  - Anonymization layer removes PII (recruiter_email, first_name, last_name, phone, s3_key, s3_bucket, dedup_key) from all responses
  - Env vars: `RECRUITER_TABLE`, `CORS_ALLOW_ORIGIN`, `DATE_INDEX_NAME`

### SES Email Flow

SES receives email → stores raw email in S3 → triggers email-parser Lambda → parses email with `internal/parser` → extracts recruiter data via OpenAI (`internal/extractor`) → sanitizes fields (`internal/sanitizer`) → writes to DynamoDB (`internal/db`) → tags S3 object with parse results (`internal/tagger`)

### DynamoDB Schema

- **Table:** `recruiter_emails`
- **Primary key:** `id` (S, partition) + `received_at` (S, sort)
- **GSI `recruiter-index`:** `recruiter_email` (HASH) + `received_at` (RANGE)
- **GSI `date-index`:** `date_year` (HASH) + `date_day` (RANGE)
- **Key attributes:** `id`, `received_at`, `recruiter_email`, `date_year`, `date_day`, plus `first_name`, `last_name`, `email`, `company`, `job_title`, `phone`, `subject`, `confidence`, `s3_bucket`, `s3_key`, `dedup_key`

### CI/CD

- **`ci.yml`** — Runs on push to `main` and all PRs. Three parallel jobs: frontend (lint + build), Go tests (`go test -v -race ./...`), Terraform validation (fmt check + init + validate)
- **`deploy.yml`** — Deploys frontend to S3 on push to `main`
- **`release.yml`** — Runs semantic-release on push to `main` using SSH deploy key to bypass branch ruleset

### Source Structure (`src/`)

- `App.jsx` — Root component with router; routes: `/` (HomePage), `/resume` (ResumeRequestPage), `*` (NotFoundPage)
- `pages/` — Page-level route components
- `components/layout/` — Header, Breadcrumbs, PageTracker (analytics)
- `components/sections/` — Content sections (e.g., SkillsSection)
- `components/ui/` — Reusable primitives (button, card)

### Key Conventions

- JSX files (not TypeScript) — all components use `.jsx` extension
- Dark mode support via Tailwind `dark:` variants
- Icons from `lucide-react`
- Conventional commits required (see [conventionalcommits.org](https://www.conventionalcommits.org/en/v1.0.0/))
- Terraform: HCL files only, module pattern (`modules/<name>/main.tf`, `variables.tf`, `outputs.tf`), see `terraform.tfvars.example` for variable defaults
- Go: `internal/` package pattern (email-parser) or flat package (api-handler), `bootstrap` binaries compiled for `linux/arm64`, colocated `_test.go` files, table-driven tests, test fixtures in `testdata/`
- Custom Claude Code commands in `.claude/commands/` (`component`, `add-skill`, `deploy-check`, `pr-fix`)

### Stale Boilerplate

- `src/App.css` and `src/assets/react.svg` are unused Vite scaffolding — do not import or extend them

### Component Patterns

- Content pages wrap in `<section className="max-w-4xl mx-auto space-y-12">` with `<Breadcrumbs />` then `<Header />` (exception: `NotFoundPage` uses a minimal layout)
- `components/ui/` uses **named exports** (e.g., `export function Button`, `export function Card`)
- All other components (pages, layout, sections) use **default exports**

### Backend API

- Resume request form POSTs to `https://api.sh3r4rd.com/requests`
- Payload fields: `firstName`, `lastName`, `email`, `phone`, `company`, `jobTitle`, `description`
- Backend infrastructure is defined in `infra/recruiter-dashboard/`

### Analytics

- Google Analytics via `window.gtag` in `PageTracker` component
- Measurement ID `G-L2852SHBRS` loaded in `index.html`

### Deployment

- GitHub Actions workflow (`.github/workflows/deploy.yml`) auto-deploys on push to `main`
- `index.html` is deployed with no-cache headers; other assets use default caching
- Images in `public/images/` are excluded from S3 sync
