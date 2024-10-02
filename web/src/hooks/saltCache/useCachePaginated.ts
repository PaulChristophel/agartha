import axios from 'axios';
import { useMemo, useState, useEffect } from 'react';

interface Cache {
  alter_time: string;
  data: Record<string, unknown>;
  id: string;
  bank: string;
  psql_key: string;
}

interface Paging {
  num_pages: number;
  count: number;
}

interface ApiResponse {
  results: Cache[];
  paging: Paging;
}

interface UseCachePaginated {
  caches: Cache[];
  isLoading: boolean;
  error: Error | null;
  currentPage: number;
  rowsPerPage: number;
  setCurrentPage: (page: number) => void;
  setRowsPerPage: (rows: number) => void;
  totalPages: number;
  totalCount: number;
}

interface QueryParams {
  bank?: string;
  key?: string;
  load_data?: boolean;
  jsonpath?: string;
  jsonpath_filter?: string;
  since?: string;
  until?: string;
  order_by?: string;
}

// // Helper function to decode base64 strings
// const decodeBase64 = (base64: string | null): Record<string, unknown> => {
//   if (!base64) return {};
//   try {
//     const decodedString = atob(base64);
//     return JSON.parse(decodedString);
//   } catch (e) {
//     console.error('Failed to decode base64 string:', e);
//     return {};
//   }
// };

const useCachePaginated = (
  queryParams: QueryParams,
  page: number = 1,
  per_page: number = 10
): UseCachePaginated => {
  const [caches, setCaches] = useState<Cache[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [currentPage, setCurrentPage] = useState(page);
  const [rowsPerPage, setRowsPerPage] = useState(per_page);
  const [totalPages, setTotalPages] = useState(0);
  const [totalCount, setTotalCount] = useState(0);

  const stableQueryParams = useMemo(() => queryParams, [queryParams]);

  useEffect(() => {
    const fetchCaches = async () => {
      setIsLoading(true);
      try {
        const { bank, key, load_data, jsonpath, jsonpath_filter, since, until, order_by } =
          stableQueryParams;
        const params = new URLSearchParams();

        if (bank) params.append('bank', bank.concat('*'));
        if (key) params.append('key', key);
        if (load_data !== undefined) params.append('load_data', String(load_data));
        if (jsonpath) params.append('jsonpath', jsonpath);
        if (jsonpath_filter) params.append('jsonpath_filter', jsonpath_filter);
        if (since) params.append('since', new Date(since).toISOString());
        if (until) params.append('until', new Date(until).toISOString());
        if (order_by) params.append('order_by', order_by);
        params.append('page', String(currentPage));
        params.append('per_page', String(rowsPerPage));

        const authToken = localStorage.getItem('authToken');
        const response = await axios.get<ApiResponse>(`/api/v1/salt_cache?${params.toString()}`, {
          headers: {
            Authorization: `${authToken}`,
          },
        });

        // Decode the base64 data field
        const decodedCaches = response.data.results.map((cache) => ({
          ...cache,
          data: cache.data,
        }));

        setCaches(decodedCaches);
        setTotalPages(response.data.paging.num_pages);
        setTotalCount(response.data.paging.count);
        setError(null); // Reset error state on successful response
      } catch (err) {
        if (axios.isAxiosError(err) && err.response && err.response.status === 404) {
          setCaches([]); // Treat 404 as empty results
          setTotalPages(0);
          setTotalCount(0);
        } else {
          setError(err as Error);
        }
      } finally {
        setIsLoading(false);
      }
    };

    fetchCaches();
  }, [currentPage, rowsPerPage, stableQueryParams]); // Ensure proper dependencies

  return {
    caches,
    isLoading,
    error,
    currentPage,
    rowsPerPage,
    setCurrentPage,
    setRowsPerPage,
    totalPages,
    totalCount,
  };
};

export default useCachePaginated;
