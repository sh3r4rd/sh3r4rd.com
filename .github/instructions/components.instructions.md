---
applyTo: "src/components/**,src/pages/**"
---

# Component Instructions

## File Naming

- Use PascalCase for pages, layout, and section component files (e.g., `SkillsSection.jsx`, `HomePage.jsx`)
- Use lowercase filenames for UI primitives in `components/ui/` (e.g., `button.jsx`, `card.jsx`)
- All files use `.jsx` extension

## Dark Mode

Always pair light and dark classes:

- Text: `text-gray-700 dark:text-gray-300` or `text-gray-900 dark:text-white`
- Backgrounds: `bg-white dark:bg-gray-800`
- Borders: `border-gray-200 dark:border-gray-700`
- Emphasis: `font-semibold dark:text-white`

## Component Categories

### `components/ui/` — Reusable primitives
- Use **named exports** (e.g., `export function Button`)
- Keep generic — no business logic, no data fetching
- Accept `className` prop for composition when appropriate

### `components/layout/` — Structural components
- Use **default exports**
- Examples: `Header`, `Breadcrumbs`, `PageTracker`

### `components/sections/` — Content sections
- Use **default exports**
- Self-contained sections displayed within pages (e.g., `SkillsSection`)

### `pages/` — Route-level components
- Use **default exports**
- Wrap in `<section className="max-w-4xl mx-auto space-y-12">` and render `<Breadcrumbs />` then `<Header />` first (exception: special pages like `NotFoundPage` use a minimal layout)

## Avoid

- Inline `style` attributes (use Tailwind classes)
- CSS modules or separate CSS files
- Component-level data fetching (fetch in pages, pass as props)
- Creating wrapper components that add no value
