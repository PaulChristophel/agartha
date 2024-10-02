import { useLocation, useNavigate } from 'react-router-dom';
import React, { useMemo, useState, useEffect, useCallback } from 'react';

import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';

import usePostConformity from 'src/hooks/conformity/usePostConformity.ts';
import useConformityStatus from 'src/hooks/conformity/useConformityStatus.ts';
import useConformityPaginated from 'src/hooks/conformity/useConformityPaginated.ts';

import ConformityTable from './ConformityTable.tsx';
import SearchConformity from './ConformitySearch.tsx';

function useQuery() {
  return new URLSearchParams(useLocation().search);
}

const parseSuccess = (success: string | null): boolean | undefined => {
  if (success === 'true') return true;
  if (success === 'false') return false;
  return undefined;
};

const ConformityView: React.FC = () => {
  const query = useQuery();
  const navigate = useNavigate();
  const { postConformity } = usePostConformity();

  const [minionID, setMinionID] = useState(query.get('id') || '');
  const [success, setSuccess] = useState(query.get('success') || '');
  const [since, setSince] = useState(query.get('since') || '');
  const [until, setUntil] = useState(query.get('until') || '');
  const [limit, setLimit] = useState<number>(Number(query.get('limit')) || 50);
  const [page, setPage] = useState<number>(Number(query.get('page')) || 1);
  const [orderBy, setOrderBy] = useState<string>(query.get('order_by') || '');

  const parsedSuccess = useMemo(() => parseSuccess(success), [success]);

  const { isPending, getConformity } = useConformityStatus(); // Destructure isPending and getConformity here

  const queryParams = useMemo(
    () => ({
      id: minionID,
      success: parsedSuccess,
      since,
      until,
      limit,
      page,
      order_by: orderBy,
    }),
    [minionID, parsedSuccess, since, until, limit, page, orderBy]
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

  useEffect(() => {
    const params = new URLSearchParams();
    if (minionID) params.set('id', minionID.concat('*'));
    if (success) params.set('success', success);
    if (since) params.set('since', since);
    if (until) params.set('until', until);
    if (limit) params.set('limit', limit.toString());
    if (page) params.set('page', page.toString());
    if (orderBy) params.set('order_by', orderBy);
    navigate({ search: params.toString() });
  }, [minionID, success, since, until, limit, page, orderBy, navigate]);

  const { fetchAllData } = useConformityPaginated(queryParams, page, limit);

  const exportToCSV = async () => {
    const data = await fetchAllData();
    const csvRows = [
      ['Minion ID', 'Successes', 'Failures', 'Changed', 'Unchanged', 'Conforming', 'Alter Time'], // Header row
      ...data.map((row) => [
        row.id,
        row.true_count,
        row.false_count,
        row.changed_count,
        row.unchanged_count,
        row.success,
        row.alter_time,
      ]),
    ];

    const csvContent = `data:text/csv;charset=utf-8,${csvRows.map((e) => e.join(',')).join('\n')}`;
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement('a');
    link.setAttribute('href', encodedUri);
    link.setAttribute('download', 'conformity_data.csv');
    document.body.appendChild(link);
    link.click();
  };

  useEffect(() => {
    getConformity(); // Call getConformity when the component mounts
  }, [getConformity]);

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Conformity
      </Typography>
      <Button variant="contained" onClick={exportToCSV} sx={{ mb: 2 }}>
        Export to CSV
      </Button>
      <Button
        variant="outlined"
        onClick={() => {
          postConformity();
          getConformity(); // Call getConformity when the button is clicked
        }}
        sx={{ mb: 2, ml: 1 }}
        disabled={isPending}
      >
        {isPending ? 'Refreshing...' : 'Refresh Table'}
      </Button>
      <SearchConformity
        minionID={minionID}
        setMinionID={setMinionID}
        success={success}
        setSuccess={setSuccess}
        since={since}
        setSince={setSince}
        until={until}
        setUntil={setUntil}
      />
      <ConformityTable
        queryParams={queryParams}
        setLimit={handleSetLimit}
        setPage={handleSetPage}
        setOrderBy={handleSetOrderBy}
      />
    </Box>
  );
};

export default ConformityView;
