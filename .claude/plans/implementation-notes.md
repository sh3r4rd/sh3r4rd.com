# Recruiter Intelligence Week 0: Stop losing data — Implementation Notes

Running log for milestone `RI-W00`: epic [#118](https://github.com/sh3r4rd/sh3r4rd.com/issues/118) and task [#131](https://github.com/sh3r4rd/sh3r4rd.com/issues/131). Spec: `docs/recruiter-intelligence/PLAN.md` § Week 0. Work done 2026-11-20 (local) / 2026-11-21 UTC.

## Decisions you made (asked up front)

| Question | Your answer |
|---|---|
| Run the production S3 copy + Terraform apply myself? | **Yes, run both now.** (Satisfies the epic's "run it yourself or approve it explicitly" gate.) |
| Base branch, given `JOURNAL.md`/`PLAN.md` exist only on the unmerged docs PR #116 | **Stack on #116:** epic branch off `docs/recruiter-intelligence-plan` |
| Epic branch name (`feat/{n}-…` vs TRACKING.md §6 `epic/ri-week-NN-<slug>`) | **`epic/ri-week-00-stop-losing-data`** |

## Branch / PR layout

```
main
 └─ docs/recruiter-intelligence-plan   (PR #116, yours, open)
     └─ epic/ri-week-00-stop-losing-data   (epic #118 PR → docs branch; auto-retargets to main when #116 merges and its branch is deleted)
         └─ feat/131-preserve-raw-emails-inventory   (task #131 PR → epic branch)
```

- The task branch is cut from the epic branch **after** the epic's own commits, so the two PRs never edit the same lines and #131 merges into the epic without conflicts.
- **Suggested merge order:** #116 → #131's PR into the epic → the epic PR into `main`.

| Issue | Branch | PR | Base | Status |
|---|---|---|---|---|
| #131 | `feat/131-preserve-raw-emails-inventory` | #143 | `epic/ri-week-00-stop-losing-data` | open, not merged |
| #118 | `epic/ri-week-00-stop-losing-data` | #144 | `docs/recruiter-intelligence-plan` (retargets to `main` after #116) | open, not merged |

- **Auto-close caveat:** GitHub applies `Closes #N` only when a PR merges into the **default** branch. #143 merges into the epic branch, so it won't close #131 by itself. The epic PR #144 therefore carries both `Closes #118` and `Closes #131`.
- PR titles use `docs(recruiter-intelligence): …` with the issue number in parentheses, per your PR-title format.

## Production operations (all done, all verified)

| Time (UTC) | Step | Result |
|---|---|---|
| 04:30:18 | Snapshot `incoming/` key list, then `aws s3 sync s3://sh3r4rd-recruiter-emails/incoming/ …/archive/raw/ --copy-props metadata-directive` | 25 → 25 objects; key sets identical |
| 04:30–04:31 | Verified **every** copy, not just one: `get-object-tagging` + `head-object` ETag/size vs the original | 0 copies with tags; 0 ETag/size mismatches; SSE AES256 preserved |
| 04:31 | `make tf-vars-pull` (no drift), set `s3_lifecycle_expiration_days = 3650`, **targeted** `plan -out` | `Plan: 0 to add, 1 to change, 0 to destroy`, only `days 30 -> 3650` |
| 04:31:36 | `terraform apply <saved plan>` | `Apply complete! Resources: 0 added, 1 changed, 0 destroyed.` `get-bucket-lifecycle-configuration` now returns `Days: 3650` |
| 04:32 | `make tf-vars-push TFVARS_ARGS=--force`, then `make tf-vars-diff` | Pushed as a new SSM version; diff exit 0 |

### Why the apply was `-target`ed (deviation from the spec's bare `plan`/`apply`)

The untargeted plan showed **4** changes, not 1. Three were pre-existing drift unrelated to week 0:

1. `module.lambda_api_handler` and 2. `module.lambda_email_parser`: `source_code_hash` differs. An untargeted apply would have **redeployed both production Lambdas from whatever build artifacts happen to be on this laptop**.
3. `module.s3.aws_s3_bucket_server_side_encryption_configuration`: it would **remove the `blocked_encryption_types = ["SSE-C"]` block** that AWS added. That's exactly the drift issue #98 owns, and week 1 requires reconciling it before the bucket gets a KMS key. Applying it here would weaken the bucket and pre-empt #98.

So I applied `-target=module.s3.aws_s3_bucket_lifecycle_configuration.email_storage` from a saved plan file, which meets the issue's AC "the applied plan changed only the lifecycle rule" exactly. After the apply, the full plan shows just those 3 residual changes (the lifecycle change is gone from it). **They're still there for #98 / week 1 to handle.** Item 3 is #98's. Items 1–2 need a rebuild from a clean checkout of `main` (`make build-lambdas`) before any untargeted apply, so a redeploy ships `main`'s code and not this laptop's artifacts. Go builds here aren't byte-reproducible, so the hash may still differ after the rebuild; that's fine once the binaries come from `main`. This step is added to #132.

### Other operational details
- The tfvars comment was also edited: `# S3 lifecycle expiration in days for raw emails (RI week 0 stopgap: 3650; week 1 restores 30, #132)`, so whoever runs week 1 sees why it's 3650. It first cited #131 (week 0's task); the restore belongs to #132, so on 21 November 2026 the comment was corrected and pushed to SSM with `make tf-vars-push`.
- Lowering the expiry back to 30 days is a data-loss step. `archive/raw/` was a one-time sync, and `archive/clean/` doesn't backfill raw bodies, so anything received after 20 Nov exists only under `incoming/`. Re-run the same `sync … --copy-props metadata-directive` and the per-copy tag check right before the week 1 apply. This step is added to #132.
- `archive/raw/` matches no lifecycle rule (the only rule filters `incoming/`), so the copies are retained indefinitely. That's intended until week 1 adds Intelligent-Tiering on `archive/`.
- The copy didn't re-trigger the parser: SES invokes the Lambda directly, with no S3 event notification (PLAN.md; the DynamoDB item count was 92 before the copy and 92 after).
- Scratch artifacts (key lists, the saved plan, the row projection) live only in the session scratchpad and were never committed. The row projection holds only `s3_key` + `received_at`.

## Inventory (aggregates only; this is what goes in JOURNAL.md)

| Metric | Value | How |
|---|---|---|
| Raw bodies preserved (`archive/raw/`) | **25** (239,397 bytes) | `aws s3 ls --recursive --summarize` |
| DynamoDB rows (all-time extractions) | **92** | `scan --select COUNT --filter-expression 'NOT begins_with(id, :p)'` with `:p = STATS#` |
| Bodies permanently lost | **67** (exact, not approximate) | every row has an `s3_key`; 25 of the 92 keys still exist, and 67 don't |
| Surviving bodies that have no DynamoDB row | 0 | All 25 also carry `parse-status=success` |
| Oldest surviving body | received 2026-10-22 | |
| Bodies lost between plan approval (25 Oct) and the copy (20 Nov) | **about 26** (≈20 of them after the milestone's 1 Nov due date) | rows with `received_at` from 2026-09-25 up to the oldest surviving body. Day granularity; `received_at` carries mixed UTC offsets, so it's approximate |

### Deviation: "COUNT minus one" is wrong, so I used a filter
PLAN.md and #131 say to take the DynamoDB `COUNT` **minus one** for the `STATS#cache` item. But whether that item exists at any moment is unpredictable. The 5-minute window (`statsCacheDuration`) is a freshness check the api-handler does itself; it rewrites the item on every `/stats` cache miss. DynamoDB's TTL on `stats_ttl` deletes expired items in the background, often days later, so an expired item can still be counted. At inventory time it **wasn't there**: a `query id = "STATS#cache"` returned 0, and the raw count was 92, not 93. "Minus one" would have reported 91. I used `NOT begins_with(id, "STATS#")` instead, which is correct whether or not the cache item exists, and corrected PLAN.md § Week 0 step 3 in the #131 PR.

## Docs changes and why they're split this way

- **#131 PR (`feat/131-…`):** fills in the `JOURNAL.md` § "Week 0: inventory" fields (the task's AC), and corrects PLAN.md § Week 0: (a) target the apply at the lifecycle rule, (b) count with the `STATS#` filter, not minus one.
- **Epic #118 PR:** TRACKING.md §6 says to write the week's `JOURNAL.md` entry *in the epic PR*. So the epic branch carries a full week-0 retrospective (Shipped / Numbers / Decisions / Surprises / Exam skills / Spend / Cut) in the `JOURNAL.md` template format, plus these notes.
- PR titles: #131's issue title uses the type `ops:`, which isn't a conventional-commit type (no commitlint here, but semantic-release's analyzer ignores it). Both PRs change only docs, so they're `docs(recruiter-intelligence): …`, which also correctly triggers **no release**.

## Things I deliberately did NOT do (yours to decide)

- **The `EXAM-COVERAGE.md` ✓ for 3.2.2 is left unticked.** That column is a self-check: "tick a row when you can explain the skill and the alternatives." It's your call, not mine. It's listed as a close-out checkbox in the epic PR.
- **Milestones were not shifted.** TRACKING.md §2 says "When a week slips, move that milestone and every later one by the same amount." Week 0 was due 1 Nov and finished 20 Nov (about a 3-week slip). RI-W01 (due 8 Nov) and RI-W02 (due 15 Nov) are already past due with all tasks open, and RI-W03 is due 22 Nov. Re-dating 13 milestones is a planning decision, so I flagged it rather than doing it.
- **Issue bodies weren't edited** (checkbox ticking, etc.). The PRs carry `Closes #131` and the verification evidence. I did remove the `needs-approval` label from #131, because TRACKING.md §3 says it's "removed once approved" and you approved it in this session.
- **The 3 residual plan diffs were not touched** (see above).
- **CI: the E2E job fails on every PR right now, unrelated to these docs-only PRs.** The `test.yml` E2E job dies at "Install Playwright" with `ERROR: Playwright does not support chromium on ubuntu26.04-x64`. GitHub's `ubuntu-latest` image moved to Ubuntu 26.04 after the last green run (26 Oct). The fix is either pinning `runs-on: ubuntu-24.04` in `test.yml` or bumping `@playwright/test` to a version that supports 26.04. Not done in the RI-W00 PRs because modifying `.github/workflows/` is an AGENTS.md ask-first item and outside RI-W00. Lint/build, unit tests and Terraform validate all pass.
  - **Update: fixed in PR #145**, merged to `main` on 21 Nov 2026 as `ci: bump Playwright to 1.65.1 for Ubuntu 26.04 runners`, which you asked for as a separate PR. The merged fix bumps `@playwright/test` to `^1.65.1` (the first version here with ubuntu26.04 Chromium downloads), keeps both `test.yml` jobs on `ubuntu-latest`, and adds the runner image (`$ImageOS`) to the Playwright browser cache key, so an image change misses the cache instead of reusing browsers downloaded on the old image. The PR's first commit pinned both jobs to `ubuntu-24.04` as a stopgap (green: run 44867376824); its last commit removed the pin. No issue number, so the branch uses a `fix/` slug, not `feat/{n}-…`.
  - **Since #145 merged:** `main` and this epic branch (rebased onto `main`) both include it. #143 merged into the epic branch, and #144's E2E check passes on `ubuntu-latest` (run 44873750416).
- The pre-existing CLAUDE.md/AGENTS.md naming mismatch was **not fixed**: the docs call the table `recruiter_emails`, but the real table (and tfvars) is `recruiter-emails`. It's harmless but worth a one-line fix someday.

## Close-out status for epic #118

- [ ] Every task sub-issue closed or cut: #131 is the only task; it closes when its PR merges
- [ ] `JOURNAL.md` week-0 entry merged with the epic PR (in the epic PR, pending review)
- [ ] Exam skills exercised ticked in `EXAM-COVERAGE.md`: **yours** (3.2.2), see above
- [x] Next wave's task issues created: end of week 0 → week 2's tasks, which already exist (#138–#141, created in wave 1)

---

# Phase 5: Testing & CI Pipeline — Implementation Notes

Running log for issue [#20](https://github.com/sh3r4rd/sh3r4rd.com/issues/20) and subtasks #38, #39, #40, #41.

Base branch: `epic/phase-5-testing-and-ci-pipeline` (off `main`, which already contains the merged Phase 4 dashboard — `5820b5f`). All subtask PRs target this epic branch.

## Issue tracker

The prompt referenced "subtasks in Linear," but this repo tracks work on **GitHub Issues** (`sh3r4rd/sh3r4rd.com`) and has no `.linear/` directory — same as Phase 4. Proceeding with GitHub. The four subtasks are GitHub issues #38 (5.1), #39 (5.2), #40 (5.3), #41 (5.4).

## Dependency graph & parallelism

```
#38 (vitest/MSW/fixtures) ──┐
                            ├──► #39 (component+integration tests, needs #38)
#40 (Playwright E2E) ───────┤            │
  (independent of #38/#39)  │            │
                            └────────────┴──► #41 (CI workflow, runs all of the above)
```

- **#38 and #40 are independent** and run in parallel: #40 uses Playwright `page.route()` mocking — it needs neither Vitest nor MSW, only the Phase 4 dashboard (already in `main`). Built #40 via a **worktree-isolated background subagent** while #38 was built in the main worktree.
- **#39 depends on #38** (needs the Vitest config + MSW infra to run). Branched `feat/39-...` **off `feat/38-...`** (stacked PR). Its PR base is `feat/38-...`; GitHub auto-retargets it to the epic when #38 merges, keeping the #39 diff clean (only the test files, not #38's infra).
- **#41 depends on #38, #39, #40** (its workflow runs `test:coverage` and `test:e2e`, which those PRs add). Branched off the epic; its `test.yml` will only go fully green once the sibling PRs merge into the epic and the scripts exist there. Flagged in the PR body. (Same pattern Phase 4 used for the dependent DashboardPage PR.)

## Agent / skill mapping

| Task | Tool | Why |
|---|---|---|
| #40 Playwright E2E | `general-purpose` agent, `isolation: "worktree"`, background | Fully independent of #38/#39; isolated worktree lets it run concurrently with #38 without file/lock contention. |
| #38, #39, #41 | Main worktree, hands-on | #38 is foundational (everything depends on it — verify carefully); #39 stacks on it; #41 is small YAML/shell glue. |
| Verification | `npm run test` / `test:coverage` after each test PR; `npx playwright test` for #40 | Each PR proves its own tests pass before opening. |
| `code-review` skill | Optional post-implementation pass | Reserved if a test file grows complex. |

Linear/`linear-workflow` skill not used (no `.linear/` dir — see above).

## API response shapes (ground truth for fixtures/mocks)

Confirmed against `infra/recruiter-dashboard/lambda-src/api-handler`:

- **`GET /recruiters`** → array of `AnonymizedItem`: `{ id, company, jobTitle, month: "YYYY-MM", recruiterLabel: "Recruiter at {company}", confidence: float }`. **No PII** (anonymized server-side).
- **`GET /stats`** → `{ totalEmails: int, uniqueCompanies: int, byMonth: {"YYYY-MM": int}, topJobTitles: {"title": int} }`.

`DashboardPage` fetches **both** endpoints. Issue #38 names only `/recruiters`, but MSW must also mock `/stats` or the integration tests' stats assertions break — **added a `/stats` handler too** (decision; superset of the issue spec).

## Decisions / deviations (Phase 5)

### #38 — Vitest / MSW / fixtures (PR #86)
- **Mocked `/stats` in addition to `/recruiters`.** The issue named only `/recruiters`, but `DashboardPage` fetches both; the #39 integration tests assert stats totals.
- **Coverage `include` scoped to the 6 dashboard files** (DashboardPage, StatsCards, FilterBar, RecruiterTable, card, filters). Including the pre-existing non-dashboard pages — which have zero Phase 5 tests — would make the "75% overall" AC unachievable in scope. Documented in `vitest.config.js`. *(Superseded by PR #89: replaced with `include: src/**` + a denylist so new files are gated by default.)*
- **`passWithNoTests: true`** so `npm run test` exits 0 with zero tests (an explicit #38 AC). A fixture/MSW smoke test (`src/mocks/__tests__/mocks.test.js`) is included regardless and validates the other #38 ACs.
- **Vitest `include` scoped to `src/**`** so it never picks up Playwright specs in `e2e/` (a Playwright spec run under Vitest throws) or specs inside agent worktrees under `.claude/`.
- **ESLint override** added for `**/*.{test,spec}.{js,jsx}` + `src/mocks/**` exposing Vitest + Node globals (the base config is browser-only). `coverage/` added to ESLint `globalIgnores` (the v8 HTML report ships JS with its own eslint-disable directives, which ESLint 9 flags by default). *(Superseded by PR #89: anchored to `coverage/**`.)*

### #39 — Component & integration tests (PR #87)
- **Stacked on #38**, not the epic: it needs the test infra to run. PR base is `feat/38-...`; GitHub auto-retargets it to the epic when #38 merges, so the #39 diff stays limited to the test files.
- Final coverage on the dashboard surface: **96% lines / 80% branches** (threshold 75%). 43 tests.
- "This Month" in StatsCards depends on `new Date()`; the test pins it with `vi.setSystemTime` for determinism. *(Hardened in PR #89: now also forces a non-UTC `TZ` and a month-boundary instant so it guards UTC-vs-local, not just determinism.)*

### #40 — Playwright E2E (PR #85)
- Built by a **worktree-isolated background subagent**, in parallel with #38.
- **`test:e2e` uses `--config e2e/playwright.config.js`** (config lives in `e2e/`, not repo root) — deviates from the issue's bare `playwright test`.
- Targeted **ESLint `e2e/**` Node-globals override** instead of disabling rules globally.
- Lockfile resolved `@playwright/test` to `1.60.0` (satisfies the issue's `^1.52.0`).
- **Subagent verification gap, then fixed by me:** the subagent's environment denied the Playwright browser download, so it shipped the suite **unrun**. I installed Chromium and ran it — 4 mobile-chrome tests failed because the desktop `<table>` is `display:none` below the `md` breakpoint (so cell/row assertions and sortable headers don't exist on mobile) and one card assertion latched onto the hidden `<option>` of the same company name in the dropdown. Fixed (commit `d34e01b`): viewport-guarded the desktop cell assertions, skipped the sort journey on mobile, and `.filter({ visible: true })` on the card assertion. Final: **18 passed, 2 skipped** across both projects. → **Lesson:** delegated test work must be executed before it's trusted; an agent that can't run the tests is writing them blind. *(Superseded by PR #89: the `if (isDesktop)`-guarded filter tests are now `test.skip` desktop-only with dedicated mobile coverage; final 17 passed, 5 skipped.)*

### #41 — CI workflow (PR #88)
- **Trigger is `push:[main]` + `pull_request:` (all PRs)**, mirroring `ci.yml` — the issue said "PRs to main", but an unfiltered PR trigger is what lets the workflow validate this epic's own subtask PRs (a `main`-only filter would skip every PR here).
- **Per-job path gating** for `tf-validate` via `dorny/paths-filter@v3` (a `changes` job output); `on.*.paths` would gate the whole workflow instead of one job. *(Superseded by PR #89: the `tf-validate`/`changes` jobs were removed — `ci.yml` already validates Terraform unconditionally.)*
- Playwright report uploaded **`if: failure()`** (per AC); coverage uploaded unconditionally.
- **Overlap with `ci.yml`** (lint/build/tf-validate appear in both) is intentional for now; consolidating the two workflows is a sensible follow-up, out of scope here. *(Done in PR #89: `test.yml` now owns only unit + e2e; lint/build/tf-validate live solely in `ci.yml`.)*
- Branched off the epic (depends on #38/#39/#40 for the scripts/tests). Its `unit-tests`/`e2e-tests` jobs stay red on its own PR until the siblings merge into the epic — same pattern Phase 4 used for its dependent PR.

### Post-merge review fixes (epic PR #89, 2026-05-30)

After the four subtask PRs merged into the epic, ran `/code-review` (high effort) on the epic→`main` PR #89 and applied the actionable findings via `/pr-fix`. These **supersede** several decisions above; the bullets they replace are marked *(Superseded by PR #89)* inline. Two follow-up commits on the epic branch:

- `62bf96f` `test(frontend): strengthen dashboard tests and coverage scope`
- `a80606c` `ci: dedupe test workflow against ci.yml and cache Playwright`

**Coverage scope — allowlist → denylist.** `vitest.config.js` now uses `include: ['src/**/*.{js,jsx}']` plus an `exclude` of the pre-Phase-5 untested surface (`App.jsx`, `main.jsx`, `components/layout/**`, `SkillsSection.jsx`, `button.jsx`, the three non-dashboard pages) and test infra (`src/mocks/**`, `__tests__`, spec files). Same covered surface and same 96/80.71 numbers as the old 6-file allowlist, but **new files are now gated by default** instead of silently exempt. Narrow the exclude list as the legacy surface gains tests.

**Test correctness.**
- `StatsCards` "This Month" test: the old `2026-05-15T12:00:00Z` (noon UTC) instant resolved to the same month in every timezone, so it passed even against a local-time regression. Now pins `2026-06-01T03:00:00Z` under `process.env.TZ = 'America/Los_Angeles'` (UTC = Jun, local = May 31) with distinct bucket values, so it genuinely guards the UTC computation.
- `FilterBar` month-range test: `userEvent.type` into a native `<input type="month">` is jsdom-version-fragile (segmented control); switched to `fireEvent.change`.
- `fixtures.js` `buildStats`: caps `topJobTitles` at the top 10 by count to mirror the api-handler's `topN(jobTitles, 10)`, so the fixture can't emit a `/stats` shape the real endpoint never returns.

**E2E (`e2e/dashboard.spec.js`).** The "Recruiter Dashboard" describe ran under **both** projects, so its row-cell filter assertions were count-only near-no-ops on mobile-chrome. The two filter tests are now `test.skip(!isDesktop(page))` (desktop-only), with explicit mobile filter coverage added to the Mobile describe. Also: `recruiterLabel` fixtures use the real `"Recruiter at {company}"` shape (was bare names) so the suite exercises the real contract — which forced the cell assertions to `{ exact: true }` (otherwise `"Initech"` matches the `"Recruiter at Initech"` labels too: 8 cells, not 4). `MD_BREAKPOINT` constant replaces the duplicated `768` literal. Final: **17 passed, 5 skipped**.

**CI — `test.yml` deduped against `ci.yml`.** Removed the `lint` job and the `changes`/`tf-validate` jobs: `ci.yml` already runs lint and validates Terraform unconditionally on the same triggers, and the path-gated `tf-validate` was strictly more fragile (a hardened token could 403 on `dorny/paths-filter` and silently skip while reporting green). `test.yml` now owns only `unit-tests` + `e2e-tests`, run **in parallel** (dropped the `lint → unit → e2e` `needs` chain), with the Playwright Chromium download **cached** (only OS deps reinstalled on a hit). The now-orphaned `scripts/validate.sh` (its only caller was the deleted job) was removed; `ci.yml` keeps its inline Terraform steps. `coverage` ignore anchored to `coverage/` (`.gitignore`) and `coverage/**` (ESLint). → **Net:** each check now runs once, in the workflow that owns it; the two workflows no longer overlap.

## PR map

| Subtask | Branch | PR | Base | Status |
|---|---|---|---|---|
| #38 | `feat/38-setup-vitest-msw-fixtures` | #86 | epic | open, not merged |
| #39 | `feat/39-component-integration-tests` | #87 | `feat/38-...` (auto-retargets to epic) | open, not merged |
| #40 | `feat/40-playwright-e2e-tests` | #85 | epic | open, not merged |
| #41 | `feat/41-ci-test-workflow` | #88 | epic | open, not merged |

**Suggested merge order:** #40 (independent) and #38 first, then #39 (after #38), then #41 last (so its CI is green on the epic).

---

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
