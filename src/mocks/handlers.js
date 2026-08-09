import { http, HttpResponse } from 'msw'
import { RECRUITERS, STATS } from './fixtures'
import { API_BASE } from '../lib/api'

// Re-exported so tests keep importing the base from here. It comes from the app
// itself, so mocks can never drift from the URLs the dashboard actually calls.
// Under Vitest (mode=test) VITE_API_BASE_URL is unset, so this is the real API
// origin — the dev-only proxy prefix never reaches tests.
// Note: the dashboard API lives on its own subdomain; the resume form's
// /requests endpoint is a separate API on api.sh3r4rd.com and is not mocked here.
export { API_BASE }

// Default happy-path handlers. Tests override these per-case with
// `server.use(...)` to simulate errors, empty results, or custom datasets.
export const handlers = [
  http.get(`${API_BASE}/recruiters`, () => HttpResponse.json(RECRUITERS)),
  http.get(`${API_BASE}/stats`, () => HttpResponse.json(STATS)),
]
