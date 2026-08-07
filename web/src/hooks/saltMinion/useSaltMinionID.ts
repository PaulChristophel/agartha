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

const useSaltMinionID = (minionID: string) => {
  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.saltMinions.byId(minionID),
    queryFn: async ({ signal }) => {
      const encodedMinionID = encodeURIComponent(minionID);
      const response = await apiClient.get<MinionData>(`/api/v1/salt_minion/${encodedMinionID}`, {
        signal,
      });
      return response.data;
    },
    enabled: Boolean(minionID),
  });

  return {
    alterTime: data?.alter_time ?? '',
    grains: data?.grains ?? {},
    pillar: data?.pillar ?? {},
    id: data?.id ?? '00000000-0000-0000-0000-000000000000',
    minionID,
    isLoading,
    error: error as Error | null,
  };
};

export default useSaltMinionID;
