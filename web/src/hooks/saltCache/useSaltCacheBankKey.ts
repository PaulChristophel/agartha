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

const useCacheBankKey = (bank: string, psqlKey: string) => {
  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.saltCache.detail(bank, psqlKey),
    queryFn: async ({ signal }) => {
      const encodedBank = encodeURIComponent(bank);
      const encodedPsqlKey = encodeURIComponent(psqlKey);
      const response = await apiClient.get<CacheData>(
        `/api/v1/salt_cache/${encodedBank}/${encodedPsqlKey}`,
        { signal }
      );
      return response.data;
    },
    enabled: Boolean(bank && psqlKey),
  });

  return {
    alterTime: data?.alter_time ?? '',
    cacheData: data?.data ?? {},
    id: data?.id ?? '00000000-0000-0000-0000-000000000000',
    psqlKey,
    bank,
    isLoading,
    error: error as Error | null,
  };
};

export default useCacheBankKey;
