import { useMemo } from 'react';
import userEvent from '@testing-library/user-event';
import { it, vi, expect, describe, beforeEach } from 'vitest';
import { Route, Routes, MemoryRouter } from 'react-router-dom';
import { render, screen, waitFor } from '@testing-library/react';

import MinionDetailsView from './MinionDetailsView.tsx';

const mocks = vi.hoisted(() => ({
  post: vi.fn().mockResolvedValue({ data: {} }),
  caches: [] as { bank: string; psql_key: string; id: string }[],
  grains: {},
  pillar: {},
}));

vi.mock('src/api/client.ts', () => ({ apiClient: { post: mocks.post } }));
vi.mock('src/hooks/saltCache/useCachePaginated.ts', () => ({
  default: function useMockCache({ bank, key }: { bank?: string; key?: string }) {
    const caches = useMemo(
      () =>
        mocks.caches.filter((cache) =>
          bank
            ? cache.bank.startsWith(bank)
            : cache.psql_key.startsWith((key || '').replace(/\*$/, ''))
        ),
      [bank, key]
    );
    return { caches, isLoading: false };
  },
}));
vi.mock('src/hooks/saltMinion/useSaltMinionID.ts', () => ({
  default: () => ({ grains: mocks.grains, pillar: mocks.pillar, isLoading: false }),
}));
vi.mock('./CodeView.tsx', () => ({
  default: ({ cacheID }: { cacheID: string }) => <div>cache row: {cacheID}</div>,
}));
vi.mock('./CommonDetails.tsx', () => ({ default: () => <div>common details</div> }));
vi.mock('./NetworkDetails.tsx', () => ({ default: () => <div>network details</div> }));
vi.mock('./HardwareDetails.tsx', () => ({ default: () => <div>hardware details</div> }));
vi.mock('./DataViewer.tsx', () => ({ default: () => <div>data viewer</div> }));
vi.mock('./ConformityDetailsView.tsx', () => ({ default: () => <div>conformity</div> }));

describe('MinionDetailsView', () => {
  beforeEach(() => {
    mocks.caches = [];
  });

  it('discovers migrated tabs, prefers migrated rows, and excludes neighboring minions', async () => {
    mocks.caches = [
      { bank: 'minions/minion-1', psql_key: 'gted_sssd', id: 'legacy' },
      { bank: 'gted_sssd', psql_key: 'minions/minion-1', id: 'migrated' },
      { bank: 'other_cache', psql_key: 'minions/minion-1', id: 'other' },
      { bank: 'minions/minion-10', psql_key: 'neighbor', id: 'neighbor' },
      { bank: 'grains', psql_key: 'minion-1', id: 'grains' },
      { bank: 'pillar', psql_key: 'minion-1:dev', id: 'pillar-dev' },
      { bank: 'minions/minion-1', psql_key: 'legacy_only', id: 'legacy-only' },
    ];
    render(
      <MemoryRouter initialEntries={['/minion/minion-1']}>
        <Routes>
          <Route path="/minion/:id" element={<MinionDetailsView />} />
        </Routes>
      </MemoryRouter>
    );

    expect(await screen.findAllByRole('tab', { name: 'gted_sssd' })).toHaveLength(1);
    expect(screen.getByRole('tab', { name: 'other_cache' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: 'legacy_only' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: 'Pillar:dev' })).toBeInTheDocument();
    expect(screen.queryByRole('tab', { name: 'neighbor' })).not.toBeInTheDocument();
    await userEvent.setup().click(screen.getByRole('tab', { name: 'gted_sssd' }));
    expect(screen.getByText('cache row: migrated')).toBeInTheDocument();
  });

  it('submits highstate for the routed minion through the API client', async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter initialEntries={['/minion/minion-1']}>
        <Routes>
          <Route path="/minion/:id" element={<MinionDetailsView />} />
        </Routes>
      </MemoryRouter>
    );

    await user.click(screen.getByRole('button', { name: 'Run highstate on minion' }));

    await waitFor(() =>
      expect(mocks.post).toHaveBeenCalledWith('/api/v1/netapi/', {
        client: 'local',
        fun: 'state.apply',
        tgt: 'minion-1',
        tgt_type: 'glob',
      })
    );
  });
});
