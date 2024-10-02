import React, { useEffect, useCallback } from 'react';

import Box from '@mui/material/Box';
import Table from '@mui/material/Table';
import Paper from '@mui/material/Paper';
import TableRow from '@mui/material/TableRow';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableContainer from '@mui/material/TableContainer';
import TableSortLabel from '@mui/material/TableSortLabel';
import TablePagination from '@mui/material/TablePagination';
import CircularProgress from '@mui/material/CircularProgress';

import useConformityPaginated from 'src/hooks/conformity/useConformityPaginated.ts';

import Row from './Row.tsx';

interface ConformityTableProps {
  queryParams: {
    id?: string;
    jid?: string;
    fun?: string;
    success?: boolean;
    since?: string;
    until?: string;
    limit?: number;
    page?: number;
    order_by?: string;
  };
  setLimit: (newLimit: number) => void;
  setPage: (newPage: number) => void;
  setOrderBy: (orderBy: string) => void;
}

const ConformityTable: React.FC<ConformityTableProps> = ({
  queryParams,
  setLimit,
  setPage,
  setOrderBy,
}) => {
  const {
    returns,
    isLoading,
    error,
    currentPage,
    rowsPerPage,
    setCurrentPage,
    setRowsPerPage,
    totalCount,
  } = useConformityPaginated(queryParams, queryParams.page || 1, queryParams.limit || 50);

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
      const isAsc = queryParams.order_by === `${property} asc`;
      setOrderBy(`${property} ${isAsc ? 'desc' : 'asc'}`);
    },
    [queryParams.order_by, setOrderBy]
  );

  if (isLoading) return <CircularProgress color="success" />;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <Box>
      <TableContainer component={Paper}>
        <Table size="small" aria-label="collapsible table">
          <TableHead>
            <TableRow>
              <TableCell />
              <TableCell>
                <TableSortLabel
                  active={queryParams.order_by?.startsWith('id')}
                  direction={queryParams.order_by?.endsWith('desc') ? 'desc' : 'asc'}
                  onClick={() => handleRequestSort('id')}
                >
                  Minion ID
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={queryParams.order_by?.startsWith('true_count')}
                  direction={queryParams.order_by?.endsWith('desc') ? 'desc' : 'asc'}
                  onClick={() => handleRequestSort('true_count')}
                >
                  Successes
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={queryParams.order_by?.startsWith('false_count')}
                  direction={queryParams.order_by?.endsWith('desc') ? 'desc' : 'asc'}
                  onClick={() => handleRequestSort('false_count')}
                >
                  Failures
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={queryParams.order_by?.startsWith('changed_count')}
                  direction={queryParams.order_by?.endsWith('desc') ? 'desc' : 'asc'}
                  onClick={() => handleRequestSort('changed_count')}
                >
                  Changed
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={queryParams.order_by?.startsWith('unchanged_count')}
                  direction={queryParams.order_by?.endsWith('desc') ? 'desc' : 'asc'}
                  onClick={() => handleRequestSort('unchanged_count')}
                >
                  Unchanged
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={queryParams.order_by?.startsWith('success')}
                  direction={queryParams.order_by?.endsWith('desc') ? 'desc' : 'asc'}
                  onClick={() => handleRequestSort('success')}
                >
                  Conforming
                </TableSortLabel>
              </TableCell>
              <TableCell>
                <TableSortLabel
                  active={queryParams.order_by?.startsWith('alter_time')}
                  direction={queryParams.order_by?.endsWith('desc') ? 'desc' : 'asc'}
                  onClick={() => handleRequestSort('alter_time')}
                >
                  Check-in Time
                </TableSortLabel>
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {returns.map((row) => (
              <Row key={`${row.id}`} row={row} />
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

export default ConformityTable;
