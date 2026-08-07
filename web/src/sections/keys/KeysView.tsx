import React from 'react';

import Box from '@mui/material/Box';

import useKeys from 'src/hooks/netapi/wheel/useKeys.ts';
import useAcceptKeyDict from 'src/hooks/netapi/wheel/useAcceptKeyDict.ts';
import useDeleteKeyDict from 'src/hooks/netapi/wheel/useDeleteKeyDict.ts';
import useRejectKeyDict from 'src/hooks/netapi/wheel/useRejectKeyDict.ts';

import KeysTable, { Order, KeyRow, KeyState } from './KeysTable.tsx';

interface SaltKeyMatch {
  minions?: string[];
  minions_denied?: string[];
  minions_pre?: string[];
  minions_rejected?: string[];
}

const stateOrder: Record<KeyState, number> = {
  pending: 4,
  denied: 3,
  rejected: 2,
  accepted: 1,
};

const saltKeyBucketByState: Record<KeyState, keyof SaltKeyMatch> = {
  accepted: 'minions',
  denied: 'minions_denied',
  pending: 'minions_pre',
  rejected: 'minions_rejected',
};

function createRow(id: number, name: string, state: KeyState): KeyRow {
  return { id, name, state };
}

function descendingComparator(a: KeyRow, b: KeyRow, orderBy: keyof KeyRow) {
  if (orderBy === 'state') {
    return stateOrder[a.state] - stateOrder[b.state];
  }
  if (b[orderBy] < a[orderBy]) return -1;
  if (b[orderBy] > a[orderBy]) return 1;
  return 0;
}

function getComparator(order: Order, orderBy: keyof KeyRow) {
  return order === 'desc'
    ? (a: KeyRow, b: KeyRow) => descendingComparator(a, b, orderBy)
    : (a: KeyRow, b: KeyRow) => -descendingComparator(a, b, orderBy);
}

function stableSort(rows: readonly KeyRow[], comparator: (a: KeyRow, b: KeyRow) => number) {
  return rows
    .map((row, index) => ({ row, index }))
    .sort((a, b) => comparator(a.row, b.row) || a.index - b.index)
    .map(({ row }) => row);
}

interface KeysViewProps {
  reload?: () => void;
}

export default function KeysView({ reload = () => window.location.reload() }: KeysViewProps = {}) {
  const [order, setOrder] = React.useState<Order>('asc');
  const [orderBy, setOrderBy] = React.useState<keyof KeyRow>('state');
  const [selected, setSelected] = React.useState<readonly number[]>([]);
  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(50);
  const [filterText, setFilterText] = React.useState('');
  const { minions, minionsDenied, minionsPre, minionsRejected, isLoading } = useKeys();
  const { acceptKeys } = useAcceptKeyDict();
  const { deleteKeys } = useDeleteKeyDict();
  const { rejectKeys } = useRejectKeyDict();

  const rows = React.useMemo(() => {
    const combinedRows = [
      ...minions.map((name, index) => createRow(index, name, 'accepted')),
      ...minionsDenied.map((name, index) => createRow(index + minions.length, name, 'denied')),
      ...minionsPre.map((name, index) =>
        createRow(index + minions.length + minionsDenied.length, name, 'pending')
      ),
      ...minionsRejected.map((name, index) =>
        createRow(
          index + minions.length + minionsDenied.length + minionsPre.length,
          name,
          'rejected'
        )
      ),
    ];
    const normalizedFilter = filterText.toLowerCase();
    return combinedRows.filter((row) => row.name.toLowerCase().includes(normalizedFilter));
  }, [filterText, minions, minionsDenied, minionsPre, minionsRejected]);

  const selectedRows = () =>
    selected
      .map((id) => rows.find((row) => row.id === id))
      .filter((row): row is KeyRow => Boolean(row));

  const selectedKeyMatch = () =>
    selectedRows().reduce<SaltKeyMatch>((match, row) => {
      const bucket = saltKeyBucketByState[row.state];
      match[bucket] = [...(match[bucket] ?? []), row.name];
      return match;
    }, {});

  const handleAcceptKeys = async () => {
    await acceptKeys({
      match: { minions: selectedRows().map((row) => row.name) },
      include_rejected: true,
      include_denied: true,
    });
    reload();
  };

  const handleDeleteKeys = async () => {
    await deleteKeys({ match: selectedKeyMatch() });
    reload();
  };

  const handleRejectKeys = async () => {
    await rejectKeys({
      match: selectedKeyMatch(),
      include_accepted: true,
      include_denied: true,
    });
    reload();
  };

  const visibleRows = React.useMemo(
    () =>
      stableSort(rows, getComparator(order, orderBy)).slice(
        page * rowsPerPage,
        page * rowsPerPage + rowsPerPage
      ),
    [order, orderBy, page, rows, rowsPerPage]
  );
  const emptyRows = page > 0 ? Math.max(0, (page + 1) * rowsPerPage - rows.length) : 0;

  return (
    <Box sx={{ width: '100%' }}>
      <KeysTable
        rows={visibleRows}
        totalRows={rows.length}
        isLoading={isLoading}
        selected={selected}
        order={order}
        orderBy={orderBy}
        page={page}
        rowsPerPage={rowsPerPage}
        emptyRows={emptyRows}
        filterText={filterText}
        onFilterTextChange={(event) => setFilterText(event.target.value)}
        onAcceptKeys={handleAcceptKeys}
        onDeleteKeys={handleDeleteKeys}
        onRejectKeys={handleRejectKeys}
        onSelectAllClick={(event) =>
          setSelected(event.target.checked ? rows.map(({ id }) => id) : [])
        }
        onRequestSort={(_event, property) => {
          setOrder(orderBy === property && order === 'asc' ? 'desc' : 'asc');
          setOrderBy(property);
        }}
        onRowClick={(_event, id) =>
          setSelected((current) =>
            current.includes(id)
              ? current.filter((selectedId) => selectedId !== id)
              : [...current, id]
          )
        }
        onPageChange={(_event, nextPage) => setPage(nextPage)}
        onRowsPerPageChange={(event) => {
          setRowsPerPage(parseInt(event.target.value, 10));
          setPage(0);
        }}
      />
    </Box>
  );
}
