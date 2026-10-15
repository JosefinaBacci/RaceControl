import type { EventKind } from '@/api/calendar';
import type { BadgeTone } from '@/components';

export const eventKindLabels: Record<EventKind, string> = {
  race: 'Carrera',
  tyre_test: 'Prueba de neumáticos',
  technical_control: 'Control técnico',
};

export const eventKindTones: Record<EventKind, BadgeTone> = {
  race: 'accent',
  tyre_test: 'info',
  technical_control: 'warning',
};

export const eventKindOptions: readonly { value: EventKind; label: string }[] = (
  Object.keys(eventKindLabels) as EventKind[]
).map((kind) => ({ value: kind, label: eventKindLabels[kind] }));

export function kindLabel(kind: EventKind): string {
  return eventKindLabels[kind];
}
