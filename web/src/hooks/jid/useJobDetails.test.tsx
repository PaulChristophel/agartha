import type { PropsWithChildren } from 'react';
import { act, renderHook } from '@testing-library/react';
import { it, vi, expect, describe, afterEach } from 'vitest';
import { QueryClient, focusManager, QueryClientProvider } from '@tanstack/react-query';

import useJobDetails from './useJobDetails.ts';

const api = vi.hoisted(() => ({ get: vi.fn() }));
vi.mock('src/api/client.ts', () => ({
  apiClient: api,
  isApiError: (error: { status?: number }) => Boolean(error.status),
}));

function wrapper() {
  const client = new QueryClient({ defaultOptions: { queries: { gcTime: 0 } } });
  return function Provider({ children }: PropsWithChildren) {
    return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
  };
}

async function advance(milliseconds: number) {
  await act(async () => {
    await vi.advanceTimersByTimeAsync(milliseconds);
  });
}

describe('job detail fetching', () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
    focusManager.setFocused(undefined);
  });
  it('polls only while visible, stops after two minutes, and permits manual refresh', async () => {
    vi.useFakeTimers();
    api.get.mockResolvedValue({ data: { jid: 'job-1' } });
    const { result, unmount } = renderHook(() => useJobDetails('job-1'), { wrapper: wrapper() });
    await advance(10);
    expect(api.get).toHaveBeenCalledTimes(1);
    await advance(5_000);
    expect(api.get).toHaveBeenCalledTimes(2);
    focusManager.setFocused(false);
    await advance(10_000);
    expect(api.get).toHaveBeenCalledTimes(2);
    focusManager.setFocused(true);
    await advance(110_000);
    expect(result.current.autoRefresh).toBe(false);
    const calls = api.get.mock.calls.length;
    await advance(20_000);
    expect(api.get).toHaveBeenCalledTimes(calls);
    await act(async () => {
      await result.current.refetch();
    });
    expect(api.get).toHaveBeenCalledTimes(calls + 1);
    unmount();
    await advance(10_000);
    expect(api.get).toHaveBeenCalledTimes(calls + 1);
  });
  it('retries a newly submitted job that has not reached the database', async () => {
    vi.useFakeTimers();
    api.get.mockRejectedValueOnce({ status: 404 }).mockResolvedValue({ data: { jid: 'new-job' } });
    const submittedAt = Date.now();
    const { result, unmount } = renderHook(() => useJobDetails('new-job', submittedAt), {
      wrapper: wrapper(),
    });
    await advance(3_100);
    expect(result.current.data?.jid).toBe('new-job');
    expect(api.get).toHaveBeenCalledTimes(2);
    unmount();
  });
  it('does not retry ordinary missing jobs', async () => {
    vi.useFakeTimers();
    api.get.mockRejectedValue({ status: 404 });
    const { result, unmount } = renderHook(() => useJobDetails('missing'), { wrapper: wrapper() });
    await advance(20_000);
    expect(result.current.isError).toBe(true);
    expect(api.get).toHaveBeenCalledTimes(1);
    unmount();
  });
});
