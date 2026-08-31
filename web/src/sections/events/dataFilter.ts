export type DataFilterLogic = 'and' | 'or';

export type DataFilterScope = 'key' | 'any_key' | 'root';

export type DataFilterOperator =
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

export type DataValueType = 'string' | 'bool' | 'int' | 'float' | 'null';

export interface DataFilterClause {
  scope: DataFilterScope;
  key: string;
  containerPath: string;
  path: string;
  operator: DataFilterOperator;
  value: string;
  valueType: DataValueType;
}

export interface DataFilterState {
  logic: DataFilterLogic;
  clauses: DataFilterClause[];
}

interface SerializedDataFilterClause {
  scope: DataFilterScope;
  key?: string;
  container_path?: string[];
  path?: string[];
  operator: DataFilterOperator;
  value?: string;
  value_type?: DataValueType;
}

const scopes: DataFilterScope[] = ['key', 'any_key', 'root'];
const operators: DataFilterOperator[] = [
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
const valueTypes: DataValueType[] = ['string', 'bool', 'int', 'float', 'null'];

export const newDataFilterClause = (): DataFilterClause => ({
  scope: 'root',
  key: '',
  containerPath: '',
  path: 'result',
  operator: 'eq',
  value: 'false',
  valueType: 'bool',
});

export const emptyDataFilter = (): DataFilterState => ({
  logic: 'and',
  clauses: [],
});

export const dataOperatorNeedsValue = (operator: DataFilterOperator): boolean =>
  operator !== 'exists' && operator !== 'not_exists';

export const parseDataPath = (path: string): string[] => {
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

const formatDataPath = (path: string[]): string =>
  path.map((component) => component.replaceAll('\\', '\\\\').replaceAll('.', '\\.')).join('.');

export const serializeDataFilter = (filter: DataFilterState): string | undefined => {
  if (filter.clauses.length === 0) return undefined;

  return JSON.stringify({
    logic: filter.logic,
    clauses: filter.clauses.map((clause): SerializedDataFilterClause => {
      const serialized: SerializedDataFilterClause = {
        scope: clause.scope,
        operator: clause.operator,
      };
      if (clause.scope === 'key') serialized.key = clause.key;
      if (clause.scope === 'any_key') {
        const containerPath = parseDataPath(clause.containerPath);
        if (containerPath.length > 0) serialized.container_path = containerPath;
      }
      const path = parseDataPath(clause.path);
      if (path.length > 0) serialized.path = path;
      if (dataOperatorNeedsValue(clause.operator)) {
        serialized.value = clause.value;
        serialized.value_type = clause.valueType;
      }
      return serialized;
    }),
  });
};

export const parseDataFilter = (value: string | null): DataFilterState => {
  if (!value) return emptyDataFilter();

  try {
    const parsed: unknown = JSON.parse(value);
    if (!parsed || typeof parsed !== 'object') return emptyDataFilter();
    const candidate = parsed as { logic?: unknown; clauses?: unknown };
    if (
      (candidate.logic !== 'and' && candidate.logic !== 'or') ||
      !Array.isArray(candidate.clauses) ||
      candidate.clauses.length === 0 ||
      candidate.clauses.length > 20
    ) {
      return emptyDataFilter();
    }

    const clauses: DataFilterClause[] = [];
    for (const rawClause of candidate.clauses) {
      if (!rawClause || typeof rawClause !== 'object') return emptyDataFilter();
      const clause = rawClause as Record<string, unknown>;
      if (
        !scopes.includes(clause.scope as DataFilterScope) ||
        !operators.includes(clause.operator as DataFilterOperator)
      ) {
        return emptyDataFilter();
      }
      const valueType = clause.value_type ?? 'string';
      if (!valueTypes.includes(valueType as DataValueType)) return emptyDataFilter();

      const path = clause.path ?? [];
      const containerPath = clause.container_path ?? [];
      if (
        !Array.isArray(path) ||
        path.some((component) => typeof component !== 'string') ||
        !Array.isArray(containerPath) ||
        containerPath.some((component) => typeof component !== 'string')
      ) {
        return emptyDataFilter();
      }

      clauses.push({
        scope: clause.scope as DataFilterScope,
        key: typeof clause.key === 'string' ? clause.key : '',
        containerPath: formatDataPath(containerPath as string[]),
        path: formatDataPath(path as string[]),
        operator: clause.operator as DataFilterOperator,
        value: typeof clause.value === 'string' ? clause.value : '',
        valueType: valueType as DataValueType,
      });
    }
    return { logic: candidate.logic, clauses };
  } catch {
    return emptyDataFilter();
  }
};
