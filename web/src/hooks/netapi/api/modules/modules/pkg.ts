import * as Core from '../core.ts';

export interface IInstallRequest extends Core.ILocalRequest {
  kwarg: {
    name: string;
    version?: string;
  };
}

export interface IInstallResponse {
  [key: string]: {
    [key: string]: {
      old: string;
      new: string;
    };
  };
}
