import jsYaml from 'js-yaml';
import { keymap } from '@codemirror/view';
import { yaml } from '@codemirror/lang-yaml';
import CodeMirror from '@uiw/react-codemirror';
import { foldKeymap } from '@codemirror/language';
import React, { useState, useEffect } from 'react';
import { autocompletion } from '@codemirror/autocomplete';
import { search, searchKeymap } from '@codemirror/search';
import { Link, useParams, useNavigate, useLocation } from 'react-router-dom';

import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';

import useHighState from 'src/hooks/highState/useHighState.ts';
import useConformity from 'src/hooks/conformity/useConformity.ts';

import formatTime from 'src/utils/formatTime.ts';

interface ConformityDetailsPageParams extends Record<string, string | undefined> {
  id: string;
}

interface ReturnData {
  result: boolean;
  changes?: Record<string, unknown>;
}

const ConformityDetailsPage: React.FC = () => {
  const { id } = useParams<ConformityDetailsPageParams>();
  const navigate = useNavigate();
  const location = useLocation();
  const query = new URLSearchParams(location.search);
  const filterFromUrl = query.get('filter');

  const { alterTime, fun, returnId, returnJid, returnData, success, isLoading, error } =
    useHighState(id || '', true, false);
  const { trueCount, falseCount, changedCount, unchangedCount } = useConformity(id || '');
  const [formattedLoad, setFormattedLoad] = useState('');
  const [currentFilter, setCurrentFilter] = useState<string | null>(filterFromUrl || null);

  useEffect(() => {
    if (returnData) {
      let filteredData: Record<string, ReturnData> = {};

      switch (currentFilter) {
        case 'succeeded':
          filteredData = Object.fromEntries(
            Object.entries(returnData).filter(([, value]) => (value as ReturnData).result === true)
          ) as Record<string, ReturnData>;
          break;
        case 'failed':
          filteredData = Object.fromEntries(
            Object.entries(returnData).filter(([, value]) => (value as ReturnData).result === false)
          ) as Record<string, ReturnData>;
          break;
        case 'changed':
          filteredData = Object.fromEntries(
            Object.entries(returnData).filter(
              ([, value]) =>
                (value as ReturnData).changes &&
                Object.keys((value as ReturnData).changes!).length > 0
            )
          ) as Record<string, ReturnData>;
          break;
        case 'unchanged':
          filteredData = Object.fromEntries(
            Object.entries(returnData).filter(
              ([, value]) =>
                !(value as ReturnData).changes ||
                Object.keys((value as ReturnData).changes!).length === 0
            )
          ) as Record<string, ReturnData>;
          break;
        default:
          filteredData = returnData as Record<string, ReturnData>;
      }

      setFormattedLoad(jsYaml.dump(filteredData));
    }
  }, [returnData, currentFilter]);

  const handleFilter = (type: string) => {
    const newFilter = currentFilter === type ? null : type;
    setCurrentFilter(newFilter);
    const newQuery = new URLSearchParams(location.search);
    if (newFilter) {
      newQuery.set('filter', newFilter);
    } else {
      newQuery.delete('filter');
    }
    navigate({ search: newQuery.toString() });
  };

  if (!id) {
    return <div>Error: ID is required</div>;
  }

  if (isLoading) {
    return <CircularProgress color="success" />;
  }

  if (error) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div
      style={{
        padding: '20px',
        border: '1px solid #ccc',
        borderRadius: '8px',
        backgroundColor: '#f9f9f9',
      }}
    >
      <h1>Conformity Details</h1>
      <div style={{ marginBottom: '10px' }}>
        Minion ID: <Link to={`/minion/${returnId}`}>{returnId}</Link>
      </div>
      <div style={{ marginBottom: '10px' }}>
        Job ID: <Link to="/returns/?jid={returnJid}">{returnJid}</Link>
      </div>
      <div style={{ marginBottom: '10px' }}>
        Function: <Link to={`/returns/?fun=${fun}`}>{fun}</Link>
      </div>
      <div style={{ marginBottom: '10px' }}>
        Status:
        <Chip
          variant="outlined"
          label={success?.toString() || 'N/A'}
          color={success?.toString() === 'true' ? 'success' : 'error'}
        />
      </div>
      <div style={{ marginBottom: '10px' }}>
        Succeeded:
        <Chip
          variant={currentFilter === 'succeeded' ? 'filled' : 'outlined'}
          label={trueCount?.toString() || 'N/A'}
          color={trueCount > 0 ? 'success' : 'error'}
          onClick={() => handleFilter('succeeded')}
        />
      </div>
      <div style={{ marginBottom: '10px' }}>
        Failed:
        <Chip
          variant={currentFilter === 'failed' ? 'filled' : 'outlined'}
          label={falseCount?.toString() || 'N/A'}
          color={falseCount === 0 ? 'success' : 'error'}
          onClick={() => handleFilter('failed')}
        />
      </div>
      <div style={{ marginBottom: '10px' }}>
        Changed:
        <Chip
          variant={currentFilter === 'changed' ? 'filled' : 'outlined'}
          label={changedCount?.toString() || 'N/A'}
          color={changedCount === 0 ? 'info' : 'warning'}
          onClick={() => handleFilter('changed')}
        />
      </div>
      <div style={{ marginBottom: '10px' }}>
        Unchanged:
        <Chip
          variant={currentFilter === 'unchanged' ? 'filled' : 'outlined'}
          label={unchangedCount?.toString() || 'N/A'}
          color={unchangedCount === 0 ? 'warning' : 'info'}
          onClick={() => handleFilter('unchanged')}
        />
      </div>
      <div style={{ marginBottom: '10px' }}>Start Time: {formatTime(alterTime)}</div>
      <div style={{ marginTop: '20px' }}>
        <h2>Return:</h2>
        <CodeMirror
          value={formattedLoad}
          extensions={[
            yaml(),
            keymap.of([...foldKeymap, ...searchKeymap]),
            autocompletion(),
            search({
              top: true, // position search bar at the top
            }),
          ]}
          theme="dark"
          readOnly
        />
      </div>
    </div>
  );
};

export default ConformityDetailsPage;
