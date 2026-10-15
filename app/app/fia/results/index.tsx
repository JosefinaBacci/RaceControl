import { router } from 'expo-router';

import { AppText, Card, ListItem, Screen } from '@/components';
import { useCalendarEvents } from '@/data/calendar';
import { formatEventDate } from '@/format';
import { RemoteContent } from '@/features/RemoteContent';

export default function FiaResultsScreen() {
  const { state, reload } = useCalendarEvents({ kind: 'race' });

  return (
    <Screen title="Puntajes" subtitle="Cargá y modificá los resultados de cada carrera.">
      <Card title="Carreras">
        <RemoteContent state={state} onRetry={reload}>
          {(races) =>
            races.length === 0 ? (
              <AppText color="textMuted">Todavía no hay carreras en el calendario.</AppText>
            ) : (
              races.map((race, index) => (
                <ListItem
                  key={race.id}
                  title={race.name}
                  subtitle={`${race.categoryName} · ${formatEventDate(race.startsAt)}`}
                  onPress={() => router.push({ pathname: '/fia/results/[id]', params: { id: String(race.id) } })}
                  isLast={index === races.length - 1}
                />
              ))
            )
          }
        </RemoteContent>
      </Card>
    </Screen>
  );
}
