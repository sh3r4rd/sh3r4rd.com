# Journal

One entry per week, written at the end of the week, in 15 minutes. This is the raw material for
the week-12 case study, and the numbers in it are the case study's credibility. Record numbers
and decisions here, not feelings.

**Never paste real email content, recruiter names or addresses here.** This file is committed.
Use aggregate numbers and synthetic examples only (D-06).

---

## Week 0: inventory

- Date:
- Raw bodies preserved (`archive/raw/` object count):
- DynamoDB rows (all-time extractions):
- Bodies permanently lost (rows − bodies, approximate):
- Lifecycle change applied (date, `terraform plan` summary):

---

## Week 0: stop losing data

Epic #118 · task #131. The inventory numbers are in § Week 0: inventory above.

**Shipped**
- A tag-free copy of every surviving raw email in `archive/raw/`, outside the expiring prefix.
- The `incoming/` expiry raised from 30 to 3650 days, a stopgap until week 1's archive.
- The inventory, recorded above.

**Numbers** (eval, cost, latency; with the dataset and its size)

| Metric | Before | After | Dataset (n) |
|---|---|---|---|
| Raw bodies held outside the expiring prefix | 0 | 25 | `incoming/` at copy time (25) |
| `incoming/` expiry (days) | 30 | 3650 | lifecycle rule `expire-raw-emails` |
| Archive copies carrying the `sender` tag | — | 0 | every copy checked (25) |
| Extractions whose body is already gone | — | 67 | DynamoDB extraction rows (92) |

**Decisions** (link the `DECISIONS.md` entry and any status change)
- No `DECISIONS.md` entry; the choices were operational.
- The apply was targeted at the lifecycle rule (`-target`, from a saved plan). The full plan also
  wanted to redeploy both Lambdas from local build artifacts and to drop the bucket's SSE-C
  block, which is #98's job.
- Counted extractions by excluding `id` values that start with `STATS#`, not by subtracting
  one. `PLAN.md` § Week 0 is corrected.

**What broke or surprised me**
- Week 0 ran about three weeks late: due 1 November, done 20 November. About 26 of the 67 lost
  bodies expired between the plan's approval on 25 October and the copy.
- 25 real bodies is under the plan's threshold of about 50. Weeks 2–5 therefore lean on synthetic
  data for volume, with the real bodies as ground truth.
- The `STATS#cache` item has a 5-minute TTL, so it was absent when I counted. "COUNT minus one"
  would have undercounted by one.

**Exam skills exercised** (IDs, then tick them in `EXAM-COVERAGE.md`)
- 3.2.2: S3 Lifecycle retention, and keeping PII-bearing object tags off the archive copies.

**Spend this week** (Cost Explorer, filtered to the project's tags)
- $0.00 (estimated, under $0.01). 25 COPY requests, about 100 GET/HEAD requests and 239 KB of
  new storage. DynamoDB scans used provisioned capacity. Not checked in Cost Explorer, which lags
  about a day.

**Cut or deferred**
- Nothing cut.
- Deferred to week 1 (#98, #132): the SSE-C drift, the Lambda code-hash drift, and setting the
  `incoming/` expiry back to 30 days.

---

## Template

### Week N: <theme>

**Shipped**
-

**Numbers** (eval, cost, latency; with the dataset and its size)

| Metric | Before | After | Dataset (n) |
|---|---|---|---|
| | | | |

**Decisions** (link the `DECISIONS.md` entry and any status change)
-

**What broke or surprised me**
-

**Exam skills exercised** (IDs, then tick them in `EXAM-COVERAGE.md`)
-

**Spend this week** (Cost Explorer, filtered to the project's tags)
- $

**Cut or deferred**
-
