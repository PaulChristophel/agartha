import yaml from 'js-yaml';
import React, { useMemo } from 'react';
import { keymap } from '@codemirror/view';
import { useParams } from 'react-router-dom';
import CodeMirror from '@uiw/react-codemirror';
import { foldKeymap } from '@codemirror/language';
import { yaml as yamlLang } from '@codemirror/lang-yaml';
import { autocompletion } from '@codemirror/autocomplete';
import { search, searchKeymap } from '@codemirror/search';

import CircularProgress from '@mui/material/CircularProgress';

import useHighState from 'src/hooks/highState/useHighState.ts';

interface ConformityDetailsPageParams extends Record<string, string | undefined> {
  id: string;
}

const ConformityDetailsPage: React.FC = () => {
  const { id } = useParams<ConformityDetailsPageParams>();

  const { fullRet, isLoading, error } = useHighState(id || '', false, true);

  const yamlData = useMemo(() => {
    try {
      if (fullRet) {
        // Create a new object to reorder the keys
        const { return: returnValue, ...rest } = fullRet;
        const reorderedObject = { ...rest, return: returnValue };
        return yaml.dump(reorderedObject);
      }
      return '';
    } catch (e) {
      console.error('Failed to convert data to YAML:', e);
      return '';
    }
  }, [fullRet]);

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
    <CodeMirror
      maxHeight="1024px"
      value={yamlData}
      extensions={[
        yamlLang(),
        keymap.of([...foldKeymap, ...searchKeymap]),
        autocompletion(),
        search({
          top: true, // position search bar at the top
        }),
      ]}
      theme="dark"
      readOnly
    />
  );
};

export default ConformityDetailsPage;
