import { useEffect, useCallback, useRef } from 'react';

export function useSSE(onMessage: () => void) {
  const eventSourceRef = useRef<EventSource | null>(null);
  const onMessageRef = useRef(onMessage);

  // Keep the callback ref updated
  onMessageRef.current = onMessage;

  const connect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const eventSource = new EventSource('/events');
    eventSourceRef.current = eventSource;

    eventSource.onmessage = (event) => {
      if (event.data === 'reload') {
        onMessageRef.current();
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
