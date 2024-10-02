import { useLocation, useNavigate } from 'react-router-dom';
import React, { useMemo, useState, useEffect, useCallback } from 'react';

import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';

import EventsTable from './EventsTable.tsx';
import EventsSearch from './EventsSearch.tsx';

function useQuery() {
  return new URLSearchParams(useLocation().search);
}

const EventsView: React.FC = () => {
  const query = useQuery();
  const navigate = useNavigate();

  const [masterID, setMasterID] = useState(query.get('master_id') || '');
  const [tag, setTag] = useState(query.get('tag') || '');
  const [since, setSince] = useState(query.get('since') || '');
  const [until, setUntil] = useState(query.get('until') || '');
  const [limit, setLimit] = useState<number>(Number(query.get('limit')) || 25);
  const [page, setPage] = useState<number>(Number(query.get('page')) || 1);
  const [orderBy, setOrderBy] = useState<string>(query.get('order_by') || '');

  const queryParams = useMemo(
    () => ({
      tag,
      master_id: masterID,
      since,
      until,
      limit,
      page,
      order_by: orderBy,
    }),
    [tag, masterID, since, until, limit, page, orderBy]
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

  // Update URL when query parameters change
  useEffect(() => {
    const params = new URLSearchParams();
    if (masterID) params.set('master_id', masterID);
    if (tag) params.set('tag', tag);
    if (since) params.set('since', since);
    if (until) params.set('until', until);
    if (limit) params.set('limit', limit.toString());
    if (page) params.set('page', page.toString());
    if (orderBy) params.set('order_by', orderBy);
    navigate({ search: params.toString() }, { replace: true });
  }, [masterID, tag, since, until, limit, page, orderBy, navigate]);

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Events
      </Typography>
      <EventsSearch
        masterID={masterID}
        setMasterID={setMasterID}
        tag={tag}
        setTag={setTag}
        since={since}
        setSince={setSince}
        until={until}
        setUntil={setUntil}
      />
      <EventsTable
        queryParams={queryParams}
        setLimit={handleSetLimit}
        setPage={handleSetPage}
        setOrderBy={handleSetOrderBy}
      />
    </Box>
  );
};

export default EventsView;
