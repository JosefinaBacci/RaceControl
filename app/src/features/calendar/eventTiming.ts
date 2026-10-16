import type { CalendarEvent } from '@/api/calendar';

// Mirrors the API, which refuses to cancel an event once it has started: from
// then on it is part of the season's history.
export function hasStarted(event: CalendarEvent): boolean {
  return new Date(event.startsAt).getTime() <= Date.now();
}
