import axios from 'axios';
import { useState, useEffect } from 'react';

interface CacheData {
  alter_time: string;
  data: Record<string, unknown>;
  bank: string;
  psql_key: string;
  id: string;
}

interface UseCache {
  alterTime: string;
  cacheData: Record<string, unknown>;
  bank: string;
  psqlKey: string;
  id: string;
  isLoading: boolean;
  error: Error | null;
}

// const decodeBase64 = (base64: string | null): Record<string, unknown> => {
//   if (!base64) return {};
//   try {
//     const decodedString = atob(base64);
//     return JSON.parse(decodedString);
//   } catch (e) {
//     console.error('Failed to decode base64 string:', e);
//     return {};
//   }
// };

const useCacheBankKey = (bank: string, psqlKey: string): UseCache => {
  const [alterTime, setAlterTime] = useState<string>('');
  const [cacheData, setCacheData] = useState<Record<string, unknown>>({});
  const [id, setID] = useState<string>('00000000-0000-0000-0000-000000000000');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const authToken = localStorage.getItem('authToken');

    const fetchCacheData = async () => {
      setIsLoading(true);
      try {
        const encodedBank = encodeURIComponent(bank);
        const encodedPsqlKey = encodeURIComponent(psqlKey);
        const { data } = await axios.get<CacheData>(
          `/api/v1/salt_cache/${encodedBank}/${encodedPsqlKey}`,
          {
            headers: {
              Authorization: `${authToken}`,
            },
          }
        );
        setAlterTime(data.alter_time);
        setCacheData(data.data);
        setID(data.id);
      } catch (err) {
        setError(err as Error);
      } finally {
        setIsLoading(false);
      }
    };

    if (bank && psqlKey) {
      fetchCacheData();
    }
  }, [bank, psqlKey]);

  return {
    alterTime,
    cacheData,
    id,
    psqlKey,
    bank,
    isLoading,
    error,
  };
};

export default useCacheBankKey;
