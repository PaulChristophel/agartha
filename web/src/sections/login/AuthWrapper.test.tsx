import { render, screen } from '@testing-library/react';
import { it, vi, expect, describe, beforeEach } from 'vitest';
import { Route, Routes, MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import { sessionStore } from 'src/api/session.ts';

import AuthWrapper from './AuthWrapper.tsx';

const mocks = vi.hoisted(() => ({ getSession: vi.fn() }));

vi.mock('src/api/auth.ts', () => ({ getSession: mocks.getSession }));

describe('AuthWrapper', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    sessionStore.reset();
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  });

  const renderRoute = () =>
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/protected']}>
          <Routes>
            <Route path="/login" element={<div>login page</div>} />
            <Route
              path="/protected"
              element={
                <AuthWrapper>
                  <div>protected page</div>
                </AuthWrapper>
              }
            />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    );

  it('renders protected content after the cookie session is validated', async () => {
    mocks.getSession.mockResolvedValue({
      id: 1,
      username: 'alice',
      date_joined: '',
      first_name: 'Alice',
      last_name: 'Example',
      email: 'alice@example.com',
      is_active: true,
      is_staff: false,
      is_superuser: false,
      last_login: '',
    });

    renderRoute();

    expect(await screen.findByText('protected page')).toBeInTheDocument();
    expect(sessionStore.getSnapshot().status).toBe('authenticated');
  });

  it('redirects to login when the cookie session has expired', async () => {
    mocks.getSession.mockRejectedValue(new Error('Unauthorized'));

    renderRoute();

    expect(await screen.findByText('login page')).toBeInTheDocument();
  });

  it('cancels session validation when the protected view unmounts', () => {
    let requestSignal: AbortSignal | undefined;
    mocks.getSession.mockImplementation((signal: AbortSignal) => {
      requestSignal = signal;
      return new Promise(() => undefined);
    });

    const view = renderRoute();
    view.unmount();

    expect(requestSignal?.aborted).toBe(true);
  });
});
