import yaml from 'js-yaml';
import React, { useMemo } from 'react';
import { keymap } from '@codemirror/view';
import CodeMirror from '@uiw/react-codemirror';
import { yaml as yamlLang } from '@codemirror/lang-yaml';
import { autocompletion } from '@codemirror/autocomplete';
import { search, searchKeymap } from '@codemirror/search';

import CircularProgress from '@mui/material/CircularProgress';

import useSaltCacheBankKey from 'src/hooks/saltCache/useSaltCacheBankKey.ts';

interface CodeViewComponentProps {
  bank: string;
  cacheKey: string;
}

const CodeView: React.FC<CodeViewComponentProps> = ({ bank, cacheKey }) => {
  const { cacheData, isLoading, error } = useSaltCacheBankKey(bank, cacheKey);

  const yamlData = useMemo(() => {
    try {
      return yaml.dump(cacheData);
    } catch (e) {
      console.error('Failed to convert data to YAML:', e);
      return '';
    }
  }, [cacheData]);

  if (isLoading) {
    return <CircularProgress color="success" />;
  }

  if (error) {
    return <div>Error loading data</div>;
  }

  return (
    <CodeMirror
      value={yamlData}
      maxHeight="1024px"
      extensions={[
        yamlLang(),
        keymap.of([...searchKeymap]),
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

export default CodeView;
