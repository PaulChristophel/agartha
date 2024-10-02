import jsYaml from 'js-yaml';
import { keymap } from '@codemirror/view';
import { useState, useEffect } from 'react';
import { yaml } from '@codemirror/lang-yaml';
import CodeMirror from '@uiw/react-codemirror';
import { foldKeymap } from '@codemirror/language';
import { Link, useParams } from 'react-router-dom';
import { autocompletion } from '@codemirror/autocomplete';
import { search, searchKeymap } from '@codemirror/search';

import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';

import useSaltReturn from 'src/hooks/saltReturn/useSaltReturn.ts';

import formatTime from 'src/utils/formatTime.ts';

export default function ReturnDetailsPage() {
  const { jid, id } = useParams<Record<string, string>>();
  const { alterTime, fun, returnId, returnJid, fullRet, success, isLoading, error } = useSaltReturn(
    jid as string,
    id as string,
    false,
    true
  );
  const [formattedLoad, setFormattedLoad] = useState<string>('');

  useEffect(() => {
    if (fullRet) {
      setFormattedLoad(jsYaml.dump(fullRet));
    }
  }, [fullRet]);

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
      <h1>Return Details</h1>
      <p>
        Minion ID: <Link to={`/minion/${returnId}`}>{returnId}</Link>
      </p>
      <p>
        Job ID: <Link to={`/returns/?jid=${returnJid}`}>{returnJid}</Link>
      </p>
      <p>
        Function: <Link to={`/returns/?fun=${fun}`}>{fun}</Link>
      </p>
      <p>
        Status:{' '}
        <Chip
          variant="outlined"
          label={success?.toString() || 'N/A'}
          color={success ? 'success' : 'error'}
        />
      </p>
      <p>Start Time: {formatTime(alterTime)}</p>
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
}
