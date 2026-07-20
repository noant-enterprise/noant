import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { OfflineBanner } from '@/components/OfflineBanner'

vi.mock('@/contexts/NetworkContext', () => ({
  useNetwork: vi.fn(),
}))

vi.mock('lucide-react', () => ({
  WifiOff: (props: any) => <svg data-testid="wifi-off" {...props} />,
  Wifi: (props: any) => <svg data-testid="wifi" {...props} />,
}))

import { useNetwork } from '@/contexts/NetworkContext'
const mockUseNetwork = vi.mocked(useNetwork)

describe('OfflineBanner', () => {
  it('shows offline banner when offline', () => {
    mockUseNetwork.mockReturnValue({ isOnline: false, wasOffline: false, lastChecked: null })
    render(<OfflineBanner />)
    expect(screen.getByRole('alert')).toBeInTheDocument()
    expect(screen.getByText(/offline/)).toBeInTheDocument()
  })

  it('returns null when online and was never offline', () => {
    mockUseNetwork.mockReturnValue({ isOnline: true, wasOffline: false, lastChecked: null })
    const { container } = render(<OfflineBanner />)
    expect(container.innerHTML).toBe('')
  })

  it('shows back online banner after reconnecting', () => {
    mockUseNetwork.mockReturnValue({ isOnline: true, wasOffline: true, lastChecked: new Date() })
    render(<OfflineBanner />)
    expect(screen.getByRole('status')).toBeInTheDocument()
    expect(screen.getByText(/Back online/)).toBeInTheDocument()
  })

  it('renders wifi-off icon when offline', () => {
    mockUseNetwork.mockReturnValue({ isOnline: false, wasOffline: false, lastChecked: null })
    render(<OfflineBanner />)
    expect(screen.getByTestId('wifi-off')).toBeInTheDocument()
  })

  it('renders wifi icon when back online', () => {
    mockUseNetwork.mockReturnValue({ isOnline: true, wasOffline: true, lastChecked: new Date() })
    render(<OfflineBanner />)
    expect(screen.getByTestId('wifi')).toBeInTheDocument()
  })
})
