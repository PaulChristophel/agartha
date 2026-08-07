import { AuthUser } from 'src/hooks/auth/authUser.ts';

import { apiClient } from './client.ts';

export interface LoginRequest {
  username: string;
  password: string;
  method: string;
}

export interface AuthMethodsResponse {
  auth_methods: string[];
}

export async function getAuthMethods(signal?: AbortSignal): Promise<string[]> {
  const { data } = await apiClient.get<AuthMethodsResponse>('/auth/method', { signal });
  return data.auth_methods;
}

export async function login(credentials: LoginRequest): Promise<void> {
  await apiClient.post('/auth/token', credentials);
}

export async function getSession(signal?: AbortSignal): Promise<AuthUser> {
  const { data } = await apiClient.get<AuthUser>('/auth/session', { signal });
  return data;
}

export async function logout(): Promise<void> {
  await apiClient.post('/auth/logout');
}
