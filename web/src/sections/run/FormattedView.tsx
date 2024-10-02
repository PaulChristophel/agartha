import yaml from 'js-yaml';
import '@xterm/xterm/css/xterm.css';
import CodeMirror from '@uiw/react-codemirror';
import React, { useState, useEffect } from 'react';
import { LanguageSupport } from '@codemirror/language';
import { json as jsonLang } from '@codemirror/lang-json';
import { yaml as yamlLang } from '@codemirror/lang-yaml';

import Grid from '@mui/material/Grid';
import Button from '@mui/material/Button';
import Switch from '@mui/material/Switch';
import Tooltip from '@mui/material/Tooltip';
import MenuItem from '@mui/material/MenuItem';
import Checkbox from '@mui/material/Checkbox';
import TextField from '@mui/material/TextField';
import InputLabel from '@mui/material/InputLabel';
import FormControl from '@mui/material/FormControl';
import FormControlLabel from '@mui/material/FormControlLabel';
import Select, { SelectChangeEvent } from '@mui/material/Select';

import { handleRun } from './handleRun.ts';

const FormattedView: React.FC<{
  aSync: boolean;
  batch: string;
  fun: string;
  tgt: string;
  tgtType: string;
  timeout: string;
  clientType: string;
  onClientTypeChange: (event: SelectChangeEvent<string>) => void;
  setOutput: (data: { output: string; status: 'success' | 'error' | 'warning' }) => void;
  showInfoMessage: (message: string) => void;
}> = ({
  aSync,
  batch,
  fun,
  tgt,
  tgtType,
  timeout,
  clientType,
  onClientTypeChange,
  setOutput,
  showInfoMessage,
}) => {
  const [asyncState, setAsyncState] = useState(aSync || false);
  const [batchState, setBatchState] = useState(batch || '');
  const [funState, setFunState] = useState(fun || '');
  const [tgtState, setTgtState] = useState(tgt || '');
  const [timeoutState, setTimeoutState] = useState(timeout || '');
  const [tgtTypeState, setTgtTypeState] = useState(tgtType || 'glob');
  const [clientTypeState, setClientTypeState] = useState(clientType || 'local');
  const [argumentsState, setArgumentsState] = useState('');
  const [kwArgumentsState, setKwArgumentsState] = useState('');
  const [pillarVisible, setPillarVisible] = useState(false);
  const [pillarValue, setPillarValue] = useState('');
  const [codeMirrorMode, setCodeMirrorMode] = useState<LanguageSupport>(yamlLang());

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    setAsyncState(params.get('async') === 'true');
    setBatchState(params.get('batch') || batch);
    setTimeoutState(params.get('timeout') || '');
    setFunState(params.get('fun') || fun);
    setTgtState(params.get('tgt') || tgt);
    setTgtTypeState(params.get('tgt_type') || tgtType);
    setClientTypeState(params.get('client') || clientType);
    setArgumentsState(params.get('args') || '');
    setKwArgumentsState(params.get('kwargs') || '');
    if (params.has('pillar')) {
      setPillarVisible(true);
      setPillarValue(params.get('pillar') || '');
    }
  }, [batch, fun, tgt, tgtType, clientType]);

  const handleCodeChange = (value: string) => {
    setPillarValue(value);
    if (value.trim().startsWith('{') || value.trim().startsWith('[')) {
      setCodeMirrorMode(jsonLang());
    } else {
      setCodeMirrorMode(yamlLang());
    }
  };

  const updateURLParams = (key: string, value: string | boolean) => {
    const params = new URLSearchParams(window.location.search);
    if (value) {
      params.set(key, value.toString());
    } else {
      params.delete(key);
    }
    window.history.replaceState({}, '', `${window.location.pathname}?${params.toString()}`);
  };

  const onRunClick = async () => {
    showInfoMessage('Command sent.');
    const authToken = localStorage.getItem('authToken') || '';
    const authSaltString = localStorage.getItem('authSalt') || '';
    try {
      const output = await handleRun(
        authToken,
        authSaltString,
        clientTypeState,
        asyncState,
        batchState,
        funState,
        tgtState,
        tgtTypeState,
        timeoutState,
        argumentsState,
        kwArgumentsState,
        pillarValue,
        pillarVisible
      );

      // Parse the JSON output to an object
      const outputObj = JSON.parse(output);

      // Convert the object to YAML with proper formatting
      const outputYaml = yaml.dump(outputObj, { indent: 2, lineWidth: 80 });

      setOutput({ output: outputYaml, status: 'success' });
    } catch (error: unknown) {
      console.error('Error running command:', error);
      setOutput({ output: 'An error occurred while running the command.', status: 'error' });
    }
  };

  const onRunTestClick = async () => {
    showInfoMessage('Test command sent.');
    const authToken = localStorage.getItem('authToken') || '';
    const authSaltString = localStorage.getItem('authSalt') || '';
    try {
      const output = await handleRun(
        authToken,
        authSaltString,
        clientTypeState,
        asyncState,
        batchState,
        funState,
        tgtState,
        tgtTypeState,
        timeoutState,
        argumentsState,
        kwArgumentsState,
        pillarValue,
        pillarVisible,
        true
      );

      // Parse the JSON output to an object
      const outputObj = JSON.parse(output);

      // Convert the object to YAML with proper formatting
      const outputYaml = yaml.dump(outputObj, { indent: 2, lineWidth: 80 });

      setOutput({ output: outputYaml, status: 'success' });
    } catch (error: unknown) {
      console.error('Error running test command:', error);
      setOutput({ output: 'An error occurred while running the test command.', status: 'error' });
    }
  };

  return (
    <Grid container spacing={2} alignItems="center">
      <Grid item xs={12} sm={6} md={3}>
        <FormControl fullWidth variant="outlined">
          <InputLabel>Client Type</InputLabel>
          <Select
            label="Client Type"
            value={clientTypeState}
            onChange={(e) => {
              setClientTypeState(e.target.value as string);
              onClientTypeChange(e);
              updateURLParams('client', e.target.value as string);
            }}
            onBlur={() => updateURLParams('client', clientTypeState)}
          >
            <MenuItem value="local">Local (salt)</MenuItem>
            <MenuItem value="runner">Runner (salt-run)</MenuItem>
            <MenuItem value="ssh">SSH (salt-ssh)</MenuItem>
            <MenuItem value="wheel">Wheel (salt-key)</MenuItem>
            <MenuItem value="cloud" className="hidden">
              Cloud (salt-cloud)
            </MenuItem>
          </Select>
        </FormControl>
      </Grid>
      <Grid
        item
        xs={6}
        sm={3}
        className={batchState !== '' || clientTypeState === 'ssh' ? 'hidden' : ''}
      >
        <FormControlLabel
          control={
            <Tooltip title="Run the salt command but don't wait for a reply.">
              <Checkbox
                checked={asyncState}
                onChange={(e) => {
                  setAsyncState(e.target.checked);
                  if (e.target.checked) {
                    updateURLParams('batch', '');
                  }
                  updateURLParams('async', e.target.checked.toString());
                }}
                onBlur={() => updateURLParams('async', asyncState.toString())}
              />
            </Tooltip>
          }
          label="Async"
        />
      </Grid>
      <Grid
        item
        xs={6}
        sm={3}
        className={
          clientTypeState === 'runner' ||
          clientTypeState === 'wheel' ||
          clientTypeState === 'ssh' ||
          asyncState
            ? 'hidden'
            : ''
        }
      >
        <Tooltip title="Execute the salt job in batch mode, pass either the number of minions to batch at a time, or the percentage of minions to have running.">
          <TextField
            label="Batch"
            variant="outlined"
            value={batchState}
            onChange={(e) => {
              const re = /^[0-9\b]+[%]{0,1}$/;
              if (re.test(e.target.value) || e.target.value === '') {
                setBatchState(e.target.value);
                updateURLParams('batch', e.target.value);
              }
            }}
            onBlur={() => updateURLParams('batch', batchState)}
          />
        </Tooltip>
      </Grid>
      <Grid
        item
        xs={12}
        sm={6}
        md={3}
        className={clientTypeState === 'runner' || clientTypeState === 'wheel' ? 'hidden' : ''}
      >
        <Tooltip title="Change the timeout, if applicable, for the running command (in seconds).">
          <TextField
            label="Timeout"
            variant="outlined"
            value={timeoutState}
            onChange={(e) => {
              const re = /^[0-9\b]+$/;
              if (re.test(e.target.value) || e.target.value === '') {
                setTimeoutState(e.target.value);
                updateURLParams('timeout', e.target.value);
              }
            }}
            onBlur={() => updateURLParams('timeout', timeoutState)}
          />
        </Tooltip>
      </Grid>
      <Grid
        item
        xs={12}
        sm={6}
        md={1}
        className={clientTypeState === 'runner' || clientTypeState === 'wheel' ? 'hidden' : ''}
      >
        <Tooltip
          placement="top-end"
          title={
            <>
              The type of tgt.
              <br />
              glob: Bash glob completion.
              <br />
              pcre: Perl style regular expression.
              <br />
              list: Python list of hosts.
              <br />
              grain: Match based on a grain comparison.
              <br />
              grain_pcre: Grain comparison with a regex.
              <br />
              pillar: Pillar data comparison.
              <br />
              pillar_pcre: Pillar data comparison with a regex.
              <br />
              nodegroup: Match on nodegroup.
              <br />
              range: Use a Range server for matching.
              <br />
              compound: Pass a compound match string.
              <br />
              ipcidr: Match based on Subnet (CIDR notation) or IPv4 address.
            </>
          }
        >
          <FormControl fullWidth variant="outlined">
            <InputLabel>Target Type</InputLabel>
            <Select
              label="Target Type"
              value={tgtTypeState}
              onChange={(e) => {
                setTgtTypeState(e.target.value as string);
                updateURLParams('tgt_type', e.target.value as string);
              }}
              onBlur={() => updateURLParams('tgt_type', tgtTypeState)}
            >
              <MenuItem value="glob">glob</MenuItem>
              <MenuItem value="pcre">pcre</MenuItem>
              <MenuItem value="list">list</MenuItem>
              <MenuItem value="grain">grain</MenuItem>
              <MenuItem value="grain_pcre">grain_pcre</MenuItem>
              <MenuItem value="pillar">pillar</MenuItem>
              <MenuItem value="pillar_pcre">pillar_pcre</MenuItem>
              <MenuItem value="range">range</MenuItem>
              <MenuItem value="compound">compound</MenuItem>
              <MenuItem value="nodegroup">nodegroup</MenuItem>
            </Select>
          </FormControl>
        </Tooltip>
      </Grid>
      <Grid
        item
        xs={12}
        sm={6}
        md={2}
        className={clientTypeState === 'runner' || clientTypeState === 'wheel' ? 'hidden' : ''}
      >
        <Tooltip title="Which minions to target for the execution. ">
          <TextField
            label="Target"
            variant="outlined"
            fullWidth
            value={tgtState}
            onChange={(e) => setTgtState(e.target.value)}
            onBlur={() => updateURLParams('tgt', tgtState)}
          />
        </Tooltip>
      </Grid>
      <Grid item xs={12} sm={6} md={2}>
        <Tooltip title="The module and function to call on the specified minions of the form module.function. For example test.ping or grains.items.">
          <TextField
            label="Function"
            variant="outlined"
            fullWidth
            value={funState}
            onChange={(e) => setFunState(e.target.value)}
            onBlur={() => updateURLParams('fun', funState)}
          />
        </Tooltip>
      </Grid>
      <Grid item xs={12} sm={6} md={2}>
        <Tooltip title="A list of arguments to pass to the remote function. If the function takes no arguments arg may be omitted except when executing a compound command.">
          <TextField
            label="Arguments"
            variant="outlined"
            fullWidth
            value={argumentsState}
            onChange={(e) => setArgumentsState(e.target.value)}
            onBlur={() => updateURLParams('args', argumentsState)}
          />
        </Tooltip>
      </Grid>
      <Grid item xs={12} sm={6} md={4}>
        <Tooltip title="Keyword arguments for the function.">
          <TextField
            label="Keyword Arguments"
            variant="outlined"
            fullWidth
            value={kwArgumentsState}
            onChange={(e) => setKwArgumentsState(e.target.value)}
            onBlur={() => updateURLParams('kwargs', kwArgumentsState)}
          />
        </Tooltip>
      </Grid>
      <Grid
        item
        xs={6}
        sm={3}
        className={clientTypeState === 'runner' || clientTypeState === 'wheel' ? 'hidden' : ''}
      >
        <FormControlLabel control={<Switch />} label="Schedule" className="hidden" />
      </Grid>
      <Grid
        item
        xs={6}
        sm={3}
        className={clientTypeState === 'runner' || clientTypeState === 'wheel' ? 'hidden' : ''}
      >
        <Tooltip title="YAML or JSON pillar data to override the default for the tgt.">
          <FormControlLabel
            control={
              <Switch
                checked={pillarVisible}
                onChange={(e) => {
                  setPillarVisible(e.target.checked);
                  updateURLParams('pillar', e.target.checked ? pillarValue : '');
                }}
              />
            }
            label="Pillar"
          />
        </Tooltip>
      </Grid>
      <Grid item xs={6} sm={3}>
        <FormControlLabel control={<Switch />} label="Save as Template" className="hidden" />
      </Grid>
      <Grid item xs={12} sm={3}>
        <Button
          variant="contained"
          color="primary"
          sx={{ marginLeft: 1 }}
          style={{ float: 'right' }}
          onClick={onRunClick}
        >
          Run
        </Button>
        <Button
          variant="contained"
          color="warning"
          sx={{ marginLeft: 1 }}
          style={{ float: 'right' }}
          onClick={onRunTestClick}
        >
          Test
        </Button>
      </Grid>
      <Grid item xs={6} sm={3} />
      <Grid item xs={6} sm={3}>
        {pillarVisible && (
          <CodeMirror
            value={pillarValue}
            extensions={[codeMirrorMode, codeMirrorMode.language]}
            onChange={(value) => handleCodeChange(value)}
            onBlur={() => updateURLParams('pillar', pillarValue)}
          />
        )}{' '}
      </Grid>
      <Grid item xs={6} sm={3} />
    </Grid>
  );
};

export default FormattedView;
