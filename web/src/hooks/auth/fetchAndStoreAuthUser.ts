import axios from 'axios';
import { jwtDecode } from 'jwt-decode';

import { AuthUser, AUTH_USER_KEY } from './authUser.ts';

interface DecodedToken {
  user_id: string;
}

export const fetchAndStoreAuthUser = async (token: string): Promise<AuthUser | null> => {
  try {
    const decodedToken = jwtDecode<DecodedToken>(token);
    const userId = decodedToken.user_id;

    // Check if the user settings are already stored in localStorage
    const cachedSettings = localStorage.getItem(AUTH_USER_KEY);
    if (cachedSettings) {
      return JSON.parse(cachedSettings);
    }

    // Fetch user settings from the server
    const response = await axios.get<AuthUser>(`/api/v1/secure/auth_user/${userId}`, {
      headers: {
        Authorization: `${token}`,
      },
    });

    const authUser = response.data;

    // Store user settings in localStorage
    localStorage.setItem(AUTH_USER_KEY, JSON.stringify(authUser));

    return authUser;
  } catch (error) {
    console.error('Failed to fetch user settings:', error);
    return null;
  }
};
