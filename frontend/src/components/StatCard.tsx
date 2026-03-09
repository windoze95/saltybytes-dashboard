interface Props {
  title: string
  value: string | number
  subtitle?: string
  color?: string
}

export default function StatCard({ title, value, subtitle, color = 'text-[#F0F0F5]' }: Props) {
  return (
    <div className="bg-[#1E1E28] rounded-lg p-4 border border-[#3A3A48]">
      <p className="text-sm text-[#F0F0F5]/60">{title}</p>
      <p className={`text-2xl font-bold mt-1 ${color}`}>{value}</p>
      {subtitle && <p className="text-xs text-[#F0F0F5]/50 mt-1">{subtitle}</p>}
    </div>
  )
}
