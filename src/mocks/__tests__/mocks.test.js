import { describe, it, expect } from 'vitest'
import { createRecruiter, RECRUITERS, STATS, buildStats } from '../fixtures'
import { API_BASE } from '../handlers'
import { REQUESTS_API_BASE } from '../../lib/api'

// Validates the test infrastructure itself (issue #38): deterministic fixtures
// and a live MSW server (started by setup.js) that serves the mock endpoints.
describe('fixtures', () => {
  it('createRecruiter produces deterministic, complete records', () => {
    expect(createRecruiter()).toEqual(createRecruiter())
    expect(createRecruiter()).toEqual({
      id: 'rec-1',
      company: 'Acme Corp',
      jobTitle: 'Software Engineer',
      month: '2026-05',
      recruiterLabel: 'Recruiter at Acme Corp',
      confidence: 0.95,
    })
  })

  it('createRecruiter derives recruiterLabel from company', () => {
    expect(createRecruiter({ company: 'Globex' }).recruiterLabel).toBe(
      'Recruiter at Globex',
    )
  })

  it('createRecruiter honors explicit overrides', () => {
    const r = createRecruiter({ id: 'x', recruiterLabel: 'Anonymous' })
    expect(r.id).toBe('x')
    expect(r.recruiterLabel).toBe('Anonymous')
  })

  it('RECRUITERS dataset has unique ids and is large enough to paginate', () => {
    const ids = RECRUITERS.map((r) => r.id)
    expect(new Set(ids).size).toBe(ids.length)
    expect(RECRUITERS.length).toBeGreaterThan(10)
  })

  it('buildStats stays consistent with the recruiter list it describes', () => {
    expect(STATS).toEqual(buildStats(RECRUITERS))
    expect(STATS.totalEmails).toBe(RECRUITERS.length)
    expect(STATS.uniqueCompanies).toBe(
      new Set(RECRUITERS.map((r) => r.company)).size,
    )
  })
})

describe('MSW server', () => {
  it('serves the recruiters fixture for GET /recruiters', async () => {
    const res = await fetch(`${API_BASE}/recruiters`)
    expect(res.ok).toBe(true)
    await expect(res.json()).resolves.toEqual(RECRUITERS)
  })

  it('serves the stats fixture for GET /stats', async () => {
    const res = await fetch(`${API_BASE}/stats`)
    expect(res.ok).toBe(true)
    await expect(res.json()).resolves.toEqual(STATS)
  })
})

// The mocks derive their base from src/lib/api.js so they can never drift from
// the URLs the app actually calls. That removes the guard the old duplicated
// literal gave us: every handler would follow a wrong base in lockstep and the
// suite would stay green while production was broken. These assertions restore
// it by pinning the values themselves rather than re-stating them in the mocks.
describe('API base URLs', () => {
  it('resolve to the production origins under test/production mode', () => {
    // A dev-only proxy prefix reaching here means .env.development leaked out of
    // mode=development — the exact failure that would ship a relative URL to S3.
    expect(API_BASE).toBe('https://dashboard-api.sh3r4rd.com')
    expect(REQUESTS_API_BASE).toBe('https://api.sh3r4rd.com')
  })
})
