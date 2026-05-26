# Phase 4: Dashboard Frontend — Implementation Notes

Running log of decisions, deviations from issue specs, and tradeoffs made while implementing issue [#19](https://github.com/sh3r4rd/sh3r4rd.com/issues/19) and subtasks #34, #35, #36, #37.

Base branch: `epic/phase-4-dashboard-frontend` (off `main`). All subtask PRs target this branch.

## Decisions

### 1. Data shape — match the real API, not the issue specs (2026-05-25)

The Phase 4 issues were drafted before the Phase 3 API was finalized, and the two diverge:

| Field referenced in issues | Field actually returned by `/recruiters` (Phase 3 `AnonymizedItem`) |
|---|---|
| `firstName`, `lastName` (with client-side `anonymizeName` → "Jane S.") | _not returned_ — stripped server-side in `anonymizer.go` |
| `dateReceived` (full date) | `month` (coarsened to `YYYY-MM`) |
| — | `recruiterLabel` ("Recruiter at {Company}") |
| — | `confidence` |

**Decision:** Render the API response verbatim — use `recruiterLabel` for the recruiter column and `month` for the date column. No client-side anonymization helper; the backend already does that.

**Why:** The Phase 3 anonymization is deliberate (issue #18 AC: "API responses NEVER contain recruiter email, phone, or full name"). Sending first/last names to the client and anonymizing there would re-introduce the very PII the backend strips. The issue specs in #36 are stale — flagged but not blocking.

**Affected subtasks:**
- **#36 (RecruiterTable):** Columns become Company, Job Title, Recruiter (renders `recruiterLabel`), Month. No `anonymizeName` helper.
- **#35 (FilterBar):** Search filter applies to `company` and `jobTitle` only (no name field exists). "Job Title dropdown" still derived from `jobTitle`.
- **#34 (StatsCards):** "This Month" computed against `month` field (YYYY-MM compare), not a Date.
- **#37 (DashboardPage):** Filtered fields are `{ search, company, jobTitle, monthFrom, monthTo }` instead of `{ dateFrom, dateTo }`. Date range inputs use `month` type, not `date`.

### 2. Access — public route, no auth (2026-05-25)

`/dashboard` is publicly accessible. No auth gate. Data is fully anonymized server-side, and the existing site has no auth infrastructure. Adding auth was deemed scope creep on Phase 4.

### 3. API base URL — hardcoded (2026-05-25)

`https://api.sh3r4rd.com` hardcoded in the fetch call, matching the existing pattern in `ResumeRequestPage.jsx`. No Vite env var introduced — single-environment site, no staging.

### 4. Branch & PR strategy (2026-05-25)

- Each subtask has its own `feat/{n}-{slug}` branch off `epic/phase-4-dashboard-frontend`.
- Each PR targets `epic/phase-4-dashboard-frontend`, NOT `main`.
- #34, #35, #36 are independent components — implemented in **parallel via subagents with worktree isolation**.
- #37 is implemented sequentially after the other three (depends on importing them).
- #37's CI/build will fail until #34, #35, #36 are merged into the epic — flagged in the PR description.

### 5. Subagent / skill mapping (2026-05-25)

| Task | Tool | Why |
|---|---|---|
| Parallel build of #34, #35, #36 | `general-purpose` agent × 3 with `isolation: "worktree"` | Independent components — no shared files; agents can work in isolated trees simultaneously. |
| Final build of #37 | Done in main worktree (sequential) | Imports the other three components; easier to verify all wired up locally before splitting commits. |
| Build / lint verification | `npm run lint && npm run build` inside each worktree | Each agent verifies their own component compiles. |
| `feature-dev:code-reviewer` | Reserved for post-implementation review of #37 if it grows complex. | Optional — only if the page logic warrants a second pass. |

Linear/`linear-workflow` skill not used — this repo tracks issues on GitHub, no `.linear/` directory.

## Open items / follow-ups

_(None yet — will track here as I encounter them.)_

## Per-subtask notes

### #34 — StatsCards
_(updates as work progresses)_

### #35 — FilterBar
_(updates as work progresses)_

### #36 — RecruiterTable
_(updates as work progresses)_

### #37 — DashboardPage
_(updates as work progresses)_
