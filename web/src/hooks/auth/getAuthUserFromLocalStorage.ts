import { AuthUser, AUTH_USER_KEY } from './authUser.ts';

export const getAuthUserFromLocalStorage = (): AuthUser | null => {
  const cachedSettings = localStorage.getItem(AUTH_USER_KEY);
  if (cachedSettings) {
    return JSON.parse(cachedSettings);
  }
  return null;
};
