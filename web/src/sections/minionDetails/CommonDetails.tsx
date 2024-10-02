import React, { useMemo, useState, useEffect } from 'react';

import Box from '@mui/material/Box';
import Tab from '@mui/material/Tab';
import Chip from '@mui/material/Chip';
import Tabs from '@mui/material/Tabs';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableRow from '@mui/material/TableRow';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import Typography from '@mui/material/Typography';
import TableContainer from '@mui/material/TableContainer';
import CircularProgress from '@mui/material/CircularProgress';

import useHighState from 'src/hooks/highState/useHighState.ts';
import useReturnPaginated from 'src/hooks/saltReturn/useReturnPaginated.ts';

import formatTime from 'src/utils/formatTime.ts';

interface GrainData {
  fqdn?: string;
  os?: string;
  oscodename?: string;
  osrelease?: string;
  lastJob?: string;
  lastHighstate?: string;
  highstateConformity?: string;
  id?: string;
  master?: string;
  saltversion?: string;
  saltpath?: string;
  pythonversion?: string[];
  saltopts?: {
    master?: string;
    saltenv?: string;
    pillarenv?: string;
    master_port?: string;
    publish_port?: string;
  };
}

interface CommonDetailsProps {
  grainData: GrainData | null; // Allow for null initially
}

const CommonDetails: React.FC<CommonDetailsProps> = ({ grainData }) => {
  const [tabValue, setTabValue] = useState(0);
  const [tabs, setTabs] = useState<{ label: string; key: string }[]>([
    { label: 'Common', key: 'common' },
    { label: 'Salt', key: 'salt' },
  ]);

  useEffect(() => {
    if (grainData && grainData.saltopts) {
      setTabs((prevTabs) => {
        if (prevTabs.some((tab) => tab.key === 'saltopts')) {
          return prevTabs;
        }
        return [...prevTabs, { label: 'SaltOpts', key: 'saltopts' }];
      });
    }
  }, [grainData]);

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const {
    alterTime,
    success,
    isLoading: isHighStateLoading,
  } = useHighState(grainData?.id || '', false, false);

  const queryParams = useMemo(
    () => ({
      id: grainData?.id || '',
      load_return: true,
    }),
    [grainData?.id]
  );

  const { returns, isLoading, error } = useReturnPaginated(queryParams, 1, 1);

  if (isLoading || !grainData || isHighStateLoading) {
    return <CircularProgress color="success" />; // Render loading state if grainData is not available
  }

  if (error) {
    return <div>Error loading data</div>; // Handle error state
  }

  const renderTabContent = (key: string) => {
    const saltopts = grainData.saltopts ?? {};
    const data = (() => {
      const lastJobTime = formatTime(returns[0]?.alter_time);
      const lastHighstateTime = formatTime(alterTime);

      switch (key) {
        case 'common':
          return [
            ['FQDN', grainData.fqdn ?? 'N/A'],
            [
              'OS',
              `${grainData.os ?? 'N/A'} ${grainData.osrelease ?? 'N/A'} (${grainData.oscodename ?? 'N/A'})`,
            ],
            ...(lastJobTime !== 'Invalid Date'
              ? [
                  [
                    'Last Job',
                    <a href={`/return/${returns[0]?.jid}/${returns[0]?.id}/`}>{lastJobTime}</a>,
                  ],
                ]
              : []),
            ...(lastHighstateTime !== 'Invalid Date'
              ? [
                  [
                    'Last Highstate',
                    <a href={`/conformity/${grainData.id}/`}>{lastHighstateTime}</a>,
                  ],
                ]
              : []),
            [
              'Highstate Conformity',
              <Chip
                variant="outlined"
                label={success?.toString() || 'N/A'}
                color={success?.toString() === 'true' ? 'success' : 'error'}
              />,
            ],
          ];
        case 'salt':
          return [
            ['ID', grainData.id ?? 'N/A'],
            ['Master', grainData.master ?? 'N/A'],
            ['Version', grainData.saltversion ?? 'N/A'],
            ['Path', grainData.saltpath ?? 'N/A'],
            ['Python', grainData.pythonversion ? grainData.pythonversion.join('.') : 'N/A'],
          ];
        case 'saltopts':
          return [
            ['Master', saltopts.master ?? 'N/A'],
            ['Saltenv', saltopts.saltenv ?? 'N/A'],
            ['Pillarenv', saltopts.pillarenv ?? 'N/A'],
            ['Master Port', saltopts.master_port ?? 'N/A'],
            ['Publish Port', saltopts.publish_port ?? 'N/A'],
          ];
        default:
          return [];
      }
    })();

    return (
      <TableContainer component={Paper}>
        <Table size="small">
          <TableBody>
            {data.map(([label, value]) => (
              <TableRow key={label as string}>
                <TableCell align="left">{label}</TableCell>
                <TableCell align="right">{value}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
    );
  };

  return (
    <Box>
      <Typography variant="h6" gutterBottom>
        Common Details
      </Typography>
      <Paper sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs value={tabValue} onChange={handleTabChange}>
          {tabs.map((tab) => (
            <Tab label={tab.label} key={tab.key} />
          ))}
        </Tabs>
      </Paper>
      <Box sx={{ padding: 0 }}>{renderTabContent(tabs[tabValue].key)}</Box>
    </Box>
  );
};

export default CommonDetails;
