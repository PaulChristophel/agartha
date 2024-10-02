import React from 'react';
import { Link } from 'react-router-dom';

// import Tooltip from '@mui/material/Tooltip';
import TableRow from '@mui/material/TableRow';
import Checkbox from '@mui/material/Checkbox';
import TableCell from '@mui/material/TableCell';
// import IconButton from '@mui/material/IconButton';
// import ContentCopyIcon from '@mui/icons-material/ContentCopy';

import formatTime from 'src/utils/formatTime.ts';

interface RowProps {
  row: {
    id: string;
    minion_id?: string;
    grains?: Record<string, unknown>;
    pillar?: Record<string, unknown>;
    alter_time: string;
  };
  selected: boolean;
  onSelect: (id: string) => void;
  onDeselect: (id: string) => void;
}

const Row: React.FC<RowProps> = ({ row, selected, onSelect, onDeselect }) => {
  const handleSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      onSelect(row.id);
    } else {
      onDeselect(row.id);
    }
  };

  // Function to copy row data to clipboard
  // const handleCopy = () => {
  //   const { pillar, grains, ...rest } = row;
  //   const rowData = {
  //     ...rest,
  //     grains: Object.keys(grains || {}).length > 0 ? grains : undefined, // Include grains only if not empty
  //     pillar: Object.keys(pillar || {}).length > 0 ? pillar : undefined, // Include pillar only if not empty
  //   };
  //   navigator.clipboard.writeText(JSON.stringify(rowData, null, 2));
  // };

  // Extract grains keys and values
  const grainEntries = row.grains ? Object.entries(row.grains) : [];

  return (
    <TableRow sx={{ '& > *': { borderBottom: 'unset' } }}>
      <TableCell padding="checkbox">
        <Checkbox color="primary" checked={selected} onChange={handleSelect} />
      </TableCell>{' '}
      <TableCell> </TableCell>
      <TableCell>
        <Link to={`/minion/${row.minion_id}/`}>{row.minion_id}</Link>
      </TableCell>
      {grainEntries.map(([key, value]) => (
        <TableCell key={key}>
          {typeof value === 'object' ? JSON.stringify(value) : String(value)}
        </TableCell>
      ))}
      <TableCell>{formatTime(row.alter_time)}</TableCell>
      {/* <TableCell sx={{ width: '50px' }}>
        <Tooltip title="Copy Row">
          <IconButton onClick={handleCopy}>
            <ContentCopyIcon />
          </IconButton>
        </Tooltip>
      </TableCell> */}
    </TableRow>
  );
};

export default Row;
