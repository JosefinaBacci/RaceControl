import { router } from 'expo-router';
import { StyleSheet, View } from 'react-native';

import { useSession } from '@/auth/SessionProvider';
import { Button, CardLink, HeroCard, Screen, StatRow, StatTile } from '@/components';
import { useTeamDrivers, useTeamInbox, useTeamPoints, useUpcomingEvents } from '@/data/hooks';
import { DriverCard, DriverGrid } from '@/features/DriverCard';
import { EventList } from '@/features/EventList';
import { MockNotice } from '@/features/MockNotice';
import { categoryLabel, teamById } from '@/mocks/catalog';
import { teamPhoto } from '@/photos';
import { spacing } from '@/theme';

export default function TeamDashboardScreen() {
  const teamId = useSession().user?.teamId ?? null;
  const team = teamById(teamId);
  const drivers = useTeamDrivers(teamId);
  const points = useTeamPoints(teamId);
  const { pendingCount: pending } = useTeamInbox(teamId);
  const upcoming = useUpcomingEvents(3);

  return (
    <Screen>
      <HeroCard
        overline={`${categoryLabel(team.category)} · ${team.country}`}
        title={team.name}
        accentColor={team.color}
        flag={team.flag}
        photo={teamPhoto(team.id, team.category)}
      >
        {pending > 0 ? (
          <View style={styles.heroAction}>
            <Button label={`${pending} pendientes de acuse`} icon="alert-circle-outline" compact onPress={() => router.push('/team/inbox')} />
          </View>
        ) : null}
      </HeroCard>
      <MockNotice />
      <StatRow>
        <StatTile label="Pilotos" value={drivers.length} />
        <StatTile label="Puntos" value={points} />
        <StatTile label="Pendientes" value={pending} highlight={pending > 0} />
      </StatRow>
      <View style={styles.sectionHeader}>
        <CardLink label="Gestionar pilotos" onPress={() => router.push('/team/drivers')} />
      </View>
      <DriverGrid>
        {drivers.map((driver) => (
          <DriverCard key={driver.id} driver={driver} teamColor={team.color} />
        ))}
      </DriverGrid>
      <EventList title="Próximos eventos" events={upcoming} />
    </Screen>
  );
}

const styles = StyleSheet.create({
  heroAction: { flexDirection: 'row', marginTop: spacing.sm },
  sectionHeader: { flexDirection: 'row', justifyContent: 'flex-end' },
});
