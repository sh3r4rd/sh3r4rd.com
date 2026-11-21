**Parent**: {{ref:RI}} (Recruiter Intelligence)

Weekly epic for **week 6** of [Recruiter Intelligence](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/README.md). The full specification is [PLAN.md § Week 6](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md#week-6-agent-v1); this issue tracks status. Task issues are attached as sub-issues.

## Exam focus

D2 2.1.1, 2.1.2 (partial), 2.1.4 (stretch), 2.1.6, 2.4.4 (stretch), 2.5.5 · D3 3.1.2 · D5 5.1.7

## Definition of done

the agent answers the 25 tasks from the command line, and the eval numbers are in the journal.

## Approval gates

Python dependencies (`strands-agents`, the AWS SDK). Each task issue that touches one carries `needs-approval`.

## Study

- [ ] ReAct and plan-and-execute patterns.
- [ ] Bedrock Agents (managed: action groups, return of control) vs Strands vs AgentCore.
- [ ] AWS Agent Squad for multi-agent routing (study-only).
- [ ] Agent memory: short-term vs long-term.

## Exam lens

- Which agent framework or service fits which constraint.
- Deterministic tools vs model reasoning for numbers (3.1.2).
- Evaluating agents by task completion, tool-use effectiveness and reasoning quality (5.1.7).

## Close-out

- [ ] Every task sub-issue is closed or cut (stretch) as not planned
- [ ] `JOURNAL.md` entry for week 6 merged with the epic PR
- [ ] Exam skills exercised ticked in `EXAM-COVERAGE.md`
- [ ] Next wave's task issues created (TRACKING.md § 7)
