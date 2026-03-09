interface Column<T> {
  key: string
  label: string
  render?: (row: T) => React.ReactNode
}

interface Props<T> {
  columns: Column<T>[]
  data: T[]
  maxRows?: number
}

export default function DataTable<T extends Record<string, unknown>>({
  columns,
  data,
  maxRows,
}: Props<T>) {
  const rows = maxRows ? data.slice(0, maxRows) : data
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-[#3A3A48]">
            {columns.map((col) => (
              <th key={col.key} className="text-left py-2 px-3 text-[#F0F0F5]/60 font-medium">
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, i) => (
            <tr key={i} className="border-b border-[#3A3A48]/50 hover:bg-[#2A2A36]/30">
              {columns.map((col) => (
                <td key={col.key} className="py-2 px-3 text-[#F0F0F5]/80">
                  {col.render ? col.render(row) : String(row[col.key] ?? '')}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
