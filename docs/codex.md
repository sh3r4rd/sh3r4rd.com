# Codex CLI

This guide explains how to use Codex CLI in this repository.

## Install

```bash
npm install -g @openai/codex
# or
brew install codex
```

Verify the installation:

```bash
codex --version
```

## Sign In

```bash
codex login
```

Choose the ChatGPT sign-in flow when prompted and complete authentication in the browser.

## Start in This Repository

From the repository root:

```bash
codex
```

Codex automatically reads [AGENTS.md](../AGENTS.md) for repository-specific guidance. A subdirectory override at [infra/recruiter-dashboard/AGENTS.md](../infra/recruiter-dashboard/AGENTS.md) provides additional context when working on backend infrastructure.

## Configuration

Create a project-level config file for repository-specific defaults:

```bash
mkdir -p .codex
```

**`.codex/config.toml`** (recommended starting point):

```toml
model = "o4-mini"

# workspace-write: read + write within workspace, network blocked
sandbox_mode = "workspace-write"

# on-request: autonomous within sandbox, asks when exceeding boundaries
approval_policy = "on-request"

# Increase from 32K default — this monorepo has detailed AGENTS.md files
project_doc_max_bytes = 65536

# Also read CLAUDE.md if no AGENTS.md exists in a subdirectory
project_doc_fallback_filenames = ["CLAUDE.md"]
```

### Profiles

Define profiles in config.toml for different work modes:

```toml
[profiles.frontend]
model = "o4-mini"

[profiles.infra]
model = "o3"
```

Use with:

```bash
codex --profile frontend
codex --profile infra
```

## Approval Modes

| Mode | Sandbox | Approval | Flag | Use Case |
|------|---------|----------|------|----------|
| Suggest | `read-only` | `untrusted` | (default) | Code review, exploration |
| Auto-edit | `workspace-write` | `on-request` | default with config above | Normal development |
| Full-auto | `workspace-write` | `never` | `--full-auto` | Trusted automation |

For full-auto mode, AGENTS.md includes a validation checklist and strict boundaries to compensate for the lack of human approval. Use `make ci` to run all checks at once.

Switch modes during a session with `/permissions`.

## Recommended Workflow

1. Start Codex from the repository root so it picks up `AGENTS.md`.
2. State the task, expected scope, and constraints up front.
3. Ask it to inspect existing files before making architectural changes.
4. For multi-file changes, use Plan mode (`/plan` or Shift+Tab) to outline the approach first.
5. Ask it to run validation before finishing: `make ci` or the applicable subset.
6. Review the resulting diff before committing.

### Example Prompts

**Frontend:**
```
Update the home page hero section copy and run npm run lint && npm run build.
```

**Backend:**
```
Add phone number validation to the sanitizer package. Write table-driven tests. Run go test -v -race ./... when done.
```

**Infrastructure:**
```
Add a TTL attribute to the DynamoDB table for auto-expiring old records. Run terraform validate and terraform fmt -check -recursive.
```

## Hierarchical AGENTS.md

Codex discovers instruction files by walking from the git root to the current directory:

```
AGENTS.md                           ← root: project overview, commands, boundaries
infra/recruiter-dashboard/AGENTS.md ← subdirectory: Go/Terraform-specific context
```

Files are concatenated root-downward. Subdirectory instructions supplement (not replace) the root file. You can also create `AGENTS.override.md` at any level to suppress the regular file at that level.

## Validation

For most code changes, ask Codex to run:

```bash
npm run lint && npm run build
```

For Go changes:

```bash
cd infra/recruiter-dashboard/lambda-src/email-parser && go vet ./... && go test -v -race ./...
```

For infrastructure changes:

```bash
terraform -chdir=infra/recruiter-dashboard fmt -check -recursive && terraform -chdir=infra/recruiter-dashboard validate
```

Or run everything at once:

```bash
make ci
```

If a validation step is skipped or fails, record that in the handoff.

## Sandbox and Approvals

Codex CLI may ask for approval before running commands that need broader filesystem or network access. In this repository, that is most likely when:

- Installing dependencies (`npm install`, `go get`)
- Running commands outside the workspace
- Accessing external services
- Performing potentially destructive actions

Read approval prompts carefully and grant them only when the action matches the task.

## Related Files

| File | Agent |
|------|-------|
| [AGENTS.md](../AGENTS.md) | Codex CLI (root) |
| [infra/recruiter-dashboard/AGENTS.md](../infra/recruiter-dashboard/AGENTS.md) | Codex CLI (backend) |
| [CLAUDE.md](../CLAUDE.md) | Claude Code |
| [GEMINI.md](../GEMINI.md) | Gemini |
| [.github/copilot-instructions.md](../.github/copilot-instructions.md) | GitHub Copilot |
| [.github/instructions/](../.github/instructions/) | GitHub Copilot (scoped) |
