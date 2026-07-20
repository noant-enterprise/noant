import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MetricRow } from '@/components/stats/MetricRow'

describe('MetricRow', () => {
  it('renders name and value', () => {
    render(
      <MetricRow
        icon={<span>icon</span>}
        iconBg="#fff"
        iconColor="#000"
        name="Total Users"
        value={1200}
      />
    )
    expect(screen.getByText('Total Users')).toBeInTheDocument()
    expect(screen.getByText('1200')).toBeInTheDocument()
  })

  it('renders string value', () => {
    render(
      <MetricRow
        icon={<span>icon</span>}
        iconBg="#fff"
        iconColor="#000"
        name="Status"
        value="Active"
      />
    )
    expect(screen.getByText('Active')).toBeInTheDocument()
  })

  it('renders the icon', () => {
    render(
      <MetricRow
        icon={<span data-testid="test-icon">⚡</span>}
        iconBg="#fff"
        iconColor="#000"
        name="Metric"
        value={10}
      />
    )
    expect(screen.getByTestId('test-icon')).toBeInTheDocument()
  })

  it('applies icon background and color styles', () => {
    render(
      <MetricRow
        icon={<span>icon</span>}
        iconBg="#ff0000"
        iconColor="#ffffff"
        name="Metric"
        value={10}
      />
    )
    const iconContainer = screen.getByText('icon').parentElement
    expect(iconContainer).toHaveStyle({ background: '#ff0000', color: '#ffffff' })
  })

  it('renders the name with correct classes', () => {
    render(
      <MetricRow
        icon={<span>icon</span>}
        iconBg="#fff"
        iconColor="#000"
        name="Revenue"
        value={5000}
      />
    )
    const nameEl = screen.getByText('Revenue')
    expect(nameEl.className).toContain('text-primary')
  })

  it('renders the value with correct classes', () => {
    render(
      <MetricRow
        icon={<span>icon</span>}
        iconBg="#fff"
        iconColor="#000"
        name="Revenue"
        value={5000}
      />
    )
    const valueEl = screen.getByText('5000')
    expect(valueEl.className).toContain('font-bold')
  })
})
