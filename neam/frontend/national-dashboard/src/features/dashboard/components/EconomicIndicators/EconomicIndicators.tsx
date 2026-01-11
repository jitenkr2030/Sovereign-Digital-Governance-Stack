import React from 'react';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';
import { Card } from '../../ui/Card';
import { LoadingSpinner } from '../../ui/LoadingSpinner';
import './EconomicIndicators.css';

export interface EconomicIndicatorData {
  id: string;
  name: string;
  value: number;
  previousValue: number;
  unit: string;
  format: 'number' | 'currency' | 'percentage' | 'decimal';
  changePeriod?: string;
  status?: 'normal' | 'warning' | 'critical';
  trend?: 'up' | 'down' | 'stable';
  sparklineData?: number[];
}

export interface EconomicIndicatorsProps {
  /** Array of indicator data */
  indicators: EconomicIndicatorData[];
  /** Is data loading */
  isLoading?: boolean;
  /** Callback when indicator is clicked */
  onIndicatorClick?: (indicator: EconomicIndicatorData) => void;
  /** Layout variant */
  variant?: 'grid' | 'list' | 'compact';
  /** Maximum number of indicators to show */
  maxItems?: number;
  /** Custom class name */
  className?: string;
}

/**
 * Format a number according to the specified format type
 */
const formatValue = (value: number, format: EconomicIndicatorData['format']): string => {
  const formatter = new Intl.NumberFormat('en-US', {
    style: format === 'currency' ? 'currency' : 'decimal',
    minimumFractionDigits: format === 'percentage' ? 1 : 0,
    maximumFractionDigits: format === 'percentage' ? 2 : 2,
    currency: 'USD',
  });

  switch (format) {
    case 'percentage':
      return `${formatter.format(value)}%`;
    case 'currency':
      return formatter.format(value);
    case 'decimal':
      return formatter.format(value);
    default:
      return formatter.format(value);
  }
};

/**
 * Calculate percentage change between two values
 */
const calculateChange = (current: number, previous: number): { value: number; direction: 'up' | 'down' | 'stable' } => {
  if (previous === 0) return { value: 0, direction: 'stable' };
  const change = ((current - previous) / Math.abs(previous)) * 100;
  const direction = change > 0.5 ? 'up' : change < -0.5 ? 'down' : 'stable';
  return { value: Math.abs(change), direction };
};

/**
 * EconomicIndicators Component
 * 
 * Displays real-time economic indicators with trend visualization.
 * Shows key metrics like GDP, inflation, unemployment, etc.
 */
export const EconomicIndicators: React.FC<EconomicIndicatorsProps> = ({
  indicators,
  isLoading = false,
  onIndicatorClick,
  variant = 'grid',
  maxItems,
  className = '',
}) => {
  const displayIndicators = maxItems ? indicators.slice(0, maxItems) : indicators;

  if (isLoading) {
    return (
      <div className={`economic-indicators ${className}`}>
        <div className="economic-indicators-grid">
          {[...Array(4)].map((_, i) => (
            <Card key={i} className="indicator-card-skeleton">
              <div className="skeleton-text skeleton-title" />
              <div className="skeleton-text skeleton-value" />
              <div className="skeleton-text skeleton-trend" />
            </Card>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className={`economic-indicators ${className}`}>
      <div className={`economic-indicators-${variant}`}>
        {displayIndicators.map((indicator) => {
          const { value: changeValue, direction } = calculateChange(
            indicator.value,
            indicator.previousValue
          );

          const TrendIcon = direction === 'up' ? TrendingUp : direction === 'down' ? TrendingDown : Minus;
          const trendClass = direction === 'up' 
            ? 'indicator-trend-up' 
            : direction === 'down' 
              ? 'indicator-trend-down' 
              : 'indicator-trend-stable';

          return (
            <Card
              key={indicator.id}
              className={`indicator-card indicator-card-${variant} indicator-card-${indicator.status || 'normal'}`}
              onClick={() => onIndicatorClick?.(indicator)}
              clickable={!!onIndicatorClick}
            >
              <div className="indicator-header">
                <span className="indicator-name">{indicator.name}</span>
                {indicator.status && (
                  <span className={`indicator-status indicator-status-${indicator.status}`}>
                    {indicator.status}
                  </span>
                )}
              </div>
              
              <div className="indicator-value-container">
                <span className="indicator-value">
                  {formatValue(indicator.value, indicator.format)}
                </span>
                {indicator.unit && (
                  <span className="indicator-unit">{indicator.unit}</span>
                )}
              </div>

              <div className="indicator-footer">
                <div className={`indicator-change ${trendClass}`}>
                  <TrendIcon size={14} />
                  <span>{changeValue.toFixed(1)}%</span>
                  {indicator.changePeriod && (
                    <span className="indicator-period">{indicator.changePeriod}</span>
                  )}
                </div>
                
                {/* Sparkline if data available */}
                {indicator.sparklineData && indicator.sparklineData.length > 0 && (
                  <svg className="indicator-sparkline" viewBox="0 0 60 20">
                    <path
                      d={generateSparklinePath(indicator.sparklineData)}
                      fill="none"
                      stroke={getSparklineColor(direction)}
                      strokeWidth="1.5"
                    />
                  </svg>
                )}
              </div>
            </Card>
          );
        })}
      </div>
    </div>
  );
};

// Helper function to generate sparkline SVG path
const generateSparklinePath = (data: number[]): string => {
  if (data.length === 0) return '';
  
  const max = Math.max(...data);
  const min = Math.min(...data);
  const range = max - min || 1;
  const width = 60;
  const height = 20;
  const padding = 2;
  
  const points = data.map((value, index) => {
    const x = (index / (data.length - 1)) * (width - padding * 2) + padding;
    const y = height - ((value - min) / range) * (height - padding * 2) - padding;
    return `${x},${y}`;
  });
  
  return `M ${points.join(' L ')}`;
};

// Helper function to get sparkline color based on trend
const getSparklineColor = (direction: 'up' | 'down' | 'stable'): string => {
  switch (direction) {
    case 'up':
      return 'var(--color-success)';
    case 'down':
      return 'var(--color-danger)';
    default:
      return 'var(--color-text-muted)';
  }
};

// Compact single indicator component
export const IndicatorCard: React.FC<{
  indicator: EconomicIndicatorData;
  onClick?: () => void;
  showTrend?: boolean;
}> = ({ indicator, onClick, showTrend = true }) => {
  const { value: changeValue, direction } = calculateChange(
    indicator.value,
    indicator.previousValue
  );

  return (
    <Card
      className={`indicator-card indicator-card-compact indicator-card-${indicator.status || 'normal'}`}
      onClick={onClick}
      clickable={!!onClick}
    >
      <div className="indicator-name">{indicator.name}</div>
      <div className="indicator-value-row">
        <span className="indicator-value">
          {formatValue(indicator.value, indicator.format)}
        </span>
        {showTrend && (
          <span className={`indicator-change indicator-change-compact ${
            direction === 'up' ? 'indicator-trend-up' : 
            direction === 'down' ? 'indicator-trend-down' : 'indicator-trend-stable'
          }`}>
            {direction === 'up' ? '+' : direction === 'down' ? '-' : ''}
            {changeValue.toFixed(1)}%
          </span>
        )}
      </div>
    </Card>
  );
};

export default EconomicIndicators;
