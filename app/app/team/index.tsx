import { router } from 'expo-router';
import { StyleSheet, View } from 'react-native';

import { useSession } from '@/auth/SessionProvider';
import { Button, CardLink, HeroCard, Screen, StatRow, StatTile } from '@/components';
import { useUpcomingEvents } from '@/data/calendar';
import { useTeamDrivers } from '@/data/drivers';
import { useTeamInbox } from '@/data/hooks';
import { useTeamPoints } from '@/data/scores';
import { usePendingScoreNotices } from '@/data/TeamScoreNoticesProvider';
import { DriverCard, DriverGrid } from '@/features/DriverCard';
import { EventList } from '@/features/EventList';
import { RemoteContent } from '@/features/RemoteContent';
import { MockNotice } from '@/features/MockNotice';
import { formatPoints } from '@/format';
import { categoryLabel, teamById } from '@/mocks/catalog';
import { teamPhoto } from '@/photos';
import { spacing } from '@/theme';

export default function TeamDashboardScreen() {
  const teamId = useSession().user?.teamId ?? null;
  const team = teamById(teamId);
  const drivers = useTeamDrivers();
  const points = useTeamPoints(teamId);
  const pendingScores = usePendingScoreNotices();
  const { pendingCount: pendingSanctions } = useTeamInbox(teamId);
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
        {pendingScores > 0 ? (
          <View style={styles.heroAction}>
            <Button
              label={`Confirmar puntajes (${pendingScores})`}
              icon="alert-circle-outline"
              compact
              onPress={() => router.push('/team/notifications')}
            />
          </View>
        ) : null}
      </HeroCard>
      <MockNotice />
      <StatRow>
        <StatTile label="Pilotos" value={drivers.state.status === 'ready' ? drivers.state.data.length : 0} />
        <StatTile label="Puntos" value={points === null ? '—' : formatPoints(points)} />
        <StatTile label="Puntajes sin confirmar" value={pendingScores} highlight={pendingScores > 0} />
        <StatTile label="Sanciones sin acuse" value={pendingSanctions} highlight={pendingSanctions > 0} />
      </StatRow>
      <View style={styles.sectionHeader}>
        <CardLink label="Gestionar pilotos" onPress={() => router.push('/team/drivers')} />
      </View>
      <RemoteContent state={drivers.state} onRetry={drivers.reload}>
        {(lineup) => (
          <DriverGrid>
            {lineup.map((driver) => (
              <DriverCard key={driver.id} driver={driver} teamColor={team.color} />
            ))}
          </DriverGrid>
        )}
      </RemoteContent>
      <RemoteContent state={upcoming.state} onRetry={upcoming.reload}>
        {(events) => <EventList title="Próximos eventos" events={events} />}
      </RemoteContent>
    </Screen>
  );
}

const styles = StyleSheet.create({
  heroAction: { flexDirection: 'row', marginTop: spacing.sm },
  sectionHeader: { flexDirection: 'row', justifyContent: 'flex-end' },
});
