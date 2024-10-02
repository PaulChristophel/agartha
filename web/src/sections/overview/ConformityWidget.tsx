import React from 'react';
import { Pie, Cell, Legend, Tooltip, PieChart, ResponsiveContainer } from 'recharts';

import { Theme, SxProps } from '@mui/material/styles';
import {
  Card,
  Table,
  Paper,
  TableRow,
  TableBody,
  TableCell,
  Typography,
  TableContainer,
  CircularProgress,
} from '@mui/material';

import useConformityPaginated from 'src/hooks/conformity/useConformityPaginated.ts';

interface ConformityWidgetProps {
  sx?: SxProps<Theme>;
  [key: string]: unknown; // To allow other props
}

const COLORS = ['#0088FE', '#FF8042'];

const ConformityWidget: React.FC<ConformityWidgetProps> = ({ sx, ...other }) => {
  const trueQueryParams = React.useMemo(
    () => ({
      success: true,
      limit: 1,
      page: 1,
    }),
    []
  );

  const falseQueryParams = React.useMemo(
    () => ({
      success: false,
      limit: 1,
      page: 1,
    }),
    []
  );

  const { isLoading: isLoadingTrue, totalCount: totalTrueCount } = useConformityPaginated(
    trueQueryParams,
    1,
    1
  );

  const { isLoading: isLoadingFalse, totalCount: totalFalseCount } = useConformityPaginated(
    falseQueryParams,
    1,
    1
  );

  const total = totalTrueCount + totalFalseCount;
  const data = [
    { name: 'Conforming', value: totalTrueCount },
    { name: 'Non-Conforming', value: totalFalseCount },
  ];

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
        Conformity
      </Typography>
      <TableContainer component={Paper}>
        <Table size="small">
          <TableBody>
            <TableRow>
              <TableCell align="left">Conforming</TableCell>
              <TableCell align="right">
                {isLoadingTrue ? <CircularProgress size={20} /> : totalTrueCount}
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell align="left">Non-Conforming</TableCell>
              <TableCell align="right">
                {isLoadingFalse ? <CircularProgress size={20} /> : totalFalseCount}
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
      <ResponsiveContainer width="100%" height={200}>
        <PieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            labelLine={false}
            outerRadius={80}
            fill="#8884d8"
            dataKey="value"
          >
            {data.map((_entry, index) => (
              <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
            ))}
          </Pie>
          <Tooltip formatter={(value: number) => `${((value / total) * 100).toFixed(2)}%`} />
          <Legend />
        </PieChart>
      </ResponsiveContainer>
    </Card>
  );
};

export default ConformityWidget;
