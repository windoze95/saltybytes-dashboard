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
  PieChart, Pie, Cell, Legend,
} from 'recharts'

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899', '#06b6d4']

export default function RecipesPage() {
  const { data: recipes, loading: rl } = useMetrics(api.recipes)
  const { data: canonical } = useMetrics(api.canonical)

  if (rl || !recipes) return <Loading />

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Recipes</h1>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard title="Total Recipes" value={formatNumber(recipes.total_recipes)} />
        <StatCard title="Created Today" value={recipes.recipes_today} />
        <StatCard title="Created This Week" value={recipes.recipes_this_week} />
        <StatCard title="Deleted Today" value={recipes.deleted_today} />
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard title="Avg Ingredients" value={recipes.avg_ingredients_per_recipe.toFixed(1)} />
        <StatCard title="Avg Cook Time" value={`${recipes.avg_cook_time.toFixed(0)} min`} />
        <StatCard title="Avg Portions" value={recipes.avg_portions.toFixed(1)} />
        <StatCard title="User Edited" value={formatPercent(recipes.user_edited_rate)} subtitle={`${recipes.user_edited_count} recipes`} />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <ChartCard title="Embedding Coverage">
          <ProgressBar value={recipes.embedding_coverage} max={100} label={`${recipes.recipes_with_embeddings} / ${recipes.total_recipes}`} color="bg-emerald-500" />
        </ChartCard>
        <ChartCard title="Image Coverage">
          <ProgressBar value={recipes.image_coverage} max={100} label={`${recipes.recipes_with_images} / ${recipes.total_recipes}`} color="bg-blue-500" />
        </ChartCard>
      </div>

      {recipes.node_type_distribution.length > 0 && (
        <ChartCard title="Node Type Distribution">
          <ResponsiveContainer width="100%" height={250}>
            <PieChart>
              <Pie data={recipes.node_type_distribution} dataKey="count" nameKey="label" cx="50%" cy="50%" outerRadius={80} label>
                {recipes.node_type_distribution.map((_, i) => (
                  <Cell key={i} fill={COLORS[i % COLORS.length]} />
                ))}
              </Pie>
              <Legend />
              <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155' }} />
            </PieChart>
          </ResponsiveContainer>
        </ChartCard>
      )}

      {recipes.import_breakdown.length > 0 && (
        <ChartCard title="Import Breakdown">
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={recipes.import_breakdown} layout="vertical">
              <XAxis type="number" stroke="#94a3b8" fontSize={12} />
              <YAxis dataKey="label" type="category" stroke="#94a3b8" fontSize={12} width={100} />
              <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155' }} />
              <Bar dataKey="count" fill="#3b82f6" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      )}

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard title="Canonical Linked" value={recipes.canonical_linked} />
        <StatCard title="Non-Diverged" value={recipes.canonical_non_diverged} subtitle="Thin refs" />
        <StatCard title="Diverged" value={recipes.canonical_diverged} subtitle="Materialized" />
        <StatCard title="Divergence Rate" value={formatPercent(recipes.divergence_rate)} />
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard title="Total Trees" value={recipes.total_trees} />
        <StatCard title="Fork Count" value={recipes.fork_count} />
        <StatCard title="Max Fork Depth" value={recipes.max_fork_depth} />
        <StatCard title="Ephemeral Nodes" value={recipes.ephemeral_nodes} />
        <StatCard title="Image Regens" value={recipes.image_regen_count} />
        <StatCard title="With Source URL" value={recipes.recipes_with_source_url} />
      </div>

      {canonical && (
        <>
          <h2 className="text-xl font-bold mt-8">Canonical Cache</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard title="Total Entries" value={canonical.total_entries} />
            <StatCard title="New Today" value={canonical.new_today} />
            <StatCard title="Avg Hit Count" value={canonical.avg_hit_count.toFixed(1)} />
            <StatCard title="Zero-Hit Entries" value={canonical.zero_hit_entries} />
          </div>

          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            <StatCard title="Approaching TTL" value={canonical.entries_approaching_ttl} />
            <StatCard title="Hot (Eligible Refresh)" value={canonical.hot_entries_eligible} />
            <StatCard title="Stale (90d)" value={canonical.stale_entries_90d} />
          </div>

          {canonical.extraction_method_distribution.length > 0 && (
            <ChartCard title="Extraction Method Distribution">
              <ResponsiveContainer width="100%" height={250}>
                <PieChart>
                  <Pie data={canonical.extraction_method_distribution} dataKey="count" nameKey="label" cx="50%" cy="50%" outerRadius={80} label>
                    {canonical.extraction_method_distribution.map((_, i) => (
                      <Cell key={i} fill={COLORS[i % COLORS.length]} />
                    ))}
                  </Pie>
                  <Legend />
                  <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155' }} />
                </PieChart>
              </ResponsiveContainer>
            </ChartCard>
          )}

          <StatCard title="AI Extractions Saved" value={canonical.total_ai_extractions_saved} subtitle="Avoided Haiku calls" color="text-emerald-400" />

          {canonical.top_by_hits.length > 0 && (
            <ChartCard title="Top Canonical Entries by Hits">
              <DataTable
                columns={[
                  { key: 'label', label: 'URL', render: (r: Record<string, unknown>) => {
                    const url = String(r.label || '')
                    return <span className="text-xs break-all">{url.length > 60 ? url.slice(0, 60) + '...' : url}</span>
                  }},
                  { key: 'count', label: 'Hits' },
                  { key: 'extra', label: 'Method' },
                ]}
                data={canonical.top_by_hits as unknown as Record<string, unknown>[]}
                maxRows={15}
              />
            </ChartCard>
          )}

          {canonical.daily_new.length > 0 && (
            <ChartCard title="Daily New Canonical Entries">
              <ResponsiveContainer width="100%" height={200}>
                <BarChart data={canonical.daily_new}>
                  <XAxis dataKey="date" tickFormatter={shortDate} stroke="#94a3b8" fontSize={12} />
                  <YAxis stroke="#94a3b8" fontSize={12} />
                  <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155' }} labelFormatter={shortDate} />
                  <Bar dataKey="count" fill="#10b981" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </ChartCard>
          )}
        </>
      )}

      {recipes.top_hashtags.length > 0 && (
        <ChartCard title="Top Hashtags">
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={recipes.top_hashtags.slice(0, 15)} layout="vertical">
              <XAxis type="number" stroke="#94a3b8" fontSize={12} />
              <YAxis dataKey="label" type="category" stroke="#94a3b8" fontSize={12} width={120} />
              <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155' }} />
              <Bar dataKey="count" fill="#8b5cf6" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      )}

      {recipes.recipes_per_user.length > 0 && (
        <ChartCard title="Recipes Per User Distribution">
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={recipes.recipes_per_user}>
              <XAxis dataKey="label" stroke="#94a3b8" fontSize={12} />
              <YAxis stroke="#94a3b8" fontSize={12} />
              <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155' }} />
              <Bar dataKey="count" fill="#f59e0b" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      )}
    </div>
  )
}
