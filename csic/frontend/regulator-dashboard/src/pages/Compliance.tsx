import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'

interface Entity {
  id: string
  name: string
  type: string
  complianceScore: number
  tier: string
  licenseStatus: string
  lastAudit: string
}

export default function Compliance() {
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState('all')

  // Mock data - in production, this would call the API
  const { data: entities } = useQuery({
    queryKey: ['entities'],
    queryFn: async (): Promise<Entity[]> => [
      { id: '1', name: 'Crypto Exchange Pro', type: 'EXCHANGE', complianceScore: 95, tier: 'GOLD', licenseStatus: 'ACTIVE', lastAudit: '2024-01-15' },
      { id: '2', name: 'Digital Assets Inc', type: 'CUSTODY', complianceScore: 78, tier: 'SILVER', licenseStatus: 'ACTIVE', lastAudit: '2024-01-10' },
      { id: '3', name: 'Mining Operations Corp', type: 'MINING', complianceScore: 45, tier: 'AT_RISK', licenseStatus: 'SUSPENDED', lastAudit: '2024-01-08' },
      { id: '4', name: 'Trading Firm LLC', type: 'TRADING', complianceScore: 88, tier: 'GOLD', licenseStatus: 'ACTIVE', lastAudit: '2024-01-14' },
      { id: '5', name: 'ATM Services Inc', type: 'ATM', complianceScore: 62, tier: 'BRONZE', licenseStatus: 'ACTIVE', lastAudit: '2024-01-05' },
    ],
  })

  const filteredEntities = entities?.filter(entity => {
    const matchesSearch = entity.name.toLowerCase().includes(searchQuery.toLowerCase())
    const matchesStatus = statusFilter === 'all' || entity.tier === statusFilter
    return matchesSearch && matchesStatus
  })

  const getTierColor = (tier: string) => {
    switch (tier) {
      case 'GOLD': return 'text-yellow-400 bg-yellow-500/20'
      case 'SILVER': return 'text-slate-300 bg-slate-500/20'
      case 'BRONZE': return 'text-amber-600 bg-amber-500/20'
      case 'AT_RISK': return 'text-orange-400 bg-orange-500/20'
      case 'CRITICAL': return 'text-red-400 bg-red-500/20'
      default: return 'text-slate-400 bg-slate-500/20'
    }
  }

  const getLicenseStatusColor = (status: string) => {
    switch (status) {
      case 'ACTIVE': return 'text-emerald-400'
      case 'SUSPENDED': return 'text-amber-400'
      case 'REVOKED': return 'text-red-400'
      default: return 'text-slate-400'
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Compliance Management</h1>
          <p className="text-slate-400">Monitor and manage entity compliance</p>
        </div>
        <button className="btn btn-primary">Generate Report</button>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-4">
        <input
          type="text"
          placeholder="Search entities..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="search-input flex-1 max-w-md"
        />
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="search-input"
        >
          <option value="all">All Tiers</option>
          <option value="GOLD">Gold</option>
          <option value="SILVER">Silver</option>
          <option value="BRONZE">Bronze</option>
          <option value="AT_RISK">At Risk</option>
          <option value="CRITICAL">Critical</option>
        </select>
      </div>

      {/* Entity Table */}
      <div className="dashboard-card overflow-hidden">
        <table className="data-table">
          <thead>
            <tr>
              <th>Entity</th>
              <th>Type</th>
              <th>Compliance Score</th>
              <th>Tier</th>
              <th>License Status</th>
              <th>Last Audit</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredEntities?.map((entity) => (
              <tr key={entity.id}>
                <td className="font-medium text-white">{entity.name}</td>
                <td className="text-slate-400">{entity.type}</td>
                <td>
                  <div className="flex items-center gap-2">
                    <div className="w-24 h-2 bg-slate-700 rounded-full overflow-hidden">
                      <div 
                        className={`h-full rounded-full ${
                          entity.complianceScore >= 90 ? 'bg-yellow-500' :
                          entity.complianceScore >= 75 ? 'bg-slate-400' :
                          entity.complianceScore >= 60 ? 'bg-amber-600' :
                          entity.complianceScore >= 40 ? 'bg-orange-500' :
                          'bg-red-500'
                        }`}
                        style={{ width: `${entity.complianceScore}%` }}
                      />
                    </div>
                    <span className="text-sm text-white">{entity.complianceScore}</span>
                  </div>
                </td>
                <td>
                  <span className={`px-2 py-1 rounded-full text-xs font-medium ${getTierColor(entity.tier)}`}>
                    {entity.tier}
                  </span>
                </td>
                <td className={getLicenseStatusColor(entity.licenseStatus)}>
                  {entity.licenseStatus}
                </td>
                <td className="text-slate-400">{entity.lastAudit}</td>
                <td>
                  <button className="text-blue-400 hover:text-blue-300 text-sm">
                    View Details
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Compliance Summary */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="dashboard-card text-center">
          <p className="text-3xl font-bold text-yellow-400">24</p>
          <p className="text-sm text-slate-400 mt-1">Gold Tier Entities</p>
        </div>
        <div className="dashboard-card text-center">
          <p className="text-3xl font-bold text-slate-300">56</p>
          <p className="text-sm text-slate-400 mt-1">Silver Tier Entities</p>
        </div>
        <div className="dashboard-card text-center">
          <p className="text-3xl font-bold text-red-400">8</p>
          <p className="text-sm text-slate-400 mt-1">At Risk / Critical</p>
        </div>
      </div>
    </div>
  )
}
