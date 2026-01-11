import React, { useCallback } from 'react';
import './ProgressBar.css';

export interface ProgressBarProps {
  /** Progress value (0-100) */
  value: number;
  /** Maximum value */
  max?: number;
  /** Show percentage label */
  showLabel?: boolean;
  /** Size */
  size?: 'sm' | 'md' | 'lg';
  /** Color variant */
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'primary';
  /** Animated */
  animated?: boolean;
  /** Custom class name */
  className?: string;
}

/**
 * ProgressBar Component
 * 
 * Visual indicator of progress or completion status.
 */
export const ProgressBar: React.FC<ProgressBarProps> = ({
  value,
  max = 100,
  showLabel = false,
  size = 'md',
  variant = 'default',
  animated = false,
  className = '',
}) => {
  const percentage = Math.min(100, Math.max(0, (value / max) * 100));

  const getColor = () => {
    if (variant !== 'default') {
      switch (variant) {
        case 'success':
          return 'var(--color-success)';
        case 'warning':
          return 'var(--color-warning)';
        case 'danger':
          return 'var(--color-danger)';
        case 'primary':
          return 'var(--color-primary)';
      }
    }
    // Auto-color based on percentage
    if (percentage >= 80) return 'var(--color-success)';
    if (percentage >= 50) return 'var(--color-warning)';
    return 'var(--color-danger)';
  };

  return (
    <div className={`progressbar-container progressbar-${size} ${className}`}>
      <div className="progressbar-track">
        <div
          className={`progressbar-fill ${animated ? 'progressbar-animated' : ''}`}
          style={{
            width: `${percentage}%`,
            backgroundColor: getColor(),
          }}
        />
      </div>
      {showLabel && (
        <span className="progressbar-label">{Math.round(percentage)}%</span>
      )}
    </div>
  );
};

export default ProgressBar;
