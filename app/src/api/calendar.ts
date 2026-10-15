import { apiRequest } from './client';

export type EventKind = 'race' | 'tyre_test' | 'technical_control';

export type EventStatus = 'scheduled' | 'finished';

export type EventPeriod = 'upcoming' | 'past';

export type CalendarEvent = {
  id: number;
  kind: EventKind;
  name: string;
  categoryCode: string;
  categoryName: string;
  seasonId: number;
  seasonYear: number;
  roundNumber: number | null;
  gpName: string | null;
  controlType: string | null;
  circuitId: number | null;
  circuitName: string | null;
  countryCode: string | null;
  startsAt: string;
  endsAt: string;
  status: EventStatus;
  isActive: boolean;
};

export type EventFilter = {
  category?: string;
  kind?: EventKind;
  period?: EventPeriod;
};

export type NewEvent = {
  kind: EventKind;
  startsAt: string;
  endsAt: string;
  seasonId: number;
  categoryCode: string;
  circuitId?: number;
  number?: number;
  gpName?: string;
  controlType?: string;
};

export type EventChanges = {
  startsAt?: string;
  endsAt?: string;
  seasonId?: number;
  categoryCode?: string;
  circuitId?: number;
  number?: number;
  gpName?: string;
  controlType?: string;
};

export type Circuit = {
  id: number;
  name: string;
  locality: string;
  countryCode: string;
};

export type Season = {
  id: number;
  year: number;
};

export type CalendarCatalogs = {
  seasons: Season[];
  circuits: Circuit[];
};

export async function listEvents(filter: EventFilter = {}): Promise<CalendarEvent[]> {
  const response = await apiRequest<{ events: CalendarEvent[] }>('GET', `/events${queryString(filter)}`);
  return response.events;
}

export function getEvent(id: number): Promise<CalendarEvent> {
  return apiRequest<CalendarEvent>('GET', `/events/${id}`);
}

export async function listCircuits(): Promise<Circuit[]> {
  const response = await apiRequest<{ circuits: Circuit[] }>('GET', '/circuits');
  return response.circuits;
}

export async function listSeasons(): Promise<Season[]> {
  const response = await apiRequest<{ seasons: Season[] }>('GET', '/seasons');
  return response.seasons;
}

export function createEvent(input: NewEvent): Promise<CalendarEvent> {
  return apiRequest<CalendarEvent>('POST', '/events', input);
}

export function updateEvent(id: number, changes: EventChanges): Promise<CalendarEvent> {
  return apiRequest<CalendarEvent>('PATCH', `/events/${id}`, changes);
}

export function cancelEvent(id: number): Promise<CalendarEvent> {
  return apiRequest<CalendarEvent>('DELETE', `/events/${id}`);
}

function queryString(filter: EventFilter): string {
  const params = { category: filter.category, kind: filter.kind, period: filter.period };
  const pairs = Object.entries(params)
    .filter((entry): entry is [string, string] => Boolean(entry[1]))
    .map(([key, value]) => `${key}=${encodeURIComponent(value)}`);
  return pairs.length > 0 ? `?${pairs.join('&')}` : '';
}
