import { useState, useEffect } from 'react';
import { Zap, TrendingUp, TrendingDown, Activity, AlertTriangle, RefreshCw } from 'lucide-react';
import { Card, CardHeader, CardTitle, Button, StatsCard } from '../components/common';
import { LineChart, BarChart, PieChart } from '../components/visualizations';
import { EnergyMetrics, RegionalEnergyData, GridStatus } from '../types';

// Mock data
const mockCurrentMetrics: EnergyMetrics = {
  timestamp: new Date().toISOString(),
  region: 'National',
  consumption: 4567,
  demand: 4890,
  supply: 5200,
  renewablePercentage: 42.5,
  carbonEmissions: 1250,
  gridFrequency: 50.02,
  voltage: 230,
};

const mockEnergyData = Array.from({ length: 24 }, (_, i) => ({
  timestamp: new Date(Date.now() - (23 - i) * 3600000).toISOString(),
  value: 3500 + Math.random() * 2000,
  category: i < 12 ? 'Consumption' : 'Consumption',
}));

const mockForecastData = Array.from({ length: 48 }, (_, i) => ({
  timestamp: new Date(Date.now() + i * 3600000).toISOString(),
  value: 4000 + Math.random() * 1500,
  category: i < 24 ? 'Forecast' : 'Long-term Forecast',
}));

const mockRegionalData: RegionalEnergyData[] = [
  { region: 'North', totalConsumption: 1200, peakDemand: 1500, averageLoad: 75, renewablePercentage: 45, numberOfNodes: 45, carbonIntensity: 0.35 },
  { region: 'South', totalConsumption: 1800, peakDemand: 2100, averageLoad: 82, renewablePercentage: 38, numberOfNodes: 62, carbonIntensity: 0.42 },
  { region: 'East', totalConsumption: 950, peakDemand: 1100, averageLoad: 68, renewablePercentage: 52, numberOfNodes: 38, carbonIntensity: 0.28 },
  { region: 'West', totalConsumption: 617, peakDemand: 750, averageLoad: 71, renewablePercentage: 48, numberOfNodes: 28, carbonIntensity: 0.32 },
];

const mockGridStatus: GridStatus[] = [
  { region: 'North', status: 'normal', frequency: 50.02, loadPercentage: 72, availableCapacity: 450, lastUpdated: new Date().toISOString() },
  { region: 'South', status: 'warning', frequency: 49.98, loadPercentage: 88, availableCapacity: 180, lastUpdated: new Date().toISOString() },
  { region: 'East', status: 'normal', frequency: 50.01, loadPercentage: 65, availableCapacity: 380, lastUpdated: new Date().toISOString() },
  { region: 'West', status: 'normal', frequency: 50.03, loadPercentage: 70, availableCapacity: 220, lastUpdated: new Date().toISOString() },
];

const mockEnergyMix = [
  { name: 'Solar', value: 18, color: '#f59e0b' },
  { name: 'Wind', value: 22, color: '#3b82f6' },
  { name: 'Hydro', value: 25, color: '#10b981' },
  { name: 'Nuclear', value: 20, color: '#8b5cf6' },
  { name: 'Natural Gas', value: 10, color: '#6b7280' },
  { name: 'Coal', value: 5, color: '#374151' },
];

const mockConsumptionTrend = [
  { label: '00:00', value: 3200 },
  { label: '04:00', value: 2800 },
  { label: '08:00', value: 4200 },
  { label: '12:00', value: 5100 },
  { label: '16:00', value: 4800 },
  { label: '20:00', value: 4400 },
];

export default function Energy() {
  const [metrics, setMetrics] = useState<EnergyMetrics | null>(null);
  const [gridStatus, setGridStatus] = useState<GridStatus[]>([]);
  const [regionalData, setRegionalData] = useState<RegionalEnergyData[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedRegion, setSelectedRegion] = useState('National');
  const [timeRange, setTimeRange] = useState('24h');

  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true);
      setTimeout(() => {
        setMetrics(mockCurrentMetrics);
        setGridStatus(mockGridStatus);
        setRegionalData(mockRegionalData);
        setIsLoading(false);
      }, 1000);
    };

    fetchData();
    const interval = setInterval(fetchData, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const formatNumber = (num: number): string => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`;
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`;
    return num.toString();
  };

  const statusColors = {
    normal: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
    warning: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
    critical: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-dark-900 dark:text-dark-50">Energy Analytics</h1>
          <p className="text-dark-500 dark:text-dark-400 mt-1">
            Monitor energy consumption, grid performance, and renewable integration
          </p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={selectedRegion}
            onChange={(e) => setSelectedRegion(e.target.value)}
            className="px-3 py-2 rounded-lg border border-dark-300 dark:border-dark-600 bg-white dark:bg-dark-800 text-sm"
          >
            <option value="National">National</option>
            <option value="North">North Region</option>
            <option value="South">South Region</option>
            <option value="East">East Region</option>
            <option value="West">West Region</option>
          </select>
          <select
            value={timeRange}
            onChange={(e) => setTimeRange(e.target.value)}
            className="px-3 py-2 rounded-lg border border-dark-300 dark:border-dark-600 bg-white dark:bg-dark-800 text-sm"
          >
            <option value="1h">Last Hour</option>
            <option value="24h">Last 24 Hours</option>
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
          </select>
          <Button variant="secondary" leftIcon={<RefreshCw className="w-4 h-4" />}>
            Refresh
          </Button>
        </div>
      </div>

      {/* Real-time Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatsCard
          title="Current Load"
          value={`${formatNumber(metrics?.demand || 0)} MW`}
          change={{ value: 2.5, type: 'increase' }}
          icon={<Activity className="w-6 h-6 text-primary-600" />}
          iconBg="bg-primary-100 dark:bg-primary-900/30"
        />
        <StatsCard
          title="Grid Frequency"
          value={`${metrics?.gridFrequency.toFixed(2) || '0'} Hz`}
          icon={<Zap className="w-6 h-6 text-green-600" />}
          iconBg="bg-green-100 dark:bg-green-900/30"
        />
        <StatsCard
          title="Renewable Mix"
          value={`${metrics?.renewablePercentage || 0}%`}
          change={{ value: 3.2, type: 'increase' }}
          icon={<TrendingUp className="w-6 h-6 text-blue-600" />}
          iconBg="bg-blue-100 dark:bg-blue-900/30"
        />
        <StatsCard
          title="Carbon Emissions"
          value={`${formatNumber(metrics?.carbonEmissions || 0)} t`}
          change={{ value: 1.8, type: 'decrease' }}
          icon={<TrendingDown className="w-6 h-6 text-yellow-600" />}
          iconBg="bg-yellow-100 dark:bg-yellow-900/30"
        />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Energy Consumption Chart */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Energy Consumption</CardTitle>
          </CardHeader>
          <CardBody>
            <LineChart
              data={mockEnergyData}
              height={300}
              showGrid={true}
              showLegend={false}
              areaFill={true}
              colors={['#3b82f6']}
              yAxisLabel="MW"
            />
          </CardBody>
        </Card>

        {/* Energy Mix */}
        <Card>
          <CardHeader>
            <CardTitle>Energy Mix</CardTitle>
          </CardHeader>
          <CardBody>
            <PieChart
              data={mockEnergyMix}
              height={280}
              donut={true}
              showLegend={true}
            />
          </CardBody>
        </Card>
      </div>

      {/* Forecast and Regional */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Load Forecast */}
        <Card>
          <CardHeader>
            <CardTitle>Load Forecast</CardTitle>
          </CardHeader>
          <CardBody>
            <LineChart
              data={mockForecastData}
              height={280}
              categories={['Forecast', 'Long-term Forecast']}
              colors={['#10b981', '#8b5cf6']}
              showLegend={true}
              areaFill={true}
              yAxisLabel="MW"
            />
          </CardBody>
        </Card>

        {/* Peak Consumption */}
        <Card>
          <CardHeader>
            <CardTitle>Peak Consumption Hours</CardTitle>
          </CardHeader>
          <CardBody>
            <BarChart
              data={mockConsumptionTrend}
              height={280}
              colors={['#3b82f6']}
              showGrid={true}
            />
          </CardBody>
        </Card>
      </div>

      {/* Grid Status and Regional Overview */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Grid Status */}
        <Card>
          <CardHeader>
            <CardTitle>Grid Status</CardTitle>
          </CardHeader>
          <CardBody>
            <div className="space-y-4">
              {gridStatus.map((region) => (
                <div
                  key={region.region}
                  className="flex items-center justify-between p-4 rounded-lg border border-dark-200 dark:border-dark-700"
                >
                  <div className="flex items-center gap-3">
                    <div
                      className={`w-2 h-2 rounded-full ${
                        region.status === 'normal'
                          ? 'bg-green-500'
                          : region.status === 'warning'
                          ? 'bg-yellow-500'
                          : 'bg-red-500'
                      }`}
                    />
                    <div>
                      <p className="font-medium">{region.region} Region</p>
                      <p className="text-sm text-dark-500">
                        Frequency: {region.frequency.toFixed(2)} Hz
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-6">
                    <div className="text-right">
                      <p className="text-sm text-dark-500">Load</p>
                      <p className="font-semibold">{region.loadPercentage}%</p>
                    </div>
                    <div className="text-right">
                      <p className="text-sm text-dark-500">Available</p>
                      <p className="font-semibold">{region.availableCapacity} MW</p>
                    </div>
                    <span className={`badge ${statusColors[region.status]}`}>
                      {region.status}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </CardBody>
        </Card>

        {/* Regional Overview */}
        <Card>
          <CardHeader>
            <CardTitle>Regional Overview</CardTitle>
          </CardHeader>
          <CardBody>
            <div className="overflow-x-auto">
              <table className="min-w-full">
                <thead>
                  <tr className="text-left text-xs font-semibold text-dark-500 uppercase">
                    <th className="pb-3">Region</th>
                    <th className="pb-3">Consumption</th>
                    <th className="pb-3">Peak Demand</th>
                    <th className="pb-3">Renewable</th>
                    <th className="pb-3">Carbon</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-dark-200 dark:divide-dark-700">
                  {regionalData.map((region) => (
                    <tr key={region.region} className="hover:bg-dark-50 dark:hover:bg-dark-800">
                      <td className="py-3 font-medium">{region.region}</td>
                      <td className="py-3">{formatNumber(region.totalConsumption)} MW</td>
                      <td className="py-3">{formatNumber(region.peakDemand)} MW</td>
                      <td className="py-3">
                        <div className="flex items-center gap-2">
                          <div className="w-16 h-2 bg-dark-200 dark:bg-dark-700 rounded-full overflow-hidden">
                            <div
                              className="h-full bg-green-500 rounded-full"
                              style={{ width: `${region.renewablePercentage}%` }}
                            />
                          </div>
                          <span className="text-sm">{region.renewablePercentage}%</span>
                        </div>
                      </td>
                      <td className="py-3 text-sm">{region.carbonIntensity.toFixed(2)} t/MWh</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardBody>
        </Card>
      </div>

      {/* Alerts Section */}
      <Card>
        <CardHeader>
          <CardTitle>Recent Grid Alerts</CardTitle>
        </CardHeader>
        <CardBody>
          <div className="space-y-3">
            {[
              {
                severity: 'warning',
                region: 'South Region',
                message: 'Load approaching 90% capacity - Consider load balancing',
                time: '5 minutes ago',
              },
              {
                severity: 'info',
                region: 'North Region',
                message: 'Solar generation peak expected at 14:00',
                time: '1 hour ago',
              },
              {
                severity: 'success',
                region: 'East Region',
                message: 'Wind farm connection restored - Full capacity available',
                time: '2 hours ago',
              },
            ].map((alert, index) => (
              <div
                key={index}
                className={`flex items-start gap-3 p-4 rounded-lg ${
                  alert.severity === 'warning'
                    ? 'bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800'
                    : alert.severity === 'success'
                    ? 'bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800'
                    : 'bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800'
                }`}
              >
                <AlertTriangle
                  className={`w-5 h-5 flex-shrink-0 mt-0.5 ${
                    alert.severity === 'warning'
                      ? 'text-yellow-600'
                      : alert.severity === 'success'
                      ? 'text-green-600'
                      : 'text-blue-600'
                  }`}
                />
                <div className="flex-1">
                  <p className="font-medium">{alert.region}</p>
                  <p className="text-sm opacity-80">{alert.message}</p>
                </div>
                <span className="text-xs opacity-60">{alert.time}</span>
              </div>
            ))}
          </div>
        </CardBody>
      </Card>
    </div>
  );
}
