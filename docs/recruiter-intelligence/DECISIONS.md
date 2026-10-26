# Decisions

Every decision starts as **Proposed**. It becomes **Accepted** when the week that depends on it
starts, or **Superseded** with a link to the decision that replaced it. Record the evidence
(an eval number, a cost figure, a doc link) next to the status change, not just the verdict.

Prices and service availability below were checked in October 2026 and **will drift**. Re-check
them in the AWS pricing pages and the Bedrock model-availability table before relying on them.

---

## D-01 — Bedrock replaces OpenAI for extraction, after the eval says it may

**Status:** Proposed · **Decided in:** week 3

The extractor calls OpenAI `gpt-5-nano` through `openai-go` with a key held in SSM
(`internal/extractor/extractor.go:18`, `internal/ssm/ssm.go`).

**Decision:** move extraction to the Bedrock **Converse API** using tool use with a JSON schema
for structured output, called from Go through `aws-sdk-go-v2/service/bedrockruntime`. The model
is chosen by the week-3 comparison, not in advance. Cutover happens only when the Bedrock
candidate **matches or beats the OpenAI baseline** on field accuracy for the golden set.

**Why:**
- The exam is about Bedrock.
- The email body (PII) stops leaving AWS. Bedrock does not use prompts or completions to train
  models.
- An IAM grant replaces a long-lived API key.
- Guardrails can attach to the same call (week 8).

**Alternatives rejected:**
- *Keep OpenAI.* It teaches nothing on the syllabus and keeps the third-party PII egress.
- *Run both permanently.* That doubles cost for no product benefit.

OpenAI stays only as the **baseline column in the eval report**. The SSM parameter and the
`openai-go` dependency are removed in the cutover PR.

**Approval gates:** a new Go dependency, IAM changes on the email-parser role, and a
Lambda-behaviour change.

---

## D-02 — Vector store: start on S3 Vectors, then let a three-way bake-off decide

**Status:** Proposed · **Decided in:** week 4 (primary) and week 5 (bake-off)

**Decision:** week 4 builds a **customer-managed Bedrock Knowledge Base on Amazon S3 Vectors**
(GA since December 2025), with **Amazon Titan Text Embeddings V2**. It is the store where you
set every knob yourself: chunking, embedding model, dimension and metadata. Those knobs are what
domain 1 tests.

Week 5 then runs a **time-boxed bake-off** on the same retrieval eval set, keeps the winner with
its numbers, and **destroys the rest the same week**:

| Candidate | Why it is in the bake-off |
|---|---|
| S3 Vectors KB (week 4) | Cheapest at rest: no always-on capacity. **No hybrid search**. AWS documents hybrid search only for RDS (which includes Aurora PostgreSQL), OpenSearch Serverless and MongoDB stores with a filterable text field |
| **Bedrock Managed Knowledge Base** | The newer KB type, which AWS now recommends by default. Bedrock runs the store and provides **hybrid and agentic retrieval** plus a **built-in reranker**. Its managed embedding and reranking carry no extra charge (check the pricing page; this plan has not verified the store's price). Trade-offs: chunking is built-in or fixed-size only, and a custom embedding model must be float32 at 1024 dimensions |
| Aurora Serverless v2 + `pgvector` | Hybrid search under your control, and **the store exam questions actually name**. S3 Vectors is not on the AIP-C01 in-scope list; Aurora and OpenSearch are. Minimum capacity 0 with auto-pause needs PostgreSQL 16.3+ (or 15.7+, 14.12+, 13.15+). Resuming takes about 15 s, or 30 s or more after a day paused, so **warm it before measuring latency** |

**Why hybrid matters here:** recruiter emails are dense with exact tokens such as company names,
"Go", "L6" and "Staff", which pure semantic search tends to blur.

**S3 Vectors limits to design around** (through a KB):
- at most 1 KB of custom metadata and 35 keys per vector, so keep `tech_stack` short
- no `startsWith` or `stringContains` filters
- hierarchical chunking is not recommended on S3 Vectors

**Alternatives rejected:**
- *OpenSearch Serverless.* It has the best retrieval feature set, but bills for minimum
  capacity around the clock, which is the wrong cost shape for this workload. Study it for the
  exam (sharding, the Neural plugin, skill 1.4.3) rather than run it.
- *DynamoDB + client-side similarity.* It teaches nothing the exam asks about.

---

## D-03 — Go for pipeline Lambdas, Python for the agent and the evals

**Status:** Proposed · **Decided in:** week 2 (evals), week 6 (agent), week 7 (MCP server)

| Component | Language | Why |
|---|---|---|
| email-parser, doc-builder, fit-analyzer | **Go** | Matches the existing Lambdas, tests and `make build-lambdas`. Converse is plain AWS SDK |
| inbox agent | **Python** (Strands Agents) | Strands' Python SDK is the most mature, and AgentCore's tooling and samples are Python-first. The exam names Strands explicitly (skills 2.1.1, 2.1.6, 2.5.5) |
| eval harness | **Python** | The evaluation ecosystem (dataframes, statistics, any RAG-metric library you adopt) is Python |
| MCP server | **decide in week 7** | Go has an official MCP SDK, which keeps it beside the other Lambdas. Python is the default if it shares code with the agent's tools |

**Alternative rejected:** *Strands TypeScript* (1.0 is out). It would avoid Python, but
`AGENTS.md:164` forbids creating `.ts`/`.tsx` files anywhere in the repository, not only under
`src/`. Choosing it would mean changing that rule first.

**Approval gates:** a new language toolchain in CI (week 11), and `pyproject.toml` / lockfiles.

---

## D-04 — Public surfaces never read recruiter email text

**Status:** Proposed · **Applies from:** week 1

Public endpoints (`/fit`, dashboard insights, the eval scorecard) may read only:
- **aggregated statistics** computed from DynamoDB (counts, shares, trends; never a row), and
- **my own experience documents**, which I write to be public.

They never read the KB, the archive, or any per-email field beyond what `/recruiters` already
returns (`api-handler/anonymizer.go:14-42`).

**Why:** a public prompt-injection or jailbreak then has nothing to exfiltrate. Defence in depth
matters, but the strongest control is that the data is not reachable. Guardrails (week 8) are
the second layer, not the first.

**Consequence:** the weekly AI market summary is generated **privately** from aggregates and
**human-approved** before publication (D-10). A model with recruiter text in its context never
writes directly to a public page.

---

## D-05 — The fit analyzer puts my experience in context, not in a vector index

**Status:** Proposed · **Decided in:** week 9

My experience documents total a few thousand tokens. Retrieval over them adds chunking, recall
and ranking failure modes and buys nothing when the whole corpus fits comfortably in context.

**Decision:** send the experience documents as a stable system-prompt prefix with **prompt
caching** enabled where the chosen model supports it. Market context comes from the precomputed
aggregates (D-04).

**Check the cache minimum.** A cache checkpoint needs between 512 and 4,096 tokens depending on
the model; some Claude models need 4,096. If the experience documents fall short, put the stable
rubric and the aggregates inside the cached prefix too, or accept no caching for that model and
record the cost difference.

**Revisit if** the experience corpus grows past about 50k tokens, or if the eval shows the model
ignoring relevant sections ("lost in the middle").

**This is itself exam material:** choosing in-context learning over RAG on cost and complexity
grounds (skills 1.1.1, 4.1.1, 4.1.4).

---

## D-06 — Real-email datasets stay out of git, and the vector index stays redacted

**Status:** Proposed · **Applies from:** week 1

- **Golden datasets built from real emails** (labels, retrieval questions, agent tasks) live in
  the private bucket under the `evals/` **S3 prefix**, encrypted with the week-1 KMS key. The
  bucket's default encryption is AES256 (`modules/s3/main.tf:10`), so week 1 applies KMS to
  `evals/` explicitly. Git holds only their **schemas**, the eval
  code and **synthetic** fixtures. CI runs on synthetic data plus aggregate metrics read back
  from S3, never on the raw rows.
- **The vector index holds redacted text.** The doc-builder masks names, email addresses, phone
  numbers and street addresses with Amazon Comprehend `DetectPiiEntities` before writing to
  `kb-source/`. Recruiter identity stays in DynamoDB. The assistant reaches it only through the
  `get_email` tool, behind Cognito.

**Why:** an embedding index is hard to audit and easy to over-share from. Keeping identifiers
out of it means a retrieval bug leaks a job description, not a person.

**Cost of the decision:** "who was the recruiter at X?" needs a tool call, not just retrieval.
That is a feature, because it is exactly the deterministic-tool-versus-retrieval split the exam
tests (skill 3.1.2).

---

## D-07 — The agent runs on Bedrock AgentCore Runtime, with Lambda as the fallback

**Status:** Proposed · **Decided in:** week 7

**Why AgentCore:** it is on the exam's in-scope list and provides session isolation, built-in
observability and long-running sessions without managing compute. Its Runtime accepts **inbound
JWT auth from a Cognito user pool**: the client sends the **access token** (not the ID token) as
a bearer token, and responses stream over SSE. That makes it the simplest route to the
browser (week 7). **Fallback:** a Python Lambda
running the same Strands agent, if AgentCore's pricing or a regional limitation makes it the
wrong choice when you get there. Two facts limit that fallback:
- Lambda's native response streaming exists only on the Node.js runtime. Python needs the Lambda
  Web Adapter or a custom runtime.
- Function URLs accept only IAM or no auth, with no Cognito authorizer. So put the fallback
  behind API Gateway REST with response streaming (`responseTransferMode=STREAM`) and a Cognito
  authorizer. Write the agent so the host is an adapter, which keeps the
choice cheap to reverse.

---

## D-08 — The private assistant uses a Cognito user pool with exactly one user

**Status:** Proposed · **Decided in:** week 7

A Cognito user pool with **self-sign-up disabled**, one admin-created user, MFA on, and an API
Gateway **Cognito authorizer** on the private routes. The existing public dashboard API stays
unauthenticated, as decided in `.claude/plans/implementation-notes.md`.

**Alternatives rejected:**
- *API keys.* They are an identifier, not authentication.
- *IAM SigV4 from the browser.* Possible through Cognito identity pools, but more moving parts
  for one user.
- *Obscurity (an unlinked route).* Ruled out.

---

## D-09 — Spend is capped in code, throttled at the edge and alarmed in billing

**Status:** Proposed · **Applies from:** week 1 (alarms) and week 9 (caps)

Five layers. Each one catches what the previous one misses:

1. **AWS Budgets** monthly alert at 50/80/100% of a figure you set in week 1, plus **Cost
   Anomaly Detection** on Bedrock.
2. **API Gateway throttling** on the public route, as the existing API already does
   (5 rps, burst 10: `modules/api-gateway/variables.tf:46-66`, applied at `modules/api-gateway/main.tf:296-297`), plus a usage plan.
3. **A per-day token budget** in DynamoDB. The fit analyzer increments it atomically and refuses
   with a friendly "come back tomorrow" once it is spent.
4. **A kill switch in AWS AppConfig** that disables the public model call without a deploy.
   This also exercises skill 1.2.2.
5. **An input size limit** (JD length) and a `maxTokens` limit on every call.

Optional: an AWS WAF rate-based rule. It has a monthly fixed cost, so add it only if abuse shows
up in the logs.

---

## D-10 — The weekly market summary is a Step Functions workflow with human approval

**Status:** Proposed · **Decided in:** week 11

Step Functions **Standard** workflow, triggered weekly by EventBridge Scheduler:
1. Compute aggregates.
2. Draft the summary with an **Amazon Bedrock Flow** (formerly Prompt Flows): a sequential
   chain with a condition node for weeks with too little data. `InvokeFlow` returns an event
   stream, and this plan found no direct Step Functions integration for it, so expect a small
   Lambda wrapper (verify when you build).
3. Apply the guardrail and a grounding check.
4. **Wait for my approval** (`waitForTaskToken`, an email link through SNS and API Gateway).
5. Publish to the insights store.

**Why:** it exercises human-in-the-loop (2.1.5), prompt chaining (1.6.6) and safeguarded
workflows (2.1.3) on something real, and enforces D-04: nothing model-written reaches the
public site without a person reading it first.

---

## D-11 — Stay in us-east-1, and call across regions only where forced

**Status:** Proposed

SES email receiving pins the stack to us-east-1. Where a Bedrock feature is missing there, use a
**cross-Region inference profile** or a targeted cross-region call, and record which data crosses
regions.

**The PII rule:** PII may move between **US Regions** through a **US-geography** inference
profile (week 3's extraction and week 8's guardrail rely on this). It never leaves the US. If a
feature would need that, skip the feature.

Checked in October 2026: **Cohere Rerank 3.5 is available in us-east-1** (Amazon Rerank 1.0 is
not), so week 5 needs no cross-region call for reranking. A Guardrails **Standard**-tier
grounding check needs cross-Region inference for the guardrail, which the rule above permits.
