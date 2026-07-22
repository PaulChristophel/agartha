import { useState, useEffect } from 'react';

import { executeWheel } from 'src/api/salt.ts';
import { apiClient as axios } from 'src/api/client.ts';

import { IListRequest, IListResponse } from '../api/modules/wheel/key.ts';

interface UseKeys {
  minions: string[];
  minionsDenied: string[];
  minionsPre: string[];
  minionsRejected: string[];
  isLoading: boolean;
  error: Error | null;
}

const useKeys = (): UseKeys => {
  const [minions, setMinions] = useState<string[]>([]);
  const [minionsDenied, setMinionsDenied] = useState<string[]>([]);
  const [minionsPre, setMinionsPre] = useState<string[]>([]);
  const [minionsRejected, setMinionsRejected] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchKeysData = async () => {
      setIsLoading(true);
      try {
        try {
          const { data } = await axios.get<IListResponse>('/api/v1/salt_keys/minion_keys');

          setMinions(data.minions);
          setMinionsDenied(data.minions_denied);
          setMinionsPre(data.minions_pre);
          setMinionsRejected(data.minions_rejected);
          setError(null);
          return;
        } catch (dbErr) {
          console.warn('Failed to load minion keys from salt_keys, falling back to Salt', dbErr);
        }

        const response = await executeWheel<IListRequest, IListResponse>({
          fun: 'key.list_all',
        });

        setMinions(response.minions);
        setMinionsDenied(response.minions_denied);
        setMinionsPre(response.minions_pre);
        setMinionsRejected(response.minions_rejected);
        setError(null);
      } catch (err) {
        setError(err as Error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchKeysData();
  }, []);

  return {
    minions,
    minionsDenied,
    minionsPre,
    minionsRejected,
    isLoading,
    error,
  };
};

export default useKeys;
