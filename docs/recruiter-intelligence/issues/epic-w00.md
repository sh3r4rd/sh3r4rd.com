**Parent**: {{ref:RI}} (Recruiter Intelligence)

Weekly epic for **week 0** of [Recruiter Intelligence](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/README.md). The full specification is [PLAN.md § Week 0](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md#week-0-stop-losing-data-do-this-today-about-45-minutes); this issue tracks status. Task issues are attached as sub-issues.

## Definition of done

`archive/raw/` holds a tag-free copy of every surviving raw email, the `incoming/` expiry is raised so nothing more is deleted, and the inventory (bodies preserved, all-time extractions, bodies already lost) is recorded in `JOURNAL.md`.

## Approval gates

A Terraform variable change and an S3 copy against production. Run it yourself or approve it explicitly.

## Close-out

- [ ] Every task sub-issue is closed or cut (stretch) as not planned
- [ ] `JOURNAL.md` entry for week 0 merged with the epic PR
- [ ] Exam skills exercised ticked in `EXAM-COVERAGE.md`
- [ ] Next wave's task issues created (TRACKING.md § 7)
