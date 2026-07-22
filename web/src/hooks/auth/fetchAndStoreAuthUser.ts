import { jwtDecode } from 'jwt-decode';

import { sessionStore } from 'src/api/session.ts';
import { apiClient as axios } from 'src/api/client.ts';

import { AuthUser } from './authUser.ts';

interface DecodedToken {
  user_id: string;
}

export const fetchAndStoreAuthUser = async (token: string): Promise<AuthUser | null> => {
  try {
    const decodedToken = jwtDecode<DecodedToken>(token);
    const userId = decodedToken.user_id;

    const cachedSettings = sessionStore.getSnapshot().authUser;
    if (cachedSettings) {
      return cachedSettings;
    }

    // Fetch user settings from the server
    const response = await axios.get<AuthUser>(`/api/v1/secure/auth_user/${userId}`);

    const authUser = response.data;

    sessionStore.setAuthUser(authUser);

    return authUser;
  } catch (error) {
    console.error('Failed to fetch user settings:', error);
    return null;
  }
};
