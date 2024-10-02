import axios from 'axios';
import { useNavigate } from 'react-router-dom';
import React, { useState, useEffect, useCallback } from 'react';

import Box from '@mui/material/Box';
import Table from '@mui/material/Table';
import Paper from '@mui/material/Paper';
import Tooltip from '@mui/material/Tooltip';
import TableRow from '@mui/material/TableRow';
import Checkbox from '@mui/material/Checkbox';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import IconButton from '@mui/material/IconButton';
import DeleteIcon from '@mui/icons-material/Delete';
import PlayArrow from '@mui/icons-material/PlayArrow';
import RefreshIcon from '@mui/icons-material/Refresh';
import TableContainer from '@mui/material/TableContainer';
import TableSortLabel from '@mui/material/TableSortLabel';
import TablePagination from '@mui/material/TablePagination';
import CircularProgress from '@mui/material/CircularProgress';

import useDeleteKeyDict from 'src/hooks/netapi/wheel/useDeleteKeyDict.ts';
import useMinionsPaginated from 'src/hooks/saltMinion/useMinionPaginated.ts';

import Row from './Row.tsx';

interface MinionsTableProps {
  queryParams: {
    minion_id?: string;
    jsonpath_grains?: string;
    jsonpath_grains_filter?: string;
    since?: string;
    until?: string;
    limit?: number;
    page?: number;
    order_by?: string;
  };
  setLimit: (newLimit: number) => void;
  setPage: (newPage: number) => void;
  setOrderBy: (orderBy: string) => void;
  // grainKeys: string[];
}

const MinionsTable: React.FC<MinionsTableProps> = ({
  queryParams,
  setLimit,
  setPage,
  setOrderBy,
  // grainKeys,
}) => {
  const {
    minions,
    isLoading,
    error,
    currentPage,
    rowsPerPage,
    setCurrentPage,
    setRowsPerPage,
    totalCount,
  } = useMinionsPaginated(queryParams, queryParams.page || 1, queryParams.limit || 50);

  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [selectAll, setSelectAll] = useState<boolean>(false);
  const [deleting, setDeleting] = useState<boolean>(false);
  const [refreshing, setRefreshing] = useState<boolean>(false);
  const { deleteKeys } = useDeleteKeyDict();

  const navigate = useNavigate();

  useEffect(() => {
    if (queryParams.limit && queryParams.limit !== rowsPerPage) {
      setRowsPerPage(queryParams.limit);
    }
    if (queryParams.page && queryParams.page !== currentPage) {
      setCurrentPage(queryParams.page);
    }
  }, [
    queryParams.limit,
    queryParams.page,
    rowsPerPage,
    currentPage,
    setRowsPerPage,
    setCurrentPage,
  ]);

  useEffect(() => {
    if (selectAll) {
      setSelectedIds(minions.map((row) => row.id));
    } else {
      setSelectedIds([]);
    }
  }, [selectAll, minions]);

  const handleChangePage = useCallback(
    (_event: React.MouseEvent<HTMLButtonElement, MouseEvent> | null, newPage: number) => {
      if (newPage + 1 !== currentPage) {
        setCurrentPage(newPage + 1);
        setPage(newPage + 1);
      }
    },
    [currentPage, setCurrentPage, setPage]
  );

  const handleChangeRowsPerPage = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const newLimit = parseInt(event.target.value, 10);
      if (newLimit !== rowsPerPage) {
        setRowsPerPage(newLimit);
        setLimit(newLimit);
        setCurrentPage(1);
        setPage(1);
      }
    },
    [rowsPerPage, setRowsPerPage, setLimit, setCurrentPage, setPage]
  );

  const handleRequestSort = useCallback(
    (property: string) => {
      // console.log(property);
      const isAsc = queryParams.order_by === `${property} asc`;
      setOrderBy(`${property} ${isAsc ? 'desc' : 'asc'}`);
    },
    [queryParams.order_by, setOrderBy]
  );

  const handleSelect = (id: string) => {
    setSelectedIds((prev) => [...prev, id]);
  };

  const handleDeselect = (id: string) => {
    setSelectedIds((prev) => prev.filter((selectedId) => selectedId !== id));
  };

  const handleSelectAll = () => {
    setSelectAll((prev) => !prev);
  };

  const handleRun = () => {
    // Implement run logic here
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    const authToken = localStorage.getItem('authToken');
    const authSaltString = localStorage.getItem('authSalt');

    const parsedAuthSalt = JSON.parse(authSaltString as string);
    const { token } = parsedAuthSalt;

    const minionIds = selectedIds.map(
      (id) => minions.find((row) => row.id === id)?.minion_id || ''
    );

    try {
      await axios.post(
        `/api/v1/netapi/`,
        [
          {
            client: 'local_async',
            fun: 'saltutil.refresh_grains',
            tgt: minionIds,
            tgt_type: 'list',
          },
          {
            client: 'local_async',
            fun: 'saltutil.refresh_pillar',
            tgt: minionIds,
            tgt_type: 'list',
          },
        ],
        {
          headers: {
            Authorization: authToken,
            'X-Auth-Token': token,
          },
        }
      );
    } catch (err) {
      console.error(`Failed to update cache for minion ids ${minionIds.join(', ')}: `, err);
    } finally {
      setRefreshing(false);
      setSelectedIds([]);
      setSelectAll(false);

      setCurrentPage(0);
      setPage(0);
    }
  };

  const handleDelete = async () => {
    setDeleting(true);
    const authToken = localStorage.getItem('authToken');

    const deleteCache = async (id: string) => {
      try {
        await axios.delete(`/api/v1/salt_cache/uuid/${id}`, {
          headers: {
            Authorization: `${authToken}`,
          },
        });
      } catch (err) {
        console.error(`Failed to delete cache with id ${id}: `, err);
      }
    };

    const selectedMinions = selectedIds.map(
      (id) => minions.find((row) => row.id === id)?.minion_id || ''
    );

    const keyDict = { match: { minions: selectedMinions } };
    await deleteKeys(keyDict);

    await Promise.all(selectedIds.map(deleteCache));
    setDeleting(false);
    setSelectedIds([]);
    setSelectAll(false);

    setCurrentPage(0);
    setPage(0);
  };

  if (isLoading) return <CircularProgress color="success" />;
  if (error) return <div>Error: {error.message}</div>;

  const grainEntries =
    minions.length > 0 && minions[0].grains ? Object.entries(minions[0].grains) : [];

  //   function searchStringInArray (str: string, strArray: string[]) {
  //     for (let j=0; j<strArray.length; j+=1) {
  //         if (strArray[j].match(str)) return strArray[j];
  //     }
  //     return "";
  // }

  return (
    <Box>
      <TableContainer component={Paper}>
        <Table size="small" aria-label="collapsible table">
          <TableHead>
            <TableRow>
              <TableCell padding="checkbox">
                <Checkbox color="primary" checked={selectAll} onChange={handleSelectAll} />
              </TableCell>
              <TableCell sx={{ width: '50px' }}>
                <Box sx={{ display: 'flex', justifyContent: 'left' }}>
                  <Tooltip title="Run Job">
                    <IconButton aria-label="job" onClick={handleRun}>
                      <PlayArrow
                        aria-label="run"
                        onClick={() =>
                          navigate(
                            `/run?tgt=${selectedIds.map(
                              (id) => minions.find((row) => row.id === id)?.minion_id || ''
                            )}&tgt_type=list`
                          )
                        }
                      />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Refresh">
                    <IconButton aria-label="refresh" onClick={handleRefresh} disabled={refreshing}>
                      <RefreshIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Delete">
                    <span>
                      <IconButton
                        aria-label="delete"
                        onClick={handleDelete}
                        disabled={selectAll || deleting}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </span>
                  </Tooltip>
                </Box>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={queryParams.order_by?.startsWith('minion_id')}
                  direction={queryParams.order_by?.endsWith('desc') ? 'desc' : 'asc'}
                  onClick={() => handleRequestSort('minion_id')}
                >
                  Minion ID
                </TableSortLabel>
              </TableCell>
              {grainEntries.map(([key]) => (
                <TableCell key={key}>
                  <TableSortLabel
                    active={queryParams.order_by?.startsWith(key)}
                    direction={queryParams.order_by?.endsWith('desc') ? 'desc' : 'asc'}
                    onClick={() => handleRequestSort(key)}
                  >
                    {String(key)}
                  </TableSortLabel>
                </TableCell>
              ))}
              <TableCell>
                <TableSortLabel
                  active={queryParams.order_by?.startsWith('alter_time')}
                  direction={queryParams.order_by?.endsWith('desc') ? 'desc' : 'asc'}
                  onClick={() => handleRequestSort('alter_time')}
                >
                  Last Check-in
                </TableSortLabel>
              </TableCell>
              {/* <TableCell sx={{ width: '50px' }}>Copy</TableCell> */}
            </TableRow>
          </TableHead>
          <TableBody>
            {minions.map((row) => (
              <Row
                key={row.id}
                row={row}
                selected={selectedIds.includes(row.id)}
                onSelect={handleSelect}
                onDeselect={handleDeselect}
              />
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      <TablePagination
        rowsPerPageOptions={[5, 10, 25, 50, 100, 200]}
        component="div"
        count={totalCount}
        rowsPerPage={rowsPerPage}
        page={currentPage - 1}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
      />
    </Box>
  );
};

export default MinionsTable;
