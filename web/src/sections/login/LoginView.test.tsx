import { MemoryRouter } from 'react-router-dom';
import userEvent from '@testing-library/user-event';
import { it, vi, expect, describe, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import { sessionStore } from 'src/api/session.ts';

import LoginView from './LoginView.tsx';

const mocks = vi.hoisted(() => ({
  getAuthMethods: vi.fn(),
  getSession: vi.fn(),
  login: vi.fn(),
  push: vi.fn(),
  postSaltAuth: vi.fn(),
}));

vi.mock('src/api/auth.ts', () => ({
  getAuthMethods: mocks.getAuthMethods,
  getSession: mocks.getSession,
  login: mocks.login,
}));
vi.mock('src/routes/hooks', () => ({ useRouter: () => ({ push: mocks.push }) }));
vi.mock('src/hooks/auth/useFetchAndStoreSaltAuth.ts', () => ({
  default: () => ({ postSaltAuth: mocks.postSaltAuth }),
}));

describe('LoginView', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    sessionStore.clear();
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    mocks.getAuthMethods.mockResolvedValue(['local', 'ldap']);
    mocks.login.mockResolvedValue(undefined);
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
    mocks.postSaltAuth.mockResolvedValue(undefined);
  });

  const renderLogin = () =>
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <LoginView />
        </MemoryRouter>
      </QueryClientProvider>
    );

  it('creates the session and redirects after a successful login', async () => {
    const user = userEvent.setup();
    renderLogin();

    await user.type(screen.getByLabelText('Username'), 'alice');
    await user.type(screen.getByLabelText('Password'), 'secret');
    await user.click(screen.getByRole('button', { name: 'Login' }));

    await waitFor(() => expect(mocks.push).toHaveBeenCalledWith('/'));
    expect(mocks.login).toHaveBeenCalledWith({
      username: 'alice',
      password: 'secret',
      method: 'ldap',
    });
    expect(sessionStore.getSnapshot().status).toBe('authenticated');
    expect(mocks.postSaltAuth).toHaveBeenCalledOnce();
  });

  it('shows a consistent error when authentication fails', async () => {
    const user = userEvent.setup();
    mocks.login.mockRejectedValue(new Error('Unauthorized'));
    renderLogin();

    await user.type(screen.getByLabelText('Username'), 'alice');
    await user.type(screen.getByLabelText('Password'), 'wrong');
    await user.click(screen.getByRole('button', { name: 'Login' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('Unable to sign in');
    expect(mocks.push).not.toHaveBeenCalled();
  });
});
