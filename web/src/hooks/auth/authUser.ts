export const AUTH_USER_KEY = 'authUser';

export interface AuthUser {
  date_joined: string;
  email: string;
  first_name: string;
  id: number;
  is_staff: boolean;
  is_active: boolean;
  is_superuser: boolean;
  last_login: string;
  last_name: string;
  password: string;
  username: string;
}
