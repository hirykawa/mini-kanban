import { useEffect, useState, useCallback } from 'react';

export type NotificationPermission = 'default' | 'denied' | 'granted';

export function useNotification() {
  const [permission, setPermission] = useState<NotificationPermission>(() => {
    if (typeof Notification === 'undefined') {
      return 'denied';
    }
    return Notification.permission;
  });

  const requestPermission = useCallback(async () => {
    if (typeof Notification === 'undefined') {
      return 'denied';
    }
    const result = await Notification.requestPermission();
    setPermission(result);
    return result;
  }, []);

  const showNotification = useCallback(
    (title: string, options?: NotificationOptions) => {
      // Always check the current permission state directly from the API
      if (typeof Notification === 'undefined' || Notification.permission !== 'granted') {
        return null;
      }
      return new Notification(title, options);
    },
    []
  );

  useEffect(() => {
    if (typeof Notification !== 'undefined') {
      setPermission(Notification.permission);
    }
  }, []);

  return {
    permission,
    requestPermission,
    showNotification,
    isSupported: typeof Notification !== 'undefined',
  };
}
