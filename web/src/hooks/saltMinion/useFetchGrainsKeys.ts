// src/hooks/saltMinion/useFetchGrainsKeys.ts
import axios from 'axios';
import { useState, useEffect } from 'react';

interface PaginatedResponse {
  paging: {
    page: number;
    per_page: number;
    total: number;
  };
  results: string[];
}

const useFetchGrainsKeys = (authToken: string, likeIncludes: string, page: number) => {
  const [grainsKeys, setGrainsKeys] = useState<string[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchGrainsKeys = async () => {
      setLoading(true);
      setError(null);

      try {
        const response = await axios.get<PaginatedResponse>('/api/v1/salt_minion/grains_keys', {
          headers: {
            Authorization: `${authToken}`,
          },
          params: {
            page,
            per_page: 50,
            like_includes: likeIncludes.replace(/"/g, "'"),
          },
        });
        // const keys = response.data.results; // .map(x => x.replace(/'/g,""));
        const keys = response.data.results.map((x) => x.replace(/'/g, '"'));
        setGrainsKeys(keys);
      } catch (err) {
        if (axios.isAxiosError(err)) {
          setError(err.message);
        } else {
          setError('An unexpected error occurred');
        }
      } finally {
        setLoading(false);
      }
    };

    fetchGrainsKeys();
  }, [authToken, likeIncludes, page]);

  return { grainsKeys, loading, error };
};

export default useFetchGrainsKeys;
