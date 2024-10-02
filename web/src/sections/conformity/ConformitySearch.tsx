import React from 'react';

import Box from '@mui/material/Box';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';

interface ConformitySearchProps {
  minionID: string;
  setMinionID: (value: string) => void;
  success: string;
  setSuccess: (value: string) => void;
  since: string;
  setSince: (value: string) => void;
  until: string;
  setUntil: (value: string) => void;
}

const ConformitySearch: React.FC<ConformitySearchProps> = ({
  minionID,
  setMinionID,
  success,
  setSuccess,
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
      label="Minion ID"
      value={minionID}
      onChange={(e) => setMinionID(e.target.value)}
      sx={{ marginRight: 2, width: '20%' }}
    />
    <TextField
      select
      label="Success"
      value={success}
      onChange={(e) => setSuccess(e.target.value)}
      sx={{ marginRight: 2, width: '20%' }}
    >
      <MenuItem value="">all</MenuItem>
      <MenuItem value="true">true</MenuItem>
      <MenuItem value="false">false</MenuItem>
    </TextField>
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

export default ConformitySearch;
