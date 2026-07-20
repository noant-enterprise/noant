import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'

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

function renderProtectedRoute(
  initialEntries: string[],
  _authOverrides?: Partial<ReturnType<typeof useAuth>>
) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <Routes>
        <Route
          path="/login"
          element={<div>Login Page</div>}
        />
        <Route
          path="/onboarding"
          element={<div>Onboarding Page</div>}
        />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <div>Protected content</div>
            </ProtectedRoute>
          }
        />
        <Route
          path="/dashboard"
          element={<div>Dashboard</div>}
        />
      </Routes>
    </MemoryRouter>
  )
}

describe('ProtectedRoute', () => {
  it('renders children when user is authenticated', () => {
    mockUseAuth.mockReturnValue({
      user: { id: '1', onboarding_status: 'complete' } as any,
      loading: false,
      signOut: vi.fn(),
      refreshUser: vi.fn(),
    } as any)
    renderProtectedRoute(['/'])
    expect(screen.getByText('Protected content')).toBeInTheDocument()
  })

  it('redirects to /login when user is not authenticated', () => {
    mockUseAuth.mockReturnValue({
      user: null,
      loading: false,
      signOut: vi.fn(),
      refreshUser: vi.fn(),
    } as any)
    renderProtectedRoute(['/'])
    expect(screen.queryByText('Protected content')).not.toBeInTheDocument()
    expect(screen.getByText('Login Page')).toBeInTheDocument()
  })

  it('shows loading skeleton when loading', () => {
    mockUseAuth.mockReturnValue({
      user: null,
      loading: true,
      signOut: vi.fn(),
      refreshUser: vi.fn(),
    } as any)
    renderProtectedRoute(['/'])
    expect(screen.getAllByTestId('skeleton')).toHaveLength(3)
    expect(screen.queryByText('Protected content')).not.toBeInTheDocument()
    expect(screen.queryByText('Login Page')).not.toBeInTheDocument()
  })

  it('redirects to /onboarding when user needs onboarding', () => {
    mockUseAuth.mockReturnValue({
      user: { id: '1', onboarding_status: 'pending' } as any,
      loading: false,
      signOut: vi.fn(),
      refreshUser: vi.fn(),
    } as any)
    renderProtectedRoute(['/'])
    expect(screen.queryByText('Protected content')).not.toBeInTheDocument()
    expect(screen.getByText('Onboarding Page')).toBeInTheDocument()
  })

  it('renders children when on onboarding route and user needs onboarding', () => {
    mockUseAuth.mockReturnValue({
      user: { id: '1', onboarding_status: 'pending' } as any,
      loading: false,
      signOut: vi.fn(),
      refreshUser: vi.fn(),
    } as any)
    render(
      <MemoryRouter initialEntries={['/onboarding']}>
        <Routes>
          <Route
            path="/onboarding"
            element={
              <ProtectedRoute>
                <div>Onboarding content</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </MemoryRouter>
    )
    expect(screen.getByText('Onboarding content')).toBeInTheDocument()
  })

  it('passes location state with from when redirecting to login', () => {
    mockUseAuth.mockReturnValue({
      user: null,
      loading: false,
      signOut: vi.fn(),
      refreshUser: vi.fn(),
    } as any)
    renderProtectedRoute(['/'])
    expect(screen.getByText('Login Page')).toBeInTheDocument()
  })
})
