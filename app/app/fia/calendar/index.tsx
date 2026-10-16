import { router } from 'expo-router';
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import type { EventPeriod } from '@/api/calendar';
import { Button, ChipGroup, Screen } from '@/components';
import { useCalendarEvents } from '@/data/calendar';
import { CalendarDownloads } from '@/features/calendar/CalendarDownloads';
import { EventAdminList } from '@/features/calendar/EventAdminList';
import { categoryFilterOptions, periodOptions } from '@/features/filters';
import { RemoteContent } from '@/features/RemoteContent';
import type { CategoryFilter } from '@/mocks/catalog';
import { spacing } from '@/theme';

export default function FiaCalendarScreen() {
  const [period, setPeriod] = useState<EventPeriod>('upcoming');
  const [category, setCategory] = useState<CategoryFilter>('all');
  const filter = { category: category === 'all' ? undefined : category, period };
  const { state, reload } = useCalendarEvents(filter);

  return (
    <Screen
      title="Gestión del calendario"
      subtitle="Alta, modificación y baja de carreras, pruebas de neumáticos y controles técnicos."
      headerAction={
        <View style={styles.actions}>
          <CalendarDownloads filter={filter} />
          <Button label="Nuevo evento" icon="add" compact onPress={() => router.push('/fia/calendar/new')} />
        </View>
      }
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

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm },
});
