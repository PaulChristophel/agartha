import { useState, useEffect } from 'react';

import { sessionStore } from 'src/api/session.ts';
import { apiClient as axios } from 'src/api/client.ts';

interface KeysResponse {
  [key: string]: string;
}

interface UseKeys {
  minion: string;
  hash: string;
  isLoading: boolean;
  error: Error | null;
}

const useKeys = (id: string): UseKeys => {
  const [minion, setMinion] = useState<string>('');
  const [hash, setHash] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchKeyData = async () => {
      setIsLoading(true);
      try {
        const token = sessionStore.getSnapshot().authSalt?.token;
        if (!token) throw new Error('Missing Salt authentication');

        const { data } = await axios.get<KeysResponse>(`/api/v1/netapi/key/${id}`, {
          params: { token },
        });

        const [minionKey, minionHash] = Object.entries(data)[0];
        setMinion(minionKey);
        setHash(minionHash);
        setError(null);
      } catch (err) {
        setError(err as Error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchKeyData();
  }, [id]);

  return {
    minion,
    hash,
    isLoading,
    error,
  };
};

export default useKeys;
