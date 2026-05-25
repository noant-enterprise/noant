import { useMemo } from 'react'
import { PieChart, Pie, Cell, Tooltip, ResponsiveContainer } from 'recharts'
import type { ChannelDistribution } from '@/types'

interface ChannelDistributionChartProps {
  data: ChannelDistribution
}

const COLORS = ['#0ea5e9', '#25D366', '#E4405F', '#5865F2', '#f59e0b', '#8b5cf6']

export function ChannelDistributionChart({ data }: ChannelDistributionChartProps) {
  const chartData = useMemo(() => {
    return Object.entries(data).map(([name, value]) => ({
      name: name.charAt(0).toUpperCase() + name.slice(1),
      value,
    }))
  }, [data])

  return (
    <div className="w-full h-full flex flex-col items-center justify-center">
      {/* Square chart area — flex-1 fills available height, aspect-square forces 1:1 width */}
      <div className="flex-1 min-h-0 aspect-square">
        <ResponsiveContainer width="100%" height="100%" minWidth={0} minHeight={0}>
          <PieChart>
            <Pie
              data={chartData}
              cx="50%"
              cy="50%"
              innerRadius="60%"
              outerRadius="80%"
              paddingAngle={4}
              dataKey="value"
              stroke="none"
            >
              {chartData.map((_, index) => (
                <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip
              contentStyle={{
                background: 'var(--bg-surface)',
                border: '1px solid var(--border-default)',
                borderRadius: '8px',
                fontSize: '12px',
                color: 'var(--text-primary)',
              }}
              itemStyle={{ color: 'var(--text-primary)' }}
            />
          </PieChart>
        </ResponsiveContainer>
      </div>

      {/* Legend */}
      <div className="flex flex-wrap justify-center gap-3 mt-2 shrink-0">
        {chartData.map((entry, index) => (
          <div key={entry.name} className="flex items-center gap-1.5">
            <div
              className="w-2 h-2 rounded-full"
              style={{ background: COLORS[index % COLORS.length] }}
            />
            <span className="text-[10px] text-secondary">{entry.name}</span>
          </div>
        ))}
      </div>
    </div>
  )
}