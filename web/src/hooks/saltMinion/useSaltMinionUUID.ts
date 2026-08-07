import { useQuery } from '@tanstack/react-query';

import { apiClient } from 'src/api/client.ts';
import { queryKeys } from 'src/api/queryKeys.ts';

interface MinionData {
  alter_time: string;
  grains: Record<string, unknown>;
  pillar: Record<string, unknown>;
  minion_id: string;
  id: string;
}

const useMinionID = (id: string) => {
  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.saltMinions.byUUID(id),
    queryFn: async ({ signal }) => {
      const response = await apiClient.get<MinionData>(`/api/v1/salt_minion/uuid/${id}`, {
        signal,
      });
      return response.data;
    },
    enabled: Boolean(id),
  });

  return {
    alterTime: data?.alter_time ?? '',
    grains: data?.grains ?? {},
    pillar: data?.pillar ?? {},
    id,
    minionID: data?.minion_id ?? '',
    isLoading,
    error: error as Error | null,
  };
};

export default useMinionID;
