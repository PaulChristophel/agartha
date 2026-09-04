// src/hooks/saltReturn/useFetchFunKeys.ts
import { useState, useEffect, useCallback } from 'react';

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
  const [isEmpty, setIsEmpty] = useState(false);
  const [revision, setRevision] = useState(0);
  const refresh = useCallback(() => setRevision((current) => current + 1), []);

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
        setIsEmpty(keys.length === 0);
      } catch (err) {
        if (controller.signal.aborted) return;
        if (axios.isAxiosError(err) && err.response?.status === 404) {
          setFunKeys([]);
          setIsEmpty(true);
        } else if (axios.isAxiosError(err)) {
          setIsEmpty(false);
          setError(err.message);
        } else {
          setIsEmpty(false);
          setError('An unexpected error occurred');
        }
      } finally {
        if (!controller.signal.aborted) setLoading(false);
      }
    };

    void fetchFunKeys();
    return () => controller.abort();
  }, [likeIncludes, page, since, until, revision]);

  return {
    funKeys,
    loading,
    error,
    isEmpty,
    refresh,
  };
};

export default useFetchFunKeys;
