import { router } from 'expo-router';
import { useState } from 'react';

import type { EventPeriod } from '@/api/calendar';
import { Button, ChipGroup, Screen } from '@/components';
import { useCalendarEvents } from '@/data/calendar';
import { EventAdminList } from '@/features/calendar/EventAdminList';
import { categoryFilterOptions, periodOptions } from '@/features/filters';
import { RemoteContent } from '@/features/RemoteContent';
import type { CategoryFilter } from '@/data/hooks';

export default function FiaCalendarScreen() {
  const [period, setPeriod] = useState<EventPeriod>('upcoming');
  const [category, setCategory] = useState<CategoryFilter>('all');
  const { state, reload } = useCalendarEvents({
    category: category === 'all' ? undefined : category,
    period,
  });

  return (
    <Screen
      title="Gestión del calendario"
      subtitle="Alta, modificación y baja de carreras, pruebas de neumáticos y controles técnicos."
      headerAction={<Button label="Nuevo evento" icon="add" compact onPress={() => router.push('/fia/calendar/new')} />}
    >
      <ChipGroup variant="underline" options={periodOptions} selected={period} onSelect={setPeriod} accessibilityLabel="Período" />
      <ChipGroup options={categoryFilterOptions} selected={category} onSelect={setCategory} accessibilityLabel="Filtrar por categoría" />
      <RemoteContent state={state} onRetry={reload}>
        {(events) => (
          <EventAdminList
            events={events}
            onEdit={(event) => router.push({ pathname: '/fia/calendar/[id]', params: { id: String(event.id) } })}
            onChanged={reload}
          />
        )}
      </RemoteContent>
    </Screen>
  );
}
