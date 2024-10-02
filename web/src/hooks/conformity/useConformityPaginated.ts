import axios from 'axios';
import { useMemo, useState, useEffect, useCallback } from 'react';

interface Conformity {
  alter_time: string;
  true_count: number;
  false_count: number;
  changed_count: number;
  unchanged_count: number;
  success: boolean;
  id: string;
}

interface Paging {
  num_pages: number;
  count: number;
}

interface ApiResponse {
  results: Conformity[];
  paging: Paging;
}

interface UseConformityPaginated {
  returns: Conformity[];
  isLoading: boolean;
  error: Error | null;
  currentPage: number;
  rowsPerPage: number;
  setCurrentPage: (page: number) => void;
  setRowsPerPage: (rows: number) => void;
  totalPages: number;
  totalCount: number;
  fetchAllData: () => Promise<Conformity[]>;
}

interface QueryParams {
  id?: string;
  success?: boolean;
  since?: string;
  until?: string;
  order_by?: string;
}

const useConformityPaginated = (
  queryParams: QueryParams,
  initialPage: number = 1,
  per_page: number = 10
): UseConformityPaginated => {
  const [returns, setConformity] = useState<Conformity[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [currentPage, setCurrentPage] = useState(initialPage);
  const [rowsPerPage, setRowsPerPage] = useState(per_page);
  const [totalPages, setTotalPages] = useState(0);
  const [totalCount, setTotalCount] = useState(0);

  const stableQueryParams = useMemo(() => queryParams, [queryParams]);

  const fetchConformity = useCallback(
    async (page: number, limit: number) => {
      const { id, success, since, until, order_by } = stableQueryParams;
      const params = new URLSearchParams();

      if (id) params.append('id', id.concat('*'));
      if (success !== undefined) params.append('success', String(success));
      if (since) params.append('since', new Date(since).toISOString());
      if (until) params.append('until', new Date(until).toISOString());
      if (order_by) params.append('order_by', order_by);
      params.append('page', String(page));
      params.append('limit', String(limit));

      const authToken = localStorage.getItem('authToken');
      const response = await axios.get<ApiResponse>(`/api/v1/conformity?${params.toString()}`, {
        headers: {
          Authorization: `${authToken}`,
        },
      });

      return response.data;
    },
    [stableQueryParams]
  );

  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true);
      try {
        const data = await fetchConformity(currentPage, rowsPerPage);
        setConformity(data.results);
        setTotalPages(data.paging.num_pages);
        setTotalCount(data.paging.count);
        setError(null); // Reset error state on successful response
      } catch (err) {
        if (axios.isAxiosError(err) && err.response && err.response.status === 404) {
          setConformity([]); // Treat 404 as empty results
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
  }, [currentPage, rowsPerPage, stableQueryParams, fetchConformity]);

  const fetchAllData = async (): Promise<Conformity[]> => {
    const firstPageData = await fetchConformity(1, rowsPerPage);
    const totalPagesToFetch = firstPageData.paging.num_pages;

    const allDataRequests = [];
    for (let page = 2; page <= totalPagesToFetch; page += 1) {
      allDataRequests.push(fetchConformity(page, rowsPerPage));
    }

    const allDataResponses = await Promise.all(allDataRequests);
    const allData = [firstPageData, ...allDataResponses].flatMap((response) => response.results);

    return allData;
  };

  return {
    returns,
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

export default useConformityPaginated;
