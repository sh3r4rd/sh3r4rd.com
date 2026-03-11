# Codex Instructions

This file provides repository-specific guidance to Codex CLI when working in this project.

## Scope

- Keep this file Codex-specific. Do not treat it as the source of truth for Claude Code or Copilot.
- Preserve existing guidance in [CLAUDE.md](/Users/sherardbailey/Code/sh3r4rd.com/CLAUDE.md) and [.github/copilot-instructions.md](/Users/sherardbailey/Code/sh3r4rd.com/.github/copilot-instructions.md).

## Project Overview

- Personal portfolio site built with React, Tailwind CSS, and Vite
- Deployed to AWS S3 and CloudFront via GitHub Actions on pushes to `main`
- Resume request form posts to `https://api.sh3r4rd.com/requests`

## Commands

- Dev server: `npm run dev`
- Build: `npm run build`
- Lint: `npm run lint`
- Preview production build: `npm run preview`
- Manual deploy: `make deploy bucket=<bucket-name>`

## Architecture

- React 19 with `react-router-dom` v7
- Vite 7 build pipeline
- Tailwind CSS 3 with the typography plugin
- ESLint 9 flat config

## Source Layout

- `src/App.jsx`: router entrypoint
- `src/pages/`: route-level pages
- `src/components/layout/`: structural layout components
- `src/components/sections/`: page sections
- `src/components/ui/`: reusable UI primitives

## Repository Conventions

- Use `.jsx` files only. Do not add `.ts` or `.tsx` files.
- Prefer Tailwind utility classes. Do not add CSS modules or inline `style` attributes.
- Pair light and dark mode classes for user-facing UI changes.
- Use icons from `lucide-react`.
- `components/ui/` uses named exports. Other components use default exports.
- Page components usually render `<Breadcrumbs />` and `<Header />` inside `<section className="max-w-4xl mx-auto space-y-12">`.
- Follow conventional commits.

## Boundaries

- Do not import or extend `src/App.css` or `src/assets/react.svg`. They are unused Vite scaffolding.
- Ask before adding dependencies, changing routes, changing deployment workflow, or changing the resume request API contract.
- Avoid broad refactors unrelated to the task.

## Validation

- Run `npm run lint` after non-trivial code changes.
- Run `npm run build` when changes could affect bundling, routing, or production output.
- If you cannot run a validation step, say so clearly.
