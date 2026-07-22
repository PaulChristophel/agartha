import { useState } from 'react';

import { executeWheel } from 'src/api/salt.ts';
import { apiClient as axios } from 'src/api/client.ts';

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
      try {
        const { data } = await axios.post<IResponse>(
          '/api/v1/salt_keys/minion_keys/delete',
          keyDict
        );

        setAcceptedMinions(data.minions);
        setError(null);
        return;
      } catch (dbErr) {
        console.warn('Failed to delete minion keys in salt_keys, falling back to Salt', dbErr);
      }

      const response = await executeWheel<IDictRequest, IResponse>({
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
