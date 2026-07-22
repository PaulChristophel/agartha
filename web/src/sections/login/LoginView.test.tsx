import { MemoryRouter } from 'react-router-dom';
import userEvent from '@testing-library/user-event';
import { it, vi, expect, describe, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';

import { sessionStore } from 'src/api/session.ts';

import LoginView from './LoginView.tsx';

const mocks = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  push: vi.fn(),
  fetchUser: vi.fn(),
  postSaltAuth: vi.fn(),
}));

vi.mock('src/api/client.ts', () => ({
  apiClient: { get: mocks.get, post: mocks.post },
}));
vi.mock('src/routes/hooks', () => ({ useRouter: () => ({ push: mocks.push }) }));
vi.mock('src/hooks/auth/fetchAndStoreAuthUser.ts', () => ({
  fetchAndStoreAuthUser: mocks.fetchUser,
}));
vi.mock('src/hooks/auth/useFetchAndStoreSaltAuth.ts', () => ({
  default: () => ({ postSaltAuth: mocks.postSaltAuth }),
}));

describe('LoginView', () => {
  beforeEach(() => {
    sessionStore.clear();
    mocks.get.mockResolvedValue({ data: { auth_methods: ['local', 'ldap'] } });
    mocks.post.mockResolvedValue({ data: { token: 'jwt-token' } });
    mocks.fetchUser.mockResolvedValue({ id: 'user-1' });
    mocks.postSaltAuth.mockResolvedValue(undefined);
  });

  it('creates the session and redirects after a successful login', async () => {
    const user = userEvent.setup();
    render(
      <MemoryRouter>
        <LoginView />
      </MemoryRouter>
    );

    await user.type(screen.getByLabelText('Username'), 'alice');
    await user.type(screen.getByLabelText('Password'), 'secret');
    await user.click(screen.getByRole('button', { name: 'Login' }));

    await waitFor(() => expect(mocks.push).toHaveBeenCalledWith('/'));
    expect(mocks.post).toHaveBeenCalledWith('/auth/token', {
      username: 'alice',
      password: 'secret',
      method: 'ldap',
    });
    expect(sessionStore.getSnapshot().authToken).toBe('jwt-token');
    expect(mocks.fetchUser).toHaveBeenCalledWith('jwt-token');
    expect(mocks.postSaltAuth).toHaveBeenCalledOnce();
  });
});
