import { useState, useEffect, useCallback, useRef } from 'react';

export interface WebSocketOptions {
  /** Enable the connection */
  enabled?: boolean;
  /** Reconnect interval in milliseconds */
  reconnectInterval?: number;
  /** Maximum reconnection attempts */
  maxReconnectAttempts?: number;
  /** Connection timeout in milliseconds */
  connectionTimeout?: number;
  /** Callback when connection opens */
  onOpen?: () => void;
  /** Callback when connection closes */
  onClose?: (event: CloseEvent) => void;
  /** Callback when error occurs */
  onError?: (error: Event) => void;
  /** Custom WebSocket constructor */
  webSocket?: typeof WebSocket;
}

export interface UseWebSocketReturn {
  /** Last received message */
  lastMessage: MessageEvent | null;
  /** Connection status */
  isOpen: boolean;
  /** Error state */
  error: Error | null;
  /** Ready state */
  readyState: number;
  /** Send a message */
  sendMessage: (data: string | object) => void;
  /** Close the connection */
  close: () => void;
  /** Reconnect manually */
  reconnect: () => void;
}

/**
 * useWebSocket Hook
 * 
 * Manages WebSocket connection with automatic reconnection and state management.
 */
export function useWebSocket(
  url: string | null,
  options: WebSocketOptions = {}
): UseWebSocketReturn {
  const {
    enabled = true,
    reconnectInterval = 5000,
    maxReconnectAttempts = 5,
    connectionTimeout = 10000,
    onOpen,
    onClose,
    onError,
    webSocket = WebSocket,
  } = options;

  const [lastMessage, setLastMessage] = useState<MessageEvent | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [readyState, setReadyState] = useState(webSocket.CONNECTING);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const timeoutRef = useRef<number | null>(null);
  const reconnectTimeoutRef = useRef<number | null>(null);

  // Clear timeouts
  const clearTimeouts = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
  }, []);

  // Close connection
  const close = useCallback(() => {
    clearTimeouts();
    reconnectAttemptsRef.current = maxReconnectAttempts; // Prevent auto-reconnect
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, [clearTimeouts, maxReconnectAttempts]);

  // Send message
  const sendMessage = useCallback((data: string | object) => {
    if (wsRef.current && wsRef.current.readyState === webSocket.OPEN) {
      const message = typeof data === 'string' ? data : JSON.stringify(data);
      wsRef.current.send(message);
    } else {
      console.warn('WebSocket is not connected');
    }
  }, [webSocket]);

  // Reconnect
  const reconnect = useCallback(() => {
    reconnectAttemptsRef.current = 0;
    close();
    // Small delay before reconnecting
    setTimeout(() => {
      // Connection will be established by the useEffect
    }, 100);
  }, [close]);

  // Establish connection
  useEffect(() => {
    if (!url || !enabled) return;

    clearTimeouts();

    const connect = () => {
      try {
        const ws = new webSocket(url);
        wsRef.current = ws;

        // Set connection timeout
        timeoutRef.current = window.setTimeout(() => {
          if (ws.readyState !== webSocket.OPEN) {
            ws.close();
            setError(new Error('Connection timeout'));
          }
        }, connectionTimeout);

        ws.onopen = () => {
          clearTimeouts();
          setIsOpen(true);
          setReadyState(webSocket.OPEN);
          setError(null);
          reconnectAttemptsRef.current = 0;
          onOpen?.();
        };

        ws.onmessage = (event) => {
          setLastMessage(event);
        };

        ws.onclose = (event) => {
          clearTimeouts();
          setIsOpen(false);
          setReadyState(webSocket.CLOSED);
          onClose?.(event);

          // Attempt reconnection if not intentionally closed
          if (reconnectAttemptsRef.current < maxReconnectAttempts) {
            reconnectAttemptsRef.current++;
            setError(new Error(`Reconnecting (${reconnectAttemptsRef.current}/${maxReconnectAttempts})...`));
            
            reconnectTimeoutRef.current = window.setTimeout(() => {
              connect();
            }, reconnectInterval);
          } else {
            setError(new Error('Max reconnection attempts reached'));
          }
        };

        ws.onerror = (event) => {
          setError(new Error('WebSocket error'));
          onError?.(event);
        };

        setReadyState(webSocket.CONNECTING);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to create WebSocket'));
      }
    };

    connect();

    return () => {
      clearTimeouts();
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [url, enabled, webSocket, reconnectInterval, maxReconnectAttempts, connectionTimeout, onOpen, onClose, onError, clearTimeouts]);

  return {
    lastMessage,
    isOpen,
    error,
    readyState,
    sendMessage,
    close,
    reconnect,
  };
}

// Connection status enum for readability
export const WebSocketReadyState = {
  CONNECTING: 0,
  OPEN: 1,
  CLOSING: 2,
  CLOSED: 3,
} as const;

export default useWebSocket;
