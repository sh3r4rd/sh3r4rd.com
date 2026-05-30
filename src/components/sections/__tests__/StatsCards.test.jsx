import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import StatsCards from '../StatsCards'

describe('StatsCards', () => {
  it('renders all four card labels', () => {
    render(<StatsCards stats={null} />)
    expect(screen.getByText('Total Requests')).toBeInTheDocument()
    expect(screen.getByText('Unique Companies')).toBeInTheDocument()
    expect(screen.getByText('Top Job Title')).toBeInTheDocument()
    expect(screen.getByText('This Month')).toBeInTheDocument()
  })

  it('falls back to zeros and N/A when stats is null', () => {
    render(<StatsCards stats={null} />)
    // Total Requests, Unique Companies, This Month all render 0.
    expect(screen.getAllByText('0').length).toBeGreaterThanOrEqual(3)
    expect(screen.getByText('N/A')).toBeInTheDocument()
  })

  it('falls back to zeros and N/A when stats is undefined (no prop)', () => {
    render(<StatsCards />)
    expect(screen.getByText('N/A')).toBeInTheDocument()
  })

  it('renders totals and the most frequent job title from stats', () => {
    const stats = {
      totalEmails: 42,
      uniqueCompanies: 7,
      byMonth: {},
      topJobTitles: {
        'Software Engineer': 5,
        'Engineering Manager': 9,
        'Staff Engineer': 2,
      },
    }
    render(<StatsCards stats={stats} />)
    expect(screen.getByText('42')).toBeInTheDocument()
    expect(screen.getByText('7')).toBeInTheDocument()
    // Top job title is the one with the highest count.
    expect(screen.getByText('Engineering Manager')).toBeInTheDocument()
  })

  describe('This Month (time-dependent)', () => {
    let originalTz
    beforeEach(() => {
      // Force a non-UTC zone AND pick an instant that lands in a different
      // month locally than in UTC: just past midnight UTC is still the prior
      // month in America/Los_Angeles. A local-time implementation would read
      // the wrong bucket, so this genuinely guards the UTC computation — a
      // noon-UTC instant resolves to the same month in every timezone and
      // would pass even against a local-time regression.
      originalTz = process.env.TZ
      process.env.TZ = 'America/Los_Angeles'
      vi.useFakeTimers()
      vi.setSystemTime(new Date('2026-06-01T03:00:00Z')) // UTC: Jun, LA: May 31
    })
    afterEach(() => {
      vi.useRealTimers()
      process.env.TZ = originalTz
    })

    it('reads the current UTC month bucket from byMonth (not local time)', () => {
      const stats = {
        totalEmails: 10,
        uniqueCompanies: 3,
        // Distinct values so the UTC bucket (2026-06 → 7) is distinguishable
        // from the local-time bucket a buggy impl would read (2026-05 → 99).
        byMonth: { '2026-06': 7, '2026-05': 99 },
        topJobTitles: { 'Software Engineer': 10 },
      }
      render(<StatsCards stats={stats} />)
      expect(screen.getByText('7')).toBeInTheDocument()
      expect(screen.queryByText('99')).toBeNull()
    })
  })
})
