import React, { useState, useEffect } from 'react';

import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Autocomplete from '@mui/material/Autocomplete';

import useDebounce from 'src/hooks/useDebounce.ts';
import useFetchFunKeys from 'src/hooks/saltReturn/useFetchFunKeys.ts';

interface ReturnsSearchProps {
  jidString: string;
  setJidString: (value: string) => void;
  minionID: string;
  setMinionID: (value: string) => void;
  fun: string;
  setFun: (value: string) => void;
  success: string;
  setSuccess: (value: string) => void;
  since: string;
  setSince: (value: string) => void;
  until: string;
  setUntil: (value: string) => void;
}

const ReturnsSearch: React.FC<ReturnsSearchProps> = ({
  jidString,
  setJidString,
  minionID,
  setMinionID,
  fun,
  setFun,
  success,
  setSuccess,
  since,
  setSince,
  until,
  setUntil,
}) => {
  const authToken = localStorage.getItem('authToken') || '';
  const [inputValue, setInputValue] = useState('');
  const [page, setPage] = useState(1);
  const [allFunKeys, setAllFunKeys] = useState<string[]>([]);
  const [hasMore, setHasMore] = useState(true);
  const debouncedInputValue = useDebounce(inputValue, 500);

  const { funKeys, loading, error } = useFetchFunKeys(
    authToken,
    debouncedInputValue,
    page,
    since,
    until
  );

  useEffect(() => {
    if (page === 1) {
      setAllFunKeys(funKeys);
    } else {
      setAllFunKeys((prev) => [...prev, ...funKeys]);
    }
    setHasMore(funKeys.length > 0); // Assuming that if the current fetch returned no results, we have no more data to load
  }, [funKeys, page]);

  const handleFunChange = (_event: React.ChangeEvent<unknown>, newValue: string[]) => {
    setFun(newValue.join(','));
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
        label="Job ID"
        value={jidString}
        onChange={(e) => {
          const re = /^[0-9\b]+$/;
          if (re.test(e.target.value) || e.target.value === '') {
            setJidString(e.target.value);
          }
        }}
        sx={{ marginRight: 2, width: '20%' }}
      />
      <TextField
        label="Minion ID"
        value={minionID}
        onChange={(e) => setMinionID(e.target.value)}
        sx={{ marginRight: 2, width: '20%' }}
      />
      <Autocomplete
        multiple
        freeSolo
        options={allFunKeys}
        loading={loading}
        value={fun ? fun.split(',') : []}
        inputValue={inputValue}
        sx={{ marginRight: 2, width: '50%' }}
        onInputChange={(_event, newInputValue) => {
          setInputValue(newInputValue);
          setAllFunKeys([]); // Reset the list of options
          setPage(1); // Reset page when input changes
        }}
        onChange={handleFunChange}
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
            label="SELECT Function"
            InputLabelProps={{
              shrink: true,
            }}
            sx={{ marginRight: 2, width: '100%' }}
            error={Boolean(error) && !hasMore}
            helperText={Boolean(error) && !hasMore ? 'Failed to load function keys' : ''}
          />
        )}
      />
      <TextField
        select
        label="Success"
        value={success}
        onChange={(e) => setSuccess(e.target.value)}
        sx={{ marginRight: 2, width: '20%' }}
      >
        <MenuItem value="">all</MenuItem>
        <MenuItem value="true">true</MenuItem>
        <MenuItem value="false">false</MenuItem>
      </TextField>
      <TextField
        label="From"
        type="datetime-local"
        value={since}
        onChange={(e) => setSince(e.target.value)}
        sx={{ marginRight: 2, width: '20%' }}
        InputLabelProps={{
          shrink: true,
        }}
      />
      <TextField
        label="To"
        type="datetime-local"
        value={until}
        onChange={(e) => setUntil(e.target.value)}
        sx={{ width: '20%' }}
        InputLabelProps={{
          shrink: true,
        }}
      />
    </Box>
  );
};

export default ReturnsSearch;
