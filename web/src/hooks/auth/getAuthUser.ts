import { sessionStore } from 'src/api/session.ts';

import { AuthUser } from './authUser.ts';

export const getAuthUser = (): AuthUser | null => sessionStore.getSnapshot().authUser;
