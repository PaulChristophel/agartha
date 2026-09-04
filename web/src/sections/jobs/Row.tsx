import { useState } from 'react';
import { Link } from 'react-router-dom';

import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import { Collapse, TableRow, TableCell, IconButton } from '@mui/material';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';

import formatTime from 'src/utils/formatTime.ts';

import JobDetailsPanel from './JobDetailsPanel.tsx';

interface RowProps {
  row: { jid: string; alter_time: string; load: Record<string, unknown> };
}

export default function Row({ row }: RowProps) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <TableRow sx={{ '& > *': { borderBottom: 'unset' } }}>
        <TableCell>
          <IconButton
            aria-label={`Job details for ${row.jid}`}
            aria-expanded={open}
            size="small"
            onClick={() => setOpen(!open)}
          >
            {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
        </TableCell>
        <TableCell component="th" scope="row">
          <Link to={`/job/${encodeURIComponent(row.jid)}`}>{row.jid}</Link>
        </TableCell>
        <TableCell>{formatTime(row.alter_time)}</TableCell>
      </TableRow>
      <TableRow>
        <TableCell sx={{ py: 0 }} colSpan={3}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            {open && <JobDetailsPanel jid={row.jid} />}
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  );
}
