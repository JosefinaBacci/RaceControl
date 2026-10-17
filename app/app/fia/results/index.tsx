import { router } from 'expo-router';

import { AppText, Badge, Card, ListItem, Screen } from '@/components';
import { useFinishedRaceScores } from '@/data/scores';
import { formatEventDate } from '@/format';
import { RemoteContent } from '@/features/RemoteContent';

// Only finished races are listed: results freeze a race's date, so the API refuses
// them before the race ends.
export default function FiaResultsScreen() {
  const { state, reload } = useFinishedRaceScores();

  return (
    <Screen title="Puntajes" subtitle="Cargá y modificá los resultados de cada carrera terminada.">
      <Card title="Carreras">
        <RemoteContent state={state} onRetry={reload}>
          {(races) =>
            races.length === 0 ? (
              <AppText color="textMuted">Todavía no terminó ninguna carrera.</AppText>
            ) : (
              races.map((race, index) => (
                <ListItem
                  key={race.id}
                  title={race.name}
                  trailing={<Badge label={race.hasResults ? 'Publicados' : 'Sin puntajes'} tone={race.hasResults ? 'neutral' : 'warning'} />}
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
