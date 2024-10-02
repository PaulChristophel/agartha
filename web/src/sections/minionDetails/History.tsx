import React from 'react';

import Box from '@mui/material/Box';
import Table from '@mui/material/Table';
import Button from '@mui/material/Button';
import TableRow from '@mui/material/TableRow';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import Typography from '@mui/material/Typography';
import TableContainer from '@mui/material/TableContainer';

const jobHistoryData = [
  {
    jid: '20240617165901007070',
    function: 'state.highstate',
    arguments: '',
    keywordArguments: '',
    user: '',
    status: 'Success',
    date: '17/06/2024, 11:59:01',
  },
  {
    jid: '20240617160449715755',
    function: 'event.fire',
    arguments: '',
    keywordArguments: '',
    user: '',
    status: 'Failed',
    date: '17/06/2024, 11:40:48',
  },
  // Continue with the rest of your job history data
];

const History: React.FC = () => (
  <Box>
    <Typography variant="h6" gutterBottom>
      Job History
    </Typography>
    <TableContainer>
      <Table size="small">
        <TableHead>
          <TableRow>
            <TableCell>Jid</TableCell>
            <TableCell>Function</TableCell>
            <TableCell>Arguments</TableCell>
            <TableCell>Keyword Arguments</TableCell>
            <TableCell>User</TableCell>
            <TableCell>Status</TableCell>
            <TableCell>Date</TableCell>
            <TableCell>Action</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {jobHistoryData.map((job, index) => (
            <TableRow key={index}>
              <TableCell>{job.jid}</TableCell>
              <TableCell>{job.function}</TableCell>
              <TableCell>{job.arguments}</TableCell>
              <TableCell>{job.keywordArguments}</TableCell>
              <TableCell>{job.user}</TableCell>
              <TableCell>{job.status}</TableCell>
              <TableCell>{job.date}</TableCell>
              <TableCell>
                <Button size="small" variant="contained" color="primary">
                  Detail
                </Button>
                <Button size="small" variant="contained" color="secondary" sx={{ marginLeft: 1 }}>
                  Rerun
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </TableContainer>
  </Box>
);

export default History;
