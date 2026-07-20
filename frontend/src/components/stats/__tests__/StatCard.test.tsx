import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { StatCard } from '@/components/stats/StatCard'

describe('StatCard', () => {
  it('renders label', () => {
    render(<StatCard label="Total Users" value={1200} />)
    expect(screen.getByText('Total Users')).toBeInTheDocument()
  })

  it('renders numeric value', () => {
    render(<StatCard label="Revenue" value={50000} />)
    expect(screen.getByText('50000')).toBeInTheDocument()
  })

  it('renders string value', () => {
    render(<StatCard label="Status" value="Active" />)
    expect(screen.getByText('Active')).toBeInTheDocument()
  })

  it('shows positive change indicator', () => {
    render(<StatCard label="Users" value={100} change={15} />)
    expect(screen.getByText('+15%')).toBeInTheDocument()
  })

  it('shows negative change indicator', () => {
    render(<StatCard label="Users" value={100} change={-8} />)
    expect(screen.getByText('-8%')).toBeInTheDocument()
  })

  it('shows zero change without +/- prefix', () => {
    render(<StatCard label="Users" value={100} change={0} />)
    expect(screen.getByText('0%')).toBeInTheDocument()
  })

  it('hides change when not provided', () => {
    render(<StatCard label="Users" value={100} />)
    expect(screen.queryByText(/%/)).not.toBeInTheDocument()
  })

  it('applies default variant by default', () => {
    render(<StatCard label="X" value="Y" />)
    expect(screen.getByText('Y').className).toContain('text-primary')
  })

  it('applies success variant', () => {
    render(<StatCard label="X" value="Y" variant="success" />)
    expect(screen.getByText('Y').className).toContain('text-emerald-600')
  })

  it('applies warning variant', () => {
    render(<StatCard label="X" value="Y" variant="warning" />)
    expect(screen.getByText('Y').className).toContain('text-amber-600')
  })

  it('applies info variant', () => {
    render(<StatCard label="X" value="Y" variant="info" />)
    expect(screen.getByText('Y').className).toContain('text-noant-sky')
  })

  it('applies error variant', () => {
    render(<StatCard label="X" value="Y" variant="error" />)
    expect(screen.getByText('Y').className).toContain('text-red-600')
  })
})
