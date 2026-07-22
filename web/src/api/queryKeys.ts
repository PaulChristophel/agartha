export type QueryKey = readonly [string, ...unknown[]];

export const queryKeys = {
  auth: {
    methods: () => ['auth', 'methods'] as const,
    user: (userId: string) => ['auth', 'user', userId] as const,
  },
  saltKeys: {
    all: () => ['salt-keys'] as const,
    detail: (id: string) => ['salt-keys', id] as const,
  },
  saltMinions: {
    all: () => ['salt-minions'] as const,
    list: (params: object) => ['salt-minions', 'list', params] as const,
    detail: (id: string) => ['salt-minions', id] as const,
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
} satisfies Record<string, Record<string, (...args: never[]) => QueryKey>>;
