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

import useSaltEvent from 'src/hooks/saltEvent/useSaltEvent.ts';

import formatTime from 'src/utils/formatTime.ts';

const createTagLink = (tag: string) => {
  const jobRegex = /salt\/job\/(\d+)\/(ret|sub)\/([^/]+)/;
  const runRegex = /salt\/run\/(\d+)\/(ret|new)|(\d{20})/;
  let match = tag.match(jobRegex);

  if (match) {
    const jobId = match[1];
    const minionId = match[3];
    return <Link to={`/return/${jobId}/${minionId}/`}>{tag}</Link>;
  }

  match = tag.match(runRegex);
  if (match) {
    const jobId = match[1] || match[3];
    return <Link to={`/job/${jobId}`}>{tag}</Link>;
  }

  return tag;
};

export default function EventDetailsView() {
  const { id } = useParams<{ id: string }>();
  const numericId = id ? parseInt(id, 10) : null;
  const { alterTime, eventData, masterID, tag, isLoading, error } = useSaltEvent(numericId ?? 0);
  const [formattedLoad, setFormattedLoad] = useState<string>('');

  useEffect(() => {
    if (eventData) {
      setFormattedLoad(jsYaml.dump(eventData));
    }
  }, [eventData]);

  if (!id || numericId === null) {
    return <div>Error: ID parameter is missing or invalid</div>;
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
      <h1>Event Details</h1>
      <p>ID: {id}</p>
      <p>Tag: {createTagLink(tag || '')}</p>
      <p>Master ID: {masterID}</p>
      <p>Start Time: {formatTime(alterTime)}</p>
      <div style={{ marginTop: '20px' }}>
        <h2>Data:</h2>
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
