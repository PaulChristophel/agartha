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

const useCacheID = (id: string): UseCache => {
  const [alterTime, setAlterTime] = useState<string>('');
  const [cacheData, setCacheData] = useState<Record<string, unknown>>({});
  const [psqlKey, setPsqlKey] = useState<string>('');
  const [bank, setBank] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const authToken = localStorage.getItem('authToken');

    const fetchCacheData = async () => {
      setIsLoading(true);
      try {
        const { data } = await axios.get<CacheData>(`/api/v1/salt_cache/uuid/${id}`, {
          headers: {
            Authorization: `${authToken}`,
          },
        });
        setAlterTime(data.alter_time);
        setCacheData(data.data);
        setPsqlKey(data.psql_key);
        setBank(data.bank);
      } catch (err) {
        setError(err as Error);
      }
      setIsLoading(false);
    };

    if (id) {
      fetchCacheData();
    }
  }, [id]);

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

export default useCacheID;
