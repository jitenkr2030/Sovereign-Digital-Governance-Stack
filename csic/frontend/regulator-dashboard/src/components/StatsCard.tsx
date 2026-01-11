import { ArrowUpIcon, ArrowDownIcon } from './Icons'

interface StatsCardProps {
  title: string
  value: string
  change: string
  trend: 'up' | 'down'
  icon: 'users' | 'activity' | 'clock' | 'alert'
  alert?: boolean
}

const iconPaths = {
  users: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z',
  activity: 'M13 10V3L4 14h7v7l9-11h-7z',
  clock: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
  alert: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
}

export default function StatsCard({ title, value, change, trend, icon, alert }: StatsCardProps) {
  const isPositive = trend === 'up'
  
  return (
    <div className={`dashboard-card ${alert ? 'border-red-500/50' : ''}`}>
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm text-slate-400">{title}</p>
          <p className={`text-3xl font-bold mt-2 ${alert ? 'text-red-400' : 'text-white'}`}>
            {value}
          </p>
          <div className="flex items-center gap-1 mt-2">
            {isPositive ? (
              <ArrowUpIcon className="w-4 h-4 text-emerald-400" />
            ) : (
              <ArrowDownIcon className="w-4 h-4 text-red-400" />
            )}
            <span className={`text-sm ${isPositive ? 'text-emerald-400' : 'text-red-400'}`}>
              {change}
            </span>
            <span className="text-sm text-slate-500">vs last week</span>
          </div>
        </div>
        <div className={`p-3 rounded-lg ${
          alert ? 'bg-red-500/20' : 'bg-slate-700/50'
        }`}>
          <svg className={`w-6 h-6 ${alert ? 'text-red-400' : 'text-slate-400'}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={iconPaths[icon]} />
          </svg>
        </div>
      </div>
    </div>
  )
}
