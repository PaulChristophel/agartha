// src/sections/minions/MinionsView.tsx
import { useLocation, useNavigate } from 'react-router-dom';
import React, { useMemo, useState, useEffect, useCallback } from 'react';

import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';

import useMinionsPaginated from 'src/hooks/saltMinion/useMinionPaginated.ts';

import { toColonNotation, toJSONPathNotation } from 'src/utils/grainKeys.ts';

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
  const grainKeysParam = query.get('grain_keys');
  const [grainKeys, setGrainKeys] = useState<string[]>(
    grainKeysParam ? grainKeysParam.split(',').map((key) => toColonNotation(key)) : []
  );
  const grainFiltersParam = query.get('grain_filters');
  const [grainFilters, setGrainFilters] = useState<string[]>(
    grainFiltersParam && grainFiltersParam.length > 0
      ? grainFiltersParam.split(',').filter((filter) => filter.length > 0)
      : []
  );
  const pillarKeysParam = query.get('pillar_keys');
  const [pillarKeys, setPillarKeys] = useState<string[]>(
    pillarKeysParam ? pillarKeysParam.split(',').map((key) => toColonNotation(key)) : []
  );
  const pillarFiltersParam = query.get('pillar_filters');
  const [pillarFilters, setPillarFilters] = useState<string[]>(
    pillarFiltersParam && pillarFiltersParam.length > 0
      ? pillarFiltersParam.split(',').filter((filter) => filter.length > 0)
      : []
  );

  const queryParams = useMemo(() => {
    const normalizedGrainKeys = grainKeys
      .map((key) => toJSONPathNotation(key))
      .filter((key) => key.length > 0);

    const normalizedGrainFilters = grainFilters
      .map((filter) => {
        const parts = filter.split('::');
        if (parts.length < 2) {
          return '';
        }
        const [pathAndValue, type, ...operatorParts] = parts;
        if (!pathAndValue || !type) {
          return '';
        }
        const operator = operatorParts.join('::');
        const lastColon = pathAndValue.lastIndexOf(':');
        if (lastColon === -1) {
          return '';
        }
        const path = pathAndValue.slice(0, lastColon);
        const value = pathAndValue.slice(lastColon + 1);
        const normalizedPath = toJSONPathNotation(path);
        if (!normalizedPath) {
          return '';
        }
        let normalizedFilter = `${normalizedPath}:${value}::${type}`;
        if (operator) {
          normalizedFilter = `${normalizedFilter}::${operator}`;
        }
        return normalizedFilter;
      })
      .filter((filter) => filter.length > 0);

    const normalizedPillarKeys = pillarKeys
      .map((key) => toJSONPathNotation(key))
      .filter((key) => key.length > 0);

    const normalizedPillarFilters = pillarFilters
      .map((filter) => {
        const parts = filter.split('::');
        if (parts.length < 2) {
          return '';
        }
        const [pathAndValue, type, ...operatorParts] = parts;
        if (!pathAndValue || !type) {
          return '';
        }
        const operator = operatorParts.join('::');
        const lastColon = pathAndValue.lastIndexOf(':');
        if (lastColon === -1) {
          return '';
        }
        const path = pathAndValue.slice(0, lastColon);
        const value = pathAndValue.slice(lastColon + 1);
        const normalizedPath = toJSONPathNotation(path);
        if (!normalizedPath) {
          return '';
        }
        let normalizedFilter = `${normalizedPath}:${value}::${type}`;
        if (operator) {
          normalizedFilter = `${normalizedFilter}::${operator}`;
        }
        return normalizedFilter;
      })
      .filter((filter) => filter.length > 0);

    return {
      minion_id: minionID,
      jsonpath_grains: normalizedGrainKeys.join(','),
      jsonpath_grains_filter: normalizedGrainFilters.join(','),
      jsonpath_pillar: normalizedPillarKeys.join(','),
      jsonpath_pillar_filter: normalizedPillarFilters.join(','),
      since,
      until,
      limit,
      page,
      order_by: orderBy,
    };
  }, [
    minionID,
    since,
    until,
    limit,
    page,
    orderBy,
    grainKeys,
    grainFilters,
    pillarKeys,
    pillarFilters,
  ]);

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

    const headers = Object.keys(data[0]).filter(
      (key) => key !== 'grains' && key !== 'pillar' && key !== 'id'
    ) as Array<keyof Minion>;

    // Extract keys from the grains object
    const grainsKeys = Object.keys(data[0].grains || {});
    const pillarKeysForExport = Object.keys(data[0].pillar || {});

    const allHeaders = [
      ...headers,
      ...grainsKeys.map((key) => `grains.${key}`),
      ...pillarKeysForExport.map((key) => `pillar.${key}`),
    ];

    // Map through each row to get the values
    const csvRows = [
      allHeaders.join(','), // Header row
      ...data.map((row) => {
        const rowValues = headers.map((header) => JSON.stringify(row[header] ?? ''));

        const grainsValues = grainsKeys.map((key) => JSON.stringify(row.grains[key] ?? ''));
        const pillarValues = pillarKeysForExport.map((key) =>
          JSON.stringify(row.pillar[key] ?? '')
        );

        return [...rowValues, ...grainsValues, ...pillarValues].join(',');
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
    if (grainFilters.length) params.set('grain_filters', grainFilters.join(','));
    if (pillarKeys.length) params.set('pillar_keys', pillarKeys.join(','));
    if (pillarFilters.length) params.set('pillar_filters', pillarFilters.join(','));
    navigate({ search: params.toString() });
  }, [
    minionID,
    since,
    until,
    limit,
    page,
    orderBy,
    grainKeys,
    grainFilters,
    pillarKeys,
    pillarFilters,
    navigate,
  ]);

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
        grainFilters={grainFilters}
        setGrainFilters={setGrainFilters}
        pillarKeys={pillarKeys}
        setPillarKeys={setPillarKeys}
        pillarFilters={pillarFilters}
        setPillarFilters={setPillarFilters}
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
