import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';

import { queryKeys } from 'src/api/queryKeys.ts';
import { apiClient, isApiError } from 'src/api/client.ts';

export interface JobReturn {
  id: string;
  jid: string;
  fun: string;
  success: boolean;
  alter_time: string | null;
  return: unknown;
  full_ret: unknown;
}

export interface JobDetails {
  jid: string;
  load: unknown;
  returns: JobReturn[];
  targeted_count: number | null;
  returned_count: number;
  successful_count: number;
  failed_count: number;
  pending_count: number | null;
  started_at: string | null;
  last_return_at: string | null;
}

export default function useJobDetails(jid: string, submittedAt?: number) {
  const [autoRefresh, setAutoRefresh] = useState(true);
  useEffect(() => {
    setAutoRefresh(true);
    const timer = window.setTimeout(() => setAutoRefresh(false), 120_000);
    return () => window.clearTimeout(timer);
  }, [jid]);

  const query = useQuery({
    queryKey: queryKeys.jobs.detail(jid),
    queryFn: async ({ signal }) => {
      const response = await apiClient.get<JobDetails>(`/api/v1/jobs/${encodeURIComponent(jid)}`, {
        signal,
      });
      return response.data;
    },
    enabled: Boolean(jid),
    staleTime: 5_000,
    refetchOnWindowFocus: false,
    refetchIntervalInBackground: false,
    refetchInterval: (current) => (autoRefresh && !current.state.error ? 5_000 : false),
    retry: (count, error) =>
      Boolean(
        submittedAt &&
        Date.now() - submittedAt < 30_000 &&
        count < 5 &&
        isApiError(error) &&
        error.status === 404
      ),
    retryDelay: 3_000,
  });

  return { ...query, autoRefresh };
}
