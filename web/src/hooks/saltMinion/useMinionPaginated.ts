import axios from 'axios';
import { useMemo, useState, useEffect, useCallback } from 'react';

// Define the interfaces
interface Minion {
  alter_time: string;
  grains: Record<string, unknown>;
  pillar: Record<string, unknown>;
  id: string;
  minion_id: string;
}

interface Paging {
  num_pages: number;
  count: number;
}

interface ApiResponse {
  results: Minion[];
  paging: Paging;
}

interface UseMinionPaginated {
  minions: Minion[];
  isLoading: boolean;
  error: Error | null;
  currentPage: number;
  rowsPerPage: number;
  setCurrentPage: (page: number) => void;
  setRowsPerPage: (rows: number) => void;
  totalPages: number;
  totalCount: number;
  fetchAllData: () => Promise<Minion[]>;
}

interface QueryParams {
  minion_id?: string;
  load_grains?: boolean;
  jsonpath_grains?: string;
  jsonpath_grains_filter?: string;
  load_pillar?: boolean;
  jsonpath_pillar?: string;
  jsonpath_pillar_filter?: string;
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

// Custom hook for paginated minions
const useMinionPaginated = (
  queryParams: QueryParams,
  initialPage: number = 1,
  per_page: number = 10
): UseMinionPaginated => {
  const [minions, setMinions] = useState<Minion[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [currentPage, setCurrentPage] = useState(initialPage);
  const [rowsPerPage, setRowsPerPage] = useState(per_page);
  const [totalPages, setTotalPages] = useState(0);
  const [totalCount, setTotalCount] = useState(0);

  const stableQueryParams = useMemo(() => queryParams, [queryParams]);

  const fetchMinions = useCallback(
    async (page: number, limit: number) => {
      const {
        minion_id,
        load_grains,
        jsonpath_grains,
        jsonpath_grains_filter,
        load_pillar,
        jsonpath_pillar,
        jsonpath_pillar_filter,
        since,
        until,
        order_by,
      } = stableQueryParams;
      const params = new URLSearchParams();

      if (minion_id) params.append('minion_id', minion_id.concat('*'));
      if (load_grains !== undefined) params.append('load_grains', String(load_grains));
      if (jsonpath_grains) params.append('jsonpath_grains', jsonpath_grains);
      if (jsonpath_grains_filter) params.append('jsonpath_grains_filter', jsonpath_grains_filter);
      if (load_pillar !== undefined) params.append('load_pillar', String(load_pillar));
      if (jsonpath_pillar) params.append('jsonpath_pillar', jsonpath_pillar);
      if (jsonpath_pillar_filter) params.append('jsonpath_pillar_filter', jsonpath_pillar_filter);
      if (since) params.append('since', new Date(since).toISOString());
      if (until) params.append('until', new Date(until).toISOString());
      if (order_by) params.append('order_by', order_by);
      params.append('page', String(page));
      params.append('per_page', String(limit));

      const authToken = localStorage.getItem('authToken');
      const response = await axios.get<ApiResponse>(`/api/v1/salt_minion?${params.toString()}`, {
        headers: {
          Authorization: `${authToken}`,
        },
      });

      // Process the response to decode base64 strings and update grains and pillar
      const decodedMinions = response.data.results.map((minion) => ({
        ...minion,
        grains: minion.grains,
        pillar: minion.pillar,
      }));

      return { results: decodedMinions, paging: response.data.paging };
    },
    [stableQueryParams]
  );

  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true);
      try {
        const data = await fetchMinions(currentPage, rowsPerPage);
        setMinions(data.results);
        setTotalPages(data.paging.num_pages);
        setTotalCount(data.paging.count);
        setError(null); // Reset error state on successful response
      } catch (err) {
        if (axios.isAxiosError(err) && err.response && err.response.status === 404) {
          setMinions([]); // Treat 404 as empty results
          setTotalPages(0);
          setTotalCount(0);
        } else {
          setError(err as Error);
        }
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [currentPage, rowsPerPage, stableQueryParams, fetchMinions]);

  const fetchAllData = async (): Promise<Minion[]> => {
    try {
      const firstPageData = await fetchMinions(1, rowsPerPage);
      const totalPagesToFetch = firstPageData.paging.num_pages;

      const allDataRequests = [];
      for (let pageNum = 2; pageNum <= totalPagesToFetch; pageNum += 1) {
        allDataRequests.push(
          fetchMinions(pageNum, rowsPerPage).catch((err) => {
            if (axios.isAxiosError(err) && err.response && err.response.status === 404) {
              return { results: [], paging: { num_pages: 0, count: 0 } };
            }
            throw err;
          })
        );
      }

      const allDataResponses = await Promise.all(allDataRequests);
      const allData = [firstPageData, ...allDataResponses].flatMap((response) => response.results);

      return allData;
    } catch (err) {
      console.error('Failed to fetch all data:', err);
      return [];
    }
  };

  return {
    minions,
    isLoading,
    error,
    currentPage,
    rowsPerPage,
    setCurrentPage,
    setRowsPerPage,
    totalPages,
    totalCount,
    fetchAllData,
  };
};

export default useMinionPaginated;
