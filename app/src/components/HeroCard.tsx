import { LinearGradient } from 'expo-linear-gradient';
import type { ReactNode } from 'react';
import { StyleSheet, View } from 'react-native';

import { colors, radius, spacing } from '@/theme';

import { AppText } from './AppText';

type HeroCardProps = {
  overline: string;
  title: string;
  subtitle?: string;
  accentColor?: string;
  children?: ReactNode;
};

export function HeroCard({ overline, title, subtitle, accentColor = colors.accent, children }: HeroCardProps) {
  return (
    <View style={styles.card}>
      <LinearGradient
        colors={[accentColor, '#3A0A0A', colors.surface]}
        locations={[0, 0.45, 1]}
        start={{ x: 1, y: 0 }}
        end={{ x: 0, y: 1 }}
        style={StyleSheet.absoluteFill}
      />
      <View style={styles.stripes}>
        {[0, 1, 2, 3].map((index) => (
          <View key={index} style={[styles.stripe, { opacity: 0.08 + index * 0.04 }]} />
        ))}
      </View>
      <View style={styles.content}>
        <View style={styles.overline}>
          <View style={[styles.flag, { backgroundColor: accentColor }]} />
          <AppText variant="overline" color="text">
            {overline}
          </AppText>
        </View>
        <AppText variant="title" accessibilityRole="header">
          {title}
        </AppText>
        {subtitle ? <AppText color="textMuted">{subtitle}</AppText> : null}
        {children}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    minHeight: 220,
    borderRadius: radius.lg,
    overflow: 'hidden',
    borderWidth: 1,
    borderColor: colors.border,
    justifyContent: 'flex-end',
    backgroundColor: colors.surface,
  },
  stripes: {
    position: 'absolute',
    top: -40,
    right: -20,
    flexDirection: 'row',
    gap: 14,
    transform: [{ skewX: '-24deg' }],
  },
  stripe: { width: 26, height: 260, backgroundColor: colors.text },
  content: { padding: spacing.xl, gap: spacing.sm },
  overline: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm },
  flag: { width: 18, height: 12, borderRadius: 2 },
});
