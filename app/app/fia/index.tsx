import { router } from 'expo-router';
import { StyleSheet, View } from 'react-native';

import { useSession } from '@/auth/SessionProvider';
import { Button, Card, CardLink, Screen, StatRow, StatTile } from '@/components';
import { useActiveAccountCount } from '@/data/accounts';
import { usePendingAcknowledgements, useUpcomingEvents } from '@/data/hooks';
import { EventList } from '@/features/EventList';
import { MockNotice } from '@/features/MockNotice';
import { spacing } from '@/theme';

export default function FiaDashboardScreen() {
  const session = useSession();
  const upcoming = useUpcomingEvents(4);
  const pending = usePendingAcknowledgements();
  const activeAccounts = useActiveAccountCount();

  return (
    <Screen title={`Hola, ${session.user?.username ?? ''}`} subtitle="Administración central de la FIA.">
      <MockNotice />
      <StatRow>
        <StatTile label="Próximos eventos" value={upcoming.length} />
        <StatTile label="Sanciones sin acuse" value={pending.sanctions} highlight={pending.sanctions > 0} />
        <StatTile label="Puntajes sin acuse" value={pending.scores} highlight={pending.scores > 0} />
        <StatTile label="Cuentas activas" value={activeAccounts ?? '—'} />
      </StatRow>
      <Card title="Acciones rápidas">
        <View style={styles.actions}>
          <Button label="Nuevo usuario" icon="person-add-outline" compact onPress={() => router.push('/fia/users/new')} />
          <Button label="Cargar evento" icon="calendar-outline" variant="secondary" compact onPress={() => router.push('/fia/calendar')} />
          <Button label="Publicar puntajes" icon="trophy-outline" variant="secondary" compact onPress={() => router.push('/fia/results')} />
          <Button label="Enviar comunicado" icon="megaphone-outline" variant="secondary" compact onPress={() => router.push('/fia/notifications')} />
        </View>
      </Card>
      <EventList title="Agenda inmediata" events={upcoming} action={<CardLink label="Ver calendario" onPress={() => router.push('/fia/calendar')} />} />
    </Screen>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm },
});
