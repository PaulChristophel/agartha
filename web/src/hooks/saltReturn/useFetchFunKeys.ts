// src/hooks/saltReturn/useFetchFunKeys.ts
import { useState, useEffect } from 'react';

import { apiClient as axios } from 'src/api/client.ts';

interface PaginatedResponse {
  paging: {
    page: number;
    per_page: number;
    total: number;
  };
  results: string[];
}

const useFetchFunKeys = (likeIncludes: string, page: number, since?: string, until?: string) => {
  const [funKeys, setFunKeys] = useState<string[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    const fetchFunKeys = async () => {
      setLoading(true);
      setError(null);

      try {
        const params = new URLSearchParams();
        params.append('page', String(page));
        params.append('per_page', '50');
        params.append('like_includes', likeIncludes);
        if (since) params.append('since', new Date(since).toISOString());
        if (until) params.append('until', new Date(until).toISOString());

        const response = await axios.get<PaginatedResponse>('/api/v1/salt_return/fun', {
          params,
          signal: controller.signal,
        });

        const keys = response.data.results.map((x) => x);
        setFunKeys(keys);
      } catch (err) {
        if (controller.signal.aborted) return;
        if (axios.isAxiosError(err)) {
          setError(err.message);
        } else {
          setError('An unexpected error occurred');
        }
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    };

    void fetchFunKeys();
    return () => controller.abort();
  }, [likeIncludes, page, since, until]);

  return { funKeys, loading, error };
};

export default useFetchFunKeys;
