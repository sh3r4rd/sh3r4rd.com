# Journal

One entry per week, written at the end of the week, in 15 minutes. This is the raw material for
the week-12 case study, and the numbers in it are the case study's credibility. Record numbers
and decisions here, not feelings.

**Never paste real email content, recruiter names or addresses here.** This file is committed.
Use aggregate numbers and synthetic examples only (D-06).

---

## Week 0: inventory

- Date: 21 November 2026 (UTC); 20 November local
- Raw bodies preserved (`archive/raw/` object count): 25 (239,397 bytes). This equals the
  `incoming/` count at copy time, the key sets are identical, and none of the 25 copies carries a
  tag.
- DynamoDB rows (all-time extractions): 92. This excludes `STATS#` items; the `STATS#cache` item
  was absent at count time.
- Bodies permanently lost (rows − bodies, approximate): 67, which is exact. All 92 rows record an
  `s3_key`, the 92 keys are distinct (no two rows share a body), and 25 of them still exist. The oldest surviving body was received on
  22 October 2026.
- Lifecycle change applied (date, `terraform plan` summary): 21 November 2026, 04:31 UTC.
  `Plan: 0 to add, 1 to change, 0 to destroy.` The only change was
  `module.s3.aws_s3_bucket_lifecycle_configuration.email_storage`, with expiration `days` going
  from 30 to 3650 on `incoming/`. It was applied with `-target` from a saved plan, and
  `make tf-vars-push` followed.

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
- The `STATS#cache` item was absent when I counted. Its 5-minute window is the api-handler's own
  freshness check, and DynamoDB's TTL deletes it in the background, so its presence at any moment
  is unpredictable. "COUNT minus one" would have undercounted by one.

**Exam skills exercised** (IDs, then tick them in `EXAM-COVERAGE.md`)
- 3.2.2: S3 Lifecycle retention, and keeping PII-bearing object tags off the archive copies.

**Spend this week** (Cost Explorer, filtered to the project's tags)
- $0.00 (estimated, under $0.01). 25 COPY requests, about 100 GET/HEAD requests and 239 KB of
  new storage. DynamoDB scans used provisioned capacity. Not checked in Cost Explorer, which lags
  about a day.

**Cut or deferred**
- Nothing cut.
- Deferred to week 1 (#132, blocked by #98):
  - The SSE-C drift, which is #98's job.
  - The Lambda code-hash drift. Before any untargeted `terraform apply`, rebuild both artifacts
    from a clean checkout of `main` with `make build-lambdas`. Otherwise the apply redeploys
    whatever `bootstrap` binaries are on the machine running it.
  - Setting the `incoming/` expiry back to 30 days. `archive/raw/` was a one-time copy on
    20 November, and `archive/clean/` holds cleaned JSON, not raw bodies. Right before lowering
    the expiry, re-run the `incoming/` → `archive/raw/` sync (`--copy-props metadata-directive`)
    and re-check that the new copies carry no tags, or raw bodies received after 20 November
    expire for good.

---

## Week 1: Macie sensitive-data discovery (stretch, #136)

Epic #119 · task #136. A one-time Amazon Macie job over the archive, run from the CLI, to compare
with Comprehend PII detection in week 4. Aggregates only.

- **Date:** 21 November 2026 (UTC). Macie was enabled at 19:43, the job ran from 19:43 to 19:55,
  and Macie was disabled at 19:57.
- **Scope:** one bucket, objects whose key starts with `archive/`, 100% sampling. At scan time
  that meant 25 raw `.eml` objects in `archive/raw/` (239,397 bytes), all SSE-S3 (AES256).
  `archive/clean/` didn't exist yet.
- **Identifier selection:** `ALL` managed data identifiers (166 at run time), with no custom
  identifiers and no allow lists. The `RECOMMENDED` set leaves out `NAME`, `ADDRESS` and
  `PHONE_NUMBER`, which are exactly the types that overlap with Comprehend's PII entities. With
  `RECOMMENDED`, the comparison would have had nothing in it. At 239 KB the larger set costs
  nothing extra.
- **Job statistics:** `ONE_TIME`, 1 run, status `COMPLETE`, last-run error code `NONE`. The job's
  CloudWatch log holds only the created, started and completed events, with no object-level
  errors, so no object was skipped. 20 of the 25 objects produced a finding (194,336 bytes, all
  classified as `message/rfc822`, result status `COMPLETE`), and 5 produced none.
- **Findings by type:** 20 × `SensitiveData:S3Object/Personal` (10 Low, 10 Medium severity).
  Policy findings for the bucket: 0. There were no other finding types.

| Category | Detection type | Objects flagged (of 25) | Occurrences |
|---|---|---|---|
| `PERSONAL_INFORMATION` | `NAME` | 17 | 35 |
| `PERSONAL_INFORMATION` | `ADDRESS` | 3 | 4 |
| `PERSONAL_INFORMATION` | `PHONE_NUMBER` | 3 | 3 |
| **Total** | | **20** (distinct) | **42** |

There were no credentials or financial detections, and nothing in any other category. Occurrences
per flagged object ranged from 1 to 6.

- **Cost:** $0.00. Macie's `get-usage-totals` reported an estimated $0 for discovery jobs,
  automated discovery, bucket inventory and object monitoring. At $1/GB, 239 KB would be about
  $0.0002 even without the free tier, plus a few dozen S3 GET and LIST requests.
- **Macie disabled:** `disable-macie` at 19:57 UTC, and afterwards `get-macie-session` returns
  `AccessDeniedException: Macie is not enabled`. Disabling Macie deleted the findings and the job
  record, which is why the aggregates above were recorded first. The service-linked role
  `AWSServiceRoleForAmazonMacie` stays in the account. It has no cost, and Macie reuses it if it
  is ever re-enabled.
- **Automated discovery:** it read `DISABLED` straight after `enable-macie`, then turned itself
  on asynchronously about 5 minutes later (`firstEnabledAt` 19:48 UTC). I set it back to
  `DISABLED` at 19:57, before disabling Macie. It ran for about 9 minutes and cost $0. It
  produced no findings: the account's 20 findings all carried this job's ID. Lesson: check
  automated discovery again a few minutes after enabling Macie, not only straight after.

**Notes**
- **Re-running after #132.** After #132 the archive is SSE-KMS with a customer-managed key, and
  Macie can't read those objects unless the key policy lets its service-linked role
  (`AWSServiceRoleForAmazonMacie`) use the key (`kms:Decrypt`). Without that grant, a re-run
  reports the objects as unclassifiable rather than scanning them. The same applies to `evals/`,
  which uses the same key.
- **What to compare against Comprehend in week 4.** Macie finds sensitive data **at rest in
  S3**. It parses the whole stored object (headers and every MIME part) and reports, for each
  object, its types and counts. Comprehend finds PII **in text you pass it**: only the cleaned
  body we send, with a confidence score and offsets for each entity. Points to check:
  - Does Comprehend find `NAME`, `ADDRESS` and `PHONE` in the same 17, 3 and 3 emails, and how
    many more or fewer occurrences?
  - Email addresses. Every `.eml` has them in its headers, but Macie has no managed identifier
    for email addresses, so it reported none. Comprehend has an `EMAIL` entity, so its count is
    a capability gap, not a disagreement. The same goes for `URL` and `IP_ADDRESS`.
  - The 5 emails Macie flagged nothing in: does Comprehend agree?
  - What it costs per email for each service at this volume.

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
