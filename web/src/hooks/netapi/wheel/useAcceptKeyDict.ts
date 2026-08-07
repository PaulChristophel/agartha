import { useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';

import { executeWheel } from 'src/api/salt.ts';
import { queryKeys } from 'src/api/queryKeys.ts';
import { apiClient as axios } from 'src/api/client.ts';

import { IResponse, IDictRequest } from '../api/modules/wheel/key.ts';

interface KeyDict {
  match: {
    minions: string[];
  };
  include_rejected?: boolean;
  include_denied?: boolean;
}

interface UseKeyDict {
  acceptedMinions: string[];
  isLoading: boolean;
  error: Error | null;
  acceptKeys: (keyDict: KeyDict) => void;
}

const useKeyDict = (): UseKeyDict => {
  const queryClient = useQueryClient();
  const [acceptedMinions, setAcceptedMinions] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const acceptKeys = async (keyDict: KeyDict) => {
    setIsLoading(true);
    try {
      try {
        const { data } = await axios.post<IResponse>(
          '/api/v1/salt_keys/minion_keys/accept',
          keyDict
        );

        setAcceptedMinions(data.minions);
        setError(null);
        await queryClient.invalidateQueries({ queryKey: queryKeys.saltKeys.all() });
        return;
      } catch (dbErr) {
        console.warn('Failed to accept minion keys in salt_keys, falling back to Salt', dbErr);
      }

      const response = await executeWheel<IDictRequest, IResponse>({
        fun: 'key.accept',
        match: keyDict.match.minions,
        include_rejected: keyDict.include_rejected,
        include_denied: keyDict.include_denied,
      });

      setAcceptedMinions(response.minions);
      setError(null);
      await queryClient.invalidateQueries({ queryKey: queryKeys.saltKeys.all() });
    } catch (err) {
      setError(err as Error);
    } finally {
      setIsLoading(false);
    }
  };

  return {
    acceptedMinions,
    isLoading,
    error,
    acceptKeys,
  };
};

export default useKeyDict;
