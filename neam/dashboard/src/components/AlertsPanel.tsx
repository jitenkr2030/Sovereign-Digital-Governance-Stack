import React from 'react';
import { useGetAlertsQuery, useAcknowledgeAlertMutation } from '../store/api/alertApi';
import type { Alert, AlertType } from '../types';
import { 
  AlertTriangle, 
  AlertCircle, 
  Flame, 
  TrendingDown,
  Clock,
  Check,
  X,
  ExternalLink
} from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

interface AlertsPanelProps {
  onClose?: () => void;
  limit?: number;
}

export const AlertsPanel: React.FC<AlertsPanelProps> = ({ onClose, limit = 10 }) => {
  const { data: alerts, isLoading, refetch } = useGetAlertsQuery({ limit });
  const [acknowledgeAlert] = useAcknowledgeAlertMutation();

  const getAlertIcon = (type: AlertType) => {
    switch (type) {
      case 'EMERGENCY': return <Flame className="h-5 w-5 text-red-500" />;
      case 'INFLATION': return <TrendingDown className="h-5 w-5 text-orange-500" />;
      case 'EMPLOYMENT': return <AlertCircle className="h-5 w-5 text-yellow-500" />;
      default: return <AlertTriangle className="h-5 w-5 text-gray-500" />;
    }
  };

  const getPriorityStyles = (priority: string) => {
    switch (priority) {
      case 'critical': return 'border-l-red-500 bg-red-50';
      case 'high': return 'border-l-orange-500 bg-orange-50';
      case 'medium': return 'border-l-yellow-500 bg-yellow-50';
      default: return 'border-l-gray-300 bg-gray-50';
    }
  };

  const getPriorityBadge = (priority: string) => {
    const styles = {
      critical: 'bg-red-600 text-white',
      high: 'bg-orange-500 text-white',
      medium: 'bg-yellow-500 text-white',
      low: 'bg-gray-400 text-white',
    };
    
    return (
      <span className={`px-2 py-0.5 rounded text-xs font-medium ${styles[priority as keyof typeof styles] || styles.low}`}>
        {priority.toUpperCase()}
      </span>
    );
  };

  const handleAcknowledge = async (alertId: string) => {
    try {
      await acknowledgeAlert({ id: alertId });
      refetch();
    } catch (error) {
      console.error('Failed to acknowledge alert:', error);
    }
  };

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow p-4">
        <div className="animate-pulse space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-16 bg-gray-200 rounded"></div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b">
        <div className="flex items-center space-x-2">
          <AlertTriangle className="h-5 w-5 text-orange-500" />
          <h3 className="font-semibold text-gray-900">Active Alerts</h3>
          {alerts && alerts.length > 0 && (
            <span className="bg-orange-100 text-orange-700 text-xs px-2 py-0.5 rounded-full">
              {alerts.length}
            </span>
          )}
        </div>
        {onClose && (
          <button 
            onClick={onClose}
            className="p-1 hover:bg-gray-100 rounded"
          >
            <X className="h-5 w-5 text-gray-400" />
          </button>
        )}
      </div>

      {/* Alert List */}
      <div className="max-h-96 overflow-y-auto">
        {alerts && alerts.length > 0 ? (
          <div className="divide-y">
            {alerts.map((alert) => (
              <AlertItem 
                key={alert.id} 
                alert={alert} 
                onAcknowledge={handleAcknowledge}
                getAlertIcon={getAlertIcon}
                getPriorityStyles={getPriorityStyles}
                getPriorityBadge={getPriorityBadge}
              />
            ))}
          </div>
        ) : (
          <div className="p-8 text-center text-gray-500">
            <Check className="h-12 w-12 mx-auto text-green-400 mb-2" />
            <p>No active alerts</p>
            <p className="text-sm">All systems operating normally</p>
          </div>
        )}
      </div>

      {/* Footer */}
      {alerts && alerts.length > 0 && (
        <div className="p-3 border-t bg-gray-50">
          <button 
            className="w-full text-center text-sm text-blue-600 hover:text-blue-800 font-medium"
          >
            View All Alerts
            <ExternalLink className="h-4 w-4 inline ml-1" />
          </button>
        </div>
      )}
    </div>
  );
};

interface AlertItemProps {
  alert: Alert;
  onAcknowledge: (id: string) => void;
  getAlertIcon: (type: AlertType) => React.ReactNode;
  getPriorityStyles: (priority: string) => string;
  getPriorityBadge: (priority: string) => React.ReactNode;
}

const AlertItem: React.FC<AlertItemProps> = ({
  alert,
  onAcknowledge,
  getAlertIcon,
  getPriorityStyles,
  getPriorityBadge,
}) => {
  return (
    <div className={`p-3 border-l-4 ${getPriorityStyles(alert.priority)} hover:bg-opacity-50 transition-colors`}>
      <div className="flex items-start justify-between">
        <div className="flex items-start space-x-3">
          {getAlertIcon(alert.type)}
          <div className="flex-1 min-w-0">
            <div className="flex items-center space-x-2 mb-1">
              <h4 className="text-sm font-medium text-gray-900 truncate">{alert.title}</h4>
              {getPriorityBadge(alert.priority)}
            </div>
            <p className="text-xs text-gray-600 line-clamp-2">{alert.message}</p>
            <div className="flex items-center space-x-3 mt-2">
              <div className="flex items-center text-xs text-gray-500">
                <Clock className="h-3 w-3 mr-1" />
                {formatDistanceToNow(new Date(alert.createdAt), { addSuffix: true })}
              </div>
              {alert.regionId && (
                <span className="text-xs bg-gray-100 text-gray-600 px-1.5 py-0.5 rounded">
                  {alert.regionId}
                </span>
              )}
            </div>
          </div>
        </div>
        
        {!alert.acknowledgedAt && (
          <button
            onClick={() => onAcknowledge(alert.id)}
            className="p-1.5 text-gray-400 hover:text-green-600 hover:bg-green-50 rounded transition-colors"
            title="Acknowledge Alert"
          >
            <Check className="h-4 w-4" />
          </button>
        )}
      </div>
    </div>
  );
};
