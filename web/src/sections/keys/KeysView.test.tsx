import { it, vi, expect, describe } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '@testing-library/react';

import KeysView from './KeysView.tsx';

const mocks = vi.hoisted(() => ({
  acceptKeys: vi.fn().mockResolvedValue(undefined),
  deleteKeys: vi.fn().mockResolvedValue(undefined),
  rejectKeys: vi.fn().mockResolvedValue(undefined),
}));

vi.mock('src/hooks/netapi/wheel/useKeys.ts', () => ({
  default: () => ({
    minions: [],
    minionsDenied: [],
    minionsPre: ['pending-minion'],
    minionsRejected: [],
    isLoading: false,
    error: null,
  }),
}));
vi.mock('src/hooks/netapi/wheel/useAcceptKeyDict.ts', () => ({
  default: () => ({ acceptKeys: mocks.acceptKeys }),
}));
vi.mock('src/hooks/netapi/wheel/useDeleteKeyDict.ts', () => ({
  default: () => ({ deleteKeys: mocks.deleteKeys }),
}));
vi.mock('src/hooks/netapi/wheel/useRejectKeyDict.ts', () => ({
  default: () => ({ rejectKeys: mocks.rejectKeys }),
}));

describe('KeysView', () => {
  it('submits selected pending keys through the accept action', async () => {
    const user = userEvent.setup();
    render(<KeysView reload={vi.fn()} />);

    await user.click(screen.getByRole('checkbox', { name: 'pending-minion' }));
    await user.click(screen.getByRole('button', { name: 'Accept' }));

    await waitFor(() =>
      expect(mocks.acceptKeys).toHaveBeenCalledWith({
        match: { minions: ['pending-minion'] },
        include_rejected: true,
        include_denied: true,
      })
    );
  });
});
