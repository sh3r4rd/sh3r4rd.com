**Parent**: {{ref:RI}} (Recruiter Intelligence)

Weekly epic for **week 5** of [Recruiter Intelligence](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/README.md). The full specification is [PLAN.md § Week 5](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md#week-5-retrieval-quality); this issue tracks status. Task issues are attached as sub-issues.

## Exam focus

D1 1.5.4, 1.5.5, 1.5.6, 1.4.3 (study) · D4 4.2.2, 4.3.5 · D5 5.1.6, 5.2.4

## Definition of done

one retrieval configuration is chosen, each component's contribution is measured, and nothing is left running that you aren't using.

## Approval gates

Terraform (the Aurora cluster and the Managed KB; both temporary, but both go through `plan`). Each task issue that touches one carries `needs-approval`.

## Study

- [ ] Hybrid search and reranking.
- [ ] Query transformation.
- [ ] OpenSearch Service vector search: sharding, multiple indexes, the Neural plugin (1.4.3 is
  study-only here).
- [ ] Amazon Kendra vs Knowledge Bases.

## Exam lens

- Why hybrid search helps on exact tokens (product names, levels).
- Where reranking sits and what it costs in latency.
- Query decomposition with Lambda or Step Functions vs the KB's built-in options.

## Close-out

- [ ] Every task sub-issue is closed or cut (stretch) as not planned
- [ ] `JOURNAL.md` entry for week 5 merged with the epic PR
- [ ] Exam skills exercised ticked in `EXAM-COVERAGE.md`
- [ ] Next wave's task issues created (TRACKING.md § 7)
