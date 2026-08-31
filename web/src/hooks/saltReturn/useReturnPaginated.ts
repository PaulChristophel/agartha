import { useMemo, useState, useEffect } from 'react';

import { apiClient as axios } from 'src/api/client.ts';

interface Return {
  alter_time: string;
  full_ret: Record<string, unknown>;
  fun: string;
  id: string;
  jid: string;
  return: Record<string, unknown>;
  success: boolean;
}

interface Paging {
  num_pages: number;
  count: number;
}

interface ApiResponse {
  results: Return[];
  paging: Paging;
}

interface UseReturnPaginated {
  returns: Return[];
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
  jid?: string;
  fun?: string;
  success?: boolean;
  load_return?: boolean;
  load_full_ret?: boolean;
  since?: string;
  until?: string;
  order_by?: string;
  return_match?: string;
  return_filter?: string;
  return_key?: string;
  return_field?: string;
  return_value?: string;
  return_value_type?: string;
  return_query?: string;
}

const useReturnPaginated = (
  queryParams: QueryParams,
  page: number = 1,
  per_page: number = 10
): UseReturnPaginated => {
  const [returns, setReturns] = useState<Return[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [currentPage, setCurrentPage] = useState(page);
  const [rowsPerPage, setRowsPerPage] = useState(per_page);
  const [totalPages, setTotalPages] = useState(0);
  const [totalCount, setTotalCount] = useState(0);

  const stableQueryParams = useMemo(() => queryParams, [queryParams]);

  useEffect(() => {
    const controller = new AbortController();
    const fetchReturns = async () => {
      setIsLoading(true);
      try {
        const {
          id,
          jid,
          fun,
          success,
          load_return,
          load_full_ret,
          since,
          until,
          order_by,
          return_match,
          return_filter,
          return_key,
          return_field,
          return_value,
          return_value_type,
          return_query,
        } = stableQueryParams;
        const params = new URLSearchParams();

        if (id) params.append('id', id.concat('*'));
        if (jid) params.append('jid', jid.concat('*'));
        if (fun) params.append('fun', fun.concat('*'));
        if (success !== undefined) params.append('success', String(success));
        if (load_return !== undefined) params.append('load_return', String(load_return));
        if (load_full_ret !== undefined) params.append('load_full_ret', String(load_full_ret));
        if (since) params.append('since', new Date(since).toISOString());
        if (until) params.append('until', new Date(until).toISOString());
        if (order_by) params.append('order_by', order_by);
        if (return_match) params.append('return_match', return_match);
        if (return_filter !== undefined) params.append('return_filter', return_filter);
        if (return_key !== undefined) params.append('return_key', return_key);
        if (return_field !== undefined) params.append('return_field', return_field);
        if (return_value !== undefined) params.append('return_value', return_value);
        if (return_value_type !== undefined) params.append('return_value_type', return_value_type);
        if (return_query !== undefined) params.append('return_query', return_query);
        params.append('page', String(currentPage));
        params.append('per_page', String(rowsPerPage));

        const response = await axios.get<ApiResponse>(`/api/v1/salt_return?${params.toString()}`, {
          signal: controller.signal,
        });

        setReturns(response.data.results);
        setTotalPages(response.data.paging.num_pages);
        setTotalCount(response.data.paging.count);
        setError(null); // Reset error state on successful response
      } catch (err) {
        if (controller.signal.aborted) return;
        if (axios.isAxiosError(err) && err.response && err.response.status === 404) {
          setReturns([]); // Treat 404 as empty results
          setTotalPages(0);
          setTotalCount(0);
        } else {
          setError(err as Error);
        }
      } finally {
        if (!controller.signal.aborted) setIsLoading(false);
      }
    };

    void fetchReturns();
    return () => controller.abort();
  }, [currentPage, rowsPerPage, stableQueryParams]); // Ensure proper dependencies

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
  };
};

export default useReturnPaginated;
