import React, { useState, useEffect } from 'react';

import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
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
  setOrderBy,
  // grainValue,
  // setGrainValue,
}) => {
  const authToken = localStorage.getItem('authToken') || '';
  const [inputValue, setInputValue] = useState('');
  const [page, setPage] = useState(1);
  const [allGrainsKeys, setAllGrainsKeys] = useState<string[]>([]);
  const [hasMore, setHasMore] = useState(true);
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

  return (
    <Box
      display="flex"
      alignItems="center"
      padding={2}
      bgcolor="background.paper"
      borderRadius={1}
      boxShadow={1}
      mb={2}
    >
      <TextField
        label="Minion ID"
        value={minionID}
        onChange={handleMinionIDChange}
        sx={{ marginRight: 2, width: '15%' }}
      />
      <Autocomplete
        multiple
        freeSolo
        options={allGrainsKeys}
        loading={loading}
        value={grainKeys}
        inputValue={inputValue}
        sx={{ marginRight: 2, width: '50%' }}
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
            sx={{ marginRight: 2, width: '100%' }}
            error={Boolean(error) && !hasMore}
            helperText={Boolean(error) && !hasMore ? 'Failed to load grains keys' : ''}
          />
        )}
      />
      {/* <TextField
        label="WHERE Grain Value"
        value={grainValue}
        onChange={(e) => setGrainValue(e.target.value)}
        sx={{ marginRight: 2, width: '25%' }}
        InputLabelProps={{
          shrink: true,
        }}
      /> */}
      <TextField
        label="From"
        type="datetime-local"
        value={since}
        onChange={handleSinceChange}
        sx={{ marginRight: 2, width: '15%' }}
        InputLabelProps={{
          shrink: true,
        }}
      />
      <TextField
        label="To"
        type="datetime-local"
        value={until}
        onChange={handleUntilChange}
        sx={{ width: '15%' }}
        InputLabelProps={{
          shrink: true,
        }}
      />
    </Box>
  );
};

export default MinionsSearch;
