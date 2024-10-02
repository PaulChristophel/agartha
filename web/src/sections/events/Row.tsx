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

import useSaltEvent from 'src/hooks/saltEvent/useSaltEvent.ts';

import formatTime from 'src/utils/formatTime.ts';

interface RowProps {
  row: {
    id: number;
    tag?: string;
    master_id?: string;
    alter_time: string;
  };
}

const Row: React.FC<RowProps> = ({ row }) => {
  const [open, setOpen] = useState(false);
  const [fetchData, setFetchData] = useState(false);
  const { eventData, isLoading, error } = useSaltEvent(fetchData ? row.id : 0);
  const [formattedLoad, setFormattedLoad] = useState<string>('');

  useEffect(() => {
    if (eventData) {
      setFormattedLoad(jsYaml.dump(eventData));
    }
  }, [eventData]);

  useEffect(() => {
    if (open) {
      setFetchData(true);
    }
  }, [open]);

  const createTagLink = (tag: string) => {
    const jobRegex = /salt\/job\/(\d+)\/(ret|sub)\/([^/]+)/;
    const runRegex = /salt\/run\/(\d+)\/(ret|new)|(\d{20})/;
    let match = tag.match(jobRegex);

    if (match) {
      const jobId = match[1];
      const minionId = match[3];
      return <Link to={`/return/${jobId}/${minionId}/`}>{tag}</Link>;
    }

    match = tag.match(runRegex);
    if (match) {
      const jobId = match[1] || match[3];
      return <Link to={`/job/${jobId}`}>{tag}</Link>;
    }

    return tag;
  };

  const extractMinionID = (tag: string) => {
    const minionRegex =
      /salt\/job\/\d+\/(ret|sub)\/([^/]+)|minion\/refresh\/([^/]+)|salt\/beacon\/([^/]+)/;
    const match = tag.match(minionRegex);

    if (match) {
      const minionId = match[2] || match[3] || match[4];
      return <Link to={`/minion/${minionId}/`}>{minionId}</Link>;
    }

    return '';
  };

  return (
    <>
      <TableRow sx={{ '& > *': { borderBottom: 'unset' } }}>
        <TableCell>
          <IconButton aria-label="expand row" size="small" onClick={() => setOpen(!open)}>
            {open ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
          </IconButton>
        </TableCell>
        <TableCell component="th" scope="row">
          <Link to={`/event/${row.id}`}>{row.id}</Link>
        </TableCell>
        <TableCell>{createTagLink(row.tag || '')}</TableCell>
        <TableCell>{extractMinionID(row.tag || '')}</TableCell>
        <TableCell>{row.master_id}</TableCell>
        <TableCell>{formatTime(row.alter_time)}</TableCell>
      </TableRow>
      <TableRow>
        <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={6}>
          <Collapse in={open} timeout="auto" unmountOnExit>
            <Box sx={{ margin: 1 }}>
              {isLoading && <CircularProgress color="success" />}
              {error && <div>Error: {error.message}</div>}
              {!isLoading && !error && eventData && (
                <CodeMirror
                  value={formattedLoad}
                  maxHeight="350px"
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
