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

import useKeys from 'src/hooks/netapi/wheel/useKeys.ts';
import useAcceptKeyDict from 'src/hooks/netapi/wheel/useAcceptKeyDict.ts';
import useDeleteKeyDict from 'src/hooks/netapi/wheel/useDeleteKeyDict.ts';
import useRejectKeyDict from 'src/hooks/netapi/wheel/useRejectKeyDict.ts';

interface Data {
  id: number;
  name: string;
  state: string;
}

function createData(id: number, name: string, state: string): Data {
  return {
    id,
    name,
    state,
  };
}

const stateOrder: { [key: string]: number } = {
  pending: 4,
  denied: 3,
  rejected: 2,
  accepted: 1,
};

function descendingComparator<T>(a: T, b: T, orderBy: keyof T) {
  if (orderBy === 'state') {
    return (
      stateOrder[a[orderBy] as unknown as string] - stateOrder[b[orderBy] as unknown as string]
    );
  }
  if (b[orderBy] < a[orderBy]) {
    return -1;
  }
  if (b[orderBy] > a[orderBy]) {
    return 1;
  }
  return 0;
}

type Order = 'asc' | 'desc';

function getComparator<Key extends keyof Data>(
  order: Order,
  orderBy: Key
): (a: Data, b: Data) => number {
  return order === 'desc'
    ? (a, b) => descendingComparator(a, b, orderBy)
    : (a, b) => -descendingComparator(a, b, orderBy);
}

function stableSort<T>(array: readonly T[], comparator: (a: T, b: T) => number) {
  const stabilizedThis = array.map((el, index) => [el, index] as [T, number]);
  stabilizedThis.sort((a, b) => {
    const order = comparator(a[0], b[0]);
    if (order !== 0) {
      return order;
    }
    return a[1] - b[1];
  });
  return stabilizedThis.map((el) => el[0]);
}

interface HeadCell {
  disablePadding: boolean;
  id: keyof Data;
  label: string;
  numeric: boolean;
}

const headCells: readonly HeadCell[] = [
  {
    id: 'name',
    numeric: false,
    disablePadding: true,
    label: 'Minion ID',
  },
  {
    id: 'state',
    numeric: false,
    disablePadding: false,
    label: 'State',
  },
];

interface KeysProps {
  numSelected: number;
  onRequestSort: (event: React.MouseEvent<unknown>, property: keyof Data) => void;
  onSelectAllClick: (event: React.ChangeEvent<HTMLInputElement>) => void;
  order: Order;
  orderBy: string;
  rowCount: number;
}

function KeysHead(props: KeysProps) {
  const { onSelectAllClick, order, orderBy, numSelected, rowCount, onRequestSort } = props;
  const createSortHandler = (property: keyof Data) => (event: React.MouseEvent<unknown>) => {
    onRequestSort(event, property);
  };

  return (
    <TableHead>
      <TableRow>
        <TableCell padding="checkbox">
          <Checkbox
            color="primary"
            indeterminate={numSelected > 0 && numSelected < rowCount}
            checked={rowCount > 0 && numSelected === rowCount}
            onChange={onSelectAllClick}
            inputProps={{
              'aria-label': 'select all minions',
            }}
          />
        </TableCell>
        {headCells.map((headCell) => (
          <TableCell
            key={headCell.id}
            align={headCell.numeric ? 'right' : 'left'}
            padding={headCell.disablePadding ? 'none' : 'normal'}
            sortDirection={orderBy === headCell.id ? order : false}
          >
            <TableSortLabel
              active={orderBy === headCell.id}
              direction={orderBy === headCell.id ? order : 'asc'}
              onClick={createSortHandler(headCell.id)}
            >
              {headCell.label}
              {orderBy === headCell.id ? (
                <Box component="span" sx={visuallyHidden}>
                  {order === 'desc' ? 'sorted descending' : 'sorted ascending'}
                </Box>
              ) : null}
            </TableSortLabel>
          </TableCell>
        ))}
      </TableRow>
    </TableHead>
  );
}

interface KeysToolbarProps {
  numSelected: number;
  filterText: string;
  onFilterTextChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
  onAcceptKeys: () => void;
  onDeleteKeys: () => void;
  onRejectKeys: () => void;
}

function KeysToolbar(props: KeysToolbarProps) {
  const { numSelected, filterText, onFilterTextChange, onAcceptKeys, onDeleteKeys, onRejectKeys } =
    props;

  return (
    <Toolbar
      sx={{
        pl: { sm: 2 },
        pr: { xs: 1, sm: 1 },
        ...(numSelected > 0 && {
          bgcolor: (theme) =>
            alpha(theme.palette.primary.main, theme.palette.action.activatedOpacity),
        }),
      }}
    >
      {numSelected > 0 ? (
        <Typography sx={{ flex: '1 1 100%' }} color="inherit" variant="subtitle1" component="div">
          {numSelected} selected
        </Typography>
      ) : (
        <Typography sx={{ flex: '1 1 100%' }} variant="h6" id="minionKeys" component="div">
          Keys
        </Typography>
      )}
      {numSelected > 0 ? (
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
            variant="outlined"
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
  );
}

export default function KeysView() {
  const [order, setOrder] = React.useState<Order>('asc');
  const [orderBy, setOrderBy] = React.useState<keyof Data>('state');
  const [selected, setSelected] = React.useState<readonly number[]>([]);
  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(50);
  const [filterText, setFilterText] = React.useState<string>('');
  const { minions, minionsDenied, minionsPre, minionsRejected, isLoading } = useKeys();
  const { acceptKeys } = useAcceptKeyDict();
  const { deleteKeys } = useDeleteKeyDict();
  const { rejectKeys } = useRejectKeyDict();

  const handleRequestSort = (_event: React.MouseEvent<unknown>, property: keyof Data) => {
    const isAsc = orderBy === property && order === 'asc';
    setOrder(isAsc ? 'desc' : 'asc');
    setOrderBy(property);
  };

  const handleSelectAllClick = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      const newSelected = minionsData.map((n) => n.id);
      setSelected(newSelected);
      return;
    }
    setSelected([]);
  };

  const handleClick = (_event: React.MouseEvent<unknown>, id: number) => {
    const selectedIndex = selected.indexOf(id);
    let newSelected: readonly number[] = [];

    if (selectedIndex === -1) {
      newSelected = newSelected.concat(selected, id);
    } else if (selectedIndex === 0) {
      newSelected = newSelected.concat(selected.slice(1));
    } else if (selectedIndex === selected.length - 1) {
      newSelected = newSelected.concat(selected.slice(0, -1));
    } else if (selectedIndex > 0) {
      newSelected = newSelected.concat(
        selected.slice(0, selectedIndex),
        selected.slice(selectedIndex + 1)
      );
    }
    setSelected(newSelected);
  };

  const handleChangePage = (_event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  const handleFilterTextChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setFilterText(event.target.value);
  };

  const handleAcceptKeys = async () => {
    const selectedMinions = selected.map(
      (id) => minionsData.find((row) => row.id === id)?.name || ''
    );
    const keyDict = {
      match: { minions: selectedMinions },
      include_rejected: true,
      include_denied: true,
    };
    await acceptKeys(keyDict);
    window.location.reload();
  };

  const handleDeleteKeys = async () => {
    const selectedMinions = selected.map(
      (id) => minionsData.find((row) => row.id === id)?.name || ''
    );
    const keyDict = { match: { minions: selectedMinions } };
    await deleteKeys(keyDict);
    window.location.reload();
  };

  const handleRejectKeys = async () => {
    const selectedMinions = selected.map(
      (id) => minionsData.find((row) => row.id === id)?.name || ''
    );
    const keyDict = {
      match: { minions: selectedMinions },
      include_accepted: true,
      include_denied: true,
    };
    await rejectKeys(keyDict);
    window.location.reload();
  };

  const isSelected = (id: number) => selected.indexOf(id) !== -1;

  const minionsData = React.useMemo(() => {
    const combinedData = [
      ...minions.map((name, index) => createData(index, name, 'accepted')),
      ...minionsDenied.map((name, index) => createData(index + minions.length, name, 'denied')),
      ...minionsPre.map((name, index) =>
        createData(index + minions.length + minionsDenied.length, name, 'pending')
      ),
      ...minionsRejected.map((name, index) =>
        createData(
          index + minions.length + minionsDenied.length + minionsPre.length,
          name,
          'rejected'
        )
      ),
    ];

    return combinedData.filter((row) => row.name.toLowerCase().includes(filterText.toLowerCase()));
  }, [minions, minionsDenied, minionsPre, minionsRejected, filterText]);

  // Avoid a layout jump when reaching the last page with empty rows.
  const emptyRows = page > 0 ? Math.max(0, (1 + page) * rowsPerPage - minionsData.length) : 0;

  const visibleRows = React.useMemo(
    () =>
      stableSort(minionsData, getComparator(order, orderBy)).slice(
        page * rowsPerPage,
        page * rowsPerPage + rowsPerPage
      ),
    [order, orderBy, page, rowsPerPage, minionsData]
  );

  return (
    <Box sx={{ width: '100%' }}>
      {isLoading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', mt: 5 }}>
          <CircularProgress color="success" />
        </Box>
      ) : (
        <Paper sx={{ width: '100%', mb: 2 }}>
          <KeysToolbar
            numSelected={selected.length}
            filterText={filterText}
            onFilterTextChange={handleFilterTextChange}
            onAcceptKeys={handleAcceptKeys}
            onDeleteKeys={handleDeleteKeys}
            onRejectKeys={handleRejectKeys}
          />
          <TableContainer>
            <Table sx={{ minWidth: 750 }} aria-labelledby="minionKeys" size="small">
              <KeysHead
                numSelected={selected.length}
                order={order}
                orderBy={orderBy}
                onSelectAllClick={handleSelectAllClick}
                onRequestSort={handleRequestSort}
                rowCount={minionsData.length}
              />
              <TableBody>
                {visibleRows.map((row, index) => {
                  const isItemSelected = isSelected(row.id);
                  const labelId = `enhanced-table-checkbox-${index}`;

                  return (
                    <TableRow
                      hover
                      onClick={(event) => handleClick(event, row.id)}
                      role="checkbox"
                      aria-checked={isItemSelected}
                      tabIndex={-1}
                      key={row.id}
                      selected={isItemSelected}
                      sx={{ cursor: 'pointer' }}
                    >
                      <TableCell padding="checkbox">
                        <Checkbox
                          color="primary"
                          checked={isItemSelected}
                          inputProps={{
                            'aria-labelledby': labelId,
                          }}
                        />
                      </TableCell>
                      <TableCell component="th" id={labelId} scope="row" padding="none">
                        {row.name}
                      </TableCell>
                      <TableCell align="left">
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
                  <TableRow
                    style={{
                      height: 33 * emptyRows,
                    }}
                  >
                    <TableCell colSpan={6} />
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>
          <TablePagination
            rowsPerPageOptions={[50, 100, 250, 500, 1000]}
            component="div"
            count={minionsData.length}
            rowsPerPage={rowsPerPage}
            page={page}
            onPageChange={handleChangePage}
            onRowsPerPageChange={handleChangeRowsPerPage}
          />
        </Paper>
      )}
    </Box>
  );
}
