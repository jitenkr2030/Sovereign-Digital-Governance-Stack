import React, { useState, useMemo } from 'react';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts';
import { Card } from '../../ui/Card';
import { Button } from '../../ui/Button';
import { LoadingSpinner } from '../../ui/LoadingSpinner';
import './TrendAnalysisCharts.css';

export interface TrendDataPoint {
  timestamp: Date | string;
  [key: string]: string | number | Date;
}

export interface TrendSeries {
  key: string;
  name: string;
  color: string;
  type?: 'line' | 'area' | 'bar';
}

export interface TrendAnalysisChartsProps {
  /** Data points for the chart */
  data: TrendDataPoint[];
  /** Series configuration */
  series: TrendSeries[];
  /** Chart title */
  title?: string;
  /** X-axis field */
  xAxisKey: string;
  /** Time range options */
  timeRanges?: { label: string; value: string }[];
  /** Selected time range */
  selectedRange?: string;
  /** On time range change */
  onRangeChange?: (range: string) => void;
  /** Show legend */
  showLegend?: boolean;
  /** Show grid lines */
  showGrid?: boolean;
  /** Show reference lines */
  referenceLines?: { y: number; label: string; color?: string }[];
  /** Y-axis label */
  yAxisLabel?: string;
  /** Height of the chart */
  height?: number;
  /** Is loading */
  isLoading?: boolean;
  /** Custom tooltips */
  customTooltip?: React.ReactNode;
  /** Callback when data point is clicked */
  onDataPointClick?: (point: TrendDataPoint) => void;
  /** Fill area under line */
  showArea?: boolean;
  /** Custom class name */
  className?: string;
}

/**
 * Custom tooltip for the chart
 */
const CustomTooltip: React.FC<{
  active?: boolean;
  payload?: Array<{ color: string; name: string; value: number; dataKey: string }>;
  label?: string | number;
  formatter?: (value: number) => string;
}> = ({ active, payload, label, formatter }) => {
  if (!active || !payload?.length) return null;

  return (
    <div className="chart-custom-tooltip">
      <p className="chart-tooltip-label">{label}</p>
      {payload.map((entry, index) => (
        <p key={index} className="chart-tooltip-value" style={{ color: entry.color }}>
          {entry.name}: {formatter ? formatter(entry.value) : entry.value.toLocaleString()}
        </p>
      ))}
    </div>
  );
};

/**
 * TrendAnalysisCharts Component
 * 
 * Interactive trend visualization with multiple series and time range selection.
 */
export const TrendAnalysisCharts: React.FC<TrendAnalysisChartsProps> = ({
  data,
  series,
  title,
  xAxisKey,
  timeRanges = [
    { label: '1H', value: '1h' },
    { label: '24H', value: '24h' },
    { label: '7D', value: '7d' },
    { label: '30D', value: '30d' },
    { label: '90D', value: '90d' },
  ],
  selectedRange = '24h',
  onRangeChange,
  showLegend = true,
  showGrid = true,
  referenceLines = [],
  yAxisLabel,
  height = 300,
  isLoading = false,
  customTooltip,
  onDataPointClick,
  showArea = true,
  className = '',
}) => {
  const [activeRange, setActiveRange] = useState(selectedRange);

  const handleRangeChange = (range: string) => {
    setActiveRange(range);
    onRangeChange?.(range);
  };

  // Format X-axis based on time range
  const formatXAxis = (value: string | number) => {
    const date = new Date(value);
    
    switch (activeRange) {
      case '1h':
      case '24h':
        return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
      case '7d':
      case '30d':
        return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
      case '90d':
        return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
      default:
        return date.toLocaleDateString();
    }
  };

  // Format tooltip values
  const formatValue = (value: number) => {
    if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`;
    if (value >= 1000) return `${(value / 1000).toFixed(1)}K`;
    return value.toFixed(2);
  };

  if (isLoading) {
    return (
      <Card className={`trend-analysis-charts ${className}`}>
        {title && <h3 className="chart-title">{title}</h3>}
        <div className="chart-loading">
          <LoadingSpinner size="lg" label="Loading chart data..." />
        </div>
      </Card>
    );
  }

  return (
    <Card className={`trend-analysis-charts ${className}`}>
      {/* Header */}
      {(title || timeRanges.length > 0) && (
        <div className="chart-header">
          {title && <h3 className="chart-title">{title}</h3>}
          
          {timeRanges.length > 0 && (
            <div className="chart-time-ranges">
              {timeRanges.map((range) => (
                <button
                  key={range.value}
                  className={`chart-range-btn ${activeRange === range.value ? 'chart-range-active' : ''}`}
                  onClick={() => handleRangeChange(range.value)}
                >
                  {range.label}
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Chart */}
      <div className="chart-container" style={{ height }}>
        <ResponsiveContainer width="100%" height="100%">
          {showArea ? (
            <AreaChart data={data} onClick={(e) => e?.activePayload?.[0] && onDataPointClick?.(e.activePayload[0].payload)}>
              <defs>
                {series.map((s) => (
                  <linearGradient key={s.key} id={`gradient-${s.key}`} x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={s.color} stopOpacity={0.3} />
                    <stop offset="95%" stopColor={s.color} stopOpacity={0} />
                  </linearGradient>
                ))}
              </defs>
              
              {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="var(--border-light)" />}
              <XAxis
                dataKey={xAxisKey}
                tickFormatter={formatXAxis}
                stroke="var(--color-text-muted)"
                fontSize={12}
                tickLine={false}
                axisLine={{ stroke: 'var(--border-light)' }}
              />
              <YAxis
                tickFormatter={formatValue}
                stroke="var(--color-text-muted)"
                fontSize={12}
                tickLine={false}
                axisLine={false}
                label={yAxisLabel ? { value: yAxisLabel, angle: -90, position: 'insideLeft', fill: 'var(--color-text-muted)' } : undefined}
              />
              <Tooltip
                content={customTooltip || <CustomTooltip formatter={formatValue} />}
              />
              {showLegend && <Legend />}
              
              {referenceLines.map((ref, index) => (
                <ReferenceLine
                  key={index}
                  y={ref.y}
                  label={ref.label}
                  stroke={ref.color || 'var(--color-text-muted)'}
                  strokeDasharray="5 5"
                />
              ))}
              
              {series.map((s) => (
                <Area
                  key={s.key}
                  type="monotone"
                  dataKey={s.key}
                  name={s.name}
                  stroke={s.color}
                  fill={`url(#gradient-${s.key})`}
                  strokeWidth={2}
                />
              ))}
            </AreaChart>
          ) : (
            <LineChart data={data} onClick={(e) => e?.activePayload?.[0] && onDataPointClick?.(e.activePayload[0].payload)}>
              {showGrid && <CartesianGrid strokeDasharray="3 3" stroke="var(--border-light)" />}
              <XAxis
                dataKey={xAxisKey}
                tickFormatter={formatXAxis}
                stroke="var(--color-text-muted)"
                fontSize={12}
                tickLine={false}
                axisLine={{ stroke: 'var(--border-light)' }}
              />
              <YAxis
                tickFormatter={formatValue}
                stroke="var(--color-text-muted)"
                fontSize={12}
                tickLine={false}
                axisLine={false}
              />
              <Tooltip content={<CustomTooltip formatter={formatValue} />} />
              {showLegend && <Legend />}
              
              {series.map((s) => (
                <Line
                  key={s.key}
                  type="monotone"
                  dataKey={s.key}
                  name={s.name}
                  stroke={s.color}
                  strokeWidth={2}
                  dot={false}
                  activeDot={{ r: 6, fill: s.color }}
                />
              ))}
            </LineChart>
          )}
        </ResponsiveContainer>
      </div>
    </Card>
  );
};

// Simple line chart for quick metrics
export const MetricLineChart: React.FC<{
  data: { time: string; value: number }[];
  color?: string;
  height?: number;
  showGrid?: boolean;
}> = ({ data, color = 'var(--color-primary)', height = 60, showGrid = false }) => (
  <div style={{ height }}>
    <ResponsiveContainer width="100%" height="100%">
      <AreaChart data={data}>
        <defs>
          <linearGradient id={`gradient-metric-${color}`} x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor={color} stopOpacity={0.3} />
            <stop offset="95%" stopColor={color} stopOpacity={0} />
          </linearGradient>
        </defs>
        <Area
          type="monotone"
          dataKey="value"
          stroke={color}
          fill={`url(#gradient-metric-${color})`}
          strokeWidth={1.5}
        />
      </AreaChart>
    </ResponsiveContainer>
  </div>
);

export default TrendAnalysisCharts;
