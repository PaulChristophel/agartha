import axios from 'axios';
import { useMemo, useState, useEffect } from 'react';

interface HighState {
  alter_time: string;
  full_ret: Record<string, unknown>;
  fun: string;
  id: string;
  return: Record<string, unknown>;
  success: boolean;
}

interface Paging {
  num_pages: number;
  count: number;
}

interface ApiResponse {
  results: HighState[];
  paging: Paging;
}

interface UseHighStatePaginated {
  highstates: HighState[];
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
  id?: string;
  fun?: string;
  success?: boolean;
  load_return?: boolean;
  load_full_ret?: boolean;
  since?: string;
  until?: string;
  order_by?: string;
}

const useHighStatePaginated = (
  queryParams: QueryParams,
  page: number = 1,
  per_page: number = 10
): UseHighStatePaginated => {
  const [highstates, setHighStates] = useState<HighState[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [currentPage, setCurrentPage] = useState(page);
  const [rowsPerPage, setRowsPerPage] = useState(per_page);
  const [totalPages, setTotalPages] = useState(0);
  const [totalCount, setTotalCount] = useState(0);

  const stableQueryParams = useMemo(() => queryParams, [queryParams]);

  useEffect(() => {
    const fetchHighStates = async () => {
      setIsLoading(true);
      try {
        const { id, fun, success, load_return, load_full_ret, since, until, order_by } =
          stableQueryParams;
        const params = new URLSearchParams();

        if (id) params.append('id', id.concat('*'));
        if (fun) params.append('fun', fun.concat('*'));
        if (success !== undefined) params.append('success', String(success));
        if (load_return !== undefined) params.append('load_return', String(load_return));
        if (load_full_ret !== undefined) params.append('load_full_ret', String(load_full_ret));
        if (since) params.append('since', new Date(since).toISOString());
        if (until) params.append('until', new Date(until).toISOString());
        if (order_by) params.append('order_by', order_by);
        params.append('page', String(currentPage));
        params.append('per_page', String(rowsPerPage));

        const authToken = localStorage.getItem('authToken');
        const response = await axios.get<ApiResponse>(`/api/v1/high_state?${params.toString()}`, {
          headers: {
            Authorization: `${authToken}`,
          },
        });

        setHighStates(response.data.results);
        setTotalPages(response.data.paging.num_pages);
        setTotalCount(response.data.paging.count);
        setError(null); // Reset error state on successful response
      } catch (err) {
        if (axios.isAxiosError(err) && err.response && err.response.status === 404) {
          setHighStates([]); // Treat 404 as empty results
          setTotalPages(0);
          setTotalCount(0);
        } else {
          setError(err as Error);
        }
      } finally {
        setIsLoading(false);
      }
    };

    fetchHighStates();
  }, [currentPage, rowsPerPage, stableQueryParams]); // Ensure proper dependencies

  return {
    highstates,
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

export default useHighStatePaginated;
