import React, { useState, useEffect } from 'react';

import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Autocomplete from '@mui/material/Autocomplete';

import useDebounce from 'src/hooks/useDebounce.ts';
import useFetchGrainsKeys from 'src/hooks/saltMinion/useFetchGrainsKeys.ts';

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
  setOrderBy: (orderBy: string) => void;

  // grainValue: string;
  // setGrainValue: (value: string) => void;
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
  setOrderBy,
  // grainValue,
  // setGrainValue,
}) => {
  const authToken = localStorage.getItem('authToken') || '';
  const [inputValue, setInputValue] = useState('');
  const [page, setPage] = useState(1);
  const [allGrainsKeys, setAllGrainsKeys] = useState<string[]>([]);
  const [hasMore, setHasMore] = useState(true);
  const [filterPath, setFilterPath] = useState('');
  const [filterValue, setFilterValue] = useState('');
  const [filterType, setFilterType] = useState('string');
  const [filterOperator, setFilterOperator] = useState<'eq' | 'not' | 'like' | 'not_like'>('eq');
  const debouncedInputValue = useDebounce(inputValue, 500);

  const { grainsKeys, loading, error } = useFetchGrainsKeys(authToken, debouncedInputValue, page);

  useEffect(() => {
    if (page === 1) {
      setAllGrainsKeys(grainsKeys);
    } else {
      setAllGrainsKeys((prev) => [...prev, ...grainsKeys]);
    }
    setHasMore(grainsKeys.length > 0); // Assuming that if the current fetch returned no results, we have no more data to load
  }, [grainsKeys, page]);

  useEffect(() => {
    if ((filterOperator === 'like' || filterOperator === 'not_like') && filterType !== 'string') {
      setFilterType('string');
    }
  }, [filterOperator, filterType]);

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

  const handleScroll = (event: React.SyntheticEvent) => {
    const listboxNode = event.currentTarget;
    if (listboxNode.scrollTop + listboxNode.clientHeight >= listboxNode.scrollHeight - 1) {
      if (!loading && hasMore) {
        setPage((prevPage) => prevPage + 1);
      }
    }
  };

  const handleAddFilter = () => {
    if (!filterPath.trim() || !filterValue.trim()) {
      return;
    }

    const normalizedPath = filterPath.trim();
    const operatorSuffix = filterOperator === 'eq' ? '' : `::${filterOperator}`;
    const newFilter = `${normalizedPath}:${filterValue.trim()}::${filterType}${operatorSuffix}`;
    setGrainFilters((prev) => {
      if (prev.includes(newFilter)) {
        return prev;
      }
      return [...prev, newFilter];
    });
    setFilterValue('');
    setFilterPath('');
  };

  const handleRemoveFilter = (targetFilter: string) => {
    setGrainFilters((prev) => prev.filter((filter) => filter !== targetFilter));
  };

  const valuePlaceholder =
    filterOperator === 'like' || filterOperator === 'not_like' ? 'Use % wildcards' : 'e.g. RedHat';

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
        loading={loading}
        value={grainKeys}
        inputValue={inputValue}
        sx={{ flex: '2 1 320px' }}
        onInputChange={(_event, newInputValue) => {
          setInputValue(newInputValue);
          setAllGrainsKeys([]); // Reset the list of options
          setPage(1); // Reset page when input changes
        }}
        onChange={handleGrainKeyChange}
        ListboxProps={{
          onScroll: handleScroll,
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
            error={Boolean(error) && !hasMore}
            helperText={Boolean(error) && !hasMore ? 'Failed to load grains keys' : ''}
          />
        )}
      />
      <Box display="flex" flexDirection="column" sx={{ flex: '1 1 100%', minWidth: 320 }}>
        <Box display="flex" gap={1} mb={1}>
          <Autocomplete
            freeSolo
            options={allGrainsKeys}
            value={filterPath}
            inputValue={filterPath}
            onInputChange={(_event, newInputValue) => {
              setFilterPath(newInputValue);
            }}
            onChange={(_event, newValue) => {
              setFilterPath(newValue || '');
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
            value={filterOperator}
            onChange={(e) =>
              setFilterOperator(e.target.value as 'eq' | 'not' | 'like' | 'not_like')
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
            value={filterValue}
            onChange={(e) => setFilterValue(e.target.value)}
            placeholder={valuePlaceholder}
            InputLabelProps={{
              shrink: true,
            }}
            sx={{ flex: 1 }}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault();
                handleAddFilter();
              }
            }}
          />
          <TextField
            select
            label="Type"
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
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
          <Button variant="outlined" onClick={handleAddFilter} sx={{ whiteSpace: 'nowrap' }}>
            Add
          </Button>
        </Box>
        <Box display="flex" flexWrap="wrap" gap={1}>
          {grainFilters.map((filter) => (
            <Chip key={filter} label={filter} onDelete={() => handleRemoveFilter(filter)} />
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
    </Box>
  );
};

export default MinionsSearch;
