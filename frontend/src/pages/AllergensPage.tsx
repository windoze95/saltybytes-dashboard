import { api } from '../lib/api'
import { useMetrics } from '../hooks/useMetrics'
import { formatNumber, formatPercent, shortDate } from '../lib/format'
import StatCard from '../components/StatCard'
import ChartCard from '../components/ChartCard'
import Loading from '../components/Loading'
import {
  BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer,
  LineChart, Line,
  PieChart, Pie, Cell, Legend,
} from 'recharts'

const COLORS = ['#FF6B85', '#B4A7FF', '#5CFFD4', '#FFDAE0', '#E8DEFF', '#B2F5EA', '#06b6d4']

export default function AllergensPage() {
  const { data: allergens, loading: al } = useMetrics(api.allergens)
  const { data: families } = useMetrics(api.families)

  if (al || !allergens) return <Loading />

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Allergens & Families</h1>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard title="Total Analyses" value={formatNumber(allergens.total_analyses)} />
        <StatCard title="Today" value={allergens.analyses_today} />
        <StatCard title="Avg Confidence" value={formatPercent(allergens.avg_confidence * 100)} />
        <StatCard title="Requires Review" value={formatPercent(allergens.requires_review_rate)} subtitle={`${allergens.requires_review_count} total`} />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <ChartCard title="Premium vs Free">
          <ResponsiveContainer width="100%" height={200}>
            <PieChart>
              <Pie data={[
                { name: 'Premium', value: allergens.premium_count },
                { name: 'Free', value: allergens.free_count },
              ]} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={70} label>
                <Cell fill="#FF6B85" />
                <Cell fill="#94a3b8" />
              </Pie>
              <Legend />
              <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} />
            </PieChart>
          </ResponsiveContainer>
        </ChartCard>

        {allergens.allergen_flags.length > 0 && (
          <ChartCard title="Most Flagged Allergens">
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={allergens.allergen_flags} layout="vertical">
                <XAxis type="number" stroke="#94a3b8" fontSize={12} />
                <YAxis dataKey="label" type="category" stroke="#94a3b8" fontSize={12} width={80} />
                <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} />
                <Bar dataKey="count" fill="#FF6B85" radius={[0, 4, 4, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </ChartCard>
        )}
      </div>

      {allergens.daily_volume.length > 0 && (
        <ChartCard title="Daily Analysis Volume">
          <ResponsiveContainer width="100%" height={200}>
            <LineChart data={allergens.daily_volume}>
              <XAxis dataKey="date" tickFormatter={shortDate} stroke="#94a3b8" fontSize={12} />
              <YAxis stroke="#94a3b8" fontSize={12} />
              <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} labelFormatter={shortDate} />
              <Line type="monotone" dataKey="count" stroke="#FF6B85" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </ChartCard>
      )}

      {allergens.confidence_distribution.length > 0 && (
        <ChartCard title="Confidence Distribution">
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={allergens.confidence_distribution}>
              <XAxis dataKey="label" stroke="#94a3b8" fontSize={11} />
              <YAxis stroke="#94a3b8" fontSize={12} />
              <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} />
              <Bar dataKey="count" fill="#B4A7FF" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      )}

      {families && (
        <>
          <h2 className="text-xl font-bold mt-8">Families</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard title="Total Families" value={families.total_families} />
            <StatCard title="Total Members" value={families.total_members} />
            <StatCard title="Avg Members/Family" value={families.avg_members_per_family.toFixed(1)} />
            <StatCard title="Dietary Profile Coverage" value={formatPercent(families.dietary_profile_coverage)} subtitle={`${families.members_with_dietary_profile} members`} />
          </div>
        </>
      )}
    </div>
  )
}
