import { router, useLocalSearchParams } from 'expo-router';

import type { CalendarEvent } from '@/api/calendar';
import type { HistoryType } from '@/api/history';
import type { TyreTestRevision } from '@/api/tyreTests';
import { AppText, Button, Card, Notice, Screen } from '@/components';
import { useTyreTestRecord } from '@/data/tyreTests';
import { dayOfInstant, formatInstantDay, todayDay } from '@/features/history/day';
import { RemoteContent } from '@/features/RemoteContent';
import { CompoundRecordCard } from '@/features/tyreTests/CompoundRecordCard';
import { NewTyreResultForm } from '@/features/tyreTests/NewTyreResultForm';
import { compoundRecords, untestedCompounds } from '@/features/tyreTests/tyreRecord';

export default function TyreTestResultsScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const { state, reload } = useTyreTestRecord(Number(id));

  return (
    <Screen
      title="Resultados de la prueba de neumáticos"
      subtitle="Un resultado por compuesto. Lo registrado no se edita: se corrige con un motivo."
      headerAction={<Button label="Volver" icon="arrow-back" variant="ghost" compact onPress={() => router.back()} />}
    >
      <RemoteContent state={state} onRetry={reload}>
        {({ event, compounds, revisions }) =>
          event.kind === 'tyre_test' ? (
            <TyreTestRecord event={event} compounds={compounds} revisions={revisions} onChanged={reload} />
          ) : (
            <Notice tone="danger" message="El evento no es una prueba de neumáticos." />
          )
        }
      </RemoteContent>
    </Screen>
  );
}

type TyreTestRecordProps = {
  event: CalendarEvent;
  compounds: HistoryType[];
  revisions: TyreTestRevision[];
  onChanged: () => void;
};

function TyreTestRecord({ event, compounds, revisions, onChanged }: TyreTestRecordProps) {
  const firstDay = dayOfInstant(event.startsAt);
  const hasStarted = todayDay() >= firstDay;

  return (
    <>
      <Card title={event.name}>
        <AppText color="textMuted">
          {`${event.circuitName ?? ''} · ${event.categoryName} ${event.seasonYear} · ${formatInstantDay(event.startsAt)}`}
        </AppText>
      </Card>
      {!event.isActive ? <Notice tone="warning" message="La prueba fue cancelada: no admite resultados." /> : null}
      {event.isActive && hasStarted ? (
        <NewTyreResultForm
          eventId={event.id}
          compounds={untestedCompounds(compounds, revisions)}
          defaultDay={firstDay}
          onSaved={onChanged}
        />
      ) : (
        <Notice tone="info" message="La prueba todavía no ocurrió: sus resultados se registran desde el primer día." />
      )}
      {compoundRecords(revisions).map((record) => (
        <CompoundRecordCard key={record.compound.code} record={record} onCorrected={onChanged} />
      ))}
    </>
  );
}
