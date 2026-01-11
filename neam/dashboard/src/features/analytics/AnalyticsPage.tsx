/**
 * NEAM Dashboard - Analytics Page
 * Advanced drill-down analytics and trend analysis
 */

import React, { useState } from 'react';
import { Section } from '../../components/layout/Layout';
import {
  TrendLineChart,
  StackedAreaChart,
  ComparisonBarChart,
  DistributionPieChart,
  MultiRadarChart,
  SectorPerformanceChart,
} from '../../components/charts/Charts';
import { MetricCard } from '../../components/common/MetricCard';
import { RegionHeatmap } from '../../components/dashboard/RegionHeatmap';
import { Filter, Download, TrendingUp, TrendingDown, Minus } from 'lucide-react';
import type { TimeSeriesPoint } from '../../types';
import clsx from 'clsx';

const mockTrendData: TimeSeriesPoint[] = Array.from({ length: 24 }, (_, i) => ({
  timestamp: `2023-${String(i + 1).padStart(2, '0')}-01`,
  value: 5 + Math.random() * 5,
}));

const mockComparisonData = [
  { name: 'Q1 2024', value: 7.2 },
  { name: 'Q1 2023', value: 6.7 },
  { name: 'Q1 2022', value: 4.0 },
  { name: 'Q1 2021', value: 1.6 },
];

const mockSectorData = [
  { name: 'Agriculture', value: 18.4 },
  { name: 'Manufacturing', value: 17.4 },
  { name: 'Construction', value: 8.5 },
  { name: 'Services', value: 55.7 },
];

const mockRadarData = [
  { subject: 'GDP Growth', A: 7.2, B: 6.5, fullMark: 10 },
  { subject: 'Inflation', A: 4.8, B: 5.2, fullMark: 10 },
  { subject: 'Employment', A: 8.1, B: 7.5, fullMark: 10 },
  { subject: 'Trade', A: 5.5, B: 6.0, fullMark: 10 },
  { subject: 'Revenue', A: 7.8, B: 7.0, fullMark: 10 },
  { subject: 'Investment', A: 6.5, B: 6.0, fullMark: 10 },
];

const AnalyticsPage: React.FC = () => {
  const [selectedPeriod, setSelectedPeriod] = useState('year');
  const [selectedMetric, setSelectedMetric] = useState('gdp');

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">Analytics</h1>
          <p className="text-slate-500 mt-1">
            Deep dive into economic trends and comparative analysis
          </p>
        </div>
        <div className="flex items-center gap-3">
          {/* Period Selector */}
          <div className="flex bg-white border border-slate-200 rounded-lg overflow-hidden">
            {['month', 'quarter', 'year', '5year'].map((period) => (
              <button
                key={period}
                onClick={() => setSelectedPeriod(period)}
                className={clsx(
                  'px-4 py-2 text-sm font-medium transition-colors',
                  selectedPeriod === period
                    ? 'bg-blue-500 text-white'
                    : 'text-slate-600 hover:bg-slate-50'
                )}
              >
                {period === '5year' ? '5 Years' : period.charAt(0).toUpperCase() + period.slice(1)}
              </button>
            ))}
          </div>
          <button className="flex items-center gap-2 px-4 py-2 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors">
            <Filter className="w-4 h-4" />
            <span className="text-sm font-medium">Filters</span>
          </button>
          <button className="flex items-center gap-2 px-4 py-2 bg-slate-800 text-white rounded-lg hover:bg-slate-700 transition-colors">
            <Download className="w-4 h-4" />
            <span className="text-sm font-medium">Export</span>
          </button>
        </div>
      </div>

      {/* Key Metrics Comparison */}
      <Section title="Key Metrics Comparison" subtitle="Current period vs previous periods">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { name: 'GDP Growth', value: 7.2, change: 0.5, trend: 'up' },
            { name: 'Inflation', value: 4.8, change: -0.3, trend: 'down' },
            { name: 'Unemployment', value: 4.2, change: -0.1, trend: 'down' },
            { name: 'Fiscal Deficit', value: 5.9, change: 0.2, trend: 'up' },
          ].map((metric) => (
            <div
              key={metric.name}
              className="p-4 bg-slate-50 rounded-lg border border-slate-200"
            >
              <p className="text-sm text-slate-500">{metric.name}</p>
              <div className="flex items-end gap-2 mt-1">
                <span className="text-2xl font-bold text-slate-800">
                  {metric.value}%
                </span>
                <span
                  className={clsx(
                    'flex items-center text-sm font-medium mb-1',
                    metric.trend === 'up'
                      ? 'text-emerald-600'
                      : metric.trend === 'down'
                      ? 'text-red-600'
                      : 'text-slate-500'
                  )}
                >
                  {metric.trend === 'up' ? (
                    <TrendingUp className="w-4 h-4" />
                  ) : metric.trend === 'down' ? (
                    <TrendingDown className="w-4 h-4" />
                  ) : (
                    <Minus className="w-4 h-4" />
                  )}
                  {Math.abs(metric.change)}%
                </span>
              </div>
            </div>
          ))}
        </div>
      </Section>

      {/* Main Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <TrendLineChart
          data={mockTrendData}
          title="Economic Growth Trend"
          xAxisKey="timestamp"
          dataKeys={[
            { key: 'value', name: 'Growth %', color: '#3B82F6' },
          ]}
          height={350}
        />
        <ComparisonBarChart
          data={mockComparisonData}
          title="GDP Growth by Quarter"
          height={350}
        />
      </div>

      {/* Sector & Regional Analysis */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <DistributionPieChart
            data={mockSectorData}
            title="Sector Contribution to GDP"
            height={350}
          />
        </div>
        <Section title="Performance Radar" subtitle="Multi-dimensional comparison" className="h-full">
          <MultiRadarChart data={mockRadarData} height={300} />
        </Section>
      </div>

      {/* Regional Performance */}
      <Section
        title="Regional Performance Analysis"
        subtitle="State-wise economic indicators comparison"
        action={
          <button className="text-sm text-blue-600 hover:text-blue-700 font-medium">
            View All Regions
          </button>
        }
      >
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <ComparisonBarChart
            data={[
              { name: 'Maharashtra', value: 8.1, comparison: 7.5 },
              { name: 'Karnataka', value: 7.8, comparison: 7.2 },
              { name: 'Tamil Nadu', value: 7.5, comparison: 7.0 },
              { name: 'Gujarat', value: 7.2, comparison: 6.8 },
              { name: 'UP', value: 5.5, comparison: 5.0 },
            ]}
            title="State GDP Growth Comparison"
            height={300}
            showComparison
            colors={{ main: '#3B82F6', comparison: '#94A3B8' }}
          />
          <TrendLineChart
            data={mockTrendData}
            title="Regional Performance Trend"
            xAxisKey="timestamp"
            dataKeys={[
              { key: 'value', name: 'National Avg', color: '#3B82F6' },
            ]}
            height={300}
          />
        </div>
      </Section>

      {/* Sector Performance Table */}
      <Section title="Detailed Sector Breakdown" subtitle="Sector-wise performance metrics">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-slate-200">
                <th className="text-left py-3 px-4 text-sm font-medium text-slate-500">Sector</th>
                <th className="text-right py-3 px-4 text-sm font-medium text-slate-500">Contribution</th>
                <th className="text-right py-3 px-4 text-sm font-medium text-slate-500">Growth</th>
                <th className="text-right py-3 px-4 text-sm font-medium text-slate-500">Employment</th>
                <th className="text-right py-3 px-4 text-sm font-medium text-slate-500">Investment</th>
                <th className="text-right py-3 px-4 text-sm font-medium text-slate-500">Trend</th>
              </tr>
            </thead>
            <tbody>
              {[
                { sector: 'Agriculture', contribution: 18.4, growth: 3.5, employment: 42, investment: 4.2, trend: 'stable' },
                { sector: 'Manufacturing', contribution: 17.4, growth: 6.8, employment: 12, investment: 8.5, trend: 'rising' },
                { sector: 'Services', contribution: 55.7, growth: 8.2, employment: 28, investment: 12.3, trend: 'rising' },
                { sector: 'Construction', contribution: 8.5, growth: 7.1, employment: 10, investment: 15.8, trend: 'rising' },
              ].map((row) => (
                <tr key={row.sector} className="border-b border-slate-100 hover:bg-slate-50">
                  <td className="py-3 px-4 text-sm font-medium text-slate-800">{row.sector}</td>
                  <td className="py-3 px-4 text-sm text-right text-slate-600">{row.contribution}%</td>
                  <td className="py-3 px-4 text-sm text-right text-slate-600">{row.growth}%</td>
                  <td className="py-3 px-4 text-sm text-right text-slate-600">{row.employment}%</td>
                  <td className="py-3 px-4 text-sm text-right text-slate-600">{row.investment}%</td>
                  <td className="py-3 px-4 text-sm text-right">
                    {row.trend === 'rising' ? (
                      <TrendingUp className="w-4 h-4 text-emerald-500 inline" />
                    ) : row.trend === 'falling' ? (
                      <TrendingDown className="w-4 h-4 text-red-500 inline" />
                    ) : (
                      <Minus className="w-4 h-4 text-slate-400 inline" />
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Section>
    </div>
  );
};

export default AnalyticsPage;
