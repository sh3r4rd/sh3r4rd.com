import { setupServer } from 'msw/node'
import { handlers } from './handlers'

// Node-side MSW server (used by Vitest/jsdom). The browser worker is not needed
// since the app itself never runs against mocks outside tests.
export const server = setupServer(...handlers)
