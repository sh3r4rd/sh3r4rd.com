**Parent**: {{ref:RI}} (Recruiter Intelligence)

Weekly epic for **week 7** of [Recruiter Intelligence](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/README.md). The full specification is [PLAN.md § Week 7](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md#week-7-the-private-assistant-goes-live); this issue tracks status. Task issues are attached as sub-issues.

## Exam focus

D2 2.1.3, 2.1.7, 2.2.1, 2.3.2, 2.3.3, 2.4.1, 2.4.2, 2.5.1 · D1 1.6.2

## Definition of done

you can sign in on your phone, ask "which companies pitched remote Staff Go roles this quarter?", watch it stream, and click through to the cited emails. An unsigned request gets 401.

## Approval gates

A new route in `App.jsx`, an npm dependency, Terraform (Cognito, AgentCore, the authorizer), and a new Lambda. Each task issue that touches one carries `needs-approval`.

## Study

- [ ] MCP (stateless Lambda servers vs ECS for complex tools).
- [ ] WebSockets vs SSE vs streaming APIs.
- [ ] Cognito authorizers vs Lambda authorizers vs IAM authorization.
- [ ] VPC endpoints and PrivateLink for Bedrock (3.2.1 is study-only, since this stack has no VPC).
- [ ] Event-driven integration with EventBridge (2.3.1, 2.3.2).

## Exam lens

- Integrating FMs into existing applications: API Gateway, webhook handlers, event-driven
  designs.
- Choosing a deployment for application needs: on-demand Lambda vs provisioned throughput vs
  SageMaker endpoints (2.2.1).

## Close-out

- [ ] Every task sub-issue is closed or cut (stretch) as not planned
- [ ] `JOURNAL.md` entry for week 7 merged with the epic PR
- [ ] Exam skills exercised ticked in `EXAM-COVERAGE.md`
- [ ] Next wave's task issues created (TRACKING.md § 7)
