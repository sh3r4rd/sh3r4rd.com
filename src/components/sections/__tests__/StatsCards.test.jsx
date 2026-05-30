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
    beforeEach(() => {
      vi.useFakeTimers()
      // Fixed instant so the UTC "YYYY-MM" key is deterministic.
      vi.setSystemTime(new Date('2026-05-15T12:00:00Z'))
    })
    afterEach(() => {
      vi.useRealTimers()
    })

    it('reads the current UTC month bucket from byMonth', () => {
      const stats = {
        totalEmails: 10,
        uniqueCompanies: 3,
        byMonth: { '2026-05': 4, '2026-04': 6 },
        topJobTitles: { 'Software Engineer': 10 },
      }
      render(<StatsCards stats={stats} />)
      // "This Month" → 2026-05 → 4
      expect(screen.getByText('4')).toBeInTheDocument()
    })
  })
})
