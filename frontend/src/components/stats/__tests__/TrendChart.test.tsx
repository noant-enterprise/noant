import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import type { TrendData } from '@/types'

vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children, height }: { children: React.ReactNode; height?: number }) => (
    <div data-testid="responsive-container" style={{ height }}>
      {children}
    </div>
  ),
  AreaChart: ({ children, data }: { children: React.ReactNode; data: unknown[] }) => (
    <div data-testid="area-chart" data-length={data.length}>
      {children}
    </div>
  ),
  Area: () => null,
  XAxis: () => null,
  YAxis: () => null,
  CartesianGrid: () => null,
  Tooltip: () => null,
  defs: () => null,
}))

import { TrendChart } from '@/components/stats/TrendChart'

const sampleData: TrendData[] = [
  { date: '2026-01-01', conversations: 10 },
  { date: '2026-01-02', conversations: 25 },
  { date: '2026-01-03', conversations: 18 },
]

describe('TrendChart', () => {
  it('renders the chart with data', () => {
    render(<TrendChart data={sampleData} />)
    expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
    expect(screen.getByTestId('area-chart')).toBeInTheDocument()
  })

  it('passes data to AreaChart', () => {
    render(<TrendChart data={sampleData} />)
    const chart = screen.getByTestId('area-chart')
    expect(chart).toHaveAttribute('data-length', '3')
  })

  it('renders with empty data array', () => {
    render(<TrendChart data={[]} />)
    const chart = screen.getByTestId('area-chart')
    expect(chart).toHaveAttribute('data-length', '0')
  })

  it('uses default height of 240', () => {
    render(<TrendChart data={sampleData} />)
    const container = screen.getByTestId('responsive-container')
    expect(container).toHaveStyle({ height: '240px' })
  })

  it('accepts custom height', () => {
    render(<TrendChart data={sampleData} height={400} />)
    const container = screen.getByTestId('responsive-container')
    expect(container).toHaveStyle({ height: '400px' })
  })
})
