import { describe, expect, it, vi } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import RecruiterTable from '../RecruiterTable'

function makeRows(n) {
  return Array.from({ length: n }, (_, i) => ({
    id: `r${i + 1}`,
    company: `Company ${String(i + 1).padStart(2, '0')}`,
    jobTitle: 'Software Engineer',
    recruiterLabel: `Recruiter at Company ${i + 1}`,
    month: `2026-${String((i % 12) + 1).padStart(2, '0')}`,
  }))
}

describe('RecruiterTable', () => {
  it('shows the empty state when there is no data', () => {
    render(<RecruiterTable data={[]} page={1} onPageChange={vi.fn()} />)
    expect(
      screen.getByText('No recruiter requests match your filters'),
    ).toBeInTheDocument()
  })

  it('renders a row per record in the desktop table', () => {
    const data = makeRows(3)
    render(<RecruiterTable data={data} page={1} onPageChange={vi.fn()} />)
    const table = screen.getByRole('table')
    const bodyRows = within(table).getAllByRole('row').slice(1) // drop header row
    expect(bodyRows).toHaveLength(3)
  })

  it('falls back to "Anonymous" when recruiterLabel is missing', () => {
    const data = [
      { id: 'x', company: 'Acme', jobTitle: 'SE', recruiterLabel: '', month: '2026-05' },
    ]
    render(<RecruiterTable data={data} page={1} onPageChange={vi.fn()} />)
    expect(screen.getAllByText('Anonymous').length).toBeGreaterThanOrEqual(1)
  })

  it('defaults to sorting by month descending', () => {
    const data = [
      { id: 'a', company: 'A', jobTitle: 'x', recruiterLabel: 'l', month: '2026-01' },
      { id: 'b', company: 'B', jobTitle: 'x', recruiterLabel: 'l', month: '2026-12' },
    ]
    render(<RecruiterTable data={data} page={1} onPageChange={vi.fn()} />)
    const monthHeader = screen.getByRole('columnheader', { name: /Month/ })
    expect(monthHeader).toHaveAttribute('aria-sort', 'descending')
    const table = screen.getByRole('table')
    const firstRow = within(table).getAllByRole('row')[1]
    expect(within(firstRow).getByText('2026-12')).toBeInTheDocument()
  })

  it('sorts ascending on first click of a new column and toggles on repeat', async () => {
    const user = userEvent.setup()
    const data = [
      { id: 'a', company: 'Zeta', jobTitle: 'x', recruiterLabel: 'l', month: '2026-05' },
      { id: 'b', company: 'Alpha', jobTitle: 'x', recruiterLabel: 'l', month: '2026-04' },
    ]
    render(<RecruiterTable data={data} page={1} onPageChange={vi.fn()} />)

    const companyHeaderBtn = screen.getByRole('button', { name: /Company/ })
    await user.click(companyHeaderBtn)

    let companyHeader = screen.getByRole('columnheader', { name: /Company/ })
    expect(companyHeader).toHaveAttribute('aria-sort', 'ascending')
    let rows = within(screen.getByRole('table')).getAllByRole('row').slice(1)
    expect(within(rows[0]).getByText('Alpha')).toBeInTheDocument()

    // Click again → descending.
    await user.click(companyHeaderBtn)
    companyHeader = screen.getByRole('columnheader', { name: /Company/ })
    expect(companyHeader).toHaveAttribute('aria-sort', 'descending')
    rows = within(screen.getByRole('table')).getAllByRole('row').slice(1)
    expect(within(rows[0]).getByText('Zeta')).toBeInTheDocument()
  })

  it('paginates: only pageSize rows render and page controls appear', () => {
    const data = makeRows(25)
    render(<RecruiterTable data={data} page={1} pageSize={10} onPageChange={vi.fn()} />)
    const bodyRows = within(screen.getByRole('table')).getAllByRole('row').slice(1)
    expect(bodyRows).toHaveLength(10)
    expect(screen.getByText('Showing 1-10 of 25')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Next' })).toBeInTheDocument()
  })

  it('calls onPageChange when a page control is clicked', async () => {
    const user = userEvent.setup()
    const onPageChange = vi.fn()
    const data = makeRows(25)
    render(<RecruiterTable data={data} page={1} pageSize={10} onPageChange={onPageChange} />)
    await user.click(screen.getByRole('button', { name: 'Next' }))
    expect(onPageChange).toHaveBeenCalledWith(2)
    await user.click(screen.getByRole('button', { name: 'Go to page 3' }))
    expect(onPageChange).toHaveBeenCalledWith(3)
  })

  it('disables Previous on the first page', () => {
    const data = makeRows(25)
    render(<RecruiterTable data={data} page={1} pageSize={10} onPageChange={vi.fn()} />)
    expect(screen.getByRole('button', { name: 'Previous' })).toBeDisabled()
  })

  it('clamps an out-of-range page down to the last valid page', () => {
    const data = makeRows(25)
    // page 99 is past the end; component clamps to page 3 (rows 21-25).
    render(<RecruiterTable data={data} page={99} pageSize={10} onPageChange={vi.fn()} />)
    expect(screen.getByText('Showing 21-25 of 25')).toBeInTheDocument()
  })
})
