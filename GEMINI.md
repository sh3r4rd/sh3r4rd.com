# Project: sh3r4rd.com Monorepo

## Project Overview

This is a monorepo for the personal website `sh3r4rd.com`. It consists of a React-based frontend and a serverless backend for a "Recruiter Dashboard".

### Key Technologies

*   **Frontend**:
    *   Framework: React
    *   Build Tool: Vite
    *   Styling: Tailwind CSS
*   **Backend (Recruiter Dashboard)**:
    *   Language: Go 1.25
    *   Platform: AWS Lambda (arm64)
    *   Services: Amazon S3 (for raw email storage), DynamoDB (for parsed data), SES (for email receiving), and API Gateway (for a REST API).
    *   AI: OpenAI API for recruiter data extraction
*   **Infrastructure as Code**:
    *   Tool: Terraform
*   **CI/CD**:
    *   Platform: GitHub Actions for continuous deployment of the frontend.

### Directory Structure

*   `src/`: Contains the React frontend source code.
*   `infra/recruiter-dashboard/`: Contains the Terraform configuration for the backend infrastructure.
*   `infra/recruiter-dashboard/lambda-src/email-parser/`: Go Lambda for parsing recruiter emails.
    *   `cmd/handler/main.go`: Entry point
    *   `internal/`: 9 packages — `handler`, `ssm`, `sanitizer`, `extractor`, `models`, `parser`, `db`, `errors`, `tagger`
    *   `testdata/`: Test fixtures
*   `infra/recruiter-dashboard/lambda-src/api-handler/`: Stub Lambda for Phase 3 API endpoints.
*   `.github/workflows/`: GitHub Actions workflows (ci, deploy, release).

## Building and Running

### Frontend

1.  **Install Dependencies**:
    ```bash
    npm install
    ```
2.  **Run Development Server**:
    ```bash
    npm run dev
    ```
3.  **Build for Production**:
    ```bash
    npm run build
    ```
    The output will be in the `dist/` directory.

4.  **Linting**:
    ```bash
    npm run lint
    ```

### Backend (Go Lambda)

1.  **Build all Lambdas**:
    ```bash
    make build-lambdas
    ```
2.  **Run email-parser tests**:
    ```bash
    cd infra/recruiter-dashboard/lambda-src/email-parser && go test -v -race ./...
    ```

### Backend (Terraform)

The backend infrastructure is managed by Terraform.

1.  **Initialize Terraform**:
    ```bash
    terraform -chdir=infra/recruiter-dashboard init
    ```
2.  **Plan Changes**:
    ```bash
    terraform -chdir=infra/recruiter-dashboard plan
    ```
3.  **Apply Changes**:
    ```bash
    terraform -chdir=infra/recruiter-dashboard apply
    ```
    *Note: You will need to create a `terraform.tfvars` file based on the `terraform.tfvars.example`.*

## Deployment

### Frontend

Three GitHub Actions workflows automate CI/CD:

*   **`ci.yml`**: Runs on all PRs and pushes to `main`. Three parallel jobs: frontend (lint + build), Go tests (`go test -v -race ./...`), Terraform validation (fmt check + init + validate).
*   **`deploy.yml`**: Deploys frontend to S3 on push to `main`.
*   **`release.yml`**: Runs semantic-release on push to `main` using SSH deploy key.

You can also deploy manually using the Makefile:
```bash
make deploy bucket=your-s3-bucket-name
```

## SES Email Flow

SES receives email → stores raw email in S3 → triggers email-parser Lambda → parses email → extracts recruiter data via OpenAI → sanitizes fields → writes to DynamoDB → tags S3 object with parse results.

## Monitoring

CloudWatch alarms and budget notifications route through SNS to email for alerting.

## Development Conventions

*   **Conventional Commits**: This repository follows the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.
*   **Dark Mode**: Always provide `dark:` Tailwind variants for user-facing UI.
*   **Export Patterns**: `components/ui/` uses named exports; all other components use default exports.
*   **Stale Boilerplate**: Do not import or extend `src/App.css` or `src/assets/react.svg` (unused Vite scaffolding).
*   **Modular Terraform**: The Terraform code is organized into 7 reusable modules located in `infra/recruiter-dashboard/modules/`: s3, dynamodb, lambda, iam, ses, api-gateway, monitoring.
*   **Go Conventions**: `internal/` package pattern, `bootstrap` binaries for `linux/arm64`, colocated `_test.go` files, table-driven tests, test fixtures in `testdata/`.
