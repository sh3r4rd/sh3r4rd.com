**Parent**: {{ref:RI}} (Recruiter Intelligence)

Weekly epic for **week 4** of [Recruiter Intelligence](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/README.md). The full specification is [PLAN.md § Week 4](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md#week-4-rag-ingestion); this issue tracks status. Task issues are attached as sub-issues.

## Exam focus

D1 1.4.1, 1.4.2, 1.4.4, 1.4.5, 1.5.1, 1.5.2, 1.5.3 · D3 3.2.3

## Definition of done

a new email is retrievable within minutes of arrival, with no manual step. The chunking and dimension table is in the journal, and the chosen configuration is justified by it.

## Approval gates

Terraform (KBs, S3 Vectors, the nightly-sync schedule), the AWS provider constraint bump, IAM (Comprehend, Bedrock KB), a new Lambda. Each task issue that touches one carries `needs-approval`.

## Study

- [ ] KB components: data sources, parsing options, chunking strategies, metadata filtering,
  supported vector stores (OpenSearch Serverless, Aurora pgvector, Neptune Analytics, S3
  Vectors, and the third-party ones).
- [ ] Embedding dimensionality trade-offs.
- [ ] Comprehend PII detection vs Bedrock Guardrails' sensitive-information filters.

## Exam lens

- The chunking strategy and when each fits. It can't be changed after a data source is created.
- Metadata frameworks for precision (1.4.2): timestamps, authorship, domain tags.
- Keeping vector stores current (1.4.5): incremental vs scheduled vs change-detection sync.

## Close-out

- [ ] Every task sub-issue is closed or cut (stretch) as not planned
- [ ] `JOURNAL.md` entry for week 4 merged with the epic PR
- [ ] Exam skills exercised ticked in `EXAM-COVERAGE.md`
- [ ] Next wave's task issues created (TRACKING.md § 7)
