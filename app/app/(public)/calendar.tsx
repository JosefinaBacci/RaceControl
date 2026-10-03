import { useState } from 'react';

import { ChipGroup, Screen } from '@/components';
import { useCalendar, type CategoryFilter, type EventPeriod } from '@/data/hooks';
import { EventList } from '@/features/EventList';
import { categoryFilterOptions, periodOptions } from '@/features/filters';

export default function PublicCalendarScreen() {
  const [period, setPeriod] = useState<EventPeriod>('upcoming');
  const [category, setCategory] = useState<CategoryFilter>('all');
  const events = useCalendar(category, period);

  return (
    <Screen title="Calendario 2026" subtitle="Carreras, pruebas de neumáticos y controles técnicos.">
      <ChipGroup variant="underline" options={periodOptions} selected={period} onSelect={setPeriod} accessibilityLabel="Período" />
      <ChipGroup options={categoryFilterOptions} selected={category} onSelect={setCategory} accessibilityLabel="Filtrar por categoría" />
      <EventList events={events} />
    </Screen>
  );
}
