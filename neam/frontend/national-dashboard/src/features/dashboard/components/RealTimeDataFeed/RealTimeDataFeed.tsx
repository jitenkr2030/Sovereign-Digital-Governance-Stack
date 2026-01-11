import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Zap, Activity, Server, Database, Wifi, WifiOff } from 'lucide-react';
import { Card } from '../../ui/Card';
import { Badge } from '../../ui/Badge';
import { Button } from '../../ui/Button';
import { LoadingSpinner } from '../../ui/LoadingSpinner';
import { useWebSocket } from '../../../hooks/useWebSocket';
import './RealTimeDataFeed.css';

export interface DataFeedItem {
  id: string;
  type: 'metric' | 'event' | 'alert' | 'transaction' | 'system';
  source: string;
  timestamp: Date;
  data: Record<string, unknown>;
  priority?: 'high' | 'normal' | 'low';
}

export interface RealTimeDataFeedProps {
  /** WebSocket URL for data stream */
  wsUrl?: string;
  /** Polling interval in ms (for fallback) */
  pollInterval?: number;
  /** Maximum items to display */
  maxItems?: number;
  /** Filter by feed types */
  feedTypes?: DataFeedItem['type'][];
  /** Auto-scroll to new items */
  autoScroll?: boolean;
  /** Pause updates */
  isPaused?: boolean;
  /** Callback when item is clicked */
  onItemClick?: (item: DataFeedItem) => void;
  /** Custom data source for manual control */
  data?: DataFeedItem[];
  /** Show connection status */
  showConnectionStatus?: boolean;
  /** Custom class name */
  className?: string;
}

/**
 * Get icon for feed type
 */
const getFeedTypeIcon = (type: DataFeedItem['type']) => {
  switch (type) {
    case 'metric':
      return <Activity size={14} />;
    case 'event':
      return <Zap size={14} />;
    case 'alert':
      return <Zap size={14} className="feed-icon-alert" />;
    case 'transaction':
      return <Database size={14} />;
    case 'system':
      return <Server size={14} />;
    default:
      return <Activity size={14} />;
  }
};

/**
 * Get color for feed type
 */
const getFeedTypeColor = (type: DataFeedItem['type'], priority?: DataFeedItem['priority']): string => {
  if (priority === 'high' || type === 'alert') {
    return 'var(--color-danger)';
  }
  switch (type) {
    case 'metric':
      return 'var(--color-primary)';
    case 'event':
      return 'var(--color-warning)';
    case 'transaction':
      return 'var(--color-success)';
    case 'system':
      return 'var(--color-info)';
    default:
      return 'var(--color-text-muted)';
  }
};

/**
 * Format timestamp
 */
const formatTimestamp = (date: Date): string => {
  return new Date(date).toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  });
};

/**
 * RealTimeDataFeed Component
 * 
 * Displays streaming data updates with connection management and filtering.
 */
export const RealTimeDataFeed: React.FC<RealTimeDataFeedProps> = ({
  wsUrl,
  pollInterval = 5000,
  maxItems = 50,
  feedTypes,
  autoScroll = true,
  isPaused = false,
  onItemClick,
  data: externalData,
  showConnectionStatus = true,
  className = '',
}) => {
  const [feedItems, setFeedItems] = useState<DataFeedItem[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const lastItemRef = useRef<HTMLDivElement>(null);

  // WebSocket connection
  const { lastMessage, isOpen, error } = useWebSocket(wsUrl, {
    enabled: !!wsUrl && !isPaused,
    reconnectInterval: 3000,
  });

  // Handle WebSocket messages
  useEffect(() => {
    if (lastMessage && !isPaused) {
      try {
        const parsed = JSON.parse(lastMessage.data);
        if (parsed.type === 'FEED_UPDATE') {
          const newItem: DataFeedItem = {
            id: `feed-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
            type: parsed.data.type || 'metric',
            source: parsed.data.source || 'Unknown',
            timestamp: new Date(),
            data: parsed.data.payload || {},
            priority: parsed.data.priority,
          };
          
          setFeedItems((prev) => {
            const updated = [newItem, ...prev];
            return updated.slice(0, maxItems);
          });
        }
      } catch (e) {
        console.error('Failed to parse feed message:', e);
      }
    }
  }, [lastMessage, isPaused, maxItems]);

  // Connection status
  useEffect(() => {
    setIsConnected(isOpen);
  }, [isOpen]);

  useEffect(() => {
    if (error) {
      setConnectionError(error.message);
    }
  }, [error]);

  // Auto-scroll to top when new items arrive
  useEffect(() => {
    if (autoScroll && lastItemRef.current) {
      lastItemRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [feedItems, autoScroll]);

  // Pause/resume handler
  const handleTogglePause = useCallback(() => {
    // This would typically update a parent state
  }, []);

  // Clear feed handler
  const handleClear = useCallback(() => {
    setFeedItems([]);
  }, []);

  // Filter items
  const filteredItems = feedTypes
    ? feedItems.filter((item) => feedTypes.includes(item.type))
    : feedItems;

  return (
    <Card className={`real-time-feed ${className}`}>
      {/* Header */}
      <div className="feed-header">
        <div className="feed-title-row">
          <h3 className="feed-title">Real-Time Data Feed</h3>
          {showConnectionStatus && (
            <div className={`feed-connection-status ${isConnected ? 'feed-connected' : 'feed-disconnected'}`}>
              {isConnected ? <Wifi size={14} /> : <WifiOff size={14} />}
              <span>{isConnected ? 'Connected' : 'Disconnected'}</span>
            </div>
          )}
        </div>

        <div className="feed-controls">
          <Badge variant={isConnected ? 'success' : 'danger'} size="sm">
            {filteredItems.length} events
          </Badge>
          <div className="feed-buttons">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleTogglePause}
              title={isPaused ? 'Resume' : 'Pause'}
            >
              {isPaused ? 'Resume' : 'Pause'}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleClear}
              disabled={filteredItems.length === 0}
            >
              Clear
            </Button>
          </div>
        </div>
      </div>

      {/* Error Display */}
      {connectionError && (
        <div className="feed-error">
          <span>Connection error: {connectionError}</span>
          <Badge variant="danger" size="sm">Reconnecting...</Badge>
        </div>
      )}

      {/* Feed List */}
      <div className="feed-list" ref={scrollRef}>
        {filteredItems.length === 0 ? (
          <div className="feed-empty">
            <Activity size={32} />
            <p>No data events</p>
          </div>
        ) : (
          filteredItems.map((item, index) => {
            const isNew = index === 0 && 
              new Date().getTime() - new Date(item.timestamp).getTime() < 5000;
            
            return (
              <div
                key={item.id}
                ref={index === filteredItems.length - 1 ? lastItemRef : null}
                className={`feed-item ${isNew ? 'feed-item-new' : ''} ${item.priority === 'high' ? 'feed-item-priority' : ''}`}
                onClick={() => onItemClick?.(item)}
              >
                <div 
                  className="feed-item-icon"
                  style={{ color: getFeedTypeColor(item.type, item.priority) }}
                >
                  {getFeedTypeIcon(item.type)}
                </div>
                
                <div className="feed-item-content">
                  <div className="feed-item-header">
                    <span className="feed-item-source">{item.source}</span>
                    <span className="feed-item-time">{formatTimestamp(item.timestamp)}</span>
                  </div>
                  
                  <div className="feed-item-data">
                    {Object.entries(item.data).slice(0, 3).map(([key, value]) => (
                      <span key={key} className="feed-data-pair">
                        <span className="feed-data-key">{key}:</span>
                        <span className="feed-data-value">{String(value)}</span>
                      </span>
                    ))}
                  </div>
                </div>

                <Badge 
                  variant={item.type === 'alert' ? 'danger' : item.type === 'system' ? 'info' : 'default'} 
                  size="sm"
                >
                  {item.type}
                </Badge>
              </div>
            );
          })
        )}
      </div>

      {/* Footer Stats */}
      <div className="feed-footer">
        <div className="feed-stats">
          <span className="feed-stat">
            <Activity size={12} />
            {filteredItems.filter(i => i.type === 'metric').length} metrics
          </span>
          <span className="feed-stat">
            <Zap size={12} />
            {filteredItems.filter(i => i.type === 'event').length} events
          </span>
          <span className="feed-stat">
            <Database size={12} />
            {filteredItems.filter(i => i.type === 'transaction').length} transactions
          </span>
        </div>
      </div>
    </Card>
  );
};

export default RealTimeDataFeed;
