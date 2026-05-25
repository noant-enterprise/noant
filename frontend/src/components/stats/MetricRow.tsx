interface MetricRowProps {
  icon: React.ReactNode
  iconBg: string
  iconColor: string
  name: string
  value: string | number
}

export function MetricRow({ icon, iconBg, iconColor, name, value }: MetricRowProps) {
  return (
    <div className="flex items-center justify-between py-4 border-b border-default last:border-b-0 transition-colors duration-300">
      <div className="flex items-center gap-3">
        <div
          className="w-9 h-9 rounded-md flex items-center justify-center text-base transition-colors duration-300"
          style={{ background: iconBg, color: iconColor }}
        >
          {icon}
        </div>
        <span className="text-sm font-semibold text-primary transition-colors duration-300">{name}</span>
      </div>
      <span className="text-lg font-bold text-primary transition-colors duration-300">{value}</span>
    </div>
  )
}