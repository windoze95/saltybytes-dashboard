import { Routes, Route, NavLink, useNavigate } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { api } from './lib/api'
import AnimatedLogo from './components/AnimatedLogo'
import LoginPage from './pages/LoginPage'
import OverviewPage from './pages/OverviewPage'
import CostCenterPage from './pages/CostCenterPage'
import RecipesPage from './pages/RecipesPage'
import SearchCachePage from './pages/SearchCachePage'
import UsersPage from './pages/UsersPage'
import AllergensPage from './pages/AllergensPage'
import InfrastructurePage from './pages/InfrastructurePage'
import DataQualityPage from './pages/DataQualityPage'

const NAV_ITEMS = [
  { path: '/', label: 'Overview' },
  { path: '/cost', label: 'Cost Center' },
  { path: '/recipes', label: 'Recipes' },
  { path: '/search', label: 'Search & Cache' },
  { path: '/users', label: 'Users & Subs' },
  { path: '/allergens', label: 'Allergens' },
  { path: '/infrastructure', label: 'Infrastructure' },
  { path: '/quality', label: 'Data Quality' },
]

function Layout() {
  const navigate = useNavigate()

  const handleLogout = async () => {
    await api.logout()
    navigate('/login')
  }

  const handleRefresh = () => {
    api.refresh()
  }

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <nav className="w-56 bg-[#121218] border-r border-[#3A3A48] flex flex-col">
        <div className="p-4 border-b border-[#3A3A48]">
          <AnimatedLogo fontSize={16} />
          <p className="text-xs text-[#F0F0F5]/50 mt-0.5">Dashboard</p>
        </div>
        <div className="flex-1 py-2">
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.path === '/'}
              className={({ isActive }) =>
                `block px-4 py-2 text-sm ${
                  isActive
                    ? 'bg-[#1E1E28] text-[#F0F0F5] border-l-2 border-[#FF6B85]'
                    : 'text-[#F0F0F5]/60 hover:text-[#F0F0F5] hover:bg-[#1E1E28]/50'
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </div>
        <div className="p-3 border-t border-[#3A3A48] space-y-2">
          <button
            onClick={handleRefresh}
            className="w-full px-3 py-1.5 text-xs bg-[#1E1E28] hover:bg-[#2A2A36] text-[#F0F0F5]/60 rounded"
          >
            Refresh Data
          </button>
          <button
            onClick={handleLogout}
            className="w-full px-3 py-1.5 text-xs bg-[#1E1E28] hover:bg-[#2A2A36] text-[#F0F0F5]/60 rounded"
          >
            Sign Out
          </button>
        </div>
      </nav>

      {/* Main content */}
      <main className="flex-1 p-6 overflow-auto">
        <Routes>
          <Route path="/" element={<OverviewPage />} />
          <Route path="/cost" element={<CostCenterPage />} />
          <Route path="/recipes" element={<RecipesPage />} />
          <Route path="/search" element={<SearchCachePage />} />
          <Route path="/users" element={<UsersPage />} />
          <Route path="/allergens" element={<AllergensPage />} />
          <Route path="/infrastructure" element={<InfrastructurePage />} />
          <Route path="/quality" element={<DataQualityPage />} />
        </Routes>
      </main>
    </div>
  )
}

export default function App() {
  const [authed, setAuthed] = useState<boolean | null>(null)

  useEffect(() => {
    api
      .checkAuth()
      .then((r) => setAuthed(r.authenticated))
      .catch(() => setAuthed(false))
  }, [])

  if (authed === null) return null

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/*" element={authed ? <Layout /> : <LoginPage />} />
    </Routes>
  )
}
