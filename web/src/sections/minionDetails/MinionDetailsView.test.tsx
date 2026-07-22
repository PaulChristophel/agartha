import { it, vi, expect, describe } from 'vitest';
import userEvent from '@testing-library/user-event';
import { Route, Routes, MemoryRouter } from 'react-router-dom';
import { render, screen, waitFor } from '@testing-library/react';

import MinionDetailsView from './MinionDetailsView.tsx';

const mocks = vi.hoisted(() => ({
  post: vi.fn().mockResolvedValue({ data: {} }),
  caches: [],
  grains: {},
  pillar: {},
}));

vi.mock('src/api/client.ts', () => ({ apiClient: { post: mocks.post } }));
vi.mock('src/hooks/saltCache/useCachePaginated.ts', () => ({
  default: () => ({ caches: mocks.caches, isLoading: false }),
}));
vi.mock('src/hooks/saltMinion/useSaltMinionID.ts', () => ({
  default: () => ({ grains: mocks.grains, pillar: mocks.pillar, isLoading: false }),
}));
vi.mock('./CommonDetails.tsx', () => ({ default: () => <div>common details</div> }));
vi.mock('./NetworkDetails.tsx', () => ({ default: () => <div>network details</div> }));
vi.mock('./HardwareDetails.tsx', () => ({ default: () => <div>hardware details</div> }));
vi.mock('./DataViewer.tsx', () => ({ default: () => <div>data viewer</div> }));
vi.mock('./ConformityDetailsView.tsx', () => ({ default: () => <div>conformity</div> }));

describe('MinionDetailsView', () => {
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
