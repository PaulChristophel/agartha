import * as Core from '../core.ts';

export interface IPingRequest extends Core.ILocalRequest {}
export interface IPingResponse extends Core.IGenericBooleanResponse {}

export interface IVersionRequest extends Core.ILocalRequest {}
export interface IVersionResponse extends Core.IGenericBooleanResponse {}
