import type { CalendarEvent, EventChanges, EventKind, NewEvent } from '@/api/calendar';

export type EventFormValues = {
  kind: EventKind;
  categoryCode: string;
  seasonId: number;
  circuitId: number | null;
  date: string;
  startTime: string;
  endTime: string;
  number: string;
  gpName: string;
  controlType: string;
};

// The names match the field the API reports on a rejection, so a 400 lands on
// the input that caused it without a translation table.
export type EventFormField =
  | 'kind'
  | 'categoryCode'
  | 'seasonId'
  | 'circuitId'
  | 'startsAt'
  | 'endsAt'
  | 'number'
  | 'gpName'
  | 'controlType';

export const eventFormFields: readonly EventFormField[] = [
  'kind',
  'categoryCode',
  'seasonId',
  'circuitId',
  'startsAt',
  'endsAt',
  'number',
  'gpName',
  'controlType',
];

export const eventKindOptions: readonly { value: EventKind; label: string }[] = [
  { value: 'race', label: 'Carrera' },
  { value: 'tyre_test', label: 'Prueba de neumáticos' },
  { value: 'technical_control', label: 'Control técnico' },
];

export function kindLabel(kind: EventKind): string {
  return eventKindOptions.find((option) => option.value === kind)?.label ?? kind;
}

export type Timing = {
  startsAt: string;
  endsAt: string;
};

export function newFormValues(seasonId: number, categoryCode: string): EventFormValues {
  return {
    kind: 'race',
    categoryCode,
    seasonId,
    circuitId: null,
    date: '',
    startTime: '',
    endTime: '',
    number: '',
    gpName: '',
    controlType: '',
  };
}

// The backend stores the timing as an instant and composes every date from it, so
// the form reads and writes the same UTC text and never disagrees with it.
export function formValuesOf(event: CalendarEvent): EventFormValues {
  return {
    kind: event.kind,
    categoryCode: event.categoryCode,
    seasonId: event.seasonId,
    circuitId: event.circuitId,
    date: event.startsAt.slice(0, 10),
    startTime: event.startsAt.slice(11, 16),
    endTime: event.endsAt.slice(11, 16),
    number: event.roundNumber === null ? '' : String(event.roundNumber),
    gpName: event.gpName ?? '',
    controlType: event.controlType ?? '',
  };
}

// Returns nil when the date or either time is missing or malformed; the semantic
// checks (span, season, round number) belong to the backend and come back as a
// field error.
export function timingOf(values: EventFormValues): Timing | null {
  const startsAt = instantOf(values.date, values.startTime);
  const endsAt = instantOf(values.date, values.endTime);
  if (!startsAt || !endsAt) {
    return null;
  }
  return { startsAt, endsAt };
}

export function toNewEvent(values: EventFormValues, timing: Timing): NewEvent {
  const input: NewEvent = {
    kind: values.kind,
    startsAt: timing.startsAt,
    endsAt: timing.endsAt,
    seasonId: values.seasonId,
    categoryCode: values.categoryCode,
  };
  if (values.kind === 'technical_control') {
    input.controlType = values.controlType.trim();
    return input;
  }
  if (values.circuitId !== null) {
    input.circuitId = values.circuitId;
  }
  if (values.kind === 'race') {
    input.number = Number(values.number);
    input.gpName = values.gpName.trim();
  }
  return input;
}

export function toEventChanges(event: CalendarEvent, values: EventFormValues, timing: Timing): EventChanges {
  const changes: EventChanges = {};
  if (timing.startsAt !== event.startsAt) {
    changes.startsAt = timing.startsAt;
  }
  if (timing.endsAt !== event.endsAt) {
    changes.endsAt = timing.endsAt;
  }
  if (values.seasonId !== event.seasonId) {
    changes.seasonId = values.seasonId;
  }
  if (values.categoryCode !== event.categoryCode) {
    changes.categoryCode = values.categoryCode;
  }
  if (values.circuitId !== null && values.circuitId !== event.circuitId) {
    changes.circuitId = values.circuitId;
  }
  if (values.kind === 'race') {
    const number = Number(values.number);
    if (number !== event.roundNumber) {
      changes.number = number;
    }
    const gpName = values.gpName.trim();
    if (gpName !== (event.gpName ?? '')) {
      changes.gpName = gpName;
    }
  }
  if (values.kind === 'technical_control') {
    const controlType = values.controlType.trim();
    if (controlType !== (event.controlType ?? '')) {
      changes.controlType = controlType;
    }
  }
  return changes;
}

export function hasChanges(changes: EventChanges): boolean {
  return Object.keys(changes).length > 0;
}

const datePattern = /^\d{4}-\d{2}-\d{2}$/;
const timePattern = /^\d{2}:\d{2}$/;

function instantOf(date: string, time: string): string | null {
  if (!isCalendarDate(date) || !timePattern.test(time)) {
    return null;
  }
  return `${date}T${time}:00Z`;
}

function isCalendarDate(value: string): boolean {
  if (!datePattern.test(value)) {
    return false;
  }
  const parsed = new Date(`${value}T00:00:00Z`);
  return !Number.isNaN(parsed.getTime()) && parsed.toISOString().slice(0, 10) === value;
}
