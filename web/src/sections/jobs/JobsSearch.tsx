import React from 'react';

import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';

interface JobsSearchProps {
  filterName: string;
  setFilterName: (value: string) => void;
  since: string;
  setSince: (value: string) => void;
  until: string;
  setUntil: (value: string) => void;
}

const JobsSearch: React.FC<JobsSearchProps> = ({
  filterName,
  setFilterName,
  since,
  setSince,
  until,
  setUntil,
}) => (
  <Box
    display="flex"
    alignItems="center"
    padding={2}
    bgcolor="background.paper"
    borderRadius={1}
    boxShadow={1}
    mb={2}
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
      sx={{ marginRight: 2, width: '20%' }}
    />
    <TextField
      label="From"
      type="datetime-local"
      value={since}
      onChange={(e) => setSince(e.target.value)}
      sx={{ marginRight: 2, width: '20%' }}
      InputLabelProps={{
        shrink: true,
      }}
    />
    <TextField
      label="To"
      type="datetime-local"
      value={until}
      onChange={(e) => setUntil(e.target.value)}
      sx={{ marginRight: 2, width: '20%' }}
      InputLabelProps={{
        shrink: true,
      }}
    />
  </Box>
);

export default JobsSearch;
