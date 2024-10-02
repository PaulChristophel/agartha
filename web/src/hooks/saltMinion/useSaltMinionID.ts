import axios from 'axios';
import { useState, useEffect } from 'react';

interface MinionData {
  alter_time: string;
  grains: Record<string, unknown>;
  pillar: Record<string, unknown>;
  minion_id: string;
  id: string;
}

interface UseMinion {
  alterTime: string;
  grains: Record<string, unknown>;
  pillar: Record<string, unknown>;
  minionID: string;
  id: string;
  isLoading: boolean;
  error: Error | null;
}

const useCacheBankKey = (minionID: string): UseMinion => {
  const [alterTime, setAlterTime] = useState<string>('');
  const [grains, setGrains] = useState<Record<string, unknown>>({});
  const [pillar, setPillar] = useState<Record<string, unknown>>({});
  const [id, setID] = useState<string>('00000000-0000-0000-0000-000000000000');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const authToken = localStorage.getItem('authToken');

    const fetchCacheData = async () => {
      setIsLoading(true);
      try {
        const encodedMinionID = encodeURIComponent(minionID);
        const { data } = await axios.get<MinionData>(`/api/v1/salt_cache/${encodedMinionID}`, {
          headers: {
            Authorization: `${authToken}`,
          },
        });
        setAlterTime(data.alter_time);
        setGrains(data.grains);
        setPillar(data.pillar);
        setID(data.id);
      } catch (err) {
        setError(err as Error);
      }
      setIsLoading(false);
    };

    if (minionID) {
      fetchCacheData();
    }
  }, [minionID]);

  return {
    alterTime,
    grains,
    pillar,
    id,
    minionID,
    isLoading,
    error,
  };
};

export default useCacheBankKey;
