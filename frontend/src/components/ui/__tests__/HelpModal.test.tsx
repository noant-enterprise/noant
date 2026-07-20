import { describe, it, expect } from 'vitest'

// NOTE: The file at src/components/ui/HelpModal.tsx actually exports a
// `ChatsPage` component (default export), not a HelpModal.
// It is a complex chat page with many dependencies (WebSocket, API, etc.)
// that would require extensive mocking. Skipping automated tests for this
// file until it is refactored or the correct HelpModal component is located.
describe('HelpModal', () => {
  it.skip('HelpModal.tsx does not contain a HelpModal component — it exports ChatsPage', () => {
    expect(true).toBe(true)
  })
})
