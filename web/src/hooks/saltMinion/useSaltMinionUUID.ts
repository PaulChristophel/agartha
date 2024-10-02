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

const useMinionID = (id: string): UseMinion => {
  const [alterTime, setAlterTime] = useState<string>('');
  const [grains, setGrains] = useState<Record<string, unknown>>({});
  const [pillar, setPillar] = useState<Record<string, unknown>>({});
  const [minionID, setMinionID] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const authToken = localStorage.getItem('authToken');

    const fetchMinionData = async () => {
      setIsLoading(true);
      try {
        const { data } = await axios.get<MinionData>(`/api/v1/salt_minion/uuid/${id}`, {
          headers: {
            Authorization: `${authToken}`,
          },
        });
        setAlterTime(data.alter_time);
        setGrains(data.grains);
        setPillar(data.pillar);
        setMinionID(data.minion_id);
      } catch (err) {
        setError(err as Error);
      }
      setIsLoading(false);
    };

    if (id) {
      fetchMinionData();
    }
  }, [id]);

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

export default useMinionID;
