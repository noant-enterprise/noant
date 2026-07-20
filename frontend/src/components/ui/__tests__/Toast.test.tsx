import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ToastProvider, useToast } from '@/components/ui/Toast'

function TestComponent() {
  const { toast } = useToast()
  return (
    <button onClick={() => toast('Test message')}>
      Show Toast
    </button>
  )
}

function renderWithProvider(ui?: React.ReactNode) {
  return render(<ToastProvider>{ui ?? <TestComponent />}</ToastProvider>)
}

describe('Toast', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('useToast returns toast function inside provider', () => {
    let toastFn: ReturnType<typeof useToast>['toast'] | undefined
    function Grabber() {
      const { toast } = useToast()
      toastFn = toast
      return null
    }
    render(
      <ToastProvider>
        <Grabber />
      </ToastProvider>
    )
    expect(toastFn).toBeDefined()
    expect(typeof toastFn).toBe('function')
  })

  it('throws when useToast is used outside provider', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})
    function Bad() {
      useToast()
      return null
    }
    expect(() => render(<Bad />)).toThrow('useToast must be inside ToastProvider')
    spy.mockRestore()
  })

  it('adding a toast renders it on screen', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
    renderWithProvider()
    await user.click(screen.getByRole('button', { name: /show toast/i }))
    expect(screen.getByText('Test message')).toBeInTheDocument()
  })

  it('toast auto-removes after timeout', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
    renderWithProvider()
    await user.click(screen.getByRole('button', { name: /show toast/i }))
    expect(screen.getByText('Test message')).toBeInTheDocument()

    act(() => {
      vi.advanceTimersByTime(4000)
    })

    expect(screen.queryByText('Test message')).not.toBeInTheDocument()
  })
})
