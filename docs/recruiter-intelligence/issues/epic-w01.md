**Parent**: {{ref:RI}} (Recruiter Intelligence)

Weekly epic for **week 1** of [Recruiter Intelligence](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/README.md). The full specification is [PLAN.md § Week 1](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md#week-1-data-you-keep-spend-you-see); this issue tracks status. Task issues are attached as sub-issues.

## Exam focus

D1 1.1.1, 1.1.3, 1.3.1 · D3 3.2.2 · D4 4.3.2 (cost anomaly)

## Definition of done

a new test email lands in `archive/clean/` within a minute. The backfill has run over the whole archive. The report exists. A budget alert email has arrived (trigger one with a low test threshold, then raise it).

## Approval gates

Terraform (S3, KMS, budget), IAM (`s3:PutObject` on `archive/*` for the parser role), an email-parser behaviour change. Each task issue that touches one carries `needs-approval`.

## Study

- [ ] Read the whole AIP-C01 exam guide once, including the in-scope services list.
- [ ] Skim the AIF-C01 domain 3 objectives as a vocabulary check.
- [ ] Bedrock fundamentals:
  - the model catalog
  - the Converse API vs InvokeModel
  - inference parameters
  - pricing modes: on-demand, batch and provisioned throughput
  - cross-Region inference profiles
- [ ] The Well-Architected Generative AI Lens.

## Exam lens

- Macie (finds sensitive data **at rest in S3**) vs Comprehend PII detection (finds PII **in text
  you pass it**).
- S3 Lifecycle as a **retention control** (3.2.2), not just a cost control.
- Glue Data Quality vs custom Lambda checks for validating data before FM consumption (1.3.1).

## Close-out

- [ ] Every task sub-issue is closed or cut (stretch) as not planned
- [ ] `JOURNAL.md` entry for week 1 merged with the epic PR
- [ ] Exam skills exercised ticked in `EXAM-COVERAGE.md`
- [ ] Next wave's task issues created (TRACKING.md § 7)
