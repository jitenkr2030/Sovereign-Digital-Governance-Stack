import { useQuery } from '@tanstack/react-query'
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  AreaChart, Area, BarChart, Bar
} from 'recharts'
import ScoreGauge from '../components/ScoreGauge'
import StatsCard from '../components/StatsCard'
import RecentActivity from '../components/RecentActivity'

// Mock data for charts
const complianceHistory = [
  { date: '2024-01-01', score: 85, gold: 12, silver: 28, bronze: 45, atRisk: 8 },
  { date: '2024-01-08', score: 87, gold: 14, silver: 30, bronze: 42, atRisk: 6 },
  { date: '2024-01-15', score: 86, gold: 13, silver: 29, bronze: 44, atRisk: 7 },
  { date: '2024-01-22', score: 88, gold: 15, silver: 31, bronze: 41, atRisk: 5 },
  { date: '2024-01-29', score: 89, gold: 16, silver: 32, bronze: 40, atRisk: 4 },
]

const energyData = [
  { time: '00:00', consumption: 450, hashrate: 120 },
  { time: '04:00', consumption: 380, hashrate: 115 },
  { time: '08:00', consumption: 520, hashrate: 140 },
  { time: '12:00', consumption: 680, hashrate: 175 },
  { time: '16:00', consumption: 720, hashrate: 185 },
  { time: '20:00', consumption: 650, hashrate: 168 },
]

export default function Dashboard() {
  // Fetch dashboard statistics
  const { data: stats } = useQuery({
    queryKey: ['dashboardStats'],
    queryFn: async () => ({
      totalEntities: 156,
      activeOperations: 89,
      pendingApplications: 12,
      overdueObligations: 5,
      averageScore: 82.5,
      goldTierCount: 24,
      atRiskCount: 8,
    }),
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Regulator Dashboard</h1>
          <p className="text-slate-400">Real-time overview of the CSIC Platform</p>
        </div>
        <div className="flex items-center gap-2 text-sm text-slate-400 live-indicator">
          Live updates enabled
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsCard
          title="Total Entities"
          value={stats?.totalEntities.toLocaleString() || '0'}
          change="+12%"
          trend="up"
          icon="users"
        />
        <StatsCard
          title="Active Operations"
          value={stats?.activeOperations.toLocaleString() || '0'}
          change="+5%"
          trend="up"
          icon="activity"
        />
        <StatsCard
          title="Pending Applications"
          value={stats?.pendingApplications.toLocaleString() || '0'}
          change="-3%"
          trend="down"
          icon="clock"
        />
        <StatsCard
          title="Overdue Obligations"
          value={stats?.overdueObligations.toLocaleString() || '0'}
          change="+2"
          trend="down"
          icon="alert"
          alert={stats?.overdueObligations > 0}
        />
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Compliance Score Overview */}
        <div className="dashboard-card lg:col-span-1">
          <div className="dashboard-card-header">
            <span>Compliance Score</span>
            <span className="text-xs text-slate-500">{stats?.averageScore || 0}% average</span>
          </div>
          <div className="flex items-center justify-center py-4">
            <ScoreGauge score={stats?.averageScore || 0} size="large" />
          </div>
          <div className="grid grid-cols-2 gap-4 mt-4 text-center">
            <div>
              <p className="text-2xl font-bold text-yellow-400">{stats?.goldTierCount || 0}</p>
              <p className="text-xs text-slate-400">Gold Tier</p>
            </div>
            <div>
              <p className="text-2xl font-bold text-red-400">{stats?.atRiskCount || 0}</p>
              <p className="text-xs text-slate-400">At Risk</p>
            </div>
          </div>
        </div>

        {/* Compliance History Chart */}
        <div className="dashboard-card lg:col-span-2">
          <div className="dashboard-card-header">
            <span>Compliance Trend</span>
            <div className="flex gap-2">
              <span className="status-indicator status-active">Gold: {complianceHistory[complianceHistory.length-1]?.gold}</span>
              <span className="status-indicator status-warning">Silver: {complianceHistory[complianceHistory.length-1]?.silver}</span>
              <span className="status-indicator status-inactive">Bronze: {complianceHistory[complianceHistory.length-1]?.bronze}</span>
            </div>
          </div>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={complianceHistory}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="date" stroke="#64748b" fontSize={12} />
                <YAxis stroke="#64748b" fontSize={12} />
                <Tooltip 
                  contentStyle={{ 
                    backgroundColor: '#1e293b', 
                    border: '1px solid #334155',
                    borderRadius: '8px'
                  }} 
                />
                <Area type="monotone" dataKey="gold" stackId="1" stroke="#eab308" fill="#eab308" fillOpacity={0.3} />
                <Area type="monotone" dataKey="silver" stackId="1" stroke="#94a3b8" fill="#94a3b8" fillOpacity={0.3} />
                <Area type="monotone" dataKey="bronze" stackId="1" stroke="#b45309" fill="#b45309" fillOpacity={0.3} />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* Energy & Hashrate */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="dashboard-card">
          <div className="dashboard-card-header">
            <span>Energy Consumption vs Hashrate</span>
            <span className="text-xs text-slate-500">Last 24 hours</span>
          </div>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={energyData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="time" stroke="#64748b" fontSize={12} />
                <YAxis yAxisId="left" stroke="#64748b" fontSize={12} />
                <YAxis yAxisId="right" orientation="right" stroke="#64748b" fontSize={12} />
                <Tooltip 
                  contentStyle={{ 
                    backgroundColor: '#1e293b', 
                    border: '1px solid #334155',
                    borderRadius: '8px'
                  }} 
                />
                <Line yAxisId="left" type="monotone" dataKey="consumption" stroke="#3b82f6" strokeWidth={2} dot={false} />
                <Line yAxisId="right" type="monotone" dataKey="hashrate" stroke="#10b981" strokeWidth={2} dot={false} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="dashboard-card">
          <div className="dashboard-card-header">
            <span>Recent Activity</span>
          </div>
          <RecentActivity />
        </div>
      </div>

      {/* Tier Distribution */}
      <div className="dashboard-card">
        <div className="dashboard-card-header">
          <span>Compliance Tier Distribution</span>
        </div>
        <div className="grid grid-cols-5 gap-4">
          {[
            { tier: 'Gold', count: complianceHistory[complianceHistory.length-1]?.gold || 0, color: 'bg-yellow-500', textColor: 'text-yellow-400' },
            { tier: 'Silver', count: complianceHistory[complianceHistory.length-1]?.silver || 0, color: 'bg-slate-400', textColor: 'text-slate-300' },
            { tier: 'Bronze', count: complianceHistory[complianceHistory.length-1]?.bronze || 0, color: 'bg-amber-700', textColor: 'text-amber-600' },
            { tier: 'At Risk', count: complianceHistory[complianceHistory.length-1]?.atRisk || 0, color: 'bg-red-500', textColor: 'text-red-400' },
            { tier: 'Critical', count: 2, color: 'bg-red-800', textColor: 'text-red-600' },
          ].map((item) => (
            <div key={item.tier} className="text-center p-4 bg-slate-700/30 rounded-lg">
              <div className={`text-3xl font-bold ${item.textColor}`}>{item.count}</div>
              <div className="text-sm text-slate-400 mt-1">{item.tier}</div>
              <div className={`h-1 mt-2 rounded-full ${item.color}`} />
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
