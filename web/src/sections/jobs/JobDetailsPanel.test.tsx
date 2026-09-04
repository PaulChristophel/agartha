import { it, vi, expect, afterEach } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, cleanup } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import JobDetailsPanel from './JobDetailsPanel.tsx';

const api = vi.hoisted(() => ({ get: vi.fn() }));
vi.mock('src/api/client.ts', () => ({ apiClient: api, isApiError: () => false }));
afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

it('keeps unknown counts distinct, filters returns locally, and preserves results on refresh errors', async () => {
  const user = userEvent.setup();
  api.get.mockResolvedValue({
    data: {
      jid: 'job-1',
      load: { fun: 'test.ping' },
      returns: [
        { id: 'alpha', success: true, alter_time: null, return: true, full_ret: {} },
        { id: 'beta', success: false, alter_time: null, return: 'failed', full_ret: {} },
      ],
      targeted_count: null,
      pending_count: null,
      returned_count: 2,
      successful_count: 1,
      failed_count: 1,
      started_at: null,
      last_return_at: null,
    },
  });
  const client = new QueryClient({ defaultOptions: { queries: { gcTime: 0 } } });
  render(
    <QueryClientProvider client={client}>
      <JobDetailsPanel jid="job-1" />
    </QueryClientProvider>
  );
  expect(await screen.findByText('Targeted: Unknown')).toBeInTheDocument();
  expect(screen.getByText('No return received: Unknown')).toBeInTheDocument();
  expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
  await user.type(screen.getByLabelText('Search minion ID'), 'beta');
  expect(screen.queryByText('alpha')).not.toBeInTheDocument();
  expect(screen.getByText('beta')).toBeInTheDocument();
  await user.click(screen.getByText('Return', { exact: true }));
  expect(await screen.findByText('"failed"')).toBeInTheDocument();
  expect(api.get).toHaveBeenCalledTimes(1);
  api.get.mockRejectedValue(new Error('offline'));
  await user.click(screen.getByRole('button', { name: 'Refresh' }));
  expect(
    await screen.findByText('Could not refresh. Showing previously loaded results.')
  ).toBeInTheDocument();
  expect(screen.getByText('beta')).toBeInTheDocument();
});
