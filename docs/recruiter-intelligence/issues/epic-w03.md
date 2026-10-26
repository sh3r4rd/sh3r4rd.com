**Parent**: {{ref:RI}} (Recruiter Intelligence)

Weekly epic for **week 3** of [Recruiter Intelligence](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/README.md). The full specification is [PLAN.md § Week 3](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md#week-3-extraction-v2-on-bedrock); this issue tracks status. Task issues are attached as sub-issues.

## Exam focus

D1 1.1.2, 1.2.1, 1.2.2, 1.2.3, 1.3.3, 1.3.4, 1.6.1, 1.6.3, 1.6.4, 1.6.5 · D3 3.1.3 · D4 4.1.1, 4.1.2, 4.2.4 · D2 2.4.3 · D5 5.2.2

## Definition of done

v2 is in production. The comparison table is in the journal. Every extraction call logs model, prompt version, tokens and latency.

## Approval gates

A Go dependency (`bedrockruntime`, `appconfigdata`), IAM, Terraform (AppConfig, Prompt Management), an API payload change, and a DynamoDB attribute addition. Each task issue that touches one carries `needs-approval`.

## Study

- [ ] Converse API tool use.
- [ ] Prompt Management: versions vs drafts, and variables.
- [ ] Cross-Region inference.
- [ ] Model customization trade-offs: prompting, RAG, fine-tuning, LoRA/adapters, distillation
  (1.2.4 is study-only, and the SageMaker Model Registry with it).
- [ ] Inference parameters (temperature, top-p, top-k, max tokens).

## Exam lens

- When fine-tuning is justified over prompt engineering and RAG, and what it costs to run
  (custom models need provisioned throughput or on-demand custom-model pricing; check which).
- AppConfig and API Gateway for model swapping without code changes (1.2.2).
- JSON Schema as a hallucination control (3.1.3).

## Close-out

- [ ] Every task sub-issue is closed or cut (stretch) as not planned
- [ ] `JOURNAL.md` entry for week 3 merged with the epic PR
- [ ] Exam skills exercised ticked in `EXAM-COVERAGE.md`
- [ ] Next wave's task issues created (TRACKING.md § 7)
