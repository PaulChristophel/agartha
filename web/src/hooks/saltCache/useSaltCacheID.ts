import { useQuery } from '@tanstack/react-query';

import { apiClient } from 'src/api/client.ts';
import { queryKeys } from 'src/api/queryKeys.ts';

interface CacheData {
  alter_time: string;
  data: Record<string, unknown>;
  bank: string;
  psql_key: string;
  id: string;
}

const useCacheID = (id: string) => {
  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.saltCache.detail('uuid', id),
    queryFn: async ({ signal }) => {
      const response = await apiClient.get<CacheData>(`/api/v1/salt_cache/uuid/${id}`, { signal });
      return response.data;
    },
    enabled: Boolean(id),
  });

  return {
    alterTime: data?.alter_time ?? '',
    cacheData: data?.data ?? {},
    id,
    psqlKey: data?.psql_key ?? '',
    bank: data?.bank ?? '',
    isLoading,
    error: error as Error | null,
  };
};

export default useCacheID;
