import { MemoryRouter } from 'react-router-dom';
import { it, vi, expect, describe } from 'vitest';
import userEvent from '@testing-library/user-event';
import { render, screen, waitFor } from '@testing-library/react';

import FormattedView from './FormattedView.tsx';

const mocks = vi.hoisted(() => ({ handleRun: vi.fn() }));

vi.mock('./handleRun.ts', () => ({ handleRun: mocks.handleRun }));

describe('FormattedView', () => {
  it('links an asynchronous submission to its job details', async () => {
    const user = userEvent.setup();
    mocks.handleRun.mockResolvedValue(
      JSON.stringify({ jid: '20260904120000000000', minions: ['alpha'] })
    );
    window.history.replaceState({}, '', '/run?async=true');
    try {
      render(
        <MemoryRouter>
          <FormattedView
            aSync
            batch=""
            fun="test.ping"
            tgt="alpha"
            tgtType="glob"
            timeout=""
            clientType="local"
            onClientTypeChange={vi.fn()}
            setOutput={vi.fn()}
            showInfoMessage={vi.fn()}
          />
        </MemoryRouter>
      );
      await user.click(screen.getByRole('button', { name: 'Run' }));
      expect(
        await screen.findByRole('link', { name: 'View job 20260904120000000000' })
      ).toHaveAttribute('href', '/job/20260904120000000000');
    } finally {
      window.history.replaceState({}, '', '/');
    }
  });
  it('submits a typed command without reading credentials from the component', async () => {
    const user = userEvent.setup();
    const setOutput = vi.fn();
    const showInfoMessage = vi.fn();
    mocks.handleRun.mockResolvedValue(JSON.stringify({ 'minion-1': true }));

    render(
      <FormattedView
        aSync={false}
        batch=""
        fun=""
        tgt=""
        tgtType="glob"
        timeout=""
        clientType="local"
        onClientTypeChange={vi.fn()}
        setOutput={setOutput}
        showInfoMessage={showInfoMessage}
      />
    );

    await user.type(screen.getByLabelText('Target'), 'minion-1');
    await user.type(screen.getByLabelText('Function'), 'test.ping');
    await user.click(screen.getByRole('button', { name: 'Run' }));

    await waitFor(() => expect(setOutput).toHaveBeenCalled());
    expect(showInfoMessage).toHaveBeenCalledWith('Command sent.');
    expect(mocks.handleRun).toHaveBeenCalledWith(
      'local',
      false,
      '',
      'test.ping',
      'minion-1',
      'glob',
      '',
      '',
      '',
      '',
      false
    );
  });
});
