import { useState } from 'react';

import { Button, ChipGroup, Screen } from '@/components';
import { useCalendar, type CategoryFilter, type EventPeriod } from '@/data/hooks';
import { EventList } from '@/features/EventList';
import { categoryFilterOptions, periodOptions } from '@/features/filters';
import { MockNotice } from '@/features/MockNotice';

export default function FiaCalendarScreen() {
  const [period, setPeriod] = useState<EventPeriod>('upcoming');
  const [category, setCategory] = useState<CategoryFilter>('all');
  const events = useCalendar(category, period);

  return (
    <Screen
      title="Gestión del calendario"
      subtitle="Alta, modificación y baja de carreras, pruebas y controles técnicos."
      headerAction={<Button label="Nuevo evento" icon="add" compact disabled />}
    >
      <MockNotice />
      <ChipGroup variant="underline" options={periodOptions} selected={period} onSelect={setPeriod} accessibilityLabel="Período" />
      <ChipGroup options={categoryFilterOptions} selected={category} onSelect={setCategory} accessibilityLabel="Filtrar por categoría" />
      <EventList events={events} />
    </Screen>
  );
}
