import { useState } from 'react';
import { Linking, StyleSheet, View } from 'react-native';

import type { EventPeriod } from '@/api/calendar';
import { Button, ChipGroup, Screen } from '@/components';
import { calendarDownloadUrl, useCalendarEvents } from '@/data/calendar';
import { EventList } from '@/features/EventList';
import { categoryFilterOptions, periodOptions } from '@/features/filters';
import { RemoteContent } from '@/features/RemoteContent';
import type { CategoryFilter } from '@/mocks/catalog';

export default function PublicCalendarScreen() {
  const [period, setPeriod] = useState<EventPeriod>('upcoming');
  const [category, setCategory] = useState<CategoryFilter>('all');
  const categoryCode = category === 'all' ? undefined : category;

  const { state, reload } = useCalendarEvents({ category: categoryCode, period });

  const download = (format: 'ics' | 'pdf') =>
    Linking.openURL(calendarDownloadUrl(format, { category: categoryCode, period }));

  return (
    <Screen
      title="Calendario 2026"
      subtitle="Carreras, pruebas de neumáticos y controles técnicos."
      headerAction={
        <View style={styles.downloads}>
          <Button label=".ics" compact variant="secondary" onPress={() => download('ics')} />
          <Button label="PDF" compact variant="secondary" onPress={() => download('pdf')} />
        </View>
      }
    >
      <ChipGroup variant="underline" options={periodOptions} selected={period} onSelect={setPeriod} accessibilityLabel="Período" />
      <ChipGroup options={categoryFilterOptions} selected={category} onSelect={setCategory} accessibilityLabel="Filtrar por categoría" />
      <RemoteContent state={state} onRetry={reload}>
        {(events) => <EventList events={events} />}
      </RemoteContent>
    </Screen>
  );
}

const styles = StyleSheet.create({
  downloads: { flexDirection: 'row', gap: 8 },
});
