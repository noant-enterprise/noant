import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { BottomNav } from '@/components/layout/BottomNav'

vi.mock('@/contexts/SidebarAlertsContext', () => ({
  useSidebarAlerts: vi.fn(() => ({
    unreadChats: 0,
    unknownQuestions: 0,
    channelIssues: 0,
    billingAlert: false,
    unreadNotifications: 0,
    total: 0,
  })),
}))

vi.mock('@/lib/utils', () => ({
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

function renderWithRouter(ui: React.ReactNode, initialEntries = ['/dashboard']) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>{ui}</MemoryRouter>
  )
}

describe('BottomNav', () => {
  it('renders all navigation links', () => {
    renderWithRouter(<BottomNav />)
    expect(screen.getByText('Home')).toBeInTheDocument()
    expect(screen.getByText('Inbox')).toBeInTheDocument()
    expect(screen.getByText('Teach')).toBeInTheDocument()
    expect(screen.getByText('Insights')).toBeInTheDocument()
    expect(screen.getByText('Settings')).toBeInTheDocument()
  })

  it('links to correct routes', () => {
    renderWithRouter(<BottomNav />)
    const links = screen.getAllByRole('link')
    const hrefs = links.map(l => l.getAttribute('href'))
    expect(hrefs).toContain('/dashboard')
    expect(hrefs).toContain('/chats')
    expect(hrefs).toContain('/teach')
    expect(hrefs).toContain('/insights')
    expect(hrefs).toContain('/settings')
  })

  it('highlights active route', () => {
    renderWithRouter(<BottomNav />, ['/dashboard'])
    const homeLink = screen.getByText('Home').closest('a')
    expect(homeLink?.className).toContain('text-noant-sky')
  })

  it('does not highlight inactive routes', () => {
    renderWithRouter(<BottomNav />, ['/dashboard'])
    const settingsLink = screen.getByText('Settings').closest('a')
    expect(settingsLink?.className).toContain('text-tertiary')
  })

  it('returns null when viewing a chat thread', () => {
    const { container } = renderWithRouter(<BottomNav />, ['/chats?id=123'])
    expect(container.innerHTML).toBe('')
  })

  it('shows badge for unread chats', async () => {
    const { useSidebarAlerts } = await import('@/contexts/SidebarAlertsContext')
    vi.mocked(useSidebarAlerts).mockReturnValue({
      unreadChats: 5,
      unknownQuestions: 0,
      channelIssues: 0,
      billingAlert: false,
      unreadNotifications: 0,
      total: 5,
    })
    renderWithRouter(<BottomNav />)
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('shows 9+ for badge count greater than 9', async () => {
    const { useSidebarAlerts } = await import('@/contexts/SidebarAlertsContext')
    vi.mocked(useSidebarAlerts).mockReturnValue({
      unreadChats: 15,
      unknownQuestions: 0,
      channelIssues: 0,
      billingAlert: false,
      unreadNotifications: 0,
      total: 15,
    })
    renderWithRouter(<BottomNav />)
    expect(screen.getByText('9+')).toBeInTheDocument()
  })
})
