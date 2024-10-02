import * as Core from '../core.ts';

export interface IHighStateRequest extends Core.ILocalRequest {}
export interface IHighStateResponse {
  [key: string]: {
    [key: string]: {
      __id__?: string;
      __run_num__?: number;
      __sls__?: string;
      result: boolean;
      comment?: string;
      duration?: number;
      start_time?: string;
      changes?: {
        [key: string]: string;
      };
    };
  };
}

export interface ISlsRequest extends Core.ILocalRequest {
  kwarg: {
    mods: string;
    test?: boolean;
    exclude?: string;
    pillar?: Record<string, unknown>;
  };
}
export interface ISlsResponse extends IHighStateResponse {}
