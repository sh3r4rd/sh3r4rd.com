**Parent**: {{ref:RI}} (Recruiter Intelligence)

Weekly epic for **week 8** of [Recruiter Intelligence](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/README.md). The full specification is [PLAN.md § Week 8](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md#week-8-guardrails-and-security); this issue tracks status. Task issues are attached as sub-issues.

## Exam focus

D3 3.1.1–3.1.5, 3.2.1 (study), 3.2.2, 3.2.3, 3.3.1, 3.3.2, 3.4.3

## Definition of done

the red-team suite runs in the eval harness, and its before-and-after injection success rate is in the journal. Guardrail interventions are logged as a metric.

## Approval gates

Terraform (guardrail, logging bucket, CloudTrail), IAM, and an email-parser change (the tagger). Each task issue that touches one carries `needs-approval`.

## Study

- [ ] Every Guardrails policy type and the `ApplyGuardrail` API.
- [ ] Amazon Macie.
- [ ] Lake Formation (study-only).
- [ ] Glue Data Catalog for lineage (study-only).
- [ ] Responsible-AI principles.

## Exam lens

- Which Guardrails policy stops which threat.
- Grounding checks vs RAG for hallucination.
- Enforcing guardrails through IAM.
- Data-retention controls.
- Programmatic model cards in SageMaker vs a document.

## Close-out

- [ ] Every task sub-issue is closed or cut (stretch) as not planned
- [ ] `JOURNAL.md` entry for week 8 merged with the epic PR
- [ ] Exam skills exercised ticked in `EXAM-COVERAGE.md`
- [ ] Next wave's task issues created (TRACKING.md § 7)
