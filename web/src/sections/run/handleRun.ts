// handleRun.ts
import axios from 'axios';
import yaml from 'js-yaml';

export const handleRun = async (
  authToken: string,
  authSaltString: string,
  clientTypeState: string,
  asyncState: boolean,
  batchState: string,
  funState: string,
  tgtState: string,
  tgtTypeState: string,
  timeoutState: string,
  argumentsState: string,
  kwArgumentsState: string,
  pillarValue: string,
  pillarVisible: boolean,
  test?: boolean
): Promise<string> => {
  const parsedAuthSalt = JSON.parse(authSaltString);
  const { token } = parsedAuthSalt;

  let client = clientTypeState;
  if (asyncState) {
    client = `${client}_async`;
  } else if (batchState !== '') {
    client = `${client}_batch`;
  }

  const payload: Record<string, unknown> = {
    client,
    fun: funState,
  };

  if (batchState !== '') {
    payload.batch = batchState;
  }

  const allowedValues = ['local', 'local_async', 'ssh', 'local_batch'];
  if (allowedValues.includes(client)) {
    if (tgtTypeState === 'list') {
      payload.tgt = tgtState.split(',');
    } else {
      payload.tgt = tgtState;
    }
    payload.tgt_type = tgtTypeState;
  }

  const timeoutValue = parseInt(timeoutState, 10);
  if (!Number.isNaN(timeoutValue)) {
    payload.timeout = timeoutValue;
  }

  if (argumentsState.trim()) {
    payload.arg = argumentsState.split(',');
  }

  const kwarg: Record<string, string | boolean | unknown> = kwArgumentsState.trim()
    ? kwArgumentsState
        .split(',')
        .reduce((acc: Record<string, string | boolean | unknown>, curr) => {
          const [key, value] = curr.split('=');
          acc[key.trim()] = value.trim();
          return acc;
        }, {})
    : {};

  if (pillarValue !== '' && pillarVisible) {
    try {
      let parsedPillar;
      if (pillarValue.trim().startsWith('{') || pillarValue.trim().startsWith('[')) {
        parsedPillar = JSON.parse(pillarValue);
      } else {
        parsedPillar = yaml.load(pillarValue);
      }
      kwarg.pillar = parsedPillar;
    } catch (error) {
      console.error('Failed to parse pillar:', error);
      return 'Failed to parse pillar';
    }
  }

  if (test) {
    kwarg.test = true;
  }

  if (Object.keys(kwarg).length !== 0) {
    payload.kwarg = kwarg;
  }

  try {
    const response = await axios.post(`/api/v1/netapi/`, [payload], {
      headers: {
        Authorization: authToken,
        'X-Auth-Token': token,
      },
    });

    if (Object.keys(response.data.return).length === 1) {
      return JSON.stringify(response.data.return[0], null, 2);
    }
    return JSON.stringify(response.data.return, null, 2);
  } catch (err) {
    console.error(`Failed to run command on tgt ${tgtState}: `, err);
    return `Failed to run command on tgt ${tgtState}: ${err}`;
  }
};
