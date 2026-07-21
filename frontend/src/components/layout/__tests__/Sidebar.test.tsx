import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { Sidebar } from '@/components/layout/Sidebar'

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

function renderWithRouter(ui: React.ReactNode, initialEntries = ['/dashboard']) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>{ui}</MemoryRouter>
  )
}

describe('Sidebar', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the noant brand', () => {
    renderWithRouter(<Sidebar />)
    expect(screen.getByText('noant')).toBeInTheDocument()
  })

  it('renders all navigation sections', () => {
    renderWithRouter(<Sidebar />)
    expect(screen.getByText('Your workspace')).toBeInTheDocument()
    expect(screen.getByText('Build your Noant')).toBeInTheDocument()
    expect(screen.getByText('Manage')).toBeInTheDocument()
  })

  it('renders all navigation links', () => {
    renderWithRouter(<Sidebar />)
    expect(screen.getByText('Overview')).toBeInTheDocument()
    expect(screen.getByText('Conversations')).toBeInTheDocument()
    expect(screen.getByText('Insights')).toBeInTheDocument()
    expect(screen.getByText('Leads')).toBeInTheDocument()
    expect(screen.getByText('Teach your Noant')).toBeInTheDocument()
    expect(screen.getByText('Your channels')).toBeInTheDocument()
    expect(screen.getByText('Web Widget')).toBeInTheDocument()
    expect(screen.getByText('Inventory')).toBeInTheDocument()
    expect(screen.getByText('Team')).toBeInTheDocument()
    expect(screen.getByText('Billing')).toBeInTheDocument()
    expect(screen.getByText('Settings')).toBeInTheDocument()
  })

  it('shows user name and initials', () => {
    renderWithRouter(<Sidebar />)
    expect(screen.getByText('John Doe')).toBeInTheDocument()
    expect(screen.getByText('JD')).toBeInTheDocument()
  })

  it('shows user plan', () => {
    renderWithRouter(<Sidebar />)
    expect(screen.getByText('pro plan')).toBeInTheDocument()
  })

  it('highlights active route', () => {
    renderWithRouter(<Sidebar />, ['/dashboard'])
    const overviewLink = screen.getByText('Overview').closest('a')
    expect(overviewLink?.className).toContain('text-noant-sky')
  })

  it('renders close button when onClose provided', () => {
    const onClose = vi.fn()
    renderWithRouter(<Sidebar onClose={onClose} />)
    expect(screen.getByLabelText('Close menu')).toBeInTheDocument()
  })

  it('does not render close button when onClose is not provided', () => {
    renderWithRouter(<Sidebar />)
    expect(screen.queryByLabelText('Close menu')).not.toBeInTheDocument()
  })

  it('links have correct href attributes', () => {
    renderWithRouter(<Sidebar />)
    const links = screen.getAllByRole('link')
    const hrefs = links.map(l => l.getAttribute('href'))
    expect(hrefs).toContain('/dashboard')
    expect(hrefs).toContain('/chats')
    expect(hrefs).toContain('/teach')
    expect(hrefs).toContain('/channels')
    expect(hrefs).toContain('/insights')
    expect(hrefs).toContain('/settings')
    expect(hrefs).toContain('/billing')
    expect(hrefs).toContain('/team')
    expect(hrefs).toContain('/widget')
    expect(hrefs).toContain('/leads')
    expect(hrefs).toContain('/inventory')
  })

  it('shows logout button', () => {
    renderWithRouter(<Sidebar />)
    const logoutBtn = screen.getByRole('button', { name: /john doe/i })
    expect(logoutBtn).toBeInTheDocument()
  })
})
