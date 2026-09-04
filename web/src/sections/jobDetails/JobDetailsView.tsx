import { Link, useParams, useLocation } from 'react-router-dom';

import { Stack, Typography } from '@mui/material';

import JobDetailsPanel from '../jobs/JobDetailsPanel.tsx';

export default function JidDetailsPage() {
  const { jid = '' } = useParams();
  const { state } = useLocation();
  const submittedAt = typeof state?.submittedAt === 'number' ? state.submittedAt : undefined;
  return (
    <Stack spacing={2}>
      <Typography variant="h4">Job details</Typography>
      <Typography sx={{ overflowWrap: 'anywhere' }}>Job ID: {jid}</Typography>
      <Link to={`/returns/?jid=${encodeURIComponent(jid)}`}>Browse returns</Link>
      <JobDetailsPanel key={jid} jid={jid} submittedAt={submittedAt} />
    </Stack>
  );
}
