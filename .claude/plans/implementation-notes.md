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

- **PR #83 will not pass CI until PRs #80, #81, #82 merge into the epic.** Its imports of the three sibling components only resolve after those land. Plan: rebase #83 once the three components are in.
- Worktrees from the parallel agent runs at `.claude/worktrees/agent-*` were locked by the agent harness when I tried to clean them up. They auto-clean on agent exit. They're now gitignored and eslint-ignored so they shouldn't interfere.
- The existing `src/pages/ResumeRequestPage.jsx` inputs use `text-black` (no dark-mode variant). Out of scope for Phase 4 but worth a small follow-up — Phase 4 inputs (FilterBar) use the dark-aware pattern `bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100`.

## Infrastructure pushed directly to the epic branch

Three small commits went straight to `epic/phase-4-dashboard-frontend` (not as subtask PRs) because they're cross-cutting setup, not feature work:

1. `chore(plans): add Phase 4 implementation notes and ignore agent worktrees` — adds this file and gitignores `.claude/worktrees/`.
2. `chore(lint): ignore .claude/ in eslint global ignores` — keeps lint from scanning agent worktree dist bundles.

These don't belong in any subtask PR's scope.

## Per-subtask notes

### #34 — StatsCards → PR #80
- Implemented by parallel subagent in worktree isolation. Lint + build clean.
- Subagent deviation worth knowing: it initially passed icons as `Icon` props and rendered `<Icon />`, but the project's eslint config (`no-unused-vars` with `varsIgnorePattern: '^[A-Z_]'`) doesn't have a React plugin to recognize JSX usage of capitalized identifiers, so `Icon` was flagged unused. Refactor: store pre-rendered `<Mail .../>` elements directly in the card config. Same visual output.

### #35 — FilterBar → PR #82
- Implemented by parallel subagent in worktree isolation. Lint + build clean.
- Date range reinterpreted as month range (`<input type="month">`) because the API exposes only `month: "YYYY-MM"`. See top-of-file decision #1.
- Two `useEffect`s for the debounce: one for local→parent push (200ms), one to sync local state back from parent when filters are cleared externally. Both have a targeted `eslint-disable-next-line react-hooks/exhaustive-deps` with a one-line rationale.

### #36 — RecruiterTable → PR #81
- Implemented by parallel subagent in worktree isolation. Lint + build clean.
- Dropped the spec's `anonymizeName(firstName, lastName)` helper — API already returns `recruiterLabel` in anonymized form. See top-of-file decision #1.
- Default sort: `month` descending. Sort state lives inside the component; pagination state is owned by the parent (per the prop contract).
- Optional mobile sort `<select>` from the spec was skipped (spec said optional).
- Defensive clamp on `currentPage` when filtered data shrinks below the current page.

### #37 — DashboardPage → PR #83
- Implemented sequentially in the main worktree after #34/#35/#36 were on their branches.
- Local verification approach: temporarily checked the three components into the working tree via `git show feat/34-stats-cards:... > ...` (one per sibling), ran `npm run lint && npm run build && npm run dev`, curled `/dashboard` (got HTTP 200 + SPA shell), then `rm`'d the temp files before staging the actual commit. Final commit has only `src/pages/DashboardPage.jsx` + `src/App.jsx`.
- Did not test against the real API end-to-end. Loading/error states verified by visual code review only; live-data flows will need verification after the components merge into the epic and the page builds with real imports.
- Filter logic decisions:
  - Search is a `toLowerCase().includes(...)` substring match across the concatenation of `company + " " + jobTitle` (no name field exists in the API).
  - Company / jobTitle dropdowns do exact equality on `item.company === filters.company` (the dropdown values come from the dataset itself, so exact equality is correct).
  - Month range: lexical comparison on `YYYY-MM` strings — correct because the format sorts naturally.
- `EMPTY_FILTERS` constant + `setFilters(EMPTY_FILTERS)` in the clear path; `setPage(1)` is called by the parent's `handleFilterChange` so the FilterBar doesn't need to know about pagination.
