import { getTeamDriver, listTeamDrivers } from '@/api/drivers';

import { useReloadOnFocus } from './useReloadOnFocus';
import { useRemote } from './useRemote';

export function useTeamDrivers() {
  const remote = useRemote(listTeamDrivers, []);
  useReloadOnFocus(remote.reload);
  return remote;
}

export function useTeamDriver(id: number) {
  return useRemote(() => getTeamDriver(id), [id]);
}
