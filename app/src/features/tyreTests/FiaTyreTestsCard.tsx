import { router } from 'expo-router';

import { AppText, Badge, Button, Card, ListItem } from '@/components';
import { useTyreTestEvents } from '@/data/tyreTests';
import { formatInstantDay } from '@/features/history/day';
import { RemoteContent } from '@/features/RemoteContent';

import { tyreTestResultsHref } from './routes';

// The tyre tests of the calendar on the FIA "Resultados" tab, each opening its
// record, and the shortcut that schedules a new one.
export function FiaTyreTestsCard() {
  const { state, reload } = useTyreTestEvents();

  return (
    <Card
      title="Pruebas de neumáticos"
      action={
        <Button
          label="Nueva prueba de neumáticos"
          icon="add"
          variant="secondary"
          compact
          onPress={() => router.push({ pathname: '/fia/calendar/new', params: { kind: 'tyre_test' } })}
        />
      }
    >
      <RemoteContent state={state} onRetry={reload}>
        {(events) =>
          events.length === 0 ? (
            <AppText color="textMuted">Todavía no hay pruebas de neumáticos en el calendario.</AppText>
          ) : (
            events.map((event, index) => (
              <ListItem
                key={event.id}
                title={event.circuitName ?? event.name}
                subtitle={`${event.categoryName} · ${formatInstantDay(event.startsAt)}`}
                trailing={<Badge label={event.status === 'finished' ? 'Finalizada' : 'Programada'} tone="neutral" />}
                onPress={() => router.push(tyreTestResultsHref(event.id))}
                isLast={index === events.length - 1}
              />
            ))
          )
        }
      </RemoteContent>
    </Card>
  );
}
