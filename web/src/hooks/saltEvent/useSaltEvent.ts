import { useState, useEffect } from 'react';

import { apiClient as axios } from 'src/api/client.ts';

interface EventData {
  alter_time: string;
  data: Record<string, unknown>;
  id: number;
  master_id: string;
  tag: string;
}

interface UseEvent {
  alterTime: string;
  eventData: Record<string, unknown>;
  id: number;
  masterID: string;
  tag: string;
  isLoading: boolean;
  error: Error | null;
}

const useEvent = (id: number): UseEvent => {
  const [alterTime, setAlterTime] = useState<string>('');
  const [eventData, setEventData] = useState<Record<string, unknown>>({});
  const [masterID, setMasterID] = useState<string>('');
  const [tag, setTag] = useState<string>('');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchEventData = async () => {
      setIsLoading(true);
      try {
        const { data } = await axios.get<EventData>(`/api/v1/salt_event/${id}`);
        setAlterTime(data.alter_time);
        setEventData(data.data);
        setMasterID(data.master_id);
        setTag(data.tag);
      } catch (err) {
        setError(err as Error);
      }
      setIsLoading(false);
    };

    if (id) {
      fetchEventData();
    }
  }, [id]);

  return {
    alterTime,
    eventData,
    id,
    masterID,
    tag,
    isLoading,
    error,
  };
};

export default useEvent;
