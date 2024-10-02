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
import Chip from '@mui/material/Chip';
import Collapse from '@mui/material/Collapse';
import TableRow from '@mui/material/TableRow';
import TableCell from '@mui/material/TableCell';
import IconButton from '@mui/material/IconButton';
import CircularProgress from '@mui/material/CircularProgress';
import KeyboardArrowUpIcon from '@mui/icons-material/KeyboardArrowUp';
import KeyboardArrowDownIcon from '@mui/icons-material/KeyboardArrowDown';

import useSaltReturn from 'src/hooks/saltReturn/useSaltReturn.ts';

import formatTime from 'src/utils/formatTime.ts';

interface RowProps {
  row: {
    alter_time: string;
    id?: string;
    jid?: string;
    fun?: string;
    success?: boolean;
    load_return?: boolean;
    load_full_ret?: boolean;
  };
}

const Row: React.FC<RowProps> = ({ row }) => {
  const [open, setOpen] = useState(false);
  const { returnData, isLoading, error } = useSaltReturn(
    open ? row.jid || '' : '',
    open ? row.id || '' : '',
    true,
    false
  );
  const [formattedLoad, setFormattedLoad] = useState<string>('');

  useEffect(() => {
    if (returnData) {
      setFormattedLoad(jsYaml.dump(returnData));
    }
  }, [returnData]);

  return (
    <>
      <TableRow sx={{ '& > *': { borderBottom: 'unset' } }}>
        <TableCell>
          <IconButton aria-label="expand row" size="small" onClick={() => setOpen(!open)}>
            {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
        </TableCell>
        <TableCell component="th" scope="row">
          <Link to={`/return/${row.jid}/${row.id}/`}>{row.jid}</Link>
        </TableCell>
        <TableCell component="th" scope="row">
          <Link to={`/minion/${row.id}/`}>{row.id}</Link>
        </TableCell>
        <TableCell>{row.fun}</TableCell>
        <TableCell>
          <Chip
            variant="outlined"
            label={row.success?.toString() || 'N/A'}
            color={row.success ? 'success' : 'error'}
          />
        </TableCell>
        <TableCell>{formatTime(row.alter_time)}</TableCell>
      </TableRow>
      <TableRow>
        <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={6}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            <Box sx={{ margin: 1 }}>
              {isLoading && <CircularProgress color="success" />}
              {error && <div>Error: {error.message}</div>}
              {!isLoading && !error && returnData && (
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
