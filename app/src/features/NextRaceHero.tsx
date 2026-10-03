import type { ReactNode } from 'react';
import { StyleSheet, View, useWindowDimensions } from 'react-native';

import { AppText, Card, Countdown, HeroCard } from '@/components';
import { useNextRace } from '@/data/hooks';
import { formatEventDate } from '@/format';
import { CircuitMap, circuitKeyFor } from '@/illustrations';
import { nextRacePhoto } from '@/photos';
import { colors, spacing } from '@/theme';

export function NextRaceHero({ children }: { children?: ReactNode }) {
  const { race, startsAt, facts } = useNextRace();
  const mapWidth = useWindowDimensions().width < 480 ? 110 : 150;

  return (
    <View style={styles.stack}>
      <HeroCard
        overline="Próxima carrera"
        title={race.name}
        subtitle={`${race.circuit} · ${formatEventDate(race.date)}`}
        flag={race.country}
        photo={nextRacePhoto(race.category)}
        illustration={<CircuitMap circuit={circuitKeyFor(race.circuit)} width={mapWidth} strokeWidth={6} color="#FFFFFF" showStart halo />}
        illustrationSide="left"
      >
        {children}
      </HeroCard>
      <Countdown target={startsAt} />
      {facts ? (
        <Card>
          <Fact label="Longitud" value={facts.length} />
          <Fact label="Vueltas" value={String(facts.laps)} />
          <Fact label="Distancia total" value={facts.distance} />
          <Fact label="Récord de vuelta" value={facts.lapRecord} isLast />
        </Card>
      ) : null}
    </View>
  );
}

function Fact({ label, value, isLast = false }: { label: string; value: string; isLast?: boolean }) {
  return (
    <View style={[styles.fact, !isLast && styles.factDivider]}>
      <AppText variant="caption" color="textMuted">
        {label}
      </AppText>
      <AppText variant="bodyStrong">{value}</AppText>
    </View>
  );
}

const styles = StyleSheet.create({
  stack: { gap: spacing.md },
  fact: { flexDirection: 'row', justifyContent: 'space-between', gap: spacing.md, paddingVertical: spacing.xs },
  factDivider: { borderBottomWidth: StyleSheet.hairlineWidth, borderBottomColor: colors.border, paddingBottom: spacing.sm },
});
