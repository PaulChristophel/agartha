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
  include_accepted?: boolean;
  include_denied?: boolean;
}

interface UseKeyDict {
  rejectedMinions: string[];
  isLoading: boolean;
  error: Error | null;
  rejectKeys: (keyDict: KeyDict) => void;
}

const useKeyDict = (): UseKeyDict => {
  const [rejectedMinions, setAcceptedMinions] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const rejectKeys = async (keyDict: KeyDict) => {
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
          '/api/v1/salt_keys/minion_keys/reject',
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
        console.warn('Failed to reject minion keys in salt_keys, falling back to Salt', dbErr);
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
        fun: 'key.reject_dict',
        match: keyDict.match,
        include_accepted: keyDict.include_accepted,
        include_denied: keyDict.include_denied,
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
    rejectedMinions,
    isLoading,
    error,
    rejectKeys,
  };
};

export default useKeyDict;
