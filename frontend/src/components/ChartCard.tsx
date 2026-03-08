import { type ReactNode } from 'react'

interface Props {
  title: string
  children: ReactNode
  className?: string
}

export default function ChartCard({ title, children, className = '' }: Props) {
  return (
    <div className={`bg-slate-800 rounded-lg p-4 border border-slate-700 ${className}`}>
      <h3 className="text-sm font-medium text-slate-400 mb-3">{title}</h3>
      {children}
    </div>
  )
}
