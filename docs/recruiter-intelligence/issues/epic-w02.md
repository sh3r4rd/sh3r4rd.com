**Parent**: {{ref:RI}} (Recruiter Intelligence)

Weekly epic for **week 2** of [Recruiter Intelligence](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/README.md). The full specification is [PLAN.md § Week 2](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md#week-2-evals-first); this issue tracks status. Task issues are attached as sub-issues.

## Exam focus

D5 5.1.3, 5.1.4 · D1 1.3.3, 1.6.4 · D4 4.3.6

## Definition of done

`make eval-extraction` (a new target) prints the per-field table for either dataset. The baseline is recorded in `JOURNAL.md`.

## Approval gates

A Python toolchain under `infra/` (D-03), new Makefile targets, and a new `cmd/` binary. Each task issue that touches one carries `needs-approval`.

## Study

- [ ] Evaluation metrics: ROUGE, BLEU and BERTScore, and why none of them fits structured
  extraction.
- [ ] LLM-as-a-judge, and its known biases (position, verbosity, self-preference).
- [ ] Golden datasets and regression testing for FMs.
- [ ] Amazon Bedrock Model Evaluation: the job types and what each needs.
- [ ] SageMaker Ground Truth and Amazon A2I for human labelling (study only).

## Exam lens

- Reference-based metrics vs judge-based evaluation, and when each applies.
- Golden datasets to detect hallucination and regression (4.3.6, 5.1.4).
- Human feedback interfaces and annotation workflows (5.1.3): you just built a small one.

## Close-out

- [ ] Every task sub-issue is closed or cut (stretch) as not planned
- [ ] `JOURNAL.md` entry for week 2 merged with the epic PR
- [ ] Exam skills exercised ticked in `EXAM-COVERAGE.md`
- [ ] Next wave's task issues created (TRACKING.md § 7)
