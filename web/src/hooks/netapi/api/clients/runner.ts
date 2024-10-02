import { SaltClient } from '../client.ts';

export class RunnerClient extends SaltClient {
  public async exec<T, U>(request: T): Promise<U> {
    await this.refreshToken();

    const command = {
      client: 'runner',
      ...request,
    };

    const response = await this.client.post('/', command, {
      headers: {
        'X-Auth-Token': this.token,
        Authorization: this.config.password,
      },
    });

    return response.data.return[0];
  }
}
