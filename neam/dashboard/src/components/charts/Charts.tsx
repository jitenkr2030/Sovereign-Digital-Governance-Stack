/**
 * NEAM Dashboard - Charts Component
 * Recharts-based chart components for economic visualization
 */

import React, { useMemo } from 'react';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  ComposedChart,
  RadarChart,
  Radar,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
} from 'recharts';
import type { TimeSeriesPoint, TrendDirection, SectorPerformance } from '../../types';
import clsx from 'clsx';

// Color palette for charts
export const CHART_COLORS = [
  '#3B82F6', // Blue 500
  '#10B981', // Emerald 500
  '#F59E0B', // Amber 500
  '#EF4444', // Red 500
  '#8B5CF6', // Violet 500
  '#EC4899', // Pink 500
  '#06B6D4', // Cyan 500
  '#84CC16', // Lime 500
];

export const STATUS_COLORS = {
  healthy: '#10B981',
  caution: '#F59E0B',
  warning: '#EF4444',
  critical: '#7C3AED',
};

// Line Chart for trends
interface TrendLineChartProps {
  data: TimeSeriesPoint[];
  title?: string;
  xAxisKey?: string;
  dataKeys?: { key: string; name: string; color?: string }[];
  height?: number;
  showGrid?: boolean;
  showLegend?: boolean;
  formatXAxis?: (value: string) => string;
  formatTooltip?: (value: number, name: string) => string;
}

export const TrendLineChart: React.FC<TrendLineChartProps> = ({
  data,
  title,
  xAxisKey = 'timestamp',
  dataKeys = [{ key: 'value', name: 'Value' }],
  height = 300,
  showGrid = true,
  showLegend = true,
  formatXAxis,
  formatTooltip,
}) => {
  return (
    <div className="bg-white rounded-xl border border-slate-200 p-4">
      {title && <h3 className="text-lg font-semibold text-slate-800 mb-4">{title}</h3>}
      <ResponsiveContainer width="100%" height={height}>
        <LineChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
          {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="#E2E8F0" />}
          <XAxis
            dataKey={xAxisKey}
            tickFormatter={formatXAxis}
            tick={{ fontSize: 12, fill: '#64748B' }}
            axisLine={{ stroke: '#E2E8F0' }}
          />
          <YAxis
            tick={{ fontSize: 12, fill: '#64748B' }}
            axisLine={{ stroke: '#E2E8F0' }}
          />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1E293B',
              border: 'none',
              borderRadius: '8px',
              color: '#fff',
            }}
            formatter={formatTooltip}
          />
          {showLegend && <Legend />}
          {dataKeys.map((dk, index) => (
            <Line
              key={dk.key}
              type="monotone"
              dataKey={dk.key}
              name={dk.name}
              stroke={dk.color || CHART_COLORS[index % CHART_COLORS.length]}
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 6, fill: dk.color || CHART_COLORS[index % CHART_COLORS.length] }}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
};

// Area Chart for stacked metrics
interface StackedAreaChartProps {
  data: any[];
  title?: string;
  xAxisKey?: string;
  dataKeys: { key: string; name: string; stackId?: string }[];
  height?: number;
  colors?: string[];
}

export const StackedAreaChart: React.FC<StackedAreaChartProps> = ({
  data,
  title,
  xAxisKey = 'date',
  dataKeys,
  height = 300,
  colors = CHART_COLORS,
}) => {
  return (
    <div className="bg-white rounded-xl border border-slate-200 p-4">
      {title && <h3 className="text-lg font-semibold text-slate-800 mb-4">{title}</h3>}
      <ResponsiveContainer width="100%" height={height}>
        <AreaChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
          <defs>
            {dataKeys.map((dk, index) => (
              <linearGradient key={dk.key} id={`gradient-${dk.key}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor={colors[index % colors.length]} stopOpacity={0.3} />
                <stop offset="95%" stopColor={colors[index % colors.length]} stopOpacity={0} />
              </linearGradient>
            ))}
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="#E2E8F0" />
          <XAxis dataKey={xAxisKey} tick={{ fontSize: 12, fill: '#64748B' }} />
          <YAxis tick={{ fontSize: 12, fill: '#64748B' }} />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1E293B',
              border: 'none',
              borderRadius: '8px',
              color: '#fff',
            }}
          />
          <Legend />
          {dataKeys.map((dk, index) => (
            <Area
              key={dk.key}
              type="monotone"
              dataKey={dk.key}
              name={dk.name}
              stackId={dk.stackId || 'stack'}
              stroke={colors[index % colors.length]}
              fill={`url(#gradient-${dk.key})`}
              fillOpacity={1}
            />
          ))}
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
};

// Bar Chart for comparisons
interface ComparisonBarChartProps {
  data: { name: string; value: number; comparison?: number }[];
  title?: string;
  height?: number;
  showComparison?: boolean;
  colors?: { main?: string; comparison?: string };
}

export const ComparisonBarChart: React.FC<ComparisonBarChartProps> = ({
  data,
  title,
  height = 300,
  showComparison = false,
  colors = {},
}) => {
  const chartData = useMemo(() => {
    return data.map((item) => ({
      ...item,
      mainValue: item.value,
      comparisonValue: item.comparison || 0,
    }));
  }, [data]);

  return (
    <div className="bg-white rounded-xl border border-slate-200 p-4">
      {title && <h3 className="text-lg font-semibold text-slate-800 mb-4">{title}</h3>}
      <ResponsiveContainer width="100%" height={height}>
        <BarChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#E2E8F0" />
          <XAxis dataKey="name" tick={{ fontSize: 12, fill: '#64748B' }} />
          <YAxis tick={{ fontSize: 12, fill: '#64748B' }} />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1E293B',
              border: 'none',
              borderRadius: '8px',
              color: '#fff',
            }}
          />
          <Legend />
          <Bar dataKey="mainValue" name="Current" fill={colors.main || CHART_COLORS[0]} radius={[4, 4, 0, 0]} />
          {showComparison && (
            <Bar dataKey="comparisonValue" name="Previous" fill={colors.comparison || CHART_COLORS[1]} radius={[4, 4, 0, 0]} />
          )}
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
};

// Pie/Donut Chart for distributions
interface DistributionPieChartProps {
  data: { name: string; value: number }[];
  title?: string;
  height?: number;
  innerRadius?: number;
  colors?: string[];
}

export const DistributionPieChart: React.FC<DistributionPieChartProps> = ({
  data,
  title,
  height = 300,
  innerRadius = 60,
  colors = CHART_COLORS,
}) => {
  return (
    <div className="bg-white rounded-xl border border-slate-200 p-4">
      {title && <h3 className="text-lg font-semibold text-slate-800 mb-4">{title}</h3>}
      <ResponsiveContainer width="100%" height={height}>
        <PieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            innerRadius={innerRadius}
            outerRadius={100}
            paddingAngle={2}
            dataKey="value"
            label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
            labelLine={false}
          >
            {data.map((entry, index) => (
              <Cell key={`cell-${index}`} fill={colors[index % colors.length]} />
            ))}
          </Pie>
          <Tooltip
            contentStyle={{
              backgroundColor: '#1E293B',
              border: 'none',
              borderRadius: '8px',
              color: '#fff',
            }}
          />
          <Legend />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
};

// Radar Chart for multi-dimensional comparison
interface RadarChartProps {
  data: { subject: string; A: number; B: number; fullMark: number }[];
  title?: string;
  height?: number;
  colors?: string[];
}

export const MultiRadarChart: React.FC<RadarChartProps> = ({
  data,
  title,
  height = 400,
  colors = [CHART_COLORS[0], CHART_COLORS[1]],
}) => {
  return (
    <div className="bg-white rounded-xl border border-slate-200 p-4">
      {title && <h3 className="text-lg font-semibold text-slate-800 mb-4">{title}</h3>}
      <ResponsiveContainer width="100%" height={height}>
        <RadarChart cx="50%" cy="50%" outerRadius="80%" data={data}>
          <PolarGrid stroke="#E2E8F0" />
          <PolarAngleAxis dataKey="subject" tick={{ fontSize: 12, fill: '#64748B' }} />
          <PolarRadiusAxis angle={30} domain={[0, 'auto']} tick={{ fontSize: 10, fill: '#64748B' }} />
          <Radar
            name="Current"
            dataKey="A"
            stroke={colors[0]}
            fill={colors[0]}
            fillOpacity={0.3}
            strokeWidth={2}
          />
          <Radar
            name="Target"
            dataKey="B"
            stroke={colors[1]}
            fill={colors[1]}
            fillOpacity={0.3}
            strokeWidth={2}
          />
          <Legend />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1E293B',
              border: 'none',
              borderRadius: '8px',
              color: '#fff',
            }}
          />
        </RadarChart>
      </ResponsiveContainer>
    </div>
  );
};

// Sector Performance Chart
interface SectorPerformanceChartProps {
  sectors: SectorPerformance[];
  height?: number;
  metric?: 'growthRate' | 'contributionToGDP' | 'employmentShare';
}

export const SectorPerformanceChart: React.FC<SectorPerformanceChartProps> = ({
  sectors,
  height = 300,
  metric = 'growthRate',
}) => {
  const data = useMemo(() => {
    return sectors.map((sector) => ({
      name: sector.sectorName,
      value: sector[metric] as number,
      fullMark: Math.max(...sectors.map((s) => s[metric] as number)) * 1.2,
    }));
  }, [sectors, metric]);

  return (
    <ResponsiveContainer width="100%" height={height}>
      <BarChart data={data} layout="vertical" margin={{ top: 5, right: 30, left: 80, bottom: 5 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="#E2E8F0" horizontal={true} vertical={false} />
        <XAxis type="number" tick={{ fontSize: 12, fill: '#64748B' }} />
        <YAxis dataKey="name" type="category" tick={{ fontSize: 12, fill: '#64748B' }} />
        <Tooltip
          contentStyle={{
            backgroundColor: '#1E293B',
            border: 'none',
            borderRadius: '8px',
            color: '#fff',
          }}
          formatter={(value: number) => `${value.toFixed(2)}%`}
        />
        <Bar dataKey="value" name={metric === 'growthRate' ? 'Growth Rate' : metric} fill={CHART_COLORS[0]} radius={[0, 4, 4, 0]} />
      </BarChart>
    </ResponsiveContainer>
  );
};

export default {
  TrendLineChart,
  StackedAreaChart,
  ComparisonBarChart,
  DistributionPieChart,
  MultiRadarChart,
  SectorPerformanceChart,
};
