import React from 'react';

import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';

interface SearchJobsProps {
  masterID: string;
  setMasterID: (value: string) => void;
  tag: string;
  setTag: (value: string) => void;
  since: string;
  setSince: (value: string) => void;
  until: string;
  setUntil: (value: string) => void;
}

const SearchJobs: React.FC<SearchJobsProps> = ({
  masterID,
  setMasterID,
  tag,
  setTag,
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
      label="Tag"
      value={tag}
      onChange={(e) => setTag(e.target.value)}
      sx={{ marginRight: 2, width: '20%' }}
    />
    <TextField
      label="Master ID"
      value={masterID}
      onChange={(e) => setMasterID(e.target.value)}
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

export default SearchJobs;
