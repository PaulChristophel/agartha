import { useState, useCallback } from 'react';

import { apiClient as axios } from 'src/api/client.ts';
interface SaltAuthSession {
  eauth: string;
  expire: number;
  perms: string[];
  start: number;
  user: string;
}

export interface UseFetchAndStoreSaltAuth {
  eauth: string;
  expire: number;
  perms: string[];
  start: number;
  user: string;
  isLoading: boolean;
  status: number | null;
  error: Error | null;
  postSaltAuth: () => Promise<void>;
}

const useFetchAndStoreSaltAuth = (): UseFetchAndStoreSaltAuth => {
  const [eauth, setEauth] = useState<string>('');
  const [expire, setExpire] = useState<number>(0);
  const [perms, setPerms] = useState<string[]>([]);
  const [start, setStart] = useState<number>(0);
  const [user, setUser] = useState<string>('');

  const [isLoading, setIsLoading] = useState(false);
  const [status, setStatus] = useState<number | null>(null);
  const [error, setError] = useState<Error | null>(null);

  const postSaltAuth = useCallback(async () => {
    setIsLoading(true);
    try {
      const { data } = await axios.post<{ return: SaltAuthSession[] }>('/api/v1/netapi/login', {});
      const authData = data.return[0];

      setEauth(authData.eauth);
      setExpire(authData.expire);
      setPerms(authData.perms);
      setStart(authData.start);
      setUser(authData.user);

      setStatus(200);
      setError(null);
    } catch (err) {
      setError(err as Error);
      if (axios.isAxiosError(err) && err.response) {
        setStatus(err.response.status);
      } else {
        setStatus(null);
      }
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  return {
    eauth,
    expire,
    perms,
    start,
    user,
    isLoading,
    status,
    error,
    postSaltAuth,
  };
};

export default useFetchAndStoreSaltAuth;
