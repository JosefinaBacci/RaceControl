import { router } from 'expo-router';
import { StyleSheet, View } from 'react-native';

import { useSession } from '@/auth/SessionProvider';
import { Button, Card, CardLink, Screen, StatRow, StatTile } from '@/components';
import { useManagedUsers, useSanctions, useScoreNotices, useUpcomingEvents } from '@/data/hooks';
import { EventList } from '@/features/EventList';
import { MockNotice } from '@/features/MockNotice';
import { spacing } from '@/theme';

export default function FiaDashboardScreen() {
  const session = useSession();
  const upcoming = useUpcomingEvents(4);
  const pendingSanctions = useSanctions('all').filter((sanction) => !sanction.acknowledged).length;
  const pendingScores = useScoreNotices().filter((notice) => !notice.acknowledged).length;
  const activeAccounts = useManagedUsers('', 'all').filter((user) => user.isActive).length;

  return (
    <Screen title={`Hola, ${session.user?.username ?? ''}`} subtitle="Administración central de la FIA.">
      <MockNotice />
      <StatRow>
        <StatTile label="Próximos eventos" value={upcoming.length} />
        <StatTile label="Sanciones sin acuse" value={pendingSanctions} highlight={pendingSanctions > 0} />
        <StatTile label="Puntajes sin acuse" value={pendingScores} highlight={pendingScores > 0} />
        <StatTile label="Cuentas activas" value={activeAccounts} />
      </StatRow>
      <Card title="Acciones rápidas">
        <View style={styles.actions}>
          <Button label="Nuevo usuario" icon="person-add-outline" compact onPress={() => router.push('/fia/users')} />
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
