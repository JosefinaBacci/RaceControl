import { useState } from 'react';

import type { EventPeriod } from '@/api/calendar';
import { ChipGroup, Screen } from '@/components';
import { useCalendarEvents } from '@/data/calendar';
import { CalendarDownloads } from '@/features/calendar/CalendarDownloads';
import { EventList } from '@/features/EventList';
import { categoryFilterOptions, periodOptions } from '@/features/filters';
import { RemoteContent } from '@/features/RemoteContent';
import type { CategoryFilter } from '@/mocks/catalog';

export default function PublicCalendarScreen() {
  const [period, setPeriod] = useState<EventPeriod>('upcoming');
  const [category, setCategory] = useState<CategoryFilter>('all');
  const categoryCode = category === 'all' ? undefined : category;

  const { state, reload } = useCalendarEvents({ category: categoryCode, period });

  return (
    <Screen
      title="Calendario 2026"
      subtitle="Carreras, pruebas de neumáticos y controles técnicos."
      headerAction={<CalendarDownloads filter={{ category: categoryCode, period }} />}
    >
      <ChipGroup variant="underline" options={periodOptions} selected={period} onSelect={setPeriod} accessibilityLabel="Período" />
      <ChipGroup options={categoryFilterOptions} selected={category} onSelect={setCategory} accessibilityLabel="Filtrar por categoría" />
      <RemoteContent state={state} onRetry={reload}>
        {(events) => <EventList events={events} />}
      </RemoteContent>
    </Screen>
  );
}
