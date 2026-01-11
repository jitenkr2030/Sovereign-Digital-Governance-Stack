import { useState, useEffect } from 'react'
import { SearchIcon, BellIcon, SettingsIcon, HsmStatusIcon } from './Icons'
import { useQuery } from '@tanstack/react-query'

interface HeaderProps {
  onSearch?: (query: string) => void
}

export default function Header({ onSearch }: HeaderProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [currentTime, setCurrentTime] = useState(new Date())

  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000)
    return () => clearInterval(timer)
  }, [])

  // Fetch system status
  const { data: systemStatus } = useQuery({
    queryKey: ['systemStatus'],
    queryFn: async () => {
      // In production, this would call the actual API
      return {
        hsmConnected: true,
        databaseConnected: true,
        servicesHealthy: 4,
        totalServices: 4,
      }
    },
  })

  // Fetch alert count
  const { data: alertCount } = useQuery({
    queryKey: ['alertCount'],
    queryFn: async () => 3, // Mock data
  })

  return (
    <header className="bg-slate-800 border-b border-slate-700 px-6 py-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="relative">
            <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <input
              type="text"
              placeholder="Search entities, transactions..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && onSearch?.(searchQuery)}
              className="search-input pl-10 w-80"
            />
          </div>
        </div>

        <div className="flex items-center gap-6">
          {/* System Status Indicators */}
          <div className="flex items-center gap-4 text-sm">
            <div className="flex items-center gap-2">
              <div className={`w-2 h-2 rounded-full ${systemStatus?.hsmConnected ? 'bg-emerald-500' : 'bg-red-500'}`} />
              <span className="text-slate-400">HSM</span>
            </div>
            <div className="flex items-center gap-2">
              <div className={`w-2 h-2 rounded-full ${systemStatus?.databaseConnected ? 'bg-emerald-500' : 'bg-red-500'}`} />
              <span className="text-slate-400">DB</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-slate-400">Services:</span>
              <span className="text-emerald-400 font-medium">
                {systemStatus?.servicesHealthy}/{systemStatus?.totalServices}
              </span>
            </div>
          </div>

          {/* Notifications */}
          <button className="relative p-2 text-slate-400 hover:text-white transition-colors">
            <BellIcon className="w-5 h-5" />
            {alertCount && alertCount > 0 && (
              <span className="absolute -top-1 -right-1 w-5 h-5 bg-red-500 rounded-full text-xs font-bold text-white flex items-center justify-center">
                {alertCount}
              </span>
            )}
          </button>

          {/* Settings */}
          <button className="p-2 text-slate-400 hover:text-white transition-colors">
            <SettingsIcon className="w-5 h-5" />
          </button>

          {/* Timestamp */}
          <div className="text-sm text-slate-400">
            {currentTime.toLocaleString()}
          </div>
        </div>
      </div>
    </header>
  )
}
