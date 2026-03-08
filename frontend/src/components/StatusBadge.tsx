interface Props {
  status: 'pass' | 'fail' | 'warn'
}

const colors = {
  pass: 'bg-green-900 text-green-300',
  fail: 'bg-red-900 text-red-300',
  warn: 'bg-yellow-900 text-yellow-300',
}

export default function StatusBadge({ status }: Props) {
  return (
    <span className={`px-2 py-0.5 rounded text-xs font-medium ${colors[status]}`}>
      {status.toUpperCase()}
    </span>
  )
}
