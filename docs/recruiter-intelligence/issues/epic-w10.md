**Parent**: {{ref:RI}} (Recruiter Intelligence)

Weekly epic for **week 10** of [Recruiter Intelligence](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/README.md). The full specification is [PLAN.md § Week 10](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md#week-10-operate-it); this issue tracks status. Task issues are attached as sub-issues.

## Exam focus

D4 4.1.1, 4.1.3, 4.2.3, 4.2.5, 4.2.6, 4.3.1–4.3.6 · D2 2.5.6 · D5 5.1.9, 5.2.2

## Definition of done

one dashboard answers "what did this cost yesterday, and was it any good?" The batch job has backfilled every archived email.

## Approval gates

Terraform (dashboard, alarms, X-Ray), and Lambda changes. Each task issue that touches one carries `needs-approval`.

## Study

- [ ] Provisioned throughput (when it pays; study-only).
- [ ] Bedrock model invocation logs.
- [ ] Output diffing and reasoning-path tracing (4.3.6).
- [ ] Capacity planning for token workloads.

## Exam lens

- Batch vs on-demand vs provisioned throughput.
- Token-burst anomaly detection.
- Golden-dataset hallucination monitoring in production.

## Close-out

- [ ] Every task sub-issue is closed or cut (stretch) as not planned
- [ ] `JOURNAL.md` entry for week 10 merged with the epic PR
- [ ] Exam skills exercised ticked in `EXAM-COVERAGE.md`
- [ ] Next wave's task issues created (TRACKING.md § 7)
