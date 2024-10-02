import React from 'react';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { solarizedDarkAtom } from 'react-syntax-highlighter/dist/esm/styles/prism';

import Fab from '@mui/material/Fab';
import DeleteSweepIcon from '@mui/icons-material/DeleteSweep';

const OutputView: React.FC<{ output: string; clearOutput: () => void }> = ({
  output,
  clearOutput,
}) => {
  const terminalStyle: React.CSSProperties = {
    backgroundColor: '#002b36',
    color: '#002b36',
    padding: '10px',
    borderRadius: '5px',
    fontFamily: 'monospace',
    fontSize: '14px',
    height: '1024px',
    maxHeight: '100%',
    width: '100%',
    overflow: 'auto',
    position: 'relative',
  };

  return (
    <div style={terminalStyle}>
      {output && (
        <Fab
          color="primary"
          aria-label="clear"
          size="small"
          style={{ float: 'right' }}
          sx={{ marginLeft: 1 }}
          onClick={clearOutput}
        >
          <DeleteSweepIcon />
        </Fab>
      )}
      <SyntaxHighlighter language="yaml" style={solarizedDarkAtom}>
        {output}
      </SyntaxHighlighter>
    </div>
  );
};

export default OutputView;
