import { getEvent, listEvents } from '@/api/calendar';
import type { HistoryQuery } from '@/api/history';
import { getTyreCompounds, getTyreTestHistory, getTyreTestRevisions } from '@/api/tyreTests';
import { useReloadOnFocus } from '@/data/useReloadOnFocus';
import { useRemote } from '@/data/useRemote';

export function useTyreCompounds() {
  return useRemote(getTyreCompounds, []);
}

export function useTyreTestHistory(query: HistoryQuery) {
  const remote = useRemote(() => getTyreTestHistory(query), [query.from, query.to, query.type]);
  useReloadOnFocus(remote.reload);
  return remote;
}

export function useTyreTestEvents() {
  const remote = useRemote(() => listEvents({ kind: 'tyre_test' }), []);
  useReloadOnFocus(remote.reload);
  return remote;
}

export function useTyreTestRecord(eventId: number) {
  return useRemote(async () => {
    const [event, compounds, revisions] = await Promise.all([
      getEvent(eventId),
      getTyreCompounds(),
      getTyreTestRevisions(eventId),
    ]);
    return { event, compounds, revisions };
  }, [eventId]);
}
