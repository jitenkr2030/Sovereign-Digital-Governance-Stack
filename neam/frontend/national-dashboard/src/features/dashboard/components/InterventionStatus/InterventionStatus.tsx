import React, { useState } from 'react';
import { Play, Pause, RotateCcw, CheckCircle, Clock, AlertTriangle, XCircle } from 'lucide-react';
import { Card } from '../../ui/Card';
import { Button } from '../../ui/Button';
import { Badge } from '../../ui/Badge';
import { ProgressBar } from '../../ui/ProgressBar';
import { ActionModal } from '../../common/ActionModal/ActionModal';
import { useWebSocket } from '../../../hooks/useWebSocket';
import './InterventionStatus.css';

export type InterventionStatus = 'pending' | 'active' | 'paused' | 'completed' | 'failed' | 'cancelled';

export interface InterventionAction {
  id: string;
  name: string;
  description: string;
  type: 'monetary' | 'fiscal' | 'regulatory' | 'emergency';
  status: InterventionStatus;
  startTime?: Date;
  endTime?: Date;
  progress?: number;
  impact?: {
    metric: string;
    current: number;
    target: number;
    unit: string;
  }[];
  assignedTo?: string;
  notes?: string;
}

export interface InterventionStatusProps {
  /** Array of interventions */
  interventions: InterventionAction[];
  /** WebSocket URL for real-time updates */
  wsUrl?: string;
  /** Callback when intervention is controlled */
  onControl?: (interventionId: string, action: 'start' | 'pause' | 'resume' | 'cancel') => void;
  /** Callback when intervention details are viewed */
  onViewDetails?: (intervention: InterventionAction) => void;
  /** Callback when intervention is selected */
  onSelect?: (intervention: InterventionAction) => void;
  /** Is loading */
  isLoading?: boolean;
  /** Custom class name */
  className?: string;
}

/**
 * Status configuration
 */
const statusConfig: Record<InterventionStatus, { color: string; icon: React.ReactNode; label: string }> = {
  pending: { color: 'var(--color-info)', icon: <Clock size={14} />, label: 'Pending' },
  active: { color: 'var(--color-success)', icon: <Play size={14} />, label: 'Active' },
  paused: { color: 'var(--color-warning)', icon: <Pause size={14} />, label: 'Paused' },
  completed: { color: 'var(--color-success)', icon: <CheckCircle size={14} />, label: 'Completed' },
  failed: { color: 'var(--color-danger)', icon: <XCircle size={14} />, label: 'Failed' },
  cancelled: { color: 'var(--color-text-muted)', icon: <XCircle size={14} />, label: 'Cancelled' },
};

/**
 * Get status badge variant
 */
const getStatusVariant = (status: InterventionStatus): 'default' | 'success' | 'warning' | 'danger' | 'info' => {
  switch (status) {
    case 'active':
    case 'completed':
      return 'success';
    case 'paused':
    case 'pending':
      return 'warning';
    case 'failed':
    case 'cancelled':
      return 'danger';
    default:
      return 'default';
  }
};

/**
 * InterventionStatus Component
 * 
 * Displays the status and progress of economic interventions with control actions.
 */
export const InterventionStatus: React.FC<InterventionStatusProps> = ({
  interventions,
  wsUrl,
  onControl,
  onViewDetails,
  onSelect,
  isLoading = false,
  className = '',
}) => {
  const [selectedIntervention, setSelectedIntervention] = useState<InterventionAction | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [modalAction, setModalAction] = useState<'pause' | 'resume' | 'cancel' | null>(null);

  // Real-time updates via WebSocket
  const { lastMessage } = useWebSocket(wsUrl, { enabled: !!wsUrl });

  React.useEffect(() => {
    if (lastMessage) {
      try {
        const data = JSON.parse(lastMessage.data);
        if (data.type === 'INTERVENTION_UPDATE') {
          // Handle real-time update
          console.log('Intervention update:', data.intervention);
        }
      } catch (e) {
        console.error('Failed to parse intervention update:', e);
      }
    }
  }, [lastMessage]);

  const handleControl = (intervention: InterventionAction, action: 'start' | 'pause' | 'resume' | 'cancel') => {
    if (action === 'pause' || action === 'cancel') {
      setSelectedIntervention(intervention);
      setModalAction(action);
      setModalOpen(true);
    } else {
      onControl?.(intervention.id, action);
    }
  };

  const confirmModalAction = () => {
    if (selectedIntervention && modalAction) {
      onControl?.(selectedIntervention.id, modalAction);
      setModalOpen(false);
      setSelectedIntervention(null);
      setModalAction(null);
    }
  };

  const activeInterventions = interventions.filter(i => i.status === 'active');
  const pendingInterventions = interventions.filter(i => i.status === 'pending');
  const completedInterventions = interventions.filter(i => i.status === 'completed');

  if (isLoading) {
    return (
      <Card className={`intervention-status ${className}`}>
        <div className="intervention-status-loading">
          <div className="skeleton-card" />
          <div className="skeleton-card" />
          <div className="skeleton-card" />
        </div>
      </Card>
    );
  }

  return (
    <div className={`intervention-status ${className}`}>
      {/* Summary Cards */}
      <div className="intervention-summary">
        <div className="intervention-summary-item">
          <span className="intervention-summary-value">{activeInterventions.length}</span>
          <span className="intervention-summary-label">Active</span>
        </div>
        <div className="intervention-summary-item">
          <span className="intervention-summary-value">{pendingInterventions.length}</span>
          <span className="intervention-summary-label">Pending</span>
        </div>
        <div className="intervention-summary-item">
          <span className="intervention-summary-value">{completedInterventions.length}</span>
          <span className="intervention-summary-label">Completed</span>
        </div>
      </div>

      {/* Interventions List */}
      <div className="intervention-list">
        {interventions.map((intervention) => {
          const config = statusConfig[intervention.status];
          
          return (
            <Card
              key={intervention.id}
              className={`intervention-card intervention-card-${intervention.status}`}
              onClick={() => onSelect?.(intervention)}
              clickable={!!onSelect}
            >
              <div className="intervention-header">
                <div className="intervention-info">
                  <span className="intervention-icon" style={{ color: config.color }}>
                    {config.icon}
                  </span>
                  <div>
                    <h4 className="intervention-name">{intervention.name}</h4>
                    <p className="intervention-type">{intervention.type}</p>
                  </div>
                </div>
                <Badge variant={getStatusVariant(intervention.status)} size="sm">
                  {config.label}
                </Badge>
              </div>

              <p className="intervention-description">{intervention.description}</p>

              {/* Progress */}
              {intervention.progress !== undefined && (
                <div className="intervention-progress">
                  <div className="intervention-progress-header">
                    <span>Progress</span>
                    <span>{intervention.progress}%</span>
                  </div>
                  <ProgressBar value={intervention.progress} color={config.color} />
                </div>
              )}

              {/* Impact Metrics */}
              {intervention.impact && intervention.impact.length > 0 && (
                <div className="intervention-impact">
                  {intervention.impact.map((metric, idx) => (
                    <div key={idx} className="intervention-metric">
                      <span className="metric-label">{metric.metric}</span>
                      <span className="metric-value">
                        {metric.current.toLocaleString()} / {metric.target.toLocaleString()} {metric.unit}
                      </span>
                    </div>
                  ))}
                </div>
              )}

              {/* Time Info */}
              <div className="intervention-time">
                {intervention.startTime && (
                  <span>Started: {new Date(intervention.startTime).toLocaleString()}</span>
                )}
                {intervention.endTime && (
                  <span>Ends: {new Date(intervention.endTime).toLocaleString()}</span>
                )}
              </div>

              {/* Actions */}
              <div className="intervention-actions">
                {intervention.status === 'pending' && (
                  <Button
                    variant="primary"
                    size="sm"
                    leftIcon={<Play size={14} />}
                    onClick={(e) => {
                      e.stopPropagation();
                      handleControl(intervention, 'start');
                    }}
                  >
                    Start
                  </Button>
                )}
                {intervention.status === 'active' && (
                  <>
                    <Button
                      variant="secondary"
                      size="sm"
                      leftIcon={<Pause size={14} />}
                      onClick={(e) => {
                        e.stopPropagation();
                        handleControl(intervention, 'pause');
                      }}
                    >
                      Pause
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        onViewDetails?.(intervention);
                      }}
                    >
                      Details
                    </Button>
                  </>
                )}
                {intervention.status === 'paused' && (
                  <>
                    <Button
                      variant="primary"
                      size="sm"
                      leftIcon={<Play size={14} />}
                      onClick={(e) => {
                        e.stopPropagation();
                        handleControl(intervention, 'resume');
                      }}
                    >
                      Resume
                    </Button>
                    <Button
                      variant="danger"
                      size="sm"
                      leftIcon={<XCircle size={14} />}
                      onClick={(e) => {
                        e.stopPropagation();
                        handleControl(intervention, 'cancel');
                      }}
                    >
                      Cancel
                    </Button>
                  </>
                )}
              </div>
            </Card>
          );
        })}
      </div>

      {/* Confirmation Modal */}
      <ActionModal
        isOpen={modalOpen}
        onClose={() => setModalOpen(false)}
        onConfirm={confirmModalAction}
        title={modalAction === 'pause' ? 'Pause Intervention' : 'Cancel Intervention'}
        confirmText={modalAction === 'pause' ? 'Pause' : 'Cancel'}
        confirmVariant="danger"
      >
        <p>
          {modalAction === 'pause'
            ? 'Are you sure you want to pause this intervention? You can resume it later.'
            : 'Are you sure you want to cancel this intervention? This action cannot be undone.'}
        </p>
        {selectedIntervention && (
          <p className="modal-intervention-name">{selectedIntervention.name}</p>
        )}
      </ActionModal>
    </div>
  );
};

export default InterventionStatus;
