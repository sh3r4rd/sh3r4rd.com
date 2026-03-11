# Codex CLI

This guide explains how to use Codex CLI in this repository. It assumes developers authenticate with ChatGPT sign-in rather than an API key.

## Install

Install Codex CLI using one of the official distribution methods:

- `npm install -g @openai/codex`
- `brew install codex`

After installation, verify the CLI is available:

```bash
codex --version
```

## Sign In

Authenticate from the terminal:

```bash
codex login
```

Choose the ChatGPT sign-in flow when prompted and complete authentication in the browser.

## Start in This Repository

From the repository root:

```bash
codex
```

Codex will automatically read repository guidance from [AGENTS.md](/Users/sherardbailey/Code/sh3r4rd.com/AGENTS.md). That file is intentionally Codex-specific. If you are using Claude Code or Copilot, use their own instruction files instead.

## Recommended Workflow

1. Start Codex from the repository root so it can pick up `AGENTS.md`.
2. Tell Codex the task, expected scope, and any constraints up front.
3. Ask it to inspect existing files before making architectural changes.
4. Prefer small, reviewable changes and ask it to run validation before finishing.
5. Review the resulting diff before committing.

Example prompts:

- `Update the home page copy and run lint.`
- `Investigate why the resume request form fails validation on mobile and fix it.`
- `Add a new section to the homepage, keep the existing visual language, and run build.`

## Repository Expectations

Codex should follow these project rules:

- Use `.jsx` files only
- Keep styling in Tailwind utility classes
- Preserve light and dark mode support
- Use `lucide-react` for icons
- Do not import `src/App.css` or `src/assets/react.svg`
- Treat route changes, dependency additions, deployment changes, and API contract changes as approval points

## Validation

For most code changes, ask Codex to run:

```bash
npm run lint
```

For changes that could affect production output, also run:

```bash
npm run build
```

If a validation step is skipped or fails, record that in the handoff.

## Sandbox and Approvals

Codex CLI may ask for approval before running commands that need broader filesystem or network access. In this repository, that is most likely when:

- Installing dependencies
- Running commands outside the workspace
- Accessing external services
- Performing potentially destructive actions

Read approval prompts carefully and grant them only when the action matches the task.

## Related Files

- Codex instructions: [AGENTS.md](/Users/sherardbailey/Code/sh3r4rd.com/AGENTS.md)
- Claude Code instructions: [CLAUDE.md](/Users/sherardbailey/Code/sh3r4rd.com/CLAUDE.md)
- Copilot instructions: [.github/copilot-instructions.md](/Users/sherardbailey/Code/sh3r4rd.com/.github/copilot-instructions.md)
