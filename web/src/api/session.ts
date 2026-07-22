import { jwtDecode } from 'jwt-decode';
import { useSyncExternalStore } from 'react';

import { AuthUser, AUTH_USER_KEY } from 'src/hooks/auth/authUser.ts';

const AUTH_TOKEN_KEY = 'authToken';
const AUTH_SALT_KEY = 'authSalt';

export interface SaltAuthSession {
  eauth: string;
  expire: number;
  perms: string[];
  start: number;
  token: string;
  user: string;
}

export interface SessionSnapshot {
  authToken: string | null;
  authSalt: SaltAuthSession | null;
  authUser: AuthUser | null;
}

interface DecodedToken {
  exp?: number;
}

const emptySession: SessionSnapshot = {
  authToken: null,
  authSalt: null,
  authUser: null,
};

const listeners = new Set<() => void>();
let snapshot = readSession();

function readJson<T>(key: string): T | null {
  const value = window.localStorage.getItem(key);
  if (!value) return null;

  try {
    return JSON.parse(value) as T;
  } catch {
    window.localStorage.removeItem(key);
    return null;
  }
}

function readSession(): SessionSnapshot {
  if (typeof window === 'undefined') return emptySession;

  return {
    authToken: window.localStorage.getItem(AUTH_TOKEN_KEY),
    authSalt: readJson<SaltAuthSession>(AUTH_SALT_KEY),
    authUser: readJson<AuthUser>(AUTH_USER_KEY),
  };
}

function emitChange() {
  snapshot = readSession();
  listeners.forEach((listener) => listener());
}

function setJson(key: string, value: unknown | null) {
  if (value === null) {
    window.localStorage.removeItem(key);
  } else {
    window.localStorage.setItem(key, JSON.stringify(value));
  }
  emitChange();
}

export const sessionStore = {
  getSnapshot: () => snapshot,
  getServerSnapshot: () => emptySession,
  subscribe(listener: () => void) {
    listeners.add(listener);
    return () => listeners.delete(listener);
  },
  setAuthToken(token: string) {
    window.localStorage.setItem(AUTH_TOKEN_KEY, token);
    window.localStorage.removeItem(AUTH_USER_KEY);
    window.localStorage.removeItem(AUTH_SALT_KEY);
    emitChange();
  },
  setAuthSalt(authSalt: SaltAuthSession | null) {
    setJson(AUTH_SALT_KEY, authSalt);
  },
  setAuthUser(authUser: AuthUser | null) {
    setJson(AUTH_USER_KEY, authUser);
  },
  clear() {
    window.localStorage.removeItem(AUTH_TOKEN_KEY);
    window.localStorage.removeItem(AUTH_SALT_KEY);
    window.localStorage.removeItem(AUTH_USER_KEY);
    emitChange();
  },
};

export function useSession(): SessionSnapshot {
  return useSyncExternalStore(
    sessionStore.subscribe,
    sessionStore.getSnapshot,
    sessionStore.getServerSnapshot
  );
}

export function isJwtExpired(token: string): boolean {
  try {
    const { exp } = jwtDecode<DecodedToken>(token);
    return !exp || exp * 1000 <= Date.now();
  } catch {
    return true;
  }
}

export function isSaltAuthExpired(authSalt: SaltAuthSession): boolean {
  return authSalt.expire * 1000 <= Date.now();
}

if (typeof window !== 'undefined') {
  window.addEventListener('storage', emitChange);
}
