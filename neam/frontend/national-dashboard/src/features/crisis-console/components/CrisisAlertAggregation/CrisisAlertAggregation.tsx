import React, { useState, useEffect, useCallback } from 'react';
import { AlertTriangle, AlertCircle, Info, Clock, CheckCircle, X, Filter, SortAsc } from 'lucide-react';
import { Card } from '../../ui/Card';
import { Badge } from '../../ui/Badge';
import { Button } from '../../ui/Button';
import { LoadingSpinner } from '../../ui/LoadingSpinner';
import { useWebSocket } from '../../../hooks/useWebSocket';
import './CrisisAlertAggregation.css';

export type CrisisSeverity = 'critical' | 'high' | 'medium' | 'low' | 'info';
export type CrisisStatus = 'new' | 'acknowledged' | 'investigating' | 'resolved' | 'closed';

export interface CrisisAlert {
  id: string;
  title: string;
  description: string;
  severity: CrisisSeverity;
  status: CrisisStatus;
  category: string;
  source: string;
  timestamp: Date;
  affectedRegions?: string[];
  affectedSystems?: string[];
  estimatedImpact?: string;
  relatedAlerts?: string[];
  escalationLevel?: number;
  assignedTo?: string;
  timeline: {
    timestamp: Date;
    action: string;
    user?: string;
    notes?: string;
  }[];
}

export interface CrisisAlertAggregationProps {
  /** Initial alerts data */
  initialAlerts?: CrisisAlert[];
  /** WebSocket URL for real-time updates */
  wsUrl?: string;
  /** Maximum alerts to display */
  maxItems?: number;
  /** Filter by severity */
  severityFilter?: CrisisSeverity[];
  /** Filter by status */
  statusFilter?: CrisisStatus[];
  /** Callback when alert is acknowledged */
  onAcknowledge?: (alertId: string) => void;
  /** Callback when alert is escalated */
  onEscalate?: (alertId: string, level: number) => void;
  /** Callback when alert status changes */
  onStatusChange?: (alertId: string, status: CrisisStatus) => void;
  /** Callback when alert is clicked */
  onAlertClick?: (alert: CrisisAlert) => void;
  /** Callback when bulk action is performed */
  onBulkAction?: (action: string, alertIds: string[]) => void;
  /** Is loading */
  isLoading?: boolean;
  /** Enable auto-refresh */
  autoRefresh?: boolean;
  /** Custom class name */
  className?: string;
}

/**
 * Severity configuration
 */
const severityConfig: Record<CrisisSeverity, { 
  color: string; 
  bgColor: string; 
  icon: React.ReactNode;
  priority: number;
}> = {
  critical: { 
    color: 'var(--color-danger)', 
    bgColor: 'var(--color-danger-bg)',
    icon: <AlertTriangle size={16} />,
    priority: 0
  },
  high: { 
    color: '#F97316', 
    bgColor: 'rgba(249, 115, 22, 0.1)',
    icon: <AlertCircle size={16} />,
    priority: 1
  },
  medium: { 
    color: 'var(--color-warning)', 
    bgColor: 'var(--color-warning-bg)',
    icon: <Info size={16} />,
    priority: 2
  },
  low: { 
    color: 'var(--color-info)', 
    bgColor: 'var(--color-info-bg)',
    icon: <Info size={16} />,
    priority: 3
  },
  info: { 
    color: 'var(--color-text-muted)', 
    bgColor: 'rgba(148, 163, 184, 0.1)',
    icon: <Info size={16} />,
    priority: 4
  },
};

/**
 * Status configuration
 */
const statusConfig: Record<CrisisStatus, { label: string; color: string }> = {
  new: { label: 'New', color: 'var(--color-danger)' },
  acknowledged: { label: 'Acknowledged', color: 'var(--color-warning)' },
  investigating: { label: 'Investigating', color: 'var(--color-info)' },
  resolved: { label: 'Resolved', color: 'var(--color-success)' },
  closed: { label: 'Closed', color: 'var(--color-text-muted)' },
};

/**
 * CrisisAlertAggregation Component
 * 
 * Aggregates and displays crisis alerts with real-time updates and bulk actions.
 */
export const CrisisAlertAggregation: React.FC<CrisisAlertAggregationProps> = ({
  initialAlerts = [],
  wsUrl,
  maxItems = 20,
  severityFilter,
  statusFilter,
  onAcknowledge,
  onEscalate,
  onStatusChange,
  onAlertClick,
  onBulkAction,
  isLoading = false,
  autoRefresh = true,
  className = '',
}) => {
  const [alerts, setAlerts] = useState<CrisisAlert[]>(initialAlerts);
  const [selectedAlerts, setSelectedAlerts] = useState<Set<string>>(new Set());
  const [filterSeverity, setFilterSeverity] = useState<CrisisSeverity | 'all'>('all');
  const [filterStatus, setFilterStatus] = useState<CrisisStatus | 'all'>('all');
  const [sortBy, setSortBy] = useState<'severity' | 'time' | 'status'>('severity');
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  // WebSocket connection
  const { lastMessage, isOpen } = useWebSocket(wsUrl, { enabled: !!wsUrl && autoRefresh });

  // Handle real-time updates
  useEffect(() => {
    if (lastMessage) {
      try {
        const data = JSON.parse(lastMessage.data);
        if (data.type === 'CRISIS_ALERT') {
          setAlerts((prev) => [data.alert, ...prev]);
          setLastUpdate(new Date());
        } else if (data.type === 'CRISIS_UPDATE') {
          setAlerts((prev) =>
            prev.map((alert) =>
              alert.id === data.alert.id ? { ...alert, ...data.alert } : alert
            )
          );
        }
      } catch (e) {
        console.error('Failed to parse crisis alert:', e);
      }
    }
  }, [lastMessage]);

  // Filter and sort alerts
  const filteredAlerts = alerts
    .filter((alert) => {
      if (filterSeverity !== 'all' && alert.severity !== filterSeverity) return false;
      if (filterStatus !== 'all' && alert.status !== filterStatus) return false;
      if (severityFilter && !severityFilter.includes(alert.severity)) return false;
      if (statusFilter && !statusFilter.includes(alert.status)) return false;
      return true;
    })
    .sort((a, b) => {
      switch (sortBy) {
        case 'severity':
          return severityConfig[a.severity].priority - severityConfig[b.severity].priority;
        case 'status':
          return statusOrder[a.status] - statusOrder[b.status];
        case 'time':
        default:
          return new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime();
      }
    });

  const displayAlerts = filteredAlerts.slice(0, maxItems);

  // Selection handling
  const toggleSelect = (alertId: string) => {
    setSelectedAlerts((prev) => {
      const next = new Set(prev);
      if (next.has(alertId)) {
        next.delete(alertId);
      } else {
        next.add(alertId);
      }
      return next;
    });
  };

  const toggleSelectAll = () => {
    if (selectedAlerts.size === displayAlerts.length) {
      setSelectedAlerts(new Set());
    } else {
      setSelectedAlerts(new Set(displayAlerts.map((a) => a.id)));
    }
  };

  // Bulk actions
  const handleBulkAcknowledge = () => {
    selectedAlerts.forEach((id) => onAcknowledge?.(id));
    setSelectedAlerts(new Set());
  };

  const handleBulkEscalate = (level: number) => {
    selectedAlerts.forEach((id) => onEscalate?.(id, level));
    setSelectedAlerts(new Set());
  };

  // Count summary
  const criticalCount = alerts.filter((a) => a.severity === 'critical' && a.status !== 'resolved').length;
  const unacknowledgedCount = alerts.filter((a) => a.status === 'new').length;

  if (isLoading) {
    return (
      <Card className={`crisis-aggregation ${className}`}>
        <div className="crisis-aggregation-loading">
          <LoadingSpinner size="lg" label="Loading crisis alerts..." />
        </div>
      </Card>
    );
  }

  return (
    <Card className={`crisis-aggregation ${className}`}>
      {/* Header with Summary */}
      <div className="crisis-aggregation-header">
        <div className="crisis-header-left">
          <h3 className="crisis-title">Crisis Alert Aggregation</h3>
          <div className="crisis-summary">
            {criticalCount > 0 && (
              <Badge variant="danger" size="md">
                <AlertTriangle size={12} />
                {criticalCount} Critical
              </Badge>
            )}
            {unacknowledgedCount > 0 && (
              <Badge variant="warning" size="md">
                {unacknowledgedCount} New
              </Badge>
            )}
          </div>
        </div>

        <div className="crisis-header-right">
          <div className={`connection-status ${isOpen ? 'connected' : 'disconnected'}`}>
            <span className="status-dot" />
            {isOpen ? 'Live' : 'Offline'}
          </div>
        </div>
      </div>

      {/* Filters and Controls */}
      <div className="crisis-controls">
        <div className="crisis-filters">
          <select
            className="crisis-filter-select"
            value={filterSeverity}
            onChange={(e) => setFilterSeverity(e.target.value as CrisisSeverity | 'all')}
          >
            <option value="all">All Severities</option>
            <option value="critical">Critical</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>

          <select
            className="crisis-filter-select"
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value as CrisisStatus | 'all')}
          >
            <option value="all">All Status</option>
            <option value="new">New</option>
            <option value="acknowledged">Acknowledged</option>
            <option value="investigating">Investigating</option>
            <option value="resolved">Resolved</option>
          </select>

          <select
            className="crisis-filter-select"
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as 'severity' | 'time' | 'status')}
          >
            <option value="severity">Sort by Severity</option>
            <option value="time">Sort by Time</option>
            <option value="status">Sort by Status</option>
          </select>
        </div>

        {/* Bulk Actions */}
        {selectedAlerts.size > 0 && (
          <div className="crisis-bulk-actions">
            <span className="bulk-count">{selectedAlerts.size} selected</span>
            <Button variant="outline" size="sm" onClick={handleBulkAcknowledge}>
              <CheckCircle size={14} />
              Acknowledge All
            </Button>
            <Button variant="warning" size="sm" onClick={() => handleBulkEscalate(2)}>
              Escalate
            </Button>
          </div>
        )}
      </div>

      {/* Alerts List */}
      <div className="crisis-alerts-list">
        {displayAlerts.length === 0 ? (
          <div className="crisis-empty">
            <CheckCircle size={32} />
            <p>No active crisis alerts</p>
          </div>
        ) : (
          displayAlerts.map((alert) => {
            const config = severityConfig[alert.severity];
            const status = statusConfig[alert.status];

            return (
              <div
                key={alert.id}
                className={`crisis-alert-item crisis-severity-${alert.severity} ${selectedAlerts.has(alert.id) ? 'selected' : ''}`}
                onClick={() => onAlertClick?.(alert)}
              >
                {/* Selection Checkbox */}
                <div className="crisis-checkbox">
                  <input
                    type="checkbox"
                    checked={selectedAlerts.has(alert.id)}
                    onChange={() => toggleSelect(alert.id)}
                    onClick={(e) => e.stopPropagation()}
                  />
                </div>

                {/* Severity Icon */}
                <div className="crisis-severity-icon" style={{ color: config.color }}>
                  {config.icon}
                </div>

                {/* Alert Content */}
                <div className="crisis-alert-content">
                  <div className="crisis-alert-header">
                    <h4 className="crisis-alert-title">{alert.title}</h4>
                    <Badge
                      style={{ backgroundColor: status.color, color: 'white' }}
                      size="sm"
                    >
                      {status.label}
                    </Badge>
                  </div>
                  
                  <p className="crisis-alert-description">{alert.description}</p>

                  <div className="crisis-alert-meta">
                    <span className="crisis-category">{alert.category}</span>
                    <span className="crisis-source">{alert.source}</span>
                    <span className="crisis-time">
                      <Clock size={12} />
                      {new Date(alert.timestamp).toLocaleString()}
                    </span>
                    {alert.assignedTo && (
                      <span className="crisis-assignee">
                        Assigned: {alert.assignedTo}
                      </span>
                    )}
                  </div>

                  {/* Affected Items */}
                  {(alert.affectedRegions?.length || alert.affectedSystems?.length) && (
                    <div className="crisis-affected">
                      {alert.affectedRegions?.map((region) => (
                        <Badge key={region} variant="default" size="sm">
                          {region}
                        </Badge>
                      ))}
                      {alert.affectedSystems?.map((system) => (
                        <Badge key={system} variant="info" size="sm">
                          {system}
                        </Badge>
                      ))}
                    </div>
                  )}
                </div>

                {/* Quick Actions */}
                <div className="crisis-alert-actions">
                  {alert.status === 'new' && (
                    <Button
                      variant="success"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        onAcknowledge?.(alert.id);
                      }}
                    >
                      Acknowledge
                    </Button>
                  )}
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={(e) => {
                      e.stopPropagation();
                      onEscalate?.(alert.id, (alert.escalationLevel || 0) + 1);
                    }}
                  >
                    Escalate
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={(e) => {
                      e.stopPropagation();
                      onStatusChange?.(alert.id, 'investigating');
                    }}
                  >
                    Investigate
                  </Button>
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Footer */}
      <div className="crisis-footer">
        <span className="crisis-count">
          {filteredAlerts.length} of {alerts.length} alerts
        </span>
        <span className="crisis-updated">
          Last updated: {lastUpdate.toLocaleTimeString()}
        </span>
      </div>
    </Card>
  );
};

// Status order for sorting
const statusOrder: Record<CrisisStatus, number> = {
  new: 0,
  acknowledged: 1,
  investigating: 2,
  resolved: 3,
  closed: 4,
};

export default CrisisAlertAggregation;
