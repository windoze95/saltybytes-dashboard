import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../lib/api'
import AnimatedLogo from '../components/AnimatedLogo'

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
      <div className="bg-[#1E1E28] p-8 rounded-lg border border-[#3A3A48] w-full max-w-sm">
        <div className="mb-6">
          <AnimatedLogo fontSize={22} />
          <p className="text-xs text-[#F0F0F5]/50 mt-1">Dashboard</p>
        </div>
        <form onSubmit={handleSubmit}>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Password"
            className="w-full px-3 py-2 bg-[#2A2A36] border border-[#3A3A48] rounded text-[#F0F0F5] mb-4 focus:outline-none focus:border-[#FF6B85]"
            autoFocus
          />
          {error && <p className="text-red-400 text-sm mb-3">{error}</p>}
          <button
            type="submit"
            className="w-full py-2 bg-[#FF6B85] hover:bg-[#E55570] text-[#F0F0F5] rounded font-medium"
          >
            Sign In
          </button>
        </form>
      </div>
    </div>
  )
}
