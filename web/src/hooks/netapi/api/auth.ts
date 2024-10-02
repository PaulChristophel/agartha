import axios from 'axios';

import { SaltClient, ISaltConfigOptions } from './client.ts';

export class SaltAuth extends SaltClient {
  constructor(config: ISaltConfigOptions) {
    super(config);
    this.client = axios.create({
      baseURL: `${this.config.endpoint}`,
      headers: {
        Authorization: `${config.password}`,
      },
    });
  }

  public async login() {
    const response = await this.client.post('/login', {});

    if (response.status === 200) {
      this.token = response.data.return[0].token;
      this.expire = response.data.return[0].expire;
      return response.data.return[0];
    }
    throw new Error('failed to login to Salt API');
  }
}
