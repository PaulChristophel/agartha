import { useState } from 'react';

import {
  Box,
  Chip,
  Stack,
  Alert,
  Table,
  Button,
  MenuItem,
  TableRow,
  TextField,
  TableBody,
  TableCell,
  TableHead,
  Typography,
  TableContainer,
  LinearProgress,
  TablePagination,
  CircularProgress,
} from '@mui/material';

import useJobDetails from 'src/hooks/jid/useJobDetails.ts';

function timestamp(value: string | null) {
  if (!value) return 'Unknown';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? 'Unknown' : date.toLocaleString();
}

function Payload({ label, value }: { label: string; value: unknown }) {
  const [open, setOpen] = useState(false);
  return (
    <Box component="details" onToggle={(event) => setOpen(event.currentTarget.open)} sx={{ my: 1 }}>
      <Box component="summary" sx={{ cursor: 'pointer' }}>
        {label}
      </Box>
      {open && (
        <Box
          component="pre"
          sx={{
            maxHeight: 350,
            overflow: 'auto',
            p: 2,
            bgcolor: 'background.default',
            whiteSpace: 'pre-wrap',
            overflowWrap: 'anywhere',
          }}
        >
          {JSON.stringify(value, null, 2)}
        </Box>
      )}
    </Box>
  );
}

export default function JobDetailsPanel({
  jid,
  submittedAt,
}: {
  jid: string;
  submittedAt?: number;
}) {
  const { data, error, isPending, isFetching, refetch, autoRefresh } = useJobDetails(
    jid,
    submittedAt
  );
  const [search, setSearch] = useState('');
  const [status, setStatus] = useState('all');
  const [page, setPage] = useState(0);
  const rows = (data?.returns ?? []).filter(
    (row) =>
      row.id.toLowerCase().includes(search.toLowerCase()) &&
      (status === 'all' || row.success === (status === 'successful'))
  );
  const currentPage = Math.min(page, Math.max(0, Math.ceil(rows.length / 25) - 1));

  return (
    <Stack spacing={2} sx={{ p: 2 }}>
      <Stack direction="row" spacing={2} alignItems="center">
        <Button onClick={() => void refetch()} disabled={isFetching}>
          Refresh
        </Button>
        <Typography variant="body2" color="text.secondary">
          {isFetching
            ? 'Updating…'
            : autoRefresh && !error
              ? 'Refreshes every 5 seconds while visible, for up to 2 minutes.'
              : 'Automatic refresh paused. Refresh to check for new returns.'}
        </Typography>
      </Stack>
      {isPending && (
        <Stack direction="row" spacing={2} alignItems="center">
          <CircularProgress size={24} />
          <Typography>{submittedAt ? 'Waiting for job data…' : 'Loading job details…'}</Typography>
        </Stack>
      )}
      {error && (
        <Alert severity="error">
          {data
            ? 'Could not refresh. Showing previously loaded results.'
            : `Could not load job details: ${error.message}`}
        </Alert>
      )}
      {data && (
        <>
          <Stack direction="row" useFlexGap flexWrap="wrap" gap={1}>
            <Chip label={`Targeted: ${data.targeted_count ?? 'Unknown'}`} />
            <Chip label={`Returned: ${data.returned_count}`} />
            <Chip
              color="success"
              variant="outlined"
              label={`Successful: ${data.successful_count}`}
            />
            <Chip
              color={data.failed_count ? 'error' : 'default'}
              variant="outlined"
              label={`Failed: ${data.failed_count}`}
            />
            <Chip label={`No return received: ${data.pending_count ?? 'Unknown'}`} />
          </Stack>
          {data.targeted_count !== null &&
            data.targeted_count > 0 &&
            data.pending_count !== null && (
              <Box>
                <Typography variant="body2">
                  Target returns received:{' '}
                  {Math.round(
                    (100 * Math.max(0, data.targeted_count - data.pending_count)) /
                      data.targeted_count
                  )}
                  %
                </Typography>
                <LinearProgress
                  aria-label="Target returns received"
                  variant="determinate"
                  value={Math.min(
                    100,
                    Math.max(
                      0,
                      (100 * (data.targeted_count - data.pending_count)) / data.targeted_count
                    )
                  )}
                />
              </Box>
            )}
          <Typography variant="body2">
            Job recorded: {timestamp(data.started_at)} · Last return recorded:{' '}
            {timestamp(data.last_return_at)}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Success reflects the stored return status. Missing returns do not indicate whether a
            minion is still running.
          </Typography>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
            <TextField
              size="small"
              label="Search minion ID"
              value={search}
              onChange={(event) => {
                setSearch(event.target.value);
                setPage(0);
              }}
            />
            <TextField
              select
              size="small"
              label="Return status"
              value={status}
              sx={{ minWidth: 170 }}
              onChange={(event) => {
                setStatus(event.target.value);
                setPage(0);
              }}
            >
              <MenuItem value="all">All returns</MenuItem>
              <MenuItem value="successful">Successful</MenuItem>
              <MenuItem value="failed">Failed</MenuItem>
            </TextField>
          </Stack>
          <TableContainer>
            <Table size="small" aria-label="Job minion returns">
              <TableHead>
                <TableRow>
                  <TableCell>Minion ID</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Return recorded</TableCell>
                  <TableCell>Details</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {rows.slice(currentPage * 25, (currentPage + 1) * 25).map((row) => (
                  <TableRow key={row.id}>
                    <TableCell sx={{ overflowWrap: 'anywhere' }}>{row.id}</TableCell>
                    <TableCell>
                      <Chip
                        size="small"
                        color={row.success ? 'success' : 'error'}
                        label={row.success ? 'Successful' : 'Failed'}
                      />
                    </TableCell>
                    <TableCell>{timestamp(row.alter_time)}</TableCell>
                    <TableCell>
                      <Payload label="Return" value={row.return} />
                      <Payload label="Full return" value={row.full_ret} />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
          {!rows.length && (
            <Typography>
              {data.returns.length ? 'No returns match these filters.' : 'No returns received yet.'}
            </Typography>
          )}
          {rows.length > 25 && (
            <TablePagination
              component="div"
              count={rows.length}
              page={currentPage}
              rowsPerPage={25}
              rowsPerPageOptions={[25]}
              onPageChange={(_, next) => setPage(next)}
            />
          )}
          <Payload label="Original job load" value={data.load} />
        </>
      )}
    </Stack>
  );
}
