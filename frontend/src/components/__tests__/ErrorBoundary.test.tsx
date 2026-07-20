import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ErrorBoundary } from '@/components/ErrorBoundary'

vi.mock('lucide-react', () => ({
  AlertTriangle: (props: any) => <svg data-testid="alert-triangle" {...props} />,
  RefreshCw: (props: any) => <svg data-testid="refresh-cw" {...props} />,
}))

vi.mock('@/components/ui/Button', () => ({
  Button: ({ children, onClick, ...props }: any) => (
    <button onClick={onClick} {...props}>{children}</button>
  ),
}))

function ThrowingChild({ shouldThrow = true }: { shouldThrow?: boolean }) {
  if (shouldThrow) throw new Error('Test error')
  return <div>Normal child</div>
}

function NoThrowChild() {
  return <div>Everything is fine</div>
}

describe('ErrorBoundary', () => {
  it('renders children normally when no error', () => {
    render(
      <ErrorBoundary>
        <NoThrowChild />
      </ErrorBoundary>
    )
    expect(screen.getByText('Everything is fine')).toBeInTheDocument()
  })

  it('shows error UI when child throws', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    render(
      <ErrorBoundary>
        <ThrowingChild />
      </ErrorBoundary>
    )
    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
    expect(screen.getByText(/unexpected error occurred/)).toBeInTheDocument()
    expect(screen.getByText('Test error')).toBeInTheDocument()
    consoleSpy.mockRestore()
  })

  it('does not render children when error occurred', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    render(
      <ErrorBoundary>
        <ThrowingChild />
      </ErrorBoundary>
    )
    expect(screen.queryByText('Normal child')).not.toBeInTheDocument()
    consoleSpy.mockRestore()
  })

  it('shows reload button', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    render(
      <ErrorBoundary>
        <ThrowingChild />
      </ErrorBoundary>
    )
    expect(screen.getByText('Reload page')).toBeInTheDocument()
    consoleSpy.mockRestore()
  })

  it('reload button triggers window.location.reload', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const reloadSpy = vi.fn()
    Object.defineProperty(window, 'location', {
      value: { reload: reloadSpy },
      writable: true,
    })
    const user = userEvent.setup()
    render(
      <ErrorBoundary>
        <ThrowingChild />
      </ErrorBoundary>
    )
    await user.click(screen.getByText('Reload page'))
    expect(reloadSpy).toHaveBeenCalledOnce()
    consoleSpy.mockRestore()
  })
})
