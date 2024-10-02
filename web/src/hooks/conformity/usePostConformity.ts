import axios from 'axios';
import { useState } from 'react';

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
      const authToken = localStorage.getItem('authToken');
      const response = await axios.post<RefreshStatus>(
        '/api/v1/conformity/refresh',
        {},
        {
          headers: {
            Authorization: `${authToken}`,
          },
        }
      );
      setStatus(response.status);
      setMessage(response.data.message);
      setError(null);
    } catch (err) {
      setError(err as Error);
      setMessage(null);
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
