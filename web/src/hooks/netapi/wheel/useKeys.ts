import axios from 'axios';
import { useState, useEffect } from 'react';

import { WheelClient } from '../api/clients/wheel.ts';
import { ISaltConfigOptions } from '../api/client.ts';
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
        const authSaltString = localStorage.getItem('authSalt');
        if (!authSaltString) {
          throw new Error('Missing authSalt in local storage');
        }

        const password = localStorage.getItem('authToken');
        if (!password) {
          throw new Error('Missing authToken in local storage');
        }

        try {
          const { data } = await axios.get<IListResponse>('/api/v1/salt_keys/minion_keys', {
            headers: {
              Authorization: `${password}`,
            },
          });

          setMinions(data.minions);
          setMinionsDenied(data.minions_denied);
          setMinionsPre(data.minions_pre);
          setMinionsRejected(data.minions_rejected);
          setError(null);
          return;
        } catch (dbErr) {
          console.warn('Failed to load minion keys from salt_keys, falling back to Salt', dbErr);
        }

        const parsedAuthSalt = JSON.parse(authSaltString);
        const { token, expire } = parsedAuthSalt;

        const config: ISaltConfigOptions = {
          endpoint: 'api/v1/netapi',
          token,
          password,
          expire,
          logger: console,
        };

        const wheelClient = new WheelClient(config);

        const response = await wheelClient.exec<IListRequest, IListResponse>({
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
