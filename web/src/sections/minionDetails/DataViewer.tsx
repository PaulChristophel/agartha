import React from 'react';
import { keymap } from '@codemirror/view';
import { yaml } from '@codemirror/lang-yaml';
import CodeMirror from '@uiw/react-codemirror';
import { autocompletion } from '@codemirror/autocomplete';
import { search, searchKeymap } from '@codemirror/search';

interface DataViewerComponentProps {
  data: string;
}

const DataViewer: React.FC<DataViewerComponentProps> = ({ data }) => (
  <CodeMirror
    value={data}
    maxHeight="1024px"
    extensions={[
      yaml(),
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

export default DataViewer;
