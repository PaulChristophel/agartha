import { useQuery } from '@tanstack/react-query';

import { apiClient } from 'src/api/client.ts';
import { queryKeys } from 'src/api/queryKeys.ts';

interface JidData {
  alter_time: string;
  jid: string;
  load: Record<string, unknown>;
}

const useJid = (jid: string) => {
  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.jids.detail(jid),
    queryFn: async ({ signal }) => {
      const response = await apiClient.get<JidData>(`/api/v1/jid/${jid}`, { signal });
      return response.data;
    },
    enabled: Boolean(jid),
  });

  return {
    alterTime: data?.alter_time ?? '',
    jid,
    load: data?.load ?? {},
    isLoading,
    error: error as Error | null,
  };
};

export default useJid;
