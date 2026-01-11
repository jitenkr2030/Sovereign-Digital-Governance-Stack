import React, { useState } from 'react';
import { LineChart, Line, ResponsiveContainer } from 'recharts';
import type { MetricCard as MetricCardType } from '../types';
import { ArrowUp, ArrowDown, Minus, AlertCircle, CheckCircle } from 'lucide-react';

interface MetricCardProps {
  metric: MetricCardType;
}

export const MetricCard: React.FC<MetricCardProps> = ({ metric }) => {
  const [showTrend, setShowTrend] = useState(false);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'critical': return 'border-l-red-500 bg-red-50';
      case 'warning': return 'border-l-yellow-500 bg-yellow-50';
      default: return 'border-l-green-500 bg-white';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'critical': return <AlertCircle className="h-5 w-5 text-red-500" />;
      case 'warning': return <AlertCircle className="h-5 w-5 text-yellow-500" />;
      default: return <CheckCircle className="h-5 w-5 text-green-500" />;
    }
  };

  const getChangeIcon = (changeType: string) => {
    switch (changeType) {
      case 'increase': return <ArrowUp className="h-4 w-4 text-red-500" />;
      case 'decrease': return <ArrowDown className="h-4 w-4 text-green-500" />;
      default: return <Minus className="h-4 w-4 text-gray-500" />;
    }
  };

  const formatValue = (value: number, unit: string) => {
    if (unit === '%') return `${value.toFixed(2)}%`;
    if (unit === '₹' || unit === '$') return `${unit}${value.toLocaleString()}`;
    if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`;
    if (value >= 1000) return `${(value / 1000).toFixed(1)}K`;
    return value.toFixed(2);
  };

  const getTrendData = () => {
    return metric.trend.map(point => ({
      value: point.value,
    }));
  };

  return (
    <div 
      className={`rounded-lg shadow border-l-4 p-4 ${getStatusColor(metric.status)} cursor-pointer transition-all hover:shadow-md`}
      onMouseEnter={() => setShowTrend(true)}
      onMouseLeave={() => setShowTrend(false)}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center space-x-2 mb-1">
            {getStatusIcon(metric.status)}
            <span className="text-sm font-medium text-gray-600">{metric.name}</span>
          </div>
          
          <div className="flex items-baseline space-x-2">
            <span className="text-2xl font-bold text-gray-900">
              {formatValue(metric.value, metric.unit)}
            </span>
            {metric.threshold && (
              <span className="text-xs text-gray-500">
                Threshold: {formatValue(metric.threshold.critical, metric.unit)}
              </span>
            )}
          </div>
          
          <div className="flex items-center space-x-1 mt-2">
            {getChangeIcon(metric.changeType)}
            <span className={`text-sm font-medium ${
              metric.changeType === 'increase' ? 'text-red-600' :
              metric.changeType === 'decrease' ? 'text-green-600' : 'text-gray-600'
            }`}>
              {metric.change > 0 ? '+' : ''}{metric.change.toFixed(2)}%
            </span>
            <span className="text-xs text-gray-500">vs last period</span>
          </div>
        </div>
        
        {/* Mini Trend Chart */}
        <div className={`w-16 h-10 transition-opacity ${showTrend ? 'opacity-100' : 'opacity-70'}`}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={getTrendData()}>
              <Line 
                type="monotone" 
                dataKey="value" 
                stroke={
                  metric.status === 'critical' ? '#ef4444' :
                  metric.status === 'warning' ? '#f59e0b' : '#10b981'
                }
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>
      
      {/* Progress bar for thresholds */}
      {metric.threshold && (
        <div className="mt-3">
          <div className="h-1.5 bg-gray-200 rounded-full overflow-hidden">
            <div 
              className={`h-full rounded-full transition-all ${
                metric.value >= metric.threshold.critical ? 'bg-red-500' :
                metric.value >= metric.threshold.warning ? 'bg-yellow-500' : 'bg-green-500'
              }`}
              style={{ width: `${Math.min((metric.value / metric.threshold.critical) * 100, 100)}%` }}
            />
          </div>
          <div className="flex justify-between mt-1">
            <span className="text-xs text-gray-400">0</span>
            <span className="text-xs text-gray-500">{metric.threshold.label}</span>
            <span className="text-xs text-gray-400">{formatValue(metric.threshold.critical, metric.unit)}</span>
          </div>
        </div>
      )}
    </div>
  );
};
