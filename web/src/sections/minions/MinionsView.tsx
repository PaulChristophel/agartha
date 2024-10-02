// src/sections/minions/MinionsView.tsx
import { useLocation, useNavigate } from 'react-router-dom';
import React, { useMemo, useState, useEffect, useCallback } from 'react';

import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';

import useMinionsPaginated from 'src/hooks/saltMinion/useMinionPaginated.ts';

import MinionsTable from './MinionsTable.tsx';
import MinionsSearch from './MinionsSearch.tsx';

function useQuery() {
  return new URLSearchParams(useLocation().search);
}

interface Minion {
  alter_time: string;
  grains: Record<string, unknown>;
  pillar: Record<string, unknown>;
  id: string;
  minion_id: string;
}

const MinionsView: React.FC = () => {
  const query = useQuery();
  const navigate = useNavigate();

  const [minionID, setMinionID] = useState(query.get('minion_id') || '');
  const [since, setSince] = useState(query.get('since') || '');
  const [until, setUntil] = useState(query.get('until') || '');
  const [limit, setLimit] = useState<number>(Number(query.get('limit')) || 50);
  const [page, setPage] = useState<number>(Number(query.get('page')) || 1);
  const [orderBy, setOrderBy] = useState<string>(query.get('order_by') || '');
  const [grainKeys, setGrainKeys] = useState<string[]>(query.get('grain_keys')?.split(',') || []);
  // const [grainValue, setGrainValue] = useState(query.get('grain_value') || '');

  const queryParams = useMemo(
    () => ({
      minion_id: minionID,
      jsonpath_grains: grainKeys.join(','),
      // jsonpath_grains_filter: grainValue,
      since,
      until,
      limit,
      page,
      order_by: orderBy,
    }),
    [minionID, since, until, limit, page, orderBy, grainKeys]
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

  const { fetchAllData } = useMinionsPaginated(queryParams, page, limit);

  const exportToCSV = async () => {
    const data = await fetchAllData();

    if (data.length === 0) {
      console.warn('No data to export');
      return;
    }

    // Extract keys from the first element to form the header row, ignoring "pillar"
    const headers = Object.keys(data[0]).filter((key) => key !== 'pillar' && key !== 'id') as Array<
      keyof Minion
    >;

    // Extract keys from the grains object
    const grainsKeys = Object.keys(data[0].grains || {});

    // Combine minion headers and grains keys
    const allHeaders = [...headers.filter((header) => header !== 'grains'), ...grainsKeys];

    // Map through each row to get the values
    const csvRows = [
      allHeaders.join(','), // Header row
      ...data.map((row) => {
        const rowValues = headers
          .filter((header) => header !== 'grains')
          .map((header) => JSON.stringify(row[header] ?? ''));

        const grainsValues = grainsKeys.map((key) => JSON.stringify(row.grains[key] ?? ''));

        return [...rowValues, ...grainsValues].join(',');
      }),
    ];

    const csvContent = `data:text/csv;charset=utf-8,${csvRows.join('\n')}`;
    const encodedUri = encodeURI(csvContent);
    const link = document.createElement('a');
    link.setAttribute('href', encodedUri);
    link.setAttribute('download', 'minion_data.csv');
    document.body.appendChild(link);
    link.click();
  };

  useEffect(() => {
    const params = new URLSearchParams();
    if (minionID) params.set('minion_id', minionID);
    if (since) params.set('since', since);
    if (until) params.set('until', until);
    if (limit) params.set('limit', limit.toString());
    if (page) params.set('page', page.toString());
    if (orderBy) params.set('order_by', orderBy);
    if (grainKeys.length) params.set('grain_keys', grainKeys.join(',')); // Updated to grain_keys
    // if (grainValue) params.set('grain_value', grainValue);
    navigate({ search: params.toString() });
  }, [minionID, since, until, limit, page, orderBy, grainKeys, navigate]);

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Minions
      </Typography>
      <Button variant="contained" onClick={exportToCSV} sx={{ mb: 2 }}>
        Export to CSV
      </Button>
      <MinionsSearch
        minionID={minionID}
        setMinionID={setMinionID}
        since={since}
        setSince={setSince}
        until={until}
        setUntil={setUntil}
        grainKeys={grainKeys} // Updated to grainKeys
        setGrainKeys={setGrainKeys}
        // grainValue={grainValue}
        // setGrainValue={setGrainValue}
        setOrderBy={handleSetOrderBy}
      />
      <MinionsTable
        queryParams={queryParams}
        setLimit={handleSetLimit}
        setPage={handleSetPage}
        setOrderBy={handleSetOrderBy}
        // grainKeys={grainKeys} // Updated to grainKeys
      />
    </Box>
  );
};

export default MinionsView;
