import axios from 'axios';
import { useState, useEffect } from 'react';

interface VersionData {
  version: string;
}

const useVersionData = () => {
  const [versionData, setVersionData] = useState<VersionData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchVersionData = async () => {
      setIsLoading(true);
      try {
        const response = await axios.get('/version');
        setVersionData(response.data);
      } catch (err) {
        setError(err as Error);
      }
      setIsLoading(false);
    };
    fetchVersionData();
  }, []);

  return {
    versionData,
    isLoading,
    error,
  };
};

export default useVersionData;
