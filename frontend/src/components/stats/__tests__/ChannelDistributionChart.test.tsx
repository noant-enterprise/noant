import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import type { ChannelDistribution } from '@/types'

vi.mock('recharts', () => ({
  ResponsiveContainer: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="responsive-container">{children}</div>
  ),
  PieChart: ({ children }: { children: React.ReactNode }) => (
    <div data-testid="pie-chart">{children}</div>
  ),
  Pie: ({ data }: { data: unknown[] }) => (
    <div data-testid="pie" data-length={data.length} />
  ),
  Cell: () => null,
  Tooltip: () => null,
}))

import { ChannelDistributionChart } from '@/components/stats/ChannelDistributionChart'

const sampleData: ChannelDistribution = {
  whatsapp: 50,
  instagram: 30,
  telegram: 20,
}

describe('ChannelDistributionChart', () => {
  it('renders the chart', () => {
    render(<ChannelDistributionChart data={sampleData} />)
    expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
    expect(screen.getByTestId('pie-chart')).toBeInTheDocument()
  })

  it('shows legend entries for each channel', () => {
    render(<ChannelDistributionChart data={sampleData} />)
    expect(screen.getByText('Whatsapp')).toBeInTheDocument()
    expect(screen.getByText('Instagram')).toBeInTheDocument()
    expect(screen.getByText('Telegram')).toBeInTheDocument()
  })

  it('capitalizes channel names in legend', () => {
    const data: ChannelDistribution = { discord: 10 }
    render(<ChannelDistributionChart data={data} />)
    expect(screen.getByText('Discord')).toBeInTheDocument()
  })

  it('renders with empty data', () => {
    render(<ChannelDistributionChart data={{}} />)
    expect(screen.getByTestId('responsive-container')).toBeInTheDocument()
    expect(screen.queryByRole('listitem')).not.toBeInTheDocument()
  })

  it('passes correct data length to Pie', () => {
    render(<ChannelDistributionChart data={sampleData} />)
    const pie = screen.getByTestId('pie')
    expect(pie).toHaveAttribute('data-length', '3')
  })
})
