# sh3r4rd.com

This is [my personal website](https://sh3r4rd.com) that I use to demonstrate professional abilities throughout the tech stack. Anyone interested will also be able to send me a 
request for my resume.

## Tools
It is built using React, Tailwind and Vite on the front-end and Go, Lambda, DynamoDB, SES, and API Gateway on the backend.

## Infrastructure

The recruiter dashboard backend is managed with Terraform under `infra/recruiter-dashboard/`. To get started:

```bash
cp infra/recruiter-dashboard/terraform.tfvars.example infra/recruiter-dashboard/terraform.tfvars
# Edit terraform.tfvars with your values
terraform -chdir=infra/recruiter-dashboard init
terraform -chdir=infra/recruiter-dashboard plan
```

See [`terraform.tfvars.example`](infra/recruiter-dashboard/terraform.tfvars.example) for required configuration.

## Recruiter Dashboard

The recruiter dashboard is built and deployed. It includes:

- **Email parsing pipeline** — SES receives recruiter emails, stores raw emails in S3, triggers a Go Lambda that parses the email, extracts recruiter data via OpenAI, and writes structured records to DynamoDB
- **API handler** — Stub Lambda behind API Gateway at `api.sh3r4rd.com`, ready for Phase 3 dashboard endpoints

## Backend Development

Lambda source code lives under `infra/recruiter-dashboard/lambda-src/`.

```bash
# Build all Lambda binaries (linux/arm64)
make build-lambdas

# Run email-parser tests
cd infra/recruiter-dashboard/lambda-src/email-parser && go test -v -race ./...
```

## Deployment

### GitHub Actions (Recommended)

The project includes a GitHub Actions workflow that automatically deploys to S3 when code is pushed to the main branch.

**Required GitHub Secrets:**
- `AWS_ACCESS_KEY_ID` - AWS access key with S3 permissions
- `AWS_SECRET_ACCESS_KEY` - AWS secret access key
- `S3_BUCKET_NAME` - Name of your S3 bucket

To set up secrets:
1. Go to your GitHub repository
2. Navigate to Settings → Secrets and variables → Actions
3. Add the three required secrets listed above

### Manual Deployment

Run the following command to deploy assets to production manually:

```bash
make deploy bucket=bucket-name
```

This requires AWS CLI to be configured locally.

## CI/CD

Three GitHub Actions workflows:

- **`ci.yml`** — Runs on all PRs and pushes to `main`. Parallel jobs: frontend lint + build, Go tests, Terraform validation
- **`deploy.yml`** — Deploys frontend to S3 on push to `main`
- **`release.yml`** — Runs semantic-release on push to `main` for automated versioning

## Conventional Commits

This repository uses [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) to describe commits and add automated tooling.

## AI Tooling

This repository includes tool-specific instructions for multiple coding assistants:

- Claude Code reads [CLAUDE.md](CLAUDE.md)
- Gemini reads [GEMINI.md](GEMINI.md)
- Codex CLI reads [AGENTS.md](AGENTS.md)
- GitHub Copilot reads [.github/copilot-instructions.md](.github/copilot-instructions.md) and any scoped instructions under [.github/instructions/](.github/instructions/)

Developers using Codex should start with [docs/codex.md](docs/codex.md) for installation, ChatGPT sign-in, repository workflow, and validation expectations.
