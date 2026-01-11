/**
 * NEAM Dashboard - Overview Page
 * Main dashboard with key metrics and regional overview
 */

import React, { useState } from 'react';
import { useAppSelector } from '../../store';
import { MetricCard, MetricCardSkeleton } from '../../components/common/MetricCard';
import { RegionHeatmap, MiniHeatmap } from '../../components/dashboard/RegionHeatmap';
import { AlertsPanel, MiniAlertsWidget } from '../../components/dashboard/AlertsPanel';
import { InterventionsPanel, MiniInterventionsWidget } from '../../components/dashboard/InterventionsPanel';
import { TrendLineChart, ComparisonBarChart, SectorPerformanceChart } from '../../components/charts/Charts';
import { Section, ExportButton } from '../../components/layout/Layout';
import { Calendar, RefreshCw, Download } from 'lucide-react';
import type { MetricCard as MetricCardType, HeatmapCell, TimeSeriesPoint } from '../../types';
import clsx from 'clsx';

// Mock data for demonstration
const mockMetrics: MetricCardType[] = [
  {
    id: 'gdp',
    name: 'GDP Growth',
    value: 7.2,
    unit: '%',
    change: 0.5,
    changeType: 'increase',
    trend: [],
    status: 'normal',
    category: 'GDP',
  },
  {
    id: 'inflation',
    name: 'Inflation Rate',
    value: 4.8,
    unit: '%',
    change: -0.3,
    changeType: 'decrease',
    trend: [],
    status: 'normal',
    category: 'INFLATION',
  },
  {
    id: 'unemployment',
    name: 'Unemployment',
    value: 4.2,
    unit: '%',
    change: -0.1,
    changeType: 'decrease',
    trend: [],
    status: 'normal',
    category: 'EMPLOYMENT',
  },
  {
    id: 'fiscal',
    name: 'Fiscal Deficit',
    value: 5.9,
    unit: '%',
    change: 0.2,
    changeType: 'increase',
    trend: [],
    status: 'warning',
    category: 'EXPENDITURE',
  },
  {
    id: 'trade',
    name: 'Trade Balance',
    value: -24.5,
    unit: '₹ Cr',
    change: 2.3,
    changeType: 'increase',
    trend: [],
    status: 'normal',
    category: 'TRADE',
  },
  {
    id: 'revenue',
    name: 'Tax Revenue',
    value: 285000,
    unit: '₹',
    change: 8.5,
    changeType: 'increase',
    trend: [],
    status: 'normal',
    category: 'REVENUE',
  },
];

const mockHeatmapData: HeatmapCell[] = [
  { regionId: 'mh', regionName: 'Maharashtra', x: 0, y: 0, value: 72, status: 'healthy', alertCount: 2, interventionCount: 1 },
  { regionId: 'dl', regionName: 'Delhi', x: 1, y: 0, value: 68, status: 'healthy', alertCount: 1, interventionCount: 0 },
  { regionId: 'ka', regionName: 'Karnataka', x: 0, y: 1, value: 75, status: 'healthy', alertCount: 0, interventionCount: 2 },
  { regionId: 'tn', regionName: 'Tamil Nadu', x: 1, y: 1, value: 65, status: 'caution', alertCount: 3, interventionCount: 1 },
  { regionId: 'up', regionName: 'Uttar Pradesh', x: 0, y: 2, value: 45, status: 'warning', alertCount: 5, interventionCount: 3 },
  { regionId: 'wb', regionName: 'West Bengal', x: 1, y: 2, value: 52, status: 'caution', alertCount: 4, interventionCount: 2 },
  { regionId: 'gj', regionName: 'Gujarat', x: 0, y: 3, value: 78, status: 'healthy', alertCount: 1, interventionCount: 0 },
  { regionId: 'rj', regionName: 'Rajasthan', x: 1, y: 3, value: 38, status: 'critical', alertCount: 7, interventionCount: 4 },
];

const mockTrendData: TimeSeriesPoint[] = Array.from({ length: 12 }, (_, i) => ({
  timestamp: `2024-${String(i + 1).padStart(2, '0')}-01`,
  value: 70 + Math.random() * 15,
}));

const OverviewPage: React.FC = () => {
  const { user } = useAppSelector((state) => state.auth);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const handleRefresh = () => {
    setIsRefreshing(true);
    setTimeout(() => setIsRefreshing(false), 1500);
  };

  return (
    <div className="space-y-6">
      {/* Welcome Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">
            Welcome back, {user?.name?.split(' ')[0] || 'User'}
          </h1>
          <p className="text-slate-500 mt-1">
            Here's what's happening with the economy today.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={handleRefresh}
            className="flex items-center gap-2 px-4 py-2 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors"
          >
            <RefreshCw className={clsx('w-4 h-4', isRefreshing && 'animate-spin')} />
            <span className="text-sm font-medium">Refresh</span>
          </button>
          <ExportButton />
        </div>
      </div>

      {/* Key Metrics Grid */}
      <Section
        title="Key Economic Indicators"
        subtitle="Latest values and period comparisons"
        action={
          <button className="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-700">
            <Calendar className="w-4 h-4" />
            Last 30 days
          </button>
        }
      >
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
          {mockMetrics.map((metric) => (
            <MetricCard key={metric.id} metric={metric} />
          ))}
        </div>
      </Section>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Regional Heatmap - Takes 2 columns */}
        <div className="lg:col-span-2">
          <RegionHeatmap
            data={mockHeatmapData}
            height={450}
            onRegionClick={(region) => console.log('Region clicked:', region)}
          />
        </div>

        {/* Side Panel - Alerts & Interventions */}
        <div className="space-y-6">
          <MiniAlertsWidget />
          <MiniInterventionsWidget />
          
          {/* Quick Stats */}
          <div className="bg-white rounded-xl border border-slate-200 p-4">
            <h3 className="font-semibold text-slate-800 mb-3">Quick Stats</h3>
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm text-slate-600">Active Alerts</span>
                <span className="px-2 py-0.5 bg-amber-100 text-amber-700 text-xs font-medium rounded-full">12</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-slate-600">Running Interventions</span>
                <span className="px-2 py-0.5 bg-emerald-100 text-emerald-700 text-xs font-medium rounded-full">5</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-slate-600">Regions Monitored</span>
                <span className="font-medium text-slate-800">28</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-slate-600">Data Freshness</span>
                <span className="text-sm text-emerald-600 font-medium">2 min ago</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <TrendLineChart
          data={mockTrendData}
          title="GDP Growth Trend"
          xAxisKey="timestamp"
          dataKeys={[{ key: 'value', name: 'GDP Growth %' }]}
          height={300}
        />
        <ComparisonBarChart
          data={[
            { name: 'Agriculture', value: 18.4 },
            { name: 'Industry', value: 25.9 },
            { name: 'Services', value: 55.7 },
          ]}
          title="Sector Contribution to GDP"
          height={300}
        />
      </div>

      {/* Alerts & Interventions Full Width */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <AlertsPanel maxItems={5} compact />
        <InterventionsPanel maxItems={3} compact />
      </div>
    </div>
  );
};

export default OverviewPage;
