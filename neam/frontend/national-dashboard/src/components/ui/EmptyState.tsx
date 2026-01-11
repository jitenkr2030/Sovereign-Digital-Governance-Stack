import React from 'react';
import { Inbox, Search, AlertCircle } from 'lucide-react';
import './EmptyState.css';

export interface EmptyStateProps {
  /** Icon variant */
  icon?: 'inbox' | 'search' | 'error' | 'custom';
  /** Custom icon */
  customIcon?: React.ReactNode;
  /** Title */
  title: string;
  /** Description */
  description?: string;
  /** Action button */
  action?: {
    label: string;
    onClick: () => void;
  };
  /** Custom class name */
  className?: string;
}

/**
 * Get default icon based on variant
 */
const getDefaultIcon = (variant: EmptyStateProps['icon']) => {
  switch (variant) {
    case 'search':
      return <Search size={48} />;
    case 'error':
      return <AlertCircle size={48} />;
    case 'inbox':
    default:
      return <Inbox size={48} />;
  }
};

/**
 * EmptyState Component
 * 
 * Displays a placeholder when content is empty or not found.
 */
export const EmptyState: React.FC<EmptyStateProps> = ({
  icon = 'inbox',
  customIcon,
  title,
  description,
  action,
  className = '',
}) => {
  return (
    <div className={`empty-state ${className}`}>
      <div className="empty-state-icon">
        {customIcon || getDefaultIcon(icon)}
      </div>
      
      <h3 className="empty-state-title">{title}</h3>
      
      {description && (
        <p className="empty-state-description">{description}</p>
      )}
      
      {action && (
        <button
          type="button"
          className="empty-state-action"
          onClick={action.onClick}
        >
          {action.label}
        </button>
      )}
    </div>
  );
};

export default EmptyState;
