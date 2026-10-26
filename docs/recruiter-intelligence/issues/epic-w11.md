**Parent**: {{ref:RI}} (Recruiter Intelligence)

Weekly epic for **week 11** of [Recruiter Intelligence](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/README.md). The full specification is [PLAN.md § Week 11](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md#week-11-continuous-evaluation); this issue tracks status. Task issues are attached as sub-issues.

## Exam focus

D5 5.1.1, 5.1.2, 5.1.4, 5.1.5, 5.1.8, 5.1.9, 5.2.1, 5.2.3, 5.2.5 · D2 2.1.5, 2.3.5, 2.5.2 · D1 1.6.4, 1.6.6 · D3 3.3.4

## Definition of done

a PR that degrades the extraction prompt fails CI. One weekly summary has gone through approval and been published.

## Approval gates

A workflow change, an IAM OIDC role, Terraform (Step Functions, Scheduler, SNS), and an API payload (the scorecard). Each task issue that touches one carries `needs-approval`.

## Study

- [ ] Bedrock evaluation types (automatic, human, LLM-as-judge, RAG).
- [ ] A/B and canary testing for FMs.
- [ ] CodePipeline and CodeBuild for GenAI (2.3.5, mostly study).
- [ ] The GenAI gateway pattern.
- [ ] **Take the official practice question set** and log every miss in `EXAM-COVERAGE.md`.

## Close-out

- [ ] Every task sub-issue is closed or cut (stretch) as not planned
- [ ] `JOURNAL.md` entry for week 11 merged with the epic PR
- [ ] Exam skills exercised ticked in `EXAM-COVERAGE.md`
- [ ] Milestone dates still accurate
