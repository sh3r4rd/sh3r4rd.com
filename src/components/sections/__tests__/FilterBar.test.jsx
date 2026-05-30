import { describe, expect, it, vi } from 'vitest'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import FilterBar from '../FilterBar'
import { EMPTY_FILTERS } from '../../../lib/filters'

const SAMPLE = [
  { id: '1', company: 'Globex', jobTitle: 'Software Engineer', month: '2026-05' },
  { id: '2', company: 'Acme Corp', jobTitle: 'Engineering Manager', month: '2026-04' },
  { id: '3', company: 'Acme Corp', jobTitle: 'Software Engineer', month: '2026-03' },
]

function renderBar(filters = EMPTY_FILTERS, onFilterChange = vi.fn()) {
  render(
    <FilterBar allData={SAMPLE} filters={filters} onFilterChange={onFilterChange} />,
  )
  return { onFilterChange }
}

describe('FilterBar', () => {
  it('renders search input and the two dropdowns', () => {
    renderBar()
    expect(screen.getByLabelText('Search company or job title')).toBeInTheDocument()
    expect(screen.getByLabelText('Filter by company')).toBeInTheDocument()
    expect(screen.getByLabelText('Filter by job title')).toBeInTheDocument()
  })

  it('derives sorted, de-duplicated company options from allData', () => {
    renderBar()
    const select = screen.getByLabelText('Filter by company')
    const options = [...select.querySelectorAll('option')].map((o) => o.textContent)
    expect(options).toEqual(['All companies', 'Acme Corp', 'Globex'])
  })

  it('derives sorted, de-duplicated job title options from allData', () => {
    renderBar()
    const select = screen.getByLabelText('Filter by job title')
    const options = [...select.querySelectorAll('option')].map((o) => o.textContent)
    expect(options).toEqual(['All job titles', 'Engineering Manager', 'Software Engineer'])
  })

  it('pushes a debounced search change up to the parent', async () => {
    const user = userEvent.setup()
    const { onFilterChange } = renderBar()
    await user.type(screen.getByLabelText('Search company or job title'), 'eng')
    await waitFor(() =>
      expect(onFilterChange).toHaveBeenCalledWith(
        expect.objectContaining({ search: 'eng' }),
      ),
    )
  })

  it('emits a company filter change immediately on select', async () => {
    const user = userEvent.setup()
    const { onFilterChange } = renderBar()
    await user.selectOptions(screen.getByLabelText('Filter by company'), 'Globex')
    expect(onFilterChange).toHaveBeenCalledWith(
      expect.objectContaining({ company: 'Globex' }),
    )
  })

  it('emits a job title filter change immediately on select', async () => {
    const user = userEvent.setup()
    const { onFilterChange } = renderBar()
    await user.selectOptions(
      screen.getByLabelText('Filter by job title'),
      'Software Engineer',
    )
    expect(onFilterChange).toHaveBeenCalledWith(
      expect.objectContaining({ jobTitle: 'Software Engineer' }),
    )
  })

  it('emits month range changes', () => {
    const { onFilterChange } = renderBar()
    // A native <input type="month"> is a segmented control; character-by-
    // character typing via userEvent is unreliable in jsdom (it can fire
    // onChange with partial/empty values). Set the value directly to mirror a
    // real month pick.
    fireEvent.change(screen.getByLabelText('From month'), {
      target: { value: '2026-01' },
    })
    expect(onFilterChange).toHaveBeenCalledWith(
      expect.objectContaining({ monthFrom: '2026-01' }),
    )
  })

  it('hides the Clear Filters button when no filters are active', () => {
    renderBar()
    expect(screen.queryByRole('button', { name: 'Clear Filters' })).toBeNull()
  })

  it('shows Clear Filters when a filter is active and resets on click', async () => {
    const user = userEvent.setup()
    const onFilterChange = vi.fn()
    render(
      <FilterBar
        allData={SAMPLE}
        filters={{ ...EMPTY_FILTERS, company: 'Globex' }}
        onFilterChange={onFilterChange}
      />,
    )
    const clear = screen.getByRole('button', { name: 'Clear Filters' })
    await user.click(clear)
    expect(onFilterChange).toHaveBeenCalledWith(EMPTY_FILTERS)
  })

  it('tolerates a missing/empty allData list', () => {
    render(<FilterBar allData={undefined} onFilterChange={vi.fn()} />)
    // Only the placeholder options exist.
    expect(
      [...screen.getByLabelText('Filter by company').querySelectorAll('option')].length,
    ).toBe(1)
  })
})
