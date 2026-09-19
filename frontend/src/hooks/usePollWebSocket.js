import { useState, useEffect, useRef, useCallback } from 'react';
import { api } from '../services/api';

export function usePollWebSocket(pollId, onUpdate) {
  const [connectionStatus, setConnectionStatus] = useState('disconnected'); // 'connecting' | 'connected' | 'disconnected'
  const [lastMessageTime, setLastMessageTime] = useState(null);
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const isMountedRef = useRef(true);

  const connect = useCallback(() => {
    if (!pollId) return;

    try {
      setConnectionStatus('connecting');
      const wsUrl = api.getWebSocketUrl(pollId);
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        if (!isMountedRef.current) return;
        setConnectionStatus('connected');
        console.log(`[WebSocket] Connected to poll: ${pollId}`);
      };

      ws.onmessage = (event) => {
        if (!isMountedRef.current) return;
        try {
          const data = JSON.parse(event.data);
          setLastMessageTime(new Date());
          if (onUpdate) {
            onUpdate(data);
          }
        } catch (err) {
          console.error('[WebSocket] Error parsing message:', err);
        }
      };

      ws.onclose = () => {
        if (!isMountedRef.current) return;
        setConnectionStatus('disconnected');
        console.log(`[WebSocket] Disconnected from poll: ${pollId}. Reconnecting in 3s...`);
        reconnectTimeoutRef.current = setTimeout(() => {
          if (isMountedRef.current) {
            connect();
          }
        }, 3000);
      };

      ws.onerror = (err) => {
        console.error('[WebSocket] Error occurred:', err);
        ws.close();
      };
    } catch (err) {
      console.error('[WebSocket] Initialization error:', err);
      setConnectionStatus('disconnected');
    }
  }, [pollId, onUpdate]);

  useEffect(() => {
    isMountedRef.current = true;
    connect();

    return () => {
      isMountedRef.current = false;
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  return { connectionStatus, lastMessageTime };
}
