import { useState, useEffect } from 'react';

import { apiClient as axios } from 'src/api/client.ts';

interface VersionData {
  version: string;
}

const useVersionData = () => {
  const [versionData, setVersionData] = useState<VersionData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    const fetchVersionData = async () => {
      setIsLoading(true);
      try {
        const response = await axios.get<VersionData>('/version', { signal: controller.signal });
        setVersionData(response.data);
      } catch (err) {
        if (controller.signal.aborted) return;
        setError(err as Error);
      }
      if (!controller.signal.aborted) setIsLoading(false);
    };
    void fetchVersionData();
    return () => controller.abort();
  }, []);

  return {
    versionData,
    isLoading,
    error,
  };
};

export default useVersionData;
