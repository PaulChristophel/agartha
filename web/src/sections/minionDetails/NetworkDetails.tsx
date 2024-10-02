import React from 'react';

import Box from '@mui/material/Box';
import Table from '@mui/material/Table';
import Paper from '@mui/material/Paper';
import Divider from '@mui/material/Divider';
import TableRow from '@mui/material/TableRow';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import Typography from '@mui/material/Typography';
import TableContainer from '@mui/material/TableContainer';
import CircularProgress from '@mui/material/CircularProgress';

interface DnsData {
  domain?: string;
  search?: string[];
  options?: string[];
  sortlist?: string[];
  nameservers?: string[];
  ip4_nameservers?: string[];
  ip6_nameservers?: string[];
}

interface GrainData {
  dns?: DnsData;
  ip_interfaces?: Record<string, string[]>;
  hwaddr_interfaces?: Record<string, string>;
  ipv4?: string[];
  ipv6?: string[];
}

interface NetworkDetailsProps {
  grainData: GrainData | null; // Allow for null initially
}

const NetworkDetails: React.FC<NetworkDetailsProps> = ({ grainData }) => {
  if (!grainData) {
    return <CircularProgress color="success" />; // Render loading state if grainData is not available
  }

  const { dns = {}, ip_interfaces = {}, hwaddr_interfaces = {}, ipv4 = [], ipv6 = [] } = grainData;

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Network Details
      </Typography>
      <Divider />
      <TableContainer component={Paper}>
        <Table size="small">
          <TableBody>
            {Object.entries(hwaddr_interfaces)
              .filter(([iface]) => iface !== 'lo') // Exclude 'lo' interface
              .map(([iface, hwaddr]) => (
                <TableRow key={iface}>
                  <TableCell align="left">{iface} MAC</TableCell>
                  <TableCell align="right">{hwaddr}</TableCell>
                </TableRow>
              ))}
            {Object.entries(ip_interfaces)
              .filter(([iface]) => iface !== 'lo') // Exclude 'lo' interface
              .map(([iface, ips]) => (
                <TableRow key={iface}>
                  <TableCell align="left">{iface} IP</TableCell>
                  <TableCell align="right">{ips.join(', ')}</TableCell>
                </TableRow>
              ))}
            {ipv4 && ipv4.length > 1 && (
              <TableRow>
                <TableCell align="left">IPv4</TableCell>
                <TableCell align="right">
                  {ipv4.filter((element: string) => element !== '127.0.0.1').join(', ') ?? 'N/A'}
                </TableCell>
              </TableRow>
            )}
            {ipv6 && ipv6.length > 1 && (
              <TableRow>
                <TableCell align="left">IPv6</TableCell>
                <TableCell align="right">
                  {ipv6.filter((element: string) => element !== '::1').join(', ') ?? 'N/A'}
                </TableCell>
              </TableRow>
            )}
            {dns.domain && dns.domain !== '' && (
              <TableRow>
                <TableCell align="left">Domain</TableCell>
                <TableCell align="right">{dns.domain ?? 'N/A'}</TableCell>
              </TableRow>
            )}
            {dns.search && dns.search.length > 0 && (
              <TableRow>
                <TableCell align="left">Search</TableCell>
                <TableCell align="right">{dns.search ? dns.search.join(', ') : 'N/A'}</TableCell>
              </TableRow>
            )}
            {dns.options && dns.options.length > 0 && (
              <TableRow>
                <TableCell align="left">Options</TableCell>
                <TableCell align="right">{dns.options ? dns.options.join(', ') : 'N/A'}</TableCell>
              </TableRow>
            )}
            {dns.sortlist && dns.sortlist.length > 0 && (
              <TableRow>
                <TableCell align="left">Sortlist</TableCell>
                <TableCell align="right">
                  {dns.sortlist ? dns.sortlist.join(', ') : 'N/A'}
                </TableCell>
              </TableRow>
            )}
            {dns.nameservers && dns.nameservers.length > 0 && (
              <TableRow>
                <TableCell align="left">Nameservers</TableCell>
                <TableCell align="right">
                  {dns.nameservers ? dns.nameservers.join(', ') : 'N/A'}
                </TableCell>
              </TableRow>
            )}
            {dns.ip4_nameservers && dns.ip4_nameservers.length > 0 && (
              <TableRow>
                <TableCell align="left">IPv4 Nameservers</TableCell>
                <TableCell align="right">
                  {dns.ip4_nameservers ? dns.ip4_nameservers.join(', ') : 'N/A'}
                </TableCell>
              </TableRow>
            )}
            {dns.ip6_nameservers && dns.ip6_nameservers.length > 0 && (
              <TableRow>
                <TableCell align="left">IPv6 Nameservers</TableCell>
                <TableCell align="right">
                  {dns.ip6_nameservers ? dns.ip6_nameservers.join(', ') : 'N/A'}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
};

export default NetworkDetails;
