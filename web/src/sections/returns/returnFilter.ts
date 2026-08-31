export type ReturnFilterLogic = 'and' | 'or';

export type ReturnFilterScope = 'key' | 'any_key' | 'root';

export type ReturnFilterOperator =
  | 'eq'
  | 'ne'
  | 'gt'
  | 'gte'
  | 'lt'
  | 'lte'
  | 'contains'
  | 'icontains'
  | 'regex'
  | 'exists'
  | 'not_exists';

export type ReturnValueType = 'string' | 'bool' | 'int' | 'float' | 'null';

export interface ReturnFilterClause {
  scope: ReturnFilterScope;
  key: string;
  path: string;
  operator: ReturnFilterOperator;
  value: string;
  valueType: ReturnValueType;
}

export interface ReturnFilterState {
  logic: ReturnFilterLogic;
  clauses: ReturnFilterClause[];
}

interface SerializedReturnFilterClause {
  scope: ReturnFilterScope;
  key?: string;
  path?: string[];
  operator: ReturnFilterOperator;
  value?: string;
  value_type?: ReturnValueType;
}

const scopes: ReturnFilterScope[] = ['key', 'any_key', 'root'];
const operators: ReturnFilterOperator[] = [
  'eq',
  'ne',
  'gt',
  'gte',
  'lt',
  'lte',
  'contains',
  'icontains',
  'regex',
  'exists',
  'not_exists',
];
const valueTypes: ReturnValueType[] = ['string', 'bool', 'int', 'float', 'null'];

export const newReturnFilterClause = (): ReturnFilterClause => ({
  scope: 'key',
  key: '',
  path: 'result',
  operator: 'eq',
  value: 'false',
  valueType: 'bool',
});

export const emptyReturnFilter = (): ReturnFilterState => ({
  logic: 'and',
  clauses: [],
});

export const returnOperatorNeedsValue = (operator: ReturnFilterOperator): boolean =>
  operator !== 'exists' && operator !== 'not_exists';

export const parseReturnPath = (path: string): string[] => {
  if (!path) return [];

  const components: string[] = [];
  let component = '';
  let escaped = false;
  for (const character of path) {
    if (escaped) {
      component += character;
      escaped = false;
    } else if (character === '\\') {
      escaped = true;
    } else if (character === '.') {
      components.push(component);
      component = '';
    } else {
      component += character;
    }
  }
  if (escaped) component += '\\';
  components.push(component);
  return components;
};

const formatReturnPath = (path: string[]): string =>
  path.map((component) => component.replaceAll('\\', '\\\\').replaceAll('.', '\\.')).join('.');

export const serializeReturnFilter = (filter: ReturnFilterState): string | undefined => {
  if (filter.clauses.length === 0) return undefined;

  return JSON.stringify({
    logic: filter.logic,
    clauses: filter.clauses.map((clause): SerializedReturnFilterClause => {
      const serialized: SerializedReturnFilterClause = {
        scope: clause.scope,
        operator: clause.operator,
      };
      if (clause.scope === 'key') serialized.key = clause.key;
      const path = parseReturnPath(clause.path);
      if (path.length > 0) serialized.path = path;
      if (returnOperatorNeedsValue(clause.operator)) {
        serialized.value = clause.value;
        serialized.value_type = clause.valueType;
      }
      return serialized;
    }),
  });
};

export const parseReturnFilter = (value: string | null): ReturnFilterState => {
  if (!value) return emptyReturnFilter();

  try {
    const parsed: unknown = JSON.parse(value);
    if (!parsed || typeof parsed !== 'object') return emptyReturnFilter();
    const candidate = parsed as { logic?: unknown; clauses?: unknown };
    if (
      (candidate.logic !== 'and' && candidate.logic !== 'or') ||
      !Array.isArray(candidate.clauses) ||
      candidate.clauses.length === 0 ||
      candidate.clauses.length > 20
    ) {
      return emptyReturnFilter();
    }

    const clauses: ReturnFilterClause[] = [];
    for (const rawClause of candidate.clauses) {
      if (!rawClause || typeof rawClause !== 'object') return emptyReturnFilter();
      const clause = rawClause as Record<string, unknown>;
      const normalizedScope =
        clause.scope === 'state' ? 'key' : clause.scope === 'any_state' ? 'any_key' : clause.scope;
      if (
        !scopes.includes(normalizedScope as ReturnFilterScope) ||
        !operators.includes(clause.operator as ReturnFilterOperator)
      ) {
        return emptyReturnFilter();
      }
      const valueType = clause.value_type ?? 'string';
      if (!valueTypes.includes(valueType as ReturnValueType)) return emptyReturnFilter();
      const path = clause.path ?? [];
      if (!Array.isArray(path) || path.some((component) => typeof component !== 'string')) {
        return emptyReturnFilter();
      }
      clauses.push({
        scope: normalizedScope as ReturnFilterScope,
        key: typeof clause.key === 'string' ? clause.key : '',
        path: formatReturnPath(path as string[]),
        operator: clause.operator as ReturnFilterOperator,
        value: typeof clause.value === 'string' ? clause.value : '',
        valueType: valueType as ReturnValueType,
      });
    }
    return { logic: candidate.logic, clauses };
  } catch {
    return emptyReturnFilter();
  }
};
