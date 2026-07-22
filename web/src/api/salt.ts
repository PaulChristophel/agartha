import { apiClient } from './client.ts';

interface SaltNetApiResponse<T> {
  return: Array<{
    data: {
      return: T;
    };
  }>;
}

export async function executeWheel<TRequest extends object, TResponse>(
  request: TRequest
): Promise<TResponse> {
  const { data } = await apiClient.post<SaltNetApiResponse<TResponse>>('/api/v1/netapi/', {
    client: 'wheel',
    ...request,
  });

  return data.return[0].data.return;
}
