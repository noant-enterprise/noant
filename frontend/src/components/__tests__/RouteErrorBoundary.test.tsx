import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import type { ReactNode } from 'react'
import { RouteErrorBoundary } from '@/components/RouteErrorBoundary'

vi.mock('lucide-react', () => ({
  AlertTriangle: (props: any) => <svg data-testid="alert-triangle" {...props} />,
  RefreshCw: (props: any) => <svg data-testid="refresh-cw" {...props} />,
  ArrowLeft: (props: any) => <svg data-testid="arrow-left" {...props} />,
}))

vi.mock('@/components/ui/Button', () => ({
  Button: ({ children, onClick, variant, ...props }: any) => (
    <button onClick={onClick} data-variant={variant} {...props}>{children}</button>
  ),
}))

function ThrowingChild(): ReactNode {
  throw new Error('Route crashed')
}

function NormalChild() {
  return <div>Page content</div>
}

describe('RouteErrorBoundary', () => {
  it('renders children normally when no error', () => {
    render(
      <RouteErrorBoundary>
        <NormalChild />
      </RouteErrorBoundary>
    )
    expect(screen.getByText('Page content')).toBeInTheDocument()
  })

  it('shows error UI when child throws', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    render(
      <RouteErrorBoundary>
        <ThrowingChild />
      </RouteErrorBoundary>
    )
    expect(screen.getByText(/failed to load/)).toBeInTheDocument()
    expect(screen.getByText('Route crashed')).toBeInTheDocument()
    consoleSpy.mockRestore()
  })

  it('shows default page name when pageName is not provided', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    render(
      <RouteErrorBoundary>
        <ThrowingChild />
      </RouteErrorBoundary>
    )
    expect(screen.getByText('Page failed to load')).toBeInTheDocument()
    consoleSpy.mockRestore()
  })

  it('shows custom page name when provided', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    render(
      <RouteErrorBoundary pageName="Dashboard">
        <ThrowingChild />
      </RouteErrorBoundary>
    )
    expect(screen.getByText('Dashboard failed to load')).toBeInTheDocument()
    consoleSpy.mockRestore()
  })

  it('has try again button', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    render(
      <RouteErrorBoundary>
        <ThrowingChild />
      </RouteErrorBoundary>
    )
    expect(screen.getByText('Try again')).toBeInTheDocument()
    consoleSpy.mockRestore()
  })

  it('has go back button', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    render(
      <RouteErrorBoundary>
        <ThrowingChild />
      </RouteErrorBoundary>
    )
    expect(screen.getByText('Go back')).toBeInTheDocument()
    consoleSpy.mockRestore()
  })

  it('does not render children when error occurred', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    render(
      <RouteErrorBoundary>
        <ThrowingChild />
      </RouteErrorBoundary>
    )
    expect(screen.queryByText('Page content')).not.toBeInTheDocument()
    consoleSpy.mockRestore()
  })
})
