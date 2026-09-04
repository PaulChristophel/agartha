import React, { useRef, useState, useEffect } from 'react';

import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Autocomplete from '@mui/material/Autocomplete';
import RefreshIcon from '@mui/icons-material/Refresh';

import useDebounce from 'src/hooks/useDebounce.ts';
import useRefreshKeys from 'src/hooks/saltMinion/useRefreshKeys.ts';
import useFetchGrainsKeys from 'src/hooks/saltMinion/useFetchGrainsKeys.ts';
import useFetchPillarKeys from 'src/hooks/saltMinion/useFetchPillarKeys.ts';

interface MinionsSearchProps {
  minionID: string;
  setMinionID: (value: string) => void;
  since: string;
  setSince: (value: string) => void;
  until: string;
  setUntil: (value: string) => void;
  grainKeys: string[];
  setGrainKeys: (value: string[]) => void;
  grainFilters: string[];
  setGrainFilters: React.Dispatch<React.SetStateAction<string[]>>;
  pillarKeys: string[];
  setPillarKeys: (value: string[]) => void;
  pillarFilters: string[];
  setPillarFilters: React.Dispatch<React.SetStateAction<string[]>>;
  setOrderBy: (orderBy: string) => void;
}

const MinionsSearch: React.FC<MinionsSearchProps> = ({
  minionID,
  setMinionID,
  since,
  setSince,
  until,
  setUntil,
  grainKeys,
  setGrainKeys,
  grainFilters,
  setGrainFilters,
  pillarKeys,
  setPillarKeys,
  pillarFilters,
  setPillarFilters,
  setOrderBy,
}) => {
  const [grainInputValue, setGrainInputValue] = useState('');
  const [grainPage, setGrainPage] = useState(1);
  const [allGrainsKeys, setAllGrainsKeys] = useState<string[]>([]);
  const [hasMoreGrains, setHasMoreGrains] = useState(true);
  const [pillarInputValue, setPillarInputValue] = useState('');
  const [pillarPage, setPillarPage] = useState(1);
  const [allPillarKeys, setAllPillarKeys] = useState<string[]>([]);
  const [hasMorePillars, setHasMorePillars] = useState(true);
  const [grainFilterPath, setGrainFilterPath] = useState('');
  const [grainFilterValue, setGrainFilterValue] = useState('');
  const [grainFilterType, setGrainFilterType] = useState('string');
  const [grainFilterOperator, setGrainFilterOperator] = useState<
    'eq' | 'not' | 'like' | 'not_like'
  >('eq');
  const [pillarFilterPath, setPillarFilterPath] = useState('');
  const [pillarFilterValue, setPillarFilterValue] = useState('');
  const [pillarFilterType, setPillarFilterType] = useState('string');
  const [pillarFilterOperator, setPillarFilterOperator] = useState<
    'eq' | 'not' | 'like' | 'not_like'
  >('eq');
  const debouncedGrainInputValue = useDebounce(grainInputValue, 500);
  const debouncedPillarInputValue = useDebounce(pillarInputValue, 500);
  const autoRefreshAttempted = useRef(false);
  const { isRefreshing, error: refreshError, message, revision, refreshKeys } = useRefreshKeys();

  const {
    grainsKeys,
    loading: grainsLoading,
    error: grainsError,
    isEmpty: grainsEmpty,
  } = useFetchGrainsKeys(debouncedGrainInputValue, grainPage, revision);
  const {
    pillarKeys: fetchedPillarKeys,
    loading: pillarsLoading,
    error: pillarsError,
    isEmpty: pillarsEmpty,
  } = useFetchPillarKeys(debouncedPillarInputValue, pillarPage, revision);

  useEffect(() => {
    const initialListsLoaded =
      !grainsLoading &&
      !pillarsLoading &&
      grainPage === 1 &&
      pillarPage === 1 &&
      debouncedGrainInputValue === '' &&
      debouncedPillarInputValue === '';

    if (initialListsLoaded && (grainsEmpty || pillarsEmpty) && !autoRefreshAttempted.current) {
      autoRefreshAttempted.current = true;
      void refreshKeys();
    }
  }, [
    debouncedGrainInputValue,
    debouncedPillarInputValue,
    grainPage,
    grainsEmpty,
    grainsLoading,
    pillarPage,
    pillarsEmpty,
    pillarsLoading,
    refreshKeys,
  ]);

  useEffect(() => {
    if (grainPage === 1) {
      setAllGrainsKeys(grainsKeys);
    } else {
      setAllGrainsKeys((prev) => [...prev, ...grainsKeys]);
    }
    setHasMoreGrains(grainsKeys.length > 0);
  }, [grainsKeys, grainPage]);

  useEffect(() => {
    if (pillarPage === 1) {
      setAllPillarKeys(fetchedPillarKeys);
    } else {
      setAllPillarKeys((prev) => [...prev, ...fetchedPillarKeys]);
    }
    setHasMorePillars(fetchedPillarKeys.length > 0);
  }, [fetchedPillarKeys, pillarPage]);

  useEffect(() => {
    if (
      (grainFilterOperator === 'like' || grainFilterOperator === 'not_like') &&
      grainFilterType !== 'string'
    ) {
      setGrainFilterType('string');
    }
  }, [grainFilterOperator, grainFilterType]);

  useEffect(() => {
    if (
      (pillarFilterOperator === 'like' || pillarFilterOperator === 'not_like') &&
      pillarFilterType !== 'string'
    ) {
      setPillarFilterType('string');
    }
  }, [pillarFilterOperator, pillarFilterType]);

  const handleMinionIDChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setMinionID(e.target.value);
  };

  const handleSinceChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSince(e.target.value);
  };

  const handleUntilChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setUntil(e.target.value);
  };

  const handleGrainKeyChange = (_event: React.ChangeEvent<unknown>, newValue: string[]) => {
    setGrainKeys(newValue);
    setOrderBy('');
  };

  const handlePillarKeyChange = (_event: React.ChangeEvent<unknown>, newValue: string[]) => {
    setPillarKeys(newValue);
    setOrderBy('');
  };

  const handleGrainsScroll = (event: React.SyntheticEvent) => {
    const listboxNode = event.currentTarget;
    if (listboxNode.scrollTop + listboxNode.clientHeight >= listboxNode.scrollHeight - 1) {
      if (!grainsLoading && hasMoreGrains) {
        setGrainPage((prevPage) => prevPage + 1);
      }
    }
  };

  const handlePillarsScroll = (event: React.SyntheticEvent) => {
    const listboxNode = event.currentTarget;
    if (listboxNode.scrollTop + listboxNode.clientHeight >= listboxNode.scrollHeight - 1) {
      if (!pillarsLoading && hasMorePillars) {
        setPillarPage((prevPage) => prevPage + 1);
      }
    }
  };

  const handleAddGrainFilter = () => {
    if (!grainFilterPath.trim() || !grainFilterValue.trim()) {
      return;
    }

    const normalizedPath = grainFilterPath.trim();
    const operatorSuffix = grainFilterOperator === 'eq' ? '' : `::${grainFilterOperator}`;
    const newFilter = `${normalizedPath}:${grainFilterValue.trim()}::${grainFilterType}${operatorSuffix}`;
    setGrainFilters((prev) => {
      if (prev.includes(newFilter)) {
        return prev;
      }
      return [...prev, newFilter];
    });
    setGrainFilterValue('');
    setGrainFilterPath('');
  };

  const handleRemoveFilter = (targetFilter: string) => {
    setGrainFilters((prev) => prev.filter((filter) => filter !== targetFilter));
  };

  const handleAddPillarFilter = () => {
    if (!pillarFilterPath.trim() || !pillarFilterValue.trim()) {
      return;
    }

    const normalizedPath = pillarFilterPath.trim();
    const operatorSuffix = pillarFilterOperator === 'eq' ? '' : `::${pillarFilterOperator}`;
    const newFilter = `${normalizedPath}:${pillarFilterValue.trim()}::${pillarFilterType}${operatorSuffix}`;
    setPillarFilters((prev) => {
      if (prev.includes(newFilter)) {
        return prev;
      }
      return [...prev, newFilter];
    });
    setPillarFilterValue('');
    setPillarFilterPath('');
  };

  const handleRemovePillarFilter = (targetFilter: string) => {
    setPillarFilters((prev) => prev.filter((filter) => filter !== targetFilter));
  };

  const grainValuePlaceholder =
    grainFilterOperator === 'like' || grainFilterOperator === 'not_like'
      ? 'Use % wildcards'
      : 'e.g. RedHat';
  const pillarValuePlaceholder =
    pillarFilterOperator === 'like' || pillarFilterOperator === 'not_like'
      ? 'Use % wildcards'
      : 'e.g. web';

  return (
    <Box
      display="flex"
      flexWrap="wrap"
      alignItems="center"
      padding={2}
      bgcolor="background.paper"
      borderRadius={1}
      boxShadow={1}
      mb={2}
      sx={{ columnGap: 2, rowGap: 2 }}
    >
      <TextField
        label="Minion ID"
        value={minionID}
        onChange={handleMinionIDChange}
        sx={{ flex: '1 1 200px' }}
      />
      <Autocomplete
        multiple
        freeSolo
        options={allGrainsKeys}
        loading={grainsLoading}
        value={grainKeys}
        inputValue={grainInputValue}
        sx={{ flex: '2 1 320px' }}
        onInputChange={(_event, newInputValue) => {
          setGrainInputValue(newInputValue);
          setAllGrainsKeys([]); // Reset the list of options
          setGrainPage(1); // Reset page when input changes
        }}
        onChange={handleGrainKeyChange}
        ListboxProps={{
          onScroll: handleGrainsScroll,
        }}
        renderTags={(value: string[], getTagProps) =>
          value.map((option: string, index: number) => (
            <Chip variant="outlined" label={option} {...getTagProps({ index })} key={option} />
          ))
        }
        renderInput={(params) => (
          <TextField
            {...params}
            label="SELECT Grain"
            InputLabelProps={{
              shrink: true,
            }}
            error={Boolean(grainsError) && !hasMoreGrains}
            helperText={Boolean(grainsError) && !hasMoreGrains ? 'Failed to load grains keys' : ''}
          />
        )}
      />
      <Autocomplete
        multiple
        freeSolo
        options={allPillarKeys}
        loading={pillarsLoading}
        value={pillarKeys}
        inputValue={pillarInputValue}
        sx={{ flex: '2 1 320px' }}
        onInputChange={(_event, newInputValue) => {
          setPillarInputValue(newInputValue);
          setAllPillarKeys([]);
          setPillarPage(1);
        }}
        onChange={handlePillarKeyChange}
        ListboxProps={{
          onScroll: handlePillarsScroll,
        }}
        renderTags={(value: string[], getTagProps) =>
          value.map((option: string, index: number) => (
            <Chip variant="outlined" label={option} {...getTagProps({ index })} key={option} />
          ))
        }
        renderInput={(params) => (
          <TextField
            {...params}
            label="SELECT Pillar"
            InputLabelProps={{
              shrink: true,
            }}
            error={Boolean(pillarsError) && !hasMorePillars}
            helperText={
              Boolean(pillarsError) && !hasMorePillars ? 'Failed to load pillar keys' : ''
            }
          />
        )}
      />
      <Box display="flex" flexDirection="column" sx={{ flex: '1 1 100%', minWidth: 320 }}>
        <Box display="flex" gap={1} mb={1}>
          <Autocomplete
            freeSolo
            options={allGrainsKeys}
            value={grainFilterPath}
            inputValue={grainFilterPath}
            onInputChange={(_event, newInputValue) => {
              setGrainFilterPath(newInputValue);
            }}
            onChange={(_event, newValue) => {
              setGrainFilterPath(newValue || '');
            }}
            sx={{ flex: 1 }}
            renderInput={(params) => (
              <TextField
                {...params}
                label="WHERE Grain"
                placeholder="e.g. system:os"
                InputLabelProps={{
                  shrink: true,
                }}
              />
            )}
          />
          <TextField
            select
            label="Operator"
            value={grainFilterOperator}
            onChange={(e) =>
              setGrainFilterOperator(e.target.value as 'eq' | 'not' | 'like' | 'not_like')
            }
            sx={{ width: 160 }}
            InputLabelProps={{
              shrink: true,
            }}
          >
            <MenuItem value="eq">Equals</MenuItem>
            <MenuItem value="not">Not Equals</MenuItem>
            <MenuItem value="like">Like</MenuItem>
            <MenuItem value="not_like">Not Like</MenuItem>
          </TextField>
          <TextField
            label="Value"
            value={grainFilterValue}
            onChange={(e) => setGrainFilterValue(e.target.value)}
            placeholder={grainValuePlaceholder}
            InputLabelProps={{
              shrink: true,
            }}
            sx={{ flex: 1 }}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault();
                handleAddGrainFilter();
              }
            }}
          />
          <TextField
            select
            label="Type"
            value={grainFilterType}
            onChange={(e) => setGrainFilterType(e.target.value)}
            sx={{ width: 120 }}
            InputLabelProps={{
              shrink: true,
            }}
          >
            <MenuItem value="string">string</MenuItem>
            <MenuItem value="int">int</MenuItem>
            <MenuItem value="float">float</MenuItem>
            <MenuItem value="bool">bool</MenuItem>
            <MenuItem value="array">array</MenuItem>
            <MenuItem value="null">null</MenuItem>
          </TextField>
          <Button variant="outlined" onClick={handleAddGrainFilter} sx={{ whiteSpace: 'nowrap' }}>
            Add
          </Button>
        </Box>
        <Box display="flex" flexWrap="wrap" gap={1}>
          {grainFilters.map((filter) => (
            <Chip key={filter} label={filter} onDelete={() => handleRemoveFilter(filter)} />
          ))}
        </Box>
      </Box>
      <Box display="flex" flexDirection="column" sx={{ flex: '1 1 100%', minWidth: 320 }}>
        <Box display="flex" gap={1} mb={1}>
          <Autocomplete
            freeSolo
            options={allPillarKeys}
            value={pillarFilterPath}
            inputValue={pillarFilterPath}
            onInputChange={(_event, newInputValue) => {
              setPillarFilterPath(newInputValue);
            }}
            onChange={(_event, newValue) => {
              setPillarFilterPath(newValue || '');
            }}
            sx={{ flex: 1 }}
            renderInput={(params) => (
              <TextField
                {...params}
                label="WHERE Pillar"
                placeholder="e.g. role"
                InputLabelProps={{
                  shrink: true,
                }}
              />
            )}
          />
          <TextField
            select
            label="Operator"
            value={pillarFilterOperator}
            onChange={(e) =>
              setPillarFilterOperator(e.target.value as 'eq' | 'not' | 'like' | 'not_like')
            }
            sx={{ width: 160 }}
            InputLabelProps={{
              shrink: true,
            }}
          >
            <MenuItem value="eq">Equals</MenuItem>
            <MenuItem value="not">Not Equals</MenuItem>
            <MenuItem value="like">Like</MenuItem>
            <MenuItem value="not_like">Not Like</MenuItem>
          </TextField>
          <TextField
            label="Value"
            value={pillarFilterValue}
            onChange={(e) => setPillarFilterValue(e.target.value)}
            placeholder={pillarValuePlaceholder}
            InputLabelProps={{
              shrink: true,
            }}
            sx={{ flex: 1 }}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault();
                handleAddPillarFilter();
              }
            }}
          />
          <TextField
            select
            label="Type"
            value={pillarFilterType}
            onChange={(e) => setPillarFilterType(e.target.value)}
            sx={{ width: 120 }}
            InputLabelProps={{
              shrink: true,
            }}
          >
            <MenuItem value="string">string</MenuItem>
            <MenuItem value="int">int</MenuItem>
            <MenuItem value="float">float</MenuItem>
            <MenuItem value="bool">bool</MenuItem>
            <MenuItem value="array">array</MenuItem>
            <MenuItem value="null">null</MenuItem>
          </TextField>
          <Button variant="outlined" onClick={handleAddPillarFilter} sx={{ whiteSpace: 'nowrap' }}>
            Add
          </Button>
        </Box>
        <Box display="flex" flexWrap="wrap" gap={1}>
          {pillarFilters.map((filter) => (
            <Chip key={filter} label={filter} onDelete={() => handleRemovePillarFilter(filter)} />
          ))}
        </Box>
      </Box>
      <TextField
        label="From"
        type="datetime-local"
        value={since}
        onChange={handleSinceChange}
        sx={{ flex: '1 1 200px' }}
        InputLabelProps={{
          shrink: true,
        }}
      />
      <TextField
        label="To"
        type="datetime-local"
        value={until}
        onChange={handleUntilChange}
        sx={{ flex: '1 1 200px' }}
        InputLabelProps={{
          shrink: true,
        }}
      />
      <Button
        variant="outlined"
        startIcon={<RefreshIcon />}
        disabled={isRefreshing}
        onClick={() => void refreshKeys()}
        sx={{ flex: '0 0 auto', whiteSpace: 'nowrap' }}
      >
        {isRefreshing ? 'Refreshing lists...' : 'Refresh dropdown lists'}
      </Button>
      {(message || refreshError) && (
        <Alert
          severity={refreshError ? 'error' : isRefreshing ? 'info' : 'success'}
          sx={{ flex: '1 1 100%' }}
        >
          {refreshError || message}
        </Alert>
      )}
    </Box>
  );
};

export default MinionsSearch;
