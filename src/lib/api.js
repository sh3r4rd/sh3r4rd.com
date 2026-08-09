// Base URL for the recruiter dashboard API.
//
// Production builds talk to the API's own origin directly. That API only sends
// `Access-Control-Allow-Origin: https://sh3r4rd.com` (see `cors_allowed_origin`
// in infra/recruiter-dashboard), so a browser on http://localhost:5173 has its
// responses blocked. `npm run dev` therefore sets VITE_API_BASE_URL to the
// relative `/dashboard-api` prefix (.env.development), which the Vite dev server
// proxies to the real API — same-origin from the browser's view, so CORS never
// applies. Point at any other backend with VITE_API_BASE_URL in `.env.local`.
export const API_BASE = import.meta.env.VITE_API_BASE_URL || "https://dashboard-api.sh3r4rd.com";

// Base URL for the resume-request API (the /requests endpoint the resume form
// POSTs to). This is a different API from the dashboard one above and it does
// send `Access-Control-Allow-Origin: *`, so localhost is not actually blocked —
// it is routed through the same dev proxy purely for consistency, so both APIs
// are configured in one place and neither depends on the CORS headers of a
// specific response. Same env-var mechanics as API_BASE.
export const REQUESTS_API_BASE =
  import.meta.env.VITE_REQUESTS_API_BASE_URL || "https://api.sh3r4rd.com";
