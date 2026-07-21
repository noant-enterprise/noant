import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { DashboardLayout } from '@/components/layout/DashboardLayout'

vi.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({
    user: {
      id: 'user-1',
      first_name: 'John',
      last_name: 'Doe',
      email: 'john@test.com',
      plan: 'pro',
    },
    signOut: vi.fn(),
  }),
}))

vi.mock('@/hooks/useOffline', () => ({
  useOffline: () => false,
}))

vi.mock('@/contexts/SidebarAlertsContext', () => ({
  useSidebarAlerts: () => ({
    unreadChats: 0,
    unknownQuestions: 0,
    channelIssues: 0,
    billingAlert: false,
    unreadNotifications: 0,
    total: 0,
  }),
}))

vi.mock('@/components/ui/ConfirmModal', () => ({
  ConfirmModal: ({ open, onClose, title }: any) =>
    open ? (
      <div role="dialog" aria-label={title}>
        <button onClick={onClose}>Cancel</button>
      </div>
    ) : null,
}))

vi.mock('@/lib/utils', () => ({
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

vi.mock('@/lib/api', () => ({
  api: {
    get: vi.fn().mockResolvedValue({ notifications: [] }),
    post: vi.fn(),
  },
}))

vi.mock('@/components/ui/Toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}))

vi.mock('@/hooks/useWebSocket', () => ({
  useWebSocket: () => ({
    subscribe: vi.fn(() => vi.fn()),
  }),
}))

function renderWithRouter(initialEntries = ['/dashboard']) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <Routes>
        <Route path="/" element={<DashboardLayout />}>
          <Route path="dashboard" element={<div>child content</div>} />
        </Route>
      </Routes>
    </MemoryRouter>
  )
}

describe('DashboardLayout', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the sidebar', () => {
    renderWithRouter()
    expect(screen.getAllByText('noant').length).toBeGreaterThanOrEqual(1)
  })

  it('renders the header with page title', () => {
    renderWithRouter()
    expect(screen.getAllByText('noant').length).toBeGreaterThanOrEqual(1)
  })

  it('renders the bottom navigation', () => {
    renderWithRouter()
    expect(screen.getByText('Home')).toBeInTheDocument()
  })

  it('renders the main content area', () => {
    renderWithRouter()
    expect(screen.getByText('child content')).toBeInTheDocument()
  })

  it('shows no offline banner when online', () => {
    renderWithRouter()
    expect(screen.queryByText('You are currently offline')).not.toBeInTheDocument()
  })

  it('renders the layout container', () => {
    const { container } = renderWithRouter()
    expect(container.querySelector('.h-screen')).toBeInTheDocument()
  })

  it('renders sidebar nav sections', () => {
    renderWithRouter()
    expect(screen.getByText('Your workspace')).toBeInTheDocument()
    expect(screen.getByText('Build your Noant')).toBeInTheDocument()
    expect(screen.getByText('Manage')).toBeInTheDocument()
  })

  it('renders user info in sidebar', () => {
    renderWithRouter()
    expect(screen.getByText('John Doe')).toBeInTheDocument()
  })

  it('shows user plan', () => {
    renderWithRouter()
    expect(screen.getByText('pro plan')).toBeInTheDocument()
  })
})
