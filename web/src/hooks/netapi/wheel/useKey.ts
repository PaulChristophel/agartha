import { useQuery } from '@tanstack/react-query';

import { queryKeys } from 'src/api/queryKeys.ts';
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
  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.saltKeys.detail(id),
    queryFn: async ({ signal }) => {
      const response = await axios.get<KeysResponse>(`/api/v1/netapi/key/${id}`, { signal });
      return Object.entries(response.data)[0];
    },
  });

  return {
    minion: data?.[0] ?? '',
    hash: data?.[1] ?? '',
    isLoading,
    error: error as Error | null,
  };
};

export default useKeys;
