import { describe, expect, it } from 'vitest'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { http, HttpResponse } from 'msw'
import DashboardPage from '../DashboardPage'
import { server } from '../../mocks/server'
import { API_BASE } from '../../mocks/handlers'
import { RECRUITERS } from '../../mocks/fixtures'

// DashboardPage renders Breadcrumbs (useLocation), so a router is required.
function renderDashboard() {
  return render(
    <MemoryRouter initialEntries={['/dashboard']}>
      <DashboardPage />
    </MemoryRouter>,
  )
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

  it('renders the correct stats totals from /stats', async () => {
    renderDashboard()
    expect(await screen.findByText('Total Requests')).toBeInTheDocument()
    // totalEmails === 12 unique enough to assert directly.
    expect(screen.getByText(String(RECRUITERS.length))).toBeInTheDocument()
    expect(screen.getByText('Unique Companies')).toBeInTheDocument()
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

  it('refreshes data when Refresh is clicked', async () => {
    const user = userEvent.setup()
    renderDashboard()
    await screen.findByRole('table')
    await user.click(screen.getByRole('button', { name: /Refresh/ }))
    // Table remains after a successful refresh.
    await waitFor(() => expect(screen.getByRole('table')).toBeInTheDocument())
  })
})
