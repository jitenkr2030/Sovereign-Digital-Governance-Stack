interface Activity {
  id: string
  type: 'license' | 'compliance' | 'obligation' | 'alert'
  title: string
  description: string
  timestamp: string
  status: 'success' | 'warning' | 'error' | 'info'
}

const mockActivities: Activity[] = [
  {
    id: '1',
    type: 'license',
    title: 'License Issued',
    description: 'Crypto Exchange Pro received exchange license',
    timestamp: '2 minutes ago',
    status: 'success',
  },
  {
    id: '2',
    type: 'compliance',
    title: 'Compliance Score Updated',
    description: 'Mining Corp score improved to 87 (Silver)',
    timestamp: '15 minutes ago',
    status: 'success',
  },
  {
    id: '3',
    type: 'obligation',
    title: 'Obligation Overdue',
    description: 'Digital Assets Inc failed to submit Q4 audit report',
    timestamp: '1 hour ago',
    status: 'error',
  },
  {
    id: '4',
    type: 'alert',
    title: 'High-Risk Transaction',
    description: 'Suspicious transaction flagged for review',
    timestamp: '2 hours ago',
    status: 'warning',
  },
  {
    id: '5',
    type: 'compliance',
    title: 'Compliance Check Passed',
    description: 'Monthly KYC review completed for 45 entities',
    timestamp: '3 hours ago',
    status: 'success',
  },
]

const statusColors = {
  success: 'bg-emerald-500/20 text-emerald-400',
  warning: 'bg-amber-500/20 text-amber-400',
  error: 'bg-red-500/20 text-red-400',
  info: 'bg-blue-500/20 text-blue-400',
}

const typeIcons = {
  license: '🎫',
  compliance: '✅',
  obligation: '📋',
  alert: '⚠️',
}

export default function RecentActivity() {
  return (
    <div className="space-y-3">
      {mockActivities.map((activity) => (
        <div 
          key={activity.id}
          className="flex items-start gap-3 p-3 rounded-lg bg-slate-700/30 hover:bg-slate-700/50 transition-colors"
        >
          <div className="text-xl">{typeIcons[activity.type]}</div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center justify-between">
              <p className="text-sm font-medium text-white truncate">{activity.title}</p>
              <span className={`text-xs px-2 py-0.5 rounded-full ${statusColors[activity.status]}`}>
                {activity.status}
              </span>
            </div>
            <p className="text-xs text-slate-400 mt-1">{activity.description}</p>
            <p className="text-xs text-slate-500 mt-1">{activity.timestamp}</p>
          </div>
        </div>
      ))}
      <button className="w-full py-2 text-sm text-slate-400 hover:text-white transition-colors">
        View all activity →
      </button>
    </div>
  )
}
