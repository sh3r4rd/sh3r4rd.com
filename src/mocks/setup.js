// Global test setup, wired in via vitest.config.js `setupFiles`.
import '@testing-library/jest-dom/vitest'
import { afterAll, afterEach, beforeAll } from 'vitest'
import { cleanup } from '@testing-library/react'
import { server } from './server'

// Start MSW before any test runs. `onUnhandledRequest: 'error'` fails loudly if
// a test triggers a request with no matching handler, so missing mocks surface
// as test failures instead of silent network attempts.
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))

// Reset per-test handler overrides and unmount React trees between tests so
// state never leaks across cases.
afterEach(() => {
  server.resetHandlers()
  cleanup()
})

afterAll(() => server.close())
