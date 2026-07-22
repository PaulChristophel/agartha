import { useState, useEffect } from 'react';

import { useSession } from 'src/api/session.ts';
// useStatus.ts
import { apiClient as axios } from 'src/api/client.ts';

const useStatus = () => {
  const [status, setStatus] = useState<string>('Loading...');
  const [isStatusLoading, setIsStatusLoading] = useState<boolean>(true);
  const { authToken, authSalt } = useSession();

  useEffect(() => {
    const fetchStatus = async () => {
      setIsStatusLoading(true);
      try {
        const response = await axios.get('/api/v1/netapi');
        if (response.data && response.data.return === 'Welcome') {
          setStatus('OK');
        } else {
          setStatus('Error');
        }
      } catch (error) {
        console.error(`Failed to fetch status: `, error);
        setStatus('Error');
      } finally {
        setIsStatusLoading(false);
      }
    };

    if (authToken && authSalt) {
      fetchStatus();
    } else {
      setStatus('Restricted');
      setIsStatusLoading(false);
    }
  }, [authToken, authSalt]);

  return { status, isStatusLoading };
};

export default useStatus;
