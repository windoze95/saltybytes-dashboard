import { api } from '../lib/api'
import { useMetrics } from '../hooks/useMetrics'
import StatCard from '../components/StatCard'
import ChartCard from '../components/ChartCard'
import StatusBadge from '../components/StatusBadge'
import Loading from '../components/Loading'

export default function DataQualityPage() {
  const { data, loading } = useMetrics(api.healthChecks)

  if (loading || !data) return <Loading />

  const passing = data.filter((c) => c.status === 'pass').length
  const failing = data.filter((c) => c.status === 'fail').length
  const warnings = data.filter((c) => c.status === 'warn').length

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Data Quality</h1>

      <div className="grid grid-cols-3 gap-4">
        <StatCard title="Passing" value={passing} color="text-green-400" />
        <StatCard title="Warnings" value={warnings} color="text-yellow-400" />
        <StatCard title="Failing" value={failing} color="text-red-400" />
      </div>

      <ChartCard title="Health Checks">
        <div className="space-y-3">
          {data.map((check) => (
            <div
              key={check.name}
              className="flex items-center justify-between py-2 border-b border-[#3A3A48]/50 last:border-0"
            >
              <div>
                <p className="text-sm text-[#F0F0F5]/80">{check.name}</p>
                <p className="text-xs text-[#F0F0F5]/50">{check.message}</p>
              </div>
              <div className="flex items-center gap-3">
                {check.count > 0 && (
                  <span className="text-sm text-[#F0F0F5]/60">{check.count}</span>
                )}
                <StatusBadge status={check.status} />
              </div>
            </div>
          ))}
        </div>
      </ChartCard>
    </div>
  )
}
