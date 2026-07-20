import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { GuestRoute } from '@/components/auth/GuestRoute'

vi.mock('@/hooks/useAuth', () => ({
  useAuth: vi.fn(),
}))

vi.mock('@/components/ui/Skeleton', () => ({
  Skeleton: ({ className }: { className?: string }) => (
    <div data-testid="skeleton" className={className} />
  ),
}))

import { useAuth } from '@/hooks/useAuth'
const mockUseAuth = vi.mocked(useAuth)

function renderGuestRoute(initialEntries = ['/']) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <Routes>
        <Route path="/dashboard" element={<div>Dashboard Page</div>} />
        <Route path="*" element={
          <GuestRoute><div>Public content</div></GuestRoute>
        } />
      </Routes>
    </MemoryRouter>
  )
}

describe('GuestRoute', () => {
  it('renders children when user is not authenticated', () => {
    mockUseAuth.mockReturnValue({ user: null, loading: false, signOut: vi.fn(), refreshUser: vi.fn() })
    renderGuestRoute()
    expect(screen.getByText('Public content')).toBeInTheDocument()
  })

  it('shows loading skeleton when loading', () => {
    mockUseAuth.mockReturnValue({ user: null, loading: true, signOut: vi.fn(), refreshUser: vi.fn() })
    renderGuestRoute()
    expect(screen.getAllByTestId('skeleton')).toHaveLength(3)
    expect(screen.queryByText('Public content')).not.toBeInTheDocument()
  })

  it('redirects to /dashboard when user is authenticated', () => {
    mockUseAuth.mockReturnValue({ user: { id: '1' } as any, loading: false, signOut: vi.fn(), refreshUser: vi.fn() })
    renderGuestRoute()
    expect(screen.queryByText('Public content')).not.toBeInTheDocument()
    expect(screen.getByText('Dashboard Page')).toBeInTheDocument()
  })
})
