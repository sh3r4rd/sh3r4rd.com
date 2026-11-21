# Well-Architected review: Generative AI Lens

A review of the Recruiter Intelligence **target architecture** (end of week 12, as described in
[`README.md`](README.md) and [`PLAN.md`](PLAN.md)) against the AWS Well-Architected Generative AI
Lens. It was done in week 1 as a stretch task (#137, exam skill 1.1.3). The high-risk findings
and the plan's response to each are recorded in [`DECISIONS.md`](DECISIONS.md) § D-14.

| | |
|---|---|
| Workload | `Recruiter Intelligence` |
| Workload ID | `598b0aa89ae8b4d921e26e62b3261d95` |
| Region | us-east-1 |
| Environment | Pre-production |
| Review owner | `sh3r4rd` |
| Lens | Generative AI Lens, version `2025-04-17` (AWS official) |
| Reviewed | 21 November 2026 |
| Milestone | 1, `RI-W01 baseline` |

## Method

- **What was graded.** The architecture the plan says will exist at the end of week 12, not what
  runs today. Today's pipeline (OpenAI extraction, no evals, no guardrail) would score far worse.
  So every "None" risk below is a commitment in `PLAN.md`, and it holds only if that week ships.
- **How each answer was chosen:**
  - **Selected** means the plan implements the practice. The week that delivers it is in the
    table.
  - **Not applicable** is used only where the practice targets something this workload doesn't
    have (training jobs, customer-managed networks), or for a deliberate single-Region or no-VPC
    choice. Each one has a reason in the tool, and the ones that are really accepted risks are
    listed below.
  - **Unselected** means the plan doesn't do it. Where a question's only practice is missing, the
    answer is "None of these".
- **Who computed the risk.** The Well-Architected Tool, from the selections. One missing
  practice is enough to make a question **High** if AWS rates that practice high-risk.
- **Done with the AWS CLI** (`aws wellarchitected …`): `create-workload`, `update-answer` for all
  29 questions with notes, `get-lens-review`, `list-lens-review-improvements` and
  `create-milestone`. The tool is free.
- **Two quirks of the tool:**
  - The core **AWS Well-Architected Framework** lens is attached to every workload and can't be
    removed (`DisassociateLenses` returns "Cannot disassociate default lens"). Its questions are
    left unanswered; this review covers the Generative AI Lens only.
  - The `genai` alias is rejected by `create-workload`. Pass the lens ARN
    (`arn:aws:wellarchitected::aws:lens/genai`, which is AWS-owned and contains no account ID).
    The tool also keeps notes only on not-applicable choices, so the reasoning for each answer is
    in the question-level notes.

## Risk counts

| Pillar | Questions | High | Medium | None |
|---|---|---|---|---|
| Operational excellence | 5 | **2** | 0 | 3 |
| Security | 6 | 0 | 0 | 6 |
| Reliability | 6 | 0 | **1** | 5 |
| Performance efficiency | 4 | 0 | 0 | 4 |
| Cost optimization | 5 | 0 | 0 | 5 |
| Sustainability | 3 | 0 | 0 | 3 |
| **Total** | **29** | **2** | **1** | **26** |

No question is unanswered, and none was marked not applicable as a whole.

## Findings

| # | Question | Risk | Missing practice | Response (D-14) |
|---|---|---|---|---|
| HRI-1 | genops1: consistent model output quality | High | GENOPS01-BP01, periodically evaluate functional performance | **Open.** The plan evaluates on every change but never on a schedule. Proposed for week 11: a scheduled run of the synthetic suites plus a monthly real-data run |
| HRI-2 | genops4: lifecycle automation | High | GENOPS04-BP02, GenAIOps | **Open, owner decision.** CI evaluates but nothing deploys the backend. Proposed: a plan-on-PR, apply-on-merge pipeline through GitHub OIDC, or accept manual applies as a recorded risk |
| MRI-1 | genrel1: throughput quotas | Medium | GENREL01-BP01, scale throughput as a function of utilization | **Open.** Proposed: check Bedrock quotas before the week 3 cutover, and alarm on `InvocationThrottles` in week 10 |

Improvement plans:
[GENOPS01-BP01](https://docs.aws.amazon.com/wellarchitected/latest/generative-ai-lens/genops01-bp01.html),
[GENOPS04-BP02](https://docs.aws.amazon.com/wellarchitected/latest/generative-ai-lens/genops04-bp02.html),
[GENREL01-BP01](https://docs.aws.amazon.com/wellarchitected/latest/generative-ai-lens/genrel01-bp01.html).

### Not-applicable choices that are really accepted risks

These don't raise the risk count, but the reason for each is a choice, not an absence:

| Choice | Why it is accepted |
|---|---|
| gensec1_2: private network communication to models | The stack has no VPC. Bedrock is called over TLS with SigV4 from Lambda and AgentCore. VPC interface endpoints would add a VPC and a per-endpoint, per-AZ hourly charge for one developer's workload. Private networking is study-only (3.2.1) |
| genrel5_2: replicate embedding data across Regions | SES email receiving pins the stack to us-east-1 (D-11). The KB is derived data that can be rebuilt from the S3 archive |
| genrel5_3: agent capabilities across Regions | One user. A us-east-1 outage takes the assistant down, which is acceptable |

Truly not applicable: genrel2_1 (no customer-managed network paths), genrel6_1, gensec6_1 and
gensus1_2 (no training or customization). gensec6's RAG analogue, crafted emails poisoning the
index, is covered by the week 2 synthetic injection emails and the week 8 red team.

## Per-question answers

"Weeks" is when the plan delivers the practice. The full notes are on each answer in the tool.

| Pillar | ID | Question | Our answer | Weeks | Risk |
|---|---|---|---|---|---|
| Operational excellence | genops1 | How do you achieve and verify consistent model output quality? | Stratified golden sets, a CI eval gate and canaries, plus `/fit` feedback. **Gap:** evals run only when something changes, never on a schedule | 2, 9, 11 | **High** |
| Operational excellence | genops2 | How do you monitor and manage the operational health of your applications? | Week 10's metrics, dashboard and traces cover every layer. Week 8 adds invocation logs. D-09 throttles and caps the public surface | 8, 9, 10 | None |
| Operational excellence | genops3 | How do you maintain traceability for your models, prompts, and assets? | Extraction prompt in Prompt Management with pinned versions. X-Ray and AgentCore spans. Other surfaces' prompts are versioned in git only | 3, 7, 10, 11 | None |
| Operational excellence | genops4 | How do you automate the lifecycle management of your generative AI workloads? | Everything in Terraform. **Gap:** CI evaluates but never deploys; the backend ships by a local `terraform apply` (week 0 found Lambda hash drift) | 1+, 11 | **High** |
| Operational excellence | genops5 | How do you determine when to execute Gen AI model customization? | Prompting, then RAG, then in-context docs (D-05). Customization studied and deliberately not built | 3, 4, 9 | None |
| Security | gensec1 | How do you manage access to generative AI endpoints? | Week 8 scopes model ARNs and makes the guardrail mandatory in IAM. D-04 and D-06 limit data access. No VPC endpoints (accepted) | 1, 7, 8 | None |
| Security | gensec2 | How do you prevent generative AI applications from generating harmful, biased, or factually incorrect responses? | Week 8 adds the guardrail with grounding checks. Schema validation, deterministic stats tools, quoted evidence and human approval back it up | 3, 6, 8, 9, 11 | None |
| Security | gensec3 | How do you monitor and audit events associated with your generative AI workloads? | Week 8 adds CloudTrail for Bedrock and invocation logs (KMS, 30 days). Data events for the KB and the archive need an explicit opt-in | 8, 10 | None |
| Security | gensec4 | How do you secure system and user prompts? | Prompt Management for extraction. Public input is validated, guardrailed and classified. Email text is spotlighted and red-teamed | 3, 8, 9 | None |
| Security | gensec5 | How do you prevent excessive agency for models? | Read-only tools, IAM boundaries on the agent role, Cognito-gated MCP, and no publishing without a human | 6, 7, 8 | None |
| Security | gensec6 | How do you detect and remediate data poisoning risks? | Not applicable: nothing is trained. The RAG analogue (crafted emails in the index) is covered by the week 8 red team | n/a (RAG analogue: 2, 4, 8) | None |
| Reliability | genrel1 | How do you determine throughput quotas (or needs) for foundation models? | **Gap:** demand is capped and measured, but never compared with the account's Bedrock quotas, and nothing alarms on throttling | open | **Medium** |
| Reliability | genrel2 | How do you maintain reliable communication between different components of your generative AI architecture? | Not applicable: no VPC or customer-managed network paths, only regional AWS endpoints | n/a | None |
| Reliability | genrel3 | How do you implement remediation actions for generative AI workload loops, retries, and failures? | Schema validation, retries, structured tool errors, graceful refusal on cap, and timeouts and step limits on the agent | 1, 3, 6, 7, 9, 11 | None |
| Reliability | genrel4 | How do you maintain versions for prompts, model parameters, and foundation models? | Prompt Management versions. AppConfig records which model and prompt version each surface uses and can roll it back | 3, 9, 11 | None |
| Reliability | genrel5 | How do you distribute inference workloads over multiple regions of availability? | Inference uses a US cross-Region profile. The KB and agent are single-Region by design (SES pins us-east-1), an accepted risk | 3 | None |
| Reliability | genrel6 | How do you design high-performance distributed computation tasks to maximize successful completion? | Not applicable: no training or distributed jobs. The batch job is managed and the backfill can be re-run | n/a | None |
| Performance efficiency | genperf1 | How do you capture and improve the performance of your generative AI models in production? | A golden set for every surface, from week 2 on. Per-call metrics from week 3, dashboard in week 10 | 2, 3, 4, 6, 9, 10 | None |
| Performance efficiency | genperf2 | How do you verify your generative AI workload maintains acceptable performance levels? | Week 10 load-tests `/fit`. Week 9 sweeps temperature and top-p. Week 3 picks the model by eval | 3, 9, 10 | None |
| Performance efficiency | genperf3 | How do you optimize computational resources required for high-performance distributed computation tasks? | No distributed computation of its own; every model and the agent run on managed services | 3, 7 | None |
| Performance efficiency | genperf4 | How do you improve the performance of data retrieval systems? | Week 4 tests chunking and dimensions. Week 5 tests stores, hybrid search, reranking and query rewriting, all scored on one eval set | 4, 5 | None |
| Cost optimization | gencost1 | How do you select the appropriate model to optimize costs? | Week 3's comparison chooses the extraction model on accuracy and $ per 1k emails. Week 9 puts a cheap classifier in front of the expensive model | 3, 9 | None |
| Cost optimization | gencost2 | How do you select a cost-effective pricing model (for example, provisioned, on-demand, hosted, or batch)? | On-demand Bedrock everywhere, batch inference for the archive backfill, no always-on capacity, losing stores destroyed the week they lose | 4, 5, 9, 10 | None |
| Cost optimization | gencost3 | How do you engineer prompts to optimize cost? | Week 3 trims reply chains and caps the body at the accuracy knee. Forced tool schemas and `maxTokens` on every call bound the output | 3, 9 | None |
| Cost optimization | gencost4 | How do you optimize vector stores for cost? | Week 4 measures 256, 512 and 1024 dimensions and keeps the winner | 4 | None |
| Cost optimization | gencost5 | How do you optimize agent workflows for cost? | Step limits, per-tool timeouts and a per-session token budget on the agent | 6, 7 | None |
| Sustainability | gensus1 | How do you minimize the computational resources needed for training, customizing, and hosting generative AI workloads? | All serverless or scale-to-zero. Nothing is customized | 1+ | None |
| Sustainability | gensus2 | How can you optimize data processing and storage to minimize energy consumption and maximize efficiency? | Intelligent-Tiering, lifecycles, smaller vectors, caching and batch, and nothing left running unused | 1, 3, 4, 5, 9, 10 | None |
| Sustainability | gensus3 | How do you maintain model efficiency and resource optimization when working with large language models? | Week 3 picks the smallest model that matches the baseline. Small models do classification and query expansion | 3, 5, 9 | None |

## Smaller observations

These answers are "None" because the plan covers the practice, but the coverage is thinner than
the lens describes. Worth a look in the week named:

- **Prompts outside Prompt Management (genops3, genrel4, gensec4).** Only the extraction prompt
  is placed in Bedrock Prompt Management (week 3). The fit analyzer, agent and weekly-summary
  prompts are versioned in git and deployed with their code. Decide in weeks 6, 9 and 11 whether
  git is the catalog for those or whether they move too.
- **Audit data events (gensec3).** Week 8 says "turn on CloudTrail for Bedrock". `InvokeModel` and
  `Converse` are management events, so they already appear in the 90-day event history, and a
  trail keeps them longer. Four other kinds are **data events**, which are opt-in and charged
  per event: KB `Retrieve`, `InvokeFlow`, `ApplyGuardrail` and Prompt Management's
  `RenderPrompt`. The same goes for S3 data events on `archive/`. Every one of them is on this
  plan's path. Name them, or decide against them, when week 8 is built
  ([Bedrock CloudTrail docs](https://docs.aws.amazon.com/bedrock/latest/userguide/logging-using-cloudtrail.html)).
- **Throttled extractions (genrel1).** Per-record failures are non-fatal, so an extraction that
  runs out of retries leaves an email that is archived but not extracted. Week 10's re-extraction
  is the recovery path, and week 10's metric is what makes the gap visible.

## Keeping it current

Re-run the review against **what was actually built** at the end of week 12. Update the
answers, record a second milestone, and compare it with milestone 1 here:

```bash
W=598b0aa89ae8b4d921e26e62b3261d95
L=arn:aws:wellarchitected::aws:lens/genai
aws wellarchitected get-lens-review --region us-east-1 --workload-id "$W" --lens-alias "$L"
aws wellarchitected list-lens-review-improvements --region us-east-1 --workload-id "$W" --lens-alias "$L"
aws wellarchitected create-milestone --region us-east-1 --workload-id "$W" --milestone-name "RI-W12 as built"
```
