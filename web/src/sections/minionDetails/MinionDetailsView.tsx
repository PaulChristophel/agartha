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

import useSaltMinionID from 'src/hooks/saltMinion/useSaltMinionID.ts';
import useCachePaginated from 'src/hooks/saltCache/useCachePaginated.ts';

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
  const [tabs, setTabs] = useState<{ label: string; key: string; bank?: string }[]>([]);
  const [refreshing, setRefreshing] = useState<boolean>(false);
  const [highstateRunning, setHighstateRunning] = useState<boolean>(false);

  const oldBankQueryParams = useMemo(
    () => ({
      bank: minionID ? `minions/${minionID}` : '',
    }),
    [minionID]
  );
  const newCacheQueryParams = useMemo(
    () => ({
      key: minionID ? `${minionID}*` : '',
    }),
    [minionID]
  );

  const { caches: oldBankCaches, isLoading: isLoadingOldBankCaches } =
    useCachePaginated(oldBankQueryParams);
  const { caches: newCacheCaches, isLoading: isLoadingNewCacheCaches } =
    useCachePaginated(newCacheQueryParams);
  const { grains, pillar, isLoading: isLoadingMinionData } = useSaltMinionID(minionID || '');

  const minionData = useMemo(() => ({ grains, pillar }), [grains, pillar]);

  useEffect(() => {
    const otherTabsByKey = new Map<string, { label: string; key: string; bank: string }>();
    const builtInKeys = new Set(['data', 'grains', 'pillar', 'conformity']);
    const isBuiltInCache = (cache: { bank: string; psql_key: string }) =>
      builtInKeys.has(cache.psql_key) ||
      Boolean(
        minionID &&
        (cache.bank === 'grains' || cache.bank === 'pillar') &&
        cache.psql_key === minionID
      );
    const tabLabel = (cache: { bank: string; psql_key: string }) => {
      const pillarEnvPrefix = `${minionID}:`;

      if (cache.bank === 'pillar' && minionID && cache.psql_key.startsWith(pillarEnvPrefix)) {
        return `Pillar:${cache.psql_key.slice(pillarEnvPrefix.length)}`;
      }

      return cache.psql_key;
    };

    for (const cache of oldBankCaches) {
      if (isBuiltInCache(cache)) {
        continue;
      }
      otherTabsByKey.set(cache.psql_key, {
        label: tabLabel(cache),
        key: cache.psql_key,
        bank: cache.bank,
      });
    }

    for (const cache of newCacheCaches) {
      if (isBuiltInCache(cache)) {
        continue;
      }
      otherTabsByKey.set(cache.psql_key, {
        label: tabLabel(cache),
        key: cache.psql_key,
        bank: cache.bank,
      });
    }

    // Grains and pillar come from /salt_minion so both old and Salt 3008 cache layouts work.
    const otherTabs = Array.from(otherTabsByKey.values()).sort((a, b) =>
      a.label.localeCompare(b.label)
    );

    // Create the final tabs array with Grains and Pillar first, other keys next, and HighState last
    const newTabs = [
      { label: 'Grains', key: 'grains' },
      { label: 'Pillar', key: 'pillar' },
      ...otherTabs,
      { label: 'HighState', key: 'conformity' },
    ];

    setTabs(newTabs);
  }, [minionID, newCacheCaches, oldBankCaches]);

  if (isLoadingOldBankCaches || isLoadingNewCacheCaches || isLoadingMinionData) {
    return <CircularProgress color="success" />; // Render loading state if grainData is not available
  }

  const handleTabChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const renderTabContent = (tab: { key: string; bank?: string }) => {
    const { key } = tab;
    if (key === 'grains' || key === 'pillar') {
      const yamlData = yaml.dump(minionData[key]);
      return <DataViewer data={yamlData} />;
    }
    if (key === 'conformity') {
      return <ConformityDetailsView />; // Render ConformityDetailsView for the new tab
    }
    return <CodeView bank={tab.bank || `minions/${minionID}`} cacheKey={`${key}`} />;
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
              <CommonDetails grainData={grains} />
            </Paper>
            <Paper sx={{ padding: 2, marginBottom: 2 }}>
              <NetworkDetails grainData={grains} />
            </Paper>
            <Paper sx={{ padding: 2 }}>
              <HardwareDetails grainData={grains} />
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
                  {renderTabContent(tab)}
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
