import jsYaml from 'js-yaml';
import { keymap } from '@codemirror/view';
import { useState, useEffect } from 'react';
import { yaml } from '@codemirror/lang-yaml';
import CodeMirror from '@uiw/react-codemirror';
import { foldKeymap } from '@codemirror/language';
import { Link, useParams } from 'react-router-dom';
import { autocompletion } from '@codemirror/autocomplete';
import { search, searchKeymap } from '@codemirror/search';

import CircularProgress from '@mui/material/CircularProgress';

import useJid from 'src/hooks/jid/useJid.ts';

import formatTime from 'src/utils/formatTime.ts';

export default function JidDetailsPage() {
  const { jid } = useParams<Record<string, string>>();
  const { alterTime, load, isLoading, error } = useJid(jid as string);
  const [formattedLoad, setFormattedLoad] = useState<string>('');

  useEffect(() => {
    if (load) {
      setFormattedLoad(jsYaml.dump(load));
    }
  }, [load]);

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
      <h1>Job Details</h1>
      <p>
        Job ID: <Link to={`/returns/?jid=${jid}`}>{jid}</Link>
      </p>
      <p>Start Time: {formatTime(alterTime)}</p>
      <div style={{ marginTop: '20px' }}>
        <h2>Load:</h2>
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
