import React, { useState, useEffect } from 'react';

import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import Button from '@mui/material/Button';
import Tooltip from '@mui/material/Tooltip';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import {
  parseDataPath,
  emptyDataFilter,
  type DataValueType,
  newDataFilterClause,
  type DataFilterState,
  type DataFilterScope,
  type DataFilterClause,
  dataOperatorNeedsValue,
  type DataFilterOperator,
} from './dataFilter.ts';

interface DataFilterBuilderProps {
  dataFilter: DataFilterState;
  setDataFilter: (value: DataFilterState) => void;
}

const editableDataFilter = (filter: DataFilterState): DataFilterState =>
  filter.clauses.length > 0 ? filter : { logic: 'and', clauses: [newDataFilterClause()] };

const numericOperators: DataFilterOperator[] = ['gt', 'gte', 'lt', 'lte'];
const stringOperators: DataFilterOperator[] = ['contains', 'icontains', 'regex'];
const hoverHelpProps = {
  arrow: true,
  placement: 'top' as const,
  enterDelay: 700,
  enterNextDelay: 700,
  leaveDelay: 0,
  disableFocusListener: true,
  disableInteractive: true,
};

const DataFilterBuilder: React.FC<DataFilterBuilderProps> = ({ dataFilter, setDataFilter }) => {
  const [draftDataFilter, setDraftDataFilter] = useState<DataFilterState>(() =>
    editableDataFilter(dataFilter)
  );

  useEffect(() => {
    setDraftDataFilter(editableDataFilter(dataFilter));
  }, [dataFilter]);

  const updateDataClause = (index: number, changes: Partial<DataFilterClause>) => {
    setDraftDataFilter((current) => ({
      ...current,
      clauses: current.clauses.map((clause, clauseIndex) =>
        clauseIndex === index ? { ...clause, ...changes } : clause
      ),
    }));
  };

  const changeDataOperator = (index: number, operator: DataFilterOperator) => {
    const current = draftDataFilter.clauses[index];
    let { valueType } = current;
    if (numericOperators.includes(operator) && valueType !== 'int' && valueType !== 'float') {
      valueType = 'float';
    } else if (stringOperators.includes(operator)) {
      valueType = 'string';
    }
    updateDataClause(index, { operator, valueType });
  };

  const clauseIsValid = (clause: DataFilterClause): boolean => {
    if (clause.scope === 'key' && !clause.key) return false;
    const containerPath = parseDataPath(clause.containerPath);
    if (
      containerPath.length > 16 ||
      containerPath.some((component) => !component) ||
      (clause.scope !== 'any_key' && containerPath.length > 0)
    ) {
      return false;
    }
    const path = parseDataPath(clause.path);
    if (path.length > 16 || path.some((component) => !component)) return false;
    if (!dataOperatorNeedsValue(clause.operator)) return true;
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

  const canApplyDataFilter =
    draftDataFilter.clauses.length > 0 && draftDataFilter.clauses.every(clauseIsValid);

  const clearDataFilter = () => {
    const emptyFilter = emptyDataFilter();
    setDraftDataFilter(editableDataFilter(emptyFilter));
    setDataFilter(emptyFilter);
  };

  return (
    <Box display="flex" flexDirection="column" gap={1.5} sx={{ flex: '1 1 100%' }}>
      <Box display="flex" alignItems="center" flexWrap="wrap" gap={1}>
        <Typography variant="subtitle2">Event data conditions</Typography>
        {draftDataFilter.clauses.length > 1 && (
          <Tooltip
            {...hoverHelpProps}
            title="AND requires every condition to match. OR includes an event when any condition matches."
          >
            <TextField
              select
              size="small"
              label="Combine with"
              value={draftDataFilter.logic}
              onChange={(event) =>
                setDraftDataFilter((current) => ({
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
          disabled={draftDataFilter.clauses.length >= 20}
          onClick={() =>
            setDraftDataFilter((current) => ({
              ...current,
              clauses: [...current.clauses, newDataFilterClause()],
            }))
          }
        >
          Add condition
        </Button>
      </Box>

      {draftDataFilter.clauses.map((clause, index) => (
        <Box
          key={`data-clause-${index}`}
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
            title="Choose the whole data value, one exact top-level object key, or every top-level object entry."
          >
            <TextField
              select
              size="small"
              label="Scope"
              value={clause.scope}
              onChange={(event) => {
                const scope = event.target.value as DataFilterScope;
                updateDataClause(index, {
                  scope,
                  containerPath: scope === 'any_key' ? clause.containerPath : '',
                });
              }}
              sx={{ flex: '1 1 170px' }}
            >
              <MenuItem value="key">Exact object key</MenuItem>
              <MenuItem value="any_key">Any object entry</MenuItem>
              <MenuItem value="root">Root data</MenuItem>
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
                onChange={(event) => updateDataClause(index, { key: event.target.value })}
                placeholder="Exact top-level key in the event data"
                sx={{ flex: '3 1 420px' }}
              />
            </Tooltip>
          )}
          {clause.scope === 'any_key' && (
            <Tooltip
              {...hoverHelpProps}
              title="Optionally select the object whose entries should be searched. For highstate event data, use return. Leave blank to search top-level data entries."
            >
              <TextField
                size="small"
                label="Container path"
                value={clause.containerPath}
                onChange={(event) =>
                  updateDataClause(index, { containerPath: event.target.value })
                }
                placeholder="return"
                sx={{ flex: '1 1 220px' }}
              />
            </Tooltip>
          )}
          <Tooltip
            {...hoverHelpProps}
            title={
              'Enter a dot-separated nested path, such as return.result or payload.changes.diff. Numeric components select array indexes. Escape a literal dot as \\.'
            }
          >
            <TextField
              size="small"
              label={clause.scope === 'any_key' ? 'Entry field path' : 'Field path'}
              value={clause.path}
              onChange={(event) => updateDataClause(index, { path: event.target.value })}
              placeholder={clause.scope === 'root' ? 'Leave blank for the root value' : 'result'}
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
                changeDataOperator(index, event.target.value as DataFilterOperator)
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
          {dataOperatorNeedsValue(clause.operator) && (
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
                    onChange={(event) => updateDataClause(index, { value: event.target.value })}
                  />
                </Box>
              </Tooltip>
              <Tooltip
                {...hoverHelpProps}
                title="Select the JSON type used to parse the comparison value. Numeric-looking identifiers should usually remain strings."
              >
                <TextField
                  select
                  size="small"
                  label="Type"
                  value={clause.valueType}
                  onChange={(event) =>
                    updateDataClause(index, {
                      valueType: event.target.value as DataValueType,
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
              setDraftDataFilter((current) => ({
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
          disabled={!canApplyDataFilter}
          onClick={() => setDataFilter(draftDataFilter)}
        >
          Apply data filter
        </Button>
        {dataFilter.clauses.length > 0 && (
          <>
            <Button variant="text" color="inherit" onClick={clearDataFilter}>
              Clear
            </Button>
            <Chip
              variant="outlined"
              label={`${dataFilter.clauses.length} condition${dataFilter.clauses.length === 1 ? '' : 's'} (${dataFilter.logic.toUpperCase()})`}
              onDelete={clearDataFilter}
            />
          </>
        )}
      </Box>
    </Box>
  );
};

export default DataFilterBuilder;
