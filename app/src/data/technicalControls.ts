import { listEvents } from '@/api/calendar';
import { getControlSheet, listControlHistory, listControlTypes, type ControlHistoryQuery } from '@/api/technicalControls';

import { useReloadOnFocus } from './useReloadOnFocus';
import { useRemote } from './useRemote';

export function useControlHistory(query: ControlHistoryQuery) {
  const remote = useRemote(() => listControlHistory(query), [query.from, query.to, query.type, query.category, query.team]);
  useReloadOnFocus(remote.reload);
  return remote;
}

export function useControlTypes() {
  return useRemote(listControlTypes, []);
}

export function useControlSheet(eventId: number) {
  const remote = useRemote(() => getControlSheet(eventId), [eventId]);
  useReloadOnFocus(remote.reload);
  return remote;
}

// Newest first: the FIA usually grades the control that just ended.
export function useGradableControls() {
  const remote = useRemote(async () => {
    const controls = await listEvents({ kind: 'technical_control', period: 'past' });
    return [...controls].reverse();
  }, []);
  useReloadOnFocus(remote.reload);
  return remote;
}
