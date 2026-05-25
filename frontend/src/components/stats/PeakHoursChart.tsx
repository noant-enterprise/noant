import { useMemo } from 'react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import type { PeakHourData } from '@/types'

interface PeakHoursChartProps {
  data: PeakHourData[]
  height?: number
}

export function PeakHoursChart({ data, height = 180 }: PeakHoursChartProps) {
  const chartData = useMemo(() => data, [data])

  return (
    <ResponsiveContainer width="100%" height={height}>
      <BarChart data={chartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="var(--border-default)" vertical={false} />
        <XAxis
          dataKey="hour"
          axisLine={false}
          tickLine={false}
          tick={{ fill: 'var(--text-tertiary)', fontSize: 10 }}
          dy={8}
        />
        <YAxis
          axisLine={false}
          tickLine={false}
          tick={{ fill: 'var(--text-tertiary)', fontSize: 10 }}
        />
        <Tooltip
          contentStyle={{
            background: 'var(--bg-surface)',
            border: '1px solid var(--border-default)',
            borderRadius: '8px',
            fontSize: '12px',
            color: 'var(--text-primary)',
          }}
          itemStyle={{ color: 'var(--text-primary)' }}
          cursor={{ fill: 'var(--bg-inset)', radius: 4 }}
        />
        <Bar
          dataKey="volume"
          fill="#0ea5e9"
          radius={[4, 4, 0, 0]}
          maxBarSize={40}
        />
      </BarChart>
    </ResponsiveContainer>
  )
}