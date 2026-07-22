import axios, { AxiosError, AxiosHeaders, AxiosResponse, InternalAxiosRequestConfig } from 'axios';

import { sessionStore } from './session.ts';

export class ApiError<T = unknown> extends Error {
  readonly status: number | null;
  readonly code: string | null;
  readonly details: T | null;
  readonly cancelled: boolean;
  readonly response: AxiosResponse<T> | undefined;

  constructor(error: AxiosError<T>) {
    super(error.message || 'Request failed', { cause: error });
    this.name = 'ApiError';
    this.status = error.response?.status ?? null;
    this.code = error.code ?? null;
    this.details = error.response?.data ?? null;
    this.cancelled = axios.isCancel(error) || error.code === AxiosError.ERR_CANCELED;
    this.response = error.response;
  }
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}

const addSessionHeaders = (config: InternalAxiosRequestConfig) => {
  const { authToken, authSalt } = sessionStore.getSnapshot();
  const headers = AxiosHeaders.from(config.headers);

  if (authToken && !headers.has('Authorization')) {
    headers.set('Authorization', authToken);
  }
  if (authSalt?.token && config.url?.startsWith('/api/v1/netapi') && !headers.has('X-Auth-Token')) {
    headers.set('X-Auth-Token', authSalt.token);
  }

  config.headers = headers;
  return config;
};

export const apiClient = Object.assign(
  axios.create({
    withCredentials: true,
  }),
  { isAxiosError: isApiError }
);

apiClient.interceptors.request.use(addSessionHeaders);
apiClient.interceptors.response.use(
  (response) => response,
  (error: unknown) => {
    if (!axios.isAxiosError(error)) {
      return Promise.reject(error);
    }

    const apiError = new ApiError(error);
    if (apiError.status === 401 && sessionStore.getSnapshot().authToken) {
      sessionStore.clear();
    }
    return Promise.reject(apiError);
  }
);
