import axios from 'axios';
import yaml from 'js-yaml';
import { useParams, useNavigate } from 'react-router-dom';
import React, { useMemo, useState, useEffect } from 'react';

import Box from '@mui/material/Box';
import Fab from '@mui/material/Fab';
import Tab from '@mui/material/Tab';
import Grid from '@mui/material/Grid';
import Tabs from '@mui/material/Tabs';
import Paper from '@mui/material/Paper';
import Tooltip from '@mui/material/Tooltip';
import RedoIcon from '@mui/icons-material/Redo';
import StartIcon from '@mui/icons-material/Start';
import RefreshIcon from '@mui/icons-material/Refresh';
import CircularProgress from '@mui/material/CircularProgress';

import useCachePaginated from 'src/hooks/saltCache/useCachePaginated.ts';
import useSaltCacheBankKey from 'src/hooks/saltCache/useSaltCacheBankKey.ts';

import CodeView from './CodeView.tsx';
import DataViewer from './DataViewer.tsx';
import CommonDetails from './CommonDetails.tsx';
import NetworkDetails from './NetworkDetails.tsx';
import HardwareDetails from './HardwareDetails.tsx';
import ConformityDetailsView from './ConformityDetailsView.tsx';

const MinionDetailsView: React.FC = () => {
  const navigate = useNavigate();
  const { id: minionID } = useParams<{ id: string }>();
  const [tabValue, setTabValue] = useState(0);
  const [tabs, setTabs] = useState<{ label: string; key: string }[]>([]);
  const [refreshing, setRefreshing] = useState<boolean>(false);
  const [highstateRunning, setHighstateRunning] = useState<boolean>(false);

  const queryParams = useMemo(
    () => ({
      bank: `minions/${minionID}` || '',
    }),
    [minionID]
  );

  const { caches, isLoading: isLoadingCaches } = useCachePaginated(queryParams);
  const { cacheData, isLoading: isLoadingCacheData } = useSaltCacheBankKey(
    `minions/${minionID}`,
    'data'
  );

  useEffect(() => {
    const uniqueKeys = Array.from(new Set(caches.map((cache) => cache.psql_key)));

    // Filter and map Grains and Pillar
    const grainsAndPillarTabs = uniqueKeys
      .filter((key) => key === 'data')
      .flatMap(() => [
        { label: 'Grains', key: 'grains' },
        { label: 'Pillar', key: 'pillar' },
      ]);

    // Filter out HighState/Conformity and other keys
    const otherTabs = uniqueKeys
      .filter((key) => key !== 'data' && key !== 'conformity')
      .sort()
      .map((key) => ({ label: key, key }));

    // Create the final tabs array with Grains and Pillar first, other keys next, and HighState last
    const newTabs = [
      ...grainsAndPillarTabs,
      ...otherTabs,
      { label: 'HighState', key: 'conformity' },
    ];

    setTabs(newTabs);
  }, [caches]);

  if (isLoadingCaches || isLoadingCacheData) {
    return <CircularProgress color="success" />; // Render loading state if grainData is not available
  }

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const renderTabContent = (key: string) => {
    if (key === 'grains' || key === 'pillar') {
      const yamlData = yaml.dump(cacheData[key]);
      return <DataViewer data={yamlData} />;
    }
    if (key === 'conformity') {
      return <ConformityDetailsView />; // Render ConformityDetailsView for the new tab
    }
    return <CodeView bank={`minions/${minionID}`} cacheKey={`${key}`} />;
  };

  const handleHighstate = async () => {
    setHighstateRunning(true);
    const authToken = localStorage.getItem('authToken');
    const authSaltString = localStorage.getItem('authSalt');

    const parsedAuthSalt = JSON.parse(authSaltString as string);
    const { token } = parsedAuthSalt;

    try {
      await axios.post(
        `/api/v1/netapi/`,
        {
          client: 'local',
          fun: 'state.apply',
          tgt: minionID,
          tgt_type: 'glob',
        },
        {
          headers: {
            Authorization: authToken,
            'X-Auth-Token': token,
          },
        }
      );
    } catch (err) {
      console.error(`Failed to run highstate on minion ${minionID}: `, err);
    } finally {
      setHighstateRunning(false);
    }
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    const authToken = localStorage.getItem('authToken');
    const authSaltString = localStorage.getItem('authSalt');

    const parsedAuthSalt = JSON.parse(authSaltString as string);
    const { token } = parsedAuthSalt;

    try {
      await axios.post(
        `/api/v1/netapi/`,
        [
          {
            client: 'local',
            fun: 'saltutil.refresh_grains',
            tgt: minionID,
            tgt_type: 'glob',
          },
          {
            client: 'local',
            fun: 'saltutil.refresh_pillar',
            tgt: minionID,
            tgt_type: 'glob',
          },
        ],
        {
          headers: {
            Authorization: authToken,
            'X-Auth-Token': token,
          },
        }
      );
    } catch (err) {
      console.error(`Failed to update cache for minion ${minionID}: `, err);
    } finally {
      setRefreshing(false);
    }
  };

  return (
    <Box sx={{ padding: 2 }}>
      <Grid container spacing={2}>
        <Grid item xs={12} md={3}>
          <>
            <Paper sx={{ padding: 2, marginBottom: 2 }}>
              <CommonDetails grainData={cacheData.grains as Record<string, unknown>} />
            </Paper>
            <Paper sx={{ padding: 2, marginBottom: 2 }}>
              <NetworkDetails grainData={cacheData.grains as Record<string, unknown>} />
            </Paper>
            <Paper sx={{ padding: 2 }}>
              <HardwareDetails grainData={cacheData.grains as Record<string, unknown>} />
            </Paper>
          </>
        </Grid>
        <Grid item xs={12} md={9}>
          <Paper sx={{ padding: 2, marginBottom: 2 }}>
            <Tabs value={tabValue} onChange={handleTabChange}>
              {tabs.map((tab) => (
                <Tab label={tab.label} key={tab.key} />
              ))}
            </Tabs>
          </Paper>
          {tabs.map(
            (tab, index) =>
              tabValue === index && (
                <Paper sx={{ padding: 2 }} key={tab.key}>
                  {renderTabContent(tab.key)}
                </Paper>
              )
          )}
          <Grid>
            <Fab
              style={{ float: 'right' }}
              color="success"
              sx={{ marginLeft: 1 }}
              onClick={handleRefresh}
              disabled={refreshing || highstateRunning}
            >
              <Tooltip title="Refresh minion details" arrow>
                <RefreshIcon />
              </Tooltip>
            </Fab>
            <Fab
              style={{ float: 'right' }}
              color="secondary"
              sx={{ marginLeft: 1 }}
              onClick={handleHighstate}
              disabled={refreshing || highstateRunning}
            >
              <Tooltip title="Run highstate on minion" arrow>
                <RedoIcon />
              </Tooltip>
            </Fab>
            <Fab
              style={{ float: 'right' }}
              color="primary"
              sx={{ marginLeft: 1 }}
              onClick={() => navigate(`/run?tgt=${minionID}`)}
            >
              <Tooltip title="Run command on minion" arrow>
                <StartIcon />
              </Tooltip>
            </Fab>
          </Grid>
        </Grid>
      </Grid>
    </Box>
  );
};

export default MinionDetailsView;
