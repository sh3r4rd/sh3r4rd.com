# Week-by-week plan

**12 weeks at about 9 hours a week:** roughly 6 hours building and 3 hours studying. Each week
has **must** items and **stretch** items. If a week runs long, drop the stretch items, never the
eval or the journal entry.

The exam's domains and weights (AIP-C01) are used throughout:

| Domain | Weight | Main weeks |
|---|---|---|
| D1 Foundation Model Integration, Data Management, and Compliance | 31% | 1, 3, 4, 5 |
| D2 Implementation and Integration | 26% | 6, 7, 9 |
| D3 AI Safety, Security, and Governance | 20% | 8, plus 4, 9 |
| D4 Operational Efficiency and Optimization | 12% | 10, plus 3, 9 |
| D5 Testing, Validation, and Troubleshooting | 11% | 2, 11, plus every week's eval |

Skill numbers such as `1.5.4` refer to the exam guide. [`EXAM-COVERAGE.md`](EXAM-COVERAGE.md)
maps all 98 of them.

## At a glance

| Wk | Theme | Ships | Domains |
|---|---|---|---|
| 0 | **Stop losing data (today)** | Raw emails no longer deleted | — |
| 1 | Data you keep, spend you see | Durable archive, data-quality report, budgets | D1, D3, D4 |
| 2 | Evals first | Golden set, eval harness, OpenAI baseline | D5, D1 |
| 3 | Extraction v2 on Bedrock | Model comparison, versioned prompt, new fields | D1, D4 |
| 4 | RAG ingestion | Redacting doc-builder, Knowledge Base, retrieval eval | D1, D3 |
| 5 | Retrieval quality | Hybrid/rerank/query-rewrite experiments with numbers | D1, D4, D5 |
| 6 | Agent v1 | Strands agent with deterministic tools, agent eval | D2, D3, D5 |
| 7 | Private assistant live | MCP server, AgentCore, Cognito, streaming UI | D2, D1, D3 |
| 8 | Guardrails and security | Guardrail, injection red-team, least privilege, model card | D3 |
| 9 | Public Job-Fit Analyzer | `/fit` live with cache, spend cap, grounding | D2, D3, D4 |
| 10 | Operate it | GenAI dashboard, traces, batch re-extraction | D4 |
| 11 | Continuous evaluation | CI eval gate, Bedrock evaluations, approved weekly summary | D5, D2 |
| 12 | Ship and sit | Case study page, demo, practice exams | all |

---

## Week 0: stop losing data (do this today, about 45 minutes)

`modules/s3/main.tf:15-30` expires everything under `incoming/` after
`s3_lifecycle_expiration_days` (default 30). Versioning is off, so expiry is permanent. The raw
`.eml` is the **only** durable copy of each email body, so every day this waits, the oldest day
of corpus is deleted.

This is a Terraform-variable change and touches production. Run it yourself, or approve it
explicitly.

1. **Copy what exists outside the expiring prefix first.** A lifecycle change can take time to
   take effect, so don't rely on it alone:
   ```bash
   aws login   # the CLI session in this environment had expired
   BUCKET=$(terraform -chdir=infra/recruiter-dashboard output -raw s3_bucket_name)
   aws s3 sync "s3://$BUCKET/incoming/" "s3://$BUCKET/archive/raw/" --copy-props metadata-directive
   ```
   `--copy-props metadata-directive` copies content type and metadata but **not tags**. The
   default (`default`) would also copy the `sender` tag, which holds the recruiter's email
   address, onto every archive copy. Spot-check one copy with
   `aws s3api get-object-tagging --bucket "$BUCKET" --key archive/raw/<id>`; it should return no
   tags. Copying into `archive/` doesn't re-run the parser, because SES invokes the Lambda
   directly (`modules/ses/main.tf:93-97`) rather than through S3 events.
   The bucket policy only *allows* SES to write and contains no `Deny`, so your admin credentials
   can still write to `archive/`. If `sync` is refused, stop and check the policy rather than
   loosening it.
2. **Stop the expiry.** Run `make tf-vars-pull`, then set `s3_lifecycle_expiration_days = 3650`
   in `terraform.tfvars`. Run `terraform -chdir=infra/recruiter-dashboard plan`; it should change
   **only** the lifecycle rule. Then run `terraform -chdir=infra/recruiter-dashboard apply` and
   `make tf-vars-push`. Week 1 replaces this stopgap with a real archive.
3. **Record the inventory** in `JOURNAL.md`:
   ```bash
   aws s3 ls "s3://$BUCKET/archive/raw/" --recursive --summarize | tail -2
   aws dynamodb scan --table-name recruiter-emails --select COUNT --region us-east-1
   ```
   The DynamoDB count **minus one** is every email ever extracted. The api-handler keeps its
   `STATS#cache` item in the same table (`api-handler/handler.go:131`). The S3 count is how many
   of those emails still have a body. The gap between them is permanent.

**Why it matters for the plan:** the S3 count decides how much of weeks 2–5 runs on real email
and how much on synthetic. Fewer than about 50 real bodies means synthetic data carries the
volume while real data stays the ground truth. The corpus grows by about 12 weeks of mail
during the project either way.

---

## Week 1: data you keep, spend you see

**Exam focus:** D1 1.1.1, 1.1.3, 1.3.1 · D3 3.2.2 · D4 4.3.2 (cost anomaly)

**Build (must)**
- [ ] **Durable archive in the email-parser.** After a successful parse, write
      `archive/clean/{messageId}.json`. It holds the cleaned body, parsed headers, `IsForwarded`
      (computed today and never persisted), the extraction result and the parser version.
      Encrypt with a new customer-managed **KMS key**. Use the same key for the `s3://…/evals/`
      prefix (D-06).
- [ ] Add a lifecycle rule on `archive/` that transitions to **S3 Intelligent-Tiering** with no
      expiry. Then set `s3_lifecycle_expiration_days` back to 30. `incoming/` is a landing zone
      again, and `archive/raw/` holds the originals.
- [ ] **Backfill command** (`cmd/backfill` in email-parser). Re-run the existing `parser` package
      over every `archive/raw/*` and write `archive/clean/*`. No LLM calls; it is
      deterministic and re-runnable.
- [ ] **Data-quality report** (`infra/recruiter-dashboard/evals/reports/data-quality.md`,
      aggregates only). It covers:
  - body count and token-length distribution
  - empty or near-empty bodies
  - forwarded share
  - language mix
  - duplicates by `dedup_key`
  - the "Unknown" rate per field in DynamoDB
  - how often `confidence < 0.5`
- [ ] **AWS Budgets** (alerts at 50/80/100%) and **Cost Anomaly Detection** for Bedrock, in a
      new `modules/budget`. Pick the monthly figure now (D-09).
- [ ] Confirm the week-3 candidate models can be invoked in us-east-1. AWS retired the Model
      Access page in October 2025, and serverless models are now enabled automatically. Only
      Anthropic models still need a one-time use-case form.

**Build (stretch)**
- [ ] A one-time **Amazon Macie** sensitive-data discovery job over `archive/`, to see what
      it flags compared with Comprehend in week 4.
- [ ] A **Well-Architected Tool** review using the Generative AI Lens. Keep the findings in
      `DECISIONS.md`.

**Definition of done:** a new test email lands in `archive/clean/` within a minute. The backfill
has run over the whole archive. The report exists. A budget alert email has arrived (trigger
one with a low test threshold, then raise it).

**Approval gates:** Terraform (S3, KMS, budget), IAM (`s3:PutObject` on `archive/*` for the
parser role), an email-parser behaviour change.

**Study (3 h)**
- Read the whole AIP-C01 exam guide once, including the in-scope services list.
- Skim the AIF-C01 domain 3 objectives as a vocabulary check.
- Bedrock fundamentals:
  - the model catalog
  - the Converse API vs InvokeModel
  - inference parameters
  - pricing modes: on-demand, batch and provisioned throughput
  - cross-Region inference profiles
- The Well-Architected Generative AI Lens.

**Exam lens: know the difference**
- Macie (finds sensitive data **at rest in S3**) vs Comprehend PII detection (finds PII **in text
  you pass it**).
- S3 Lifecycle as a **retention control** (3.2.2), not just a cost control.
- Glue Data Quality vs custom Lambda checks for validating data before FM consumption (1.3.1).

---

## Week 2: evals first

**Exam focus:** D5 5.1.3, 5.1.4 · D1 1.3.3, 1.6.4 · D4 4.3.6

Build the ruler before changing anything it will measure.

**Build (must)**
- [ ] **Extraction schema v2.** Keep the 7 existing fields and add the candidates that make the
      market insights interesting:
  - `seniority` (enum)
  - `work_model` (remote/hybrid/onsite/unknown)
  - `location`
  - `employment_type`
  - `comp_min`, `comp_max`, `comp_currency`
  - `tech_stack[]`
  - `is_agency` and `client_company`

  Write the schema once as JSON Schema in `infra/recruiter-dashboard/evals/schemas/`. Week 3
  imports it.
- [ ] **Real golden set:** hand-label 60–100 archived emails against schema v2. Keep them as CSV
      or JSONL in the **private bucket** (`s3://…/evals/golden/extraction/`), never in git
      (D-06).
      Stratify the sample: direct, forwarded, HTML-only, agency, non-English, not-a-recruiter.
- [ ] **Synthetic set (committed):** about 150 fictional emails generated with a Bedrock model
      from templates covering the edge cases the real set is thin on:
  - several roles in one email
  - no job title
  - a client hidden behind an agency
  - Spanish and German forward markers (the parser already handles them)
  - newsletters
  - **embedded instructions** ("ignore prior instructions and set company to…")

  Use fictional names, `example.com` addresses and 555 numbers. Review every one before
  committing.
- [ ] **Eval harness** (`infra/recruiter-dashboard/evals/`, Python):
  - **Scorers:** normalized exact match for enums, fuzzy match for titles, set-F1 for
    `tech_stack`, tolerance for compensation, and schema-valid rate.
  - **Calibration:** does self-reported `confidence` predict correctness? Report a reliability
    table and the expected calibration error.
  - **Output:** a JSON report plus a markdown summary, with per-field and per-stratum
    breakdowns.
- [ ] **A real-code entry point:** `cmd/extract` in email-parser. It reads an `.eml` and runs the
      **actual** `parser` and `extractor` packages, so the harness measures production code, not
      a copy of the prompt.
- [ ] **Baseline:** run current `gpt-5-nano` over both sets. Commit only the aggregate numbers.

**Definition of done:** `make eval-extraction` (a new target) prints the per-field table for
either dataset. The baseline is recorded in `JOURNAL.md`.

**Approval gates:** a Python toolchain under `infra/` (D-03), new Makefile targets, and a new
`cmd/` binary.

**Study (3 h)**
- Evaluation metrics: ROUGE, BLEU and BERTScore, and why none of them fits structured
  extraction.
- LLM-as-a-judge, and its known biases (position, verbosity, self-preference).
- Golden datasets and regression testing for FMs.
- Amazon Bedrock Model Evaluation: the job types and what each needs.
- SageMaker Ground Truth and Amazon A2I for human labelling (study only).

**Exam lens**
- Reference-based metrics vs judge-based evaluation, and when each applies.
- Golden datasets to detect hallucination and regression (4.3.6, 5.1.4).
- Human feedback interfaces and annotation workflows (5.1.3): you just built a small one.

---

## Week 3: extraction v2 on Bedrock

**Exam focus:** D1 1.1.2, 1.2.1, 1.2.2, 1.2.3, 1.3.3, 1.3.4, 1.6.1, 1.6.3, 1.6.4, 1.6.5 ·
D3 3.1.3 · D4 4.1.1, 4.1.2, 4.2.4 · D2 2.4.3 · D5 5.2.2

**Build (must)**
- [ ] **Bedrock extractor in Go.** Use the Converse API with one tool, `record_extraction`,
      whose `inputSchema` is schema v2. Force the tool through `toolChoice` where the model
      supports it, and validate the result against the schema either way. Check whether the
      chosen model offers native structured output; if it does, compare it.
- [ ] **Prompt in Bedrock Prompt Management.** Create a parameterized template, publish
      numbered versions, and have the Lambda pin a **version** ARN, never the draft.
- [ ] **Token budget:** cut quoted reply chains and signatures, then cap the body with the
      currently unused `sanitizer.Truncate`. Plot field accuracy against the cap (for example
      500, 1,000, 2,000 and 4,000 tokens) and pick the knee of the curve. `Truncate` cuts by
      runes (`maxChars`), not tokens, so calibrate the conversion on the corpus using
      Converse's `usage.inputTokens`, or add a token-based cap.
- [ ] **Model comparison** on both golden sets, at temperature 0, for three Bedrock candidates
      of different sizes plus the OpenAI baseline:

      | Model | Field acc. | Schema-valid | p50 / p95 ms | $ per 1k emails |
      |---|---|---|---|---|

      Cost comes from Converse's `usage` token counts multiplied by list price. Record the winner
      in D-01.
- [ ] **Swap models without code changes.** Read the model ID and prompt version from **AWS
      AppConfig** (1.2.2). Use a **US-geography cross-Region inference profile** for resilience
      (1.2.3). It keeps the email bodies inside US Regions (D-11). Use
      the SDK retryer with backoff for throttling (2.4.3).
- [ ] **Put `confidence` to use.** Using week 2's calibration, set a threshold below which a row
      gets `needs_review = true`.
- [ ] **Shadow mode:** run v1 and v2 side by side for at least 5 days. Write v2's result to an
      `extraction_v2` attribute and compare it on live traffic. Then cut over and remove the
      OpenAI dependency and the SSM key (D-01).
- [ ] Extend the `api-handler` allowlist with the new **non-PII** fields (`seniority`,
      `workModel`, `techStack`, `employmentType`, `isAgency`), with tests.

**Definition of done:** v2 is in production. The comparison table is in the journal. Every
extraction call logs model, prompt version, tokens and latency.

**Approval gates:** a Go dependency (`bedrockruntime`, `appconfigdata`), IAM, Terraform
(AppConfig, Prompt Management), an API payload change, and a DynamoDB attribute addition.

**Study (3 h)**
- Converse API tool use.
- Prompt Management: versions vs drafts, and variables.
- Cross-Region inference.
- Model customization trade-offs: prompting, RAG, fine-tuning, LoRA/adapters, distillation
  (1.2.4 is study-only, and the SageMaker Model Registry with it).
- Inference parameters (temperature, top-p, top-k, max tokens).

**Exam lens**
- When fine-tuning is justified over prompt engineering and RAG, and what it costs to run
  (custom models need provisioned throughput or on-demand custom-model pricing; check which).
- AppConfig and API Gateway for model swapping without code changes (1.2.2).
- JSON Schema as a hallucination control (3.1.3).

---

## Week 4: RAG ingestion

**Exam focus:** D1 1.4.1, 1.4.2, 1.4.4, 1.4.5, 1.5.1, 1.5.2, 1.5.3 · D3 3.2.3

**Build (must)**
- [ ] **`doc-builder` Lambda (Go).** It runs on S3 events from `archive/clean/`:
  1. Call Comprehend `DetectPiiEntities`.
  2. Mask names, email addresses, phone numbers and addresses.
  3. Write `kb-source/{id}.txt` plus a `{id}.txt.metadata.json` sidecar (`metadataAttributes`)
     carrying `company`, `received_month`, `seniority`, `work_model`, `tech_stack`, `is_agency`.

  Backfill it over the whole archive.
- [ ] **A customer-managed Knowledge Base on S3 Vectors** (D-02) with Titan Text Embeddings
      V2, using `aws_s3vectors_*` and `aws_bedrockagent_knowledge_base`'s
      `storage_configuration.s3_vectors_configuration`. That block arrived in AWS provider
      6.27.0. The lock file already pins 6.34.0, but `main.tf:7` still says `>= 5.31.0`, so
      tighten it to `>= 6.27`. Keep each metadata sidecar under S3 Vectors' limit of 1 KB and 35
      keys per vector.
- [ ] **Keep it current (1.4.5).** The doc-builder pushes each new document with the
      `IngestKnowledgeBaseDocuments` direct-ingestion API, which makes it searchable in minutes
      without a full sync (check that it supports your data-source type). A **nightly**
      `StartIngestionJob` reconciles anything the direct path missed. Only one ingestion job
      runs at a time per KB and per data source, with 5 per account, so the reconciliation
      can't be per-object.
- [ ] **Chunking and dimension experiment.** Chunking can't be changed after a data source is
      created. Every data source in one KB also writes into the **same** vector index, so two
      strategies side by side in one KB would mix their chunks. **Each configuration therefore
      gets its own KB and its own S3 vector index.** Run two stages to keep the grid small:
  1. At 1024 dimensions, compare **no chunking** (most emails are one short document),
     **fixed-size** and **semantic**. Semantic chunking costs extra model calls at ingestion.
     Skip hierarchical, which AWS doesn't recommend on S3 Vectors because its parent/child
     context lands in non-filterable metadata and can hit the size limit.
  2. With the winning strategy, compare **256 and 512** dimensions against 1024.

  That makes five KBs. Ingest them one after another (the account allows 5 concurrent jobs),
  record the table, and **delete all but the winner**.
- [ ] **Retrieval eval set:** 40–60 questions, each with the IDs of the emails that answer it.
      Real ones stay private and synthetic ones are committed. Report **recall@5, MRR and
      nDCG@5** per configuration.

**Definition of done:** a new email is retrievable within minutes of arrival, with no manual
step. The chunking and dimension table is in the journal, and the chosen configuration is
justified by it.

**Approval gates:** Terraform (KBs, S3 Vectors, the nightly-sync schedule), the AWS provider
constraint bump, IAM (Comprehend, Bedrock KB), a new Lambda.

**Study (3 h)**
- KB components: data sources, parsing options, chunking strategies, metadata filtering,
  supported vector stores (OpenSearch Serverless, Aurora pgvector, Neptune Analytics, S3
  Vectors, and the third-party ones).
- Embedding dimensionality trade-offs.
- Comprehend PII detection vs Bedrock Guardrails' sensitive-information filters.

**Exam lens**
- The chunking strategy and when each fits. It can't be changed after a data source is created.
- Metadata frameworks for precision (1.4.2): timestamps, authorship, domain tags.
- Keeping vector stores current (1.4.5): incremental vs scheduled vs change-detection sync.

---

## Week 5: retrieval quality

**Exam focus:** D1 1.5.4, 1.5.5, 1.5.6, 1.4.3 (study) · D4 4.2.2, 4.3.5 · D5 5.1.6, 5.2.4

**Build (must)**
- [ ] **The vector-store bake-off** (D-02), on the same eval set and the same table:
  - the week-4 S3 Vectors KB, which is semantic only
  - a **Bedrock Managed Knowledge Base**, with built-in hybrid retrieval and reranking
  - **Aurora Serverless v2 + `pgvector`** at minimum capacity 0, with **hybrid search**. Warm
    it before measuring, because a resume from auto-pause adds 15 to 30 seconds or more

  Keep the winner, **destroy the others this week**, and record what each cost.
- [ ] **Reranking:** add **Cohere Rerank 3.5** (available in us-east-1; D-11) to the
      customer-managed path and measure the change. Compare it with the Managed KB's built-in
      reranker.
- [ ] **Query handling:** compare raw query, query expansion with a cheap model, and
      decomposition for compound questions ("Go roles in 2026 vs 2025, remote only"). Measure
      each.
- [ ] **RetrieveAndGenerate (with citations) vs Retrieve plus your own prompt.** Pick one and
      write down why. Factor in week 8: an IAM policy that **requires** a guardrail through
      `bedrock:GuardrailIdentifier` denies RetrieveAndGenerate, which calls the model without
      one. Retrieve plus Converse keeps that option open.
- [ ] **Package retrieval as a function-calling tool** (1.5.6): `search_emails(query, filters)`.
      Week 6's agent and week 7's MCP server both reuse it.
- [ ] **Troubleshooting drill (5.2.4):** break something on purpose, such as a terrible chunk
      size, a stale index or a mismatched embedding model, and diagnose it from metrics alone.
      Write it up as a runbook entry.

**Definition of done:** one retrieval configuration is chosen, each component's contribution is
measured, and nothing is left running that you aren't using.

**Approval gates:** Terraform (the Aurora cluster and the Managed KB; both temporary, but both
go through `plan`).

**Study (3 h)**
- Hybrid search and reranking.
- Query transformation.
- OpenSearch Service vector search: sharding, multiple indexes, the Neural plugin (1.4.3 is
  study-only here).
- Amazon Kendra vs Knowledge Bases.

**Exam lens**
- Why hybrid search helps on exact tokens (product names, levels).
- Where reranking sits and what it costs in latency.
- Query decomposition with Lambda or Step Functions vs the KB's built-in options.

---

## Week 6: agent v1

**Exam focus:** D2 2.1.1, 2.1.2 (partial), 2.1.4 (stretch), 2.1.6, 2.4.4 (stretch), 2.5.5 ·
D3 3.1.2 · D5 5.1.7

**Build (must)**
- [ ] **A Strands agent (Python), run locally first**, with these tools:
  - `search_emails(query, filters)`: from week 5.
  - `market_stats(dimension, filters, period)`: **deterministic**, computed from DynamoDB. The
    model never does arithmetic over retrieved text. This is the text-to-query pattern from
    3.1.2.
  - `get_email(id)`: returns identity fields from DynamoDB, which the redacted index does not
    hold (D-06).
- [ ] **Tool hygiene** (2.1.6): strict input schemas, validation, structured error results the
      model can recover from, and idempotent reads only.
- [ ] **Agent safeguards** (preview of 2.1.3): a maximum number of tool calls per turn, a timeout
      per tool, and a per-session token budget.
- [ ] **Agent eval:** 25 tasks, each with an expected tool sequence and a reference answer.
      Report:
  - task success rate (LLM-as-judge against the reference)
  - tool-selection accuracy
  - median steps
  - tokens and $ per task

**Build (stretch)**
- [ ] **Agents as tools** (2.1.4): a cheap routing model dispatches to a "stats" sub-agent or a
      "research" sub-agent. Measure whether it beats a single agent on cost at equal success.

**Definition of done:** the agent answers the 25 tasks from the command line, and the eval
numbers are in the journal.

**Approval gates:** Python dependencies (`strands-agents`, the AWS SDK).

**Study (3 h)**
- ReAct and plan-and-execute patterns.
- Bedrock Agents (managed: action groups, return of control) vs Strands vs AgentCore.
- AWS Agent Squad for multi-agent routing (study-only).
- Agent memory: short-term vs long-term.

**Exam lens**
- Which agent framework or service fits which constraint.
- Deterministic tools vs model reasoning for numbers (3.1.2).
- Evaluating agents by task completion, tool-use effectiveness and reasoning quality (5.1.7).

---

## Week 7: the private assistant goes live

**Exam focus:** D2 2.1.3, 2.1.7, 2.2.1, 2.3.2, 2.3.3, 2.4.1, 2.4.2, 2.5.1 · D1 1.6.2

**Build (must)**
- [ ] **MCP server** as a **stateless Lambda** (2.1.7). It exposes `search_emails`,
      `market_stats` and `get_email` and requires Cognito-issued tokens. Connect Claude Code to
      it as a demonstration. Decide its language per D-03.
- [ ] **Host the agent on AgentCore Runtime** (D-07), with conversation history in AgentCore
      Memory or a DynamoDB table (1.6.2).
- [ ] **Cognito user pool** (D-08): self-sign-up off, one user, MFA.
- [ ] **Streaming to the browser** (2.4.2). The default path is **AgentCore Runtime's endpoint
      with inbound JWT auth** from the Cognito user pool. The browser sends the **access
      token** as a bearer token, and responses stream over SSE (D-07). If you fall back to
      Lambda, use API Gateway REST with `responseTransferMode=STREAM` and a Cognito authorizer.
      A Function URL doesn't fit: it accepts only IAM or no auth, and native streaming exists
      only on Node.js.
- [ ] **Frontend `/dashboard/assistant`**: Cognito sign-in, a chat UI showing streaming tokens,
      **tool-call traces and citations** (preview of 3.4.1), and `noindex`. For OAuth, prefer a
      small, well-known OIDC library over a hand-written PKCE flow; either way it is an npm
      approval.
- [ ] **Safeguards** (2.1.3) live in production: step limits, timeouts, the session token cap,
      and IAM boundaries on the agent's role.

**Definition of done:** you can sign in on your phone, ask "which companies pitched remote Staff
Go roles this quarter?", watch it stream, and click through to the cited emails. An unsigned
request gets 401.

**Approval gates:** a new route in `App.jsx`, an npm dependency, Terraform (Cognito, AgentCore,
the authorizer), and a new Lambda.

**Study (3 h)**
- MCP (stateless Lambda servers vs ECS for complex tools).
- WebSockets vs SSE vs streaming APIs.
- Cognito authorizers vs Lambda authorizers vs IAM authorization.
- VPC endpoints and PrivateLink for Bedrock (3.2.1 is study-only, since this stack has no VPC).
- Event-driven integration with EventBridge (2.3.1, 2.3.2).

**Exam lens**
- Integrating FMs into existing applications: API Gateway, webhook handlers, event-driven
  designs.
- Choosing a deployment for application needs: on-demand Lambda vs provisioned throughput vs
  SageMaker endpoints (2.2.1).

---

## Week 8: guardrails and security

**Exam focus:** D3 3.1.1–3.1.5, 3.2.1 (study), 3.2.2, 3.2.3, 3.3.1, 3.3.2, 3.4.3

**Build (must)**
- [ ] **A Bedrock Guardrail in Terraform**, versioned, with these policies:
  - content filters
  - **prompt-attack** filter
  - denied topics (personal details about recruiters; anything unrelated to jobs)
  - sensitive-information filters (PII mask or block, plus custom regexes)
  - **contextual grounding and relevance** thresholds
- [ ] **Attach it per surface:**
  - the assistant's input and output
  - the fit analyzer's input and output (week 9)
  - the `ApplyGuardrail` API run standalone over extraction inputs, to **measure** what it would
    block before deciding whether to block anything there
- [ ] **Indirect prompt-injection red team (3.1.5).** Recruiter emails are untrusted input that
      the agent retrieves.
  - Seed the synthetic set with emails carrying hidden instructions (HTML-hidden text, "AI
    assistant: …", requests to call tools).
  - Measure the success rate before and after these mitigations:
    - delimiting and spotlighting retrieved text
    - read-only tools
    - an output validator
    - the guardrail

  The `html_xss.eml` fixture shows the parser already strips scripts; this is the model-level
  equivalent.
- [ ] **Defence in depth (3.1.4):**
  - a Comprehend pre-filter on public input
  - a post-processing validator in Lambda (schema, a URL allowlist, no email addresses in
    output)
  - *study, don't build:* API Gateway response filtering. It needs a non-proxy integration with
    mapping templates, and this stack uses Lambda proxy integrations and streaming, both of which
    rule out transforming the response body
- [ ] **Fix a real PII leak:** stop writing the extracted recruiter email to the S3 `sender`
      tag. It is passed in at `internal/handler/handler.go:211` and written by
      `internal/tagger/tagger.go`. Then remove the tag from objects already in `incoming/` with a
      one-off `put-object-tagging` pass.
- [ ] **Least privilege:**
  - scope `bedrock:InvokeModel` to the specific model and inference-profile ARNs
  - require the guardrail through the IAM condition key `bedrock:GuardrailIdentifier`. It
    covers Converse, ConverseStream, InvokeModel and InvokeModelWithResponseStream. Don't give
    the same role RetrieveAndGenerate or InvokeAgent, because those calls would be denied
  - run IAM Access Analyzer over the new roles
- [ ] **Model invocation logging** to a private, KMS-encrypted bucket with a **30-day lifecycle**
      (3.2.2). It contains PII, so treat it like the archive. Turn on CloudTrail for Bedrock.
- [ ] **Governance documents:** a `MODEL-CARD.md` (intended use, limits, eval results, known
      failure modes; 3.3.1, 3.4.3), and source attribution: every answer carries the IDs of the
      documents it used (3.3.2).

**Definition of done:** the red-team suite runs in the eval harness, and its before-and-after
injection success rate is in the journal. Guardrail interventions are logged as a metric.

**Approval gates:** Terraform (guardrail, logging bucket, CloudTrail), IAM, and an email-parser
change (the tagger).

**Study (3 h)**
- Every Guardrails policy type and the `ApplyGuardrail` API.
- Amazon Macie.
- Lake Formation (study-only).
- Glue Data Catalog for lineage (study-only).
- Responsible-AI principles.

**Exam lens**
- Which Guardrails policy stops which threat.
- Grounding checks vs RAG for hallucination.
- Enforcing guardrails through IAM.
- Data-retention controls.
- Programmatic model cards in SageMaker vs a document.

---

## Week 9: the public Job-Fit Analyzer

**Exam focus:** D2 2.2.3, 2.4.3, 2.5.1 · D1 1.1.1, 1.2.2, 1.6.5 · D3 3.1.3, 3.4.1, 3.4.2 ·
D4 4.1.2, 4.1.4, 4.2.1, 4.2.4 · D5 5.1.1, 5.1.3

**Build (must)**
- [ ] **Experience documents** (`lambda-src/fit-analyzer/experience/*.md`). You write them, they
      are public by design, and they are committed.
- [ ] **Market aggregates**, precomputed from the enriched fields: share remote, seniority mix,
      stack frequency, compensation bands where present. Aggregates only (D-04).
- [ ] **The `fit-analyzer` Lambda (Go)** pipeline:
  1. Validate input: length and language.
  2. Run `ApplyGuardrail` on the input.
  3. A **cheap model classifies "is this a job description?"** and rejects non-JDs before the
     expensive call (cascade; 2.2.3, 4.1.2). The same call condenses the JD into a
     **requirements summary under 1,000 characters**, which step 6 needs.
  4. The main model receives the experience documents as a **prompt-cached prefix** (D-05) plus
     the aggregates.
  5. It returns **structured JSON**:
     - a fit score against a published rubric
     - strengths, each **quoting the experience line it rests on**
     - gaps
     - market context with numbers taken from the aggregates
     - questions to ask the recruiter
  6. Run the **contextual grounding check** through `ApplyGuardrail`, and size its three inputs
     to the documented limits:
     - the **query** is step 3's requirements summary (limit 1,000 characters; a raw JD is
       usually longer)
     - the **grounding source** is the experience documents **plus the aggregates** (limit
       100,000). Leave the aggregates out and the market numbers fail as ungrounded
     - the **response** is checked **one claim at a time** (limit 5,000 characters per
       response; the whole JSON answer can exceed it)

     Compare the Classic and Standard tiers. Standard handles reasoning-style text better but
     needs cross-Region inference for the guardrail (permitted by D-11).
- [ ] **Caching (4.1.4):** a deterministic hash of the normalized JD, stored in DynamoDB with a
      TTL. Semantic caching is a stretch item.
- [ ] **Parameter sweep (4.2.4):** temperature 0, 0.3 and 0.7, and two top-p values, measured on
      the consistency and judge scores below. Extraction fixed temperature at 0; this is the
      surface where the trade-off is real.
- [ ] **Spend controls (D-09):** throttling, the per-day token budget, the AppConfig kill switch
      and size limits. Prove each one with a test that trips it.
- [ ] **Frontend `/fit`:** the input, the structured result, **the evidence behind each claim**
      (3.4.1), and a clear "AI-generated" note. Replace the homepage countdown teaser with a
      link to it.
- [ ] **User feedback (5.1.3):** 👍/👎 plus a reason picked from a fixed list (inaccurate, not
      grounded, unhelpful, other). Store it in DynamoDB keyed by the JD hash, not beside a copy
      of the JD. The only copy of the JD is the cache entry, which expires with its TTL.
      Review feedback weekly. To keep a 👎 case for the fit eval set, copy it before the cache
      entry expires, and only after reading it for anything personal a visitor pasted in.
- [ ] **Fit eval:** 30 JDs (synthetic plus a few public postings, used privately). Measure:
  - groundedness: every strength traceable to an experience line, LLM-as-judge plus a string
    check of the quotes
  - **consistency:** score variance over 5 runs of the same JD
  - **invariance (3.4.2):** swapping the company name or gendered wording must not move the
    score beyond a set tolerance

**Definition of done:** `/fit` is live. Each spend control has been tripped deliberately in a
test. The eval numbers are in the journal.

**Approval gates:** a new route, an API payload (a new endpoint), Terraform, MSW handlers and
fixtures, and an e2e spec.

**Study (3 h)**
- Model cascading and routing.
- Prompt caching vs semantic caching vs edge caching.
- Latency-optimized inference.
- A/B testing prompts with Prompt Management and Prompt Flows (3.4.2).
- Transparency patterns (3.4.1).

**Exam lens**
- Tiered model use by query complexity.
- Deterministic request hashing vs semantic caching: when each one is wrong.
- Graceful degradation when a model is throttled (2.4.3).

---

## Week 10: operate it

**Exam focus:** D4 4.1.1, 4.1.3, 4.2.3, 4.2.5, 4.2.6, 4.3.1–4.3.6 · D2 2.5.6 · D5 5.1.9, 5.2.2

**Build (must)**
- [ ] **Embedded-metric-format metrics from every Lambda and the agent:**
  - tokens in and out
  - estimated $
  - latency
  - cache hit rate
  - guardrail interventions
  - tool calls
  - extraction status

  Extraction status is what finally feeds the `ParseFailures` alarm
  (`modules/monitoring/main.tf:63-81`), which can never fire today.
- [ ] **Surface swallowed failures:** per-record failures in `email-parser` stay non-fatal, but
      they now emit a metric, so the `Errors` alarm is no longer the only signal.
- [ ] **A GenAI CloudWatch dashboard** (cost per surface, p95 latency, token burst
      anomaly-detection bands, guardrail rate, KB ingestion failures; 4.3.5). Add **saved Logs
      Insights queries** over the invocation logs (2.5.6).
- [ ] **Tracing:** X-Ray from API Gateway through Lambda to Bedrock, and AgentCore observability
      for agent spans (4.3.4).
- [ ] **Batch inference (4.1.3, 4.2.3):** re-extract the **whole archive** with the final v2
      prompt through a Bedrock batch inference job. This backfills the new fields across all
      of history, so the insights cover more than the weeks since week 3. Three constraints:
  - A job using the **Converse** format **can't use tool calling**: records with `toolConfig`
    fail with a 400. Use the **InvokeModel** format with the model's native tool-use request,
    or structured outputs.
  - A job needs **at least 100 records**. If the archive is smaller, run the backfill on demand
    with bounded concurrency (still a 4.2.3 exercise) and study batch instead.
  - The discount is 50% for select models only, and prompt caching doesn't apply to batch jobs.
- [ ] **A load test** against `/fit` from your laptop. Confirm throttling and the daily cap hold
      and that nothing scales in a way you didn't expect.

**Definition of done:** one dashboard answers "what did this cost yesterday, and was it any
good?" The batch job has backfilled every archived email.

**Approval gates:** Terraform (dashboard, alarms, X-Ray), and Lambda changes.

**Study (3 h)**
- Provisioned throughput (when it pays; study-only).
- Bedrock model invocation logs.
- Output diffing and reasoning-path tracing (4.3.6).
- Capacity planning for token workloads.

**Exam lens**
- Batch vs on-demand vs provisioned throughput.
- Token-burst anomaly detection.
- Golden-dataset hallucination monitoring in production.

---

## Week 11: continuous evaluation

**Exam focus:** D5 5.1.1, 5.1.2, 5.1.4, 5.1.5, 5.1.8, 5.1.9, 5.2.1, 5.2.3, 5.2.5 ·
D2 2.1.5, 2.3.5, 2.5.2 · D1 1.6.4, 1.6.6 · D3 3.3.4

**Build (must)**
- [ ] **A CI eval gate:** a new `evals.yml` workflow, run on PRs that touch prompts, model
      configuration, retrieval or the agent.
  - It uses a **GitHub OIDC** role allowed to invoke only the eval models.
  - It runs the synthetic suites and fails on regression past set thresholds.
  - Real-data evals run from your laptop and publish **aggregates** to S3 for the scorecard
    (D-06).
- [ ] **Bedrock Model Evaluation and RAG evaluation jobs:** run the managed LLM-as-judge and RAG
      evaluation against the KB, and compare their verdicts with your harness. Where they
      disagree, read the cases.
- [ ] **Canary prompt rollout (5.1.2):** AppConfig serves prompt version N+1 to a percentage of
      extraction calls. Compare it on live metrics, then promote or roll back.
- [ ] **The weekly market summary (D-10):** a Step Functions Standard workflow (aggregates →
      Bedrock Flow draft → guardrail and grounding check → **`waitForTaskToken` approval
      email** → publish). This covers 2.1.5 and 1.6.6.
- [ ] **A public eval scorecard:** the latest aggregate metrics per surface, published to the
      insights store and rendered on `/dashboard`.

- [ ] **Troubleshooting drills (5.2.1, 5.2.3, 5.2.5):** context-window overflow, a
      prompt-format regression and schema drift. One runbook entry each.

**Definition of done:** a PR that degrades the extraction prompt fails CI. One weekly summary has
gone through approval and been published.

**Approval gates:** a workflow change, an IAM OIDC role, Terraform (Step Functions, Scheduler,
SNS), and an API payload (the scorecard).

**Study (3 h)**
- Bedrock evaluation types (automatic, human, LLM-as-judge, RAG).
- A/B and canary testing for FMs.
- CodePipeline and CodeBuild for GenAI (2.3.5, mostly study).
- The GenAI gateway pattern.
- **Take the official practice question set** and log every miss in `EXAM-COVERAGE.md`.

---

## Week 12: ship and sit

**Exam focus:** review across all domains, with the study-only rows of `EXAM-COVERAGE.md`
first.

**Build (must)**
- [ ] **A case study page** at `/projects/recruiter-intelligence`. It covers the problem, the
      architecture diagram, **the numbers** (eval results, cost per surface, the injection
      red-team result), what failed, and links to `/fit`, the insights and the repository.
- [ ] **Homepage:** a Projects entry. The current `projects` shape (`HomePage.jsx:8-45`) has no
      link field, so add one. Remove the countdown teaser. Add skills through `/add-skill`;
      note that its template omits `category`, so a new skill falls into "Other"
      (`SkillsSection.jsx:101`) unless you add the field by hand.
- [ ] **Tests:** MSW handlers and fixtures for each new endpoint (`onUnhandledRequest: 'error'`
      fails anything missing), Playwright specs for `/fit` and the insights, and coverage held
      at 75% or above.
- [ ] **Documentation:** update `docs/architecture-recruiter-dashboard.md`. While you are there,
      fix the stale items found in the week-0 read: the anonymizer is an allowlist, and the
      cache write is now permitted.
- [ ] A two-minute demo video and a written post. `JOURNAL.md` is the draft.

**Exam**
- [ ] Two timed full-length practice runs. Review every miss against the exam guide's skill
      text, not just the answer key.
- [ ] Book the exam for week 13 or 14, while it's fresh.

**Definition of done:** a recruiter who lands on the homepage can reach a working demo and a
write-up with real numbers in two clicks.

**Approval gates:** a new route, a homepage data-shape change, and npm or e2e changes.

---

## Risks and how the plan absorbs them

| Risk | Mitigation |
|---|---|
| The real corpus is small | Synthetic data carries volume and real data is ground truth (week 2). Week 10's batch backfill and 12 weeks of new mail grow it |
| A service feature is missing in us-east-1 or in the Terraform provider | Every such spot names a fallback (D-02, D-07, D-11, week 4). Record the gap; don't click-ops around it |
| Cost overrun | The D-09 layers, Budgets from week 1, Aurora destroyed in week 5, and no always-on vector capacity |
| Recruiter PII leaks through a public surface | D-04: the data isn't reachable from public surfaces. Guardrails, validators and the red team come second |
| Scope creep | Must vs stretch per week. Cut stretch items, never evals or the journal |
| A week slips | Weeks 10 and 11 have the most stretch to absorb it. Week 12's build list is the minimum showcase |
