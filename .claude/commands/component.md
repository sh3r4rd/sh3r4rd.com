Create a new React component based on the type argument: $ARGUMENTS

Determine the component type from the argument. Valid types: `ui`, `layout`, `section`, `page`.

Follow these rules based on type:

## `ui` — `src/components/ui/<name>.jsx`
- Use a **named export**: `export function ComponentName`
- Keep it generic with no business logic
- Accept a `className` prop if it renders a wrapper element
- Example: `export function Badge({ children, className = "" }) { ... }`

## `layout` — `src/components/layout/<Name>.jsx`
- Use a **default export**: `export default function ComponentName`
- Structural components used across pages

## `section` — `src/components/sections/<Name>.jsx`
- Use a **default export**: `export default function ComponentName`
- Self-contained content section rendered inside a page

## `page` — `src/pages/<Name>.jsx`
- Use a **default export**: `export default function PageName`
- Wrap content in `<section className="max-w-4xl mx-auto space-y-12">`
- Start with `<Breadcrumbs />` and `<Header />`:

```jsx
import Breadcrumbs from "../components/layout/Breadcrumbs";
import Header from "../components/layout/Header";

export default function PageName() {
  return (
    <section className="max-w-4xl mx-auto space-y-12">
      <Breadcrumbs />
      <Header />
      {/* page content here */}
    </section>
  );
}
```

- After creating the page, remind the user to add a route in `src/App.jsx`

## General Rules
- Always use `.jsx` extension
- Always include `dark:` Tailwind variants for colors
- Use `lucide-react` for icons
