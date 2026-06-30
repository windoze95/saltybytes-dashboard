import { api } from '../lib/api'
import { useMetrics } from '../hooks/useMetrics'
import { formatNumber, formatPercent, shortDate } from '../lib/format'
import StatCard from '../components/StatCard'
import ChartCard from '../components/ChartCard'
import DataTable from '../components/DataTable'
import Loading from '../components/Loading'
import {
  AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer,
} from 'recharts'

function formatMs(n: number): string {
  return `${Math.round(n)} ms`
}

function tooltipCalls(value: unknown): [string, string] {
  return [formatNumber(Number(value)), 'Calls']
}

// Higher is better: emerald when essentially clean, amber when a little lossy,
// red once failures are material.
function successColor(rate: number): string {
  if (rate >= 99) return 'text-emerald-400'
  if (rate >= 95) return 'text-amber-400'
  return 'text-red-400'
}

// A DataTable cell renderer that formats a numeric millisecond column.
const msCell = (key: string) => (r: Record<string, unknown>) => formatMs(Number(r[key]))

export default function AIOpsPage() {
  const { data, loading } = useMetrics(api.aiOps)

  if (loading && !data) return <Loading />

  // Endpoint returns null (or zero activity) until usage accrues.
  if (!data || data.total_calls === 0) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">AI Operations</h1>
        <div className="bg-[#1E1E28] rounded-lg p-12 border border-[#3A3A48] text-center">
          <p className="text-[#F0F0F5]/70">No AI activity recorded yet</p>
          <p className="text-sm text-[#F0F0F5]/40 mt-1">
            Latency, reliability and throughput will appear here once AI calls are logged.
          </p>
        </div>
      </div>
    )
  }

  const latencyByOperation = [...(data.latency_by_operation ?? [])].sort((a, b) => b.avg_ms - a.avg_ms)
  const latencyByModel = [...(data.latency_by_model ?? [])].sort((a, b) => b.avg_ms - a.avg_ms)
  const reliabilityByOperation = [...(data.reliability_by_operation ?? [])].sort((a, b) => b.failures - a.failures)
  const reliabilityByModel = [...(data.reliability_by_model ?? [])].sort((a, b) => b.failures - a.failures)
  const callsPerDay = data.calls_per_day ?? []
  const slowest = [...(data.slowest_operations ?? [])].sort((a, b) => b.avg_ms - a.avg_ms)
  const maxSlow = slowest.reduce((m, s) => Math.max(m, s.avg_ms), 0)

  const reliabilityColumns = (firstLabel: string) => [
    { key: 'label', label: firstLabel },
    { key: 'calls', label: 'Calls', render: (r: Record<string, unknown>) => formatNumber(Number(r.calls)) },
    { key: 'successes', label: 'Successes', render: (r: Record<string, unknown>) => formatNumber(Number(r.successes)) },
    {
      key: 'failures',
      label: 'Failures',
      render: (r: Record<string, unknown>) => {
        const f = Number(r.failures)
        return <span className={f > 0 ? 'text-red-400' : 'text-[#F0F0F5]/80'}>{f}</span>
      },
    },
    {
      key: 'error_rate',
      label: 'Error Rate',
      render: (r: Record<string, unknown>) => {
        const e = Number(r.error_rate)
        return <span className={e > 0 ? 'text-red-400' : 'text-emerald-400'}>{formatPercent(e)}</span>
      },
    },
  ]

  const latencyColumns = (firstLabel: string) => [
    { key: 'label', label: firstLabel },
    { key: 'calls', label: 'Calls', render: (r: Record<string, unknown>) => formatNumber(Number(r.calls)) },
    { key: 'avg_ms', label: 'Avg', render: msCell('avg_ms') },
    { key: 'p50_ms', label: 'p50', render: msCell('p50_ms') },
    { key: 'p95_ms', label: 'p95', render: msCell('p95_ms') },
    { key: 'p99_ms', label: 'p99', render: msCell('p99_ms') },
  ]

  return (
    <div className="space-y-6">
      <div className="flex items-baseline justify-between">
        <h1 className="text-2xl font-bold">AI Operations</h1>
        <span className="text-sm text-[#F0F0F5]/50">Last {data.period_days} days</span>
      </div>

      {/* KPI cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard
          title="Total Calls"
          value={formatNumber(data.total_calls)}
          subtitle={`Last ${data.period_days} days`}
        />
        <StatCard
          title="Success Rate"
          value={formatPercent(data.success_rate)}
          color={successColor(data.success_rate)}
        />
        <StatCard title="Avg Latency" value={formatMs(data.avg_latency_ms)} />
        <StatCard title="p95 Latency" value={formatMs(data.p95_latency_ms)} />
      </div>

      {/* Centerpiece: which operations are slowest by average latency */}
      {slowest.length > 0 && (
        <ChartCard title="Slowest Operations">
          <p className="text-xs text-[#F0F0F5]/50 mb-4 -mt-1">
            Top operations by average latency. Longer bars are slower; p95 shows the tail.
          </p>
          <div className="space-y-4">
            {slowest.map((s) => {
              const width = maxSlow > 0 ? Math.max((s.avg_ms / maxSlow) * 100, 2) : 0
              return (
                <div key={s.label}>
                  <div className="flex items-center justify-between mb-1.5 text-sm">
                    <span className="text-[#F0F0F5]/85">{s.label}</span>
                    <div className="flex items-center gap-3">
                      <span className="tabular-nums text-[#F0F0F5]/85">{formatMs(s.avg_ms)}</span>
                      <span className="tabular-nums w-28 text-right text-[#F0F0F5]/40">
                        p95 {formatMs(s.p95_ms)}
                      </span>
                    </div>
                  </div>
                  <div className="h-2.5 rounded-full bg-[#2A2A36] overflow-hidden">
                    <div
                      className="h-full rounded-full"
                      style={{ width: `${width}%`, backgroundColor: '#FFB85C' }}
                    />
                  </div>
                </div>
              )
            })}
          </div>
        </ChartCard>
      )}

      {/* Latency by operation */}
      {latencyByOperation.length > 0 && (
        <ChartCard title="Latency by Operation">
          <DataTable
            columns={latencyColumns('Operation')}
            data={latencyByOperation as unknown as Record<string, unknown>[]}
          />
        </ChartCard>
      )}

      {/* Latency by model */}
      {latencyByModel.length > 0 && (
        <ChartCard title="Latency by Model">
          <DataTable
            columns={latencyColumns('Model')}
            data={latencyByModel as unknown as Record<string, unknown>[]}
          />
        </ChartCard>
      )}

      {/* Reliability by operation */}
      {reliabilityByOperation.length > 0 && (
        <ChartCard title="Reliability by Operation">
          <DataTable
            columns={reliabilityColumns('Operation')}
            data={reliabilityByOperation as unknown as Record<string, unknown>[]}
          />
        </ChartCard>
      )}

      {/* Reliability by model */}
      {reliabilityByModel.length > 0 && (
        <ChartCard title="Reliability by Model">
          <DataTable
            columns={reliabilityColumns('Model')}
            data={reliabilityByModel as unknown as Record<string, unknown>[]}
          />
        </ChartCard>
      )}

      {/* Throughput */}
      {callsPerDay.length > 0 && (
        <ChartCard title="Calls per Day">
          <ResponsiveContainer width="100%" height={220}>
            <AreaChart data={callsPerDay}>
              <defs>
                <linearGradient id="aiOpsCallsGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#FF6B85" stopOpacity={0.4} />
                  <stop offset="95%" stopColor="#FF6B85" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis dataKey="date" tickFormatter={shortDate} stroke="#94a3b8" fontSize={12} />
              <YAxis stroke="#94a3b8" fontSize={12} allowDecimals={false} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }}
                labelFormatter={shortDate}
                formatter={tooltipCalls}
              />
              <Area
                type="monotone"
                dataKey="count"
                stroke="#FF6B85"
                strokeWidth={2}
                fill="url(#aiOpsCallsGradient)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </ChartCard>
      )}
    </div>
  )
}
