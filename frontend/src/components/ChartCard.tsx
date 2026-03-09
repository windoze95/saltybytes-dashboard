import { type ReactNode } from 'react'

interface Props {
  title: string
  children: ReactNode
  className?: string
}

export default function ChartCard({ title, children, className = '' }: Props) {
  return (
    <div className={`bg-[#1E1E28] rounded-lg p-4 border border-[#3A3A48] ${className}`}>
      <h3 className="text-sm font-medium text-[#F0F0F5]/60 mb-3">{title}</h3>
      {children}
    </div>
  )
}
