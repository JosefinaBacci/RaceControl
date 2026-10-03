import type { ReactNode } from 'react';
import { LinearGradient } from 'expo-linear-gradient';
import { StyleSheet, View } from 'react-native';

import { AppText, Badge } from '@/components';
import type { Driver } from '@/mocks/data';
import { colors, radius, spacing } from '@/theme';

export function DriverCard({ driver, teamColor }: { driver: Driver; teamColor: string }) {
  const [firstName, ...lastNames] = driver.name.split(' ');

  return (
    <View style={styles.card}>
      <LinearGradient colors={[teamColor, colors.surface]} locations={[0, 0.85]} start={{ x: 1, y: 0 }} end={{ x: 0, y: 1 }} style={StyleSheet.absoluteFill} />
      <AppText style={styles.number}>{driver.number}</AppText>
      <View style={styles.footer}>
        <AppText variant="caption" color="textMuted">
          {firstName}
        </AppText>
        <AppText variant="heading">{lastNames.join(' ')}</AppText>
        <Badge label={driver.status === 'starter' ? 'Titular' : 'Suplente'} tone={driver.status === 'starter' ? 'accent' : 'neutral'} />
      </View>
    </View>
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
