import { it, vi, expect, describe, beforeEach } from 'vitest';
import { act, waitFor, renderHook } from '@testing-library/react';

import useFetchFunKeys from './useFetchFunKeys.ts';

const api = vi.hoisted(() => ({
  get: vi.fn(),
  isAxiosError: vi.fn(),
}));

vi.mock('src/api/client.ts', () => ({
  apiClient: api,
}));

describe('useFetchFunKeys', () => {
  beforeEach(() => {
    api.isAxiosError.mockImplementation(
      (error: unknown) => typeof error === 'object' && error !== null && 'response' in error
    );
  });

  it('can reload the function list on demand', async () => {
    api.get
      .mockResolvedValueOnce({ data: { results: ['test.version'] } })
      .mockResolvedValueOnce({ data: { results: ['test.version', 'state.apply'] } });
    const { result } = renderHook(() => useFetchFunKeys('', 1));

    await waitFor(() => expect(result.current.funKeys).toEqual(['test.version']));
    act(() => result.current.refresh());
    await waitFor(() => expect(result.current.funKeys).toEqual(['test.version', 'state.apply']));

    expect(api.get).toHaveBeenCalledTimes(2);
  });

  it('reports an empty list without treating the API 404 as a failure', async () => {
    api.get.mockRejectedValue({ response: { status: 404 } });
    const { result } = renderHook(() => useFetchFunKeys('', 1));

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.isEmpty).toBe(true);
    expect(result.current.error).toBeNull();
  });
});
