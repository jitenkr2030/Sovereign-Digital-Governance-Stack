import { useState, useEffect } from 'react';
import { TrendingUp, TrendingDown, AlertTriangle, Shield, Zap, FileText, CheckCircle, Clock } from 'lucide-react';
import { Card, CardHeader, CardTitle, StatsCard } from '../components/common';
import { LineChart, BarChart, PieChart } from '../components/visualizations';
import { DashboardMetrics, ActivityItem } from '../types';

// Mock data for demonstration
const mockMetrics: DashboardMetrics = {
  totalLicenses: 1247,
  activeLicenses: 1089,
  pendingApplications: 23,
  complianceScore: 94.5,
  totalEnergyConsumption: 4567,
  averageGridLoad: 72.3,
  activeAlerts: 5,
  reportsGenerated: 156,
  recentActivity: [],
};

const mockEnergyData = [
  { timestamp: '2024-01-01T00:00:00Z', value: 120, category: 'Actual' },
  { timestamp: '2024-01-01T04:00:00Z', value: 95, category: 'Actual' },
  { timestamp: '2024-01-01T08:00:00Z', value: 180, category: 'Actual' },
  { timestamp: '2024-01-01T12:00:00Z', value: 220, category: 'Actual' },
  { timestamp: '2024-01-01T16:00:00Z', value: 250, category: 'Actual' },
  { timestamp: '2024-01-01T20:00:00Z', value: 200, category: 'Actual' },
  { timestamp: '2024-01-02T00:00:00Z', value: 130, category: 'Actual' },
  { timestamp: '2024-01-01T00:00:00Z', value: 115, category: 'Predicted' },
  { timestamp: '2024-01-01T04:00:00Z', value: 100, category: 'Predicted' },
  { timestamp: '2024-01-01T08:00:00Z', value: 175, category: 'Predicted' },
  { timestamp: '2024-01-01T12:00:00Z', value: 215, category: 'Predicted' },
  { timestamp: '2024-01-01T16:00:00Z', value: 245, category: 'Predicted' },
  { timestamp: '2024-01-01T20:00:00Z', value: 195, category: 'Predicted' },
  { timestamp: '2024-01-02T00:00:00Z', value: 125, category: 'Predicted' },
];

const mockLicenseData = [
  { label: 'Active', value: 1089 },
  { label: 'Pending', value: 89 },
  { label: 'Suspended', value: 45 },
  { label: 'Revoked', value: 24 },
];

const mockEnergyMixData = [
  { name: 'Solar', value: 25, color: '#f59e0b' },
  { name: 'Wind', value: 20, color: '#3b82f6' },
  { name: 'Hydro', value: 30, color: '#10b981' },
  { name: 'Nuclear', value: 15, color: '#8b5cf6' },
  { name: 'Fossil', value: 10, color: '#6b7280' },
];

const mockActivityData = [
  { label: 'Mon', value: 12 },
  { label: 'Tue', value: 19 },
  { label: 'Wed', value: 15 },
  { label: 'Thu', value: 22 },
  { label: 'Fri', value: 18 },
  { label: 'Sat', value: 8 },
  { label: 'Sun', value: 5 },
];

export default function Dashboard() {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Simulate API call
    const fetchData = async () => {
      setIsLoading(true);
      // In production, this would call the actual API
      setTimeout(() => {
        setMetrics(mockMetrics);
        setIsLoading(false);
      }, 1000);
    };

    fetchData();
  }, []);

  const formatNumber = (num: number): string => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
    return num.toString();
  };

  return (
    <div className="space-y-6">
      {/* KPI Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsCard
          title="Total Licenses"
          value={formatNumber(metrics?.totalLicenses || 0)}
          change={{ value: 12, type: 'increase' }}
          icon={<Shield className="w-6 h-6 text-primary-600" />}
          iconBg="bg-primary-100 dark:bg-primary-900/30"
        />
        <StatsCard
          title="Compliance Score"
          value={`${metrics?.complianceScore || 0}%`}
          change={{ value: 2.3, type: 'increase' }}
          icon={<CheckCircle className="w-6 h-6 text-green-600" />}
          iconBg="bg-green-100 dark:bg-green-900/30"
        />
        <StatsCard
          title="Energy Consumption"
          value={`${formatNumber(metrics?.totalEnergyConsumption || 0)} MW`}
          change={{ value: 5.2, type: 'decrease' }}
          icon={<Zap className="w-6 h-6 text-yellow-600" />}
          iconBg="bg-yellow-100 dark:bg-yellow-900/30"
        />
        <StatsCard
          title="Active Alerts"
          value={metrics?.activeAlerts || 0}
          icon={<AlertTriangle className="w-6 h-6 text-red-600" />}
          iconBg="bg-red-100 dark:bg-red-900/30"
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Energy Consumption Chart */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Energy Consumption vs Forecast</CardTitle>
          </CardHeader>
          <CardBody>
            <LineChart
              data={mockEnergyData}
              height={300}
              categories={['Actual', 'Predicted']}
              colors={['#3b82f6', '#10b981']}
              showLegend={true}
              areaFill={true}
              yAxisLabel="MW"
            />
          </CardBody>
        </Card>

        {/* License Status Distribution */}
        <Card>
          <CardHeader>
            <CardTitle>License Status</CardTitle>
          </CardHeader>
          <CardBody>
            <PieChart
              data={mockLicenseData}
              height={300}
              donut={true}
              showLegend={true}
            />
          </CardBody>
        </Card>
      </div>

      {/* Second Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Weekly Activity */}
        <Card>
          <CardHeader>
            <CardTitle>Weekly Activity</CardTitle>
          </CardHeader>
          <CardBody>
            <BarChart
              data={mockActivityData}
              height={280}
              colors={['#3b82f6']}
              showGrid={true}
              animate={true}
            />
          </CardBody>
        </Card>

        {/* Energy Mix */}
        <Card>
          <CardHeader>
            <CardTitle>Energy Source Mix</CardTitle>
          </CardHeader>
          <CardBody>
            <div className="flex items-center gap-8">
              <div className="w-1/2">
                <PieChart
                  data={mockEnergyMixData}
                  height={220}
                  donut={true}
                  showLegend={false}
                />
              </div>
              <div className="w-1/2 space-y-3">
                {mockEnergyMixData.map((item) => (
                  <div key={item.name} className="flex items-center gap-3">
                    <div
                      className="w-3 h-3 rounded-full"
                      style={{ backgroundColor: item.color }}
                    />
                    <span className="text-sm text-dark-600 dark:text-dark-400">{item.name}</span>
                    <span className="text-sm font-medium ml-auto">{item.value}%</span>
                  </div>
                ))}
              </div>
            </div>
          </CardBody>
        </Card>
      </div>

      {/* Recent Activity and Pending Actions */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Activity */}
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
          </CardHeader>
          <CardBody>
            <div className="space-y-4">
              {[
                { type: 'license', action: 'New license approved', entity: 'CryptoCorp Exchange', time: '2 hours ago', icon: Shield },
                { type: 'energy', action: 'Grid alert resolved', entity: 'Region North', time: '4 hours ago', icon: Zap },
                { type: 'report', action: 'Monthly report generated', entity: 'Compliance Report', time: '6 hours ago', icon: FileText },
                { type: 'license', action: 'License renewal submitted', entity: 'SecureVault CASP', time: '8 hours ago', icon: Shield },
                { type: 'system', action: 'System backup completed', entity: 'Automated', time: '12 hours ago', icon: CheckCircle },
              ].map((activity, index) => (
                <div
                  key={index}
                  className="flex items-center gap-4 p-3 rounded-lg hover:bg-dark-50 dark:hover:bg-dark-800 transition-colors"
                >
                  <div className="p-2 rounded-lg bg-primary-100 dark:bg-primary-900/30">
                    <activity.icon className="w-5 h-5 text-primary-600" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-dark-900 dark:text-dark-100">
                      {activity.action}
                    </p>
                    <p className="text-sm text-dark-500 dark:text-dark-400 truncate">
                      {activity.entity}
                    </p>
                  </div>
                  <div className="flex items-center gap-1 text-dark-400 text-xs">
                    <Clock className="w-3 h-3" />
                    {activity.time}
                  </div>
                </div>
              ))}
            </div>
          </CardBody>
        </Card>

        {/* Pending Actions */}
        <Card>
          <CardHeader>
            <CardTitle>Pending Actions</CardTitle>
          </CardHeader>
          <CardBody>
            <div className="space-y-4">
              {[
                { title: 'License Application Review', count: 23, priority: 'high', color: 'red' },
                { title: 'Compliance Audits Due', count: 8, priority: 'medium', color: 'yellow' },
                { title: 'Report Approvals', count: 12, priority: 'low', color: 'blue' },
                { title: 'Renewal Reminders', count: 45, priority: 'medium', color: 'yellow' },
              ].map((action, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between p-4 rounded-lg border border-dark-200 dark:border-dark-700 hover:border-primary-500 transition-colors cursor-pointer"
                >
                  <div className="flex items-center gap-3">
                    <div
                      className={`w-2 h-2 rounded-full bg-${
                        action.color === 'red' ? 'red' : action.color === 'yellow' ? 'yellow' : 'blue'
                      }-500`}
                    />
                    <span className="font-medium text-dark-900 dark:text-dark-100">
                      {action.title}
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-2xl font-bold text-dark-900 dark:text-dark-100">
                      {action.count}
                    </span>
                    <button className="btn-secondary btn-sm">View</button>
                  </div>
                </div>
              ))}
            </div>
          </CardBody>
        </Card>
      </div>

      {/* Quick Stats Grid */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card padding="sm">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-green-100 dark:bg-green-900/30 mb-3">
              <TrendingUp className="w-6 h-6 text-green-600" />
            </div>
            <p className="text-2xl font-bold text-dark-900 dark:text-dark-50">1,089</p>
            <p className="text-sm text-dark-500 dark:text-dark-400">Active Licenses</p>
          </div>
        </Card>
        <Card padding="sm">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-blue-100 dark:bg-blue-900/30 mb-3">
              <FileText className="w-6 h-6 text-blue-600" />
            </div>
            <p className="text-2xl font-bold text-dark-900 dark:text-dark-50">156</p>
            <p className="text-sm text-dark-500 dark:text-dark-400">Reports This Month</p>
          </div>
        </Card>
        <Card padding="sm">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-yellow-100 dark:bg-yellow-900/30 mb-3">
              <Zap className="w-6 h-6 text-yellow-600" />
            </div>
            <p className="text-2xl font-bold text-dark-900 dark:text-dark-50">72.3%</p>
            <p className="text-sm text-dark-500 dark:text-dark-400">Average Grid Load</p>
          </div>
        </Card>
        <Card padding="sm">
          <div className="text-center">
            <div className="inline-flex items-center justify-center w-12 h-12 rounded-full bg-purple-100 dark:bg-purple-900/30 mb-3">
              <CheckCircle className="w-6 h-6 text-purple-600" />
            </div>
            <p className="text-2xl font-bold text-dark-900 dark:text-dark-50">94.5%</p>
            <p className="text-sm text-dark-500 dark:text-dark-400">Compliance Rate</p>
          </div>
        </Card>
      </div>
    </div>
  );
}
