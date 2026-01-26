import { useEffect, useCallback, useRef } from 'react';

export interface SSENotification {
  type: 'status_change' | 'overdue';
  taskId: number;
  taskTitle: string;
  newStatus?: string;
  oldStatus?: string;
}

export interface SSEMessage {
  action: 'reload' | 'notify';
  notification?: SSENotification;
}

export function useSSE(onReload: () => void, onNotification?: (notification: SSENotification) => void) {
  const eventSourceRef = useRef<EventSource | null>(null);
  const onReloadRef = useRef(onReload);
  const onNotificationRef = useRef(onNotification);

  // Keep the callback ref updated
  onReloadRef.current = onReload;
  onNotificationRef.current = onNotification;

  const connect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const eventSource = new EventSource('/events');
    eventSourceRef.current = eventSource;

    eventSource.onmessage = (event) => {
      // Handle legacy reload message
      if (event.data === 'reload') {
        onReloadRef.current();
        return;
      }

      // Try to parse JSON message
      try {
        const msg: SSEMessage = JSON.parse(event.data);
        if (msg.action === 'reload') {
          onReloadRef.current();
        }
        if (msg.action === 'notify' && msg.notification && onNotificationRef.current) {
          onNotificationRef.current(msg.notification);
        }
      } catch {
        // Ignore parse errors for non-JSON messages
        if (event.data !== 'connected') {
          console.log('Unknown SSE message:', event.data);
        }
      }
    };

    eventSource.onerror = () => {
      console.log('SSE connection lost, retrying...');
      eventSource.close();
      // Reconnect after 3 seconds
      setTimeout(connect, 3000);
    };

    return eventSource;
  }, []);

  useEffect(() => {
    const eventSource = connect();

    return () => {
      eventSource.close();
    };
  }, [connect]);
}
