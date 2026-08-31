import React from 'react';

import Box from '@mui/material/Box';
import Tooltip from '@mui/material/Tooltip';
import TextField from '@mui/material/TextField';

import type { DataFilterState } from './dataFilter.ts';
import DataFilterBuilder from './DataFilterBuilder.tsx';

interface EventsSearchProps {
  masterID: string;
  setMasterID: (value: string) => void;
  tag: string;
  setTag: (value: string) => void;
  since: string;
  setSince: (value: string) => void;
  until: string;
  setUntil: (value: string) => void;
  dataFilter: DataFilterState;
  setDataFilter: (value: DataFilterState) => void;
}

const hoverHelpProps = {
  arrow: true,
  placement: 'top' as const,
  enterDelay: 700,
  enterNextDelay: 700,
  leaveDelay: 0,
  disableFocusListener: true,
  disableInteractive: true,
};

const EventsSearch: React.FC<EventsSearchProps> = ({
  masterID,
  setMasterID,
  tag,
  setTag,
  since,
  setSince,
  until,
  setUntil,
  dataFilter,
  setDataFilter,
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
    <Tooltip {...hoverHelpProps} title="Filter by event tag. Partial text is matched as a prefix.">
      <TextField
        label="Tag"
        value={tag}
        onChange={(event) => setTag(event.target.value)}
        sx={{ flex: '1 1 260px' }}
      />
    </Tooltip>
    <Tooltip {...hoverHelpProps} title="Filter by the ID of the master that received the event.">
      <TextField
        label="Master ID"
        value={masterID}
        onChange={(event) => setMasterID(event.target.value)}
        sx={{ flex: '1 1 260px' }}
      />
    </Tooltip>
    <Tooltip
      {...hoverHelpProps}
      title="Only include events recorded at or after this date and time."
    >
      <TextField
        label="From"
        type="datetime-local"
        value={since}
        onChange={(event) => setSince(event.target.value)}
        sx={{ flex: '1 1 220px' }}
        InputLabelProps={{ shrink: true }}
      />
    </Tooltip>
    <Tooltip
      {...hoverHelpProps}
      title="Only include events recorded at or before this date and time."
    >
      <TextField
        label="To"
        type="datetime-local"
        value={until}
        onChange={(event) => setUntil(event.target.value)}
        sx={{ flex: '1 1 220px' }}
        InputLabelProps={{ shrink: true }}
      />
    </Tooltip>
    <DataFilterBuilder dataFilter={dataFilter} setDataFilter={setDataFilter} />
  </Box>
);

export default EventsSearch;
