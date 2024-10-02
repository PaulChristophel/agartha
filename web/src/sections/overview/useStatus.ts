// useStatus.ts
import axios from 'axios';
import { useState, useEffect } from 'react';

const useStatus = (authToken: string, authSaltString: string) => {
  const [status, setStatus] = useState<string>('Loading...');
  const [isStatusLoading, setIsStatusLoading] = useState<boolean>(true);

  const parsedAuthSalt = JSON.parse(authSaltString);
  const { token } = parsedAuthSalt;

  useEffect(() => {
    const fetchStatus = async () => {
      setIsStatusLoading(true);
      try {
        const response = await axios.get('/api/v1/netapi', {
          headers: {
            Authorization: authToken,
            'X-Auth-Token': token,
          },
        });
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

    if (authToken && token) {
      fetchStatus();
    } else {
      setStatus('Error');
      setIsStatusLoading(false);
    }
  }, [authToken, token]);

  return { status, isStatusLoading };
};

export default useStatus;
