import axios from 'axios';
import { useState, useEffect } from 'react';

interface JidData {
  alter_time: string;
  jid: string;
  load: Record<string, unknown>;
}

interface UseJid {
  alterTime: string;
  jid: string;
  load: Record<string, unknown>;
  isLoading: boolean;
  error: Error | null;
}

const useJid = (jid: string): UseJid => {
  const [alterTime, setAlterTime] = useState<string>('');
  const [load, setLoad] = useState<Record<string, unknown>>({});
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const authToken = localStorage.getItem('authToken');

    const fetchJidData = async () => {
      setIsLoading(true);
      try {
        const { data } = await axios.get<JidData>(`/api/v1/jid/${jid}`, {
          headers: {
            Authorization: `${authToken}`,
          },
        });
        setAlterTime(data.alter_time);
        setLoad(data.load);
      } catch (err) {
        setError(err as Error);
      }
      setIsLoading(false);
    };

    if (jid) {
      fetchJidData();
    }
  }, [jid]);

  return {
    alterTime,
    jid,
    load,
    isLoading,
    error,
  };
};

export default useJid;
