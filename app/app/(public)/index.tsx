import { router } from 'expo-router';
import { StyleSheet, View } from 'react-native';

import { Button, CardLink, Screen } from '@/components';
import { useSanctions, useStandings } from '@/data/hooks';
import { NextRaceHero } from '@/features/NextRaceHero';
import { SanctionList } from '@/features/SanctionList';
import { StandingsTable } from '@/features/StandingsTable';
import { spacing } from '@/theme';

export default function PublicHomeScreen() {
  const topDrivers = useStandings('f1', 4);
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
          <StandingsTable rows={topDrivers} action={<CardLink label="Ver todas" onPress={() => router.push('/standings')} />} />
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
