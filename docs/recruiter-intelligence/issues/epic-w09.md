**Parent**: {{ref:RI}} (Recruiter Intelligence)

Weekly epic for **week 9** of [Recruiter Intelligence](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/README.md). The full specification is [PLAN.md § Week 9](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md#week-9-the-public-job-fit-analyzer); this issue tracks status. Task issues are attached as sub-issues.

## Exam focus

D2 2.2.3, 2.4.3, 2.5.1 · D1 1.1.1, 1.2.2, 1.6.5 · D3 3.1.3, 3.4.1, 3.4.2 · D4 4.1.2, 4.1.4, 4.2.1, 4.2.4 · D5 5.1.1, 5.1.3

## Definition of done

`/fit` is live. Each spend control has been tripped deliberately in a test. The eval numbers are in the journal.

## Approval gates

A new route, an API payload (a new endpoint), Terraform, MSW handlers and fixtures, and an e2e spec. Each task issue that touches one carries `needs-approval`.

## Study

- [ ] Model cascading and routing.
- [ ] Prompt caching vs semantic caching vs edge caching.
- [ ] Latency-optimized inference.
- [ ] A/B testing prompts with Prompt Management and Prompt Flows (3.4.2).
- [ ] Transparency patterns (3.4.1).

## Exam lens

- Tiered model use by query complexity.
- Deterministic request hashing vs semantic caching: when each one is wrong.
- Graceful degradation when a model is throttled (2.4.3).

## Close-out

- [ ] Every task sub-issue is closed or cut (stretch) as not planned
- [ ] `JOURNAL.md` entry for week 9 merged with the epic PR
- [ ] Exam skills exercised ticked in `EXAM-COVERAGE.md`
- [ ] Next wave's task issues created (TRACKING.md § 7)
