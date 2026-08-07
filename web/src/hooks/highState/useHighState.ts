import { useQuery } from '@tanstack/react-query';

import { apiClient } from 'src/api/client.ts';
import { queryKeys } from 'src/api/queryKeys.ts';

interface HighStateData {
  alter_time: string;
  full_ret: Record<string, unknown>;
  fun: string;
  id: string;
  jid: string;
  return: Record<string, unknown>;
  success: boolean;
}

const useHighState = (id: string, loadReturn: boolean, loadFullRet: boolean) => {
  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.highStates.detail(id, loadReturn, loadFullRet),
    queryFn: async ({ signal }) => {
      const response = await apiClient.get<HighStateData>(`/api/v1/high_state/${id}`, {
        params: { load_return: loadReturn, load_full_ret: loadFullRet },
        signal,
      });
      return response.data;
    },
    enabled: Boolean(id),
  });

  return {
    alterTime: data?.alter_time ?? '',
    fullRet: data?.full_ret ?? {},
    fun: data?.fun ?? '',
    returnId: data?.id ?? '',
    returnJid: data?.jid ?? '',
    returnData: data?.return ?? {},
    success: data?.success ?? false,
    isLoading,
    error: error as Error | null,
  };
};

export default useHighState;
