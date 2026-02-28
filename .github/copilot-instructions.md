# Copilot Instructions

## Project

React 19 portfolio site using Vite 7, Tailwind CSS 3, and JSX (no TypeScript). Deployed to AWS S3/CloudFront via GitHub Actions.

## Commands

- Dev server: `npm run dev`
- Build: `npm run build`
- Lint: `npm run lint`
- Deploy: `make deploy bucket=<bucket-name>` (builds then syncs to S3)

## Code Rules

- Use `.jsx` files only — never create `.ts` or `.tsx` files
- Style with Tailwind utility classes only — no custom CSS, no CSS modules, no inline `style` attributes
- Always provide `dark:` variants for colors (e.g., `text-gray-700 dark:text-gray-300`, `bg-white dark:bg-gray-800`)
- Use icons exclusively from `lucide-react`
- Use conventional commits (e.g., `feat:`, `fix:`, `docs:`, `chore:`)

## Component Structure

Pages follow this layout pattern:

```jsx
import Breadcrumbs from "../components/layout/Breadcrumbs";
import Header from "../components/layout/Header";

export default function MyPage() {
  return (
    <section className="max-w-4xl mx-auto space-y-12">
      <Breadcrumbs />
      <Header />
      {/* page content */}
    </section>
  );
}
```

## UI Primitives

Use existing UI components from `components/ui/`:

```jsx
import { Button } from "../components/ui/button";
import { Card, CardContent } from "../components/ui/card";
```

- `Button` accepts `size` prop: `"sm"`, `"md"` (default), `"lg"`
- `Card` wraps content; use `CardContent` inside with `className="p-4"`
- UI components use **named exports**; all other components use **default exports**

## Forms

- Use DOM-based validation (access fields via `form.elements[id]`)
- Include a honeypot field: `<input name="zip" style={{ display: 'none' }} type="text" />`
- API endpoint: `POST https://api.sh3r4rd.com/requests`
- Payload fields: `firstName`, `lastName`, `email`, `phone`, `company`, `jobTitle`, `description`

## Commits

Follow conventional commits format:

```
type(optional-scope): short description

Optional body with more detail.
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `ci`

## Boundaries

**Never:**
- Create `.ts` or `.tsx` files
- Add new icon libraries (use `lucide-react`)
- Add state management libraries (Redux, Zustand, etc.)
- Import or extend `src/App.css` or `src/assets/react.svg` (unused Vite scaffolding)

**Ask before:**
- Adding new npm dependencies
- Creating new routes in `App.jsx`
- Modifying the deployment workflow (`.github/workflows/deploy.yml`)
- Changing the API endpoint or payload structure
