import { useState } from 'react';

import { apiClient as axios } from 'src/api/client.ts';

interface UsePostConformity {
  isLoading: boolean;
  status: number | null;
  error: Error | null;
  message: string | null;
  postConformity: () => void;
}

interface RefreshStatus {
  status: string;
  message: string;
}

const usePostConformity = (): UsePostConformity => {
  const [isLoading, setIsLoading] = useState(false);
  const [status, setStatus] = useState<number | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const postConformity = async () => {
    setIsLoading(true);
    try {
      const response = await axios.post<RefreshStatus>('/api/v1/conformity/refresh', {});
      setStatus(response.status);
      setMessage(response.data.message);
      setError(null);
    } catch (err) {
      setError(err as Error);
      setMessage(null);
      if (axios.isAxiosError(err) && err.response) {
        setStatus(err.response.status);
        const responseData = err.response.data as Partial<RefreshStatus>;
        if (typeof responseData.message === 'string') {
          setMessage(responseData.message);
        }
      } else {
        setStatus(null);
      }
    } finally {
      setIsLoading(false);
    }
  };

  return {
    isLoading,
    status,
    error,
    message,
    postConformity,
  };
};

export default usePostConformity;
