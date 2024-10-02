import * as Core from '../core.ts';

export interface ISetRequest extends Core.ILocalRequest {
  kwarg: {
    key: string;
    val: string;
    force?: boolean;
    destructive?: boolean;
    delimiter?: string;
  };
}

export interface ISetResponse {
  [key: string]: {
    comment: string;
    changes: [key: string];
    result: boolean;
  };
}
