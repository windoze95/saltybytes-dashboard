import { api } from '../lib/api'
import { useMetrics } from '../hooks/useMetrics'
import { formatNumber, formatPercent, shortDate } from '../lib/format'
import StatCard from '../components/StatCard'
import ChartCard from '../components/ChartCard'
import DataTable from '../components/DataTable'
import Loading from '../components/Loading'
import {
  BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell, Legend,
} from 'recharts'

const COLORS = ['#FF6B85', '#B4A7FF', '#5CFFD4', '#FFDAE0', '#E8DEFF', '#B2F5EA', '#06b6d4']

// Map known recipe statuses to semantic colors; fall back to the palette.
const STATUS_COLORS: Record<string, string> = {
  ready: '#34d399', // emerald-400
  failed: '#f87171', // red-400
  generating: '#fbbf24', // amber-400
}

function statusColor(label: string, i: number): string {
  return STATUS_COLORS[label] ?? COLORS[i % COLORS.length]
}

export default function RecipeQualityPage() {
  const { data, loading } = useMetrics(api.recipeQuality)

  if (loading && !data) return <Loading />

  // Endpoint returns null until the cache warms, or zero recipes on a fresh DB.
  if (!data || data.total_recipes === 0) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">Recipe Quality</h1>
        <div className="bg-[#1E1E28] rounded-lg p-12 border border-[#3A3A48] text-center">
          <p className="text-[#F0F0F5]/70">No recipes yet</p>
          <p className="text-sm text-[#F0F0F5]/40 mt-1">
            Extraction-quality and completeness signals will appear here once recipes are created.
          </p>
        </div>
      </div>
    )
  }

  const extractionDist = data.extraction_method_distribution ?? []
  const statusDist = data.status_distribution ?? []
  const typeDist = data.type_distribution ?? []
  const dailyOutcome = data.daily_outcome ?? []
  const totalCanonical = data.total_canonical
  const totalExtractions = data.free_extraction_count + data.paid_extraction_count

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Recipe Quality</h1>

      {/* Headline */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard title="Total Recipes" value={formatNumber(data.total_recipes)} />
        <StatCard
          title="Canonical Extractions"
          value={formatNumber(data.total_canonical)}
          subtitle="URL-keyed master copies"
        />
        <StatCard
          title="Free Extraction"
          value={formatPercent(data.free_extraction_rate)}
          subtitle={`${formatNumber(data.free_extraction_count)} json_ld`}
          color="text-emerald-400"
        />
        <StatCard
          title="Paid AI Extraction"
          value={formatPercent(data.paid_extraction_rate)}
          subtitle={`${formatNumber(data.paid_extraction_count)} haiku`}
          color="text-[#FF6B85]"
        />
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard
          title="Failure Rate"
          value={formatPercent(data.failure_rate)}
          subtitle={`${formatNumber(data.failed_recipes)} failed`}
          color={data.failed_recipes > 0 ? 'text-red-400' : 'text-[#F0F0F5]'}
        />
        <StatCard title="Avg Ingredients" value={data.avg_ingredients.toFixed(1)} />
        <StatCard title="Avg Steps" value={data.avg_steps.toFixed(1)} />
        <StatCard
          title="Missing Ingredients"
          value={formatNumber(data.recipes_missing_ingredients)}
          subtitle="Empty ingredient lists"
          color={data.recipes_missing_ingredients > 0 ? 'text-amber-400' : 'text-[#F0F0F5]'}
        />
      </div>

      {/* Free vs paid proportion bar — the cost story */}
      {totalExtractions > 0 && (
        <ChartCard title="Free vs Paid Extraction">
          <p className="text-xs text-[#F0F0F5]/50 mb-3 -mt-1">
            Share of canonical extractions served by free JSON-LD parsing vs paid AI (Haiku).
          </p>
          <div className="flex h-3 rounded-full overflow-hidden bg-[#2A2A36]">
            <div className="h-full bg-emerald-400" style={{ width: `${data.free_extraction_rate}%` }} />
            <div className="h-full bg-[#FF6B85]" style={{ width: `${data.paid_extraction_rate}%` }} />
          </div>
          <div className="flex items-center justify-between mt-2 text-sm">
            <span className="text-emerald-400">
              Free {formatPercent(data.free_extraction_rate)} ({formatNumber(data.free_extraction_count)})
            </span>
            <span className="text-[#FF6B85]">
              Paid {formatPercent(data.paid_extraction_rate)} ({formatNumber(data.paid_extraction_count)})
            </span>
          </div>
        </ChartCard>
      )}

      {/* Extraction method mix: chart + table */}
      {extractionDist.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <ChartCard title="Extraction Method Mix">
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie data={extractionDist} dataKey="count" nameKey="label" cx="50%" cy="50%" outerRadius={80} label>
                  {extractionDist.map((_, i) => (
                    <Cell key={i} fill={COLORS[i % COLORS.length]} />
                  ))}
                </Pie>
                <Legend />
                <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} />
              </PieChart>
            </ResponsiveContainer>
          </ChartCard>
          <ChartCard title="Extraction Methods">
            <DataTable
              columns={[
                { key: 'label', label: 'Method' },
                { key: 'count', label: 'Count', render: (r: Record<string, unknown>) => formatNumber(Number(r.count)) },
                {
                  key: 'share',
                  label: 'Share',
                  render: (r: Record<string, unknown>) =>
                    formatPercent(totalCanonical > 0 ? (Number(r.count) / totalCanonical) * 100 : 0),
                },
              ]}
              data={extractionDist as unknown as Record<string, unknown>[]}
            />
          </ChartCard>
        </div>
      )}

      {/* Recipe status distribution */}
      {statusDist.length > 0 && (
        <ChartCard title="Recipe Status Distribution">
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie data={statusDist} dataKey="count" nameKey="label" cx="50%" cy="50%" outerRadius={80} label>
                {statusDist.map((d, i) => (
                  <Cell key={i} fill={statusColor(d.label, i)} />
                ))}
              </Pie>
              <Legend />
              <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} />
            </PieChart>
          </ResponsiveContainer>
        </ChartCard>
      )}

      {/* Source/type mix — only rendered when recipes exposes a type column */}
      {typeDist.length > 0 && (
        <ChartCard title="Recipe Source / Type Mix">
          <ResponsiveContainer width="100%" height={Math.max(200, typeDist.length * 32)}>
            <BarChart data={typeDist} layout="vertical">
              <XAxis type="number" stroke="#94a3b8" fontSize={12} />
              <YAxis dataKey="label" type="category" stroke="#94a3b8" fontSize={12} width={120} />
              <Tooltip contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }} />
              <Bar dataKey="count" fill="#B4A7FF" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      )}

      {/* Creation outcome over time */}
      {dailyOutcome.length > 0 && (
        <ChartCard title="Creation Outcome (last 30 days)">
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={dailyOutcome}>
              <XAxis dataKey="date" tickFormatter={shortDate} stroke="#94a3b8" fontSize={12} />
              <YAxis stroke="#94a3b8" fontSize={12} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }}
                labelFormatter={shortDate}
              />
              <Legend />
              <Bar dataKey="succeeded" stackId="a" fill="#34d399" />
              <Bar dataKey="failed" stackId="a" fill="#f87171" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      )}
    </div>
  )
}
