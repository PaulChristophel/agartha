import { useRef, useState, useEffect, useCallback } from 'react';

import { apiClient as axios } from 'src/api/client.ts';

interface RefreshStatus {
  status: 'available' | 'failed' | 'pending' | 'success';
  message: string;
}

const pollIntervalMilliseconds = 250;
const refreshTimeoutMilliseconds = 120_000;

const useRefreshKeys = () => {
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [revision, setRevision] = useState(0);
  const controllerRef = useRef<AbortController | null>(null);
  const inFlightRef = useRef<Promise<void> | null>(null);

  useEffect(
    () => () => {
      controllerRef.current?.abort();
    },
    []
  );

  const refreshKeys = useCallback((): Promise<void> => {
    if (inFlightRef.current) return inFlightRef.current;

    const controller = new AbortController();
    controllerRef.current = controller;
    setIsRefreshing(true);
    setError(null);
    setMessage('Refreshing dropdown lists...');

    const refresh = async () => {
      try {
        await axios.post<RefreshStatus>(
          '/api/v1/salt_minion/keys/refresh',
          {},
          {
            signal: controller.signal,
          }
        );

        const deadline = Date.now() + refreshTimeoutMilliseconds;
        while (Date.now() < deadline) {
          await new Promise<void>((resolve) => {
            window.setTimeout(resolve, pollIntervalMilliseconds);
          });
          if (controller.signal.aborted) return;

          const response = await axios.get<RefreshStatus>('/api/v1/salt_minion/keys/refresh', {
            signal: controller.signal,
          });
          if (response.data.status !== 'pending') {
            if (response.data.status === 'failed') {
              throw new Error(response.data.message);
            }
            setMessage('Dropdown lists refreshed');
            setRevision((current) => current + 1);
            return;
          }
        }

        throw new Error('Timed out waiting for the dropdown lists to refresh');
      } catch (err) {
        if (controller.signal.aborted) return;
        setMessage(null);
        setError(err instanceof Error ? err.message : 'Failed to refresh dropdown lists');
      } finally {
        if (!controller.signal.aborted) setIsRefreshing(false);
        if (controllerRef.current === controller) controllerRef.current = null;
        inFlightRef.current = null;
      }
    };

    inFlightRef.current = refresh();
    return inFlightRef.current;
  }, []);

  return { isRefreshing, error, message, revision, refreshKeys };
};

export default useRefreshKeys;
