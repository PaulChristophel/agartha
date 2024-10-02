import axios from 'axios';
import { useState, useCallback } from 'react';

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
      const authToken = localStorage.getItem('authToken');
      const response = await axios.get<RefreshStatus>('/api/v1/conformity/refresh', {
        headers: {
          Authorization: `${authToken}`,
        },
      });
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
        if (err.response.data && typeof err.response.data.message === 'string') {
          setMessage(err.response.data.message);
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
