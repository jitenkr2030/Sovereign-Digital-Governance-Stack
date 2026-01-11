import React, { useState, useCallback, useEffect } from 'react';
import { AlertCircle, CheckCircle, XCircle, Clock, Filter, RefreshCw } from 'lucide-react';
import { Card } from '../../ui/Card';
import { Button } from '../../ui/Button';
import { Badge } from '../../ui/Badge';
import { LoadingSpinner } from '../../ui/LoadingSpinner';
import { useWebSocket } from '../../../hooks/useWebSocket';
import './AnomalyAlerts.css';

export type AlertSeverity = 'critical' | 'high' | 'medium' | 'low' | 'info';

export interface AnomalyAlert {
  id: string;
  title: string;
  description: string;
  severity: AlertSeverity;
  category: string;
  region?: string;
  timestamp: Date;
  acknowledged: boolean;
  metrics?: {
    name: string;
    value: number;
    threshold: number;
    unit: string;
  }[];
  source: string;
  relatedAlerts?: string[];
}

export interface AnomalyAlertsProps {
  /** Initial alerts data */
  initialAlerts?: AnomalyAlert[];
  /** WebSocket URL for real-time updates */
  wsUrl?: string;
  /** Filter by severity */
  severityFilter?: AlertSeverity[];
  /** Filter by acknowledged status */
  showAcknowledged?: boolean;
  /** Maximum number of alerts to display */
  maxItems?: number;
  /** Is loading */
  isLoading?: boolean;
  /** Callback when alert is acknowledged */
  onAcknowledge?: (alertId: string) => void;
  /** Callback when alert is dismissed */
  onDismiss?: (alertId: string) => void;
  /** Callback when alert is clicked */
  onAlertClick?: (alert: AnomalyAlert) => void;
  /** Auto-refresh interval in ms (0 = disabled) */
  refreshInterval?: number;
  /** Custom class name */
  className?: string;
}

/**
 * Severity configuration
 */
const severityConfig: Record<AlertSeverity, { color: string; icon: React.ReactNode; label: string }> = {
  critical: { 
    color: 'var(--color-danger)', 
    icon: <XCircle size={16} />,
    label: 'Critical'
  },
  high: { 
    color: 'var(--color-warning)', 
    icon: <AlertCircle size={16} />,
    label: 'High'
  },
  medium: { 
    color: '#F97316', // Orange
    icon: <AlertCircle size={16} />,
    label: 'Medium'
  },
  low: { 
    color: 'var(--color-info)', 
    icon: <AlertCircle size={16} />,
    label: 'Low'
  },
  info: { 
    color: 'var(--color-text-muted)', 
    icon: <CheckCircle size={16} />,
    label: 'Info'
  },
};

/**
 * Format timestamp to relative time
 */
const formatRelativeTime = (date: Date): string => {
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  return `${diffDays}d ago`;
};

/**
 * AnomalyAlerts Component
 * 
 * Displays real-time anomaly alerts with filtering, sorting, and actions.
 */
export const AnomalyAlerts: React.FC<AnomalyAlertsProps> = ({
  initialAlerts = [],
  wsUrl,
  severityFilter,
  showAcknowledged = false,
  maxItems,
  isLoading = false,
  onAcknowledge,
  onDismiss,
  onAlertClick,
  refreshInterval = 0,
  className = '',
}) => {
  const [alerts, setAlerts] = useState<AnomalyAlert[]>(initialAlerts);
  const [filter, setFilter] = useState<AlertSeverity | 'all'>('all');
  const [sortBy, setSortBy] = useState<'severity' | 'time'>('severity');
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  // WebSocket connection for real-time updates
  const { lastMessage } = useWebSocket(wsUrl, {
    enabled: !!wsUrl,
    reconnectInterval: 5000,
  });

  // Handle incoming WebSocket messages
  useEffect(() => {
    if (lastMessage) {
      try {
        const data = JSON.parse(lastMessage.data);
        if (data.type === 'NEW_ALERT') {
          setAlerts((prev) => [data.alert, ...prev]);
          setLastUpdate(new Date());
        } else if (data.type === 'ALERT_UPDATED') {
          setAlerts((prev) =>
            prev.map((alert) =>
              alert.id === data.alert.id ? { ...alert, ...data.alert } : alert
            )
          );
        }
      } catch (e) {
        console.error('Failed to parse alert message:', e);
      }
    }
  }, [lastMessage]);

  // Filter and sort alerts
  const filteredAlerts = alerts
    .filter((alert) => {
      if (!showAcknowledged && alert.acknowledged) return false;
      if (filter !== 'all' && alert.severity !== filter) return false;
      if (severityFilter && !severityFilter.includes(alert.severity)) return false;
      return true;
    })
    .sort((a, b) => {
      if (sortBy === 'severity') {
        const severityOrder = { critical: 0, high: 1, medium: 2, low: 3, info: 4 };
        return severityOrder[a.severity] - severityOrder[b.severity];
      }
      return new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime();
    });

  const displayAlerts = maxItems ? filteredAlerts.slice(0, maxItems) : filteredAlerts;

  const handleAcknowledge = useCallback((alertId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setAlerts((prev) =>
      prev.map((alert) =>
        alert.id === alertId ? { ...alert, acknowledged: true } : alert
      )
    );
    onAcknowledge?.(alertId);
  }, [onAcknowledge]);

  const handleDismiss = useCallback((alertId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    setAlerts((prev) => prev.filter((alert) => alert.id !== alertId));
    onDismiss?.(alertId);
  }, [onDismiss]);

  if (isLoading) {
    return (
      <Card className={`anomaly-alerts ${className}`}>
        <div className="anomaly-alerts-header">
          <h3 className="anomaly-alerts-title">Anomaly Alerts</h3>
        </div>
        <div className="anomaly-alerts-loading">
          <LoadingSpinner size="lg" label="Loading alerts..." />
        </div>
      </Card>
    );
  }

  return (
    <Card className={`anomaly-alerts ${className}`}>
      <div className="anomaly-alerts-header">
        <div className="anomaly-alerts-title-row">
          <h3 className="anomaly-alerts-title">Anomaly Alerts</h3>
          <div className="live-indicator">Live</div>
        </div>
        
        <div className="anomaly-alerts-controls">
          {/* Severity Filter */}
          <select
            className="anomaly-filter-select"
            value={filter}
            onChange={(e) => setFilter(e.target.value as AlertSeverity | 'all')}
          >
            <option value="all">All Severities</option>
            <option value="critical">Critical</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
            <option value="info">Info</option>
          </select>

          {/* Sort */}
          <select
            className="anomaly-filter-select"
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as 'severity' | 'time')}
          >
            <option value="severity">Sort by Severity</option>
            <option value="time">Sort by Time</option>
          </select>
        </div>
      </div>

      {/* Alert Count */}
      <div className="anomaly-alerts-count">
        {filteredAlerts.length} alert{filteredAlerts.length !== 1 ? 's' : ''}
        {lastUpdate && (
          <span className="anomaly-last-update">
            Updated {formatRelativeTime(lastUpdate)}
          </span>
        )}
      </div>

      {/* Alerts List */}
      <div className="anomaly-alerts-list">
        {displayAlerts.length === 0 ? (
          <div className="anomaly-alerts-empty">
            <CheckCircle size={32} />
            <p>No active alerts</p>
          </div>
        ) : (
          displayAlerts.map((alert) => {
            const config = severityConfig[alert.severity];
            
            return (
              <div
                key={alert.id}
                className={`anomaly-alert-item ${
                  alert.acknowledged ? 'anomaly-alert-acknowledged' : ''
                } anomaly-alert-${alert.severity}`}
                onClick={() => onAlertClick?.(alert)}
              >
                <div className="anomaly-alert-icon" style={{ color: config.color }}>
                  {config.icon}
                </div>
                
                <div className="anomaly-alert-content">
                  <div className="anomaly-alert-header">
                    <span className="anomaly-alert-title">{alert.title}</span>
                    <Badge 
                      variant={alert.severity === 'critical' ? 'danger' : alert.severity === 'high' ? 'warning' : 'default'}
                      size="sm"
                    >
                      {config.label}
                    </Badge>
                  </div>
                  
                  <p className="anomaly-alert-description">{alert.description}</p>
                  
                  <div className="anomaly-alert-meta">
                    <span className="anomaly-alert-category">{alert.category}</span>
                    {alert.region && (
                      <span className="anomaly-alert-region">{alert.region}</span>
                    )}
                    <span className="anomaly-alert-time">
                      <Clock size={12} />
                      {formatRelativeTime(new Date(alert.timestamp))}
                    </span>
                  </div>

                  {/* Metrics Preview */}
                  {alert.metrics && alert.metrics.length > 0 && (
                    <div className="anomaly-alert-metrics">
                      {alert.metrics.slice(0, 3).map((metric, idx) => (
                        <div key={idx} className="anomaly-metric">
                          <span className="anomaly-metric-name">{metric.name}:</span>
                          <span className="anomaly-metric-value">
                            {metric.value.toFixed(1)} {metric.unit}
                          </span>
                          <span className="anomaly-metric-threshold">
                            (threshold: {metric.threshold})
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                <div className="anomaly-alert-actions">
                  {!alert.acknowledged && onAcknowledge && (
                    <button
                      type="button"
                      className="anomaly-action-btn"
                      onClick={(e) => handleAcknowledge(alert.id, e)}
                      title="Acknowledge"
                    >
                      <CheckCircle size={16} />
                    </button>
                  )}
                  {onDismiss && (
                    <button
                      type="button"
                      className="anomaly-action-btn"
                      onClick={(e) => handleDismiss(alert.id, e)}
                      title="Dismiss"
                    >
                      <XCircle size={16} />
                    </button>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Footer */}
      {filteredAlerts.length > maxItems && (
        <div className="anomaly-alerts-footer">
          <Button variant="ghost" size="sm">
            View all {filteredAlerts.length} alerts
          </Button>
        </div>
      )}
    </Card>
  );
};

export default AnomalyAlerts;
