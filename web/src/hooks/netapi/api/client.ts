import axios, { AxiosInstance } from 'axios';

import { Logger } from './logger.ts';

export interface ISaltConfigOptions {
  endpoint: string;
  token?: string;
  expire?: number;
  password: string;
  logger?: Logger; // Use Logger interface here
}

export abstract class SaltClient {
  protected config: ISaltConfigOptions;

  protected token: string | null = null;

  protected expire: number | null = null;

  protected client: AxiosInstance;

  constructor(config: ISaltConfigOptions) {
    this.config = config;

    if ('token' in config && typeof config.token === 'string') {
      this.token = config.token;
    }

    if ('expire' in config && typeof config.expire === 'number') {
      this.expire = config.expire;
    }

    this.client = axios.create({
      baseURL: `${this.config.endpoint}`,
      headers: {
        Authorization: `${config.password}`,
      },
    });
  }

  protected async refreshToken(force?: boolean) {
    if (!this.token || !this.expire || this.expire <= new Date().getTime() / 1000 || force) {
      this.config.logger?.debug('refreshing token');

      const results = await this.client.post('/login', {});

      if (results.status === 200) {
        this.config.logger?.debug('login successful');

        this.token = results.data.return[0].token;
        this.expire = results.data.return[0].expire;
      } else {
        throw new Error('failed to login to the Salt API');
      }
    }
  }
}
