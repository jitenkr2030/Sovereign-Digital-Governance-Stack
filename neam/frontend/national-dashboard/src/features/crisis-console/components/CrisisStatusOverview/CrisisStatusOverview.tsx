import React from 'react';
import { AlertTriangle, CheckCircle, Clock, Users, Signal, Activity } from 'lucide-react';
import { Card } from '../../ui/Card';
import { Badge } from '../../ui/Badge';
import { ProgressBar } from '../../ui/ProgressBar';
import './CrisisStatusOverview.css';

export interface CrisisMetrics {
  alertCount: number;
  activeInterventions: number;
  affectedRegions: number;
  systemStatus: 'operational' | 'degraded' | 'outage';
  responseTime: number; // average in minutes
  lastIncident?: Date;
}

export interface CrisisStatusOverviewProps {
  /** Current crisis level */
  crisisLevel: 1 | 2 | 3 | 4 | 5;
  /** Crisis metrics */
  metrics: CrisisMetrics;
  /** Active team members count */
  activeTeamMembers?: number;
  /** On duty commander */
  commanderOnDuty?: string;
  /** Elapsed time since crisis started */
  elapsedTime?: string;
  /** Callback when view details is clicked */
  onViewDetails?: () => void;
  /** Custom class name */
  className?: string;
}

/**
 * Crisis level configuration
 */
const crisisLevelConfig: Record<number, { 
  label: string; 
  color: string; 
  bgColor: string; 
  description: string;
}> = {
  1: {
    label: 'Monitor',
    color: 'var(--color-info)',
    bgColor: 'var(--color-info-bg)',
    description: 'Enhanced monitoring active',
  },
  2: {
    label: 'Elevated',
    color: '#F97316',
    bgColor: 'rgba(249, 115, 22, 0.1)',
    description: 'Potential threats identified',
  },
  3: {
    label: 'High Alert',
    color: 'var(--color-warning)',
    bgColor: 'var(--color-warning-bg)',
    description: 'Active threat management',
  },
  4: {
    label: 'Severe',
    color: '#DC2626',
    bgColor: 'rgba(220, 38, 38, 0.1)',
    description: 'Significant impact detected',
  },
  5: {
    label: 'Critical',
    color: 'var(--color-danger)',
    bgColor: 'var(--color-danger-bg)',
    description: 'Maximum response activated',
  },
};

/**
 * System status configuration
 */
const systemStatusConfig: Record<CrisisMetrics['systemStatus'], { 
  label: string; 
  color: string; 
  icon: React.ReactNode;
}> = {
  operational: {
    label: 'Operational',
    color: 'var(--color-success)',
    icon: <CheckCircle size={14} />,
  },
  degraded: {
    label: 'Degraded',
    color: 'var(--color-warning)',
    icon: <Activity size={14} />,
  },
  outage: {
    label: 'Outage',
    color: 'var(--color-danger)',
    icon: <AlertTriangle size={14} />,
  },
};

/**
 * CrisisStatusOverview Component
 * 
 * High-level overview of the current crisis status with key metrics.
 */
export const CrisisStatusOverview: React.FC<CrisisStatusOverviewProps> = ({
  crisisLevel,
  metrics,
  activeTeamMembers = 0,
  commanderOnDuty,
  elapsedTime,
  onViewDetails,
  className = '',
}) => {
  const levelConfig = crisisLevelConfig[crisisLevel];

  return (
    <Card className={`crisis-status-overview ${className}`}>
      {/* Header with Crisis Level */}
      <div className="crisis-overview-header" style={{ backgroundColor: levelConfig.bgColor }}>
        <div className="crisis-level-indicator">
          <span className="crisis-level-number">{crisisLevel}</span>
          <div className="crisis-level-info">
            <h3 className="crisis-level-label" style={{ color: levelConfig.color }}>
              Level {crisisLevel}: {levelConfig.label}
            </h3>
            <p className="crisis-level-description">{levelConfig.description}</p>
          </div>
        </div>

        {elapsedTime && (
          <div className="crisis-elapsed">
            <Clock size={14} />
            <span>{elapsedTime}</span>
          </div>
        )}
      </div>

      {/* Key Metrics Grid */}
      <div className="crisis-metrics-grid">
        {/* Alert Count */}
        <div className="crisis-metric-card">
          <div className="metric-icon" style={{ color: 'var(--color-warning)' }}>
            <AlertTriangle size={20} />
          </div>
          <div className="metric-content">
            <span className="metric-value">{metrics.alertCount}</span>
            <span className="metric-label">Active Alerts</span>
          </div>
        </div>

        {/* Active Interventions */}
        <div className="crisis-metric-card">
          <div className="metric-icon" style={{ color: 'var(--color-primary)' }}>
            <Activity size={20} />
          </div>
          <div className="metric-content">
            <span className="metric-value">{metrics.activeInterventions}</span>
            <span className="metric-label">Active Interventions</span>
          </div>
        </div>

        {/* Affected Regions */}
        <div className="crisis-metric-card">
          <div className="metric-icon" style={{ color: 'var(--color-danger)' }}>
            <Signal size={20} />
          </div>
          <div className="metric-content">
            <span className="metric-value">{metrics.affectedRegions}</span>
            <span className="metric-label">Affected Regions</span>
          </div>
        </div>

        {/* System Status */}
        <div className="crisis-metric-card">
          <div 
            className="metric-icon"
            style={{ color: systemStatusConfig[metrics.systemStatus].color }}
          >
            {systemStatusConfig[metrics.systemStatus].icon}
          </div>
          <div className="metric-content">
            <span 
              className="metric-value"
              style={{ color: systemStatusConfig[metrics.systemStatus].color }}
            >
              {systemStatusConfig[metrics.systemStatus].label}
            </span>
            <span className="metric-label">System Status</span>
          </div>
        </div>
      </div>

      {/* Response Time Progress */}
      <div className="crisis-response-time">
        <div className="response-time-header">
          <span>Average Response Time</span>
          <span className="response-time-value">{metrics.responseTime} min</span>
        </div>
        <ProgressBar 
          value={metrics.responseTime} 
          max={60} 
          variant={metrics.responseTime > 30 ? 'danger' : metrics.responseTime > 15 ? 'warning' : 'success'}
        />
        <span className="response-time-target">Target: &lt;15 min</span>
      </div>

      {/* Team and Commander Info */}
      <div className="crisis-team-info">
        <div className="team-member">
          <Users size={14} />
          <span>{activeTeamMembers} team members active</span>
        </div>
        {commanderOnDuty && (
          <div className="commander-info">
            <span className="commander-label">Commander on duty:</span>
            <span className="commander-name">{commanderOnDuty}</span>
          </div>
        )}
      </div>

      {/* Action Button */}
      <div className="crisis-actions">
        <button className="crisis-details-btn" onClick={onViewDetails}>
          View Full Status
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M9 18l6-6-6-6" />
          </svg>
        </button>
      </div>
    </Card>
  );
};

export default CrisisStatusOverview;
