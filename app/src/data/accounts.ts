import { useFocusEffect } from 'expo-router';
import { useCallback, useRef } from 'react';

import { getAccount, listAccounts, type AccountFilter } from '@/api/users';
import { listTeams } from '@/api/teams';

import { useDebouncedValue } from './useDebouncedValue';
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

// Coming back from the create or edit screen must show the change, but the first
// focus is already covered by the initial load.
function useReloadOnFocus(reload: () => void) {
  const latestReload = useRef(reload);
  latestReload.current = reload;
  const isFirstFocus = useRef(true);

  useFocusEffect(
    useCallback(() => {
      if (isFirstFocus.current) {
        isFirstFocus.current = false;
        return;
      }
      latestReload.current();
    }, []),
  );
}
