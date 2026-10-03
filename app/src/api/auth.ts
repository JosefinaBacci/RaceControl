import { apiRequest } from './client';
import { tokenStore } from './tokenStore';

export type Role = 'fia_admin' | 'team_admin';

export type SessionUser = {
  id: number;
  username: string;
  role: Role;
  teamId: number | null;
};

type LoginResponse = {
  token?: string;
  user: SessionUser;
};

export async function login(username: string, password: string): Promise<SessionUser> {
  const response = await apiRequest<LoginResponse>('POST', '/auth/login', { username, password });
  if (response.token) {
    await tokenStore.save(response.token);
  }
  return response.user;
}

export async function logout(): Promise<void> {
  try {
    await apiRequest<void>('POST', '/auth/logout');
  } finally {
    await tokenStore.clear();
  }
}

export function changeOwnPassword(currentPassword: string, newPassword: string): Promise<void> {
  return apiRequest<void>('POST', '/account/password', { currentPassword, newPassword });
}

export function fetchCurrentUser(): Promise<SessionUser> {
  return apiRequest<SessionUser>('GET', '/auth/me');
}
