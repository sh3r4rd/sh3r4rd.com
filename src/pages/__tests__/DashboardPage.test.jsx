import { describe, expect, it } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { http, HttpResponse } from 'msw'
import DashboardPage from '../DashboardPage'
import { server } from '../../mocks/server'
import { API_BASE } from '../../mocks/handlers'
import { createRecruiter, RECRUITERS } from '../../mocks/fixtures'

function renderDashboard() {
  return render(<DashboardPage />)
}

// Body rows of the desktop <table> (excludes the header row).
function tableBodyRows() {
  return within(screen.getByRole('table')).getAllByRole('row').slice(1)
}

describe('DashboardPage (integration)', () => {
  it('shows a loading state before data arrives', () => {
    renderDashboard()
    expect(screen.getByText('Loading recruiter data...')).toBeInTheDocument()
  })

  it('renders recruiter data once the fetch resolves', async () => {
    renderDashboard()
    await waitFor(() =>
      expect(screen.queryByText('Loading recruiter data...')).toBeNull(),
    )
    // 12 fixture rows, paginated to 10 per page.
    expect(tableBodyRows()).toHaveLength(10)
    expect(screen.getByText(`${RECRUITERS.length} results`)).toBeInTheDocument()
  })

  it('renders stats totals sourced from /stats', async () => {
    // Distinctive totals that do NOT match the /recruiters count (12) or any
    // rendered page number, so a passing assertion proves the values come from
    // the /stats response and not from the recruiter list.
    server.use(
      http.get(`${API_BASE}/stats`, () =>
        HttpResponse.json({
          totalEmails: 42,
          uniqueCompanies: 7,
          byMonth: {},
          topJobTitles: {},
        }),
      ),
    )
    renderDashboard()

    // Scope each assertion to its own stat card so a coincidental duplicate
    // elsewhere on the page (e.g. a pagination button) can't satisfy the match.
    const totalCard = (await screen.findByText('Total Requests')).closest('div')
    expect(within(totalCard).getByText('42')).toBeInTheDocument()

    const companiesCard = screen.getByText('Unique Companies').closest('div')
    expect(within(companiesCard).getByText('7')).toBeInTheDocument()
  })

  it('shows an error state and recovers via Try Again', async () => {
    let attempt = 0
    server.use(
      http.get(`${API_BASE}/recruiters`, () => {
        attempt += 1
        if (attempt === 1) {
          return new HttpResponse(null, { status: 500 })
        }
        return HttpResponse.json(RECRUITERS)
      }),
    )
    const user = userEvent.setup()
    renderDashboard()

    expect(await screen.findByText(/Request failed: 500/)).toBeInTheDocument()
    const tryAgain = screen.getByRole('button', { name: 'Try Again' })
    await user.click(tryAgain)

    // Second attempt succeeds → table renders.
    await waitFor(() => expect(screen.getByRole('table')).toBeInTheDocument())
  })

  it('handles an empty API response', async () => {
    server.use(http.get(`${API_BASE}/recruiters`, () => HttpResponse.json([])))
    renderDashboard()
    expect(
      await screen.findByText('No recruiter requests match your filters'),
    ).toBeInTheDocument()
    expect(screen.getByText('0 results')).toBeInTheDocument()
  })

  it('filters recruiters when the user types in search', async () => {
    const user = userEvent.setup()
    renderDashboard()
    await screen.findByRole('table')

    await user.type(
      screen.getByLabelText('Search company or job title'),
      'Globex',
    )

    await waitFor(() => {
      const rows = tableBodyRows()
      // Only the 2 Globex rows remain.
      expect(rows).toHaveLength(2)
    })
    expect(
      screen.getByText(/results \(filtered from 12\)/),
    ).toBeInTheDocument()
  })

  it('filters by company when the dropdown changes', async () => {
    const user = userEvent.setup()
    renderDashboard()
    await screen.findByRole('table')

    await user.selectOptions(screen.getByLabelText('Filter by company'), 'Initech')

    await waitFor(() => expect(tableBodyRows()).toHaveLength(2))
    const table = screen.getByRole('table')
    expect(within(table).queryByText('Globex')).toBeNull()
  })

  it('shows the empty state when filters match nothing', async () => {
    const user = userEvent.setup()
    renderDashboard()
    await screen.findByRole('table')

    await user.type(
      screen.getByLabelText('Search company or job title'),
      'no-such-company-xyz',
    )

    await waitFor(() =>
      expect(
        screen.getByText('No recruiter requests match your filters'),
      ).toBeInTheDocument(),
    )
  })

  it('refetches data when Refresh is clicked', async () => {
    // Return the original dataset first, then a changed dataset on the refetch,
    // so we can prove Refresh actually re-hits /recruiters and re-renders.
    let calls = 0
    server.use(
      http.get(`${API_BASE}/recruiters`, () => {
        calls += 1
        if (calls === 1) return HttpResponse.json(RECRUITERS)
        return HttpResponse.json([
          createRecruiter({
            id: 'rec-refreshed',
            company: 'Refreshed Co',
            jobTitle: 'Refreshed Role',
            month: '2026-05',
          }),
        ])
      }),
    )
    const user = userEvent.setup()
    renderDashboard()

    const table = await screen.findByRole('table')
    // The refreshed company is absent from the initial dataset.
    expect(within(table).queryByText('Refreshed Co')).toBeNull()

    await user.click(screen.getByRole('button', { name: /Refresh/ }))

    // The changed dataset from the second fetch now renders: a single row whose
    // company is scoped to the table (it also appears in the mobile-card view).
    await waitFor(() => expect(tableBodyRows()).toHaveLength(1))
    expect(
      within(screen.getByRole('table')).getByText('Refreshed Co'),
    ).toBeInTheDocument()
    expect(calls).toBe(2)
  })
})
