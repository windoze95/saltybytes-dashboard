import { Routes, Route, NavLink, useNavigate } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { api } from './lib/api'
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
      <nav className="w-56 bg-slate-900 border-r border-slate-800 flex flex-col">
        <div className="p-4 border-b border-slate-800">
          <h1 className="text-lg font-bold text-white">SaltyBytes</h1>
          <p className="text-xs text-slate-500">Dashboard</p>
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
                    ? 'bg-slate-800 text-white border-l-2 border-blue-500'
                    : 'text-slate-400 hover:text-white hover:bg-slate-800/50'
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </div>
        <div className="p-3 border-t border-slate-800 space-y-2">
          <button
            onClick={handleRefresh}
            className="w-full px-3 py-1.5 text-xs bg-slate-800 hover:bg-slate-700 text-slate-400 rounded"
          >
            Refresh Data
          </button>
          <button
            onClick={handleLogout}
            className="w-full px-3 py-1.5 text-xs bg-slate-800 hover:bg-slate-700 text-slate-400 rounded"
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
