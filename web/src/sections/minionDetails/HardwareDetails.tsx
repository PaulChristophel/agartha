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

interface GPU {
  model: string;
  vendor: string;
}

interface GrainData {
  cpu_model?: string;
  gpus?: GPU[];
  num_cpus?: number;
  num_gpus?: number;
  mem_total?: number;
  swap_total?: number;
  host?: string;
  manufacturer?: string;
  ipv4?: string[];
  ipv6?: string[];
}

interface HardwareDetailsProps {
  grainData: GrainData | null; // Allow for null initially
}

const HardwareDetails: React.FC<HardwareDetailsProps> = ({ grainData }) => {
  if (!grainData) {
    return <CircularProgress color="success" />; // Render loading state if grainData is not available
  }

  const data = [
    ['Manufacturer', grainData.manufacturer ?? 'N/A'],
    ['CPU Model', grainData.cpu_model ?? 'N/A'],
    ['CPUs', grainData.num_cpus ?? 'N/A'],
    ...((grainData.gpus ?? []).map((gpu, index) => [`GPU ${index + 1} Model`, gpu.model]) ?? []),
    ['GPUs', grainData.num_gpus ?? 'N/A'],
    ['Memory', grainData.mem_total ?? 'N/A'],
    ['Swap', grainData.swap_total ?? 'N/A'],
  ];

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Hardware
      </Typography>
      <Divider />
      <TableContainer component={Paper}>
        <Table size="small">
          <TableBody>
            {data.map(([label, value]) => (
              <TableRow key={label}>
                <TableCell align="left">{label}</TableCell>
                <TableCell align="right">{value}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
};

export default HardwareDetails;
