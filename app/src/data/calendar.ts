import { apiBaseUrl, queryString } from '@/api/client';
import { getEvent, listCircuits, listEvents, listSeasons, type EventFilter } from '@/api/calendar';
import { useReloadOnFocus } from '@/data/useReloadOnFocus';
import { useRemote } from '@/data/useRemote';

export function useCalendarEvents(filter: EventFilter) {
  const remote = useRemote(() => listEvents(filter), [filter.category, filter.kind, filter.period]);
  useReloadOnFocus(remote.reload);
  return remote;
}

export function useUpcomingEvents(limit: number, filter: Omit<EventFilter, 'period'> = {}) {
  const remote = useRemote(async () => {
    const events = await listEvents({ ...filter, period: 'upcoming' });
    return events.slice(0, limit);
  }, [limit, filter.category, filter.kind]);
  useReloadOnFocus(remote.reload);
  return remote;
}

export function useNextRace(filter: Omit<EventFilter, 'period' | 'kind'> = {}) {
  const remote = useRemote(async () => {
    const events = await listEvents({ ...filter, period: 'upcoming', kind: 'race' });
    return events[0] ?? null;
  }, [filter.category]);
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

export function calendarDownloadUrl(format: 'ics' | 'pdf', filter: EventFilter = {}): string {
  const path = format === 'ics' ? '/calendar.ics' : '/calendar.pdf';
  return `${apiBaseUrl}${path}${queryString(filter)}`;
}
