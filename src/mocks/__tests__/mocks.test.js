import { describe, it, expect } from 'vitest'
import { createRecruiter, RECRUITERS, STATS, buildStats } from '../fixtures'
import { API_BASE } from '../handlers'

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
