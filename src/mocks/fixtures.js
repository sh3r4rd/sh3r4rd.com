// Deterministic test fixtures for the recruiter dashboard.
//
// Shapes mirror the real API responses (see api-handler):
//   GET /recruiters -> AnonymizedItem[]:
//     { id, company, jobTitle, month: "YYYY-MM", recruiterLabel, confidence }
//   GET /stats -> { totalEmails, uniqueCompanies, byMonth, topJobTitles }

// createRecruiter builds a single anonymized recruiter record. Defaults are
// fixed (no randomness, no time) so fixtures are deterministic across runs.
// `recruiterLabel` is derived from `company` unless explicitly overridden,
// matching the server which formats it as "Recruiter at {company}".
export function createRecruiter(overrides = {}) {
  const company = overrides.company ?? 'Acme Corp'
  return {
    id: 'rec-1',
    company,
    jobTitle: 'Software Engineer',
    month: '2026-05',
    recruiterLabel: `Recruiter at ${company}`,
    confidence: 0.95,
    ...overrides,
  }
}

// A varied, deterministic dataset large enough to exercise pagination
// (PAGE_SIZE = 10), sorting, and company/job-title/month filtering.
export const RECRUITERS = [
  createRecruiter({ id: 'rec-1', company: 'Acme Corp', jobTitle: 'Software Engineer', month: '2026-05' }),
  createRecruiter({ id: 'rec-2', company: 'Acme Corp', jobTitle: 'Senior Software Engineer', month: '2026-04' }),
  createRecruiter({ id: 'rec-3', company: 'Globex', jobTitle: 'Software Engineer', month: '2026-05' }),
  createRecruiter({ id: 'rec-4', company: 'Globex', jobTitle: 'Engineering Manager', month: '2026-03' }),
  createRecruiter({ id: 'rec-5', company: 'Initech', jobTitle: 'Backend Engineer', month: '2026-02' }),
  createRecruiter({ id: 'rec-6', company: 'Initech', jobTitle: 'Software Engineer', month: '2026-05' }),
  createRecruiter({ id: 'rec-7', company: 'Umbrella', jobTitle: 'Platform Engineer', month: '2026-01' }),
  createRecruiter({ id: 'rec-8', company: 'Umbrella', jobTitle: 'Software Engineer', month: '2026-04' }),
  createRecruiter({ id: 'rec-9', company: 'Hooli', jobTitle: 'Frontend Engineer', month: '2026-03' }),
  createRecruiter({ id: 'rec-10', company: 'Hooli', jobTitle: 'Software Engineer', month: '2026-02' }),
  createRecruiter({ id: 'rec-11', company: 'Stark Industries', jobTitle: 'Staff Engineer', month: '2026-05' }),
  createRecruiter({ id: 'rec-12', company: 'Wayne Enterprises', jobTitle: 'Security Engineer', month: '2026-01' }),
]

// buildStats derives a /stats response from a recruiter list, keeping the stats
// fixture consistent with the recruiter fixture it describes.
export function buildStats(recruiters = RECRUITERS) {
  const byMonth = {}
  const byJobTitle = {}
  const companies = new Set()
  for (const r of recruiters) {
    companies.add(r.company)
    byMonth[r.month] = (byMonth[r.month] ?? 0) + 1
    byJobTitle[r.jobTitle] = (byJobTitle[r.jobTitle] ?? 0) + 1
  }
  return {
    totalEmails: recruiters.length,
    uniqueCompanies: companies.size,
    byMonth,
    topJobTitles: byJobTitle,
  }
}

export const STATS = buildStats(RECRUITERS)
