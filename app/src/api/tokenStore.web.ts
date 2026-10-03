export const tokenStore = {
  usesCookies: true,
  read: async (): Promise<string | null> => null,
  save: async (_token: string): Promise<void> => undefined,
  clear: async (): Promise<void> => undefined,
};
