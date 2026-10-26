# Tracking: GitHub Project and issues

How the 12-week plan in [`PLAN.md`](PLAN.md) is tracked on GitHub. **Status: approved
25 Oct 2026. The manifest, issue bodies and script exist and have been dry-run; nothing has been
created on GitHub yet.** See § Next steps.

## Decisions (approved 25 Oct 2026)

| # | Decision |
|---|---|
| 1 | **Week 1 starts Monday 2 November 2026.** Week 0 is the week of 26 October |
| 2 | **About 50 PR-sized task issues.** One issue = one PR, not one per `PLAN.md` checkbox |
| 3 | **Rolling waves:** create the scaffolding and weeks 0–2 now, then each week's tasks two weeks ahead |
| 4 | **Absorb #59 and #98:** #59 becomes a week-10 task, and #98 blocks week 1's archive issue |

## Conventions this keeps

The repository already tracks work in GitHub Issues: epic issues (#17, #19, #20) with subtask
bodies that start `**Parent**: #N`, `epic/phase-N-…` branches, and `feat/{n}-{slug}` PRs into the
epic (`.claude/plans/implementation-notes.md`). Semantic-release adds a `released` label. Before
this plan there were no milestones, Projects, issue templates or native sub-issues.

## 1. Structure

```
Initiative issue: "Recruiter Intelligence" (1)
 └─ Weekly epics: "RI Week N: <theme>" (13: weeks 0–12), one per milestone
     └─ Task issues: one per PR-sized deliverable (~3–5 per week, ~50 total)
```

- **Native sub-issues** link initiative → epics → tasks. Each body also keeps a
  `**Parent**: #N` first line for continuity with older issues.
- **Stretch items** become task issues labelled `stretch`, in the same milestone. Cutting one
  means closing it as "not planned".
- **Study items are not issues.** They are a checklist in each weekly epic's body. Exam progress
  lives in [`EXAM-COVERAGE.md`](EXAM-COVERAGE.md)'s ✓ column.

## 2. Milestones are the weeks

Each milestone is due on the Sunday that ends its week. The Project groups by milestone, and
there is deliberately no iteration field, so dates can't drift between two places.

| Milestone | Theme | Due |
|---|---|---|
| `RI-W00` | Stop losing data | Sun 1 Nov 2026 |
| `RI-W01` | Data you keep, spend you see | Sun 8 Nov 2026 |
| `RI-W02` | Evals first | Sun 15 Nov 2026 |
| `RI-W03` | Extraction v2 on Bedrock | Sun 22 Nov 2026 |
| `RI-W04` | RAG ingestion | Sun 29 Nov 2026 |
| `RI-W05` | Retrieval quality | Sun 6 Dec 2026 |
| `RI-W06` | Agent v1 | Sun 13 Dec 2026 |
| `RI-W07` | Private assistant live | Sun 20 Dec 2026 |
| `RI-W08` | Guardrails and security | Sun 27 Dec 2026 |
| `RI-W09` | Public Job-Fit Analyzer | Sun 3 Jan 2027 |
| `RI-W10` | Operate it | Sun 10 Jan 2027 |
| `RI-W11` | Continuous evaluation | Sun 17 Jan 2027 |
| `RI-W12` | Ship and sit | Sun 24 Jan 2027 |

Thanksgiving (week 4) and the holidays (weeks 8–9) fall inside this window. **When a week
slips, move that milestone and every later one by the same amount.** Never quietly overlap two
weeks.

## 3. Labels

**Reused:** `backend`, `frontend`, `infrastructure`, `terraform`, `lambda`, `api`, `testing`,
`e2e`, `ci`, `documentation`, `performance`, `recruiter-dashboard`.

**New:**

| Label | Colour | Purpose |
|---|---|---|
| `recruiter-intelligence` | `#5319E7` | Every issue in the initiative. The Project's auto-add rule keys on it |
| `epic` | `#3E4B9E` | The 13 weekly epics, plus the initiative issue |
| `aip:d1` | `#C5DEF5` | Exam D1: FM integration, data, compliance |
| `aip:d2` | `#C5DEF5` | Exam D2: implementation and integration |
| `aip:d3` | `#C5DEF5` | Exam D3: safety, security, governance |
| `aip:d4` | `#C5DEF5` | Exam D4: operational efficiency |
| `aip:d5` | `#C5DEF5` | Exam D5: testing, validation, troubleshooting |
| `eval` | `#BFD4F2` | Eval harness, datasets, scorecards |
| `security` | `#B60205` | Guardrails, IAM, PII. The repo had no such label |
| `stretch` | `#EDEDED` | Optional, and can be cut |
| `needs-approval` | `#FBCA04` | Touches an ask-first item in `AGENTS.md:152-160`. Removed once approved |

## 4. The Project

Name it **"Recruiter Intelligence"**, owned by the `sh3r4rd` user account, linked to
`sh3r4rd/sh3r4rd.com`, with **public** visibility (the repo and its issues are public anyway).

**Fields**
- **Status:** Todo / In progress / In review / Blocked / Done
- **Estimate (h):** number. The weekly sum should be about 6 for build work
- **Exam skills:** text, such as `1.5.4, 4.2.2`
- **Kind:** single select (Build / Eval / Ops / Docs)

**Views**
1. **Board:** by Status, filtered to the current milestone
2. **Plan:** a table grouped by milestone, with Estimate summed
3. **Approvals:** `label:needs-approval is:open`
4. **Exam:** grouped by label, filtered to `aip:*`
5. **Roadmap:** by milestone due date

**Built-in workflows**
- Auto-add from `sh3r4rd/sh3r4rd.com` with filter `is:issue label:recruiter-intelligence`.
  The free plan allows one auto-add rule, which is enough.
- Item added → Todo.
- Item closed, or PR merged → Done.

## 5. Issue template

Add `.github/ISSUE_TEMPLATE/recruiter-intelligence-task.yml`, an issue form with these fields:
- **Parent epic** (`#N`)
- **Week**
- **Exam skills**
- **Context**, with links to the `PLAN.md` section and any `DECISIONS.md` entries
- **Tasks**
- **Acceptance criteria**: the slice of the week's definition of done this issue delivers
- **Approval gates**: checkboxes, one per ask-first item touched
- **Eval evidence**: the number that must appear in the PR description

This follows #59's Problem / Solution / Acceptance Criteria shape. The form applies the
`recruiter-intelligence` label automatically.

## 6. Branches and PRs

- One epic branch per week: `epic/ri-week-NN-<slug>`, off `main`.
- Task PRs `feat/{n}-{slug}` (or `fix/`, `docs/`, `chore/`) → the epic branch, with
  `Closes #n` in the body.
- An epic PR → `main` at the end of the week. Run `/code-review` there, and write the week's
  `JOURNAL.md` entry in that PR.
- Commits stay conventional, scoped by area (`feat(email-parser): …`, `feat(infra): …`).

## 7. Waves, the manifest and the script

| When | Create |
|---|---|
| **Wave 1 (now)** | Labels, 13 milestones, the Project, the issue template, the initiative issue, all 13 weekly epics, and the task issues for **weeks 0–2**. Attach #59 and #98 (§ 8) |
| **End of week N** | The task issues for week N+2 |

**Manifest:** `docs/recruiter-intelligence/issues.json` is version-controlled, with one entry
per issue: key, title, milestone, labels, parent key, blocked-by, estimate, exam skills, kind and
`body_file`. It also lists the labels, the milestones, the edits to existing issues (`existing`)
and the Project's fields. JSON keeps the script free of dependencies.

**Bodies:** one markdown file per issue in `docs/recruiter-intelligence/issues/`.
`{{ref:KEY}}` placeholders become issue numbers at creation. The 13 epic bodies were generated
once from `PLAN.md` (exam focus, definition of done, approval gates, study checklist, exam lens),
so **edit an epic's body file if `PLAN.md` changes before that epic exists**, and edit the live
issue after.

**Script:** `scripts/ri-issues.sh` (bash 3.2-compatible, because that is what macOS ships;
`gh` + `jq`). Its phases are `labels`, `milestones`, `issues`, `links`, `existing` and
`project`. Use `--phase <name>` to run one.
- `--dry-run` is the **default** and prints every action.
- `--apply` performs them.
- It is **idempotent**: labels and milestones are matched by name and issues by exact title, so
  a re-run creates nothing twice.
- It writes each created issue's number back to the manifest, so parent and blocked-by
  references resolve.
- Sub-issues go through `POST /repos/{owner}/{repo}/issues/{number}/sub_issues` with the child's
  numeric issue **id**, not its number.
- Moving an issue that already has a parent (#59 is a native sub-issue of the closed #17)
  uses that endpoint's `replace_parent=true`.
- Blocked-by uses GitHub's issue-dependency API
  (`POST /repos/{owner}/{repo}/issues/{number}/dependencies/blocked_by` with the blocker's
  numeric `issue_id`). Verified available on 25 Oct 2026. Bodies also carry a
  `**Blocked by:**` line for readers.
- **The Project is created through GraphQL** (`createProjectV2` with `repositoryId`,
  `createProjectV2Field`, `addProjectV2ItemById`, `updateProjectV2ItemFieldValue`). The
  `gh project` subcommands also demand `read:org`, which the token doesn't have.
- **Views, extra Status options and built-in workflows have no API.** The script prints them as
  manual steps for the web UI.

**Keeping one source of truth:** `PLAN.md` stays the specification, and issues become the source
of truth for **status**. When a week's task issues exist, replace each `- [ ]` in that week of
`PLAN.md` with a link to its issue.

## 8. Existing issues

| Issue | Action |
|---|---|
| **#59** emit `ParseFailures` metric | Becomes a sub-issue of the **RI Week 10** epic, with the `recruiter-intelligence`, `aip:d4` and `needs-approval` labels and milestone `RI-W10`. Edit its first line from `**Parent**: #17` to the Week 10 epic. Its acceptance criteria stand, plus: emit through embedded metric format (EMF) if week 10 standardizes on it, in which case the `PutMetricData` IAM grant isn't needed |
| **#98** S3 encryption drift (SSE-C) | Becomes a sub-issue of the **RI Week 1** epic, milestone `RI-W01`, and **blocks** the week-1 archive issue (W01-1 below), so drift is reconciled before the bucket gets a KMS key |
| #34–#37, #40, #43, #44 | Out of scope: open, but apparently delivered or stale. Triage separately |

## 9. Ground rules for issue content (the repo is public)

- No real email content, recruiter names or addresses, and no company tied to a person.
- Corpus sizes and eval results as aggregates only, the same rule as `JOURNAL.md`.
- Link to the docs rather than pasting sensitive detail.

## 10. Wave 1 task issues (weeks 0–2)

Keys are manifest keys. `NA` means the issue carries `needs-approval`. Every issue also carries
`recruiter-intelligence`.

| Key | Title | Labels | Est. h | Skills |
|---|---|---|---|---|
| W00-1 | ops: preserve raw emails before the 30-day expiry and record the inventory | infrastructure, terraform, security, aip:d3, NA | 1 | 3.2.2 |
| W01-1 | feat(email-parser): durable `archive/clean` with KMS, Intelligent-Tiering, and `incoming/` back to 30 days | backend, lambda, infrastructure, terraform, security, aip:d1, aip:d3, NA | 3 | 1.3.1, 3.2.2 |
| W01-2 | feat(email-parser): backfill command that re-parses `archive/raw` into `archive/clean` | backend, aip:d1 | 1.5 | 1.3.1 |
| W01-3 | feat(evals): data-quality report over the archive (aggregates only) | eval, aip:d1, NA (Python toolchain, D-03) | 1.5 | 1.3.1 |
| W01-4 | feat(infra): AWS Budgets and Cost Anomaly Detection; confirm week-3 models are invocable | infrastructure, terraform, aip:d4, NA | 1 | 4.3.2 |
| W01-5 | chore: one-time Macie sensitive-data discovery over the archive | security, aip:d3, stretch | 1 | 3.2.2 |
| W01-6 | docs: Well-Architected Generative AI Lens review | documentation, aip:d1, stretch | 1 | 1.1.3 |
| W02-1 | feat(evals): extraction schema v2 and the committed synthetic dataset | eval, aip:d1, aip:d5 | 2 | 1.3.3, 5.1.4 |
| W02-2 | feat(evals): eval harness (scorers, calibration, reports) and `make eval-extraction` | eval, testing, aip:d5, NA (Makefile, Python) | 2.5 | 5.1.4, 4.3.6 |
| W02-3 | feat(email-parser): `cmd/extract`, which runs the real parser and extractor over an `.eml` | backend, aip:d5 | 1 | 5.1.4 |
| W02-4 | chore(evals): private real golden set and the recorded OpenAI baseline | eval, aip:d5 | 1.5 | 5.1.3, 1.6.4 |

**Dependencies:** W01-1 ← #98. W01-2 ← W01-1. W01-3 ← W01-2. W02-2 ← W02-1. W02-4 ← W02-2 and
W02-3.

Each weekly epic's body holds the theme, a link to its `PLAN.md` section, the definition of done,
the study checklist and exam lens, and a summary of its approval gates.

## Next steps

1. **Commit the plan docs.** They are uncommitted on `main` in the working tree. Branch
   `docs/recruiter-intelligence-plan`, commit `docs/recruiter-intelligence/` (including this file),
   and open a PR to `main`.
2. **Write the issue template, `issues.json` and `scripts/ri-issues.sh`** in the same PR or a
   follow-up.
3. **Show the dry run** and wait for approval. Nothing is created on GitHub until then.
4. **Run `scripts/ri-issues.sh --apply`:** labels → milestones → the initiative issue →
   epics → wave-1 tasks → sub-issue links → dependencies → #59 and #98 → the Project, its
   fields and items. Then commit the issue numbers it wrote into `issues.json`.
5. **Finish the Project in the web UI:** the views, the two extra Status options and the
   built-in workflows the script lists.
6. **Replace week 0–2 checkboxes in `PLAN.md`** with links to their issues.

### Token

Resolved 25 Oct 2026: `GITHUB_TOKEN` now carries `repo` and `project`, which the script needs.
It lacks `read:org`, so the `gh project` CLI subcommands fail; the script doesn't use them.
