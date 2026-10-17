import { router } from 'expo-router';
import { StyleSheet, View } from 'react-native';

import { useSession } from '@/auth/SessionProvider';
import { Button, Card, CardLink, Notice, Screen, StatRow, StatTile } from '@/components';
import { useActiveAccountCount } from '@/data/accounts';
import { useUpcomingEvents } from '@/data/calendar';
import { usePendingSanctions } from '@/data/hooks';
import { usePendingScoreConfirmations, useFinishedRaceScores } from '@/data/scores';
import { EventList } from '@/features/EventList';
import { RemoteContent } from '@/features/RemoteContent';
import { MockNotice } from '@/features/MockNotice';
import { spacing } from '@/theme';

export default function FiaDashboardScreen() {
  const session = useSession();
  const { state: upcomingState, reload: reloadUpcoming } = useUpcomingEvents(4);
  const pendingSanctions = usePendingSanctions();
  const pendingScores = usePendingScoreConfirmations();
  const finishedRaces = useFinishedRaceScores();
  const missingResults = finishedRaces.state.status === 'ready' ? finishedRaces.state.data.filter((race) => !race.hasResults).length : 0;
  const activeAccounts = useActiveAccountCount();

  return (
    <Screen title={`Hola, ${session.user?.username ?? ''}`} subtitle="Administración central de la FIA.">
      <MockNotice />
      {missingResults > 0 ? <Card title="Puntajes pendientes" action={<CardLink label="Cargar resultados" onPress={() => router.push('/fia/results')} />}>
        <Notice tone="warning" message={`Hay ${missingResults} carreras terminadas sin puntajes publicados.`} />
      </Card> : null}
      <StatRow>
        <StatTile label="Próximos eventos" value={upcomingState.status === 'ready' ? upcomingState.data.length : 0} />
        <StatTile label="Sanciones sin acuse" value={pendingSanctions} highlight={pendingSanctions > 0} />
        <StatTile label="Puntajes sin acuse" value={pendingScores} highlight={pendingScores > 0} />
        <StatTile label="Cuentas activas" value={activeAccounts ?? '—'} />
      </StatRow>
      <Card title="Acciones rápidas">
        <View style={styles.actions}>
          <Button label="Nuevo usuario" icon="person-add-outline" compact onPress={() => router.push('/fia/users/new', { withAnchor: true })} />
          <Button label="Cargar evento" icon="calendar-outline" variant="secondary" compact onPress={() => router.push('/fia/calendar')} />
          <Button label="Publicar puntajes" icon="trophy-outline" variant="secondary" compact onPress={() => router.push('/fia/results')} />
          <Button label="Acuses de puntajes" icon="checkmark-done-outline" variant="secondary" compact onPress={() => router.push('/fia/notifications')} />
        </View>
      </Card>
      <RemoteContent state={upcomingState} onRetry={reloadUpcoming}>
        {(upcoming) => (
          <EventList title="Agenda inmediata" events={upcoming} action={<CardLink label="Ver calendario" onPress={() => router.push('/fia/calendar')} />} />
        )}
      </RemoteContent>
    </Screen>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm },
});
