A 12-week GenAI project built on the Recruiter Dashboard, run alongside studying for the **AWS Certified Generative AI Developer – Professional (AIP-C01)** exam. Week 1 starts **Monday 2 November 2026**; week 12 ends **Sunday 24 January 2027**.

## What ships

1. **Job-Fit Analyzer** (public, `/fit`): paste a job description, get a structured, cited fit analysis.
2. **Market Insights** (public, on `/dashboard`): enriched-extraction trends and a human-approved weekly AI summary.
3. **Inbox Assistant** (private, Cognito): RAG + agent over the recruiter corpus, also exposed as an MCP server.
4. A public **eval scorecard** and a **case study** page.

## Documents

| | |
|---|---|
| [README](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/README.md) | Goal, architecture, starting-point gaps |
| [PLAN](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md) | Week-by-week specification |
| [EXAM-COVERAGE](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/EXAM-COVERAGE.md) | All 98 exam skills mapped to weeks |
| [DECISIONS](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/DECISIONS.md) | Architecture decisions |
| [TRACKING](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/TRACKING.md) | How this is tracked: milestones `RI-W00`–`RI-W12`, labels, the Project, waves |
| [JOURNAL](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/JOURNAL.md) | Weekly log |

## How to read the tracking

- The 13 weekly epics are sub-issues of this issue; each week's task issues are sub-issues of its epic.
- `needs-approval` marks work that touches an `AGENTS.md` ask-first item.
- Task issues are created in rolling waves, two weeks ahead.

**This repository is public.** Issues carry no email content, no recruiter identities, and only aggregate numbers.
