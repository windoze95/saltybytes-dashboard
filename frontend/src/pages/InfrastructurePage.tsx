import { api } from '../lib/api'
import { useMetrics } from '../hooks/useMetrics'
import { formatBytes, formatNumber, formatDollars, tooltipBytes } from '../lib/format'
import StatCard from '../components/StatCard'
import ChartCard from '../components/ChartCard'
import DataTable from '../components/DataTable'
import Loading from '../components/Loading'
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'

export default function InfrastructurePage() {
  const { data, loading } = useMetrics(api.infrastructure)

  if (loading || !data) return <Loading />

  const tableData = data.table_sizes
    .filter((t) => t.size_bytes > 0)
    .map((t) => ({
      ...t,
      size_display: formatBytes(t.size_bytes),
      rows_display: formatNumber(t.rows),
    }))

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Infrastructure</h1>

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard title="Database Size" value={`${data.database_size_mb.toFixed(1)} MB`} />
        <StatCard title="Connections" value={data.connection_count} />
        <StatCard title="S3 Images" value={formatNumber(data.s3_image_count)} />
        <StatCard title="Est. S3 Cost" value={formatDollars(data.s3_estimated_cost)} subtitle={`${data.s3_estimated_size_mb.toFixed(0)} MB`} />
      </div>

      {tableData.length > 0 && (
        <ChartCard title="Table Sizes">
          <ResponsiveContainer width="100%" height={Math.max(200, tableData.length * 28)}>
            <BarChart data={tableData} layout="vertical">
              <XAxis type="number" stroke="#94a3b8" fontSize={12} tickFormatter={(v) => formatBytes(v)} />
              <YAxis dataKey="name" type="category" stroke="#94a3b8" fontSize={11} width={160} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1E1E28', border: '1px solid #3A3A48' }}
                formatter={tooltipBytes}
              />
              <Bar dataKey="size_bytes" fill="#FF6B85" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </ChartCard>
      )}

      {data.table_sizes.length > 0 && (
        <ChartCard title="Table Row Counts">
          <DataTable
            columns={[
              { key: 'name', label: 'Table' },
              { key: 'rows', label: 'Rows', render: (r: Record<string, unknown>) => formatNumber(Number(r.rows)) },
              { key: 'size_bytes', label: 'Size', render: (r: Record<string, unknown>) => formatBytes(Number(r.size_bytes)) },
            ]}
            data={data.table_sizes.filter(t => t.rows > 0) as unknown as Record<string, unknown>[]}
          />
        </ChartCard>
      )}

      {data.index_sizes.length > 0 && (
        <ChartCard title="Index Sizes">
          <DataTable
            columns={[
              { key: 'name', label: 'Index' },
              { key: 'table', label: 'Table' },
              { key: 'size_bytes', label: 'Size', render: (r: Record<string, unknown>) => formatBytes(Number(r.size_bytes)) },
            ]}
            data={data.index_sizes.filter(i => i.size_bytes > 0) as unknown as Record<string, unknown>[]}
          />
        </ChartCard>
      )}
    </div>
  )
}
