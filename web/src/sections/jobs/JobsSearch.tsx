import React from 'react';

import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';

import type { DataFilterState } from '../events/dataFilter.ts';
import DataFilterBuilder from '../events/DataFilterBuilder.tsx';

interface JobsSearchProps {
  filterName: string;
  setFilterName: (value: string) => void;
  since: string;
  setSince: (value: string) => void;
  until: string;
  setUntil: (value: string) => void;
  loadFilter: DataFilterState;
  setLoadFilter: (value: DataFilterState) => void;
}

const JobsSearch: React.FC<JobsSearchProps> = ({
  filterName,
  setFilterName,
  since,
  setSince,
  until,
  setUntil,
  loadFilter,
  setLoadFilter,
}) => (
  <Box
    display="flex"
    flexWrap="wrap"
    alignItems="center"
    padding={2}
    bgcolor="background.paper"
    borderRadius={1}
    boxShadow={1}
    mb={2}
    sx={{ columnGap: 2, rowGap: 2 }}
  >
    <TextField
      label="JID Search"
      value={filterName}
      onChange={(e) => {
        const re = /^[0-9\b]+$/;
        if (re.test(e.target.value) || e.target.value === '') {
          setFilterName(e.target.value);
        }
      }}
      sx={{ flex: '1 1 260px' }}
    />
    <TextField
      label="From"
      type="datetime-local"
      value={since}
      onChange={(e) => setSince(e.target.value)}
      sx={{ flex: '1 1 220px' }}
      InputLabelProps={{
        shrink: true,
      }}
    />
    <TextField
      label="To"
      type="datetime-local"
      value={until}
      onChange={(e) => setUntil(e.target.value)}
      sx={{ flex: '1 1 220px' }}
      InputLabelProps={{
        shrink: true,
      }}
    />
    <DataFilterBuilder
      dataFilter={loadFilter}
      setDataFilter={setLoadFilter}
      fieldLabel="Job load"
    />
  </Box>
);

export default JobsSearch;
