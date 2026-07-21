import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { Header } from '@/components/layout/Header'

vi.mock('@/hooks/useAuth', () => ({
  useAuth: () => ({
    user: {
      id: 'user-1',
      first_name: 'John',
      last_name: 'Doe',
      email: 'john@test.com',
      plan: 'pro',
    },
  }),
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

vi.mock('@/lib/api', () => ({
  api: {
    get: vi.fn().mockResolvedValue({ notifications: [] }),
    post: vi.fn(),
  },
}))

function renderWithRouter(ui: React.ReactNode, initialEntries = ['/']) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>{ui}</MemoryRouter>
  )
}

describe('Header', () => {
  const defaultProps = {
    onMenuClick: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the page title for dashboard', () => {
    renderWithRouter(<Header {...defaultProps} />, ['/'])
    expect(screen.getByText('Overview')).toBeInTheDocument()
  })

  it('renders the page title for chats', () => {
    renderWithRouter(<Header {...defaultProps} />, ['/chats'])
    expect(screen.getByText('Conversations')).toBeInTheDocument()
  })

  it('renders the page title for channels', () => {
    renderWithRouter(<Header {...defaultProps} />, ['/channels'])
    expect(screen.getByText('Your channels')).toBeInTheDocument()
  })

  it('renders the page title for settings', () => {
    renderWithRouter(<Header {...defaultProps} />, ['/settings'])
    expect(screen.getByText('Settings')).toBeInTheDocument()
  })

  it('renders noant as fallback for unknown routes', () => {
    renderWithRouter(<Header {...defaultProps} />, ['/unknown-page'])
    expect(screen.getAllByText('noant').length).toBeGreaterThanOrEqual(1)
  })

  it('renders notification bell', () => {
    renderWithRouter(<Header {...defaultProps} />)
    expect(screen.getByLabelText('Notifications')).toBeInTheDocument()
  })

  it('renders theme toggle button', () => {
    renderWithRouter(<Header {...defaultProps} />)
    expect(screen.getByLabelText('Toggle theme')).toBeInTheDocument()
  })

  it('renders mobile menu button', () => {
    renderWithRouter(<Header {...defaultProps} />)
    expect(screen.getByLabelText('Open menu')).toBeInTheDocument()
  })

  it('calls onMenuClick when menu button is clicked', async () => {
    const onMenuClick = vi.fn()
    const user = userEvent.setup()
    renderWithRouter(<Header {...defaultProps} onMenuClick={onMenuClick} />)
    await user.click(screen.getByLabelText('Open menu'))
    expect(onMenuClick).toHaveBeenCalledOnce()
  })

  it('shows user initials', () => {
    renderWithRouter(<Header {...defaultProps} />)
    expect(screen.getByText('JD')).toBeInTheDocument()
  })
})
