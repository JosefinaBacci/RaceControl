import { LinearGradient } from 'expo-linear-gradient';
import type { ReactNode } from 'react';
import { Pressable, StyleSheet, View } from 'react-native';

import type { TeamDriver } from '@/api/drivers';
import { AppText, Badge } from '@/components';
import { colors, radius, spacing } from '@/theme';

type DriverCardProps = {
  driver: TeamDriver;
  teamColor: string;
  onPress?: () => void;
};

export function DriverCard({ driver, teamColor, onPress }: DriverCardProps) {
  return (
    <Pressable
      accessibilityRole={onPress ? 'button' : undefined}
      accessibilityLabel={onPress ? `Editar ${driver.firstName} ${driver.lastName}` : undefined}
      onPress={onPress}
      style={styles.card}
    >
      <LinearGradient colors={[teamColor, colors.surface]} locations={[0, 0.85]} start={{ x: 1, y: 0 }} end={{ x: 0, y: 1 }} style={StyleSheet.absoluteFill} />
      <AppText style={styles.number}>{driver.number ?? '—'}</AppText>
      <View style={styles.footer}>
        <AppText variant="caption" color="textMuted">
          {driver.firstName}
        </AppText>
        <AppText variant="heading">{driver.lastName}</AppText>
        <Badge label={driver.isStarter ? 'Titular' : 'Suplente'} tone={driver.isStarter ? 'accent' : 'neutral'} />
      </View>
    </Pressable>
  );
}

export function DriverGrid({ children }: { children: ReactNode }) {
  return <View style={styles.grid}>{children}</View>;
}

const styles = StyleSheet.create({
  grid: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.md },
  card: {
    flexGrow: 1,
    flexBasis: 150,
    maxWidth: 260,
    minHeight: 190,
    borderRadius: radius.lg,
    borderWidth: 1,
    borderColor: colors.border,
    overflow: 'hidden',
    padding: spacing.lg,
    justifyContent: 'space-between',
    backgroundColor: colors.surface,
  },
  number: { fontSize: 48, fontWeight: '900', fontStyle: 'italic', color: colors.text },
  footer: { gap: spacing.xs },
});
