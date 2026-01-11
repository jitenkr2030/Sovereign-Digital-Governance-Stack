import React, { useState, useCallback } from 'react';
import { Play, Pause, Square, RotateCcw, Settings, ChevronRight, AlertTriangle, Shield, Zap } from 'lucide-react';
import { Card } from '../../ui/Card';
import { Button } from '../../ui/Button';
import { Slider } from '../../ui/Slider';
import { ActionModal } from '../../common/ActionModal/ActionModal';
import { Input } from '../../ui/Input';
import { Badge } from '../../ui/Badge';
import { useAlert } from '../../feedback/AlertNotification/AlertNotification';
import './InterventionControlPanel.css';

export interface InterventionControl {
  id: string;
  name: string;
  description: string;
  type: 'monetary' | 'fiscal' | 'regulatory' | 'emergency';
  enabled: boolean;
  parameters: {
    key: string;
    label: string;
    value: number | string;
    min?: number;
    max?: number;
    step?: number;
    unit?: string;
    options?: { label: string; value: string | number }[];
  }[];
  confirmationRequired: boolean;
  cooldown?: number; // minutes
  lastUsed?: Date;
  maxUsesPerDay?: number;
  currentUsesToday?: number;
}

export interface InterventionControlPanelProps {
  /** Available interventions */
  controls: InterventionControl[];
  /** Callback when intervention is triggered */
  onTrigger?: (controlId: string, parameters: Record<string, unknown>) => void;
  /** Callback when intervention is enabled/disabled */
  onToggle?: (controlId: string, enabled: boolean) => void;
  /** Callback when parameters change */
  onParameterChange?: (controlId: string, key: string, value: unknown) => void;
  /** Callback when settings are opened */
  onOpenSettings?: (control: InterventionControl) => void;
  /** Custom class name */
  className?: string;
}

/**
 * InterventionControlPanel Component
 * 
 * Direct control panel for triggering economic interventions during crisis situations.
 */
export const InterventionControlPanel: React.FC<InterventionControlPanelProps> = ({
  controls,
  onTrigger,
  onToggle,
  onParameterChange,
  onOpenSettings,
  className = '',
}) => {
  const [activeControl, setActiveControl] = useState<InterventionControl | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [parameters, setParameters] = useState<Record<string, unknown>>({});
  const [isProcessing, setIsProcessing] = useState(false);
  const { success, error } = useAlert();

  // Handle trigger button click
  const handleTrigger = useCallback((control: InterventionControl) => {
    if (control.confirmationRequired) {
      setActiveControl(control);
      // Initialize parameters with current values
      const initialParams: Record<string, unknown> = {};
      control.parameters.forEach((p) => {
        initialParams[p.key] = p.value;
      });
      setParameters(initialParams);
      setModalOpen(true);
    } else {
      // Execute immediately
      setIsProcessing(true);
      const params: Record<string, unknown> = {};
      control.parameters.forEach((p) => {
        params[p.key] = p.value;
      });
      
      onTrigger?.(control.id, params);
      
      setTimeout(() => {
        setIsProcessing(false);
        success(`Intervention "${control.name}" executed successfully`);
      }, 1000);
    }
  }, [onTrigger, success]);

  // Confirm intervention execution
  const confirmTrigger = useCallback(async () => {
    if (!activeControl) return;
    
    setIsProcessing(true);
    try {
      await onTrigger?.(activeControl.id, parameters);
      success(`Intervention "${activeControl.name}" executed successfully`);
      setModalOpen(false);
      setActiveControl(null);
    } catch (e) {
      error('Failed to execute intervention');
    } finally {
      setIsProcessing(false);
    }
  }, [activeControl, parameters, onTrigger, success, error]);

  // Update parameter value
  const updateParameter = useCallback((key: string, value: unknown) => {
    setParameters((prev) => ({ ...prev, [key]: value }));
    onParameterChange?.(activeControl!.id, key, value);
  }, [activeControl, onParameterChange]);

  // Get icon for intervention type
  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'monetary':
        return <Zap size={16} />;
      case 'fiscal':
        return <Shield size={16} />;
      case 'regulatory':
        return <Settings size={16} />;
      case 'emergency':
        return <AlertTriangle size={16} />;
      default:
        return <Settings size={16} />;
    }
  };

  // Get type color
  const getTypeColor = (type: string): string => {
    switch (type) {
      case 'monetary':
        return 'var(--color-primary)';
      case 'fiscal':
        return 'var(--color-success)';
      case 'regulatory':
        return 'var(--color-warning)';
      case 'emergency':
        return 'var(--color-danger)';
      default:
        return 'var(--color-text-muted)';
    }
  };

  const canUse = (control: InterventionControl): boolean => {
    if (!control.enabled) return false;
    if (control.maxUsesPerDay && control.currentUsesToday && control.currentUsesToday >= control.maxUsesPerDay) {
      return false;
    }
    return true;
  };

  return (
    <div className={`intervention-control-panel ${className}`}>
      {/* Header */}
      <div className="control-panel-header">
        <h3 className="control-panel-title">Direct Intervention Controls</h3>
        <Badge variant="warning" size="md">
          <AlertTriangle size={12} />
          Crisis Mode Active
        </Badge>
      </div>

      {/* Controls Grid */}
      <div className="controls-grid">
        {controls.map((control) => {
          const typeColor = getTypeColor(control.type);
          const isAvailable = canUse(control);

          return (
            <Card
              key={control.id}
              className={`control-card ${control.enabled ? 'control-enabled' : 'control-disabled'}`}
            >
              {/* Control Header */}
              <div className="control-header">
                <div 
                  className="control-type-icon"
                  style={{ backgroundColor: `${typeColor}20`, color: typeColor }}
                >
                  {getTypeIcon(control.type)}
                </div>
                <div className="control-info">
                  <h4 className="control-name">{control.name}</h4>
                  <span className="control-type">{control.type}</span>
                </div>
                <label className="control-toggle">
                  <input
                    type="checkbox"
                    checked={control.enabled}
                    onChange={(e) => onToggle?.(control.id, e.target.checked)}
                    disabled={!isAvailable}
                  />
                  <span className="toggle-slider" />
                </label>
              </div>

              {/* Description */}
              <p className="control-description">{control.description}</p>

              {/* Parameters Preview */}
              {control.parameters.length > 0 && control.enabled && (
                <div className="control-parameters">
                  {control.parameters.slice(0, 2).map((param) => (
                    <div key={param.key} className="parameter-preview">
                      <span className="parameter-label">{param.label}:</span>
                      <span className="parameter-value">
                        {param.value} {param.unit}
                      </span>
                    </div>
                  ))}
                  {control.parameters.length > 2 && (
                    <span className="parameter-more">
                      +{control.parameters.length - 2} more
                    </span>
                  )}
                </div>
              )}

              {/* Usage Stats */}
              {control.maxUsesPerDay && (
                <div className="control-usage">
                  <span>Daily usage: {control.currentUsesToday || 0}/{control.maxUsesPerDay}</span>
                  <div className="usage-bar">
                    <div 
                      className="usage-fill"
                      style={{ 
                        width: `${((control.currentUsesToday || 0) / control.maxUsesPerDay) * 100}%`,
                        backgroundColor: (control.currentUsesToday || 0) >= control.maxUsesPerDay 
                          ? 'var(--color-danger)' 
                          : typeColor
                      }}
                    />
                  </div>
                </div>
              )}

              {/* Actions */}
              <div className="control-actions">
                <Button
                  variant="primary"
                  size="sm"
                  leftIcon={<Play size={14} />}
                  onClick={() => handleTrigger(control)}
                  disabled={!control.enabled || !isAvailable || isProcessing}
                  loading={isProcessing}
                  fullWidth
                >
                  {control.confirmationRequired ? 'Configure & Execute' : 'Execute'}
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  leftIcon={<Settings size={14} />}
                  onClick={() => onOpenSettings?.(control)}
                >
                  Settings
                </Button>
              </div>
            </Card>
          );
        })}
      </div>

      {/* Confirmation Modal */}
      <ActionModal
        isOpen={modalOpen}
        onClose={() => setModalOpen(false)}
        onConfirm={confirmTrigger}
        title={`Execute: ${activeControl?.name}`}
        confirmText="Execute Intervention"
        confirmVariant="danger"
        isLoading={isProcessing}
      >
        {activeControl && (
          <div className="intervention-modal-content">
            <div className="intervention-warning">
              <AlertTriangle size={24} />
              <p>
                This intervention will be executed immediately. Please confirm all parameters below.
              </p>
            </div>

            {/* Parameter Inputs */}
            <div className="modal-parameters">
              {activeControl.parameters.map((param) => (
                <div key={param.key} className="modal-parameter">
                  <label className="modal-parameter-label">
                    {param.label}
                    {param.unit && <span className="parameter-unit">({param.unit})</span>}
                  </label>
                  
                  {param.options ? (
                    <select
                      className="modal-parameter-select"
                      value={parameters[param.key] as string}
                      onChange={(e) => updateParameter(param.key, e.target.value)}
                    >
                      {param.options.map((opt) => (
                        <option key={opt.value} value={opt.value}>
                          {opt.label}
                        </option>
                      ))}
                    </select>
                  ) : param.min !== undefined && param.max !== undefined ? (
                    <div className="modal-parameter-slider">
                      <Slider
                        value={parameters[param.key] as number ?? param.value as number}
                        min={param.min}
                        max={param.max}
                        step={param.step || 1}
                        onChange={(value) => updateParameter(param.key, value)}
                      />
                      <span className="slider-value">
                        {parameters[param.key] ?? param.value} {param.unit}
                      </span>
                    </div>
                  ) : (
                    <Input
                      type="number"
                      value={parameters[param.key] as number ?? param.value as number}
                      onChange={(e) => updateParameter(param.key, parseFloat(e.target.value))}
                      min={param.min}
                      max={param.max}
                    />
                  )}
                </div>
              ))}
            </div>

            {/* Notes */}
            <div className="modal-notes">
              <label className="modal-parameter-label">Execution Notes (Optional)</label>
              <textarea
                className="modal-notes-textarea"
                placeholder="Add notes about this intervention..."
                rows={3}
              />
            </div>
          </div>
        )}
      </ActionModal>
    </div>
  );
};

export default InterventionControlPanel;
