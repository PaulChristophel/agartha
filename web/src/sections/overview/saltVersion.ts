// saltVersion.ts
import { apiClient as axios } from 'src/api/client.ts';

const saltVersion = async (): Promise<string> => {
  try {
    const response = await axios.post(`/api/v1/netapi`, [
      {
        client: 'runner',
        fun: 'salt_version.version',
      },
    ]);

    return response.data.return;
  } catch (err) {
    console.error(`Failed to get salt version: `, err);
    return `Failed to get salt version: ${err}`;
  }
};

export default saltVersion;
