import { getEvent, listCircuits, listEvents, listSeasons, type EventFilter } from '@/api/calendar';
import { useReloadOnFocus } from '@/data/useReloadOnFocus';
import { useRemote } from '@/data/useRemote';

export function useCalendarEvents(filter: EventFilter) {
  const remote = useRemote(() => listEvents(filter), [filter.category, filter.kind, filter.period]);
  useReloadOnFocus(remote.reload);
  return remote;
}

// The create and edit screens need the same two catalogues on top of the event,
// so one load answers both and the screen shows a single loading state.
export function useEventForm(id?: number) {
  return useRemote(async () => {
    const [seasons, circuits] = await Promise.all([listSeasons(), listCircuits()]);
    return { event: id === undefined ? null : await getEvent(id), seasons, circuits };
  }, [id]);
}
