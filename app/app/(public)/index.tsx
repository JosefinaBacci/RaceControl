import { router } from 'expo-router';
import { StyleSheet, View } from 'react-native';

import { Button, CardLink, Screen } from '@/components';
import { useSanctions } from '@/data/hooks';
import { useStandings } from '@/data/scores';
import { NextRaceHero } from '@/features/NextRaceHero';
import { RemoteContent } from '@/features/RemoteContent';
import { SanctionList } from '@/features/SanctionList';
import { driverRows } from '@/features/scores/standingsRows';
import { StandingsTable } from '@/features/StandingsTable';
import { spacing } from '@/theme';

const topDriverCount = 4;

export default function PublicHomeScreen() {
  const standings = useStandings('f1');
  const latestSanctions = useSanctions('all').slice(0, 3);

  return (
    <Screen>
      <View style={styles.columns}>
        <View style={styles.column}>
          <NextRaceHero>
            <View style={styles.heroAction}>
              <Button label="Ver calendario" compact trailingIcon="arrow-forward" onPress={() => router.push('/calendar')} />
            </View>
          </NextRaceHero>
        </View>
        <View style={styles.column}>
          <RemoteContent state={standings.state} onRetry={standings.reload}>
            {(data) => (
              <StandingsTable
                rows={driverRows(data.drivers.slice(0, topDriverCount))}
                action={<CardLink label="Ver todas" onPress={() => router.push('/standings')} />}
              />
            )}
          </RemoteContent>
          <SanctionList
            title="Últimas sanciones"
            sanctions={latestSanctions}
            action={<CardLink label="Ver todas" onPress={() => router.push('/sanctions')} />}
          />
        </View>
      </View>
    </Screen>
  );
}

const styles = StyleSheet.create({
  columns: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.lg },
  column: { flexGrow: 1, flexBasis: 360, gap: spacing.lg },
  heroAction: { flexDirection: 'row', marginTop: spacing.sm },
});
