import axios from 'axios';
import { useState } from 'react';

import { WheelClient } from '../api/clients/wheel.ts';
import { ISaltConfigOptions } from '../api/client.ts';
import { IResponse, IDictRequest } from '../api/modules/wheel/key.ts';

interface KeyDict {
  match: {
    minions?: string[];
    minions_denied?: string[];
    minions_pre?: string[];
    minions_rejected?: string[];
  };
}

interface UseKeyDict {
  deletedMinions: string[];
  isLoading: boolean;
  error: Error | null;
  deleteKeys: (keyDict: KeyDict) => void;
}

const useKeyDict = (): UseKeyDict => {
  const [deletedMinions, setAcceptedMinions] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const deleteKeys = async (keyDict: KeyDict) => {
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
        const { data } = await axios.post<IResponse>(
          '/api/v1/salt_keys/minion_keys/delete',
          keyDict,
          {
            headers: {
              Authorization: `${password}`,
            },
          }
        );

        setAcceptedMinions(data.minions);
        setError(null);
        return;
      } catch (dbErr) {
        console.warn('Failed to delete minion keys in salt_keys, falling back to Salt', dbErr);
      }

      const parsedAuthSalt = JSON.parse(authSaltString);
      const { token, expire } = parsedAuthSalt;

      const config: ISaltConfigOptions = {
        endpoint: 'api/v1/netapi',
        token,
        password,
        expire,
        logger: console, // Assuming console implements Logger interface
      };

      const wheelClient = new WheelClient(config);

      const response = await wheelClient.exec<IDictRequest, IResponse>({
        fun: 'key.delete_dict',
        match: keyDict.match,
      });

      setAcceptedMinions(response.minions);
      setError(null);
    } catch (err) {
      setError(err as Error);
    } finally {
      setIsLoading(false);
    }
  };

  return {
    deletedMinions,
    isLoading,
    error,
    deleteKeys,
  };
};

export default useKeyDict;
