import { useState } from 'react'

interface AuditLog {
  id: string
  timestamp: string
  entity: string
  action: string
  actor: string
  resource: string
  status: string
}

export default function AuditLogs() {
  const [filters, setFilters] = useState({
    entity: '',
    action: '',
    dateFrom: '',
    dateTo: '',
  })

  const mockLogs: AuditLog[] = [
    { id: '1', timestamp: '2024-01-29 14:30:00', entity: 'Crypto Exchange Pro', action: 'LICENSE_ISSUED', actor: 'Admin', resource: 'License #LCC-2024-001', status: 'SUCCESS' },
    { id: '2', timestamp: '2024-01-29 14:15:00', entity: 'Mining Corp', action: 'HASHREPORT_RECEIVED', actor: 'System', resource: 'Hashrate: 150 TH/s', status: 'SUCCESS' },
    { id: '3', timestamp: '2024-01-29 13:45:00', entity: 'Digital Assets Inc', action: 'LICENSE_SUSPENDED', actor: 'Compliance Officer', resource: 'License #LCC-2024-002', status: 'SUCCESS' },
    { id: '4', timestamp: '2024-01-29 13:30:00', entity: 'Trading Firm LLC', action: 'OBLIGATION_CREATED', actor: 'System', resource: 'Q4 Audit Report', status: 'SUCCESS' },
    { id: '5', timestamp: '2024-01-29 12:00:00', entity: 'ATM Services Inc', action: 'COMPLIANCE_VIOLATION', actor: 'System', resource: 'Overdue obligation detected', status: 'WARNING' },
  ]

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Audit Logs</h1>
          <p className="text-slate-400">View complete audit trail of all system activities</p>
        </div>
        <button className="btn btn-secondary">Export to CSV</button>
      </div>

      {/* Filters */}
      <div className="dashboard-card">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <input
            type="text"
            placeholder="Entity ID or Name"
            value={filters.entity}
            onChange={(e) => setFilters({ ...filters, entity: e.target.value })}
            className="search-input"
          />
          <select
            value={filters.action}
            onChange={(e) => setFilters({ ...filters, action: e.target.value })}
            className="search-input"
          >
            <option value="">All Actions</option>
            <option value="LICENSE_ISSUED">License Issued</option>
            <option value="LICENSE_SUSPENDED">License Suspended</option>
            <option value="HASHREPORT_RECEIVED">Hashrate Report</option>
            <option value="OBLIGATION_CREATED">Obligation Created</option>
          </select>
          <input
            type="date"
            value={filters.dateFrom}
            onChange={(e) => setFilters({ ...filters, dateFrom: e.target.value })}
            className="search-input"
          />
          <input
            type="date"
            value={filters.dateTo}
            onChange={(e) => setFilters({ ...filters, dateTo: e.target.value })}
            className="search-input"
          />
        </div>
      </div>

      {/* Audit Log Table */}
      <div className="dashboard-card overflow-hidden">
        <table className="data-table">
          <thead>
            <tr>
              <th>Timestamp</th>
              <th>Entity</th>
              <th>Action</th>
              <th>Actor</th>
              <th>Resource</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {mockLogs.map((log) => (
              <tr key={log.id}>
                <td className="text-slate-400 font-mono text-xs">{log.timestamp}</td>
                <td className="text-white">{log.entity}</td>
                <td>
                  <span className="px-2 py-1 bg-slate-700 rounded text-xs font-mono">
                    {log.action}
                  </span>
                </td>
                <td className="text-slate-400">{log.actor}</td>
                <td className="text-slate-400 text-xs truncate max-w-xs" title={log.resource}>
                  {log.resource}
                </td>
                <td>
                  <span className={`status-indicator ${
                    log.status === 'SUCCESS' ? 'status-active' :
                    log.status === 'WARNING' ? 'status-warning' :
                    'status-critical'
                  }`}>
                    {log.status}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
