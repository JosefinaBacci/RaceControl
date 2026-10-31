import type { Href } from 'expo-router';

import type { EventKind } from '@/api/calendar';
import { controlResultsHref } from '@/features/technicalControls/routes';
import { tyreTestResultsHref } from '@/features/tyreTests/routes';

// Where the FIA records the official results of an event, per kind. The event
// screen offers the shortcut only for the kinds listed here.
export const recordResultRoutes: Partial<Record<EventKind, { label: string; href: (eventId: number) => Href }>> = {
  technical_control: { label: 'Registrar resultados del control', href: controlResultsHref },
  tyre_test: { label: 'Resultados por compuesto', href: tyreTestResultsHref },
};
