import { it, expect, describe } from 'vitest';

import {
  parseReturnPath,
  parseReturnFilter,
  serializeReturnFilter,
  newReturnFilterClause,
} from './returnFilter.ts';

describe('return filter serialization', () => {
  it('starts new clauses with empty string inputs', () => {
    expect(newReturnFilterClause()).toMatchObject({ path: '', value: '', valueType: 'string' });
  });

  it('serializes multiple typed clauses and escaped paths', () => {
    const serialized = serializeReturnFilter({
      logic: 'and',
      clauses: [
        {
          scope: 'key',
          key: 'file_|-Install root Bashrc_|-/root/.bashrc_|-managed',
          path: 'duration',
          operator: 'gt',
          value: '10',
          valueType: 'float',
        },
        {
          scope: 'any_key',
          key: '',
          path: 'changes.file\\.name',
          operator: 'icontains',
          value: 'BASHRC',
          valueType: 'string',
        },
      ],
    });

    expect(JSON.parse(serialized!)).toEqual({
      logic: 'and',
      clauses: [
        {
          scope: 'key',
          key: 'file_|-Install root Bashrc_|-/root/.bashrc_|-managed',
          path: ['duration'],
          operator: 'gt',
          value: '10',
          value_type: 'float',
        },
        {
          scope: 'any_key',
          path: ['changes', 'file.name'],
          operator: 'icontains',
          value: 'BASHRC',
          value_type: 'string',
        },
      ],
    });
  });

  it('normalizes legacy scope aliases while preserving clauses', () => {
    const serialized = JSON.stringify({
      logic: 'or',
      clauses: [
        { scope: 'root', path: ['return', 'message'], operator: 'not_exists' },
        {
          scope: 'any_state',
          path: ['result'],
          operator: 'eq',
          value: 'false',
          value_type: 'bool',
        },
      ],
    });

    expect(JSON.parse(serializeReturnFilter(parseReturnFilter(serialized))!)).toEqual({
      logic: 'or',
      clauses: [
        { scope: 'root', path: ['return', 'message'], operator: 'not_exists' },
        {
          scope: 'any_key',
          path: ['result'],
          operator: 'eq',
          value: 'false',
          value_type: 'bool',
        },
      ],
    });
  });

  it('parses escaped dots and backslashes in field paths', () => {
    expect(parseReturnPath('changes.file\\.name.directory\\\\child')).toEqual([
      'changes',
      'file.name',
      'directory\\child',
    ]);
  });

  it('rejects malformed URL state', () => {
    expect(parseReturnFilter('{')).toEqual({ logic: 'and', clauses: [] });
  });
});
