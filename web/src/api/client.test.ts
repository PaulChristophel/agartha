import { AxiosError, AxiosHeaders } from 'axios';
import { it, expect, describe, afterEach, beforeEach } from 'vitest';

import { sessionStore } from './session.ts';
import { ApiError, apiClient } from './client.ts';

const originalAdapter = apiClient.defaults.adapter;

describe('apiClient', () => {
  beforeEach(() => {
    sessionStore.clear();
    sessionStore.setAuthToken('jwt-token');
    sessionStore.setAuthSalt({
      eauth: 'pam',
      expire: 9999999999,
      perms: [],
      start: 1,
      token: 'salt-token',
      user: 'alice',
    });
  });

  afterEach(() => {
    apiClient.defaults.adapter = originalAdapter;
    sessionStore.clear();
  });

  it('adds JWT and Salt authorization headers to netapi requests', async () => {
    let headers = new AxiosHeaders();
    apiClient.defaults.adapter = async (config) => {
      headers = AxiosHeaders.from(config.headers);
      return { config, data: {}, headers: {}, status: 200, statusText: 'OK' };
    };

    await apiClient.get('/api/v1/netapi/jobs');

    expect(headers.get('Authorization')).toBe('jwt-token');
    expect(headers.get('X-Auth-Token')).toBe('salt-token');
  });

  it('normalizes 401 errors and expires the local session', async () => {
    apiClient.defaults.adapter = async (config) => {
      throw new AxiosError('Unauthorized', AxiosError.ERR_BAD_REQUEST, config, undefined, {
        config,
        data: { message: 'expired' },
        headers: {},
        status: 401,
        statusText: 'Unauthorized',
      });
    };

    await expect(apiClient.get('/api/v1/secure/auth_user/user-1')).rejects.toMatchObject({
      name: 'ApiError',
      status: 401,
    });
    expect(sessionStore.getSnapshot().authToken).toBeNull();
  });

  it('marks cancelled requests consistently', () => {
    const error = new ApiError(new AxiosError('cancelled', AxiosError.ERR_CANCELED));
    expect(error.cancelled).toBe(true);
  });
});
