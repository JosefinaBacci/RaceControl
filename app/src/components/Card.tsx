import Ionicons from '@expo/vector-icons/Ionicons';
import type { ReactNode } from 'react';
import { Pressable, StyleSheet, View, type ViewStyle } from 'react-native';

import { colors, radius, spacing } from '@/theme';

import { AppText } from './AppText';

type CardProps = {
  title?: string;
  action?: ReactNode;
  children: ReactNode;
  style?: ViewStyle;
  flush?: boolean;
};

export function Card({ title, action, children, style, flush = false }: CardProps) {
  return (
    <View style={[styles.card, flush && styles.flush, style]}>
      {title ? (
        <View style={[styles.header, flush && styles.flushHeader]}>
          <AppText variant="heading">{title}</AppText>
          {action}
        </View>
      ) : null}
      {children}
    </View>
  );
}

export function CardLink({ label, onPress }: { label: string; onPress: () => void }) {
  return (
    <Pressable accessibilityRole="link" onPress={onPress} hitSlop={8} style={styles.link}>
      <AppText variant="caption" color="accent">
        {label}
      </AppText>
      <Ionicons name="chevron-forward" size={14} color={colors.accent} />
    </Pressable>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: colors.surface,
    borderRadius: radius.lg,
    borderWidth: 1,
    borderColor: colors.border,
    padding: spacing.lg,
    gap: spacing.md,
  },
  flush: { padding: 0, gap: 0, overflow: 'hidden' },
  flushHeader: { padding: spacing.lg },
  header: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: spacing.sm },
  link: { flexDirection: 'row', alignItems: 'center', gap: 2 },
});
