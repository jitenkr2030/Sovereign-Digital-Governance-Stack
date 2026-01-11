import React, { useState } from 'react';
import { 
  ArrowUp, 
  ArrowDown, 
  User, 
  Clock, 
  CheckCircle, 
  AlertCircle,
  MessageSquare
} from 'lucide-react';
import { Card } from '../../ui/Card';
import { Badge } from '../../ui/Badge';
import { Button } from '../../ui/Button';
import { Textarea } from '../../ui/Textarea';
import { ActionModal } from '../../common/ActionModal/ActionModal';
import './EscalationWorkflow.css';

export interface EscalationLevel {
  id: number;
  name: string;
  description: string;
  responseTime: number; // minutes
  authorities: string[];
  autoEscalateAfter?: number; // minutes
}

export interface EscalationEntry {
  id: string;
  level: number;
  timestamp: Date;
  reason: string;
  escalatedBy: string;
  status: 'pending' | 'acknowledged' | 'resolved';
  notes?: string;
}

export interface EscalationWorkflowProps {
  /** Escalation levels configuration */
  levels: EscalationLevel[];
  /** Current escalation entries */
  entries: EscalationEntry[];
  /** Current active level */
  currentLevel: number;
  /** Maximum level */
  maxLevel: number;
  /** Callback when escalation changes */
  onEscalate?: (level: number, reason: string) => void;
  /** Callback when de-escalate */
  onDeEscalate?: (level: number) => void;
  /** Callback when entry is acknowledged */
  onAcknowledge?: (entryId: string) => void;
  /** Callback when notes are added */
  onAddNotes?: (entryId: string, notes: string) => void;
  /** Custom class name */
  className?: string;
}

/**
 * EscalationWorkflow Component
 * 
 * Visualizes and manages the escalation workflow during crisis situations.
 */
export const EscalationWorkflow: React.FC<EscalationWorkflowProps> = ({
  levels,
  entries,
  currentLevel,
  maxLevel,
  onEscalate,
  onDeEscalate,
  onAcknowledge,
  onAddNotes,
  className = '',
}) => {
  const [showEscalateModal, setShowEscalateModal] = useState(false);
  const [showDeEscalateModal, setShowDeEscalateModal] = useState(false);
  const [escalationReason, setEscalationReason] = useState('');
  const [selectedEntry, setSelectedEntry] = useState<EscalationEntry | null>(null);
  const [notes, setNotes] = useState('');

  const canEscalate = currentLevel < maxLevel;
  const canDeEscalate = currentLevel > 1;

  const handleEscalate = () => {
    if (escalationReason.trim()) {
      onEscalate?.(currentLevel + 1, escalationReason);
      setShowEscalateModal(false);
      setEscalationReason('');
    }
  };

  const handleDeEscalate = () => {
    onDeEscalate?.(currentLevel - 1);
    setShowDeEscalateModal(false);
  };

  const handleAddNotes = () => {
    if (selectedEntry && notes.trim()) {
      onAddNotes?.(selectedEntry.id, notes);
      setSelectedEntry(null);
      setNotes('');
    }
  };

  // Get current level config
  const currentLevelConfig = levels.find(l => l.id === currentLevel);
  const nextLevelConfig = levels.find(l => l.id === currentLevel + 1);

  return (
    <div className={`escalation-workflow ${className}`}>
      {/* Current Level Indicator */}
      <Card className="current-level-card">
        <div className="current-level-header">
          <h3 className="current-level-title">Current Escalation Level</h3>
          <Badge 
            variant={currentLevel >= 4 ? 'danger' : currentLevel >= 3 ? 'warning' : 'info'}
            size="lg"
          >
            Level {currentLevel}: {currentLevelConfig?.name || 'Unknown'}
          </Badge>
        </div>

        <div className="current-level-info">
          <p className="level-description">{currentLevelConfig?.description}</p>
          
          <div className="level-meta">
            <span className="meta-item">
              <Clock size={14} />
              Response time: {currentLevelConfig?.responseTime} min
            </span>
            {currentLevelConfig?.authorities.length > 0 && (
              <span className="meta-item">
                <User size={14} />
                Authorities: {currentLevelConfig.authorities.join(', ')}
              </span>
            )}
          </div>
        </div>

        {/* Escalation Controls */}
        <div className="escalation-controls">
          <Button
            variant="danger"
            size="lg"
            leftIcon={<ArrowUp size={16} />}
            onClick={() => setShowEscalateModal(true)}
            disabled={!canEscalate}
            fullWidth
          >
            Escalate to Level {currentLevel + 1}
            {nextLevelConfig && `: ${nextLevelConfig.name}`}
          </Button>
          
          <Button
            variant="outline"
            size="lg"
            leftIcon={<ArrowDown size={16} />}
            onClick={() => setShowDeEscalateModal(true)}
            disabled={!canDeEscalate}
            fullWidth
          >
            De-escalate to Level {currentLevel - 1}
          </Button>
        </div>
      </Card>

      {/* Escalation Timeline */}
      <Card className="timeline-card">
        <h4 className="timeline-title">Escalation History</h4>
        
        <div className="timeline">
          {entries.length === 0 ? (
            <div className="timeline-empty">
              <Clock size={24} />
              <p>No escalation history</p>
            </div>
          ) : (
            entries.map((entry, index) => {
              const levelConfig = levels.find(l => l.id === entry.level);
              
              return (
                <div key={entry.id} className={`timeline-entry entry-level-${entry.level}`}>
                  <div className="timeline-marker">
                    <div className="timeline-dot" />
                    {index < entries.length - 1 && <div className="timeline-line" />}
                  </div>
                  
                  <div className="timeline-content">
                    <div className="timeline-header">
                      <Badge 
                        variant={entry.level >= 4 ? 'danger' : entry.level >= 3 ? 'warning' : 'info'}
                        size="sm"
                      >
                        Level {entry.level}: {levelConfig?.name}
                      </Badge>
                      <span className="timeline-time">
                        {new Date(entry.timestamp).toLocaleString()}
                      </span>
                    </div>
                    
                    <p className="timeline-reason">{entry.reason}</p>
                    
                    <div className="timeline-meta">
                      <span className="escalated-by">
                        <User size={12} />
                        {entry.escalatedBy}
                      </span>
                      
                      <Badge 
                        variant={entry.status === 'resolved' ? 'success' : entry.status === 'acknowledged' ? 'warning' : 'default'}
                        size="sm"
                      >
                        {entry.status}
                      </Badge>
                    </div>

                    {/* Actions */}
                    <div className="timeline-actions">
                      {entry.status === 'pending' && (
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => onAcknowledge?.(entry.id)}
                        >
                          <CheckCircle size={14} />
                          Acknowledge
                        </Button>
                      )}
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => {
                          setSelectedEntry(entry);
                          setNotes(entry.notes || '');
                        }}
                      >
                        <MessageSquare size={14} />
                        {entry.notes ? 'Edit Notes' : 'Add Notes'}
                      </Button>
                    </div>

                    {entry.notes && (
                      <div className="entry-notes">
                        <strong>Notes:</strong> {entry.notes}
                      </div>
                    )}
                  </div>
                </div>
              );
            })
          )}
        </div>
      </Card>

      {/* Escalate Modal */}
      <ActionModal
        isOpen={showEscalateModal}
        onClose={() => setShowEscalateModal(false)}
        onConfirm={handleEscalate}
        title="Escalate Crisis Level"
        confirmText="Confirm Escalation"
        confirmVariant="danger"
      >
        <div className="escalate-modal-content">
          <p>
            You are about to escalate from <strong>Level {currentLevel}</strong> to{' '}
            <strong>Level {currentLevel + 1}</strong>.
          </p>
          
          {nextLevelConfig && (
            <div className="next-level-info">
              <h5>Level {currentLevel + 1}: {nextLevelConfig.name}</h5>
              <p>{nextLevelConfig.description}</p>
              <ul>
                <li>Response time: {nextLevelConfig.responseTime} minutes</li>
                {nextLevelConfig.authorities.length > 0 && (
                  <li>Authorities: {nextLevelConfig.authorities.join(', ')}</li>
                )}
              </ul>
            </div>
          )}
          
          <Textarea
            label="Reason for escalation"
            placeholder="Describe why this escalation is necessary..."
            value={escalationReason}
            onChange={(e) => setEscalationReason(e.target.value)}
            required
          />
        </div>
      </ActionModal>

      {/* De-escalate Modal */}
      <ActionModal
        isOpen={showDeEscalateModal}
        onClose={() => setShowDeEscalateModal(false)}
        onConfirm={handleDeEscalate}
        title="De-escalate Crisis Level"
        confirmText="Confirm De-escalation"
        confirmVariant="primary"
      >
        <p>
          You are about to de-escalate from <strong>Level {currentLevel}</strong> to{' '}
          <strong>Level {currentLevel - 1}</strong>.
        </p>
        <p className="modal-warning">
          Only de-escalate if the situation has been stabilized and the risk has reduced.
        </p>
      </ActionModal>

      {/* Notes Modal */}
      <ActionModal
        isOpen={!!selectedEntry}
        onClose={() => setSelectedEntry(null)}
        onConfirm={handleAddNotes}
        title="Add Notes"
        confirmText="Save Notes"
      >
        <Textarea
          label="Notes"
          placeholder="Add notes about this escalation level..."
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          rows={4}
        />
      </ActionModal>
    </div>
  );
};

export default EscalationWorkflow;
