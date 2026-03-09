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
  LineChart, Line,
  PieChart, Pie, Cell, Legend,
} from 'recharts'

const COLORS = ['#FF6B85', '#B4A7FF', '#5CFFD4', '#FFDAE0', '#E8DEFF', '#B2F5EA']

export default function UsersPage() {
  const { data: users, loading: ul } = useMetrics(api.users)
  const { data: subs } = useMetrics(api.subscriptions)

  if (ul || !users) return <Loading />

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Users & Subscriptions</h1>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard title="Total Users" value={formatNumber(users.total_users)} />
        <StatCard title="New Today" value={users.new_users_today} />
        <StatCard title="New This Week" value={users.new_users_this_week} />
        <StatCard title="New This Month" value={users.new_users_this_month} />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <ChartCard title="Personalization Adoption">
          <ProgressBar
            value={users.users_with_personalization}
            max={users.total_users}
            label={`${users.users_with_personalization} / ${users.total_users} users`}
            color="bg-emerald-500"
          />
        </ChartCard>
        <ChartCard title="Email Set">
          <ProgressBar
            value={users.users_with_email}
            max={users.total_users}
            label={`${users.users_with_email} / ${users.total_users} users`}
            color="bg-[#FF6B85]"
          />
        </ChartCard>
      </div>

      {users.daily_registrations.length > 0 && (
        <ChartCard title="Daily Registrations (30 Days)">
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={users.daily_registrations}>
              <XAxis dataKey="date" tickFormatter={shortDate} stroke="#94a3b8" fontSize={12} />
              <YAxis stroke="#94a3b8" fontSize={12} />
              <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} labelFormatter={shortDate} />
              <Line type="monotone" dataKey="count" stroke="#FF6B85" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </ChartCard>
      )}

      {subs && (
        <>
          <h2 className="text-xl font-bold mt-8">Subscriptions</h2>

          {subs.tier_distribution.length > 0 && (
            <ChartCard title="Tier Distribution">
              <ResponsiveContainer width="100%" height={250}>
                <PieChart>
                  <Pie data={subs.tier_distribution} dataKey="count" nameKey="label" cx="50%" cy="50%" outerRadius={80} label>
                    {subs.tier_distribution.map((_, i) => (
                      <Cell key={i} fill={COLORS[i % COLORS.length]} />
                    ))}
                  </Pie>
                  <Legend />
                  <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} />
                </PieChart>
              </ResponsiveContainer>
            </ChartCard>
          )}

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <StatCard title="Avg Allergen Analyses Used" value={subs.avg_allergen_used.toFixed(1)} subtitle="Free tier" />
            <StatCard title="Avg Web Searches Used" value={subs.avg_searches_used.toFixed(1)} subtitle="Free tier" />
            <StatCard title="Avg AI Generations Used" value={subs.avg_ai_generations_used.toFixed(1)} subtitle="Free tier" />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <StatCard title="Near Allergen Limit" value={subs.users_near_allergen_limit} subtitle=">= 18/20" color="text-yellow-400" />
            <StatCard title="Near Search Limit" value={subs.users_near_search_limit} subtitle=">= 18/20" color="text-yellow-400" />
            <StatCard title="Near AI Gen Limit" value={subs.users_near_ai_gen_limit} subtitle=">= 18/20" color="text-yellow-400" />
          </div>

          {subs.allergen_distribution.length > 0 && (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <ChartCard title="Allergen Usage Distribution">
                <ResponsiveContainer width="100%" height={150}>
                  <BarChart data={subs.allergen_distribution}>
                    <XAxis dataKey="label" stroke="#94a3b8" fontSize={10} />
                    <YAxis stroke="#94a3b8" fontSize={10} />
                    <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} />
                    <Bar dataKey="count" fill="#FF6B85" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </ChartCard>
              <ChartCard title="Search Usage Distribution">
                <ResponsiveContainer width="100%" height={150}>
                  <BarChart data={subs.search_distribution}>
                    <XAxis dataKey="label" stroke="#94a3b8" fontSize={10} />
                    <YAxis stroke="#94a3b8" fontSize={10} />
                    <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} />
                    <Bar dataKey="count" fill="#B4A7FF" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </ChartCard>
              <ChartCard title="AI Gen Usage Distribution">
                <ResponsiveContainer width="100%" height={150}>
                  <BarChart data={subs.ai_gen_distribution}>
                    <XAxis dataKey="label" stroke="#94a3b8" fontSize={10} />
                    <YAxis stroke="#94a3b8" fontSize={10} />
                    <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} />
                    <Bar dataKey="count" fill="#5CFFD4" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </ChartCard>
            </div>
          )}

          {subs.users_at_limit.length > 0 && (
            <ChartCard title="Users Near Limits">
              <DataTable
                columns={[
                  { key: 'username', label: 'Username' },
                  { key: 'tier', label: 'Tier' },
                  { key: 'allergen_analyses_used', label: 'Allergen' },
                  { key: 'web_searches_used', label: 'Search' },
                  { key: 'ai_generations_used', label: 'AI Gen' },
                ]}
                data={subs.users_at_limit as unknown as Record<string, unknown>[]}
                maxRows={20}
              />
            </ChartCard>
          )}
        </>
      )}

      {users.unit_system_distribution.length > 0 && (
        <ChartCard title="Unit System Distribution">
          <ResponsiveContainer width="100%" height={200}>
            <PieChart>
              <Pie data={users.unit_system_distribution} dataKey="count" nameKey="label" cx="50%" cy="50%" outerRadius={70} label>
                {users.unit_system_distribution.map((_, i) => (
                  <Cell key={i} fill={COLORS[i % COLORS.length]} />
                ))}
              </Pie>
              <Legend />
              <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} />
            </PieChart>
          </ResponsiveContainer>
        </ChartCard>
      )}
    </div>
  )
}
