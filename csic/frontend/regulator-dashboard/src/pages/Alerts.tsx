export default function Alerts() {
  const mockAlerts = [
    { id: '1', type: 'critical', title: 'Compliance Violation', entity: 'Mining Operations Corp', message: 'License suspended due to quota violation', time: '2 hours ago' },
    { id: '2', type: 'warning', title: 'Obligation Overdue', entity: 'Digital Assets Inc', message: 'Q4 audit report not submitted', time: '4 hours ago' },
    { id: '3', type: 'info', title: 'License Expiring', entity: 'Trading Firm LLC', message: 'License expires in 30 days', time: '1 day ago' },
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-white">Alerts</h1>
        <p className="text-slate-400">Monitor and manage system alerts</p>
      </div>

      <div className="space-y-4">
        {mockAlerts.map((alert) => (
          <div key={alert.id} className={`dashboard-card border-l-4 ${
            alert.type === 'critical' ? 'border-l-red-500' :
            alert.type === 'warning' ? 'border-l-amber-500' :
            'border-l-blue-500'
          }`}>
            <div className="flex items-start justify-between">
              <div>
                <h3 className="font-semibold text-white">{alert.title}</h3>
                <p className="text-sm text-slate-400 mt-1">{alert.entity}</p>
                <p className="text-sm text-slate-500 mt-2">{alert.message}</p>
              </div>
              <span className="text-xs text-slate-500">{alert.time}</span>
            </div>
            <div className="flex gap-2 mt-4">
              <button className="btn btn-primary text-sm">View Details</button>
              <button className="btn btn-secondary text-sm">Dismiss</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
