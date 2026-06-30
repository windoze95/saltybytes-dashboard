import { api } from '../lib/api'
import { useMetrics } from '../hooks/useMetrics'
import { formatNumber, formatPercent, shortDate } from '../lib/format'
import StatCard from '../components/StatCard'
import ChartCard from '../components/ChartCard'
import ProgressBar from '../components/ProgressBar'
import DataTable from '../components/DataTable'
import Loading from '../components/Loading'
import {
  BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer,
  AreaChart, Area,
} from 'recharts'

export default function GrowthPage() {
  const { data, loading } = useMetrics(api.growth)

  if (loading || !data) return <Loading />

  const totalUsers = data.total_users
  const dailySignups = data.daily_signups ?? []
  const cumulativeUsers = data.cumulative_users ?? []
  const dailyRecipes = data.daily_recipes ?? []
  const buckets = data.recipes_per_user ?? []
  const tierMix = data.tier_distribution ?? []

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Growth & Engagement</h1>

      {/* Headline */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard title="Active Users" value={formatNumber(data.total_users)} />
        <StatCard title="New Today" value={formatNumber(data.new_users_today)} />
        <StatCard title="New (7d)" value={formatNumber(data.new_users_7d)} />
        <StatCard title="New (30d)" value={formatNumber(data.new_users_30d)} />
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard title="Total Recipes" value={formatNumber(data.total_recipes)} />
        <StatCard
          title="Recipes / Active User"
          value={data.recipes_per_active_user.toFixed(2)}
          subtitle="Total recipes / active users"
        />
        <StatCard
          title="Activated Users"
          value={formatNumber(data.users_with_recipes)}
          subtitle="Created >= 1 recipe"
        />
        <StatCard
          title="Activation Rate"
          value={formatPercent(data.activation_rate)}
          color="text-emerald-400"
        />
      </div>

      {/* Activation funnel: signup -> first recipe */}
      <ChartCard title="Activation Funnel — Signup to First Recipe">
        <ProgressBar
          value={data.users_with_recipes}
          max={totalUsers}
          label={`${data.users_with_recipes} / ${totalUsers} users created a recipe`}
          color="bg-emerald-500"
        />
      </ChartCard>

      {/* Signup trend + cumulative growth */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {dailySignups.length > 0 && (
          <ChartCard title="Daily Signups (30 Days)">
            <ResponsiveContainer width="100%" height={220}>
              <BarChart data={dailySignups}>
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

        {cumulativeUsers.length > 0 && (
          <ChartCard title="Cumulative Users (30 Days)">
            <ResponsiveContainer width="100%" height={220}>
              <AreaChart data={cumulativeUsers}>
                <defs>
                  <linearGradient id="growthCumulativeGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#5CFFD4" stopOpacity={0.4} />
                    <stop offset="95%" stopColor="#5CFFD4" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <XAxis dataKey="date" tickFormatter={shortDate} stroke="#94a3b8" fontSize={12} />
                <YAxis stroke="#94a3b8" fontSize={12} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }}
                  labelFormatter={shortDate}
                />
                <Area
                  type="monotone"
                  dataKey="count"
                  stroke="#5CFFD4"
                  strokeWidth={2}
                  fill="url(#growthCumulativeGradient)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </ChartCard>
        )}
      </div>

      {dailyRecipes.length > 0 && (
        <ChartCard title="Daily Recipes Created (30 Days)">
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={dailyRecipes}>
              <XAxis dataKey="date" tickFormatter={shortDate} stroke="#94a3b8" fontSize={12} />
              <YAxis stroke="#94a3b8" fontSize={12} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }}
                labelFormatter={shortDate}
              />
              <Bar dataKey="count" fill="#B4A7FF" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      )}

      {/* Engagement distribution: recipes-per-user buckets */}
      {buckets.length > 0 && (
        <ChartCard title="Engagement — Recipes per User">
          <DataTable
            columns={[
              { key: 'label', label: 'Recipes' },
              {
                key: 'count',
                label: 'Users',
                render: (r: Record<string, unknown>) => formatNumber(Number(r.count)),
              },
              {
                key: 'share',
                label: 'Share of Users',
                render: (r: Record<string, unknown>) =>
                  formatPercent(totalUsers > 0 ? (Number(r.count) / totalUsers) * 100 : 0),
              },
            ]}
            data={buckets as unknown as Record<string, unknown>[]}
          />
        </ChartCard>
      )}

      {/* Subscription tier mix */}
      {tierMix.length > 0 && (
        <ChartCard title="Subscription Tier Mix">
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={tierMix} layout="vertical">
              <XAxis type="number" stroke="#94a3b8" fontSize={12} />
              <YAxis dataKey="label" type="category" stroke="#94a3b8" fontSize={12} width={100} />
              <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} />
              <Bar dataKey="count" fill="#FF6B85" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      )}
    </div>
  )
}
