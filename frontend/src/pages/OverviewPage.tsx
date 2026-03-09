import { api } from '../lib/api'
import { useMetrics } from '../hooks/useMetrics'
import { formatNumber, formatDollars, formatPercent, shortDate } from '../lib/format'
import StatCard from '../components/StatCard'
import ChartCard from '../components/ChartCard'
import ProgressBar from '../components/ProgressBar'
import StatusBadge from '../components/StatusBadge'
import Loading from '../components/Loading'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'

export default function OverviewPage() {
  const { data, loading } = useMetrics(api.overview)
  const { data: healthChecks } = useMetrics(api.healthChecks)

  if (loading || !data) return <Loading />

  const healthColor =
    data.health_checks_passing === data.health_checks_total
      ? 'text-green-400'
      : data.health_checks_passing >= data.health_checks_total - 2
        ? 'text-yellow-400'
        : 'text-red-400'

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Overview</h1>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard title="Total Users" value={formatNumber(data.total_users)} />
        <StatCard title="Total Recipes" value={formatNumber(data.total_recipes)} />
        <StatCard title="Canonical Entries" value={formatNumber(data.total_canonicals)} />
        <StatCard
          title="Health Checks"
          value={`${data.health_checks_passing}/${data.health_checks_total}`}
          color={healthColor}
        />
      </div>

      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
        <StatCard
          title="Today's Est. Cost"
          value={formatDollars(data.today_estimated_cost)}
          color="text-green-400"
        />
        <StatCard
          title="Month-to-Date"
          value={formatDollars(data.month_to_date_cost)}
          color="text-green-400"
        />
        <StatCard
          title="Month Projection"
          value={formatDollars(data.month_projection)}
          subtitle="Linear extrapolation"
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <ChartCard title="Search Cache Hit Rate">
          <div className="flex items-center gap-4">
            <ProgressBar
              value={data.search_cache_hit_rate}
              max={100}
              label="Hit Rate"
              color="bg-emerald-500"
            />
          </div>
        </ChartCard>

        <ChartCard title="Firecrawl Credits">
          <ProgressBar
            value={data.firecrawl_credits_used}
            max={data.firecrawl_credits_used + data.firecrawl_credits_left}
            label={`${data.firecrawl_credits_left} remaining`}
            color={data.firecrawl_credits_left < 100 ? 'bg-red-500' : 'bg-[#FF6B85]'}
          />
        </ChartCard>
      </div>

      {data.recipes_created_week && data.recipes_created_week.length > 0 && (
        <ChartCard title="Recipes Created (7 Days)">
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={data.recipes_created_week}>
              <XAxis dataKey="date" tickFormatter={shortDate} stroke="#94a3b8" fontSize={12} />
              <YAxis stroke="#94a3b8" fontSize={12} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }}
                labelFormatter={shortDate}
              />
              <Bar dataKey="count" fill="#FF6B85" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      )}

      {healthChecks && (
        <ChartCard title="Health Checks">
          <div className="space-y-2">
            {healthChecks.map((hc) => (
              <div key={hc.name} className="flex items-center justify-between text-sm">
                <span className="text-[#F0F0F5]/80">{hc.name}</span>
                <div className="flex items-center gap-2">
                  {hc.count > 0 && <span className="text-[#F0F0F5]/50">{hc.count}</span>}
                  <StatusBadge status={hc.status} />
                </div>
              </div>
            ))}
          </div>
        </ChartCard>
      )}
    </div>
  )
}
