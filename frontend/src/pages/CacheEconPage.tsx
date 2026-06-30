import { api } from '../lib/api'
import { useMetrics } from '../hooks/useMetrics'
import { formatNumber, formatDollars, formatPercent, shortDate } from '../lib/format'
import StatCard from '../components/StatCard'
import ChartCard from '../components/ChartCard'
import DataTable from '../components/DataTable'
import Loading from '../components/Loading'
import {
  AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer,
} from 'recharts'

// Estimate to 4dp — savings per extraction are sub-cent, so formatDollars
// ($x.xx) would collapse the assumption to "$0.00".
function estDollars(n: number): string {
  return '$' + n.toFixed(4)
}

function tooltipNew(value: unknown): [string, string] {
  return [formatNumber(Number(value)), 'New entries']
}

// FREE = json_ld (no AI). PAID = haiku / firecrawl_haiku (AI). Anything else
// (e.g. firecrawl_json_ld — structured data via a paid scrape) is neutral.
function methodColor(label: string): string {
  if (label === 'json_ld') return '#5CFFD4'
  if (label === 'haiku' || label === 'firecrawl_haiku') return '#FF6B85'
  return '#7A7A8C'
}

function methodKind(label: string): string {
  if (label === 'json_ld') return 'FREE'
  if (label === 'haiku' || label === 'firecrawl_haiku') return 'PAID · AI'
  return 'other'
}

export default function CacheEconPage() {
  const { data, loading } = useMetrics(api.cacheEconomics)

  if (loading && !data) return <Loading />

  // Nothing cached yet => nothing to quantify.
  if (!data || (data.total_canonical_entries === 0 && data.total_search_cache_entries === 0)) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">Cache Economics</h1>
        <div className="bg-[#1E1E28] rounded-lg p-12 border border-[#3A3A48] text-center">
          <p className="text-[#F0F0F5]/70">No cache entries yet</p>
          <p className="text-sm text-[#F0F0F5]/40 mt-1">
            Savings from the extraction and search caches will appear here once recipes are cached.
          </p>
        </div>
      </div>
    )
  }

  const mix = [...(data.extraction_method_mix ?? [])].sort((a, b) => b.count - a.count)
  const maxMixCount = mix.reduce((m, r) => Math.max(m, r.count), 0)
  const dailyNew = data.daily_new_canonical ?? []

  // Savings components, for the breakdown bars.
  const savingsRows = [
    {
      label: 'Free extractions (skipped AI)',
      detail: `${formatNumber(data.free_entries)} × ${estDollars(data.assumed_ai_extraction_usd)}`,
      value: data.estimated_free_savings_usd,
    },
    ...(data.has_hit_count
      ? [
          {
            label: 'Cache reuse (skipped re-extraction)',
            detail: `${formatNumber(data.ai_reuse_hits)} AI hits × ${estDollars(data.assumed_ai_extraction_usd)}`,
            value: data.estimated_reuse_savings_usd,
          },
        ]
      : []),
  ]
  const maxSaving = savingsRows.reduce((m, r) => Math.max(m, r.value), 0)

  return (
    <div className="space-y-6">
      <div className="flex items-baseline justify-between">
        <h1 className="text-2xl font-bold">Cache Economics</h1>
        <span className="text-sm text-[#F0F0F5]/50">Estimated savings · all $ figures are estimates</span>
      </div>

      {/* Headline KPIs */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
        <StatCard title="Canonical Entries" value={formatNumber(data.total_canonical_entries)} subtitle="cached extractions" />
        <StatCard
          title="Extracted Free"
          value={formatPercent(data.free_pct)}
          subtitle={`${formatNumber(data.free_entries)} via json_ld`}
          color="text-emerald-400"
        />
        <StatCard
          title="Paid AI Extraction"
          value={formatPercent(data.paid_ai_pct)}
          subtitle={`${formatNumber(data.paid_ai_entries)} via Haiku`}
          color="text-[#FF6B85]"
        />
        <StatCard
          title="Est. $ Saved"
          value={formatDollars(data.estimated_total_saved_usd)}
          subtitle="estimate"
          color="text-emerald-400"
        />
        <StatCard title="Search Cache" value={formatNumber(data.total_search_cache_entries)} subtitle="cached queries" />
      </div>

      {/* Estimated savings breakdown */}
      <ChartCard title="Estimated Savings Breakdown">
        <p className="text-xs text-[#F0F0F5]/50 mb-4 -mt-1">
          ESTIMATE only. Assumes one avoided AI extraction is worth{' '}
          <span className="text-[#F0F0F5]/80">{estDollars(data.assumed_ai_extraction_usd)}</span> (avg Haiku extraction).
        </p>
        <div className="space-y-4">
          {savingsRows.map((row) => {
            const width = maxSaving > 0 ? Math.max((row.value / maxSaving) * 100, 2) : 0
            return (
              <div key={row.label}>
                <div className="flex items-center justify-between mb-1.5 text-sm">
                  <div className="flex items-center gap-2">
                    <span className="text-[#F0F0F5]/85">{row.label}</span>
                    <span className="px-1.5 py-0.5 rounded text-[10px] uppercase tracking-wide bg-[#5CFFD4]/15 text-emerald-400 border border-[#5CFFD4]/30">
                      Est
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-[#F0F0F5]/40 text-xs">{row.detail}</span>
                    <span className="tabular-nums text-emerald-400 w-16 text-right">{formatDollars(row.value)}</span>
                  </div>
                </div>
                <div className="h-2.5 rounded-full bg-[#2A2A36] overflow-hidden">
                  <div className="h-full rounded-full" style={{ width: `${width}%`, backgroundColor: '#5CFFD4' }} />
                </div>
              </div>
            )
          })}
          <div className="flex items-center justify-between pt-2 border-t border-[#3A3A48] text-sm">
            <span className="text-[#F0F0F5]/85 font-medium">Total estimated saved</span>
            <span className="tabular-nums text-emerald-400 font-semibold">
              {formatDollars(data.estimated_total_saved_usd)}
            </span>
          </div>
        </div>
      </ChartCard>

      {/* Extraction-method mix */}
      {mix.length > 0 && (
        <ChartCard title="Extraction Method Mix">
          <p className="text-xs text-[#F0F0F5]/50 mb-4 -mt-1">
            How each cached recipe was extracted. <span className="text-emerald-400">Green</span> is free (structured data);{' '}
            <span className="text-[#FF6B85]">pink</span> used a paid AI model.
          </p>
          <div className="space-y-4 mb-5">
            {mix.map((row) => {
              const width = maxMixCount > 0 ? Math.max((row.count / maxMixCount) * 100, 2) : 0
              const share = data.total_canonical_entries > 0 ? (row.count / data.total_canonical_entries) * 100 : 0
              return (
                <div key={row.label}>
                  <div className="flex items-center justify-between mb-1.5 text-sm">
                    <div className="flex items-center gap-2">
                      <span className="text-[#F0F0F5]/85">{row.label || '(unknown)'}</span>
                      <span className="text-[10px] uppercase tracking-wide text-[#F0F0F5]/40">{methodKind(row.label)}</span>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="tabular-nums text-[#F0F0F5]/85">{formatNumber(row.count)}</span>
                      <span className="tabular-nums w-14 text-right text-[#F0F0F5]/40">{formatPercent(share)}</span>
                    </div>
                  </div>
                  <div className="h-2.5 rounded-full bg-[#2A2A36] overflow-hidden">
                    <div className="h-full rounded-full" style={{ width: `${width}%`, backgroundColor: methodColor(row.label) }} />
                  </div>
                </div>
              )
            })}
          </div>
          <DataTable
            columns={[
              { key: 'label', label: 'Method', render: (r: Record<string, unknown>) => String(r.label || '(unknown)') },
              { key: 'kind', label: 'Type', render: (r: Record<string, unknown>) => methodKind(String(r.label)) },
              { key: 'count', label: 'Entries', render: (r: Record<string, unknown>) => formatNumber(Number(r.count)) },
              {
                key: 'share',
                label: 'Share',
                render: (r: Record<string, unknown>) =>
                  formatPercent(
                    data.total_canonical_entries > 0 ? (Number(r.count) / data.total_canonical_entries) * 100 : 0,
                  ),
              },
            ]}
            data={mix as unknown as Record<string, unknown>[]}
          />
        </ChartCard>
      )}

      {/* Multi-page markers vs real cached recipes */}
      {data.has_multi_page && (
        <ChartCard title="Collection Markers vs Cached Recipes">
          <p className="text-xs text-[#F0F0F5]/50 mb-4 -mt-1">
            Multi-page collection/listicle URLs are stored as markers (no recipe body), separate from real cached recipes.
          </p>
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-[#121218] rounded-lg p-4 border border-[#3A3A48]">
              <p className="text-sm text-[#F0F0F5]/60">Real cached recipes</p>
              <p className="text-2xl font-bold mt-1 text-emerald-400">{formatNumber(data.single_recipe_entries)}</p>
            </div>
            <div className="bg-[#121218] rounded-lg p-4 border border-[#3A3A48]">
              <p className="text-sm text-[#F0F0F5]/60">Collection markers</p>
              <p className="text-2xl font-bold mt-1 text-[#F0F0F5]/70">{formatNumber(data.multi_page_markers)}</p>
            </div>
          </div>
        </ChartCard>
      )}

      {/* Canonical cache growth */}
      {dailyNew.length > 0 && (
        <ChartCard title="Canonical Cache Growth (Last 30 Days)">
          <ResponsiveContainer width="100%" height={220}>
            <AreaChart data={dailyNew}>
              <defs>
                <linearGradient id="cacheGrowthGradient" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#FF6B85" stopOpacity={0.4} />
                  <stop offset="95%" stopColor="#FF6B85" stopOpacity={0} />
                </linearGradient>
              </defs>
              <XAxis dataKey="date" tickFormatter={shortDate} stroke="#94a3b8" fontSize={12} />
              <YAxis stroke="#94a3b8" fontSize={12} allowDecimals={false} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }}
                labelFormatter={shortDate}
                formatter={tooltipNew}
              />
              <Area type="monotone" dataKey="count" stroke="#FF6B85" strokeWidth={2} fill="url(#cacheGrowthGradient)" />
            </AreaChart>
          </ResponsiveContainer>
        </ChartCard>
      )}
    </div>
  )
}
