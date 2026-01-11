import React, { useState } from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LineChart, Line, PieChart, Pie, Cell } from 'recharts';
import { Calendar, TrendingUp, Clock, AlertTriangle, CheckCircle, Activity } from 'lucide-react';

const Analytics: React.FC = () => {
  const [timeRange, setTimeRange] = useState('7d');

  const incidentTrendData = [
    { period: 'Week 1', critical: 2, high: 5, medium: 12, low: 8, resolved: 18 },
    { period: 'Week 2', critical: 3, high: 4, medium: 10, low: 15, resolved: 22 },
    { period: 'Week 3', critical: 1, high: 7, medium: 8, low: 10, resolved: 20 },
    { period: 'Week 4', critical: 4, high: 3, medium: 15, low: 12, resolved: 25 },
  ];

  const resolutionTimeData = [
    { category: '< 1 hour', count: 45 },
    { category: '1-4 hours', count: 32 },
    { category: '4-24 hours', count: 28 },
    { category: '1-3 days', count: 15 },
    { category: '> 3 days', count: 8 },
  ];

  const incidentTypeData = [
    { name: 'Security', value: 35, color: '#dc2626' },
    { name: 'Compliance', value: 25, color: '#f59e0b' },
    { name: 'Technical', value: 20, color: '#3b82f6' },
    { name: 'Fraud', value: 12, color: '#8b5cf6' },
    { name: 'Operational', value: 8, color: '#10b981' },
  ];

  const teamPerformanceData = [
    { member: 'John D.', resolved: 28, avgTime: '2.5h', satisfaction: 4.8 },
    { member: 'Jane S.', resolved: 24, avgTime: '3.2h', satisfaction: 4.6 },
    { member: 'Mike J.', resolved: 19, avgTime: '4.1h', satisfaction: 4.5 },
    { member: 'Sarah K.', resolved: 15, avgTime: '2.8h', satisfaction: 4.9 },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Analytics</h1>
          <p className="text-gray-400 mt-1">Incident response metrics and trends</p>
        </div>
        <div className="flex gap-2">
          {['24h', '7d', '30d', '90d'].map((range) => (
            <button
              key={range}
              onClick={() => setTimeRange(range)}
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                timeRange === range
                  ? 'bg-primary-600 text-white'
                  : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
              }`}
            >
              {range}
            </button>
          ))}
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <div className="flex items-center justify-between mb-4">
            <div className="p-3 bg-primary-500/20 rounded-lg">
              <Activity className="w-6 h-6 text-primary-400" />
            </div>
            <span className="text-green-400 text-sm flex items-center gap-1">
              <TrendingUp size={14} /> 12%
            </span>
          </div>
          <h3 className="text-gray-400 text-sm">Avg Response Time</h3>
          <p className="text-2xl font-bold text-white mt-1">2.4h</p>
        </div>

        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <div className="flex items-center justify-between mb-4">
            <div className="p-3 bg-green-500/20 rounded-lg">
              <CheckCircle className="w-6 h-6 text-green-400" />
            </div>
            <span className="text-green-400 text-sm flex items-center gap-1">
              <TrendingUp size={14} /> 8%
            </span>
          </div>
          <h3 className="text-gray-400 text-sm">Resolution Rate</h3>
          <p className="text-2xl font-bold text-white mt-1">94%</p>
        </div>

        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <div className="flex items-center justify-between mb-4">
            <div className="p-3 bg-yellow-500/20 rounded-lg">
              <Clock className="w-6 h-6 text-yellow-400" />
            </div>
            <span className="text-red-400 text-sm flex items-center gap-1">
              <TrendingUp size={14} /> 5%
            </span>
          </div>
          <h3 className="text-gray-400 text-sm">Avg Time to Detect</h3>
          <p className="text-2xl font-bold text-white mt-1">18m</p>
        </div>

        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <div className="flex items-center justify-between mb-4">
            <div className="p-3 bg-red-500/20 rounded-lg">
              <AlertTriangle className="w-6 h-6 text-red-400" />
            </div>
            <span className="text-green-400 text-sm flex items-center gap-1">
              <TrendingUp size={14} /> 15%
            </span>
          </div>
          <h3 className="text-gray-400 text-sm">False Positive Rate</h3>
          <p className="text-2xl font-bold text-white mt-1">12%</p>
        </div>
      </div>

      {/* Charts Row 1 */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Incident Trend */}
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-4">Incident Trend</h3>
          <div className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={incidentTrendData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                <XAxis dataKey="period" stroke="#9ca3af" />
                <YAxis stroke="#9ca3af" />
                <Tooltip
                  contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '8px' }}
                  labelStyle={{ color: '#f3f4f6' }}
                />
                <Bar dataKey="critical" fill="#dc2626" radius={[4, 4, 0, 0]} />
                <Bar dataKey="high" fill="#ea580c" radius={[4, 4, 0, 0]} />
                <Bar dataKey="medium" fill="#ca8a04" radius={[4, 4, 0, 0]} />
                <Bar dataKey="low" fill="#16a34a" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Resolution Time Distribution */}
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-4">Resolution Time Distribution</h3>
          <div className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={resolutionTimeData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={100}
                  paddingAngle={2}
                  dataKey="count"
                  nameKey="category"
                  label={({ category, percent }) => `${category} (${(percent * 100).toFixed(0)}%)`}
                  labelLine={false}
                >
                  {resolutionTimeData.map((entry, index) => (
                    <Cell
                      key={`cell-${index}`}
                      fill={`hsl(${index * 50}, 70%, 50%)`}
                    />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '8px' }}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* Charts Row 2 */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Incident Types */}
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-4">By Type</h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={incidentTypeData}
                  cx="50%"
                  cy="50%"
                  outerRadius={80}
                  dataKey="value"
                  nameKey="name"
                >
                  {incidentTypeData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '8px' }}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="grid grid-cols-2 gap-2 mt-4">
            {incidentTypeData.map((item) => (
              <div key={item.name} className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
                <span className="text-sm text-gray-400">{item.name}</span>
                <span className="text-sm text-white font-medium ml-auto">{item.value}%</span>
              </div>
            ))}
          </div>
        </div>

        {/* Team Performance */}
        <div className="lg:col-span-2 bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold text-white mb-4">Team Performance</h3>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-sm text-gray-400 border-b border-gray-700">
                  <th className="pb-3 font-medium">Team Member</th>
                  <th className="pb-3 font-medium">Resolved</th>
                  <th className="pb-3 font-medium">Avg Time</th>
                  <th className="pb-3 font-medium">Satisfaction</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-700">
                {teamPerformanceData.map((member, index) => (
                  <tr key={index}>
                    <td className="py-3 text-white font-medium">{member.member}</td>
                    <td className="py-3 text-gray-300">{member.resolved}</td>
                    <td className="py-3 text-gray-300">{member.avgTime}</td>
                    <td className="py-3">
                      <div className="flex items-center gap-2">
                        <div className="flex-1 h-2 bg-gray-700 rounded-full overflow-hidden">
                          <div
                            className="h-full bg-green-500 rounded-full"
                            style={{ width: `${(member.satisfaction / 5) * 100}%` }}
                          />
                        </div>
                        <span className="text-white text-sm">{member.satisfaction}</span>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Export Section */}
      <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold text-white">Export Analytics Report</h3>
            <p className="text-gray-400 text-sm mt-1">
              Generate a detailed analytics report for the selected time period
            </p>
          </div>
          <div className="flex gap-3">
            <button className="btn btn-secondary">Export PDF</button>
            <button className="btn btn-primary">Export CSV</button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Analytics;
