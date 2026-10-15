import { getAccount, listAccounts, type AccountFilter } from '@/api/users';
import { listTeams } from '@/api/teams';

import { useDebouncedValue } from './useDebouncedValue';
import { useReloadOnFocus } from './useReloadOnFocus';
import { useRemote } from './useRemote';

const searchDelayMs = 300;

export function useAccounts(filter: AccountFilter) {
  const search = useDebouncedValue(filter.search ?? '', searchDelayMs);
  const remote = useRemote(() => listAccounts({ ...filter, search }), [search, filter.role, filter.status]);
  useReloadOnFocus(remote.reload);
  return remote;
}

export function useAccount(id: number) {
  return useRemote(() => getAccount(id), [id]);
}

export function useActiveAccountCount(): number | null {
  const { state, reload } = useRemote(() => listAccounts({ status: 'active' }), []);
  useReloadOnFocus(reload);
  return state.status === 'ready' ? state.data.length : null;
}

export function useTeams() {
  return useRemote(listTeams, []);
}
