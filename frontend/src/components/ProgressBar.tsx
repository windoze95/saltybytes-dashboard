interface Props {
  value: number
  max: number
  label?: string
  color?: string
}

export default function ProgressBar({ value, max, label, color = 'bg-blue-500' }: Props) {
  const pct = max > 0 ? Math.min((value / max) * 100, 100) : 0
  return (
    <div>
      {label && (
        <div className="flex justify-between text-sm mb-1">
          <span className="text-slate-400">{label}</span>
          <span className="text-slate-300">{pct.toFixed(1)}%</span>
        </div>
      )}
      <div className="w-full bg-slate-700 rounded-full h-2">
        <div className={`${color} rounded-full h-2 transition-all`} style={{ width: `${pct}%` }} />
      </div>
    </div>
  )
}
