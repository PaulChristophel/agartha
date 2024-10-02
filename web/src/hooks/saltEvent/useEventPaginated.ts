import axios from 'axios';
import { useMemo, useState, useEffect } from 'react';

interface Event {
  alter_time: string;
  data: Record<string, unknown>;
  id: number;
  master_id: string;
  tag: string;
}

interface Paging {
  num_pages: number;
  count: number;
}

interface ApiResponse {
  results: Event[];
  paging: Paging;
}

interface UseJidPaginated {
  events: Event[];
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
  tag?: string;
  masterID?: string;
  load_data?: boolean;
  since?: string;
  until?: string;
  order_by?: string;
}

const useJidPaginated = (
  queryParams: QueryParams,
  page: number = 1,
  per_page: number = 10
): UseJidPaginated => {
  const [events, setEvents] = useState<Event[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const [currentPage, setCurrentPage] = useState(page);
  const [rowsPerPage, setRowsPerPage] = useState(per_page);
  const [totalPages, setTotalPages] = useState(0);
  const [totalCount, setTotalCount] = useState(0);

  const stableQueryParams = useMemo(() => queryParams, [queryParams]);

  useEffect(() => {
    const fetchEvents = async () => {
      setIsLoading(true);
      try {
        const { tag, masterID, load_data, since, until, order_by } = stableQueryParams;
        const params = new URLSearchParams();

        if (tag) params.append('tag', tag.concat('*'));
        if (masterID) params.append('master_id', masterID);
        if (load_data !== undefined) params.append('load_data', String(load_data));
        if (since) params.append('since', new Date(since).toISOString());
        if (until) params.append('until', new Date(until).toISOString());
        if (order_by) params.append('order_by', order_by);
        params.append('page', String(currentPage));
        params.append('per_page', String(rowsPerPage));

        const authToken = localStorage.getItem('authToken');
        const response = await axios.get<ApiResponse>(`/api/v1/salt_event?${params.toString()}`, {
          headers: {
            Authorization: `${authToken}`,
          },
        });

        setEvents(response.data.results);
        setTotalPages(response.data.paging.num_pages);
        setTotalCount(response.data.paging.count);
        setError(null); // Reset error state on successful response
      } catch (err) {
        if (axios.isAxiosError(err) && err.response && err.response.status === 404) {
          setEvents([]); // Treat 404 as empty results
          setTotalPages(0);
          setTotalCount(0);
        } else {
          setError(err as Error);
        }
      } finally {
        setIsLoading(false);
      }
    };

    fetchEvents();
  }, [currentPage, rowsPerPage, stableQueryParams]); // Ensure proper dependencies

  return {
    events,
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
