import { api } from '../lib/api'
import { useMetrics } from '../hooks/useMetrics'
import { formatNumber, formatDollars, formatPercent, shortDate } from '../lib/format'
import StatCard from '../components/StatCard'
import ChartCard from '../components/ChartCard'
import DataTable from '../components/DataTable'
import ProgressBar from '../components/ProgressBar'
import Loading from '../components/Loading'
import {
  ComposedChart, Area, Bar, XAxis, YAxis, Tooltip, Legend, ResponsiveContainer,
} from 'recharts'

// Pretty platform names; falls back to the raw value so a new platform still shows.
const PLATFORM_LABELS: Record<string, string> = {
  tiktok: 'TikTok',
  instagram: 'Instagram',
  facebook: 'Facebook',
  youtube: 'YouTube',
  pinterest: 'Pinterest',
}

function platformLabel(p: string): string {
  return PLATFORM_LABELS[(p ?? '').toLowerCase()] ?? (p || 'Unknown')
}

// Lifecycle status -> friendly label + color (matches VideoImportStatus values).
const STATUS_LABELS: Record<string, string> = {
  done: 'Success',
  failed: 'Failed',
  processing: 'Processing',
  queued: 'Queued',
}

function statusLabel(s: string): string {
  return STATUS_LABELS[(s ?? '').toLowerCase()] ?? (s || 'Unknown')
}

function statusColor(s: string): string {
  const k = (s ?? '').toLowerCase()
  if (k === 'done') return 'text-emerald-400'
  if (k === 'failed') return 'text-red-400'
  return 'text-[#F0F0F5]/80'
}

function tooltipFormatter(value: unknown, name: unknown): [string, string] {
  const label = String(name)
  if (label === 'Spend') return [formatDollars(Number(value)), label]
  return [formatNumber(Number(value)), label]
}

export default function VideoEconPage() {
  const { data, loading } = useMetrics(api.videoEconomics)

  if (loading && !data) return <Loading />

  // Table only exists once premium video import is enabled, so until then (or
  // before any imports) show a clean empty state.
  if (!data || data.total_imports === 0) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">Video Economics</h1>
        <div className="bg-[#1E1E28] rounded-lg p-12 border border-[#3A3A48] text-center">
          <p className="text-[#F0F0F5]/70">No video imports recorded yet</p>
          <p className="text-sm text-[#F0F0F5]/40 mt-1">
            Premium video-to-recipe spend, cache savings, and success rates will appear here once the feature is used.
          </p>
        </div>
      </div>
    )
  }

  const byPlatform = [...(data.by_platform ?? [])].sort((a, b) => b.imports - a.imports)
  const byStatus = [...(data.by_status ?? [])].sort((a, b) => b.count - a.count)
  const daily = data.daily ?? []

  const budget = data.daily_budget_usd
  const overBudget = budget > 0 && data.today_spend_usd >= budget

  return (
    <div className="space-y-6">
      <div className="flex items-baseline justify-between">
        <h1 className="text-2xl font-bold">Video Economics</h1>
        <span className="text-sm text-[#F0F0F5]/50">Last {data.period_days} days</span>
      </div>

      {/* KPI cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard
          title="Video Imports"
          value={formatNumber(data.total_imports)}
          subtitle={`${formatNumber(data.paid_imports)} paid / ${formatNumber(data.cache_hits)} cached`}
        />
        <StatCard
          title="Total Spend"
          value={formatDollars(data.total_spend_usd)}
          subtitle={`avg ${formatDollars(data.avg_cost_per_paid_import)} / paid import`}
          color="text-green-400"
        />
        <StatCard
          title="Success Rate"
          value={formatPercent(data.success_rate)}
          subtitle={`${formatNumber(data.success_count)} ok / ${formatNumber(data.failed_count)} failed`}
        />
        <StatCard
          title="Cache-Hit Rate"
          value={formatPercent(data.cache_hit_rate)}
          subtitle={`${formatNumber(data.cache_hits)} served free`}
          color="text-emerald-400"
        />
      </div>

      {/* Today's spend vs the daily budget kill switch */}
      <ChartCard title="Today's Spend vs Daily Budget">
        <div className="flex items-baseline justify-between mb-3">
          <span className={`text-2xl font-bold ${overBudget ? 'text-red-400' : 'text-[#F0F0F5]'}`}>
            {formatDollars(data.today_spend_usd)}
          </span>
          <span className="text-sm text-[#F0F0F5]/50">of {formatDollars(budget)} budget</span>
        </div>
        <ProgressBar
          value={data.today_spend_usd}
          max={budget}
          color={overBudget ? 'bg-red-500' : 'bg-[#FF6B85]'}
        />
        {overBudget && (
          <p className="text-xs text-red-400 mt-2">Daily budget reached — new metered imports are blocked.</p>
        )}
      </ChartCard>

      {/* Daily spend + volume */}
      {daily.length > 0 && (
        <ChartCard title="Daily Spend & Imports">
          <ResponsiveContainer width="100%" height={240}>
            <ComposedChart data={daily}>
              <defs>
                <linearGradient id="videoSpendGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#FF6B85" stopOpacity={0.4} />
                  <stop offset="95%" stopColor="#FF6B85" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis dataKey="date" tickFormatter={shortDate} stroke="#94a3b8" fontSize={12} />
              <YAxis
                yAxisId="spend"
                stroke="#94a3b8"
                fontSize={12}
                tickFormatter={(v) => '$' + Number(v).toFixed(2)}
              />
              <YAxis
                yAxisId="count"
                orientation="right"
                stroke="#94a3b8"
                fontSize={12}
                tickFormatter={(v) => formatNumber(Number(v))}
              />
              <Tooltip
                contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }}
                labelFormatter={shortDate}
                formatter={tooltipFormatter}
              />
              <Legend wrapperStyle={{ fontSize: 12 }} />
              <Bar yAxisId="count" dataKey="imports" name="Imports" fill="#34D399" opacity={0.5} radius={[2, 2, 0, 0]} />
              <Area
                yAxisId="spend"
                type="monotone"
                dataKey="spend"
                name="Spend"
                stroke="#FF6B85"
                strokeWidth={2}
                fill="url(#videoSpendGradient)"
              />
            </ComposedChart>
          </ResponsiveContainer>
        </ChartCard>
      )}

      {/* Per-platform economics */}
      {byPlatform.length > 0 && (
        <ChartCard title="Spend by Platform">
          <DataTable
            columns={[
              {
                key: 'platform',
                label: 'Platform',
                render: (r: Record<string, unknown>) => platformLabel(String(r.platform)),
              },
              { key: 'imports', label: 'Imports', render: (r: Record<string, unknown>) => formatNumber(Number(r.imports)) },
              { key: 'cost_usd', label: 'Spend', render: (r: Record<string, unknown>) => formatDollars(Number(r.cost_usd)) },
              { key: 'cache_hits', label: 'Cache Hits', render: (r: Record<string, unknown>) => formatNumber(Number(r.cache_hits)) },
              {
                key: 'failures',
                label: 'Failures',
                render: (r: Record<string, unknown>) => {
                  const f = Number(r.failures)
                  return <span className={f > 0 ? 'text-red-400' : 'text-[#F0F0F5]/80'}>{formatNumber(f)}</span>
                },
              },
            ]}
            data={byPlatform as unknown as Record<string, unknown>[]}
          />
        </ChartCard>
      )}

      {/* Lifecycle status breakdown */}
      {byStatus.length > 0 && (
        <ChartCard title="Imports by Status">
          <DataTable
            columns={[
              {
                key: 'label',
                label: 'Status',
                render: (r: Record<string, unknown>) => (
                  <span className={statusColor(String(r.label))}>{statusLabel(String(r.label))}</span>
                ),
              },
              { key: 'count', label: 'Imports', render: (r: Record<string, unknown>) => formatNumber(Number(r.count)) },
            ]}
            data={byStatus as unknown as Record<string, unknown>[]}
          />
        </ChartCard>
      )}
    </div>
  )
}
