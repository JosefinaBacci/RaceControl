import { getDriver, getTeamDriver, listDrivers, listTeamDrivers, type DriverFilter } from '@/api/drivers';

import { useDebouncedValue } from './useDebouncedValue';
import { useReloadOnFocus } from './useReloadOnFocus';
import { useRemote } from './useRemote';

const searchDelayMs = 300;

export function useTeamDrivers() {
  const remote = useRemote(listTeamDrivers, []);
  useReloadOnFocus(remote.reload);
  return remote;
}

export function useTeamDriver(id: number) {
  return useRemote(() => getTeamDriver(id), [id]);
}

export function usePublicDrivers(filter: DriverFilter) {
  const search = useDebouncedValue(filter.search ?? '', searchDelayMs);
  const remote = useRemote(() => listDrivers({ ...filter, search }), [filter.category, search]);
  useReloadOnFocus(remote.reload);
  return remote;
}

export function usePublicDriver(id: number) {
  return useRemote(() => getDriver(id), [id]);
}
