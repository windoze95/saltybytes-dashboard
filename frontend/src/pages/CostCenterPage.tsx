import { useState } from 'react'
import { api, type RateCard } from '../lib/api'
import { useMetrics } from '../hooks/useMetrics'
import { formatDollars } from '../lib/format'
import StatCard from '../components/StatCard'
import ChartCard from '../components/ChartCard'
import ProgressBar from '../components/ProgressBar'
import Loading from '../components/Loading'

export default function CostCenterPage() {
  const { data, loading, refresh } = useMetrics(api.costCenter)
  const [editing, setEditing] = useState(false)
  const [rateCard, setRateCard] = useState<RateCard | null>(null)
  const [saving, setSaving] = useState(false)

  if (loading || !data) return <Loading />

  const handleEdit = () => {
    setRateCard({ ...data.rate_card })
    setEditing(true)
  }

  const handleSave = async () => {
    if (!rateCard) return
    setSaving(true)
    try {
      await api.updateRateCard(rateCard)
      setEditing(false)
      refresh()
    } finally {
      setSaving(false)
    }
  }

  const rateCardFields: { key: keyof RateCard; label: string; prefix?: string }[] = [
    { key: 'anthropic_sonnet_input_per_mtok', label: 'Sonnet Input/MTok', prefix: '$' },
    { key: 'anthropic_sonnet_output_per_mtok', label: 'Sonnet Output/MTok', prefix: '$' },
    { key: 'anthropic_haiku_input_per_mtok', label: 'Haiku Input/MTok', prefix: '$' },
    { key: 'anthropic_haiku_output_per_mtok', label: 'Haiku Output/MTok', prefix: '$' },
    { key: 'openai_dalle_per_image', label: 'DALL-E Per Image', prefix: '$' },
    { key: 'openai_whisper_per_minute', label: 'Whisper Per Minute', prefix: '$' },
    { key: 'openai_embedding_per_mtok', label: 'Embedding/MTok', prefix: '$' },
    { key: 'brave_monthly_plan', label: 'Brave Monthly', prefix: '$' },
    { key: 'firecrawl_monthly_credits', label: 'Firecrawl Credits/Mo' },
    { key: 'firecrawl_credits_per_scrape', label: 'Firecrawl Credits/Scrape' },
    { key: 'aws_rds_monthly', label: 'AWS RDS Monthly', prefix: '$' },
    { key: 'aws_ecs_monthly', label: 'AWS ECS Monthly', prefix: '$' },
    { key: 'aws_s3_per_gb', label: 'S3 Per GB', prefix: '$' },
  ]

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Cost Center</h1>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard
          title="Cost Per User"
          value={formatDollars(data.cost_per_user)}
          subtitle="Monthly fixed / total users"
          color="text-green-400"
        />
        <StatCard
          title="Cost Per Recipe"
          value={formatDollars(data.cost_per_recipe)}
          subtitle="Monthly fixed / total recipes"
          color="text-green-400"
        />
        <StatCard
          title="Monthly Fixed Costs"
          value={formatDollars(data.monthly_fixed_costs)}
          color="text-yellow-400"
        />
        <StatCard
          title="Total Savings"
          value={formatDollars(data.total_savings)}
          subtitle="Cache + canonical"
          color="text-emerald-400"
        />
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <StatCard
          title="Search Cache Savings"
          value={formatDollars(data.search_cache_savings)}
          color="text-emerald-400"
        />
        <StatCard
          title="Canonical Cache Savings"
          value={formatDollars(data.canonical_cache_savings)}
          color="text-emerald-400"
        />
        <ChartCard title="Firecrawl Credits">
          <ProgressBar
            value={data.firecrawl_credits_used}
            max={data.firecrawl_credits_max}
            label={`${data.firecrawl_credits_max - Number(data.firecrawl_credits_used)} remaining`}
            color={
              data.firecrawl_credits_used > data.firecrawl_credits_max * 0.8
                ? 'bg-red-500'
                : 'bg-[#FF6B85]'
            }
          />
        </ChartCard>
      </div>

      <ChartCard title="Provider Rate Card">
        <div className="flex justify-end mb-4">
          {!editing ? (
            <button
              onClick={handleEdit}
              className="px-3 py-1 bg-[#FF6B85] hover:bg-[#E55570] text-[#F0F0F5] rounded text-sm"
            >
              Edit
            </button>
          ) : (
            <div className="flex gap-2">
              <button
                onClick={() => setEditing(false)}
                className="px-3 py-1 bg-[#3A3A48] hover:bg-[#2A2A36] text-[#F0F0F5] rounded text-sm"
              >
                Cancel
              </button>
              <button
                onClick={handleSave}
                disabled={saving}
                className="px-3 py-1 bg-green-600 hover:bg-green-700 text-[#F0F0F5] rounded text-sm disabled:opacity-50"
              >
                {saving ? 'Saving...' : 'Save'}
              </button>
            </div>
          )}
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {rateCardFields.map(({ key, label, prefix }) => (
            <div key={key} className="flex items-center justify-between">
              <span className="text-sm text-[#F0F0F5]/60">{label}</span>
              {editing && rateCard ? (
                <input
                  type="number"
                  step="any"
                  value={rateCard[key]}
                  onChange={(e) =>
                    setRateCard({ ...rateCard, [key]: parseFloat(e.target.value) || 0 })
                  }
                  className="w-28 px-2 py-1 bg-[#2A2A36] border border-[#3A3A48] rounded text-[#F0F0F5] text-sm text-right"
                />
              ) : (
                <span className="text-sm text-[#F0F0F5]/80">
                  {prefix}
                  {data.rate_card[key]}
                </span>
              )}
            </div>
          ))}
        </div>
      </ChartCard>
    </div>
  )
}
