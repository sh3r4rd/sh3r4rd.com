# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Personal portfolio website (sh3r4rd.com) built with React, Tailwind CSS, and Vite. Deployed to AWS S3/CloudFront via GitHub Actions on push to `main`.

## Commands

- **Dev server:** `npm run dev` (or `make server`)
- **Build:** `npm run build` (or `make build`)
- **Lint:** `npm run lint`
- **Deploy (manual):** `make deploy bucket=<bucket-name>` (builds then syncs to S3)
- **Preview production build:** `npm run preview`

## Architecture

- **React 19** with **react-router-dom v7** for client-side routing (BrowserRouter)
- **Vite 7** as build tool with React plugin
- **Tailwind CSS 3** with typography plugin, PostCSS/Autoprefixer
- **ESLint 9** flat config with react-hooks and react-refresh plugins

### Source Structure (`src/`)

- `App.jsx` — Root component with router; routes: `/` (HomePage), `/resume` (ResumeRequestPage), `*` (NotFoundPage)
- `pages/` — Page-level route components
- `components/layout/` — Header, Breadcrumbs, PageTracker (analytics)
- `components/sections/` — Content sections (e.g., SkillsSection)
- `components/ui/` — Reusable primitives (button, card)

### Key Conventions

- JSX files (not TypeScript) — all components use `.jsx` extension
- Dark mode support via Tailwind `dark:` variants
- Icons from `lucide-react`
- Conventional commits required (see [conventionalcommits.org](https://www.conventionalcommits.org/en/v1.0.0/))

### Deployment

- GitHub Actions workflow (`.github/workflows/deploy.yml`) auto-deploys on push to `main`
- `index.html` is deployed with no-cache headers; other assets use default caching
- Images in `public/images/` are excluded from S3 sync
