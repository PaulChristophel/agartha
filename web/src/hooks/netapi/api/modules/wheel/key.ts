import { Core } from '../../index.ts';

export interface IListRequest extends Core.IRunnerRequest {
  match?: string;
}
export interface IListResponse {
  minions: string[];
  minions_pre: string[];
  minions_rejected: string[];
  minions_denied: string[];
  local: string[];
}

// key.accept
export interface IRequest extends Core.IRunnerRequest {
  match: string;
  include_rejected?: boolean;
  include_denied?: boolean;
}
export interface IResponse {
  minions: string[];
}

export interface Match {
  minions?: string[];
  minions_pre?: string[];
}

// key.accept
export interface IDictRequest extends Core.IRunnerRequest {
  match: Match | string[];
  include_accepted?: boolean;
  include_rejected?: boolean;
  include_denied?: boolean;
}
