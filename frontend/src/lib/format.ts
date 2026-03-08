export function formatNumber(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K'
  return n.toLocaleString()
}

export function formatBytes(bytes: number): string {
  if (bytes >= 1024 * 1024 * 1024) return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
  if (bytes >= 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
  if (bytes >= 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return bytes + ' B'
}

export function formatDollars(n: number): string {
  return '$' + n.toFixed(2)
}

export function formatPercent(n: number): string {
  return n.toFixed(1) + '%'
}

export function formatDate(d: string): string {
  return new Date(d).toLocaleDateString()
}

export function shortDate(d: unknown): string {
  const date = new Date(String(d))
  return `${date.getMonth() + 1}/${date.getDate()}`
}

export function tooltipBytes(value: unknown): [string, string] {
  return [formatBytes(Number(value)), 'Size']
}
