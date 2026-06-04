import { http, HttpResponse } from 'msw'
import { RECRUITERS, STATS } from './fixtures'

// Base URL of the production API the dashboard talks to. The dashboard
// hardcodes this origin (no env var), so the mocks target it directly.
// Note: the dashboard API lives on its own subdomain; the resume form's
// /requests endpoint is a separate API on api.sh3r4rd.com and is not mocked here.
export const API_BASE = 'https://dashboard-api.sh3r4rd.com'

// Default happy-path handlers. Tests override these per-case with
// `server.use(...)` to simulate errors, empty results, or custom datasets.
export const handlers = [
  http.get(`${API_BASE}/recruiters`, () => HttpResponse.json(RECRUITERS)),
  http.get(`${API_BASE}/stats`, () => HttpResponse.json(STATS)),
]
