import { useLocation, useNavigate } from 'react-router-dom';
import React, { useMemo, useState, useEffect, useCallback } from 'react';

import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';

import ReturnsTable from './ReturnsTable.tsx';
import ReturnsSearch from './ReturnsSearch.tsx';

function useQuery() {
  return new URLSearchParams(useLocation().search);
}

const parseSuccess = (success: string | null): boolean | undefined => {
  if (success === 'true') return true;
  if (success === 'false') return false;
  return undefined;
};

const ReturnsView: React.FC = () => {
  const query = useQuery();
  const navigate = useNavigate();

  const [jid, setJid] = useState(query.get('jid') || '');
  const [minionID, setMinionID] = useState(query.get('id') || '');
  const [fun, setFun] = useState(query.get('fun') || '');
  const [success, setSuccess] = useState(query.get('success') || '');
  const [since, setSince] = useState(query.get('since') || '');
  const [until, setUntil] = useState(query.get('until') || '');
  const [limit, setLimit] = useState<number>(Number(query.get('limit')) || 25);
  const [page, setPage] = useState<number>(Number(query.get('page')) || 1);
  const [orderBy, setOrderBy] = useState<string>(query.get('order_by') || '');

  const parsedSuccess = useMemo(() => parseSuccess(success), [success]);

  const queryParams = useMemo(
    () => ({
      id: minionID,
      jid,
      fun,
      success: parsedSuccess,
      load_return: false,
      load_full_ret: false,
      since,
      until,
      limit,
      page,
      order_by: orderBy,
    }),
    [minionID, jid, fun, parsedSuccess, since, until, limit, page, orderBy]
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
    if (jid) params.set('jid', jid);
    if (minionID) params.set('id', minionID);
    if (fun) params.set('fun', fun);
    if (success) params.set('success', success);
    if (since) params.set('since', since);
    if (until) params.set('until', until);
    if (limit) params.set('limit', limit.toString());
    if (page) params.set('page', page.toString());
    if (orderBy) params.set('order_by', orderBy);
    navigate({ search: params.toString() }, { replace: true });
  }, [minionID, jid, fun, success, since, until, limit, page, orderBy, navigate]);

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Returns
      </Typography>
      <ReturnsSearch
        minionID={minionID}
        setMinionID={setMinionID}
        jidString={jid}
        setJidString={setJid}
        fun={fun}
        setFun={setFun}
        success={success}
        setSuccess={setSuccess}
        since={since}
        setSince={setSince}
        until={until}
        setUntil={setUntil}
      />
      <ReturnsTable
        queryParams={queryParams}
        setLimit={handleSetLimit}
        setPage={handleSetPage}
        setOrderBy={handleSetOrderBy}
      />
    </Box>
  );
};

export default ReturnsView;
