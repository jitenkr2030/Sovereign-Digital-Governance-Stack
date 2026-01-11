import React from 'react';
import { useGetActiveInterventionsQuery, useApproveInterventionMutation, useCancelInterventionMutation } from '../store/api/interventionApi';
import type { Intervention, InterventionStatus } from '../types';
import { 
  Play, 
  Pause, 
  CheckCircle, 
  XCircle, 
  Clock,
  ArrowRight,
  ExternalLink,
  X
} from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

interface InterventionsPanelProps {
  onClose?: () => void;
}

export const InterventionsPanel: React.FC<InterventionsPanelProps> = ({ onClose }) => {
  const { data: interventions, isLoading, refetch } = useGetActiveInterventionsQuery();
  const [approveIntervention] = useApproveInterventionMutation();
  const [cancelIntervention] = useCancelInterventionMutation();

  const getStatusIcon = (status: InterventionStatus) => {
    switch (status) {
      case 'EXECUTING': return <Play className="h-4 w-4 text-blue-500" />;
      case 'PENDING': return <Clock className="h-4 w-4 text-yellow-500" />;
      case 'APPROVED': return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'COMPLETED': return <CheckCircle className="h-4 w-4 text-gray-500" />;
      case 'FAILED': return <XCircle className="h-4 w-4 text-red-500" />;
      case 'CANCELLED': return <XCircle className="h-4 w-4 text-gray-400" />;
      default: return <Clock className="h-4 w-4 text-gray-500" />;
    }
  };

  const getStatusStyles = (status: InterventionStatus) => {
    switch (status) {
      case 'EXECUTING': return 'border-l-blue-500 bg-blue-50';
      case 'PENDING': return 'border-l-yellow-500 bg-yellow-50';
      case 'APPROVED': return 'border-l-green-500 bg-green-50';
      case 'COMPLETED': return 'border-l-gray-400 bg-gray-50';
      case 'FAILED': return 'border-l-red-500 bg-red-50';
      case 'CANCELLED': return 'border-l-gray-300 bg-gray-50';
      default: return 'border-l-gray-300 bg-gray-50';
    }
  };

  const getStatusBadge = (status: InterventionStatus) => {
    const styles = {
      EXECUTING: 'bg-blue-100 text-blue-700',
      PENDING: 'bg-yellow-100 text-yellow-700',
      APPROVED: 'bg-green-100 text-green-700',
      COMPLETED: 'bg-gray-100 text-gray-700',
      FAILED: 'bg-red-100 text-red-700',
      CANCELLED: 'bg-gray-100 text-gray-500',
    };
    
    return (
      <span className={`px-2 py-0.5 rounded text-xs font-medium ${styles[status] || styles.PENDING}`}>
        {status}
      </span>
    );
  };

  const handleApprove = async (interventionId: string) => {
    try {
      await approveIntervention({ 
        id: interventionId, 
        approver: 'current_user',
        notes: 'Approved via dashboard' 
      });
      refetch();
    } catch (error) {
      console.error('Failed to approve intervention:', error);
    }
  };

  const handleCancel = async (interventionId: string) => {
    try {
      await cancelIntervention({ 
        id: interventionId, 
        reason: 'Cancelled via dashboard' 
      });
      refetch();
    } catch (error) {
      console.error('Failed to cancel intervention:', error);
    }
  };

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow p-4">
        <div className="animate-pulse space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-20 bg-gray-200 rounded"></div>
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
          <Play className="h-5 w-5 text-blue-500" />
          <h3 className="font-semibold text-gray-900">Active Interventions</h3>
          {interventions && interventions.length > 0 && (
            <span className="bg-blue-100 text-blue-700 text-xs px-2 py-0.5 rounded-full">
              {interventions.length}
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

      {/* Intervention List */}
      <div className="max-h-96 overflow-y-auto">
        {interventions && interventions.length > 0 ? (
          <div className="divide-y">
            {interventions.map((intervention) => (
              <InterventionItem 
                key={intervention.id}
                intervention={intervention}
                onApprove={handleApprove}
                onCancel={handleCancel}
                getStatusIcon={getStatusIcon}
                getStatusStyles={getStatusStyles}
                getStatusBadge={getStatusBadge}
              />
            ))}
          </div>
        ) : (
          <div className="p-8 text-center text-gray-500">
            <Play className="h-12 w-12 mx-auto text-gray-300 mb-2" />
            <p>No active interventions</p>
            <p className="text-sm">All policies operating normally</p>
          </div>
        )}
      </div>

      {/* Footer */}
      {interventions && interventions.length > 0 && (
        <div className="p-3 border-t bg-gray-50">
          <button 
            className="w-full text-center text-sm text-blue-600 hover:text-blue-800 font-medium"
          >
            View All Interventions
            <ExternalLink className="h-4 w-4 inline ml-1" />
          </button>
        </div>
      )}
    </div>
  );
};

interface InterventionItemProps {
  intervention: Intervention;
  onApprove: (id: string) => void;
  onCancel: (id: string) => void;
  getStatusIcon: (status: InterventionStatus) => React.ReactNode;
  getStatusStyles: (status: InterventionStatus) => string;
  getStatusBadge: (status: InterventionStatus) => React.ReactNode;
}

const InterventionItem: React.FC<InterventionItemProps> = ({
  intervention,
  onApprove,
  onCancel,
  getStatusIcon,
  getStatusStyles,
  getStatusBadge,
}) => {
  return (
    <div className={`p-3 border-l-4 ${getStatusStyles(intervention.status)}`}>
      <div className="flex items-start justify-between mb-2">
        <div>
          <h4 className="text-sm font-medium text-gray-900">{intervention.policyName}</h4>
          <div className="flex items-center space-x-2 mt-1">
            {getStatusIcon(intervention.status)}
            {getStatusBadge(intervention.status)}
          </div>
        </div>
        
        {/* Progress Circle */}
        <div className="relative w-12 h-12">
          <svg className="w-12 h-12 transform -rotate-90">
            <circle
              cx="24"
              cy="24"
              r="20"
              stroke="#e5e7eb"
              strokeWidth="4"
              fill="none"
            />
            <circle
              cx="24"
              cy="24"
              r="20"
              stroke="#3b82f6"
              strokeWidth="4"
              fill="none"
              strokeDasharray={`${intervention.progress * 1.256} 125.6`}
              strokeLinecap="round"
            />
          </svg>
          <span className="absolute inset-0 flex items-center justify-center text-xs font-semibold text-gray-700">
            {Math.round(intervention.progress)}%
          </span>
        </div>
      </div>

      {/* Current Step */}
      <div className="text-xs text-gray-600 mb-2">
        <span className="font-medium">Current Step:</span> {intervention.currentStep}
      </div>

      {/* Region and Timing */}
      <div className="flex items-center justify-between text-xs text-gray-500 mb-3">
        <span>{intervention.regionId}</span>
        <span>{formatDistanceToNow(new Date(intervention.triggeredAt), { addSuffix: true })}</span>
      </div>

      {/* Actions */}
      {intervention.status === 'PENDING' && (
        <div className="flex items-center space-x-2">
          <button
            onClick={() => onApprove(intervention.id)}
            className="flex-1 flex items-center justify-center space-x-1 px-3 py-1.5 bg-green-600 text-white rounded text-xs font-medium hover:bg-green-700"
          >
            <CheckCircle className="h-3.5 w-3.5" />
            <span>Approve</span>
          </button>
          <button
            onClick={() => onCancel(intervention.id)}
            className="flex items-center justify-center px-3 py-1.5 bg-red-100 text-red-700 rounded text-xs font-medium hover:bg-red-200"
          >
            <XCircle className="h-3.5 w-3.5" />
          </button>
        </div>
      )}

      {intervention.status === 'EXECUTING' && (
        <div className="flex items-center space-x-2">
          <button
            onClick={() => onCancel(intervention.id)}
            className="flex-1 flex items-center justify-center space-x-1 px-3 py-1.5 bg-red-100 text-red-700 rounded text-xs font-medium hover:bg-red-200"
          >
            <Pause className="h-3.5 w-3.5" />
            <span>Cancel</span>
          </button>
          <button className="flex items-center justify-center px-3 py-1.5 bg-blue-100 text-blue-700 rounded text-xs font-medium hover:bg-blue-200">
            <ArrowRight className="h-3.5 w-3.5" />
          </button>
        </div>
      )}
    </div>
  );
};
