import { router, useLocalSearchParams } from 'expo-router';

import { AppText, Button, Card, Notice, Screen } from '@/components';
import { useEventDetail } from '@/data/calendar';
import { RemoteContent } from '@/features/RemoteContent';
import { StandingsTable } from '@/features/StandingsTable';
import { driverRows } from '@/features/scores/standingsRows';
import { formatDateTime } from '@/format';

export default function EventDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const { state, reload } = useEventDetail(Number(id));
  return (
    <Screen title="Detalle del evento" headerAction={
      <Button label="Volver al calendario" icon="arrow-back" variant="ghost"
        onPress={() => router.canGoBack() ? router.back() : router.replace('/calendar')} />
    }>
      <RemoteContent state={state} onRetry={reload}>
        {({ event, results }) => <>
          <Card title={event.name}>
            <AppText>{event.categoryName} · Temporada {event.seasonYear}</AppText>
            <AppText>{event.circuitName ?? event.controlType}</AppText>
            <AppText>Inicio: {formatDateTime(event.startsAt)}</AppText>
            <AppText>Fin: {formatDateTime(event.endsAt)}</AppText>
          </Card>
          {results ? results.results.length > 0 ? (
            <StandingsTable title="Resultados de la carrera" rows={driverRows(results.results)} />
          ) : <Notice message="Todavía no hay puntajes publicados para esta carrera." /> : null}
        </>}
      </RemoteContent>
    </Screen>
  );
}
