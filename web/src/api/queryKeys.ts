export type QueryKey = readonly [string, ...unknown[]];

export const queryKeys = {
  auth: {
    methods: () => ['auth', 'methods'] as const,
    session: () => ['auth', 'session'] as const,
    user: (userId: string) => ['auth', 'user', userId] as const,
  },
  saltKeys: {
    all: () => ['salt-keys'] as const,
    detail: (id: string) => ['salt-keys', id] as const,
  },
  saltMinions: {
    all: () => ['salt-minions'] as const,
    list: (params: object) => ['salt-minions', 'list', params] as const,
    byId: (id: string) => ['salt-minions', 'id', id] as const,
    byUUID: (id: string) => ['salt-minions', 'uuid', id] as const,
  },
  saltCache: {
    all: () => ['salt-cache'] as const,
    list: (params: object) => ['salt-cache', 'list', params] as const,
    detail: (bank: string, key: string) => ['salt-cache', bank, key] as const,
  },
  commands: {
    all: () => ['salt-commands'] as const,
    submission: (client: string, fun: string, target?: string) =>
      ['salt-commands', client, fun, target] as const,
  },
  jids: {
    all: () => ['jids'] as const,
    detail: (id: string) => ['jids', id] as const,
  },
  saltEvents: {
    all: () => ['salt-events'] as const,
    detail: (id: number) => ['salt-events', id] as const,
  },
  saltReturns: {
    all: () => ['salt-returns'] as const,
    detail: (jid: string, id: string, loadReturn: boolean, loadFullRet: boolean) =>
      ['salt-returns', jid, id, loadReturn, loadFullRet] as const,
  },
  highStates: {
    all: () => ['high-states'] as const,
    detail: (id: string, loadReturn: boolean, loadFullRet: boolean) =>
      ['high-states', id, loadReturn, loadFullRet] as const,
  },
  conformity: {
    all: () => ['conformity'] as const,
    detail: (id: string) => ['conformity', id] as const,
  },
} satisfies Record<string, Record<string, (...args: never[]) => QueryKey>>;
