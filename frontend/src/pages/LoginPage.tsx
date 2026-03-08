import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../lib/api'

export default function LoginPage() {
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await api.login(password)
      navigate('/')
    } catch {
      setError('Invalid password')
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="bg-slate-800 p-8 rounded-lg border border-slate-700 w-full max-w-sm">
        <h1 className="text-xl font-bold text-white mb-6">SaltyBytes Dashboard</h1>
        <form onSubmit={handleSubmit}>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Password"
            className="w-full px-3 py-2 bg-slate-700 border border-slate-600 rounded text-white mb-4 focus:outline-none focus:border-blue-500"
            autoFocus
          />
          {error && <p className="text-red-400 text-sm mb-3">{error}</p>}
          <button
            type="submit"
            className="w-full py-2 bg-blue-600 hover:bg-blue-700 text-white rounded font-medium"
          >
            Sign In
          </button>
        </form>
      </div>
    </div>
  )
}
