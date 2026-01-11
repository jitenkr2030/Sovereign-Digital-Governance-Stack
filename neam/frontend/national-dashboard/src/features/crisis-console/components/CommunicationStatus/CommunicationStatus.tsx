import React, { useState, useCallback } from 'react';
import { 
  Phone, 
  Mail, 
  MessageSquare, 
  Radio, 
  Users, 
  CheckCircle, 
  XCircle,
  AlertTriangle
} from 'lucide-react';
import { Card } from '../../ui/Card';
import { Badge } from '../../ui/Badge';
import { Button } from '../../ui/Button';
import { ProgressBar } from '../../ui/ProgressBar';
import './CommunicationStatus.css';

export interface CommunicationChannel {
  id: string;
  name: string;
  type: 'phone' | 'email' | 'sms' | 'radio' | 'internal';
  status: 'operational' | 'degraded' | 'outage' | 'maintenance';
  latency?: number; // ms
  uptime: number; // percentage
  lastChecked: Date;
  messageCount?: number;
  responseRate?: number;
}

export interface TeamMember {
  id: string;
  name: string;
  role: string;
  status: 'available' | 'busy' | 'offline' | 'break';
  location?: string;
  lastActive: Date;
  assignedTasks?: number;
}

export interface CommunicationStatusProps {
  /** Communication channels */
  channels: CommunicationChannel[];
  /** Team members */
  teamMembers?: TeamMember[];
  /** Active broadcasts */
  activeBroadcasts?: number;
  /** On-call personnel count */
  onCallCount?: number;
  /** Callback to send broadcast */
  onSendBroadcast?: (message: string, channels: string[]) => void;
  /** Callback to contact team member */
  onContactMember?: (memberId: string) => void;
  /** Custom class name */
  className?: string;
}

/**
 * Get icon for channel type
 */
const getChannelIcon = (type: CommunicationChannel['type']) => {
  switch (type) {
    case 'phone':
      return <Phone size={16} />;
    case 'email':
      return <Mail size={16} />;
    case 'sms':
      return <MessageSquare size={16} />;
    case 'radio':
      return <Radio size={16} />;
    case 'internal':
      return <MessageSquare size={16} />;
    default:
      return <MessageSquare size={16} />;
  }
};

/**
 * Get status configuration
 */
const getStatusConfig = (status: CommunicationChannel['status']) => {
  switch (status) {
    case 'operational':
      return { color: 'var(--color-success)', label: 'Operational' };
    case 'degraded':
      return { color: 'var(--color-warning)', label: 'Degraded' };
    case 'outage':
      return { color: 'var(--color-danger)', label: 'Outage' };
    case 'maintenance':
      return { color: 'var(--color-info)', label: 'Maintenance' };
    default:
      return { color: 'var(--color-text-muted)', label: status };
  }
};

/**
 * CommunicationStatus Component
 * 
 * Displays the status of communication channels and team availability.
 */
export const CommunicationStatus: React.FC<CommunicationStatusProps> = ({
  channels,
  teamMembers = [],
  activeBroadcasts = 0,
  onCallCount = 0,
  onSendBroadcast,
  onContactMember,
  className = '',
}) => {
  const [broadcastMessage, setBroadcastMessage] = useState('');
  const [selectedChannels, setSelectedChannels] = useState<Set<string>>(new Set());

  const toggleChannel = (channelId: string) => {
    setSelectedChannels((prev) => {
      const next = new Set(prev);
      if (next.has(channelId)) {
        next.delete(channelId);
      } else {
        next.add(channelId);
      }
      return next;
    });
  };

  const handleSendBroadcast = () => {
    if (broadcastMessage.trim() && selectedChannels.size > 0) {
      onSendBroadcast?.(broadcastMessage, Array.from(selectedChannels));
      setBroadcastMessage('');
      setSelectedChannels(new Set());
    }
  };

  // Count operational channels
  const operationalCount = channels.filter((c) => c.status === 'operational').length;

  return (
    <div className={`communication-status ${className}`}>
      {/* Header */}
      <div className="communication-header">
        <h3 className="communication-title">Communication Status</h3>
        <div className="communication-summary">
          <Badge variant={operationalCount === channels.length ? 'success' : 'warning'} size="md">
            {operationalCount}/{channels.length} Channels Operational
          </Badge>
        </div>
      </div>

      {/* Broadcast Panel */}
      {onSendBroadcast && (
        <Card className="broadcast-panel">
          <h4 className="broadcast-title">Emergency Broadcast</h4>
          <textarea
            className="broadcast-input"
            placeholder="Enter broadcast message..."
            value={broadcastMessage}
            onChange={(e) => setBroadcastMessage(e.target.value)}
            rows={2}
          />
          <div className="broadcast-footer">
            <div className="channel-select">
              {channels
                .filter((c) => c.status === 'operational')
                .map((channel) => (
                  <button
                    key={channel.id}
                    className={`channel-btn ${selectedChannels.has(channel.id) ? 'selected' : ''}`}
                    onClick={() => toggleChannel(channel.id)}
                  >
                    {getChannelIcon(channel.type)}
                    {channel.name}
                  </button>
                ))}
            </div>
            <Button
              variant="danger"
              size="sm"
              onClick={handleSendBroadcast}
              disabled={!broadcastMessage.trim() || selectedChannels.size === 0}
              leftIcon={<Radio size={14} />}
            >
              Send Broadcast
            </Button>
          </div>
        </Card>
      )}

      {/* Channels Grid */}
      <div className="channels-grid">
        {channels.map((channel) => {
          const status = getStatusConfig(channel.status);
          
          return (
            <Card key={channel.id} className="channel-card">
              <div className="channel-header">
                <div 
                  className="channel-icon"
                  style={{ color: status.color }}
                >
                  {getChannelIcon(channel.type)}
                </div>
                <Badge 
                  variant={channel.status === 'operational' ? 'success' : channel.status === 'outage' ? 'danger' : 'warning'}
                  size="sm"
                >
                  {status.label}
                </Badge>
              </div>
              
              <div className="channel-info">
                <h5 className="channel-name">{channel.name}</h5>
                {channel.latency && (
                  <span className="channel-latency">
                    Latency: {channel.latency}ms
                  </span>
                )}
              </div>

              <div className="channel-metrics">
                <div className="metric-row">
                  <span>Uptime</span>
                  <span className="metric-value">{channel.uptime}%</span>
                </div>
                <ProgressBar 
                  value={channel.uptime} 
                  size="sm"
                  variant={channel.uptime > 99 ? 'success' : channel.uptime > 95 ? 'warning' : 'danger'}
                />
              </div>

              <div className="channel-footer">
                <span className="last-checked">
                  Last checked: {new Date(channel.lastChecked).toLocaleTimeString()}
                </span>
              </div>
            </Card>
          );
        })}
      </div>

      {/* Team Availability */}
      {teamMembers.length > 0 && (
        <Card className="team-panel">
          <div className="team-header">
            <h4 className="team-title">
              <Users size={16} />
              Team Availability
            </h4>
            <Badge variant="info" size="sm">
              {onCallCount} on call
            </Badge>
          </div>

          <div className="team-grid">
            {teamMembers.map((member) => (
              <div key={member.id} className="team-member-card">
                <div className={`member-status status-${member.status}`} />
                <div className="member-info">
                  <span className="member-name">{member.name}</span>
                  <span className="member-role">{member.role}</span>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onContactMember?.(member.id)}
                >
                  Contact
                </Button>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* Active Broadcasts Summary */}
      {activeBroadcasts > 0 && (
        <div className="active-broadcasts">
          <AlertTriangle size={14} />
          <span>{activeBroadcasts} active broadcast(s)</span>
        </div>
      )}
    </div>
  );
};

export default CommunicationStatus;
