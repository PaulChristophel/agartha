import { act, renderHook } from '@testing-library/react';
import { it, vi, expect, describe, afterEach } from 'vitest';

import useRefreshKeys from './useRefreshKeys.ts';

const api = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
}));

vi.mock('src/api/client.ts', () => ({
  apiClient: api,
}));

describe('useRefreshKeys', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it('waits for the refresh to finish before reloading dropdown data', async () => {
    vi.useFakeTimers();
    api.post.mockResolvedValue({ data: { status: 'success' } });
    api.get
      .mockResolvedValueOnce({ data: { status: 'pending' } })
      .mockResolvedValueOnce({ data: { status: 'available' } });
    const { result } = renderHook(() => useRefreshKeys());

    let refresh: Promise<void>;
    act(() => {
      refresh = result.current.refreshKeys();
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(500);
      await refresh;
    });

    expect(api.post).toHaveBeenCalledWith(
      '/api/v1/salt_minion/keys/refresh',
      {},
      expect.objectContaining({ signal: expect.any(AbortSignal) })
    );
    expect(api.get).toHaveBeenCalledTimes(2);
    expect(result.current.revision).toBe(1);
    expect(result.current.message).toBe('Dropdown lists refreshed');
  });

  it('surfaces a database refresh failure without reloading the lists', async () => {
    vi.useFakeTimers();
    api.post.mockResolvedValue({ data: { status: 'success' } });
    api.get.mockResolvedValue({
      data: { status: 'failed', message: 'Failed to refresh the grains dropdown list' },
    });
    const { result } = renderHook(() => useRefreshKeys());

    let refresh: Promise<void>;
    act(() => {
      refresh = result.current.refreshKeys();
    });
    await act(async () => {
      await vi.advanceTimersByTimeAsync(250);
      await refresh;
    });

    expect(result.current.revision).toBe(0);
    expect(result.current.error).toBe('Failed to refresh the grains dropdown list');
  });
});
