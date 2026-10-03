import * as SecureStore from 'expo-secure-store';

const sessionTokenKey = 'rc_session';

export const tokenStore = {
  usesCookies: false,
  read: () => SecureStore.getItemAsync(sessionTokenKey),
  save: (token: string) => SecureStore.setItemAsync(sessionTokenKey, token),
  clear: () => SecureStore.deleteItemAsync(sessionTokenKey),
};
