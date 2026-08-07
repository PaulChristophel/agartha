import { useQuery } from '@tanstack/react-query';

import { apiClient } from 'src/api/client.ts';
import { queryKeys } from 'src/api/queryKeys.ts';

interface EventData {
  alter_time: string;
  data: Record<string, unknown>;
  id: number;
  master_id: string;
  tag: string;
}

const useEvent = (id: number) => {
  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.saltEvents.detail(id),
    queryFn: async ({ signal }) => {
      const response = await apiClient.get<EventData>(`/api/v1/salt_event/${id}`, { signal });
      return response.data;
    },
    enabled: Boolean(id),
  });

  return {
    alterTime: data?.alter_time ?? '',
    eventData: data?.data ?? {},
    id,
    masterID: data?.master_id ?? '',
    tag: data?.tag ?? '',
    isLoading,
    error: error as Error | null,
  };
};

export default useEvent;
