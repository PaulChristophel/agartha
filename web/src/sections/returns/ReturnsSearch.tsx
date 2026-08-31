import React, { useState, useEffect } from 'react';

import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import Tooltip from '@mui/material/Tooltip';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import Autocomplete from '@mui/material/Autocomplete';

import useDebounce from 'src/hooks/useDebounce.ts';
import useFetchFunKeys from 'src/hooks/saltReturn/useFetchFunKeys.ts';

import {
  parseReturnPath,
  emptyReturnFilter,
  type ReturnValueType,
  newReturnFilterClause,
  type ReturnFilterState,
  type ReturnFilterScope,
  type ReturnFilterClause,
  returnOperatorNeedsValue,
  type ReturnFilterOperator,
} from './returnFilter.ts';

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
  returnFilter: ReturnFilterState;
  setReturnFilter: (value: ReturnFilterState) => void;
}

const editableReturnFilter = (filter: ReturnFilterState): ReturnFilterState =>
  filter.clauses.length > 0 ? filter : { logic: 'and', clauses: [newReturnFilterClause()] };

const numericOperators: ReturnFilterOperator[] = ['gt', 'gte', 'lt', 'lte'];
const stringOperators: ReturnFilterOperator[] = ['contains', 'icontains', 'regex'];
const hoverHelpProps = {
  arrow: true,
  placement: 'top' as const,
  enterDelay: 700,
  enterNextDelay: 700,
  leaveDelay: 0,
  disableFocusListener: true,
  disableInteractive: true,
};

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
  returnFilter,
  setReturnFilter,
}) => {
  const [inputValue, setInputValue] = useState('');
  const [page, setPage] = useState(1);
  const [allFunKeys, setAllFunKeys] = useState<string[]>([]);
  const [hasMore, setHasMore] = useState(true);
  const [draftReturnFilter, setDraftReturnFilter] = useState<ReturnFilterState>(() =>
    editableReturnFilter(returnFilter)
  );
  const debouncedInputValue = useDebounce(inputValue, 500);

  const { funKeys, loading, error } = useFetchFunKeys(debouncedInputValue, page, since, until);

  useEffect(() => {
    if (page === 1) {
      setAllFunKeys(funKeys);
    } else {
      setAllFunKeys((prev) => [...prev, ...funKeys]);
    }
    setHasMore(funKeys.length > 0);
  }, [funKeys, page]);

  useEffect(() => {
    setDraftReturnFilter(editableReturnFilter(returnFilter));
  }, [returnFilter]);

  const handleFunChange = (_event: React.ChangeEvent<unknown>, newValue: string[]) => {
    setFun(newValue.join(','));
  };

  const handleScroll = (event: React.SyntheticEvent) => {
    const listboxNode = event.currentTarget;
    if (listboxNode.scrollTop + listboxNode.clientHeight >= listboxNode.scrollHeight - 1) {
      if (!loading && hasMore) setPage((prevPage) => prevPage + 1);
    }
  };

  const updateReturnClause = (index: number, changes: Partial<ReturnFilterClause>) => {
    setDraftReturnFilter((current) => ({
      ...current,
      clauses: current.clauses.map((clause, clauseIndex) =>
        clauseIndex === index ? { ...clause, ...changes } : clause
      ),
    }));
  };

  const changeReturnOperator = (index: number, operator: ReturnFilterOperator) => {
    const current = draftReturnFilter.clauses[index];
    let { valueType } = current;
    if (numericOperators.includes(operator) && valueType !== 'int' && valueType !== 'float') {
      valueType = 'float';
    } else if (stringOperators.includes(operator)) {
      valueType = 'string';
    }
    updateReturnClause(index, { operator, valueType });
  };

  const clauseIsValid = (clause: ReturnFilterClause): boolean => {
    if (clause.scope === 'key' && !clause.key) return false;
    const path = parseReturnPath(clause.path);
    if (path.length > 16 || path.some((component) => !component)) return false;
    if (!returnOperatorNeedsValue(clause.operator)) return true;
    if (clause.valueType !== 'null' && !clause.value) return false;
    if (numericOperators.includes(clause.operator)) {
      return clause.valueType === 'int' || clause.valueType === 'float';
    }
    if (stringOperators.includes(clause.operator)) return clause.valueType === 'string';
    if (clause.valueType === 'bool') return clause.value === 'true' || clause.value === 'false';
    if (clause.valueType === 'int') return /^[-+]?\d+$/.test(clause.value);
    if (clause.valueType === 'float') return Number.isFinite(Number(clause.value));
    return true;
  };

  const canApplyReturnFilter =
    draftReturnFilter.clauses.length > 0 && draftReturnFilter.clauses.every(clauseIsValid);

  const clearReturnFilter = () => {
    const emptyFilter = emptyReturnFilter();
    setDraftReturnFilter(editableReturnFilter(emptyFilter));
    setReturnFilter(emptyFilter);
  };

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
      <Tooltip {...hoverHelpProps} title="Filter by the return row's job ID.">
        <TextField
          label="Job ID"
          value={jidString}
          onChange={(event) => {
            const jobID = event.target.value;
            if (/^[0-9\b]+$/.test(jobID) || jobID === '') setJidString(jobID);
          }}
          sx={{ flex: '1 1 180px' }}
        />
      </Tooltip>
      <Tooltip
        {...hoverHelpProps}
        title="Filter by minion ID. Partial text is matched as a prefix."
      >
        <TextField
          label="Minion ID"
          value={minionID}
          onChange={(event) => setMinionID(event.target.value)}
          sx={{ flex: '1 1 220px' }}
        />
      </Tooltip>
      <Tooltip
        {...hoverHelpProps}
        title="Select one or more Salt functions. Multiple functions are combined with OR."
      >
        <Autocomplete
          multiple
          freeSolo
          options={allFunKeys}
          loading={loading}
          value={fun ? fun.split(',') : []}
          inputValue={inputValue}
          sx={{ flex: '2 1 320px' }}
          onInputChange={(_event, newInputValue) => {
            setInputValue(newInputValue);
            setAllFunKeys([]);
            setPage(1);
          }}
          onChange={handleFunChange}
          ListboxProps={{ onScroll: handleScroll }}
          renderTags={(value: string[], getTagProps) =>
            value.map((option: string, index: number) => (
              <Chip variant="outlined" label={option} {...getTagProps({ index })} key={option} />
            ))
          }
          renderInput={(params) => (
            <TextField
              {...params}
              label="SELECT Function"
              InputLabelProps={{ shrink: true }}
              sx={{ width: '100%' }}
              error={Boolean(error) && !hasMore}
              helperText={Boolean(error) && !hasMore ? 'Failed to load function keys' : ''}
            />
          )}
        />
      </Tooltip>
      <Tooltip {...hoverHelpProps} title="Filter by the success column: true, false, or all rows.">
        <TextField
          select
          label="Success"
          value={success}
          onChange={(event) => setSuccess(event.target.value)}
          sx={{ flex: '1 1 140px' }}
        >
          <MenuItem value="">all</MenuItem>
          <MenuItem value="true">true</MenuItem>
          <MenuItem value="false">false</MenuItem>
        </TextField>
      </Tooltip>
      <Tooltip
        {...hoverHelpProps}
        title="Only include returns recorded at or after this date and time."
      >
        <TextField
          label="From"
          type="datetime-local"
          value={since}
          onChange={(event) => setSince(event.target.value)}
          sx={{ flex: '1 1 220px' }}
          InputLabelProps={{ shrink: true }}
        />
      </Tooltip>
      <Tooltip
        {...hoverHelpProps}
        title="Only include returns recorded at or before this date and time."
      >
        <TextField
          label="To"
          type="datetime-local"
          value={until}
          onChange={(event) => setUntil(event.target.value)}
          sx={{ flex: '1 1 220px' }}
          InputLabelProps={{ shrink: true }}
        />
      </Tooltip>

      <Box display="flex" flexDirection="column" gap={1.5} sx={{ flex: '1 1 100%' }}>
        <Box display="flex" alignItems="center" flexWrap="wrap" gap={1}>
          <Typography variant="subtitle2">Return conditions</Typography>
          {draftReturnFilter.clauses.length > 1 && (
            <Tooltip
              {...hoverHelpProps}
              title="AND requires every condition to match. OR includes a return when any condition matches."
            >
              <TextField
                select
                size="small"
                label="Combine with"
                value={draftReturnFilter.logic}
                onChange={(event) =>
                  setDraftReturnFilter((current) => ({
                    ...current,
                    logic: event.target.value as 'and' | 'or',
                  }))
                }
                sx={{ width: 140 }}
              >
                <MenuItem value="and">AND</MenuItem>
                <MenuItem value="or">OR</MenuItem>
              </TextField>
            </Tooltip>
          )}
          <Button
            size="small"
            disabled={draftReturnFilter.clauses.length >= 20}
            onClick={() =>
              setDraftReturnFilter((current) => ({
                ...current,
                clauses: [...current.clauses, newReturnFilterClause()],
              }))
            }
          >
            Add condition
          </Button>
        </Box>

        {draftReturnFilter.clauses.map((clause, index) => (
          <Box
            key={`return-clause-${index}`}
            display="flex"
            flexWrap="wrap"
            alignItems="center"
            gap={1}
            padding={1}
            border={1}
            borderColor="divider"
            borderRadius={1}
          >
            <Tooltip
              {...hoverHelpProps}
              title="Choose the whole return, one exact top-level object key, or every top-level object entry."
            >
              <TextField
                select
                size="small"
                label="Scope"
                value={clause.scope}
                onChange={(event) =>
                  updateReturnClause(index, { scope: event.target.value as ReturnFilterScope })
                }
                sx={{ flex: '1 1 170px' }}
              >
                <MenuItem value="key">Exact object key</MenuItem>
                <MenuItem value="any_key">Any object entry</MenuItem>
                <MenuItem value="root">Root return</MenuItem>
              </TextField>
            </Tooltip>
            {clause.scope === 'key' && (
              <Tooltip
                {...hoverHelpProps}
                title="Enter the exact top-level JSON object key. Dots, pipes, and hyphens do not need escaping here."
              >
                <TextField
                  size="small"
                  label="Object key"
                  value={clause.key}
                  onChange={(event) => updateReturnClause(index, { key: event.target.value })}
                  placeholder="Exact top-level key in the return object"
                  sx={{ flex: '3 1 420px' }}
                />
              </Tooltip>
            )}
            <Tooltip
              {...hoverHelpProps}
              title={
                'Enter a dot-separated nested path, such as fun_args.0 or changes.diff. Numeric components select array indexes. Escape a literal dot as \\.'
              }
            >
              <TextField
                size="small"
                label="Field path"
                value={clause.path}
                onChange={(event) => updateReturnClause(index, { path: event.target.value })}
                placeholder={
                  clause.scope === 'root' ? 'Leave blank for the root value' : 'duration'
                }
                sx={{ flex: '1 1 200px' }}
              />
            </Tooltip>
            <Tooltip
              {...hoverHelpProps}
              title="Choose how the selected JSON value is compared. Numeric and string comparisons enforce compatible value types."
            >
              <TextField
                select
                size="small"
                label="Comparison"
                value={clause.operator}
                onChange={(event) =>
                  changeReturnOperator(index, event.target.value as ReturnFilterOperator)
                }
                sx={{ flex: '0 1 170px' }}
              >
                <MenuItem value="eq">equals</MenuItem>
                <MenuItem value="ne">does not equal</MenuItem>
                <MenuItem value="gt">greater than</MenuItem>
                <MenuItem value="gte">greater or equal</MenuItem>
                <MenuItem value="lt">less than</MenuItem>
                <MenuItem value="lte">less or equal</MenuItem>
                <MenuItem value="contains">contains</MenuItem>
                <MenuItem value="icontains">contains (ignore case)</MenuItem>
                <MenuItem value="regex">regular expression</MenuItem>
                <MenuItem value="exists">exists</MenuItem>
                <MenuItem value="not_exists">does not exist</MenuItem>
              </TextField>
            </Tooltip>
            {returnOperatorNeedsValue(clause.operator) && (
              <>
                <Tooltip
                  {...hoverHelpProps}
                  title="Enter the comparison value. It is interpreted using the selected JSON type."
                >
                  <Box component="span" sx={{ flex: '1 1 160px' }}>
                    <TextField
                      fullWidth
                      size="small"
                      label="Value"
                      value={clause.valueType === 'null' ? '' : clause.value}
                      disabled={clause.valueType === 'null'}
                      onChange={(event) => updateReturnClause(index, { value: event.target.value })}
                    />
                  </Box>
                </Tooltip>
                <Tooltip
                  {...hoverHelpProps}
                  title="Select the JSON type used to parse the comparison value. JIDs and other numeric-looking identifiers should usually remain strings."
                >
                  <TextField
                    select
                    size="small"
                    label="Type"
                    value={clause.valueType}
                    onChange={(event) =>
                      updateReturnClause(index, {
                        valueType: event.target.value as ReturnValueType,
                        value: event.target.value === 'null' ? '' : clause.value,
                      })
                    }
                    sx={{ flex: '0 1 120px' }}
                  >
                    <MenuItem value="string">string</MenuItem>
                    <MenuItem value="bool">bool</MenuItem>
                    <MenuItem value="int">int</MenuItem>
                    <MenuItem value="float">float</MenuItem>
                    <MenuItem value="null">null</MenuItem>
                  </TextField>
                </Tooltip>
              </>
            )}
            <Button
              size="small"
              color="inherit"
              onClick={() =>
                setDraftReturnFilter((current) => ({
                  ...current,
                  clauses: current.clauses.filter((_item, clauseIndex) => clauseIndex !== index),
                }))
              }
            >
              Remove
            </Button>
          </Box>
        ))}

        <Box display="flex" alignItems="center" gap={1}>
          <Button
            variant="outlined"
            disabled={!canApplyReturnFilter}
            onClick={() => setReturnFilter(draftReturnFilter)}
          >
            Apply return filter
          </Button>
          {returnFilter.clauses.length > 0 && (
            <>
              <Button variant="text" color="inherit" onClick={clearReturnFilter}>
                Clear
              </Button>
              <Chip
                variant="outlined"
                label={`${returnFilter.clauses.length} condition${returnFilter.clauses.length === 1 ? '' : 's'} (${returnFilter.logic.toUpperCase()})`}
                onDelete={clearReturnFilter}
              />
            </>
          )}
        </Box>
      </Box>
    </Box>
  );
};

export default ReturnsSearch;
