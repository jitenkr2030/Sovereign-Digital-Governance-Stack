import React from 'react';
import { createPortal } from 'react-dom';
import { AlertCircle, CheckCircle, Info, AlertTriangle, X } from 'lucide-react';
import './AlertNotification.css';

export type AlertType = 'success' | 'error' | 'warning' | 'info';

export interface AlertNotificationProps {
  /** Alert type determining color and icon */
  type: AlertType;
  /** Alert title */
  title?: string;
  /** Alert message */
  message: string | React.ReactNode;
  /** Callback when alert is dismissed */
  onDismiss?: () => void;
  /** Auto-dismiss delay in milliseconds (0 = no auto-dismiss) */
  dismissDelay?: number;
  /** Custom action button */
  action?: {
    label: string;
    onClick: () => void;
  };
  /** Show icon */
  showIcon?: boolean;
  /** Custom class name */
  className?: string;
  /** Persistent (cannot be dismissed) */
  persistent?: boolean;
}

/**
 * AlertNotification Component
 * 
 * Toast-style notification alerts for feedback and status updates.
 */
export const AlertNotification: React.FC<AlertNotificationProps> = ({
  type,
  title,
  message,
  onDismiss,
  dismissDelay = 5000,
  action,
  showIcon = true,
  className = '',
  persistent = false,
}) => {
  React.useEffect(() => {
    if (persistent || !onDismiss) return;

    const timer = setTimeout(() => {
      onDismiss();
    }, dismissDelay);

    return () => clearTimeout(timer);
  }, [persistent, onDismiss, dismissDelay]);

  const icons = {
    success: <CheckCircle className="alert-icon" />,
    error: <AlertCircle className="alert-icon" />,
    warning: <AlertTriangle className="alert-icon" />,
    info: <Info className="alert-icon" />,
  };

  return (
    <div className={`alert-notification alert-${type} ${className}`} role="alert">
      {showIcon && <div className="alert-icon-container">{icons[type]}</div>}
      
      <div className="alert-content">
        {title && <div className="alert-title">{title}</div>}
        <div className="alert-message">{message}</div>
        {action && (
          <button
            type="button"
            className="alert-action"
            onClick={action.onClick}
          >
            {action.label}
          </button>
        )}
      </div>

      {!persistent && onDismiss && (
        <button
          type="button"
          className="alert-dismiss"
          onClick={onDismiss}
          aria-label="Dismiss notification"
        >
          <X size={16} />
        </button>
      )}
    </div>
  );
};

// Container for managing multiple alerts
interface AlertItem extends AlertNotificationProps {
  id: string;
}

interface AlertContainerProps {
  alerts: AlertItem[];
  onDismiss: (id: string) => void;
  position?: 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left' | 'top-center' | 'bottom-center';
}

export const AlertContainer: React.FC<AlertContainerProps> = ({
  alerts,
  onDismiss,
  position = 'top-right',
}) => {
  if (typeof document === 'undefined') return null;

  return createPortal(
    <div className={`alert-container alert-position-${position}`}>
      {alerts.map((alert) => (
        <div
          key={alert.id}
          className="alert-wrapper"
          style={{ animation: 'slideInRight 0.3s ease-out' }}
        >
          <AlertNotification {...alert} onDismiss={() => onDismiss(alert.id)} />
        </div>
      ))}
    </div>,
    document.body
  );
};

// Hook for managing alerts
interface UseAlertOptions {
  maxAlerts?: number;
  defaultDismissDelay?: number;
}

interface AlertDispatch extends Omit<AlertItem, 'id'> {
  id?: string;
}

export function useAlert(options: UseAlertOptions = {}) {
  const { maxAlerts = 5, defaultDismissDelay = 5000 } = options;
  const [alerts, setAlerts] = React.useState<AlertItem[]>([]);

  const addAlert = React.useCallback((alert: AlertDispatch) => {
    const id = alert.id || `alert-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    
    setAlerts((prev) => {
      const newAlerts = [
        ...prev,
        { ...alert, id, dismissDelay: alert.dismissDelay ?? defaultDismissDelay },
      ];
      return newAlerts.slice(-maxAlerts);
    });

    return id;
  }, [maxAlerts, defaultDismissDelay]);

  const dismissAlert = React.useCallback((id: string) => {
    setAlerts((prev) => prev.filter((alert) => alert.id !== id));
  }, []);

  const success = React.useCallback((message: string | React.ReactNode, title?: string) => {
    return addAlert({ type: 'success', message, title });
  }, [addAlert]);

  const error = React.useCallback((message: string | React.ReactNode, title?: string) => {
    return addAlert({ type: 'error', message, title });
  }, [addAlert]);

  const warning = React.useCallback((message: string | React.ReactNode, title?: string) => {
    return addAlert({ type: 'warning', message, title });
  }, [addAlert]);

  const info = React.useCallback((message: string | React.ReactNode, title?: string) => {
    return addAlert({ type: 'info', message, title });
  }, [addAlert]);

  return {
    alerts,
    addAlert,
    dismissAlert,
    success,
    error,
    warning,
    info,
    AlertContainer: () => <AlertContainer alerts={alerts} onDismiss={dismissAlert} />,
  };
}

export default AlertNotification;
