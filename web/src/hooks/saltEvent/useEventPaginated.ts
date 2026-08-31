import { useMemo, useState, useEffect } from 'react';

import { apiClient as axios } from 'src/api/client.ts';

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
  master_id?: string;
  load_data?: boolean;
  since?: string;
  until?: string;
  order_by?: string;
  data_match?: string;
  data_filter?: string;
  data_key?: string;
  data_field?: string;
  data_value?: string;
  data_value_type?: string;
  data_query?: string;
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
    const controller = new AbortController();
    const fetchEvents = async () => {
      setIsLoading(true);
      try {
        const {
          tag,
          master_id,
          load_data,
          since,
          until,
          order_by,
          data_match,
          data_filter,
          data_key,
          data_field,
          data_value,
          data_value_type,
          data_query,
        } = stableQueryParams;
        const params = new URLSearchParams();

        if (tag) params.append('tag', tag.concat('*'));
        if (master_id) params.append('master_id', master_id);
        if (load_data !== undefined) params.append('load_data', String(load_data));
        if (since) params.append('since', new Date(since).toISOString());
        if (until) params.append('until', new Date(until).toISOString());
        if (order_by) params.append('order_by', order_by);
        if (data_match) params.append('data_match', data_match);
        if (data_filter !== undefined) params.append('data_filter', data_filter);
        if (data_key !== undefined) params.append('data_key', data_key);
        if (data_field !== undefined) params.append('data_field', data_field);
        if (data_value !== undefined) params.append('data_value', data_value);
        if (data_value_type !== undefined) params.append('data_value_type', data_value_type);
        if (data_query !== undefined) params.append('data_query', data_query);
        params.append('page', String(currentPage));
        params.append('per_page', String(rowsPerPage));

        const response = await axios.get<ApiResponse>(`/api/v1/salt_event?${params.toString()}`, {
          signal: controller.signal,
        });

        setEvents(response.data.results);
        setTotalPages(response.data.paging.num_pages);
        setTotalCount(response.data.paging.count);
        setError(null); // Reset error state on successful response
      } catch (err) {
        if (controller.signal.aborted) return;
        if (axios.isAxiosError(err) && err.response && err.response.status === 404) {
          setEvents([]); // Treat 404 as empty results
          setTotalPages(0);
          setTotalCount(0);
        } else {
          setError(err as Error);
        }
      } finally {
        if (!controller.signal.aborted) setIsLoading(false);
      }
    };

    void fetchEvents();
    return () => controller.abort();
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
