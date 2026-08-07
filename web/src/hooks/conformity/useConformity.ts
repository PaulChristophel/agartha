import { useQuery } from '@tanstack/react-query';

import { apiClient } from 'src/api/client.ts';
import { queryKeys } from 'src/api/queryKeys.ts';

interface ConformityData {
  alter_time: string;
  true_count: number;
  false_count: number;
  changed_count: number;
  unchanged_count: number;
  success: boolean;
  id: string;
}

const useSaltConformity = (id: string) => {
  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.conformity.detail(id),
    queryFn: async ({ signal }) => {
      const response = await apiClient.get<ConformityData>(`/api/v1/conformity/${id}`, { signal });
      return response.data;
    },
    enabled: Boolean(id),
  });

  return {
    alterTime: data?.alter_time ?? '',
    trueCount: data?.true_count ?? 0,
    falseCount: data?.false_count ?? 0,
    changedCount: data?.changed_count ?? 0,
    unchangedCount: data?.unchanged_count ?? 0,
    success: data?.success ?? false,
    conformityId: data?.id ?? '',
    isLoading,
    error: error as Error | null,
  };
};

export default useSaltConformity;
