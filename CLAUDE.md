# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Personal portfolio website (sh3r4rd.com) built with React, Tailwind CSS, and Vite. Deployed to AWS S3/CloudFront via GitHub Actions on push to `main`.

## Commands

### Frontend

- **Dev server:** `npm run dev` (or `make server`)
- **Build:** `npm run build` (or `make build`)
- **Lint:** `npm run lint`
- **Deploy (manual):** `make deploy bucket=<bucket-name>` (builds then syncs to S3)
- **Preview production build:** `npm run preview`

### Infrastructure (Terraform)

All commands run from `infra/recruiter-dashboard/`:

- **Init:** `terraform -chdir=infra/recruiter-dashboard init`
- **Validate:** `terraform -chdir=infra/recruiter-dashboard validate`
- **Format:** `terraform -chdir=infra/recruiter-dashboard fmt -recursive`
- **Plan:** `terraform -chdir=infra/recruiter-dashboard plan`

## Architecture

### Frontend

- **React 19** with **react-router-dom v7** for client-side routing (BrowserRouter)
- **Vite 7** as build tool with React plugin
- **Tailwind CSS 3** with typography plugin, PostCSS/Autoprefixer
- **ESLint 9** flat config with react-hooks and react-refresh plugins

### Infrastructure (`infra/recruiter-dashboard/`)

- **Terraform >= 1.5.0** with **AWS provider >= 5.31.0**
- Directory structure: root module + `modules/s3`, `modules/dynamodb`
- Local state (single-developer project)
- AWS region: us-east-1 (required for SES email receiving)
- DynamoDB: provisioned billing mode (15/15 RCU/WCU, under 25 free tier)
- S3: AES-256 encryption, 30-day lifecycle, full public access block
- Default tags applied via provider `default_tags` block

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
- Terraform: HCL files only, module pattern (`modules/<name>/main.tf`, `variables.tf`, `outputs.tf`), see `terraform.tfvars.example` for variable defaults

### Stale Boilerplate

- `src/App.css` and `src/assets/react.svg` are unused Vite scaffolding — do not import or extend them

### Component Patterns

- Content pages wrap in `<section className="max-w-4xl mx-auto space-y-12">` with `<Breadcrumbs />` then `<Header />` (exception: `NotFoundPage` uses a minimal layout)
- `components/ui/` uses **named exports** (e.g., `export function Button`, `export function Card`)
- All other components (pages, layout, sections) use **default exports**

### Backend API

- Resume request form POSTs to `https://api.sh3r4rd.com/requests`
- Payload fields: `firstName`, `lastName`, `email`, `phone`, `company`, `jobTitle`, `description`
- Backend infrastructure is defined in `infra/recruiter-dashboard/`

### Analytics

- Google Analytics via `window.gtag` in `PageTracker` component
- Measurement ID `G-L2852SHBRS` loaded in `index.html`

### Deployment

- GitHub Actions workflow (`.github/workflows/deploy.yml`) auto-deploys on push to `main`
- `index.html` is deployed with no-cache headers; other assets use default caching
- Images in `public/images/` are excluded from S3 sync
