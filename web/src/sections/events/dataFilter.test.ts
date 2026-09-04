import { it, expect, describe } from 'vitest';

import {
  parseDataPath,
  parseDataFilter,
  serializeDataFilter,
  newDataFilterClause,
} from './dataFilter.ts';

describe('event data filter serialization', () => {
  it('starts new clauses with empty string inputs', () => {
    expect(newDataFilterClause()).toMatchObject({ path: '', value: '', valueType: 'string' });
  });

  it('serializes multiple typed clauses and escaped paths', () => {
    const serialized = serializeDataFilter({
      logic: 'or',
      clauses: [
        {
          scope: 'any_key',
          key: '',
          containerPath: 'return',
          path: 'result',
          operator: 'eq',
          value: 'false',
          valueType: 'bool',
        },
        {
          scope: 'root',
          key: '',
          containerPath: '',
          path: 'retcode',
          operator: 'gt',
          value: '0',
          valueType: 'int',
        },
      ],
    });

    expect(JSON.parse(serialized!)).toEqual({
      logic: 'or',
      clauses: [
        {
          scope: 'any_key',
          container_path: ['return'],
          path: ['result'],
          operator: 'eq',
          value: 'false',
          value_type: 'bool',
        },
        {
          scope: 'root',
          path: ['retcode'],
          operator: 'gt',
          value: '0',
          value_type: 'int',
        },
      ],
    });
  });

  it('round trips event data filter URL state', () => {
    const value = JSON.stringify({
      logic: 'and',
      clauses: [{ scope: 'root', path: ['minion_id'], operator: 'exists' }],
    });

    expect(JSON.parse(serializeDataFilter(parseDataFilter(value))!)).toEqual(JSON.parse(value));
    expect(parseDataPath('payload.file\\.name')).toEqual(['payload', 'file.name']);
  });

  it('preserves a nested container for arbitrary highstate IDs', () => {
    const value = JSON.stringify({
      logic: 'and',
      clauses: [
        {
          scope: 'any_key',
          container_path: ['return'],
          path: ['result'],
          operator: 'eq',
          value: 'false',
          value_type: 'bool',
        },
      ],
    });

    expect(JSON.parse(serializeDataFilter(parseDataFilter(value))!)).toEqual(JSON.parse(value));
  });

  it('rejects malformed URL state', () => {
    expect(parseDataFilter('{')).toEqual({ logic: 'and', clauses: [] });
  });
});
