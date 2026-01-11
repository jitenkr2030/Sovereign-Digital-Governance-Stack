/**
 * NEAM Dashboard - MetricCard Component
 * Displays key economic indicators with trends
 */

import React from 'react';
import { LineChart, Line, ResponsiveContainer } from 'recharts';
import { TrendingUp, TrendingDown, Minus, AlertTriangle, CheckCircle } from 'lucide-react';
import type { MetricCard as MetricCardType } from '../../types';
import clsx from 'clsx';

interface MetricCardProps {
  metric: MetricCardType;
  onClick?: () => void;
  className?: string;
  showTrend?: boolean;
  compact?: boolean;
}

export const MetricCard: React.FC<MetricCardProps> = ({
  metric,
  onClick,
  className,
  showTrend = true,
  compact = false,
}) => {
  const isPositive = metric.changeType === 'increase';
  const isNegative = metric.changeType === 'decrease';
  const isStable = metric.changeType === 'neutral';

  const getStatusColor = () => {
    switch (metric.status) {
      case 'critical':
        return 'text-red-600 bg-red-50 border-red-200';
      case 'warning':
        return 'text-amber-600 bg-amber-50 border-amber-200';
      default:
        return 'text-emerald-600 bg-emerald-50 border-emerald-200';
    }
  };

  const formatValue = (value: number, unit: string) => {
    if (unit === '%') {
      return `${value.toFixed(2)}%`;
    }
    if (unit === '₹' || unit === '₹ Cr') {
      if (value >= 100000) {
        return `₹${(value / 100000).toFixed(1)}L Cr`;
      }
      return `₹${value.toFixed(1)} Cr`;
    }
    if (Math.abs(value) >= 1000000) {
      return `${(value / 1000000).toFixed(1)}M`;
    }
    if (Math.abs(value) >= 1000) {
      return `${(value / 1000).toFixed(1)}K`;
    }
    return value.toFixed(2);
  };

  const formatChange = (change: number) => {
    const sign = change >= 0 ? '+' : '';
    return `${sign}${change.toFixed(2)}%`;
  };

  return (
    <div
      className={clsx(
        'bg-white rounded-xl border border-slate-200 p-4 transition-all duration-200',
        'hover:shadow-lg hover:border-slate-300 cursor-pointer',
        getStatusColor(),
        className
      )}
      onClick={onClick}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <p className={clsx(
            'text-slate-500 font-medium',
            compact ? 'text-xs' : 'text-sm'
          )}>
            {metric.name}
          </p>
          <div className="mt-2">
            <span className={clsx(
              'font-bold text-slate-800',
              compact ? 'text-xl' : 'text-2xl'
            )}>
              {formatValue(metric.value, metric.unit)}
            </span>
            <span className="text-slate-400 text-sm ml-1">
              {metric.unit}
            </span>
          </div>
          {showTrend && !compact && (
            <div className="mt-3 flex items-center gap-2">
              {isPositive && <TrendingUp className="w-4 h-4 text-emerald-500" />}
              {isNegative && <TrendingDown className="w-4 h-4 text-red-500" />}
              {isStable && <Minus className="w-4 h-4 text-slate-400" />}
              <span className={clsx(
                'text-sm font-medium',
                isPositive ? 'text-emerald-600' : isNegative ? 'text-red-600' : 'text-slate-500'
              )}>
                {formatChange(metric.change)}
              </span>
              <span className="text-slate-400 text-xs">
                vs last period
              </span>
            </div>
          )}
        </div>
        
        {metric.status !== 'normal' && (
          <div className={clsx(
            'p-2 rounded-lg',
            metric.status === 'critical' ? 'bg-red-100' : 'bg-amber-100'
          )}>
            <AlertTriangle className={clsx(
              'w-5 h-5',
              metric.status === 'critical' ? 'text-red-600' : 'text-amber-600'
            )} />
          </div>
        )}
        
        {metric.status === 'normal' && (
          <div className="p-2 rounded-lg bg-emerald-100">
            <CheckCircle className="w-5 h-5 text-emerald-600" />
          </div>
        )}
      </div>

      {showTrend && metric.trend && metric.trend.length > 0 && !compact && (
        <div className="mt-4 h-12">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={metric.trend.slice(-14)}>
              <Line
                type="monotone"
                dataKey="value"
                stroke={isPositive ? '#10B981' : isNegative ? '#EF4444' : '#64748B'}
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
    </div>
  );
};

// Compact version for grid layouts
export const MetricCardCompact: React.FC<MetricCardProps> = (props) => (
  <MetricCard {...props} compact showTrend={false} />
);

// Loading skeleton
export const MetricCardSkeleton: React.FC<{ className?: string }> = ({ className }) => (
  <div className={clsx('bg-white rounded-xl border border-slate-200 p-4 animate-pulse', className)}>
    <div className="h-3 bg-slate-200 rounded w-24 mb-3" />
    <div className="h-8 bg-slate-200 rounded w-32 mb-2" />
    <div className="h-4 bg-slate-200 rounded w-16" />
  </div>
);

export default MetricCard;
