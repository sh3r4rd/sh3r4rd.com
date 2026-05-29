# [1.1.0](https://github.com/sh3r4rd/sh3r4rd.com/compare/v1.0.2...v1.1.0) (2026-05-29)


### Features

* **dashboard:** recruiter dashboard frontend (Phase 4) ([#84](https://github.com/sh3r4rd/sh3r4rd.com/issues/84)) ([5820b5f](https://github.com/sh3r4rd/sh3r4rd.com/commit/5820b5fb0024a2bc453158585420e43413f40802)), closes [#80](https://github.com/sh3r4rd/sh3r4rd.com/issues/80) [#81](https://github.com/sh3r4rd/sh3r4rd.com/issues/81) [#82](https://github.com/sh3r4rd/sh3r4rd.com/issues/82) [#83](https://github.com/sh3r4rd/sh3r4rd.com/issues/83) [#34](https://github.com/sh3r4rd/sh3r4rd.com/issues/34) [#80](https://github.com/sh3r4rd/sh3r4rd.com/issues/80)

## [1.0.2](https://github.com/sh3r4rd/sh3r4rd.com/compare/v1.0.1...v1.0.2) (2026-03-23)


### Performance Improvements

* **lambda:** replace full table scan in /stats with TTL-cached aggregation ([#79](https://github.com/sh3r4rd/sh3r4rd.com/issues/79)) ([82ddb10](https://github.com/sh3r4rd/sh3r4rd.com/commit/82ddb10f30569faa4436f9e2557c2b4d77608954)), closes [STATS#cache](https://github.com/STATS/issues/cache)

## [1.0.1](https://github.com/sh3r4rd/sh3r4rd.com/compare/v1.0.0...v1.0.1) (2026-03-22)


### Performance Improvements

* **lambda:** replace in-memory company filtering with DynamoDB FilterExpression ([#78](https://github.com/sh3r4rd/sh3r4rd.com/issues/78)) ([d7ff7ab](https://github.com/sh3r4rd/sh3r4rd.com/commit/d7ff7ab42e6d11f9e8837c6da6d6763d39597a39))

# 1.0.0 (2026-03-16)


### Bug Fixes

* **ci:** use PAT for checkout to bypass branch ruleset ([#67](https://github.com/sh3r4rd/sh3r4rd.com/issues/67)) ([d99dc57](https://github.com/sh3r4rd/sh3r4rd.com/commit/d99dc57a1926479ac879bf96a17af048329a2770)), closes [hi#privilege](https://github.com/hi/issues/privilege)
* **ci:** use PAT for semantic-release to bypass branch protection ([#66](https://github.com/sh3r4rd/sh3r4rd.com/issues/66)) ([761ac48](https://github.com/sh3r4rd/sh3r4rd.com/commit/761ac48172ee624ba018529c6c4a5eefbe8c1f08)), closes [hi#privilege](https://github.com/hi/issues/privilege)
* **ci:** use SSH deploy key to bypass branch ruleset ([#69](https://github.com/sh3r4rd/sh3r4rd.com/issues/69)) ([dbf730c](https://github.com/sh3r4rd/sh3r4rd.com/commit/dbf730ca6f789e41b84f6c0232f801f4c590914d))
* update aws cli installation ([#10](https://github.com/sh3r4rd/sh3r4rd.com/issues/10)) ([899780c](https://github.com/sh3r4rd/sh3r4rd.com/commit/899780c1794189998f0f83671b5d068ad538141d))
* update header subtitle to replace Kafka with GCP ([#14](https://github.com/sh3r4rd/sh3r4rd.com/issues/14)) ([1b7e39c](https://github.com/sh3r4rd/sh3r4rd.com/commit/1b7e39cc4c722c579b378e43543dc5e5181c6639))


### Features

* add favicon and apple touch icon tags ([#13](https://github.com/sh3r4rd/sh3r4rd.com/issues/13)) ([ffaf223](https://github.com/sh3r4rd/sh3r4rd.com/commit/ffaf223e830ae2e72399e559de90ea9cced9d6fe))
* Recruiter Dashboard — Phases 1 & 2 + CI/CD pipeline ([#65](https://github.com/sh3r4rd/sh3r4rd.com/issues/65)) ([be7dca1](https://github.com/sh3r4rd/sh3r4rd.com/commit/be7dca12cb7e1a41537f7913476e8a849558966d)), closes [#22](https://github.com/sh3r4rd/sh3r4rd.com/issues/22) [#45](https://github.com/sh3r4rd/sh3r4rd.com/issues/45) [#45](https://github.com/sh3r4rd/sh3r4rd.com/issues/45)
