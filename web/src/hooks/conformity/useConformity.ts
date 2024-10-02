import axios from 'axios';
import { useState, useEffect } from 'react';

interface ConformityData {
  alter_time: string;
  true_count: number;
  false_count: number;
  changed_count: number;
  unchanged_count: number;
  success: boolean;
  id: string;
}

interface UseSaltConformity {
  alterTime: string;
  trueCount: number;
  falseCount: number;
  changedCount: number;
  unchangedCount: number;
  conformityId: string;
  success: boolean;
  isLoading: boolean;
  error: Error | null;
}

const useSaltConformity = (id: string): UseSaltConformity => {
  const [alterTime, setAlterTime] = useState<string>('');
  const [conformityId, setConformityIdTime] = useState<string>('');
  const [trueCount, setTrueCount] = useState<number>(0);
  const [falseCount, setFalseCount] = useState<number>(0);
  const [changedCount, setChangedCount] = useState<number>(0);
  const [unchangedCount, setUnchangedCount] = useState<number>(0);
  const [success, setSuccess] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const authToken = localStorage.getItem('authToken');

    const fetchConformityData = async () => {
      setIsLoading(true);
      try {
        const { data } = await axios.get<ConformityData>(`/api/v1/conformity/${id}`, {
          headers: {
            Authorization: `${authToken}`,
          },
        });
        setAlterTime(data.alter_time);
        setTrueCount(data.true_count);
        setFalseCount(data.false_count);
        setChangedCount(data.changed_count);
        setUnchangedCount(data.unchanged_count);
        setSuccess(data.success);
        setConformityIdTime(data.id);
      } catch (err) {
        setError(err as Error);
      }
      setIsLoading(false);
    };

    if (id) {
      fetchConformityData();
    }
  }, [alterTime, id, trueCount, falseCount, changedCount, unchangedCount, success]);

  return {
    alterTime,
    trueCount,
    falseCount,
    changedCount,
    unchangedCount,
    success,
    conformityId,
    isLoading,
    error,
  };
};

export default useSaltConformity;
