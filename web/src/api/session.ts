import { useSyncExternalStore } from 'react';

import { AuthUser } from 'src/hooks/auth/authUser.ts';

export type SessionStatus = 'unknown' | 'authenticated' | 'anonymous';

export interface SessionSnapshot {
  status: SessionStatus;
  authUser: AuthUser | null;
}

const listeners = new Set<() => void>();
let snapshot: SessionSnapshot = { status: 'unknown', authUser: null };

function emitChange(next: SessionSnapshot) {
  snapshot = next;
  listeners.forEach((listener) => listener());
}

export const sessionStore = {
  getSnapshot: () => snapshot,
  getServerSnapshot: (): SessionSnapshot => ({ status: 'unknown', authUser: null }),
  subscribe(listener: () => void) {
    listeners.add(listener);
    return () => listeners.delete(listener);
  },
  setAuthenticated(authUser: AuthUser) {
    emitChange({ status: 'authenticated', authUser });
  },
  clear() {
    emitChange({ status: 'anonymous', authUser: null });
  },
  reset() {
    emitChange({ status: 'unknown', authUser: null });
  },
};

export function useSession(): SessionSnapshot {
  return useSyncExternalStore(
    sessionStore.subscribe,
    sessionStore.getSnapshot,
    sessionStore.getServerSnapshot
  );
}
