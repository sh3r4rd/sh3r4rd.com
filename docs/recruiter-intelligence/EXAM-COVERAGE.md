# AIP-C01 skill coverage

Every skill in the [AIP-C01 exam guide](https://docs.aws.amazon.com/aws-certification/latest/ai-professional-01/ai-professional-01.html),
with the week that exercises it.

| Mode | Meaning |
|---|---|
| **Build** | The project implements it |
| **Partial** | The project implements part of it; the rest is study |
| **Study** | Deliberately not built, because of cost, no fit, or no VPC/on-prem in this stack. Learn it from the docs |

The skill text is abbreviated. Read the exam guide's wording before an exam session, because
the examples in each skill hint at the answer choices.

**This table is the source of truth.** Each week's "Exam focus" line in `PLAN.md` names that
week's *primary* skills only, so a skill can appear here against a week whose focus line omits it.

Use the **✓** column as a self-check: tick a row when you can explain the skill *and* the
alternatives its examples name.

## Domain 1: Foundation Model Integration, Data Management, and Compliance (31%)

| Skill | Abbreviated | Week | Mode | ✓ |
|---|---|---|---|---|
| 1.1.1 | Architecture aligned to business needs and constraints | 1, 9 | Build | |
| 1.1.2 | Proof of concept on Bedrock to validate feasibility, performance and value | 3 | Build | |
| 1.1.3 | Standardized components, Well-Architected GenAI Lens | 1 (stretch) | Partial | |
| 1.2.1 | Assess and choose FMs: benchmarks, capabilities, limits | 3 | Build | |
| 1.2.2 | Swap models or providers without code changes (Lambda, API Gateway, AppConfig) | 3, 9 | Build | |
| 1.2.3 | Resilience: circuit breakers, Cross-Region Inference, graceful degradation | 3, 9 | Build | |
| 1.2.4 | Customization lifecycle: SageMaker deploys, LoRA/adapters, Model Registry, rollback | 3 | Study | |
| 1.3.1 | Data validation workflows (Glue Data Quality, Data Wrangler, Lambda, CloudWatch) | 1 | Build | |
| 1.3.2 | Multimodal processing (Bedrock multimodal, SageMaker Processing, Transcribe) | — | Study | |
| 1.3.3 | Format input for model-specific requirements (Bedrock JSON, dialog formats) | 2, 3 | Build | |
| 1.3.4 | Improve input quality (reformatting, Comprehend entities, normalization) | 3, 4 | Build | |
| 1.4.1 | Vector database architecture (KB, OpenSearch Neural, RDS + S3, DynamoDB) | 4 | Build | |
| 1.4.2 | Metadata frameworks for precision (timestamps, authorship, domain tags) | 4 | Build | |
| 1.4.3 | High-performance vector stores (OpenSearch sharding, multi-index, hierarchical) | 5 | Study | |
| 1.4.4 | Integration components for document systems, KBs and wikis | 4 | Partial | |
| 1.4.5 | Keeping vector stores current (incremental, change detection, scheduled) | 4 | Build | |
| 1.5.1 | Document segmentation and chunking strategies | 4 | Build | |
| 1.5.2 | Embedding selection (Titan dimensionality, domain fit, batch generation) | 4 | Build | |
| 1.5.3 | Deploying vector search (OpenSearch, Aurora pgvector, KB managed stores) | 4, 5 | Build | |
| 1.5.4 | Advanced search: semantic, hybrid, Bedrock rerankers | 5 | Build | |
| 1.5.5 | Query handling: expansion, decomposition, transformation | 5 | Build | |
| 1.5.6 | Consistent access: function calling, MCP clients, standard retrieval APIs | 5, 7 | Build | |
| 1.6.1 | Instruction frameworks: Prompt Management roles, Guardrails, templates | 3, 8 | Build | |
| 1.6.2 | Interactive systems: clarification flows, intent, conversation history | 7 | Partial | |
| 1.6.3 | Prompt governance: parameterized templates, approval, CloudTrail, CloudWatch Logs | 3, 8 | Build | |
| 1.6.4 | Prompt QA: expected-output checks, edge cases, prompt regression | 2, 3, 11 | Build | |
| 1.6.5 | Iterative refinement: structured inputs, output specs, chain of thought, feedback | 3, 9 | Build | |
| 1.6.6 | Complex prompt systems: Prompt Flows, conditional branching, pre/post-processing | 11 | Build | |

## Domain 2: Implementation and Integration (26%)

| Skill | Abbreviated | Week | Mode | ✓ |
|---|---|---|---|---|
| 2.1.1 | Agents with memory and state (Strands, Agent Squad, MCP) | 6, 7 | Build | |
| 2.1.2 | Structured reasoning: ReAct, chain of thought (Step Functions) | 6 | Partial | |
| 2.1.3 | Safeguarded workflows: stopping conditions, timeouts, IAM bounds, circuit breakers | 6, 7, 11 | Build | |
| 2.1.4 | Model coordination: specialized FMs, ensembles, model selection | 6 (stretch), 9 | Partial | |
| 2.1.5 | Human-in-the-loop: Step Functions review/approval, feedback collection | 11 | Build | |
| 2.1.6 | Reliable tool integrations: Strands API, function definitions, validation, error handling | 6 | Build | |
| 2.1.7 | MCP servers: stateless on Lambda, complex on ECS, client libraries | 7 | Partial | |
| 2.2.1 | Deploying FMs: on-demand Lambda, provisioned throughput, SageMaker endpoints | 7 | Partial | |
| 2.2.2 | LLM-specific container deployments: memory, GPU, model loading | — | Study | |
| 2.2.3 | Balancing performance and resources: smaller models, cascading | 3, 9 | Build | |
| 2.3.1 | Enterprise connectivity: legacy APIs, event-driven loose coupling, data sync | 4 | Partial | |
| 2.3.2 | Adding GenAI to existing apps (API Gateway, webhook Lambdas, EventBridge) | 3, 7 | Build | |
| 2.3.3 | Secure access: federation, RBAC, least-privilege FM access | 7, 8 | Build | |
| 2.3.4 | Cross-environment and jurisdiction: Outposts, Wavelength, hybrid routing | — | Study | |
| 2.3.5 | CI/CD and GenAI gateways (CodePipeline, CodeBuild, security scans, rollback) | 11 | Partial | |
| 2.4.1 | Flexible interaction: sync Bedrock APIs, SDKs + SQS async, API Gateway validation | 3, 7 | Partial | |
| 2.4.2 | Real-time: streaming APIs, WebSockets/SSE, chunked transfer | 7 | Build | |
| 2.4.3 | Resilience: SDK backoff, rate limiting, fallbacks, X-Ray | 3, 9, 10 | Build | |
| 2.4.4 | Model routing: static, content-based (Step Functions), metric-based, API Gateway request transformations | 3, 6 (stretch) | Partial | |
| 2.5.1 | FM API interfaces: streaming, token limits, timeout retries | 7, 9 | Build | |
| 2.5.2 | Accessible interfaces: Amplify, OpenAPI-first, Prompt Flows no-code | 11 | Partial | |
| 2.5.3 | Business enhancements: CRM Lambdas, document processing, Bedrock Data Automation | 1–4 | Partial | |
| 2.5.4 | Developer productivity with Amazon Q Developer | — | Study | |
| 2.5.5 | Advanced apps: Strands, Agent Squad, Step Functions agent patterns, prompt chaining | 6, 11 | Build | |
| 2.5.6 | Troubleshooting: Logs Insights on prompts, X-Ray, Q Developer error patterns | 10 | Build | |

## Domain 3: AI Safety, Security, and Governance (20%)

| Skill | Abbreviated | Week | Mode | ✓ |
|---|---|---|---|---|
| 3.1.1 | Input safety: Guardrails, custom moderation (Step Functions, Lambda) | 8 | Build | |
| 3.1.2 | Output safety: Guardrails, toxicity evals, text-to-SQL for deterministic results | 6, 8 | Build | |
| 3.1.3 | Hallucination reduction: KB grounding, confidence scoring, JSON Schema | 3, 5, 8, 9 | Build | |
| 3.1.4 | Defence in depth: Comprehend pre-filter, Guardrails, Lambda post-checks, API Gateway | 8 | Build | |
| 3.1.5 | Threat detection: prompt injection, jailbreaks, adversarial testing | 8 | Build | |
| 3.2.1 | Protected environments: VPC endpoints, IAM, Lake Formation, access monitoring | 8 | Partial | |
| 3.2.2 | Privacy: Comprehend and Macie PII, Bedrock privacy, Guardrails, S3 Lifecycle retention | 0, 1, 4, 8 | Build | |
| 3.2.3 | Privacy with utility: masking, anonymization | 4, 8 | Build | |
| 3.3.1 | Compliance: programmatic model cards, Glue lineage, decision logs | 8 | Partial | |
| 3.3.2 | Source tracking: Glue Data Catalog, attribution metadata, CloudTrail | 4, 8 | Partial | |
| 3.3.3 | Organizational governance frameworks | 8 | Partial | |
| 3.3.4 | Continuous governance: misuse, drift and bias-drift detection, alerting, redaction, response logs | 10, 11 | Build | |
| 3.4.1 | Transparency: reasoning displays, confidence metrics, citations, agent traces | 7, 9 | Build | |
| 3.4.2 | Fairness: metrics, A/B via Prompt Management and Flows, LLM-as-judge | 9 | Build | |
| 3.4.3 | Policy compliance: Guardrails from policy, model cards, automated checks | 8 | Build | |

## Domain 4: Operational Efficiency and Optimization (12%)

| Skill | Abbreviated | Week | Mode | ✓ |
|---|---|---|---|---|
| 4.1.1 | Token efficiency: estimation, context pruning, response limits, compression | 3, 10 | Build | |
| 4.1.2 | Cost-effective selection: tiered models, price-to-performance | 3, 9 | Build | |
| 4.1.3 | Throughput and utilization: batching, capacity planning, provisioned throughput | 10 | Partial | |
| 4.1.4 | Caching: semantic, fingerprinting, edge, request hashing, prompt caching | 9 | Build | |
| 4.2.1 | Responsiveness: precomputation, latency-optimized models, parallelism, streaming | 7, 9 | Build | |
| 4.2.2 | Retrieval performance: index tuning, query preprocessing, hybrid scoring | 5 | Build | |
| 4.2.3 | Throughput: token processing, batch inference, concurrency management | 10 | Build | |
| 4.2.4 | Model parameters (temperature, top-k/top-p) and A/B tests | 3, 9 | Build | |
| 4.2.5 | Resource allocation: token capacity planning, GenAI-aware auto scaling | 10 | Partial | |
| 4.2.6 | System performance: call profiling, vector query tuning, LLM latency techniques | 5, 10 | Build | |
| 4.3.1 | Holistic observability: operational, tracing, business dashboards | 10 | Build | |
| 4.3.2 | GenAI monitoring: tokens, quality, invocation logs, cost anomalies | 1, 8, 10 | Build | |
| 4.3.3 | Integrated observability: audit, forensic traceability, behaviour patterns | 10, 11 | Build | |
| 4.3.4 | Tool performance: call patterns, tool-call and multi-agent observability | 10 | Build | |
| 4.3.5 | Vector store operations: monitoring, index optimization, data quality | 5, 10 | Build | |
| 4.3.6 | FM-specific troubleshooting: golden datasets, output diffing, reasoning traces | 2, 10 | Build | |

## Domain 5: Testing, Validation, and Troubleshooting (11%)

| Skill | Abbreviated | Week | Mode | ✓ |
|---|---|---|---|---|
| 5.1.1 | Assessment frameworks: relevance, factuality, consistency, fluency | 9, 11 | Build | |
| 5.1.2 | Model evaluation systems: Bedrock evaluations, A/B, canary, cost-performance | 3, 11 | Build | |
| 5.1.3 | User-centred evaluation: feedback UIs, ratings, annotation | 2, 9 | Build | |
| 5.1.4 | Quality assurance: continuous evaluation, regression, deployment gates | 2, 11 | Build | |
| 5.1.5 | Multi-perspective assessment: RAG evaluation, LLM-as-judge, human feedback | 5, 11 | Build | |
| 5.1.6 | Retrieval quality testing: relevance, context matching, latency | 4, 5 | Build | |
| 5.1.7 | Agent performance: task completion, tool effectiveness, Bedrock Agent evaluations, reasoning quality | 6 | Build | |
| 5.1.8 | Reporting: visualizations, automated reports, model comparisons | 11, 12 | Build | |
| 5.1.9 | Deployment validation: synthetic workflows, hallucination and drift checks | 10, 11 | Build | |
| 5.2.1 | Content handling: context overflow, dynamic chunking, truncation errors | 3, 11 | Build | |
| 5.2.2 | Integration issues: error logging, request validation, response analysis | 3, 10 | Build | |
| 5.2.3 | Prompt problems: testing frameworks, version comparison, refinement | 11 | Build | |
| 5.2.4 | Retrieval issues: embedding quality, drift, chunking, vector search performance | 5 | Build | |
| 5.2.5 | Prompt maintenance: template tests, Logs Insights, X-Ray, schema validation | 11 | Build | |

## Summary

<!-- coverage-summary:start -->
| Domain | Skills | Build | Partial | Study |
|---|---|---|---|---|
| D1 | 28 | 22 | 3 | 3 |
| D2 | 25 | 12 | 10 | 3 |
| D3 | 15 | 11 | 4 | 0 |
| D4 | 16 | 14 | 2 | 0 |
| D5 | 14 | 14 | 0 | 0 |
| **Total** | **98** | **73** | **19** | **6** |

Review the Study rows first in week 12 (1.2.4, 1.3.2, 1.4.3, 2.2.2, 2.3.4, 2.5.4), then the Partial ones, which are half-built by design.
<!-- coverage-summary:end -->

## Practice-exam misses

Log every miss from practice questions here. A miss on a **Build** row means the project
taught the mechanism but not the AWS service choice the exam wanted. That's worth a second look
at the skill's examples.

| Date | Source | Skill | What I picked | Right answer | Why |
|---|---|---|---|---|---|
