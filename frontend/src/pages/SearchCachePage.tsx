import { api } from '../lib/api'
import { useMetrics } from '../hooks/useMetrics'
import { formatNumber, formatPercent, shortDate } from '../lib/format'
import StatCard from '../components/StatCard'
import ChartCard from '../components/ChartCard'
import ProgressBar from '../components/ProgressBar'
import DataTable from '../components/DataTable'
import Loading from '../components/Loading'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'

export default function SearchCachePage() {
  const { data, loading } = useMetrics(api.searchCache)

  if (loading || !data) return <Loading />

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Search & Cache</h1>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard title="Cached Queries" value={formatNumber(data.total_cached_queries)} />
        <StatCard title="Avg Hit Count" value={data.avg_hit_count.toFixed(1)} />
        <StatCard title="Zero-Hit Entries" value={data.zero_hit_entries} />
        <StatCard title="Avg Results/Query" value={data.avg_results_per_query.toFixed(1)} />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <ChartCard title="Embedding Coverage">
          <ProgressBar
            value={data.embedding_coverage}
            max={100}
            label={`${data.entries_with_embeddings} / ${data.total_cached_queries}`}
            color="bg-emerald-500"
          />
        </ChartCard>

        <ChartCard title="Cache Health">
          <div className="space-y-3">
            <div className="flex justify-between text-sm">
              <span className="text-slate-400">Approaching TTL (22-24h)</span>
              <span className="text-yellow-400">{data.entries_approaching_ttl}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-slate-400">Stale (past 24h)</span>
              <span className="text-red-400">{data.stale_entries}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-slate-400">Hot queries (eligible refresh)</span>
              <span className="text-blue-400">{data.hot_queries_eligible}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-slate-400">Stale 30d (pending cleanup)</span>
              <span className="text-slate-500">{data.stale_entries_30d}</span>
            </div>
          </div>
        </ChartCard>
      </div>

      {data.daily_volume.length > 0 && (
        <ChartCard title="Daily New Cache Entries">
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={data.daily_volume}>
              <XAxis dataKey="date" tickFormatter={shortDate} stroke="#94a3b8" fontSize={12} />
              <YAxis stroke="#94a3b8" fontSize={12} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155' }}
                labelFormatter={shortDate}
              />
              <Bar dataKey="count" fill="#3b82f6" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      )}

      {data.top_queries.length > 0 && (
        <ChartCard title="Top Queries by Hit Count">
          <DataTable
            columns={[
              { key: 'label', label: 'Query' },
              { key: 'count', label: 'Hits' },
            ]}
            data={data.top_queries as unknown as Record<string, unknown>[]}
            maxRows={25}
          />
        </ChartCard>
      )}
    </div>
  )
}
