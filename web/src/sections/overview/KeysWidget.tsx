import React from 'react';

import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import Table from '@mui/material/Table';
import Paper from '@mui/material/Paper';
import TableRow from '@mui/material/TableRow';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import Typography from '@mui/material/Typography';
import CheckIcon from '@mui/icons-material/Check';
import { Theme, SxProps } from '@mui/material/styles';
import TableContainer from '@mui/material/TableContainer';
import DoDisturbIcon from '@mui/icons-material/DoDisturb';
import CircularProgress from '@mui/material/CircularProgress';
import PendingActionsIcon from '@mui/icons-material/PendingActions';
import ThumbDownOffAltIcon from '@mui/icons-material/ThumbDownOffAlt';

import useKeys from 'src/hooks/netapi/wheel/useKeys.ts';

interface KeysWidgetProps {
  sx?: SxProps<Theme>;
  [key: string]: unknown; // To allow other props
}

const KeysWidget: React.FC<KeysWidgetProps> = ({ sx, ...other }) => {
  const { minions, minionsDenied, minionsPre, minionsRejected, isLoading } = useKeys();

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
      <Box>
        <Typography variant="h6" gutterBottom>
          Keys
        </Typography>
        <TableContainer component={Paper}>
          <Table size="small">
            <TableBody>
              <TableRow>
                <TableCell align="left">
                  <Box display="flex" alignItems="center">
                    <CheckIcon />
                    <Typography variant="body2" color="text.secondary" ml={1}>
                      Accepted
                    </Typography>
                  </Box>
                </TableCell>
                <TableCell align="right">
                  {isLoading ? <CircularProgress size={20} /> : minions.length}
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell align="left">
                  <Box display="flex" alignItems="center">
                    <DoDisturbIcon />
                    <Typography variant="body2" color="text.secondary" ml={1}>
                      Rejected
                    </Typography>
                  </Box>
                </TableCell>
                <TableCell align="right">
                  {isLoading ? <CircularProgress size={20} /> : minionsRejected.length}
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell align="left">
                  <Box display="flex" alignItems="center">
                    <ThumbDownOffAltIcon />
                    <Typography variant="body2" color="text.secondary" ml={1}>
                      Denied
                    </Typography>
                  </Box>
                </TableCell>
                <TableCell align="right">
                  {isLoading ? <CircularProgress size={20} /> : minionsDenied.length}
                </TableCell>
              </TableRow>
              <TableRow>
                <TableCell align="left">
                  <Box display="flex" alignItems="center">
                    <PendingActionsIcon />
                    <Typography variant="body2" color="text.secondary" ml={1}>
                      Unaccepted
                    </Typography>
                  </Box>
                </TableCell>
                <TableCell align="right">
                  {isLoading ? <CircularProgress size={20} /> : minionsPre.length}
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </TableContainer>
      </Box>
    </Card>
  );
};

export default KeysWidget;
