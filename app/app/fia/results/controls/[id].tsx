import { router, useLocalSearchParams } from 'expo-router';

import type { ControlSheet } from '@/api/technicalControls';
import { AppText, Button, Card, Notice, Screen } from '@/components';
import { useControlSheet } from '@/data/technicalControls';
import { dayOfInstant, formatInstantDay } from '@/features/history/day';
import { RemoteContent } from '@/features/RemoteContent';
import { ControlTeamCard } from '@/features/technicalControls/ControlTeamCard';

export default function ControlResultsScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const { state, reload } = useControlSheet(Number(id));

  return (
    <Screen
      title="Resultados del control técnico"
      subtitle="Indicá si el auto de cada escudería pasó el control. Un resultado registrado no se borra: se corrige."
      headerAction={<Button label="Volver" icon="arrow-back" variant="ghost" compact onPress={() => router.back()} />}
    >
      <RemoteContent state={state} onRetry={reload}>
        {(sheet) => <SheetContent sheet={sheet} onSaved={reload} />}
      </RemoteContent>
    </Screen>
  );
}

function SheetContent({ sheet, onSaved }: { sheet: ControlSheet; onSaved: () => void }) {
  const { control } = sheet;
  const days = `${formatInstantDay(control.startsAt)} – ${formatInstantDay(control.endsAt)}`;

  return (
    <>
      <Card title={control.controlType}>
        <AppText color="textMuted">{`${control.categoryName} ${control.seasonYear} · ${days}`}</AppText>
        {!control.isActive ? <Notice tone="warning" message="El control está cancelado: no admite resultados." /> : null}
      </Card>
      {sheet.teams.map((team) => (
        <ControlTeamCard
          key={team.id}
          eventId={control.eventId}
          team={team}
          defaultDay={dayOfInstant(control.startsAt)}
          onSaved={onSaved}
        />
      ))}
    </>
  );
}
