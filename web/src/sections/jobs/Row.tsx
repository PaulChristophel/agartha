import jsYaml from 'js-yaml';
import { Link } from 'react-router-dom';
import { keymap } from '@codemirror/view';
import { yaml } from '@codemirror/lang-yaml';
import CodeMirror from '@uiw/react-codemirror';
import { foldKeymap } from '@codemirror/language';
import React, { useState, useEffect } from 'react';
import { autocompletion } from '@codemirror/autocomplete';
import { search, searchKeymap } from '@codemirror/search';

import Box from '@mui/material/Box';
import Collapse from '@mui/material/Collapse';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import IconButton from '@mui/material/IconButton';
import CircularProgress from '@mui/material/CircularProgress';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';

import useJid from 'src/hooks/jid/useJid.ts';

import formatTime from 'src/utils/formatTime.ts';

interface RowProps {
  row: {
    jid: string;
    alter_time: string;
    load: Record<string, unknown>;
  };
}

const Row: React.FC<RowProps> = ({ row }) => {
  const [open, setOpen] = useState(false);
  const { load, isLoading, error } = useJid(open ? row.jid : '');
  const [formattedLoad, setFormattedLoad] = useState<string>('');

  useEffect(() => {
    if (load) {
      setFormattedLoad(jsYaml.dump(load));
    }
  }, [load]);

  return (
    <>
      <TableRow sx={{ '& > *': { borderBottom: 'unset' } }}>
        <TableCell>
          <IconButton aria-label="expand row" size="small" onClick={() => setOpen(!open)}>
            {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
        </TableCell>
        <TableCell component="th" scope="row">
          <Link to={`/job/${row.jid}`}>{row.jid}</Link>
        </TableCell>
        <TableCell>{formatTime(row.alter_time)}</TableCell>
      </TableRow>
      <TableRow>
        <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={3}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            <Box sx={{ margin: 1 }}>
              {isLoading && <CircularProgress color="success" />}
              {error && <div>Error: {error.message}</div>}
              {!isLoading && !error && load && (
                <CodeMirror
                  maxHeight="350px"
                  value={formattedLoad}
                  extensions={[
                    yaml(),
                    keymap.of([...foldKeymap, ...searchKeymap]),
                    autocompletion(),
                    search({
                      top: true, // position search bar at the top
                    }),
                  ]}
                  theme="dark"
                  readOnly
                />
              )}
            </Box>
          </Collapse>
        </TableCell>
      </TableRow>
    </>
  );
};

export default Row;
