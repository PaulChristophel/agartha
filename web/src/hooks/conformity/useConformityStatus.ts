import { useState, useCallback } from 'react';

import { apiClient as axios } from 'src/api/client.ts';

interface UsePostConformity {
  isLoading: boolean;
  status: number | null;
  error: Error | null;
  isPending: boolean;
  message: string | null;
  getConformity: () => void;
}

interface RefreshStatus {
  status: string;
  message: string;
}

const useConformityStatus = (): UsePostConformity => {
  const [isLoading, setIsLoading] = useState(false);
  const [status, setStatus] = useState<number | null>(null);
  const [error, setError] = useState<Error | null>(null);
  const [isPending, setIsPending] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  const getConformity = useCallback(async () => {
    setIsLoading(true);
    try {
      const response = await axios.get<RefreshStatus>('/api/v1/conformity/refresh');
      setStatus(response.status);
      setMessage(response.data.message);
      setIsPending(response.data.status === 'pending');
      setError(null);
    } catch (err) {
      setError(err as Error);
      setMessage(null);
      setIsPending(false);
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
  }, []);

  return {
    isLoading,
    status,
    error,
    isPending,
    message,
    getConformity,
  };
};

export default useConformityStatus;
