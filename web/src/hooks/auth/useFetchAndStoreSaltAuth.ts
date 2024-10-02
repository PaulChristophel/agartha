import axios from 'axios';
import { useState } from 'react';

import { SaltAuth } from '../netapi/api/auth.ts';

export interface UseFetchAndStoreSaltAuth {
  eauth: string;
  expire: number;
  perms: string[];
  start: number;
  token: string;
  user: string;
  isLoading: boolean;
  status: number | null;
  error: Error | null;
  postSaltAuth: (password: string) => void;
}

const useFetchAndStoreSaltAuth = (): UseFetchAndStoreSaltAuth => {
  const [eauth, setEauth] = useState<string>('');
  const [expire, setExpire] = useState<number>(0);
  const [perms, setPerms] = useState<string[]>([]);
  const [start, setStart] = useState<number>(0);
  const [token, setToken] = useState<string>('');
  const [user, setUser] = useState<string>('');

  const [isLoading, setIsLoading] = useState(false);
  const [status, setStatus] = useState<number | null>(null);
  const [error, setError] = useState<Error | null>(null);

  const postSaltAuth = async (password: string) => {
    setIsLoading(true);
    try {
      const config = {
        endpoint: '/api/v1/netapi',
        password,
        logger: console,
      };

      const authClient = new SaltAuth(config);
      const authData = await authClient.login();

      setEauth(authData.eauth);
      setExpire(authData.expire);
      setPerms(authData.perms);
      setStart(authData.start);
      setToken(authData.token);
      setUser(authData.user);

      localStorage.setItem('authSalt', JSON.stringify(authData));

      setStatus(200);
      setError(null);
    } catch (err) {
      setError(err as Error);
      if (axios.isAxiosError(err) && err.response) {
        setStatus(err.response.status);
      } else {
        setStatus(null);
      }
    } finally {
      setIsLoading(false);
    }
  };

  return {
    eauth,
    expire,
    perms,
    start,
    token,
    user,
    isLoading,
    status,
    error,
    postSaltAuth,
  };
};

export default useFetchAndStoreSaltAuth;
