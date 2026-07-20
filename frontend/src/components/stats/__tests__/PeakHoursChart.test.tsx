import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import type { PeakHourData } from '@/types'

vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children, height }: { children: React.ReactNode; height?: number }) => (
    <div data-testid="responsive-container" style={{ height }}>
      {children}
    </div>
  ),
  BarChart: ({ children, data }: { children: React.ReactNode; data: unknown[] }) => (
    <div data-testid="bar-chart" data-length={data.length}>
      {children}
    </div>
  ),
  Bar: () => null,
  XAxis: () => null,
  YAxis: () => null,
  CartesianGrid: () => null,
  Tooltip: () => null,
}))

import { PeakHoursChart } from '@/components/stats/PeakHoursChart'

const sampleData: PeakHourData[] = [
  { hour: '9 AM', volume: 15 },
  { hour: '12 PM', volume: 40 },
  { hour: '3 PM', volume: 25 },
]

describe('PeakHoursChart', () => {
  it('renders the chart', () => {
    render(<PeakHoursChart data={sampleData} />)
    expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
    expect(screen.getByTestId('bar-chart')).toBeInTheDocument()
  })

  it('passes data to BarChart', () => {
    render(<PeakHoursChart data={sampleData} />)
    const chart = screen.getByTestId('bar-chart')
    expect(chart).toHaveAttribute('data-length', '3')
  })

  it('renders with empty data', () => {
    render(<PeakHoursChart data={[]} />)
    const chart = screen.getByTestId('bar-chart')
    expect(chart).toHaveAttribute('data-length', '0')
  })

  it('uses default height of 180', () => {
    render(<PeakHoursChart data={sampleData} />)
    const container = screen.getByTestId('responsive-container')
    expect(container).toHaveStyle({ height: '180px' })
  })

  it('accepts custom height', () => {
    render(<PeakHoursChart data={sampleData} height={300} />)
    const container = screen.getByTestId('responsive-container')
    expect(container).toHaveStyle({ height: '300px' })
  })
})
