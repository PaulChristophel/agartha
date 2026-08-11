import { useState, useEffect } from 'react';

import { toColonNotation } from 'src/utils/grainKeys.ts';

import { apiClient as axios } from 'src/api/client.ts';

interface PaginatedResponse {
  paging: {
    page: number;
    per_page: number;
    total: number;
  };
  results: string[];
}

const useFetchPillarKeys = (likeIncludes: string, page: number) => {
  const [pillarKeys, setPillarKeys] = useState<string[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    const fetchPillarKeys = async () => {
      setLoading(true);
      setError(null);

      try {
        const response = await axios.get<PaginatedResponse>('/api/v1/salt_minion/pillar_keys', {
          params: {
            page,
            per_page: 50,
            like_includes: likeIncludes,
          },
          signal: controller.signal,
        });
        const keys = response.data.results.map((key) => toColonNotation(key));
        setPillarKeys(keys);
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

    void fetchPillarKeys();
    return () => controller.abort();
  }, [likeIncludes, page]);

  return { pillarKeys, loading, error };
};

export default useFetchPillarKeys;
