import React, { useState } from 'react';

import Box from '@mui/material/Box';
import Fab from '@mui/material/Fab';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import Grid from '@mui/material/Grid';
import Alert from '@mui/material/Alert';
import AddIcon from '@mui/icons-material/Add';
import Typography from '@mui/material/Typography';
import { SelectChangeEvent } from '@mui/material/Select';
import Snackbar, { SnackbarCloseReason } from '@mui/material/Snackbar';

import CLIView from './CLIView.tsx';
import OutputView from './OutputView.tsx';
import FormattedView from './FormattedView.tsx';

const RunView: React.FC = () => {
  const [tab, setTab] = useState(0);
  const [clientType, setClientType] = useState('local');
  const [output, setOutput] = useState<string[]>([]);
  const [alertStatus, setAlertStatus] = useState<'success' | 'error' | 'warning' | 'info' | null>(
    null
  );
  const [open, setOpen] = useState(false);
  const [infoMessage, setInfoMessage] = useState<string | null>(null);

  const handleChange = (_event: React.SyntheticEvent, newValue: number) => {
    setTab(newValue);
  };

  const handleClientTypeChange = (event: SelectChangeEvent<string>) => {
    setClientType(event.target.value as string);
  };

  const clearOutput = () => {
    setOutput([]);
    setAlertStatus(null);
  };

  const handleSetOutput = ({
    output: newOutput,
    status,
  }: {
    output: string;
    status: 'success' | 'error' | 'warning';
  }) => {
    setAlertStatus(status);
    setOutput((prevOutput) => [...prevOutput, newOutput]);
    setOpen(true);
  };

  const handleClose = (_event: React.SyntheticEvent | Event, reason?: SnackbarCloseReason) => {
    if (reason === 'clickaway') {
      return;
    }
    setOpen(false);
  };

  const showInfoMessage = (message: string) => {
    setInfoMessage(message);
    setAlertStatus('info');
    setOpen(true);
  };

  const getAlertMessage = () => {
    if (alertStatus === 'success') {
      return 'Command executed successfully!';
    }
    if (alertStatus === 'info') {
      return infoMessage;
    }
    return 'An error occurred.';
  };

  return (
    <Grid container spacing={2}>
      <Grid container item spacing={1}>
        <Box sx={{ p: 3, borderRadius: 1, border: '1px solid #444' }}>
          <Fab
            color="primary"
            aria-label="add"
            size="small"
            style={{ float: 'right' }}
            sx={{ marginLeft: 1 }}
            className="hidden"
          >
            <AddIcon />
          </Fab>
          <Typography variant="h5" gutterBottom>
            Run
          </Typography>
          <Tabs value={tab} onChange={handleChange} sx={{ p: 2, gap: 2 }}>
            <Tab label="Formatted" />
            <Tab label="CLI" className="hidden" />
          </Tabs>
          {tab === 0 ? (
            <FormattedView
              clientType={clientType}
              onClientTypeChange={handleClientTypeChange}
              tgt=""
              tgtType="glob"
              fun=""
              timeout=""
              aSync={false}
              batch=""
              setOutput={handleSetOutput}
              showInfoMessage={showInfoMessage}
            />
          ) : (
            <CLIView />
          )}
        </Box>
      </Grid>
      <Grid container item spacing={1}>
        <OutputView output={output.join('\n')} clearOutput={clearOutput} />
      </Grid>
      {alertStatus && (
        <Snackbar open={open} autoHideDuration={6000} onClose={handleClose}>
          <Alert onClose={handleClose} severity={alertStatus} sx={{ width: '100%' }}>
            {alertStatus && (
              <Snackbar open={open} autoHideDuration={6000} onClose={handleClose}>
                <Alert onClose={handleClose} severity={alertStatus} sx={{ width: '100%' }}>
                  {getAlertMessage()}
                </Alert>
              </Snackbar>
            )}
          </Alert>
        </Snackbar>
      )}
    </Grid>
  );
};

export default RunView;
