import { api } from '../lib/api'
import { useMetrics } from '../hooks/useMetrics'
import { formatNumber, formatDollars, shortDate } from '../lib/format'
import StatCard from '../components/StatCard'
import ChartCard from '../components/ChartCard'
import DataTable from '../components/DataTable'
import Loading from '../components/Loading'
import {
  AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer,
} from 'recharts'

// Signed percent for counterfactual deltas: -79.2 -> "−79%" (uses U+2212 minus).
function formatSignedPercent(n: number, decimals = 0): string {
  const v = Number(n.toFixed(decimals))
  if (v === 0) return '0%'
  return (v > 0 ? '+' : '−') + Math.abs(v).toFixed(decimals) + '%'
}

function tooltipSpend(value: unknown): [string, string] {
  return [formatDollars(Number(value)), 'Spend']
}

export default function AIModelsPage() {
  const { data, loading } = useMetrics(api.aiModels)

  if (loading && !data) return <Loading />

  // Endpoint returns null (or zero activity) until usage accrues.
  if (!data || data.total_calls === 0) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">AI Models</h1>
        <div className="bg-[#1E1E28] rounded-lg p-12 border border-[#3A3A48] text-center">
          <p className="text-[#F0F0F5]/70">No AI usage recorded yet</p>
          <p className="text-sm text-[#F0F0F5]/40 mt-1">
            Model spend and counterfactuals will appear here once AI calls are logged.
          </p>
        </div>
      </div>
    )
  }

  const totalTokens = data.total_input_tokens + data.total_output_tokens

  // Counterfactuals: cheapest -> priciest. vs_actual_pct === 0 marks the current/actual model.
  const sortedCf = [...(data.counterfactuals ?? [])].sort((a, b) => a.cost_usd - b.cost_usd)
  const maxCfCost = sortedCf.reduce((m, c) => Math.max(m, c.cost_usd), 0)
  const cheapest = sortedCf[0] ?? null
  const hasSavings = !!cheapest && cheapest.vs_actual_pct < 0

  const byModel = [...(data.by_model ?? [])].sort((a, b) => b.cost_usd - a.cost_usd)
  const byOperation = [...(data.by_operation ?? [])].sort((a, b) => b.cost_usd - a.cost_usd)
  const dailySpend = data.daily_spend ?? []

  return (
    <div className="space-y-6">
      <div className="flex items-baseline justify-between">
        <h1 className="text-2xl font-bold">AI Models</h1>
        <span className="text-sm text-[#F0F0F5]/50">Last {data.period_days} days</span>
      </div>

      {/* KPI cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard
          title="Total AI Spend"
          value={formatDollars(data.total_cost_usd)}
          subtitle={`Last ${data.period_days} days`}
          color="text-green-400"
        />
        <StatCard title="Total Calls" value={formatNumber(data.total_calls)} />
        <StatCard
          title="Total Tokens"
          value={formatNumber(totalTokens)}
          subtitle={`${formatNumber(data.total_input_tokens)} in / ${formatNumber(data.total_output_tokens)} out`}
        />
        {cheapest ? (
          <StatCard
            title={hasSavings ? 'Biggest Savings' : 'Cheapest Alternative'}
            value={hasSavings ? formatSignedPercent(cheapest.vs_actual_pct) : 'Actual is lowest'}
            subtitle={hasSavings ? `${cheapest.label} vs actual` : undefined}
            color={hasSavings ? 'text-emerald-400' : 'text-[#F0F0F5]'}
          />
        ) : (
          <StatCard title="Output Tokens" value={formatNumber(data.total_output_tokens)} />
        )}
      </div>

      {/* Centerpiece: what the recorded token volume would have cost on each candidate model */}
      {sortedCf.length > 0 && (
        <ChartCard title="What it would've cost">
          <p className="text-xs text-[#F0F0F5]/50 mb-4 -mt-1">
            This period's token volume priced on each model, cheapest first. Green is cheaper than what we actually paid.
          </p>
          <div className="space-y-4">
            {sortedCf.map((c) => {
              const isActual = c.vs_actual_pct === 0
              const barColor = isActual ? '#FF6B85' : c.vs_actual_pct < 0 ? '#5CFFD4' : '#FF8A8A'
              const width = maxCfCost > 0 ? Math.max((c.cost_usd / maxCfCost) * 100, 2) : 0
              return (
                <div key={c.label}>
                  <div className="flex items-center justify-between mb-1.5 text-sm">
                    <div className="flex items-center gap-2">
                      <span className="text-[#F0F0F5]/85">{c.label}</span>
                      {isActual && (
                        <span className="px-1.5 py-0.5 rounded text-[10px] uppercase tracking-wide bg-[#FF6B85]/20 text-[#FF6B85] border border-[#FF6B85]/40">
                          Actual
                        </span>
                      )}
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="tabular-nums text-[#F0F0F5]/85">{formatDollars(c.cost_usd)}</span>
                      <span
                        className={`tabular-nums w-16 text-right ${
                          isActual
                            ? 'text-[#F0F0F5]/40'
                            : c.vs_actual_pct < 0
                              ? 'text-emerald-400'
                              : 'text-red-400'
                        }`}
                      >
                        {isActual ? '—' : formatSignedPercent(c.vs_actual_pct, 1)}
                      </span>
                    </div>
                  </div>
                  <div className="h-2.5 rounded-full bg-[#2A2A36] overflow-hidden">
                    <div
                      className="h-full rounded-full"
                      style={{ width: `${width}%`, backgroundColor: barColor }}
                    />
                  </div>
                </div>
              )
            })}
          </div>
        </ChartCard>
      )}

      {/* Spend by model */}
      {byModel.length > 0 && (
        <ChartCard title="Spend by Model">
          <DataTable
            columns={[
              { key: 'model', label: 'Model' },
              { key: 'provider', label: 'Provider' },
              { key: 'calls', label: 'Calls', render: (r: Record<string, unknown>) => formatNumber(Number(r.calls)) },
              { key: 'input_tokens', label: 'Input', render: (r: Record<string, unknown>) => formatNumber(Number(r.input_tokens)) },
              { key: 'output_tokens', label: 'Output', render: (r: Record<string, unknown>) => formatNumber(Number(r.output_tokens)) },
              { key: 'cost_usd', label: 'Cost', render: (r: Record<string, unknown>) => formatDollars(Number(r.cost_usd)) },
              { key: 'avg_latency_ms', label: 'Avg Latency', render: (r: Record<string, unknown>) => `${Math.round(Number(r.avg_latency_ms))} ms` },
              {
                key: 'failures',
                label: 'Failures',
                render: (r: Record<string, unknown>) => {
                  const f = Number(r.failures)
                  return <span className={f > 0 ? 'text-red-400' : 'text-[#F0F0F5]/80'}>{f}</span>
                },
              },
            ]}
            data={byModel as unknown as Record<string, unknown>[]}
          />
        </ChartCard>
      )}

      {/* Spend by operation */}
      {byOperation.length > 0 && (
        <ChartCard title="Spend by Operation">
          <DataTable
            columns={[
              { key: 'operation', label: 'Operation' },
              { key: 'calls', label: 'Calls', render: (r: Record<string, unknown>) => formatNumber(Number(r.calls)) },
              { key: 'cost_usd', label: 'Cost', render: (r: Record<string, unknown>) => formatDollars(Number(r.cost_usd)) },
            ]}
            data={byOperation as unknown as Record<string, unknown>[]}
          />
        </ChartCard>
      )}

      {/* Daily spend */}
      {dailySpend.length > 0 && (
        <ChartCard title="Daily Spend">
          <ResponsiveContainer width="100%" height={220}>
            <AreaChart data={dailySpend}>
              <defs>
                <linearGradient id="aiSpendGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#FF6B85" stopOpacity={0.4} />
                  <stop offset="95%" stopColor="#FF6B85" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis dataKey="date" tickFormatter={shortDate} stroke="#94a3b8" fontSize={12} />
              <YAxis stroke="#94a3b8" fontSize={12} tickFormatter={(v) => '$' + Number(v).toFixed(2)} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }}
                labelFormatter={shortDate}
                formatter={tooltipSpend}
              />
              <Area
                type="monotone"
                dataKey="amount"
                stroke="#FF6B85"
                strokeWidth={2}
                fill="url(#aiSpendGradient)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </ChartCard>
      )}
    </div>
  )
}
