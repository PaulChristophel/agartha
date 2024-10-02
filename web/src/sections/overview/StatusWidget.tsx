import React, { useState, useEffect } from 'react';

import { Theme, SxProps } from '@mui/material/styles';
import {
  Card,
  Chip,
  Table,
  Paper,
  TableRow,
  TableBody,
  TableCell,
  Typography,
  TableContainer,
  CircularProgress,
} from '@mui/material';

import useJidPaginated from 'src/hooks/jid/useJidPaginated.ts';
import useEventPaginated from 'src/hooks/saltEvent/useEventPaginated.ts';
import useReturnPaginated from 'src/hooks/saltReturn/useReturnPaginated.ts';

import { fShortenNumber } from 'src/utils/formatNumber.ts';

import { Version } from 'src/config.ts';

import useStatus from './useStatus.ts';
import saltVersion from './saltVersion.ts';

interface StatusWidgetProps {
  sx?: SxProps<Theme>;
  [key: string]: unknown; // To allow other props
}

const StatusWidget: React.FC<StatusWidgetProps> = ({ sx, ...other }) => {
  const [saltVersionData, setSaltVersionData] = useState<string | null>(null);
  const [isSaltVersionLoading, setIsSaltVersionLoading] = useState<boolean>(true);
  const [saltVersionError, setSaltVersionError] = useState<string | null>(null);
  const authToken = localStorage.getItem('authToken') || '';
  const authSaltString = localStorage.getItem('authSalt') || '';

  const { status, isStatusLoading } = useStatus(authToken, authSaltString);

  const queryParams = React.useMemo(() => ({}), []);

  const { isLoading: isJIDLoading, totalCount: totalJIDCount } = useJidPaginated(queryParams, 1, 1);
  const { isLoading: isReturnLoading, totalCount: totalReturnCount } = useReturnPaginated(
    queryParams,
    1,
    1
  );
  const { isLoading: isEventLoading, totalCount: totalEventCount } = useEventPaginated(
    queryParams,
    1,
    1
  );

  useEffect(() => {
    const fetchSaltVersion = async () => {
      setIsSaltVersionLoading(true);
      try {
        const output = await saltVersion(authToken, authSaltString);
        setSaltVersionData(output);
      } catch (error) {
        console.error(`Failed to get salt version: `, error);
        setSaltVersionError(`Failed to get salt version: ${error}`);
      } finally {
        setIsSaltVersionLoading(false);
      }
    };

    if (authToken && authSaltString) {
      fetchSaltVersion();
    } else {
      setSaltVersionError('Missing auth token');
      setIsSaltVersionLoading(false);
    }
  }, [authToken, authSaltString]);

  return (
    <Card
      sx={{
        px: 3,
        py: 3,
        borderRadius: 2,
        ...sx,
      }}
      {...other}
    >
      <Typography variant="h6" gutterBottom>
        Status
      </Typography>
      <TableContainer component={Paper}>
        <Table size="small">
          <TableBody>
            <TableRow>
              <TableCell align="left">Salt Status</TableCell>
              <TableCell align="right">
                {isStatusLoading ? (
                  <CircularProgress size={20} />
                ) : (
                  <Chip
                    label={status}
                    color={status === 'OK' ? 'success' : 'error'}
                    variant="outlined"
                  />
                )}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell align="left">Salt</TableCell>
              <TableCell align="right">
                {isSaltVersionLoading ? (
                  <CircularProgress size={20} />
                ) : (
                  saltVersionError || saltVersionData
                )}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell align="left">Agartha</TableCell>
              <TableCell align="right">{Version}</TableCell>
            </TableRow>
            <TableRow>
              <TableCell align="left">Jobs</TableCell>
              <TableCell align="right">
                {isJIDLoading ? <CircularProgress size={20} /> : fShortenNumber(totalJIDCount)}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell align="left">Events</TableCell>
              <TableCell align="right">
                {isEventLoading ? <CircularProgress size={20} /> : fShortenNumber(totalEventCount)}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell align="left">Returns</TableCell>
              <TableCell align="right">
                {isReturnLoading ? (
                  <CircularProgress size={20} />
                ) : (
                  fShortenNumber(totalReturnCount)
                )}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    </Card>
  );
};

export default StatusWidget;
