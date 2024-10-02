import axios from 'axios';
import { useState, useEffect } from 'react';

interface HighStateData {
  alter_time: string;
  full_ret: Record<string, unknown>; // Assuming full_ret is an object with unknown properties
  fun: string;
  id: string;
  jid: string;
  return: Record<string, unknown>; // Assuming return is an object with unknown properties
  success: boolean;
}

interface UseHighState {
  alterTime: string;
  fullRet: Record<string, unknown>;
  fun: string;
  returnId: string;
  returnJid: string;
  returnData: Record<string, unknown>;
  success: boolean;
  isLoading: boolean;
  error: Error | null;
}

const useHighState = (id: string, load_return: boolean, load_full_ret: boolean): UseHighState => {
  const [alterTime, setAlterTime] = useState<string>('');
  const [fullRet, setFullRet] = useState<Record<string, unknown>>({});
  const [fun, setFun] = useState<string>('');
  const [returnId, setHighStateId] = useState<string>('');
  const [returnJid, setHighStateJid] = useState<string>('');
  const [returnData, setHighStateData] = useState<Record<string, unknown>>({});
  const [success, setSuccess] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const authToken = localStorage.getItem('authToken');

    const fetchHighStateData = async () => {
      setIsLoading(true);
      try {
        const { data } = await axios.get<HighStateData>(
          `/api/v1/high_state/${id}?load_return=${load_return}&load_full_ret=${load_full_ret}`,
          {
            headers: {
              Authorization: `${authToken}`,
            },
          }
        );
        setAlterTime(data.alter_time);
        setFullRet(data.full_ret);
        setFun(data.fun);
        setHighStateId(data.id);
        setHighStateJid(data.jid);
        setHighStateData(data.return);
        setSuccess(data.success);
      } catch (err) {
        setError(err as Error);
      }
      setIsLoading(false);
    };

    if (id) {
      fetchHighStateData();
    }
  }, [id, load_return, load_full_ret]);

  return {
    alterTime,
    fullRet,
    fun,
    returnId,
    returnJid,
    returnData,
    success,
    isLoading,
    error,
  };
};

export default useHighState;
