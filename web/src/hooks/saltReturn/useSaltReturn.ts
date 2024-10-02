import axios from 'axios';
import { useState, useEffect } from 'react';

interface ReturnData {
  alter_time: string;
  full_ret: Record<string, unknown>; // Assuming full_ret is an object with unknown properties
  fun: string;
  id: string;
  jid: string;
  return: Record<string, unknown>; // Assuming return is an object with unknown properties
  success: boolean;
}

interface UseSaltReturn {
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

const useSaltReturn = (
  jid: string,
  id: string,
  load_return: boolean,
  load_full_ret: boolean
): UseSaltReturn => {
  const [alterTime, setAlterTime] = useState<string>('');
  const [fullRet, setFullRet] = useState<Record<string, unknown>>({});
  const [fun, setFun] = useState<string>('');
  const [returnId, setReturnId] = useState<string>('');
  const [returnJid, setReturnJid] = useState<string>('');
  const [returnData, setReturnData] = useState<Record<string, unknown>>({});
  const [success, setSuccess] = useState<boolean>(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const authToken = localStorage.getItem('authToken');

    const fetchReturnData = async () => {
      setIsLoading(true);
      try {
        const { data } = await axios.get<ReturnData>(
          `/api/v1/salt_return/${jid}/${id}?load_return=${load_return}&load_full_ret=${load_full_ret}`,
          {
            headers: {
              Authorization: `${authToken}`,
            },
          }
        );
        setAlterTime(data.alter_time);
        setFullRet(data.full_ret);
        setFun(data.fun);
        setReturnId(data.id);
        setReturnJid(data.jid);
        setReturnData(data.return);
        setSuccess(data.success);
      } catch (err) {
        setError(err as Error);
      }
      setIsLoading(false);
    };

    if (jid && id) {
      fetchReturnData();
    }
  }, [jid, id, load_return, load_full_ret]);

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

export default useSaltReturn;
