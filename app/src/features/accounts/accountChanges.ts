import type { Account, AccountChanges } from '@/api/users';

import type { AccountFieldValues } from './AccountFields';

export function initialFieldValues(account: Account): AccountFieldValues {
  return { email: account.email ?? '', role: account.role, teamId: account.teamId };
}

export function accountChanges(account: Account, values: AccountFieldValues, password: string): AccountChanges {
  const changes: AccountChanges = {};
  const email = values.email.trim();
  if (email !== (account.email ?? '')) {
    changes.email = email;
  }
  if (values.role !== account.role) {
    changes.role = values.role;
  }
  if (values.role === 'team_admin' && values.teamId !== null && values.teamId !== account.teamId) {
    changes.teamId = values.teamId;
  }
  if (password) {
    changes.password = password;
  }
  return changes;
}

export function hasChanges(changes: AccountChanges): boolean {
  return Object.keys(changes).length > 0;
}
