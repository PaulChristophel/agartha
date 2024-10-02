// saltVersion.ts
import axios from 'axios';

const saltVersion = async (authToken: string, authSaltString: string): Promise<string> => {
  const parsedAuthSalt = JSON.parse(authSaltString);
  const { token } = parsedAuthSalt;

  try {
    const response = await axios.post(
      `/api/v1/netapi`,
      [
        {
          client: 'runner',
          fun: 'salt_version.version',
        },
      ],
      {
        headers: {
          Authorization: authToken,
          'X-Auth-Token': token,
        },
      }
    );

    return response.data.return;
  } catch (err) {
    console.error(`Failed to get salt version: `, err);
    return `Failed to get salt version: ${err}`;
  }
};

export default saltVersion;
