
---

## Update: moved to Recruiter Intelligence week 10

This is now tracked as a week-10 task of [Recruiter Intelligence](https://github.com/sh3r4rd/sh3r4rd.com/blob/main/docs/recruiter-intelligence/PLAN.md#week-10-operate-it). The acceptance criteria above stand, with one change: if week 10 standardizes on CloudWatch embedded metric format (EMF) for all GenAI metrics, emit `ParseFailures` through EMF instead, and the `cloudwatch:PutMetricData` IAM grant and Go dependency are not needed.
