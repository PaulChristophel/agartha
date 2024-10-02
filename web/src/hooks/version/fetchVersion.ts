import axios from 'axios';

interface VersionData {
  build: string;
  version: string;
  compile_time: string;
  go_version: string;
  platform: string;
  path: string;
}

export const fetchVersion = async (): Promise<VersionData> => {
  try {
    const response = await axios.get('/version');
    return response.data;
  } catch (error) {
    console.error('Error fetching version:', error);
    throw new Error('Failed to fetch version data');
  }
};
