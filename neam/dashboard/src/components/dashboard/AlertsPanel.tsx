/**
 * NEAM Dashboard - Alerts Panel Component
 * Display and manage economic alerts
 */

import React, { useState } from 'react';
import { useAppSelector, useAppDispatch } from '../../store';
import { acknowledgeAlert, setSelectedAlert } from '../../store/slices/alertSlice';
import {
  AlertTriangle,
  AlertCircle,
  CheckCircle,
  Clock,
  Filter,
  Search,
  ChevronRight,
  X,
} from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import type { Alert, AlertType, AlertPriority } from '../../types';
import clsx from 'clsx';

// Priority icons and colors
const PRIORITY_CONFIG = {
  critical: { icon: AlertTriangle, color: 'text-red-600 bg-red-50 border-red-200' },
  high: { icon: AlertCircle, color: 'text-orange-600 bg-orange-50 border-orange-200' },
  medium: { icon: Clock, color: 'text-amber-600 bg-amber-50 border-amber-200' },
  low: { icon: CheckCircle, color: 'text-blue-600 bg-blue-50 border-blue-200' },
};

// Alert type display names
const ALERT_TYPE_NAMES: Record<AlertType, string> = {
  INFLATION: 'Inflation',
  EMPLOYMENT: 'Employment',
  AGRICULTURE: 'Agriculture',
  TRADE: 'Trade',
  ENERGY: 'Energy',
  FINANCIAL: 'Financial',
  EMERGENCY: 'Emergency',
  THRESHOLD_BREACH: 'Threshold Breach',
  TREND_REVERSAL: 'Trend Reversal',
  DATA_ANOMALY: 'Data Anomaly',
  REGIONAL_DISPARITY: 'Regional Disparity',
};

interface AlertCardProps {
  alert: Alert;
  onClick?: () => void;
  onAcknowledge?: () => void;
}

const AlertCard: React.FC<AlertCardProps> = ({ alert, onClick, onAcknowledge }) => {
  const priority = PRIORITY_CONFIG[alert.priority as keyof typeof PRIORITY_CONFIG] || PRIORITY_CONFIG.low;
  const PriorityIcon = priority.icon;
  const isAcknowledged = !!alert.acknowledgedAt;

  return (
    <div
      className={clsx(
        'p-4 rounded-lg border transition-all cursor-pointer hover:shadow-md',
        priority.color,
        isAcknowledged && 'opacity-60'
      )}
      onClick={onClick}
    >
      <div className="flex items-start gap-3">
        <PriorityIcon className="w-5 h-5 flex-shrink-0 mt-0.5" />
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-2">
            <div>
              <p className="font-medium text-slate-800">{alert.title}</p>
              <p className="text-sm text-slate-600 mt-1 line-clamp-2">{alert.message}</p>
            </div>
            {!isAcknowledged && onAcknowledge && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onAcknowledge();
                }}
                className="flex-shrink-0 p-1.5 bg-white/50 rounded-lg hover:bg-white transition-colors"
              >
                <CheckCircle className="w-4 h-4" />
              </button>
            )}
          </div>
          <div className="flex items-center gap-3 mt-3">
            <span className="text-xs font-medium px-2 py-0.5 bg-white/50 rounded-full">
              {ALERT_TYPE_NAMES[alert.type] || alert.type}
            </span>
            {alert.regionName && (
              <span className="text-xs text-slate-500">{alert.regionName}</span>
            )}
            <span className="text-xs text-slate-400">
              {formatDistanceToNow(new Date(alert.createdAt), { addSuffix: true })}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};

interface AlertsPanelProps {
  maxItems?: number;
  showFilters?: boolean;
  compact?: boolean;
}

export const AlertsPanel: React.FC<AlertsPanelProps> = ({
  maxItems,
  showFilters = true,
  compact = false,
}) => {
  const dispatch = useAppDispatch();
  const { alerts, filters, unreadCount } = useAppSelector((state) => state.alerts);
  const [searchTerm, setSearchTerm] = useState('');
  const [showFilterPanel, setShowFilterPanel] = useState(false);

  // Filter alerts
  const filteredAlerts = useState(() => {
    let result = [...alerts];

    // Search filter
    if (searchTerm) {
      result = result.filter(
        (alert) =>
          alert.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
          alert.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
          alert.regionName?.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Type filter
    if (filters.types && filters.types.length > 0) {
      result = result.filter((alert) => filters.types?.includes(alert.type as AlertType));
    }

    // Priority filter
    if (filters.priority && filters.priority.length > 0) {
      result = result.filter((alert) =>
        filters.priority?.includes(alert.priority)
      );
    }

    // Sort by priority and date
    const priorityOrder = { critical: 0, high: 1, medium: 2, low: 3 };
    result.sort((a, b) => {
      const priorityDiff = priorityOrder[a.priority as keyof typeof priorityOrder] - 
                          priorityOrder[b.priority as keyof typeof priorityOrder];
      if (priorityDiff !== 0) return priorityDiff;
      return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
    });

    return maxItems ? result.slice(0, maxItems) : result;
  })[0];

  const handleAcknowledge = (alertId: string) => {
    dispatch(acknowledgeAlert(alertId));
  };

  const handleAlertClick = (alert: Alert) => {
    dispatch(setSelectedAlert(alert));
  };

  if (compact) {
    return (
      <div className="space-y-3">
        {filteredAlerts.slice(0, 5).map((alert) => (
          <AlertCard
            key={alert.id}
            alert={alert}
            onAcknowledge={() => handleAcknowledge(alert.id)}
          />
        ))}
        {filteredAlerts.length === 0 && (
          <p className="text-center text-slate-500 py-4">No alerts</p>
        )}
      </div>
    );
  }

  return (
    <div className="bg-white rounded-xl border border-slate-200">
      {/* Header */}
      <div className="p-4 border-b border-slate-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <h3 className="text-lg font-semibold text-slate-800">Alerts</h3>
            {unreadCount > 0 && (
              <span className="px-2 py-0.5 bg-red-100 text-red-600 text-xs font-medium rounded-full">
                {unreadCount} new
              </span>
            )}
          </div>
          {showFilters && (
            <button
              onClick={() => setShowFilterPanel(!showFilterPanel)}
              className={clsx(
                'p-2 rounded-lg transition-colors',
                showFilterPanel ? 'bg-slate-100' : 'hover:bg-slate-50'
              )}
            >
              <Filter className="w-4 h-4 text-slate-600" />
            </button>
          )}
        </div>

        {/* Search */}
        {showFilters && (
          <div className="mt-3 relative">
            <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
            <input
              type="text"
              placeholder="Search alerts..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        )}
      </div>

      {/* Filter Panel */}
      {showFilterPanel && showFilters && (
        <div className="p-4 bg-slate-50 border-b border-slate-200">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">
                Alert Type
              </label>
              <select
                multiple
                className="w-full p-2 border border-slate-200 rounded-lg text-sm"
                value={filters.types || []}
                onChange={(e) => {
                  const selected = Array.from(e.target.selectedOptions, (opt) => opt.value as AlertType);
                  dispatch(setFilters({ types: selected.length > 0 ? selected : undefined }));
                }}
              >
                {Object.entries(ALERT_TYPE_NAMES).map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">
                Priority
              </label>
              <select
                className="w-full p-2 border border-slate-200 rounded-lg text-sm"
                value={filters.priority?.[0] || ''}
                onChange={(e) =>
                  dispatch(setFilters({
                    priority: e.target.value ? [e.target.value as AlertPriority] : undefined
                  }))
                }
              >
                <option value="">All Priorities</option>
                <option value="critical">Critical</option>
                <option value="high">High</option>
                <option value="medium">Medium</option>
                <option value="low">Low</option>
              </select>
            </div>
          </div>
        </div>
      )}

      {/* Alert List */}
      <div className="p-4 space-y-3 max-h-96 overflow-y-auto">
        {filteredAlerts.length > 0 ? (
          filteredAlerts.map((alert) => (
            <AlertCard
              key={alert.id}
              alert={alert}
              onClick={() => handleAlertClick(alert)}
              onAcknowledge={() => handleAcknowledge(alert.id)}
            />
          ))
        ) : (
          <div className="text-center py-8">
            <CheckCircle className="w-12 h-12 text-slate-300 mx-auto mb-3" />
            <p className="text-slate-500">No alerts to display</p>
          </div>
        )}
      </div>

      {/* Footer */}
      {filteredAlerts.length > 0 && maxItems && alerts.length > maxItems && (
        <div className="p-4 border-t border-slate-200">
          <button className="w-full flex items-center justify-center gap-2 text-sm text-blue-600 hover:text-blue-700 font-medium">
            View all {alerts.length} alerts
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
      )}
    </div>
  );
};

// Mini Alerts Widget for sidebar
export const MiniAlertsWidget: React.FC = () => {
  const { alerts, unreadCount } = useAppSelector((state) => state.alerts);
  const criticalAlerts = alerts.filter((a) => a.priority === 'critical' && !a.acknowledgedAt);

  return (
    <div className="p-3 bg-white rounded-lg border border-slate-200">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-medium text-slate-700">Alerts</span>
        {unreadCount > 0 && (
          <span className="px-1.5 py-0.5 bg-red-100 text-red-600 text-xs font-medium rounded">
            {unreadCount}
          </span>
        )}
      </div>
      {criticalAlerts.length > 0 ? (
        <div className="space-y-2">
          {criticalAlerts.slice(0, 3).map((alert) => (
            <div key={alert.id} className="flex items-center gap-2 text-sm">
              <div className="w-2 h-2 bg-red-500 rounded-full" />
              <span className="text-slate-600 truncate flex-1">{alert.title}</span>
            </div>
          ))}
        </div>
      ) : (
        <p className="text-xs text-slate-400">All clear</p>
      )}
    </div>
  );
};

export default AlertsPanel;
