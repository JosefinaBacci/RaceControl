import type { Role } from './auth';
import { apiRequest } from './client';

export type Account = {
  id: number;
  username: string;
  email: string | null;
  role: Role;
  teamId: number | null;
  teamName: string | null;
  isActive: boolean;
  deactivatedAt: string | null;
  createdAt: string;
  updatedAt: string;
};

export type AccountStatus = 'active' | 'inactive';

export type AccountFilter = {
  search?: string;
  role?: Role;
  status?: AccountStatus;
};

export type NewAccount = {
  username: string;
  email: string;
  password: string;
  role: Role;
  teamId: number | null;
};

export type AccountChanges = {
  email?: string;
  role?: Role;
  teamId?: number;
  password?: string;
};

export async function listAccounts(filter: AccountFilter): Promise<Account[]> {
  const response = await apiRequest<{ users: Account[] }>('GET', `/users${queryString(filter)}`);
  return response.users;
}

export function getAccount(id: number): Promise<Account> {
  return apiRequest<Account>('GET', `/users/${id}`);
}

export function createAccount(input: NewAccount): Promise<Account> {
  return apiRequest<Account>('POST', '/users', input);
}

export function updateAccount(id: number, changes: AccountChanges): Promise<Account> {
  return apiRequest<Account>('PATCH', `/users/${id}`, changes);
}

export function deactivateAccount(id: number): Promise<Account> {
  return apiRequest<Account>('POST', `/users/${id}/deactivate`);
}

export function reactivateAccount(id: number): Promise<Account> {
  return apiRequest<Account>('POST', `/users/${id}/reactivate`);
}

function queryString(filter: AccountFilter): string {
  const params = { q: filter.search, role: filter.role, status: filter.status };
  const pairs = Object.entries(params)
    .filter((entry): entry is [string, string] => Boolean(entry[1]))
    .map(([key, value]) => `${key}=${encodeURIComponent(value)}`);
  return pairs.length > 0 ? `?${pairs.join('&')}` : '';
}
