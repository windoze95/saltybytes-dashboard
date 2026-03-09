interface Props {
  value: number
  max: number
  label?: string
  color?: string
}

export default function ProgressBar({ value, max, label, color = 'bg-[#FF6B85]' }: Props) {
  const pct = max > 0 ? Math.min((value / max) * 100, 100) : 0
  return (
    <div>
      {label && (
        <div className="flex justify-between text-sm mb-1">
          <span className="text-[#F0F0F5]/60">{label}</span>
          <span className="text-[#F0F0F5]/80">{pct.toFixed(1)}%</span>
        </div>
      )}
      <div className="w-full bg-[#2A2A36] rounded-full h-2">
        <div className={`${color} rounded-full h-2 transition-all`} style={{ width: `${pct}%` }} />
      </div>
    </div>
  )
}
