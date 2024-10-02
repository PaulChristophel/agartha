import axios from 'axios';
import { useState, useEffect } from 'react';

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
    const authToken = localStorage.getItem('authToken');
    const authSaltString = localStorage.getItem('authSalt');

    const fetchKeyData = async () => {
      setIsLoading(true);
      try {
        if (!authToken || !authSaltString) {
          throw new Error('Missing authToken or authSalt in local storage');
        }

        const parsedAuthSalt = JSON.parse(authSaltString);
        const { token } = parsedAuthSalt;

        const { data } = await axios.get<KeysResponse>(`/api/v1/netapi/key/${id}?token=${token}`, {
          headers: {
            Authorization: `${authToken}`,
          },
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
