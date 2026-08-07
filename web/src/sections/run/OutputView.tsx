import React, { lazy, Suspense } from 'react';

import Fab from '@mui/material/Fab';
import DeleteSweepIcon from '@mui/icons-material/DeleteSweep';

const OutputRenderer = lazy(() => import('./OutputRenderer.tsx'));

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
      {output && (
        <Suspense fallback={<pre>{output}</pre>}>
          <OutputRenderer output={output} />
        </Suspense>
      )}
    </div>
  );
};

export default OutputView;
