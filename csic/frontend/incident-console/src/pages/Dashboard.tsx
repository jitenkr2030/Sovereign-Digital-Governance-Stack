import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  AlertTriangle,
  TrendingUp,
  TrendingDown,
  Clock,
  CheckCircle,
  Activity,
  Shield,
  AlertCircle,
  ArrowRight,
} from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { useIncidentStore } from '../store/incidentStore';
import { IncidentSeverity, IncidentStatus } from '../types/incident';

const Dashboard: React.FC = () => {
  const { dashboardData, incidents, isLoading } = useIncidentStore();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const stats = dashboardData?.stats || {
    total: incidents.length,
    bySeverity: { critical: 0, high: 0, medium: 0, low: 0, info: 0 },
    byStatus: { open: 0, investigating: 0, containment: 0, resolved: 0, closed: 0 },
    byType: { security: 0, compliance: 0, technical: 0, operational: 0, fraud: 0 },
    averageResolutionTime: 180,
    openIncidentsCount: 0,
    criticalIncidentsCount: 0,
  };

  const severityColors: Record<IncidentSeverity, string> = {
    critical: '#dc2626',
    high: '#ea580c',
    medium: '#ca8a04',
    low: '#16a34a',
    info: '#0284c7',
  };

  const statusColors: Record<IncidentStatus, string> = {
    open: '#3b82f6',
    investigating: '#8b5cf6',
    containment: '#f59e0b',
    resolved: '#10b981',
    closed: '#6b7280',
  };

  const trendData = dashboardData?.trendData || [
    { date: 'Mon', critical: 2, high: 5, medium: 8, low: 12, resolved: 15 },
    { date: 'Tue', critical: 3, high: 4, medium: 10, low: 8, resolved: 18 },
    { date: 'Wed', critical: 1, high: 6, medium: 7, low: 10, resolved: 22 },
    { date: 'Thu', critical: 4, high: 3, medium: 12, low: 6, resolved: 20 },
    { date: 'Fri', critical: 2, high: 7, medium: 9, low: 9, resolved: 25 },
    { date: 'Sat', critical: 1, high: 2, medium: 5, low: 15, resolved: 12 },
    { date: 'Sun', critical: 0, high: 3, medium: 6, low: 11, resolved: 14 },
  ];

  const severityPieData = Object.entries(stats.bySeverity).map(([name, value]) => ({
    name: name.charAt(0).toUpperCase() + name.slice(1),
    value,
    color: severityColors[name as IncidentSeverity],
  })).filter(d => d.value > 0);

  const StatCard: React.FC<{
    title: string;
    value: string | number;
    change?: number;
    icon: React.ReactNode;
    color: string;
  }> = ({ title, value, change, icon, color }) => (
    <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
      <div className="flex items-center justify-between mb-4">
        <div className={`p-3 rounded-lg ${color}`}>{icon}</div>
        {change !== undefined && (
          <div className={`flex items-center gap-1 text-sm ${change >= 0 ? 'text-green-400' : 'text-red-400'}`}>
            {change >= 0 ? <TrendingUp size={16} /> : <TrendingDown size={16} />}
            <span>{Math.abs(change)}%</span>
          </div>
        )}
      </div>
      <h3 className="text-gray-400 text-sm">{title}</h3>
      <p className="text-2xl font-bold text-white mt-1">{value}</p>
    </div>
  );

  return (
    <div className={`space-y-6 transition-opacity duration-500 ${mounted ? 'opacity-100' : 'opacity-0'}`}>
      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Incidents"
          value={stats.total}
          change={12}
          icon={<AlertTriangle size={20} className="text-white" />}
          color="bg-primary-600"
        />
        <StatCard
          title="Critical Incidents"
          value={stats.criticalIncidentsCount}
          change={-8}
          icon={<AlertCircle size={20} className="text-white" />}
          color="bg-red-600"
        />
        <StatCard
          title="Open Incidents"
          value={stats.openIncidentsCount}
          change={5}
          icon={<Clock size={20} className="text-white" />}
          color="bg-yellow-600"
        />
        <StatCard
          title="Avg Resolution Time"
          value={`${Math.floor(stats.averageResolutionTime / 60)}h ${stats.averageResolutionTime % 60}m`}
          change={-15}
          icon={<CheckCircle size={20} className="text-white" />}
          color="bg-green-600"
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Trend Chart */}
        <div className="lg:col-span-2 bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-4">Incident Trend (Last 7 Days)</h3>
          <div className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={trendData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                <XAxis dataKey="date" stroke="#9ca3af" />
                <YAxis stroke="#9ca3af" />
                <Tooltip
                  contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '8px' }}
                  labelStyle={{ color: '#f3f4f6' }}
                />
                <Line type="monotone" dataKey="critical" stroke="#dc2626" strokeWidth={2} dot={{ fill: '#dc2626' }} />
                <Line type="monotone" dataKey="high" stroke="#ea580c" strokeWidth={2} dot={{ fill: '#ea580c' }} />
                <Line type="monotone" dataKey="medium" stroke="#ca8a04" strokeWidth={2} dot={{ fill: '#ca8a04' }} />
                <Line type="monotone" dataKey="low" stroke="#16a34a" strokeWidth={2} dot={{ fill: '#16a34a' }} />
                <Line type="monotone" dataKey="resolved" stroke="#10b981" strokeWidth={2} dot={{ fill: '#10b981' }} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Severity Distribution */}
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-4">By Severity</h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={severityPieData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={80}
                  paddingAngle={5}
                  dataKey="value"
                >
                  {severityPieData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '8px' }}
                  labelStyle={{ color: '#f3f4f6' }}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="grid grid-cols-2 gap-2 mt-4">
            {severityPieData.map((item) => (
              <div key={item.name} className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
                <span className="text-sm text-gray-400">{item.name}</span>
                <span className="text-sm text-white font-medium ml-auto">{item.value}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Recent Incidents */}
      <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-white">Recent Incidents</h3>
          <Link
            to="/incidents"
            className="text-sm text-primary-400 hover:text-primary-300 flex items-center gap-1"
          >
            View all <ArrowRight size={16} />
          </Link>
        </div>
        <div className="overflow-x-auto">
          <table className="data-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Title</th>
                <th>Severity</th>
                <th>Status</th>
                <th>Type</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {incidents.slice(0, 5).map((incident) => (
                <tr key={incident.id}>
                  <td className="font-mono text-primary-400">{incident.id}</td>
                  <td className="text-white max-w-xs truncate">{incident.title}</td>
                  <td>
                    <span className={`status-badge status-${incident.severity}`}>
                      {incident.severity}
                    </span>
                  </td>
                  <td>
                    <span className="status-badge" style={{
                      backgroundColor: `${statusColors[incident.status]}20`,
                      color: statusColors[incident.status],
                      borderColor: `${statusColors[incident.status]}30`
                    }}>
                      {incident.status}
                    </span>
                  </td>
                  <td className="text-gray-400 capitalize">{incident.type}</td>
                  <td className="text-gray-400">
                    {new Date(incident.createdAt).toLocaleDateString()}
                  </td>
                  <td>
                    <Link
                      to={`/incidents/${incident.id}`}
                      className="text-primary-400 hover:text-primary-300 text-sm"
                    >
                      View
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Link
          to="/incidents/new"
          className="flex items-center gap-4 p-4 bg-gray-800 rounded-xl border border-gray-700 hover:border-primary-500/50 transition-colors"
        >
          <div className="p-3 bg-primary-600 rounded-lg">
            <AlertTriangle className="w-6 h-6 text-white" />
          </div>
          <div>
            <h4 className="font-medium text-white">Report Incident</h4>
            <p className="text-sm text-gray-400">Create a new incident report</p>
          </div>
        </Link>
        <Link
          to="/analytics"
          className="flex items-center gap-4 p-4 bg-gray-800 rounded-xl border border-gray-700 hover:border-primary-500/50 transition-colors"
        >
          <div className="p-3 bg-green-600 rounded-lg">
            <Activity className="w-6 h-6 text-white" />
          </div>
          <div>
            <h4 className="font-medium text-white">View Analytics</h4>
            <p className="text-sm text-gray-400">Detailed incident analysis</p>
          </div>
        </Link>
        <Link
          to="/settings"
          className="flex items-center gap-4 p-4 bg-gray-800 rounded-xl border border-gray-700 hover:border-primary-500/50 transition-colors"
        >
          <div className="p-3 bg-purple-600 rounded-lg">
            <Shield className="w-6 h-6 text-white" />
          </div>
          <div>
            <h4 className="font-medium text-white">Response Playbooks</h4>
            <p className="text-sm text-gray-400">Configure response procedures</p>
          </div>
        </Link>
      </div>
    </div>
  );
};

export default Dashboard;
