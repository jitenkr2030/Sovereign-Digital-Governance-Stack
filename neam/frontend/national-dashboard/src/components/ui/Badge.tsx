import React from 'react';
import './Badge.css';

export type BadgeVariant = 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info';
export type BadgeSize = 'sm' | 'md' | 'lg';

export interface BadgeProps {
  /** Badge content */
  children: React.ReactNode;
  /** Visual variant */
  variant?: BadgeVariant;
  /** Size */
  size?: BadgeSize;
  /** Dot indicator */
  dot?: boolean;
  /** Icon */
  icon?: React.ReactNode;
  /** Clickable */
  onClick?: () => void;
  /** Custom class name */
  className?: string;
}

/**
 * Badge Component
 * 
 * A compact status indicator for labels, counts, and categorical metadata.
 */
export const Badge: React.FC<BadgeProps> = ({
  children,
  variant = 'default',
  size = 'md',
  dot = false,
  icon,
  onClick,
  className = '',
}) => {
  const classes = [
    'neam-badge',
    `neam-badge-${variant}`,
    `neam-badge-${size}`,
    onClick ? 'neam-badge-clickable' : '',
    className,
  ].filter(Boolean).join(' ');

  return (
    <span className={classes} onClick={onClick} role={onClick ? 'button' : undefined}>
      {dot && <span className="neam-badge-dot" />}
      {icon && <span className="neam-badge-icon">{icon}</span>}
      {children}
    </span>
  );
};

// Status Badge convenience component
export const StatusBadge: React.FC<{
  status: 'healthy' | 'warning' | 'critical' | 'unknown';
  label?: string;
}> = ({ status, label }) => {
  const variantMap: Record<string, BadgeVariant> = {
    healthy: 'success',
    warning: 'warning',
    critical: 'danger',
    unknown: 'default',
  };

  return (
    <Badge variant={variantMap[status]} dot>
      {label || status}
    </Badge>
  );
};

export default Badge;
