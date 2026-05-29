import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

// Vitest auto-prefers this file over vite.config.js. We re-declare the React
// plugin here (rather than merging vite.config.js) so the test transform mirrors
// the dev/build transform without coupling the two configs.
export default defineConfig({
  plugins: [react()],
  test: {
    // `globals: true` exposes describe/it/expect/vi without per-file imports,
    // matching the Testing Library conventions used by the dashboard tests.
    // eslint.config.js declares these globals for test/mock files.
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/mocks/setup.js'],
    // Vitest owns unit/integration tests under src/. Playwright owns e2e/ specs
    // (a Playwright spec run through Vitest throws). Scoping `include` to src/
    // also keeps Vitest from discovering specs inside agent worktrees (.claude/).
    include: ['src/**/*.{test,spec}.{js,jsx}'],
    // Tailwind/PostCSS isn't needed in jsdom; skipping CSS keeps tests fast.
    css: false,
    // AC #38: `npm run test` must succeed even before any tests exist.
    passWithNoTests: true,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov', 'html'],
      reportsDirectory: './coverage',
      // Scope coverage to the Phase 5 testing surface — the recruiter dashboard
      // feature. Pre-existing pages/components outside this phase (HomePage,
      // ResumeRequestPage, layout, etc.) have no Phase 5 tests and would dilute
      // the "75% overall" target below into something unachievable in-scope.
      include: [
        'src/pages/DashboardPage.jsx',
        'src/components/sections/StatsCards.jsx',
        'src/components/sections/FilterBar.jsx',
        'src/components/sections/RecruiterTable.jsx',
        'src/components/ui/card.jsx',
        'src/lib/filters.js',
      ],
      // AC #20: enforce 75% overall. Satisfied once #39's tests land; on this
      // (#38) branch coverage is below threshold because tests come in #39.
      thresholds: {
        lines: 75,
        functions: 75,
        branches: 75,
        statements: 75,
      },
    },
  },
})
