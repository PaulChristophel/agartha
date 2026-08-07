import { useQuery } from '@tanstack/react-query';

import { apiClient } from 'src/api/client.ts';
import { queryKeys } from 'src/api/queryKeys.ts';

interface ReturnData {
  alter_time: string;
  full_ret: Record<string, unknown>;
  fun: string;
  id: string;
  jid: string;
  return: Record<string, unknown>;
  success: boolean;
}

const useSaltReturn = (jid: string, id: string, loadReturn: boolean, loadFullRet: boolean) => {
  const { data, isLoading, error } = useQuery({
    queryKey: queryKeys.saltReturns.detail(jid, id, loadReturn, loadFullRet),
    queryFn: async ({ signal }) => {
      const response = await apiClient.get<ReturnData>(`/api/v1/salt_return/${jid}/${id}`, {
        params: { load_return: loadReturn, load_full_ret: loadFullRet },
        signal,
      });
      return response.data;
    },
    enabled: Boolean(jid && id),
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

export default useSaltReturn;
