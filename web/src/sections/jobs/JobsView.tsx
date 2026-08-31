import { useLocation, useNavigate } from 'react-router-dom';
import React, { useMemo, useState, useEffect, useCallback } from 'react';

import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';

import JobsTable from './JobsTable.tsx';
import JobsSearch from './JobsSearch.tsx';
import {
  parseDataFilter,
  serializeDataFilter,
  type DataFilterState,
} from '../events/dataFilter.ts';

function useQuery() {
  return new URLSearchParams(useLocation().search);
}

const JobsView: React.FC = () => {
  const query = useQuery();
  const navigate = useNavigate();

  const [filterName, setFilterName] = useState(query.get('filter') || '');
  const [since, setSince] = useState(query.get('since') || '');
  const [until, setUntil] = useState(query.get('until') || '');
  const [limit, setLimit] = useState<number>(Number(query.get('limit')) || 25);
  const [page, setPage] = useState<number>(Number(query.get('page')) || 1);
  const [orderBy, setOrderBy] = useState<string>(query.get('order_by') || '');
  const [loadFilter, setLoadFilter] = useState<DataFilterState>(() =>
    parseDataFilter(query.get('load_query'))
  );

  const queryParams = useMemo(
    () => ({
      filter: filterName,
      load_load: false,
      since,
      until,
      limit,
      page,
      order_by: orderBy,
      load_query: serializeDataFilter(loadFilter),
    }),
    [filterName, since, until, limit, page, orderBy, loadFilter]
  );

  const handleSetLimit = useCallback((newLimit: number) => {
    setLimit(newLimit);
  }, []);

  const handleSetPage = useCallback((newPage: number) => {
    setPage(newPage);
  }, []);

  const handleSetOrderBy = useCallback((newOrderBy: string) => {
    setOrderBy(newOrderBy);
  }, []);

  const handleSetLoadFilter = useCallback((newLoadFilter: DataFilterState) => {
    setLoadFilter(newLoadFilter);
    setPage(1);
  }, []);

  // Update URL when query parameters change
  useEffect(() => {
    const params = new URLSearchParams();
    if (filterName) params.set('filter', filterName);
    if (since) params.set('since', since);
    if (until) params.set('until', until);
    if (limit) params.set('limit', limit.toString());
    if (page) params.set('page', page.toString());
    if (orderBy) params.set('order_by', orderBy);
    const serializedLoadFilter = serializeDataFilter(loadFilter);
    if (serializedLoadFilter) params.set('load_query', serializedLoadFilter);
    navigate({ search: params.toString() }, { replace: true });
  }, [filterName, since, until, limit, page, orderBy, loadFilter, navigate]);

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Jobs
      </Typography>
      <JobsSearch
        filterName={filterName}
        setFilterName={setFilterName}
        since={since}
        setSince={setSince}
        until={until}
        setUntil={setUntil}
        loadFilter={loadFilter}
        setLoadFilter={handleSetLoadFilter}
      />
      <JobsTable
        queryParams={queryParams}
        setLimit={handleSetLimit}
        setPage={handleSetPage}
        setOrderBy={handleSetOrderBy}
      />
    </Box>
  );
};

export default JobsView;
