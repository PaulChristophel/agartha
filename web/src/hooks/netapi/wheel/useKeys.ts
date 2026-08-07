import { useQuery } from '@tanstack/react-query';

import { executeWheel } from 'src/api/salt.ts';
import { queryKeys } from 'src/api/queryKeys.ts';
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
  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.saltKeys.all(),
    queryFn: async ({ signal }) => {
      try {
        const response = await axios.get<IListResponse>('/api/v1/salt_keys/minion_keys', {
          signal,
        });
        return response.data;
      } catch (dbErr) {
        if (signal.aborted) throw dbErr;
        console.warn('Failed to load minion keys from salt_keys, falling back to Salt', dbErr);
      }

      return executeWheel<IListRequest, IListResponse>({ fun: 'key.list_all' }, signal);
    },
  });

  return {
    minions: data?.minions ?? [],
    minionsDenied: data?.minions_denied ?? [],
    minionsPre: data?.minions_pre ?? [],
    minionsRejected: data?.minions_rejected ?? [],
    isLoading,
    error: error as Error | null,
  };
};

export default useKeys;
