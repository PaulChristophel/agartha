import React from 'react';

import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Table from '@mui/material/Table';
import Paper from '@mui/material/Paper';
import Toolbar from '@mui/material/Toolbar';
import Tooltip from '@mui/material/Tooltip';
import { visuallyHidden } from '@mui/utils';
import { alpha } from '@mui/material/styles';
import TableRow from '@mui/material/TableRow';
import Checkbox from '@mui/material/Checkbox';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import DeleteIcon from '@mui/icons-material/Delete';
import CancelIcon from '@mui/icons-material/Cancel';
import TableContainer from '@mui/material/TableContainer';
import TableSortLabel from '@mui/material/TableSortLabel';
import TablePagination from '@mui/material/TablePagination';
import FilterListIcon from '@mui/icons-material/FilterList';
import CircularProgress from '@mui/material/CircularProgress';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';

export type KeyState = 'accepted' | 'denied' | 'pending' | 'rejected';
export type Order = 'asc' | 'desc';

export interface KeyRow {
  id: number;
  name: string;
  state: KeyState;
}

interface KeysTableProps {
  rows: KeyRow[];
  totalRows: number;
  isLoading: boolean;
  selected: readonly number[];
  order: Order;
  orderBy: keyof KeyRow;
  page: number;
  rowsPerPage: number;
  emptyRows: number;
  filterText: string;
  onFilterTextChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  onAcceptKeys: () => void;
  onDeleteKeys: () => void;
  onRejectKeys: () => void;
  onSelectAllClick: (event: React.ChangeEvent<HTMLInputElement>) => void;
  onRequestSort: (event: React.MouseEvent<unknown>, property: keyof KeyRow) => void;
  onRowClick: (event: React.MouseEvent<unknown>, id: number) => void;
  onPageChange: (event: unknown, page: number) => void;
  onRowsPerPageChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
}

const headCells: readonly { id: keyof KeyRow; label: string; disablePadding: boolean }[] = [
  { id: 'name', label: 'Minion ID', disablePadding: true },
  { id: 'state', label: 'State', disablePadding: false },
];

export default function KeysTable({
  rows,
  totalRows,
  isLoading,
  selected,
  order,
  orderBy,
  page,
  rowsPerPage,
  emptyRows,
  filterText,
  onFilterTextChange,
  onAcceptKeys,
  onDeleteKeys,
  onRejectKeys,
  onSelectAllClick,
  onRequestSort,
  onRowClick,
  onPageChange,
  onRowsPerPageChange,
}: KeysTableProps) {
  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', mt: 5 }}>
        <CircularProgress color="success" />
      </Box>
    );
  }

  return (
    <Paper sx={{ width: '100%', mb: 2 }}>
      <Toolbar
        sx={{
          pl: { sm: 2 },
          pr: { xs: 1, sm: 1 },
          ...(selected.length > 0 && {
            bgcolor: (theme) =>
              alpha(theme.palette.primary.main, theme.palette.action.activatedOpacity),
          }),
        }}
      >
        <Typography sx={{ flex: '1 1 100%' }} variant={selected.length ? 'subtitle1' : 'h6'}>
          {selected.length ? `${selected.length} selected` : 'Keys'}
        </Typography>
        {selected.length ? (
          <>
            <Tooltip title="Accept">
              <IconButton onClick={onAcceptKeys}>
                <CheckCircleIcon />
              </IconButton>
            </Tooltip>
            <Tooltip title="Reject">
              <IconButton onClick={onRejectKeys}>
                <CancelIcon />
              </IconButton>
            </Tooltip>
            <Tooltip title="Delete">
              <IconButton onClick={onDeleteKeys}>
                <DeleteIcon />
              </IconButton>
            </Tooltip>
          </>
        ) : (
          <>
            <TextField
              value={filterText}
              onChange={onFilterTextChange}
              label="Filter by Minion ID"
              size="small"
            />
            <Tooltip title="Filter list">
              <IconButton>
                <FilterListIcon />
              </IconButton>
            </Tooltip>
          </>
        )}
      </Toolbar>
      <TableContainer>
        <Table sx={{ minWidth: 750 }} aria-label="minion keys" size="small">
          <TableHead>
            <TableRow>
              <TableCell padding="checkbox">
                <Checkbox
                  color="primary"
                  indeterminate={selected.length > 0 && selected.length < totalRows}
                  checked={totalRows > 0 && selected.length === totalRows}
                  onChange={onSelectAllClick}
                  inputProps={{ 'aria-label': 'select all minions' }}
                />
              </TableCell>
              {headCells.map((headCell) => (
                <TableCell
                  key={headCell.id}
                  padding={headCell.disablePadding ? 'none' : 'normal'}
                  sortDirection={orderBy === headCell.id ? order : false}
                >
                  <TableSortLabel
                    active={orderBy === headCell.id}
                    direction={orderBy === headCell.id ? order : 'asc'}
                    onClick={(event) => onRequestSort(event, headCell.id)}
                  >
                    {headCell.label}
                    {orderBy === headCell.id && (
                      <Box component="span" sx={visuallyHidden}>
                        {order === 'desc' ? 'sorted descending' : 'sorted ascending'}
                      </Box>
                    )}
                  </TableSortLabel>
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {rows.map((row, index) => {
              const isSelected = selected.includes(row.id);
              const labelId = `enhanced-table-checkbox-${index}`;

              return (
                <TableRow
                  hover
                  onClick={(event) => onRowClick(event, row.id)}
                  role="checkbox"
                  aria-checked={isSelected}
                  tabIndex={-1}
                  key={row.id}
                  selected={isSelected}
                  sx={{ cursor: 'pointer' }}
                >
                  <TableCell padding="checkbox">
                    <Checkbox
                      color="primary"
                      checked={isSelected}
                      inputProps={{ 'aria-labelledby': labelId }}
                    />
                  </TableCell>
                  <TableCell component="th" id={labelId} scope="row" padding="none">
                    {row.name}
                  </TableCell>
                  <TableCell>
                    <Chip
                      variant="outlined"
                      label={row.state}
                      color={row.state === 'accepted' ? 'success' : 'error'}
                    />
                  </TableCell>
                </TableRow>
              );
            })}
            {emptyRows > 0 && (
              <TableRow style={{ height: 33 * emptyRows }}>
                <TableCell colSpan={6} />
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
      <TablePagination
        rowsPerPageOptions={[50, 100, 250, 500, 1000]}
        component="div"
        count={totalRows}
        rowsPerPage={rowsPerPage}
        page={page}
        onPageChange={onPageChange}
        onRowsPerPageChange={onRowsPerPageChange}
      />
    </Paper>
  );
}
