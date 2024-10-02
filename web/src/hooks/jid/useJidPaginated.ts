import axios from 'axios';
import { useMemo, useState, useEffect } from 'react';

interface Job {
  jid: string;
  alter_time: string;
  load: Record<string, unknown>;
}

interface Paging {
  num_pages: number;
  count: number;
}

interface ApiResponse {
  results: Job[];
  paging: Paging;
}

interface UseJidPaginated {
  jids: Job[];
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
  filter?: string;
  load_load?: boolean;
  since?: string;
  until?: string;
  order_by?: string;
}

const useJidPaginated = (
  queryParams: QueryParams,
  page: number = 1,
  per_page: number = 10
): UseJidPaginated => {
  const [jids, setJobs] = useState<Job[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [currentPage, setCurrentPage] = useState(page);
  const [rowsPerPage, setRowsPerPage] = useState(per_page);
  const [totalPages, setTotalPages] = useState(0);
  const [totalCount, setTotalCount] = useState(0);

  const stableQueryParams = useMemo(() => queryParams, [queryParams]);

  useEffect(() => {
    const fetchJobs = async () => {
      setIsLoading(true);
      try {
        const { filter, load_load, since, until, order_by } = stableQueryParams;
        const params = new URLSearchParams();

        if (filter) params.append('jid', filter.concat('*'));
        if (load_load !== undefined) params.append('load_load', String(load_load));
        if (since) params.append('since', new Date(since).toISOString());
        if (until) params.append('until', new Date(until).toISOString());
        if (order_by) params.append('order_by', order_by);
        params.append('page', String(currentPage));
        params.append('per_page', String(rowsPerPage));

        const authToken = localStorage.getItem('authToken');
        const response = await axios.get<ApiResponse>(`/api/v1/jid?${params.toString()}`, {
          headers: {
            Authorization: `${authToken}`,
          },
        });

        setJobs(response.data.results);
        setTotalPages(response.data.paging.num_pages);
        setTotalCount(response.data.paging.count);
        setError(null); // Reset error state on successful response
      } catch (err) {
        if (axios.isAxiosError(err) && err.response && err.response.status === 404) {
          setJobs([]); // Treat 404 as empty results
          setTotalPages(0);
          setTotalCount(0);
        } else {
          setError(err as Error);
        }
      } finally {
        setIsLoading(false);
      }
    };

    fetchJobs();
  }, [currentPage, rowsPerPage, stableQueryParams]); // Ensure proper dependencies

  return {
    jids,
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

export default useJidPaginated;
