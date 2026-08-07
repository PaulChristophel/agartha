import { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { it, expect, describe, afterEach, beforeEach } from 'vitest';

import { sessionStore } from './session.ts';
import { ApiError, apiClient } from './client.ts';

const originalAdapter = apiClient.defaults.adapter;

describe('apiClient', () => {
  beforeEach(() => {
    sessionStore.setAuthenticated({
      id: 1,
      username: 'alice',
      date_joined: '',
      first_name: 'Alice',
      last_name: 'Example',
      email: 'alice@example.com',
      is_active: true,
      is_staff: false,
      is_superuser: false,
      last_login: '',
    });
  });

  afterEach(() => {
    apiClient.defaults.adapter = originalAdapter;
    sessionStore.clear();
  });

  it('uses credentials without adding JavaScript-readable authorization headers', async () => {
    let requestConfig: InternalAxiosRequestConfig | undefined;
    apiClient.defaults.adapter = async (config) => {
      requestConfig = config;
      return { config, data: {}, headers: {}, status: 200, statusText: 'OK' };
    };

    await apiClient.get('/api/v1/netapi/jobs');

    expect(requestConfig).toBeDefined();
    expect(requestConfig!.withCredentials).toBe(true);
    expect(requestConfig!.headers.Authorization).toBeUndefined();
    expect(requestConfig!.headers['X-Auth-Token']).toBeUndefined();
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
    expect(sessionStore.getSnapshot().status).toBe('anonymous');
  });

  it('marks cancelled requests consistently', () => {
    const error = new ApiError(new AxiosError('cancelled', AxiosError.ERR_CANCELED));
    expect(error.cancelled).toBe(true);
  });
});
